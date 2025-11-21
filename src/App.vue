<script setup lang="ts">
import { computed, onMounted, onUnmounted, provide, ref, watch } from "vue";
import { toast, Toaster } from "vue-sonner";
import "vue-sonner/style.css";
import type {
    Chat,
    ConfirmDialogOptions,
    LinkPreview,
    Message,
    UserDetails,
} from "./types";
import {
    generateChatTitle,
    extractUrls,
    getTransaction,
    createErrorContext,
    getErrorStatus,
    SPINDLE_URL,
} from "./lib/globals";
import { nextTick } from "vue";
import { detectAndProcessVideo } from "./lib/videoProcessing";
import ConfirmDialog from "./components/ConfirmDialog.vue";
import type { Theme } from "vue-sonner/src/packages/types.js";
import UpdateModal from "./components/Modals/UpdateModal.vue";
import { usePlatformError } from "./composables/usePlatformError";
import { useApiCall } from "@/composables/useApiCall";
import type { PlatformError } from "./types";
import { useDebounceFn } from "@vueuse/core";
import DemoToast from "./components/DemoToast.vue";
import { useRouter } from "vue-router";
import { useChat } from "./composables/useChat";
import { useCache } from "./composables/useCache";
import { useSync } from "./composables/useSync";
import { useHandlePaste } from "./composables/useHandlePaste";

const { reportError } = usePlatformError();
const router = useRouter();
const screenWidth = ref(screen.width);
const isDarkMode = ref(false);
const scrollableElem = ref<HTMLElement | null>(null);
const showScrollDownButton = ref(false);
const activeRequests = ref<Map<string, AbortController>>(new Map());
const requestChatMap = ref<Map<string, string>>(new Map());
const pendingResponses = ref<Map<string, { prompt: string; chatId: string }>>(
    new Map(),
);
const chatDrafts = ref<Map<string, string>>(new Map());
const pastePreviews = ref<
    Map<
        string,
        {
            content: string;
            wordCount: number;
            charCount: number;
            show: boolean;
        }
    >
>(new Map());

const confirmDialog = ref<ConfirmDialogOptions>({
    visible: false,
    title: "",
    message: "",
    type: undefined,
    confirmText: "",
    cancelText: "",
    onConfirm: () => {},
    onCancel: () => {},
});
const isCollapsed = ref(true);
const isSidebarHidden = ref(true);
const authData = ref({
    username: "",
    email: "",
    password: "",
    agreeToTerms: false,
});
const FREE_REQUEST_LIMIT = 20;
const isOpenTextHighlightPopover = ref(false);

const parsedUserDetailsNullValues: UserDetails = {
    agreeToTerms: false,
    createdAt: new Date(),
    email: "",
    emailSubscribed: false,
    emailVerified: false,
    userId: "",
    syncEnabled: true,
    username: "",
    theme: "system",
};

let userDetails: any = localStorage.getItem("userdetails");
const parsedUserDetails = ref<UserDetails>(
    (() => {
        try {
            return userDetails ? JSON.parse(userDetails) : null;
        } catch (error) {
            console.error("Invalid user details in localStorage:", error);
            localStorage.removeItem("userdetails");
            return null;
        }
    })(),
);

const syncEnabled = ref(parsedUserDetails.value?.syncEnabled !== false);

const isDemoMode = computed(() => {
    const user = parsedUserDetails.value;
    if (!user || typeof user !== "object") return false;
    return user.email === "demo@example.com";
});

const isAuthenticated = computed(() => {
    const user = parsedUserDetails.value;
    if (!user || typeof user !== "object") return false;

    return !!(
        user.email &&
        user.username &&
        user.sessionId &&
        /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(user.email)
    );
});

const currentChatId = ref<string>("");
const chats = ref<Chat[]>([]);
const isLoading = ref(false);
const showProfileMenu = ref(false);
const now = ref(Date.now());

// State for preview sidebar
const showPreviewSidebar = ref(false);
const previewCode = ref("");
const previewLanguage = ref("html");
const metadata = ref<
    | {
          wordCount: number;
          charCount: number;
      }
    | undefined
>(undefined);

// Function to open preview with code
const openPreview = (
    code: string,
    language: string = "html",
    data?: {
        wordCount: number;
        charCount: number;
    },
) => {
    previewCode.value = code;
    previewLanguage.value = language;
    showPreviewSidebar.value = true;
    metadata.value = data;
};

// Function to close preview
const closePreview = () => {
    showPreviewSidebar.value = false;
};

const planStatus = computed(() => {
    if (!parsedUserDetails.value || !parsedUserDetails.value.expiryTimestamp) {
        return {
            status: "no-plan",
            timeLeft: "",
            expiryDate: "",
            isExpired: false,
        };
    }

    const expiryMs =
        parsedUserDetails.value.expiryTimestamp < 1e12
            ? parsedUserDetails.value.expiryTimestamp * 1000
            : parsedUserDetails.value.expiryTimestamp;

    const diff = expiryMs - now.value;
    const isExpired = diff <= 0;

    if (isExpired) {
        return {
            status: "expired",
            timeLeft: "Expired",
            expiryDate: "",
            isExpired: true,
        };
    }

    const days = Math.floor(diff / (1000 * 60 * 60 * 24));
    const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
    const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));

    let timeLeft = "";
    if (days > 0) {
        timeLeft = `${days}d ${hours}h ${minutes}m`;
    } else if (hours > 0) {
        timeLeft = `${hours}h ${minutes}m`;
    } else {
        timeLeft = `${minutes}m`;
    }

    const expiryDate = new Date(expiryMs).toLocaleString("en-KE", {
        weekday: "short",
        year: "numeric",
        month: "short",
        day: "numeric",
        hour: "2-digit",
        minute: "2-digit",
    });

    return { status: "active", timeLeft, expiryDate, isExpired: false };
});

const userPlanStatus = computed(() => {
    if (!parsedUserDetails.value || !parsedUserDetails.value.expiryTimestamp) {
        return { status: "no-plan", isExpired: false };
    }

    const expiryMs =
        parsedUserDetails.value.expiryTimestamp < 1e12
            ? parsedUserDetails.value.expiryTimestamp * 1000
            : parsedUserDetails.value.expiryTimestamp;

    const diff = expiryMs - now.value;
    const isExpired = diff <= 0;

    if (isExpired) {
        return { status: "expired", isExpired: true };
    }

    return { status: "active", isExpired: false };
});

const userHasRequestLimits = computed(() => {
    if (!parsedUserDetails.value) return true;

    const hasFreePlan =
        !parsedUserDetails.value.plan ||
        parsedUserDetails.value.plan === "free" ||
        parsedUserDetails.value.plan === "" ||
        userPlanStatus.value.status === "no-plan";

    return hasFreePlan || userPlanStatus.value.isExpired;
});

const requestLimitInfo = computed(() => {
    const hasLimits = userHasRequestLimits.value;
    const currentCount = parsedUserDetails.value.requestCount?.count || 0;

    return {
        hasLimits,
        currentCount,
        limit: FREE_REQUEST_LIMIT,
        isExceeded: hasLimits && currentCount >= FREE_REQUEST_LIMIT,
        shouldShowUpgradePrompt:
            hasLimits &&
            currentCount >= FREE_REQUEST_LIMIT - 2 &&
            currentCount < FREE_REQUEST_LIMIT,
        remaining: hasLimits
            ? Math.max(0, FREE_REQUEST_LIMIT - currentCount)
            : Infinity,
        status: hasLimits
            ? currentCount >= FREE_REQUEST_LIMIT
                ? "exceeded"
                : currentCount >= FREE_REQUEST_LIMIT - 2
                  ? "warning"
                  : "normal"
            : "unlimited",
    };
});

const isRequestLimitExceeded = computed(
    () => requestLimitInfo.value.isExceeded,
);
const shouldShowUpgradePrompt = computed(
    () => requestLimitInfo.value.shouldShowUpgradePrompt,
);
const requestsRemaining = computed(() => requestLimitInfo.value.remaining);
const isFreeUser = computed(() => userHasRequestLimits.value);
const shouldHaveLimit =
    isFreeUser.value ||
    planStatus.value.isExpired ||
    planStatus.value.status === "no-plan" ||
    planStatus.value.status === "expired";
const requestCount = computed(() => requestLimitInfo.value.currentCount);
const copiedIndex = ref<number | null>(null);
const showErrorSection = ref(false);
const fallbackChatId = ref("");

const currentChat = computed(() => {
    if (!currentChatId.value || !chats.value.length) {
        return undefined;
    }
    return chats.value.find((chat) => chat.id === currentChatId.value);
});

const currentMessages = computed(() => {
    return currentChat.value?.messages || [];
});

const isCurrentChatValid = computed(() => {
    return currentChatId.value && currentChat.value !== undefined;
});

// Composables
const { linkPreviewCache, loadLinkPreviewCache, saveLinkPreviewCache } =
    useCache();

const { syncStatus, showSyncIndicator, hideSyncIndicator, updateSyncProgress } =
    useSync();

const { apiCall, checkInternetConnection } = useApiCall({
    parsedUserDetails,
    syncStatus,
});

const isOnline = ref(true);
const connectionStatus = ref<"online" | "offline" | "checking">("online");

const {
    isChatLoading,
    updateChat,
    loadChat,
    createNewChat,
    deleteChat,
    renameChat,
    saveChats,
    clearAllChats,
    expanded,
    activeChatMenu,
    toggleChatMenu,
    copyResponse,
    shareResponse,
    autoGrow,
    saveChatDrafts,
    loadChatDrafts,
    clearCurrentDraft,
    autoSaveDraft,
    processLinksInResponse,
    processLinksInUserPrompt,
    mergeChats,
    draftSaveTimeout,
    apiCall: apiCallFromChat,
    loadChats,
} = useChat({
    isOnline: isOnline.value,
    copiedIndex,
    chats,
    currentChatId,
    updateExpandedArray,
    linkPreviewCache,
    fetchLinkPreview,
    chatDrafts,
    pastePreviews,
    parsedUserDetails,
    isAuthenticated,
    saveLinkPreviewCache,
    confirmDialog,
});

const { handlePaste, removePastePreview } = useHandlePaste({
    currentChatId,
    pastePreviews,
    chatDrafts,
    saveChatDrafts,
    autoGrow,
});

async function logout(options = { skipConfirm: false }) {
    const executeLogout = async () => {
        try {
            isLoading.value = true;

            const syncEnabledValue = parsedUserDetails.value?.syncEnabled;
            const hasUnsyncedChanges = syncStatus.value.hasUnsyncedChanges;

            if (
                hasUnsyncedChanges &&
                syncEnabledValue &&
                !syncStatus.value.syncing
            ) {
                try {
                    showSyncIndicator("Syncing your data before logout...", 50);
                    await syncToServer();
                    hideSyncIndicator();
                } catch (syncError: any) {
                    reportError({
                        action: "sync error in logout",
                        message:
                            "Sync failed during logout: " + syncError.message,
                        status: getErrorStatus(syncError),
                        userId: parsedUserDetails.value?.userId || "unknown",
                    } as PlatformError);
                    hideSyncIndicator();
                }
            }

            try {
                chats.value = [];
                currentChatId.value = "";
                expanded.value = [];
                isCollapsed.value = true;

                syncStatus.value = {
                    lastSync: null,
                    syncing: false,
                    hasUnsyncedChanges: false,
                    lastError: null,
                    retryCount: 0,
                    maxRetries: 3,
                    showSyncIndicator: false,
                    syncMessage: "",
                    syncProgress: 0,
                };

                parsedUserDetails.value = parsedUserDetailsNullValues;

                let keysToRemove = [
                    "userdetails",
                    "chatDrafts",
                    "pastePreviews",
                    "linkPreviews",
                    "chats",
                    "isCollapsed",
                ];

                linkPreviewCache.value.clear();

                keysToRemove.forEach((key) => {
                    try {
                        localStorage.removeItem(key);
                    } catch (error) {
                        console.error(
                            `Failed to remove ${key} from localStorage:`,
                            error,
                        );
                    }
                });
            } catch (stateError) {
                console.error("Error clearing application state:", stateError);

                throw new Error(
                    "Failed to clear application state during logout",
                );
            }

            toast.success("Logged out successfully", {
                duration: 3000,
                description: "Ready to log back in anytime",
            });
        } catch (error: any) {
            reportError({
                action: "logout",
                message:
                    "Critical error during logout process: " + error.message,
                description:
                    "Some cleanup operations may not have completed. Please refresh the page.",
                status: getErrorStatus(error),
                userId: parsedUserDetails.value?.userId || "unknown",
                severity: "critical",
            } as PlatformError);
        } finally {
            isLoading.value = false;
            if (confirmDialog.value.visible) {
                confirmDialog.value.visible = false;
            }
        }
    };

    // If skipConfirm is true, execute immediately
    if (options.skipConfirm) {
        await executeLogout();
        return;
    }

    // Otherwise, show confirmation dialog
    confirmDialog.value = {
        visible: true,
        title: "Logout Confirmation",
        message:
            "Are you sure you want to logout?" +
            (parsedUserDetails.value?.syncEnabled
                ? ""
                : " Your unsynced data will be permanently lost."),
        type: "warning",
        confirmText: "Logout",
        cancelText: "Cancel",
        onConfirm: executeLogout,
        onCancel: () => {
            confirmDialog.value.visible = false;
        },
    };
}

function updateExpandedArray() {
    try {
        const messagesLength = currentMessages.value?.length || 0;
        expanded.value = new Array(messagesLength).fill(false);
    } catch (error) {
        console.error("Error updating expanded array:", error);
        expanded.value = [];
    }
}

function scrollToBottom(behavior: ScrollBehavior = "smooth") {
    if (!scrollableElem.value) return;

    try {
        nextTick(() => {
            setTimeout(() => {
                if (!scrollableElem.value) return;

                const container = scrollableElem.value;
                const scrollHeight = container.scrollHeight;
                const clientHeight = container.clientHeight;

                if (scrollHeight > clientHeight) {
                    container.scrollTo({
                        top: scrollHeight,
                        behavior,
                    });
                }

                setTimeout(() => {
                    handleScroll();
                }, 150);
            }, 50);
        });
    } catch (error) {
        console.error("Error scrolling to bottom:", error);
    }
}

function scrollToLastMessage() {
    if (!scrollableElem.value) return;

    nextTick(() => {
        const messages =
            scrollableElem.value?.querySelectorAll(".chat-message");
        if (messages && messages.length > 0) {
            const lastMessage = messages[messages.length - 2] as HTMLElement;
            if (lastMessage) {
                const offsetTop = lastMessage.offsetTop - 10;
                scrollableElem.value?.scrollTo({
                    top: offsetTop,
                    behavior: "smooth",
                });
            }
        }
    });
}

function handleSrollIntoView(id: string) {
    try {
        const element = document.getElementById(id);
        if (element) {
            element.scrollIntoView({
                behavior: "smooth",
                block: "start",
                inline: "nearest",
            });
        } else {
            console.warn(`Element with id "${id}" not found for scroll.`);
        }
    } catch (error) {
        console.error("Error scrolling to element:", error);
    }
}

function handleScroll() {
    try {
        if (isOpenTextHighlightPopover.value) {
            isOpenTextHighlightPopover.value = false;
        }

        const elem = scrollableElem.value;
        if (!elem) return;

        const scrollTop = elem.scrollTop;
        const scrollHeight = elem.scrollHeight;
        const clientHeight = elem.clientHeight;

        const currentScrollPosition = scrollTop + clientHeight;
        const totalScrollableHeight = scrollHeight;

        const threshold = 148;
        const isAtBottom =
            Math.abs(currentScrollPosition - totalScrollableHeight) <=
            threshold;

        const hasSubstantialContent = scrollHeight > currentScrollPosition;

        showScrollDownButton.value = !isAtBottom && hasSubstantialContent;
    } catch (error) {
        console.error("Error handling scroll:", error);
    }
}

function hideSidebar() {
    try {
        isSidebarHidden.value = !isSidebarHidden.value;
    } catch (error) {
        console.error("Error toggling sidebar:", error);
    }
}

async function deleteMessage(messageIndex: number) {
    if (isLoading.value || !currentChat.value) return;

    try {
        if (
            messageIndex < 0 ||
            messageIndex >= currentChat.value.messages.length
        ) {
            toast.error("Invalid message");
            return;
        }

        const message = currentChat.value.messages[messageIndex];
        const messageContent =
            message?.prompt || message?.response || "this message";
        const preview =
            messageContent.slice(0, 50) +
            (messageContent.length > 50 ? "..." : "");

        confirmDialog.value = {
            visible: true,
            title: "Delete Message",
            message: `Are you sure you want to delete this message?\n"${preview}"\n\nThis action cannot be undone.`,
            type: "danger",
            confirmText: "Delete",
            cancelText: "Cancel",
            onConfirm: async () => {
                try {
                    if (
                        !currentChat.value ||
                        messageIndex >= currentChat.value.messages.length
                    ) {
                        toast.error("Message no longer exists");
                        return;
                    }

                    const chat = currentChat.value;
                    const messageToDelete = chat.messages[messageIndex];
                    let deleteSuccess = true;

                    if (isAuthenticated.value && messageToDelete.id) {
                        try {
                            await apiCallFromChat(
                                `/messages/${messageToDelete.id}`,
                                {
                                    method: "DELETE",
                                },
                            );
                        } catch (apiError) {
                            console.error(
                                "Failed to delete message from server:",
                                apiError,
                            );
                            toast.warning(
                                "Failed to delete message, try again later",
                            );
                            deleteSuccess = false;
                        }
                    }

                    if (deleteSuccess) {
                        const promptUrls = extractUrls(
                            messageToDelete.prompt || "",
                        );
                        const responseUrls = extractUrls(
                            messageToDelete.response || "",
                        );
                        const urls = [
                            ...new Set([...promptUrls, ...responseUrls]),
                        ];

                        chat.messages.splice(messageIndex, 1);
                        expanded.value.splice(messageIndex, 1);

                        chat.updated_at = new Date().toISOString();
                        chat.message_count = chat.messages.length;

                        if (messageIndex === 0 && chat.messages.length > 0) {
                            const firstMessage = chat.messages[0];
                            const firstContent =
                                firstMessage.prompt || firstMessage.response;
                            if (firstContent) {
                                chat.title = generateChatTitle(firstContent);
                            }
                        } else if (chat.messages.length === 0) {
                            chat.title = "New Chat";
                        }

                        if (urls.length > 0) {
                            urls.forEach((url) => {
                                linkPreviewCache.value.delete(url);
                            });
                            saveLinkPreviewCache();
                        }

                        saveChats();

                        if (
                            isAuthenticated.value &&
                            parsedUserDetails.value?.syncEnabled
                        ) {
                            setTimeout(() => {
                                performSmartSync();
                            }, 1000);
                        }

                        toast.success("Message deleted");
                    }
                } catch (error: any) {
                    reportError({
                        action: "deleteMessage",
                        message: "Error in deleteMessage: " + error.message,
                        description: "Failed to delete this message.",
                        status: getErrorStatus(error),
                        userId: parsedUserDetails.value?.userId || "unknown",
                        severity: "critical",
                    } as PlatformError);
                } finally {
                    confirmDialog.value.visible = false;
                }
            },
            onCancel: () => {
                confirmDialog.value.visible = false;
            },
        };
    } catch (error: any) {
        reportError({
            action: "deleteMessage",
            message: "Error in deleteMessage: " + error.message,
            description: "Failed to delete this message.",
            status: getErrorStatus(error),
            userId: parsedUserDetails.value?.userId || "unknown",
            severity: "critical",
        } as PlatformError);
    }
}

async function fetchLinkPreview(
    url: string,
    options: {
        cache?: boolean;
    } = {},
): Promise<LinkPreview> {
    const { cache = true } = options;

    try {
        new URL(url);
    } catch (error) {
        console.error("Invalid URL provided:", url);
        return {
            url,
            title: "Invalid URL",
            domain: "Invalid",
            loading: false,
            error: true,
        };
    }

    if (cache && linkPreviewCache.value.has(url)) {
        return linkPreviewCache.value.get(url)!;
    }

    const preview: LinkPreview = {
        url,
        title: "",
        domain: new URL(url).hostname,
        loading: true,
        error: false,
    };

    if (cache) {
        linkPreviewCache.value.set(url, preview);
    }

    try {
        const lang = "en";
        const proxyUrl = `${SPINDLE_URL}/scrape?url=${encodeURIComponent(url)}&lang=${lang}`;

        const response = await fetch(proxyUrl, {
            signal: AbortSignal.timeout(15000),
        });

        if (!response.ok) {
            throw new Error(`HTTP ${response.status}`);
        }

        const results = await response.json();
        const domain = new URL(url).hostname;

        let videoInfo: any = {};
        try {
            videoInfo = await detectAndProcessVideo(url, results);
        } catch (videoError) {
            console.error("Error processing video:", videoError);
        }

        const updatedPreview: LinkPreview = {
            url,
            title: results.title?.slice(0, 100) || domain,
            description: results.description?.slice(0, 200) || "",
            images: Array.isArray(results.images) ? results.images : [],
            previewImage:
                videoInfo.thumbnail ||
                results.preview_image ||
                results.images?.[0] ||
                "",
            domain,
            favicon:
                results.favicon ||
                `https://www.google.com/s2/favicons?domain=${domain}`,
            links: Array.isArray(results.links) ? results.links : [],
            video: videoInfo.videoUrl || results.video || "",
            videoType: videoInfo.type,
            videoDuration: videoInfo.duration,
            videoThumbnail: videoInfo.thumbnail,
            embedHtml: videoInfo.embedHtml,
            loading: false,
            error: false,
        };

        if (cache) {
            linkPreviewCache.value.set(url, updatedPreview);
            saveLinkPreviewCache();
        }
        return updatedPreview;
    } catch (error: any) {
        reportError({
            action: "fetchLinkPreview",
            message: "Error in fetching link preview: " + error.message,
            description: `Failed to fetch link preview for ${url}`,
            status: getErrorStatus(error),
            userId: parsedUserDetails.value?.userId || "unknown",
            severity: "low",
        } as PlatformError);

        const fallbackPreview: LinkPreview = {
            url,
            title: new URL(url).hostname,
            domain: new URL(url).hostname,
            loading: false,
            error: true,
        };

        if (cache) {
            linkPreviewCache.value.set(url, fallbackPreview);
            saveLinkPreviewCache();
        }
        return fallbackPreview;
    }
}

function handleClickOutside() {
    try {
        activeChatMenu.value = null;
        showProfileMenu.value = false;
    } catch (error) {
        console.error("Error handling click outside:", error);
    }
}

async function syncFromServer() {
    if (!parsedUserDetails.value?.userId) {
        console.log("❌ syncFromServer: No user ID");
        return;
    }

    const shouldSync = syncEnabled.value;
    if (!shouldSync) {
        console.log("❌ syncFromServer: Sync disabled");
        await loadChats(); // ✅ Load local chat if sync disabled
        return;
    }

    try {
        syncStatus.value.syncing = true;
        syncStatus.value.lastError = null;
        isLoading.value = true;
        showSyncIndicator("Syncing data from server...", 30);

        let data: any;
        if (isOnline.value && navigator.onLine) {
            try {
                console.log("📡 Fetching data from server...");
                updateSyncProgress("Fetching data from server...", 50);
                const response = await apiCall("/sync", {
                    method: "GET",
                });
                data = response.data;
                isLoading.value =
                    data.chats && data.chats !== "[]" ? true : false;
            } catch (error: any) {
                console.warn(
                    "⚠️ No data received from server: " + error.message,
                );
            }
        }

        if (!data || !data.chats) {
            //load local chats
            await loadChats();
            isLoading.value = false;
            return;
        }

        console.log("📥 Server data received");

        updateSyncProgress("Processing chats...", 70);

        if (data.chats && data.chats !== "[]") {
            const serverChatsData =
                typeof data.chats === "string"
                    ? JSON.parse(data.chats)
                    : data.chats;

            if (Array.isArray(serverChatsData)) {
                const localChats = chats.value;
                const mergedChats = mergeChats(serverChatsData, localChats);

                chats.value = mergedChats;
                saveChats();
                console.log(
                    `✅ Synced ${mergedChats.length} chats from server`,
                );
            }
            isLoading.value = false;
        } else {
            console.log("📭 No chats data from server");
            isLoading.value = false;
        }

        updateSyncProgress("Processing link previews...", 85);

        if (data.link_previews && data.link_previews !== "{}") {
            try {
                const serverPreviewsData =
                    typeof data.link_previews === "string"
                        ? JSON.parse(data.link_previews)
                        : data.link_previews;

                if (
                    typeof serverPreviewsData === "object" &&
                    serverPreviewsData !== null
                ) {
                    const localPreviews = Object.fromEntries(
                        linkPreviewCache.value,
                    );
                    const mergedPreviews = {
                        ...localPreviews,
                        ...serverPreviewsData,
                    };

                    linkPreviewCache.value = new Map(
                        Object.entries(mergedPreviews),
                    );
                    localStorage.setItem(
                        "linkPreviews",
                        JSON.stringify(mergedPreviews),
                    );
                    console.log(
                        `✅ Synced ${Object.keys(mergedPreviews).length} link previews from server`,
                    );
                }
            } catch (parseError: any) {
                reportError({
                    action: "syncFromServer",
                    message:
                        "Error parsing server link previews: " +
                        parseError.message,
                    status: getErrorStatus(parseError),
                    context: createErrorContext({
                        dataType: "link_previews",
                        errorName: parseError.name,
                    }),
                    userId: parsedUserDetails.value?.userId || "unknown",
                    severity: "low",
                } as PlatformError);
            }
        }

        updateSyncProgress("Updating preferences...", 95);

        if (
            data.sync_enabled !== undefined ||
            data.preferences ||
            data.theme ||
            data.work_function ||
            data.phone_number ||
            data.plan
        ) {
            const updatedUserDetails: UserDetails = {
                ...parsedUserDetails.value,
                preferences:
                    data.preferences || parsedUserDetails.value.preferences,
                theme: data.theme || parsedUserDetails?.value?.theme,
                workFunction:
                    data.work_function || parsedUserDetails.value.workFunction,
                phoneNumber:
                    data.phone_number || parsedUserDetails.value.phoneNumber,
                plan: data.plan || parsedUserDetails.value.plan,
                planName: data.plan_name || parsedUserDetails.value.planName,
                amount: data.amount ?? parsedUserDetails.value.amount,
                duration: data.duration || parsedUserDetails.value.duration,
                price: data.price || parsedUserDetails.value.price,
                responseMode:
                    data.response_mode || parsedUserDetails.value.responseMode,
                requestCount:
                    data.request_count || parsedUserDetails.value.requestCount,
                expiryTimestamp:
                    data.expiry_timestamp ||
                    parsedUserDetails.value.expiryTimestamp,
                expireDuration:
                    data.expire_duration ||
                    parsedUserDetails.value.expireDuration,
                syncEnabled:
                    data.sync_enabled ?? parsedUserDetails.value.syncEnabled,
            };

            parsedUserDetails.value = updatedUserDetails;
            localStorage.setItem(
                "userdetails",
                JSON.stringify(updatedUserDetails),
            );

            parsedUserDetails.value.userTransactions =
                data.user_transactions || [];
            console.log("✅ User details updated from server");
        }

        syncStatus.value.lastSync = new Date();
        syncStatus.value.hasUnsyncedChanges = false;
        syncStatus.value.retryCount = 0;
        updateExpandedArray();

        updateSyncProgress("Sync complete!", 100);
        setTimeout(() => {
            hideSyncIndicator();
        }, 1000);

        console.log("✅ Successfully synced data from server");
    } catch (error: any) {
        isLoading.value = false;
        reportError({
            action: "syncFromServer",
            message: "Sync from server failed: " + error.message,
            description: "Using local stored data instead.",
            status: getErrorStatus(error),
            userId: parsedUserDetails.value?.userId || "unknown",
            context: createErrorContext({
                syncEnabled: parsedUserDetails.value?.syncEnabled,
                errorName: error.name,
                errorStack: error.stack,
            }),
            severity: "low",
        } as PlatformError);
        syncStatus.value.lastError = error.message;
        hideSyncIndicator();

        if (
            error.message.includes("HTTP 404") ||
            error.message.includes("HTTP 401") ||
            error.message.includes("HTTP 403") ||
            error.message.includes("HTTP 409")
        ) {
            // Skip confirmation dialog and logout immediately
            await logout({ skipConfirm: true });
        }

        if (
            !error.message.includes("NetworkError") &&
            !error.message.includes("TypeError")
        ) {
            toast.warning("Failed to sync data from server", {
                duration: 3000,
                description: "Using local data instead",
            });
        }

        throw error;
    } finally {
        syncStatus.value.syncing = false;
    }
}

async function syncToServer() {
    if (
        !parsedUserDetails.value?.userId ||
        parsedUserDetails.value.syncEnabled === false
    ) {
        return;
    }

    try {
        syncStatus.value.syncing = true;
        syncStatus.value.lastError = null;
        showSyncIndicator("Syncing data to server...", 20);

        if (!Array.isArray(chats.value)) {
            throw new Error("Chats data is not valid");
        }

        if (!parsedUserDetails.value.username) {
            throw new Error("User data is incomplete");
        }

        updateSyncProgress("Preparing sync data...", 40);

        const syncData = {
            link_previews: JSON.stringify(
                Object.fromEntries(linkPreviewCache.value),
            ),
            username: parsedUserDetails.value.username,
            preferences: parsedUserDetails.value.preferences || "",
            work_function: parsedUserDetails.value.workFunction || "",
            theme: parsedUserDetails?.value?.theme || "system",
            sync_enabled: parsedUserDetails?.value?.syncEnabled,
            response_mode:
                parsedUserDetails.value.responseMode || "light-response",
            request_count: {
                count: parsedUserDetails.value.requestCount?.count || 0,
                timestamp:
                    parsedUserDetails.value.requestCount?.timestamp ||
                    Date.now(),
            },
        };

        const syncDataSize = JSON.stringify(syncData).length;
        if (syncDataSize > 10000000) {
            throw new Error(
                "Sync data is too large. Please clear some old chats.",
            );
        }

        console.log(
            `Syncing ${chats.value.length} chats to server (${(syncDataSize / 1024).toFixed(1)}KB)`,
        );

        updateSyncProgress("Sending data to server...", 70);

        const response = await apiCall("/sync", {
            method: "POST",
            body: JSON.stringify(syncData),
        });

        syncStatus.value.lastSync = new Date();
        syncStatus.value.hasUnsyncedChanges = false;
        syncStatus.value.retryCount = 0;

        updateSyncProgress("Sync complete!", 100);
        setTimeout(() => {
            hideSyncIndicator();
        }, 1000);

        console.log("✅ Successfully synced data to server");

        return response;
    } catch (error: any) {
        reportError({
            action: "syncToServer",
            message: "Sync to server failed: " + error.message,
            description: "Changes have been reverted to local.",
            status: getErrorStatus(error),
            userId: parsedUserDetails.value?.userId || "unknown",
            severity: "low",
        } as PlatformError);
        syncStatus.value.lastError = error.message;
        syncStatus.value.hasUnsyncedChanges = true;
        hideSyncIndicator();

        syncStatus.value.retryCount = Math.min(
            syncStatus.value.retryCount + 1,
            syncStatus.value.maxRetries,
        );

        if (error.message.includes("too large")) {
            toast.error("Data too large to sync", {
                duration: 5000,
                description: error.message,
            });
        } else if (
            !error.message.includes("NetworkError") &&
            !error.message.includes("TypeError") &&
            !error.message.includes("already in progress") &&
            !error.message.includes("AbortError")
        ) {
            toast.error("Failed to sync data to server", {
                duration: 3000,
                description: "Your data is saved locally",
            });
        }

        throw error;
    } finally {
        syncStatus.value.syncing = false;
    }
}

/**
 * Sync individual chat to server
 * More efficient than syncing all chats
 */
async function syncChatToServer(chatId: string) {
    if (!isAuthenticated.value || !parsedUserDetails.value?.syncEnabled) {
        return;
    }

    try {
        const chat = chats.value.find((c) => c.id === chatId);
        if (!chat) {
            console.error(`Chat ${chatId} not found for sync`);
            return;
        }

        console.log(`🔄 Syncing individual chat: ${chatId}`);

        // Check if chat exists on server, if not create it
        try {
            await apiCall(`/chats/${chatId}`, { method: "GET" });
            // Chat exists, update it
            await apiCall(`/chats/${chatId}`, {
                method: "PUT",
                body: JSON.stringify({
                    title: chat.title,
                    is_archived: chat.is_archived,
                }),
            });
        } catch (error: any) {
            if (
                error.message.includes("not found") ||
                error.message.includes("404")
            ) {
                // Chat doesn't exist, create it
                await apiCall("/chats", {
                    method: "POST",
                    body: JSON.stringify({
                        title: chat.title,
                    }),
                });
            }
        }

        console.log(`✅ Chat ${chatId} synced successfully`);
    } catch (error: any) {
        console.error(`Failed to sync chat ${chatId}:`, error);
        reportError({
            action: "syncChatToServer",
            message: `Failed to sync chat: ${error.message}`,
            status: getErrorStatus(error),
            userId: parsedUserDetails.value?.userId || "unknown",
            severity: "low",
        } as PlatformError);
    }
}

// Syncs only the new message, not the entire chat history
async function onMessageAdded(message: Message) {
    if (!currentChatId.value) return;

    try {
        // Then sync the specific message to server
        if (isAuthenticated.value) {
            console.log(`🔄 Syncing message to chat: ${currentChatId.value}`);

            // First ensure chat exists on server
            await syncChatToServer(currentChatId.value);

            // Then add the message
            const response = await apiCall(
                `/chats/${currentChatId.value}/messages`,
                {
                    method: "POST",
                    body: JSON.stringify({
                        prompt: message.prompt,
                        response: message.response,
                        model: message.model || "gemini",
                        references: message.references || [],
                    }),
                },
            );

            console.log(response);
            if (!response.ok) {
                throw new Error(
                    `Failed to sync message: ${response.statusText}`,
                );
            }

            if (response.success === true) {
                // Save locally only if sync was successful
                saveChats();
                console.log(`✅ Message synced to chat ${currentChatId.value}`);
            }
        }
    } catch (error: any) {
        console.error(
            `Failed to sync message to ${currentChatId.value}:`,
            error,
        );
        reportError({
            action: "syncMessageToServer",
            message: `Failed to sync message: ${error.message}`,
            status: getErrorStatus(error),
            userId: parsedUserDetails.value?.userId || "unknown",
            severity: "low",
        } as PlatformError);
    }
}

/**
 * Called when chat metadata changes (title, etc)
 * Syncs only the specific chat, not all chats
 */
async function onChatUpdated(chatId: string) {
    try {
        // Save locally first
        saveChats();

        // Then sync the specific chat to server
        if (isAuthenticated.value && parsedUserDetails.value?.syncEnabled) {
            await syncChatToServer(chatId);
        }
    } catch (error: any) {
        console.error("Failed to sync chat update:", error);
        syncStatus.value.hasUnsyncedChanges = true;
    }
}

async function performSmartSync() {
    if (syncStatus.value.syncing) {
        console.log("⏳ Sync already in progress, skipping...");
        return;
    }

    console.log("🔄 Performing SmartSync (userDetails only)");

    if (!isAuthenticated.value || !parsedUserDetails.value?.userId) {
        console.log("❌ Sync skipped: not authenticated or no user ID");
        return;
    }

    try {
        const hasPendingLocalChanges = syncStatus.value.hasUnsyncedChanges;

        if (hasPendingLocalChanges) {
            console.log(
                "📤 Has pending local user details changes - syncing TO server",
            );

            const hadUnsyncedChangesBeforeSync =
                syncStatus.value.hasUnsyncedChanges;

            syncStatus.value.hasUnsyncedChanges = false;

            try {
                syncStatus.value.syncing = true;

                const syncData = {
                    username: parsedUserDetails.value.username,
                    preferences: parsedUserDetails.value.preferences || "",
                    work_function: parsedUserDetails.value.workFunction || "",
                    theme: parsedUserDetails?.value?.theme || "system",
                    sync_enabled: parsedUserDetails?.value?.syncEnabled,
                    response_mode:
                        parsedUserDetails.value.responseMode ||
                        "light-response",
                    request_count: {
                        count: parsedUserDetails.value.requestCount?.count || 0,
                        timestamp:
                            parsedUserDetails.value.requestCount?.timestamp ||
                            Date.now(),
                    },
                };

                await apiCall("/sync", {
                    method: "POST",
                    body: JSON.stringify(syncData),
                });

                console.log("✅ Successfully synced user details to server");
            } catch (error: any) {
                reportError({
                    action: "performSmartSync",
                    message: "Failed to sync to server: " + error.message,
                    description: "Using local stored instead.",
                    status: getErrorStatus(error),
                    userId: parsedUserDetails.value?.userId || "unknown",
                    severity: "low",
                } as PlatformError);

                if (hadUnsyncedChangesBeforeSync) {
                    syncStatus.value.hasUnsyncedChanges = true;
                    console.log(
                        "🔄 Marked changes as unsynced due to sync failure",
                    );
                }

                if (
                    error.message?.includes("Network") ||
                    error.message?.includes("timeout")
                ) {
                    console.log("🔄 Network error - will retry sync");
                    setTimeout(() => {
                        performSmartSync().catch(console.error);
                    }, 5000);
                }
            }
        } else {
            console.log("🔍 No pending changes - skipping server updates");
        }
    } catch (error: any) {
        reportError({
            action: "performSmartSync",
            message: "Critical error in Sync: " + error.message,
            description:
                "Changes have been reverted, using local data instead.",
            status: getErrorStatus(error),
            userId: parsedUserDetails.value?.userId || "unknown",
            severity: "critical",
        } as PlatformError);
    } finally {
        console.log("🔓 User details sync completed");
        syncStatus.value.syncing = false;
    }
}

async function toggleSync() {
    const targetSyncValue = !parsedUserDetails.value.syncEnabled;

    const originalSyncValue = parsedUserDetails.value.syncEnabled;
    const originalUserDetails = { ...parsedUserDetails.value };

    try {
        parsedUserDetails.value.syncEnabled = targetSyncValue;
        syncEnabled.value = targetSyncValue;

        console.log("Attempting sync toggle:", {
            current: syncEnabled.value,
            target: targetSyncValue,
        });

        if (targetSyncValue) {
            try {
                showSyncIndicator("Uploading your local data...", 20);

                await syncToServer();

                console.log("Local data uploaded to server successfully");

                updateSyncProgress("Checking for server updates...", 70);
                await syncFromServer();

                hideSyncIndicator();

                toast.success("Sync enabled and data synchronized", {
                    duration: 3000,
                    description: "Your data is now syncing across devices",
                });
            } catch (error: any) {
                reportError({
                    action: "toggleSync",
                    message: "Failed to sync when enabling: " + error.message,
                    status: getErrorStatus(error),
                    userId: parsedUserDetails.value?.userId || "unknown",
                    severity: "low",
                } as PlatformError);
                hideSyncIndicator();

                parsedUserDetails.value.syncEnabled = originalSyncValue;
                syncEnabled.value = originalSyncValue;
                parsedUserDetails.value = originalUserDetails;
                localStorage.setItem(
                    "userdetails",
                    JSON.stringify(originalUserDetails),
                );

                reportError({
                    action: "toggleSync",
                    message: "Failed to enable sync: " + error.message,
                    description:
                        "Could not upload your data. Please try again.",
                    status: error.status,
                    userId: parsedUserDetails.value?.userId || "unknown",
                    severity: "high",
                } as PlatformError);

                throw error;
            }
        } else {
            try {
                showSyncIndicator("Updating sync preference...", 50);

                await apiCall("/sync", {
                    method: "POST",
                    body: JSON.stringify({
                        sync_enabled: false,
                    }),
                });

                hideSyncIndicator();

                localStorage.setItem(
                    "userdetails",
                    JSON.stringify(parsedUserDetails.value),
                );

                toast.info("Sync disabled", {
                    duration: 3000,
                    description:
                        "Your data will only be saved locally on this device",
                });
            } catch (error: any) {
                reportError({
                    action: "toggleSync",
                    message:
                        "Failed to disable sync on server: " + error.message,
                    description:
                        "Changes have been reverted. Please try again.",
                    status: getErrorStatus(error),
                    userId: parsedUserDetails.value?.userId || "unknown",
                    severity: "medium",
                } as PlatformError);
                hideSyncIndicator();

                localStorage.setItem(
                    "userdetails",
                    JSON.stringify(parsedUserDetails.value),
                );

                toast.warning("Sync disabled locally", {
                    duration: 3000,
                    description:
                        "Server update failed, but sync is disabled on this device",
                });
            }
        }
    } catch (error: any) {
        parsedUserDetails.value.syncEnabled = originalSyncValue;
        syncEnabled.value = originalSyncValue;
        parsedUserDetails.value = originalUserDetails;
        localStorage.setItem(
            "userdetails",
            JSON.stringify(originalUserDetails),
        );

        reportError({
            action: "toggleSync",
            message: "Failed to update sync setting: " + error.message,
            description: "Changes have been reverted. Please try again.",
            status: getErrorStatus(error),
            userId: parsedUserDetails.value?.userId || "unknown",
            severity: "high",
        } as PlatformError);
        throw error;
    }
}

function isLocalDataEmpty(): boolean {
    try {
        if (chats.value.length === 0) {
            return true;
        }

        const hasMeaningfulData = chats.value.some((chat) => {
            return (
                (chat && chat.messages && chat.messages?.length > 0) ||
                (chat.title && chat.title !== "New Chat" && chat.title !== "")
            );
        });

        return !hasMeaningfulData;
    } catch (error: any) {
        reportError({
            action: "isLocalDataEmpty",
            message: "Error checking local data state: " + error.message,
            status: getErrorStatus(error),
            userId: parsedUserDetails.value?.userId || "unknown",
            severity: "low",
        } as PlatformError);
        return false;
    }
}

function cancelChatRequests(chatId: string) {
    const requestsToCancel: string[] = [];

    requestChatMap.value.forEach((requestChatId, requestId) => {
        if (requestChatId === chatId) {
            requestsToCancel.push(requestId);
        }
    });

    requestsToCancel.forEach((requestId) => {
        const controller = activeRequests.value.get(requestId);
        if (controller) {
            controller.abort();
            activeRequests.value.delete(requestId);
            requestChatMap.value.delete(requestId);
        }
    });
}

function cancelAllRequests() {
    activeRequests.value.forEach((controller) => {
        controller.abort();
    });
    activeRequests.value.clear();
    requestChatMap.value.clear();
}

const hasActiveRequestsForCurrentChat = computed(() => {
    let hasRequests = false;
    requestChatMap.value.forEach((chatId) => {
        if (chatId === currentChatId.value) {
            hasRequests = true;
        }
    });
    return hasRequests;
});

function setupConnectionListeners() {
    window.addEventListener("online", async () => {
        console.log("Browser reports online, verifying...");
        const [isUserOnline, message] = await checkInternetConnection();
        isOnline.value = isUserOnline;
        connectionStatus.value = message;
        if (isOnline.value) {
            if (parsedUserDetails.value?.syncEnabled) {
                console.log("Connection restored - syncing unsaved changes");
                setTimeout(() => {
                    performSmartSync().catch((error) => {
                        console.error(
                            "Auto-sync after connection recovery failed:",
                            error,
                        );
                    });
                }, 3000);
            }
        }
    });

    window.addEventListener("offline", async () => {
        console.log("Browser reports offline");
        const [isUserOnline, message] = await checkInternetConnection();
        isOnline.value = isUserOnline;
        connectionStatus.value = message;
    });

    let connectionCheckInterval: any;
    const startConnectionMonitoring = () => {
        connectionCheckInterval = setInterval(async () => {
            const [isUserOnline, message] = await checkInternetConnection();
            isOnline.value = isUserOnline;
            connectionStatus.value = message;
        }, 30 * 1000); // every 30sec
    };

    const stopConnectionMonitoring = () => {
        if (connectionCheckInterval) {
            clearInterval(connectionCheckInterval);
        }
    };

    window.addEventListener("online", stopConnectionMonitoring);
    window.addEventListener("offline", startConnectionMonitoring);
}

function loadRequestCount() {
    try {
        if (!parsedUserDetails.value) return;

        if (!parsedUserDetails.value.requestCount) {
            parsedUserDetails.value.requestCount = {
                count: 0,
                timestamp: Date.now(),
            };
            return;
        }

        const data = parsedUserDetails.value.requestCount;

        if (
            typeof data !== "object" ||
            typeof data.timestamp !== "number" ||
            typeof data.count !== "number"
        ) {
            console.warn("Invalid request count data format, resetting");
            parsedUserDetails.value.requestCount = {
                count: 0,
                timestamp: Date.now(),
            };
            return;
        }

        const now = Date.now();
        const timeDiff = now - data.timestamp;
        const twentyFourHours = 24 * 60 * 60 * 1000;

        if (timeDiff > twentyFourHours) {
            parsedUserDetails.value.requestCount = { count: 0, timestamp: now };
        } else {
            parsedUserDetails.value.requestCount.count = Math.max(
                0,
                Math.min(data.count, FREE_REQUEST_LIMIT),
            );
        }
    } catch (error) {
        console.error("Failed to load request count:", error);
        if (parsedUserDetails.value) {
            parsedUserDetails.value.requestCount = {
                count: 0,
                timestamp: Date.now(),
            };
        }
    }
}

function checkRequestLimitBeforeSubmit(): boolean {
    try {
        if (!userHasRequestLimits.value) {
            return true;
        }

        if (!parsedUserDetails.value.requestCount) {
            return true;
        }

        const currentCount = parsedUserDetails.value.requestCount?.count || 0;

        if (currentCount >= FREE_REQUEST_LIMIT) {
            if (userPlanStatus.value.isExpired) {
                toast.error("Your plan has expired", {
                    duration: 5000,
                    description:
                        "Please renew your plan to continue using the service.",
                });
            } else {
                toast.warning("Free requests exhausted", {
                    duration: 4000,
                    description: "Please upgrade to continue chatting.",
                });
            }
            return false;
        }

        return true;
    } catch (error: any) {
        reportError({
            action: "checkRequestLimitBeforeSubmit",
            message: `Error checking request limit for user ${parsedUserDetails.value.username}: ${error.message}`,
            status: getErrorStatus(error),
            userId: parsedUserDetails.value?.userId || "unknown",
            severity: "low",
        } as PlatformError);
        return true;
    }
}

function incrementRequestCount() {
    try {
        if (!userHasRequestLimits.value) {
            return;
        }

        if (!parsedUserDetails.value) {
            return;
        }

        if (!parsedUserDetails.value.requestCount) {
            parsedUserDetails.value.requestCount = {
                count: 0,
                timestamp: Date.now(),
            };
        }

        const currentCount = parsedUserDetails.value.requestCount.count || 0;

        if (currentCount < FREE_REQUEST_LIMIT) {
            parsedUserDetails.value.requestCount.count = currentCount + 1;
        }
    } catch (error: any) {
        reportError({
            action: "incrementRequestCount",
            message: "Failed to increment request count: " + error.message,
            status: getErrorStatus(error),
            userId: parsedUserDetails.value?.userId || "unknown",
            severity: "low",
        } as PlatformError);
    }
}

function resetRequestCount() {
    try {
        if (parsedUserDetails.value) {
            parsedUserDetails.value.requestCount = {
                count: 0,
                timestamp: Date.now(),
            };
        }
    } catch (error: any) {
        reportError({
            action: "resetRequestCount",
            message: "Failed to reset request count: " + error.message,
            status: getErrorStatus(error),
            userId: parsedUserDetails.value?.userId || "unknown",
            severity: "low",
        } as PlatformError);
    }
}

let userDetailsDebounceTimer: any = null;
let chatsDebounceTimer: any = null;
let previousUserDetails: any = null;

let userDetailsSyncTimeout: any = null;
watch(
    () => parsedUserDetails.value,
    (newUserDetails) => {
        if (!isAuthenticated.value || !newUserDetails?.syncEnabled) return;

        const oldUserDetails = previousUserDetails;

        if (!oldUserDetails) {
            previousUserDetails = JSON.parse(JSON.stringify(newUserDetails));
            return;
        }

        if (userDetailsSyncTimeout) {
            clearTimeout(userDetailsSyncTimeout);
        }

        const hasChanges = hasUserDetailsChangedMeaningfully(
            newUserDetails,
            oldUserDetails,
        );

        if (!hasChanges) {
            console.log("No meaningful user details changes detected");
            previousUserDetails = JSON.parse(JSON.stringify(newUserDetails));
            return;
        }

        console.log("User details changed meaningfully");

        userDetailsSyncTimeout = setTimeout(async () => {
            const tempNewState = JSON.parse(JSON.stringify(newUserDetails));

            try {
                console.log("Syncing user details changes to server...");

                localStorage.setItem(
                    "userdetails",
                    JSON.stringify(tempNewState),
                );

                syncStatus.value.hasUnsyncedChanges = true;

                // Attempt sync
                await performSmartSync();

                console.log("User details synced successfully");
                previousUserDetails = tempNewState;
            } catch (error: any) {
                reportError({
                    action: `parsedUserDetails-watcher`,
                    message:
                        "Sync after user details changed failed: " +
                        error.message,
                    status: getErrorStatus(error),
                    userId: parsedUserDetails.value?.userId || "unknown",
                    severity: "low",
                } as PlatformError);

                // On sync failure, we keep the local changes
                previousUserDetails = tempNewState;

                // Don't show error for lock conflicts - these are temporary
                if (!error.message?.includes("already in progress")) {
                    toast.warning("Failed to sync user details", {
                        duration: 4000,
                        description:
                            "Changes saved locally. Will retry automatically.",
                    });
                } else {
                    reportError({
                        action: `parsedUserDetails-watcher`,
                        message: "Sync conflict:" + error.message,
                        description: `Sync conflict - changes saved locally, will retry later`,
                        status: getErrorStatus(error),
                        userId: parsedUserDetails.value?.userId || "unknown",
                        severity: "medium",
                    } as PlatformError);
                }
            }
        }, 1500);
    },
    { deep: true, immediate: false },
);

function hasUserDetailsChangedMeaningfully(
    newDetails: any,
    oldDetails: any,
): boolean {
    if (!oldDetails || !newDetails) return false;

    // Ignore timestamp-only changes
    const keysToCheck = [
        "preferences",
        "theme",
        "workFunction",
        "phoneNumber",
        "syncEnabled",
        "responseMode",
        "requestCount",
    ];

    return keysToCheck.some((key) => {
        const oldValue = oldDetails[key];
        const newValue = newDetails[key];

        // Handle undefined/null comparisons properly
        if (oldValue === undefined || oldValue === null) {
            return newValue !== undefined && newValue !== null;
        }

        const changed = JSON.stringify(newValue) !== JSON.stringify(oldValue);
        return changed;
    });
}

watch(
    () => currentChatId.value,
    async (newChatId, oldChatId) => {
        if (!isAuthenticated.value) {
            return;
        }

        // Only navigate if chat ID actually changed and is valid
        if (newChatId && newChatId !== oldChatId) {
            const currentPath = router.currentRoute.value.path;
            const targetPath = `/chat/${newChatId}`;

            try {
                // First, await the loadChat promise to resolve.
                const [success, message] = await loadChat();

                if (success) {
                    showErrorSection.value = false; // ✅ Ensure error section is hidden
                    updateExpandedArray();
                    nextTick(() => {
                        loadChatDrafts();
                    });
                    console.log(message);
                } else {
                    // Chat genuinely doesn't exist after loading everything
                    console.warn(message);

                    // ✅ Show error section
                    showErrorSection.value = true;

                    // ✅ Find the most recent chat as fallback
                    if (chats.value.length > 0) {
                        const sortedChats = [...chats.value].sort((a, b) => {
                            const dateA = new Date(
                                a.last_message_at ||
                                    a.updated_at ||
                                    a.created_at,
                            ).getTime();
                            const dateB = new Date(
                                b.last_message_at ||
                                    b.updated_at ||
                                    b.created_at,
                            ).getTime();
                            return dateB - dateA;
                        });
                        fallbackChatId.value = sortedChats[0].id;
                        console.log(
                            `⚠️ Fallback chat ID set to: ${fallbackChatId.value}`,
                        );
                    } else {
                        console.warn(`⚠️ No chats available to fallback to`);
                    }
                }
            } catch (error) {
                console.error(
                    `🚨 Error during or after loading chat ${currentChatId.value}:`,
                    error,
                );
                toast.error(`Failed to load chat ${currentChatId.value}`, {
                    duration: 3000,
                });
                // ✅ Show error section
                showErrorSection.value = true;
            }

            // Prevent redundant navigation
            if (currentPath !== targetPath) {
                console.log(`🔄 Navigating to chat: ${newChatId}`);
                router.push(targetPath);
            }
        }
    },
    { immediate: false },
);

function applyTheme(theme: Theme) {
    const prefersDark = window.matchMedia(
        "(prefers-color-scheme: dark)",
    ).matches;

    if (theme === "dark" || (theme === "system" && prefersDark)) {
        isDarkMode.value = true;
        document.documentElement.classList.add("dark");
    } else {
        isDarkMode.value = false;
        document.documentElement.classList.remove("dark");
    }
}

function toggleTheme(newTheme?: Theme) {
    if (newTheme && ["light", "dark", "system"].includes(newTheme)) {
        parsedUserDetails.value.theme = newTheme;
    } else {
        if (parsedUserDetails.value.theme === "system") {
            parsedUserDetails.value.theme = "light";
        } else if (parsedUserDetails.value?.theme === "light") {
            parsedUserDetails.value.theme = "dark";
        } else {
            parsedUserDetails.value.theme = "system";
        }
    }

    // Apply the theme
    applyTheme(parsedUserDetails?.value?.theme || "system");
}

// watch for parsedUserDetails to handle theme changes:
watch(
    () => parsedUserDetails.value?.theme,
    (newTheme) => {
        if (newTheme && ["light", "dark", "system"].includes(newTheme)) {
            applyTheme(newTheme);
        }
    },
    { immediate: true },
);

function toggleSidebar(value?: boolean) {
    isCollapsed.value = value || !isCollapsed.value;
}

// watch for isCollapsed
watch(
    () => isCollapsed.value,
    (newVal, oldVal) => {
        if (newVal !== oldVal) {
            try {
                localStorage.setItem("isCollapsed", String(newVal));
            } catch (error) {
                console.error("Error saving collapsed state:", error);
            }
        }
    },
    { immediate: false }, // Set to false to avoid unnecessary initial save, onMounted
);

// auth state watcher
watch(
    () => isAuthenticated.value,
    async (isAuth, wasAuth) => {
        // Only act on actual auth state changes
        if (isAuth === wasAuth) return;

        const currentRoute = router.currentRoute.value;

        if (!isAuth) {
            // User just logged out
            console.log("❌ User logged out");

            // Only navigate to home if not already there
            if (currentRoute.path !== "/") {
                router.push(`/?from=${currentRoute.path.replace(/^\//, "")}`);
            }
        } else {
            try {
                await syncFromServer();
            } catch (error: unknown) {
                console.error(error);
            }
        }
    },
    { immediate: false },
);

let debouncedResize: any = null;
let systemThemeListener: ((e: MediaQueryListEvent) => void) | null = null;
let darkModeQuery: MediaQueryList | null = null;

onMounted(async () => {
    try {
        console.log("App mounting...");

        // Theme setup
        const savedTheme = parsedUserDetails.value?.theme || "system";

        // Apply theme immediately
        applyTheme(savedTheme);

        // System theme listener
        systemThemeListener = (e: MediaQueryListEvent) => {
            const currentTheme = parsedUserDetails?.value?.theme;
            if (currentTheme === "system" || !currentTheme) {
                isDarkMode.value = e.matches;
                if (e.matches) {
                    document.documentElement.classList.add("dark");
                } else {
                    document.documentElement.classList.remove("dark");
                }
            }
        };

        darkModeQuery = window.matchMedia("(prefers-color-scheme: dark)");
        darkModeQuery.addEventListener("change", systemThemeListener);

        // Collapsed state
        try {
            const storedIsCollapsed = localStorage.getItem("isCollapsed");
            if (storedIsCollapsed !== null) {
                isCollapsed.value = storedIsCollapsed === "true";
            }
        } catch (error) {
            console.error("Error loading collapsed state:", error);
        }

        // Scroll listener
        window.addEventListener("scroll", handleScroll, { passive: true });

        // Resize handler
        screenWidth.value = window.innerWidth; // update screenWidth when app is mounted
        const handleResize = () => {
            screenWidth.value = window.innerWidth; //update screenWidth when user resize screen
            if (screenWidth.value <= 768 && !isCollapsed.value) {
                isCollapsed.value = true;
            }
        };
        debouncedResize = useDebounceFn(handleResize, 200);
        window.addEventListener("resize", debouncedResize);
        handleResize();

        // Authentication and data loading
        if (isAuthenticated.value) {
            console.log("User is authenticated, initializing...");

            try {
                const localExt = localStorage.getItem("external_reference");
                if (localExt) {
                    try {
                        const ext = JSON.parse(localExt);
                        Promise.race([
                            getTransaction(ext),
                            new Promise((_, reject) =>
                                setTimeout(
                                    () => reject(new Error("Timeout")),
                                    5000,
                                ),
                            ),
                        ]).catch((extError) => {
                            console.error(
                                "Error processing external reference:",
                                extError,
                            );
                            localStorage.removeItem("external_reference");
                        });
                    } catch (extError) {
                        console.error(
                            "Error processing external reference:",
                            extError,
                        );
                        localStorage.removeItem("external_reference");
                    }
                }

                if (syncEnabled.value) {
                    await syncFromServer();
                }
            } catch (syncError: any) {
                reportError({
                    action: `App-onmounted`,
                    message: "Error during initial sync: " + syncError.message,
                    description: `Loading local data instead`,
                    status: getErrorStatus(syncError),
                    userId: parsedUserDetails.value?.userId || "unknown",
                    severity: "low",
                } as PlatformError);
            }
        }

        // Connection setup
        checkInternetConnection()
            .then(([isUserOnline, message]) => {
                isOnline.value = isUserOnline;
                connectionStatus.value = message;
                console.log(`Initial connection check: ${message}`);
            })
            .catch((error: any) => {
                isOnline.value = false;
                connectionStatus.value = "offline";
                console.log(
                    "Error checking internet connection: " + error.message,
                );
            });

        setupConnectionListeners();

        if (parsedUserDetails.value) {
            loadRequestCount();
            previousUserDetails = JSON.parse(
                JSON.stringify(parsedUserDetails.value),
            );
        }

        // Visibility change listener
        document.addEventListener("visibilitychange", () => {
            if (!document.hidden && !navigator.onLine) {
                setTimeout(() => {
                    checkInternetConnection();
                }, 1000);
            }
        });

        console.log("App mounted successfully");
    } catch (error: any) {
        reportError({
            action: `App-onmounted`,
            message: "Failed to initialize application: " + error.message,
            description: `Some features may not work correctly`,
            status: getErrorStatus(error),
            userId: parsedUserDetails.value?.userId || "unknown",
            severity: "critical",
        } as PlatformError);
    }
});

onUnmounted(() => {
    // Clean up event listeners
    if (debouncedResize) {
        window.removeEventListener("resize", debouncedResize);
    }

    if (darkModeQuery && systemThemeListener) {
        darkModeQuery.removeEventListener("change", systemThemeListener);
    }

    window.removeEventListener("online", setupConnectionListeners);
    window.removeEventListener("offline", setupConnectionListeners);

    document.removeEventListener("visibilitychange", () => {});

    // Clean up intervals and timeouts
    const intervals = Object.values(globalThis).filter(
        (value) => value && typeof value === "object" && "refresh" in value,
    );
    intervals.forEach((interval) => clearInterval(interval as any));

    if (draftSaveTimeout) {
        clearTimeout(draftSaveTimeout);
    }

    if (userDetailsSyncTimeout) {
        clearTimeout(userDetailsSyncTimeout);
    }

    if (chatsDebounceTimer) {
        clearTimeout(chatsDebounceTimer);
    }

    if (userDetailsDebounceTimer) {
        clearTimeout(userDetailsDebounceTimer);
    }

    window.removeEventListener("scroll", handleScroll);
});

// Global state object with all functions and reactive references
const globalState = {
    // Reactive references
    copiedIndex,
    isChatLoading,
    isOpenTextHighlightPopover,
    syncEnabled,
    isDemoMode,
    FREE_REQUEST_LIMIT,
    requestCount,
    userDetailsDebounceTimer,
    chatsDebounceTimer,
    activeRequests,
    requestChatMap,
    pendingResponses,
    chatDrafts,
    pastePreviews,
    screenWidth,
    confirmDialog,
    isCollapsed,
    isSidebarHidden,
    authData,
    syncStatus,
    isAuthenticated,
    parsedUserDetails,
    currentChatId,
    requestsRemaining,
    shouldShowUpgradePrompt,
    isRequestLimitExceeded,
    chats,
    isLoading,
    expanded,
    isFreeUser,
    scrollableElem,
    showScrollDownButton,
    linkPreviewCache,
    currentChat,
    currentMessages,
    isCurrentChatValid,
    activeChatMenu,
    showProfileMenu,
    planStatus,
    isDarkMode,
    isOnline,
    connectionStatus,
    hasActiveRequestsForCurrentChat,
    shouldHaveLimit,
    showErrorSection,
    fallbackChatId,

    // State for preview sidebar
    showPreviewSidebar,
    previewCode,
    previewLanguage,
    metadata,

    // Function to open preview with code
    openPreview,

    // Function to close preview
    closePreview,

    // Core functions
    updateChat,
    onChatUpdated,
    onMessageAdded,
    copyResponse,
    shareResponse,
    loadChat,
    processLinksInUserPrompt,
    processLinksInResponse,
    cancelAllRequests,
    cancelChatRequests,
    checkInternetConnection,
    apiCall,
    logout,
    updateExpandedArray,
    createNewChat,
    deleteChat,
    renameChat,
    deleteMessage,
    clearAllChats,
    loadChatDrafts,
    clearCurrentDraft,
    saveChatDrafts,
    autoSaveDraft,
    resetRequestCount,
    incrementRequestCount,
    loadRequestCount,
    checkRequestLimitBeforeSubmit,
    handlePaste,
    removePastePreview,

    // UI functions
    toggleTheme,
    scrollToBottom,
    handleScroll,
    hideSidebar,
    toggleSidebar,
    toggleChatMenu,
    handleClickOutside,
    autoGrow,
    performSmartSync,
    scrollToLastMessage,
    handleSrollIntoView,

    // Sync UI functions
    showSyncIndicator,
    hideSyncIndicator,
    updateSyncProgress,

    // Data persistence functions
    saveChats,
    isLocalDataEmpty,
    toggleSync,

    // Link preview functions
    fetchLinkPreview,
    loadLinkPreviewCache,
    saveLinkPreviewCache,

    // Sync functions
    syncFromServer,
    syncToServer,
};

// Provide global state to child components
provide("globalState", globalState);
</script>

<template>
    <div @click="handleClickOutside">
        <Toaster
            position="top-right"
            :closeButton="true"
            closeButtonPosition="top-left"
            :theme="parsedUserDetails ? parsedUserDetails.theme : 'system'"
        />
        <ConfirmDialog
            v-if="confirmDialog.visible"
            :confirmDialog="confirmDialog"
        />
        <UpdateModal />
        <DemoToast />
        <RouterView />
    </div>
</template>
