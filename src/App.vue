<script setup lang="ts">
import {
    computed,
    onMounted,
    onUnmounted,
    provide,
    ref,
    watch,
    type ComputedRef,
} from "vue";
import { toast, Toaster } from "vue-sonner";
import "vue-sonner/style.css";
import type {
    Chat,
    ConfirmDialogOptions,
    CurrentChat,
    LinkPreview,
    UserDetails,
} from "./types";
import {
    generateChatTitle,
    extractUrls,
    validateCredentials,
    getTransaction,
    createErrorContext,
    getErrorStatus,
    SPINDLE_URL,
} from "./utils/globals";
import { nextTick } from "vue";
import { detectAndProcessVideo } from "./utils/videoProcessing";
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

const { reportError } = usePlatformError();
const router = useRouter();
const isUserOnline = ref(navigator.onLine);
const connectionStatus = ref<"online" | "offline" | "checking">("online");
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
const currentChat: ComputedRef<CurrentChat | undefined> = computed(() => {
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

// computed property to determine if user has limits
const userHasRequestLimits = computed(() => {
    if (!parsedUserDetails.value) return true;

    const hasFreePlan =
        !parsedUserDetails.value.plan ||
        parsedUserDetails.value.plan === "free" ||
        parsedUserDetails.value.plan === "" ||
        userPlanStatus.value.status === "no-plan";

    return hasFreePlan || userPlanStatus.value.isExpired;
});

// Consolidated request limit computations
const requestLimitInfo = computed(() => {
    const hasLimits = userHasRequestLimits.value;
    const currentCount = parsedUserDetails.value.requestCount?.count || 0;

    return {
        // Core limits
        hasLimits,
        currentCount,
        limit: FREE_REQUEST_LIMIT,

        // Derived states
        isExceeded: hasLimits && currentCount >= FREE_REQUEST_LIMIT,
        shouldShowUpgradePrompt:
            hasLimits &&
            currentCount >= FREE_REQUEST_LIMIT - 2 &&
            currentCount < FREE_REQUEST_LIMIT,
        remaining: hasLimits
            ? Math.max(0, FREE_REQUEST_LIMIT - currentCount)
            : Infinity,

        // Status messages
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

// composables
const { linkPreviewCache, loadLinkPreviewCache, saveLinkPreviewCache } =
    useCache();

const { syncStatus, showSyncIndicator, hideSyncIndicator, updateSyncProgress } =
    useSync();

const { apiCall, unsecureApiCall, checkInternetConnection } = useApiCall({
    isUserOnline,
    connectionStatus,
    parsedUserDetails,
    syncStatus,
});

const {
    loadChats,
    createNewChat,
    sendMessage,
    deleteChat,
    renameChat,
    manualSync,
    processSyncQueue,
    syncQueue,
    isOnline,
} = useChat({
    chats,
    currentChatId,
    isAuthenticated,
    parsedUserDetails,
    syncStatus,
    isLoading,
    confirmDialog,
});

function showConfirmDialog(options: ConfirmDialogOptions) {
    confirmDialog.value = {
        visible: true,
        title: options.title,
        message: options.message,
        type: options.type || "info",
        confirmText: options.confirmText || "Confirm",
        cancelText: options.cancelText || "Cancel",
        onConfirm: () => {
            try {
                options.onConfirm();
            } catch (error: any) {
                reportError({
                    action: `showConfirmDialog`,
                    message: "Error in confirm callback :" + error.message,
                    description: `An error occurred while processing your request`,
                    status: getErrorStatus(error),
                    context: createErrorContext({
                        errorName: error.name,
                        errorStack: error.stack,
                    }),
                    userId: parsedUserDetails.value?.userId || "unknown",
                    severity: "high",
                } as PlatformError);
            } finally {
                nextTick(() => {
                    confirmDialog.value.visible = false;
                });
            }
        },
        onCancel: () => {
            try {
                options.onCancel?.();
            } catch (error: any) {
                reportError({
                    action: `showConfirmDialog`,
                    message: "Error in cancel callback :" + error.message,
                    status: getErrorStatus(error),
                    context: createErrorContext({
                        errorName: error.name,
                        errorStack: error.stack,
                    }),
                    userId: parsedUserDetails.value?.userId || "unknown",
                    severity: "low",
                } as PlatformError);
            } finally {
                nextTick(() => {
                    confirmDialog.value.visible = false;
                });
            }
        },
    };
}

async function logout() {
    showConfirmDialog({
        visible: true,
        title: "Logout Confirmation",
        message:
            "Are you sure you want to logout?" +
            (parsedUserDetails.value?.syncEnabled
                ? ""
                : " Your unsynced data will permanently lost."),
        type: "warning",
        confirmText: "Logout",
        cancelText: "Cancel",
        onConfirm: async () => {
            try {
                isLoading.value = true;

                const syncEnabled = parsedUserDetails.value?.syncEnabled;
                const hasUnsyncedChanges = syncStatus.value.hasUnsyncedChanges;

                if (
                    hasUnsyncedChanges &&
                    syncEnabled &&
                    !syncStatus.value.syncing
                ) {
                    try {
                        showSyncIndicator(
                            "Syncing your data before logout...",
                            50,
                        );
                        await syncToServer();
                        hideSyncIndicator();
                    } catch (syncError: any) {
                        reportError({
                            action: `sync error in logout`,
                            message:
                                "Sync failed during logout" + syncError.message,
                            status: getErrorStatus(syncError),
                            userId:
                                parsedUserDetails.value?.userId || "unknown",
                        } as PlatformError);
                        hideSyncIndicator();
                    }
                }

                const userBackup = { ...parsedUserDetails.value };
                const hasChats = chats.value.length > 0;

                try {
                    // Clear state
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

                    // ✅ Clear user details to trigger isAuthenticated = false
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

                    // ✅ Don't call router.replace() here
                    // The isAuthenticated watcher will handle navigation
                } catch (stateError) {
                    console.error(
                        "Error clearing application state:",
                        stateError,
                    );
                    try {
                        parsedUserDetails.value = userBackup;
                        if (!syncEnabled) {
                            loadLocalData();
                        }
                    } catch (restoreError) {
                        console.error(
                            "Failed to restore state after logout error:",
                            restoreError,
                        );
                    }

                    throw new Error(
                        "Failed to clear application state during logout",
                    );
                }

                if (syncEnabled) {
                    toast.success("Logged out successfully", {
                        duration: 3000,
                        description: hasUnsyncedChanges
                            ? "Your data has been synced to the cloud"
                            : "Ready to log back in anytime",
                    });
                } else {
                    toast.success("Logged out successfully", {
                        duration: 3000,
                        description: hasChats
                            ? "Your chats are saved locally on this device"
                            : "Ready to start fresh when you return",
                    });
                }
            } catch (error: any) {
                reportError({
                    action: `logout`,
                    message:
                        "Critical error during logout process :" +
                        error.message,
                    description: `Some cleanup operations may not have completed. Please refresh the page.`,
                    status: getErrorStatus(error),
                    userId: parsedUserDetails.value?.userId || "unknown",
                    severity: "critical",
                } as PlatformError);
            } finally {
                isLoading.value = false;
            }
        },
    });
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
            const lastMessage = messages[messages.length - 2] as HTMLElement; // Get user's prompt (second to last)
            if (lastMessage) {
                // Scroll so the last message pair starts at the top with some padding
                const offsetTop = lastMessage.offsetTop - 10; // 10px padding from top
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
        // Close text highlight popover on scroll
        if (isOpenTextHighlightPopover.value) {
            isOpenTextHighlightPopover.value = false;
        }

        // Get the scrollable element
        const elem = scrollableElem.value;
        if (!elem) return;

        // Get scroll properties
        const scrollTop = elem.scrollTop;
        const scrollHeight = elem.scrollHeight;
        const clientHeight = elem.clientHeight;

        // Calculate current scroll position and total scrollable height
        const currentScrollPosition = scrollTop + clientHeight;
        const totalScrollableHeight = scrollHeight;

        // Define a threshold to determine if the user is at the bottom
        const threshold = 148;
        const isAtBottom =
            Math.abs(currentScrollPosition - totalScrollableHeight) <=
            threshold;

        // Determine if there's substantial content to warrant a scroll-down button
        const hasSubstantialContent = scrollHeight > currentScrollPosition;

        // Show the scroll-down button if not at the bottom and content is substantial
        // hasSubstantialContent ensures the scroll-down button is only visible when there's enough content to scroll
        // !isAtBottom makes sure it disappears when the user is already at the bottom
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

//  deleteMessage function
function deleteMessage(messageIndex: number) {
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

        showConfirmDialog({
            visible: true,
            title: "Delete Message",
            message: `Are you sure you want to delete this message?\n"${preview}"\n\nThis action cannot be undone.`,
            type: "danger",
            confirmText: "Delete",
            onConfirm: () => {
                try {
                    if (
                        !currentChat.value ||
                        messageIndex >= currentChat.value.messages.length
                    ) {
                        toast.error("Message no longer exists");
                        return;
                    }

                    const messageToDelete =
                        currentChat.value.messages[messageIndex];
                    const responseUrls = extractUrls(
                        messageToDelete.response || "",
                    );
                    const promptUrls = extractUrls(
                        messageToDelete.prompt || "",
                    );
                    const urls = [...new Set([...responseUrls, ...promptUrls])];

                    currentChat.value.messages.splice(messageIndex, 1);
                    expanded.value.splice(messageIndex, 1);

                    currentChat.value.updatedAt = new Date().toISOString();

                    if (
                        messageIndex === 0 &&
                        currentChat.value.messages.length > 0
                    ) {
                        const firstMessage =
                            currentChat.value.messages[0].prompt ||
                            currentChat.value.messages[0].response;
                        if (firstMessage) {
                            currentChat.value.title =
                                generateChatTitle(firstMessage);
                        }
                    } else if (currentChat.value.messages.length === 0) {
                        currentChat.value.title = "New Chat";
                    }

                    if (urls.length > 0) {
                        urls.forEach((url) => {
                            linkPreviewCache.value.delete(url);
                        });
                        saveLinkPreviewCache();
                    }

                    saveChats();

                    // Trigger sync after deleting message
                    if (
                        isAuthenticated.value &&
                        parsedUserDetails.value?.syncEnabled
                    ) {
                        setTimeout(() => {
                            performSmartSync();
                        }, 1000);
                    }
                } catch (error: any) {
                    reportError({
                        action: `deleteMessage`,
                        message: "Error in deleteMessage: " + error.message,
                        description: `Failed to delete this message.`,
                        status: getErrorStatus(error),
                        userId: parsedUserDetails.value?.userId || "unknown",
                        severity: "critical",
                    } as PlatformError);
                }
            },
        });
    } catch (error: any) {
        reportError({
            action: `deleteMessage`,
            message: "Error in deleteMessage: " + error.message,
            description: `Failed to delete this message.`,
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

    // Only return from cache if caching is enabled
    if (cache && linkPreviewCache.value.has(url)) {
        return linkPreviewCache.value.get(url)!;
    }

    // Create preview object but don't cache it yet if cache is false
    const preview: LinkPreview = {
        url,
        title: "",
        domain: new URL(url).hostname,
        loading: true,
        error: false,
    };

    // Only set in cache if caching is enabled
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

        // Only update cache if caching is enabled
        if (cache) {
            linkPreviewCache.value.set(url, updatedPreview);
            saveLinkPreviewCache();
        }
        return updatedPreview;
    } catch (error: any) {
        reportError({
            action: `fetchLinkPreview`,
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

        // Only update cache if caching is enabled
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

async function syncFromServer(serverData?: any) {
    if (!parsedUserDetails.value?.userId) {
        console.log("❌ syncFromServer: No user ID");
        return;
    }

    const shouldSync =
        parsedUserDetails.value?.syncEnabled !== false || serverData;
    if (!shouldSync) {
        console.log("❌ syncFromServer: Sync disabled");
        return;
    }

    try {
        syncStatus.value.syncing = true;
        syncStatus.value.lastError = null;
        showSyncIndicator("Syncing data from server...", 30);

        let data = serverData;
        if (!data) {
            console.log("📡 Fetching data from server...");
            updateSyncProgress("Fetching data from server...", 50);
            const response = await apiCall("/sync", { method: "GET" });
            data = response.data;
        }

        if (!data) {
            console.warn("⚠️ No data received from server");
            return;
        }

        console.log("📥 Server data received:");

        updateSyncProgress("Processing chats...", 70);

        // Process chats
        if (data.chats && data.chats !== "[]") {
            try {
                const serverChatsData =
                    typeof data.chats === "string"
                        ? JSON.parse(data.chats)
                        : data.chats;

                if (Array.isArray(serverChatsData)) {
                    const localChats = chats.value;
                    const mergedChats = mergeChats(serverChatsData, localChats);

                    chats.value = mergedChats;
                    localStorage.setItem("chats", JSON.stringify(mergedChats));
                    console.log(
                        `✅ Synced ${mergedChats.length} chats from server`,
                    );
                }
            } catch (parseError: any) {
                reportError({
                    action: `syncFromServer`,
                    message:
                        "Error parsing server chats: " + parseError.message,
                    status: getErrorStatus(parseError),
                    userId: parsedUserDetails.value?.userId || "unknown",
                    context: createErrorContext({
                        dataType: "chats",
                        rawDataLength: data.chats?.length,
                        errorName: parseError.name,
                    }),
                    severity: "low",
                } as PlatformError);
            }
        } else {
            console.log("📭 No chats data from server");
        }

        updateSyncProgress("Processing link previews...", 85);

        // Process link previews
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
                    action: `syncFromMessage`,
                    message:
                        "Error parsing server link previews: " +
                        parseError.message,
                    status: getErrorStatus(parseError),
                    context: createErrorContext({
                        dataType: "chats",
                        rawDataLength: data.chats?.length,
                        errorName: parseError.name,
                    }),
                    userId: parsedUserDetails.value?.userId || "unknown",
                    severity: "low",
                } as PlatformError);
            }
        }

        updateSyncProgress("Updating preferences...", 95);

        // Update user details if provided
        if (
            data.syncEnabled !== undefined ||
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
                    data.sync_enabled || parsedUserDetails.value.syncEnabled,
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
        reportError({
            action: `syncFromServer`,
            message: "Sync from server failed: " + error.message,
            description: `Using local stored data instead.`,
            status: getErrorStatus(error),
            userId: parsedUserDetails.value?.userId || "unknown",
            context: createErrorContext({
                hasServerData: !!serverData,
                syncEnabled: parsedUserDetails.value?.syncEnabled,
                errorName: error.name,
                errorStack: error.stack,
            }),
            severity: "low",
        } as PlatformError);
        syncStatus.value.lastError = error.message;
        hideSyncIndicator();

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
            chats: JSON.stringify(chats.value),
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
            action: `syncToServer`,
            message: "Syn to server failed: " + error.message,
            description: `Changes have been reverted to local.`,
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

async function performSmartSync() {
    if (syncStatus.value.syncing) {
        console.log("⏳ Sync already in progress, skipping...");
        return;
    }

    console.log("🔄 Performing SmartSync");

    if (!isAuthenticated.value || !parsedUserDetails.value?.userId) {
        console.log("❌ Sync skipped: not authenticated or no user ID");
        return;
    }

    try {
        const isLocalDataEmpty =
            chats.value.length === 0 ||
            (chats.value.length === 1 && chats.value[0].messages.length === 0);

        // Check if there are pending local changes (like theme changes)
        const hasPendingLocalChanges = syncStatus.value.hasUnsyncedChanges;
        if (isLocalDataEmpty && !hasPendingLocalChanges) {
            // Only sync FROM server if there are no pending local changes
            console.log(
                "📥 Local data empty and no pending changes - syncing FROM server",
            );

            try {
                syncStatus.value.syncing = true;
                await syncFromServer();
                console.log("✅ Successfully synced data from server");
            } catch (error: any) {
                reportError({
                    action: `performSmartSync`,
                    message: "Failed to sync from server: " + error.message,
                    description: `Using local stored data instead.`,
                    status: getErrorStatus(error),
                    userId: parsedUserDetails.value?.userId || "unknown",
                    severity: "low",
                } as PlatformError);
            }
        } else if (hasPendingLocalChanges) {
            // Always sync TO server first if there are local changes (including theme)
            console.log(
                "📤 Has pending local changes (including user details) - syncing TO server",
            );

            const hadUnsyncedChangesBeforeSync =
                syncStatus.value.hasUnsyncedChanges;

            // Clear the flag BEFORE attempting sync
            syncStatus.value.hasUnsyncedChanges = false;

            try {
                syncStatus.value.syncing = true;
                await syncToServer();
                console.log("✅ Successfully synced changes to server");
            } catch (error: any) {
                reportError({
                    action: `performSmartSync`,
                    message: "Failed to sync to server: " + error.message,
                    description: `Using local stored instead.`,
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

                // Auto-retry for network errors
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
            console.log("🔍 No pending changes - checking for server updates");
            try {
                syncStatus.value.syncing = true;
                await syncFromServer();
                console.log("✅ Server data is current");
            } catch (error: any) {
                reportError({
                    action: `performSmartSync`,
                    message:
                        "Failed to check for server updates: " + error.message,
                    description: `Using local data instead.`,
                    status: getErrorStatus(error),
                    userId: parsedUserDetails.value?.userId || "unknown",
                    severity: "low",
                } as PlatformError);
            }
        }
    } catch (error: any) {
        reportError({
            action: `performSmartSync`,
            message: "Critical error in Sync: " + error.message,
            description: `Changes have been reverted, using local data instead.`,
            status: getErrorStatus(error),
            userId: parsedUserDetails.value?.userId || "unknown",
            severity: "critical",
        } as PlatformError);
    } finally {
        console.log("🔓 Sync completed");
        syncStatus.value.syncing = false;
    }
}

async function handleAuth(data: {
    username: string;
    email: string;
    password: string;
    agreeToTerms: boolean;
}) {
    const { username, email, password, agreeToTerms } = data;

    try {
        const validationError = validateCredentials(
            username,
            email,
            password,
            agreeToTerms,
        );
        if (validationError) {
            throw new Error(validationError);
        }

        let response;
        let isLogin = false;

        try {
            console.log("Attempting login...");
            response = await unsecureApiCall("/login", {
                method: "POST",
                body: JSON.stringify({
                    username,
                    email,
                    password,
                    agree_to_terms: agreeToTerms,
                    user_agent: navigator.userAgent,
                }),
            });
            isLogin = true;
        } catch (loginError: any) {
            console.log("Login failed, attempting registration...");

            try {
                response = await unsecureApiCall("/register", {
                    method: "POST",
                    body: JSON.stringify({
                        username,
                        email,
                        password,
                        agree_to_terms: agreeToTerms,
                        user_agent: navigator.userAgent,
                    }),
                });

                toast.success("Account created successfully!", {
                    duration: 3000,
                    description: `Welcome ${response.data.username}!`,
                });
            } catch (registerError: any) {
                if (
                    loginError.message?.includes("Connection") ||
                    loginError.message?.includes("Network")
                ) {
                    throw loginError;
                } else {
                    throw registerError;
                }
            }
        }

        if (!response || !response.data) {
            throw new Error("Invalid response from server");
        }

        if (isLogin) {
            toast.success("Welcome back!", {
                duration: 3000,
                description: `Logged in as ${response.data.username}`,
            });
        }

        const userData: UserDetails = {
            userId: response.data.user_id,
            username: response.data.username,
            email: response.data.email,
            createdAt: response.data.created_at,
            sessionId: btoa(email + ":" + password + ":" + username),
            workFunction: response.data.work_function || "",
            preferences: response.data.preferences || "",
            theme: response.data.theme || "system",
            syncEnabled: response.data.sync_enabled,
            phoneNumber: response.data.phone_number || "",
            plan: response.data.plan || "free",
            planName: response.data.plan_name || "",
            amount: response.data.amount || 0,
            duration: response.data.duration || "",
            price: response.data.price || 0,
            responseMode: response.data.response_mode || "light-response",
            expiryTimestamp: response.data.expiry_timestamp || null,
            expireDuration: response.data.expire_duration || "",
            emailVerified: response.data.email_verified || false,
            emailSubscribed: response.data.email_subscribed || true,
            requestCount: response.data.request_count || {
                count: 0,
                timestamp: Date.now(),
            },
        };

        // ✅ Set user details FIRST (this will trigger isAuthenticated to become true)
        parsedUserDetails.value = userData;
        previousUserDetails = JSON.parse(JSON.stringify(userData));

        console.log(
            `Authentication successful for user: ${userData.username} (sync: ${userData.syncEnabled})`,
        );

        if (userData.syncEnabled) {
            try {
                await performSmartSync();
                console.log("Initial smart sync completed");
                localStorage.setItem("userdetails", JSON.stringify(userData));
                console.log("User details saved locally after successful sync");
            } catch (syncError) {
                console.error("Initial sync failed:", syncError);
                syncStatus.value.hasUnsyncedChanges = false;
                loadLocalData();
                localStorage.setItem("userdetails", JSON.stringify(userData));

                toast.warning("Synced with server but failed to merge data", {
                    duration: 4000,
                    description: "Your local data is preserved",
                });
            }
        } else {
            localStorage.setItem("userdetails", JSON.stringify(userData));
            loadLocalData();
            console.log("Sync disabled, loaded local data only");
        }

        // ✅ Don't call createNewChat() here - let the watcher handle navigation
        // The isAuthenticated watcher will trigger and navigate appropriately

        return response;
    } catch (error: any) {
        console.error("Authentication error:", error);
        throw error;
    }
}

function loadLocalData() {
    try {
        console.log("Loading data from localStorage...");

        const storedChats = localStorage.getItem("chats");
        if (storedChats) {
            try {
                const parsedChats = JSON.parse(storedChats);
                if (Array.isArray(parsedChats)) {
                    const validChats = parsedChats.filter(
                        (chat) =>
                            chat &&
                            typeof chat === "object" &&
                            chat.id &&
                            typeof chat.id === "string" &&
                            Array.isArray(chat.messages),
                    );
                    chats.value = validChats;
                    console.log(
                        `Loaded ${validChats.length} valid chats from localStorage`,
                    );
                } else {
                    console.warn(
                        "Stored chats is not an array, resetting to empty",
                    );
                    chats.value = [];
                }
            } catch (parseError: any) {
                reportError({
                    action: `loadLocalData`,
                    message:
                        "Error parsing local stored chats: " +
                        parseError.message,
                    status: getErrorStatus(parseError),
                    userId: parsedUserDetails.value?.userId || "unknown",
                    severity: "low",
                } as PlatformError);
                chats.value = [];
                localStorage.removeItem("chats");
            }
        }

        // DON'T automatically set currentChatId or navigate here
        // Let the route handler in ChatView manage this
        console.log("Successfully loaded local data");

        // Only initialize other caches
        loadLinkPreviewCache();
        updateExpandedArray();
    } catch (error: any) {
        reportError({
            action: `loadLocalData`,
            message: "Failed to load local data: " + error.message,
            description: `Some data may not be available`,
            status: getErrorStatus(error),
            userId: parsedUserDetails.value?.userId || "unknown",
            severity: "critical",
        } as PlatformError);
    }
}

async function toggleSync() {
    const targetSyncValue = !parsedUserDetails.value.syncEnabled;

    // Store original state for rollback
    const originalSyncValue = parsedUserDetails.value.syncEnabled;
    const originalUserDetails = { ...parsedUserDetails.value };

    try {
        // Update in-memory state
        parsedUserDetails.value.syncEnabled = targetSyncValue;
        syncEnabled.value = targetSyncValue;

        console.log("Attempting sync toggle:", {
            current: syncEnabled.value,
            target: targetSyncValue,
        });

        if (targetSyncValue) {
            // ENABLING sync - sync TO server first to preserve local data
            try {
                showSyncIndicator("Uploading your local data...", 20);

                // Always sync TO server first when enabling sync
                await syncToServer();

                console.log("Local data uploaded to server successfully");

                // Then optionally sync FROM server to get any additional server data
                updateSyncProgress("Checking for server updates...", 70);
                await syncFromServer();

                hideSyncIndicator();

                toast.success("Sync enabled and data synchronized", {
                    duration: 3000,
                    description: "Your data is now syncing across devices",
                });
            } catch (error: any) {
                reportError({
                    action: `toggleSync`,
                    message: "Failed to sync when enabling: " + error.message,
                    status: getErrorStatus(error),
                    userId: parsedUserDetails.value?.userId || "unknown",
                    severity: "low",
                } as PlatformError);
                hideSyncIndicator();

                // Rollback on failure
                parsedUserDetails.value.syncEnabled = originalSyncValue;
                syncEnabled.value = originalSyncValue;
                parsedUserDetails.value = originalUserDetails;
                localStorage.setItem(
                    "userdetails",
                    JSON.stringify(originalUserDetails),
                );

                reportError({
                    action: `toggleSync`,
                    message: "Failed to enable sync: " + error.message,
                    description: `Could not upload your data. Please try again.`,
                    status: error.status,
                    userId: parsedUserDetails.value?.userId || "unknown",
                    severity: "high",
                } as PlatformError);

                throw error;
            }
        } else {
            // DISABLING sync - just update server preference
            try {
                showSyncIndicator("Updating sync preference...", 50);

                // Update server to disable sync (with empty data)
                await apiCall("/sync", {
                    method: "POST",
                    body: JSON.stringify({
                        username: parsedUserDetails.value.username,
                        sync_enabled: false,
                        chats: "[]",
                        link_previews: "{}",
                    }),
                });

                hideSyncIndicator();

                // Save to localStorage after successful server update
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
                    action: `toggleSync`,
                    message:
                        "Failed to disable sync on server: " + error.message,
                    description: `Changes have been reverted. Please try again.`,
                    status: getErrorStatus(error),
                    userId: parsedUserDetails.value?.userId || "unknown",
                    severity: "medium",
                } as PlatformError);
                hideSyncIndicator();

                // Even if server update fails, allow local disable
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
        // Rollback: Revert both in-memory AND localStorage
        parsedUserDetails.value.syncEnabled = originalSyncValue;
        syncEnabled.value = originalSyncValue;
        parsedUserDetails.value = originalUserDetails;
        localStorage.setItem(
            "userdetails",
            JSON.stringify(originalUserDetails),
        );

        reportError({
            action: `toggleSync`,
            message: "Failed to update sync setting: " + error.message,
            description: `Changes have been reverted. Please try again.`,
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
                chat.messages.length > 0 ||
                (chat.title && chat.title !== "New Chat" && chat.title !== "")
            );
        });

        return !hasMeaningfulData;
    } catch (error: any) {
        reportError({
            action: `isLoadDataEmpty`,
            message: "Error checking local data state: " + error.message,
            status: getErrorStatus(error),
            userId: parsedUserDetails.value?.userId || "unknown",
            severity: "low",
        } as PlatformError);
        return false;
    }
}

function showConnectionStatus() {
    if (!isUserOnline.value) {
        toast.error("You are offline", {
            duration: 4000,
            description: "Please check your internet connection",
        });
    } else {
        toast.success("Connection restored", {
            duration: 3000,
            description: "You are back online",
        });
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
        const isActuallyOnline = await checkInternetConnection();
        if (isActuallyOnline) {
            showConnectionStatus();

            // Auto-sync when coming back online with retry logic
            if (
                // syncStatus.value.hasUnsyncedChanges &&
                parsedUserDetails.value?.syncEnabled
            ) {
                console.log("Connection restored - syncing unsaved changes");
                setTimeout(() => {
                    performSmartSync().catch((error) => {
                        console.error(
                            "Auto-sync after connection recovery failed:",
                            error,
                        );
                        // Don't show error toast for auto-sync failures
                    });
                }, 3000);
            }
        }
    });

    window.addEventListener("offline", () => {
        console.log("Browser reports offline");
        isUserOnline.value = false;
        connectionStatus.value = "offline";
        showConnectionStatus();
    });

    let connectionCheckInterval: any;
    const startConnectionMonitoring = () => {
        connectionCheckInterval = setInterval(async () => {
            if (!isUserOnline.value) {
                await checkInternetConnection();
                if (isUserOnline.value) {
                    showConnectionStatus();
                }
            }
        }, 30000);
    };

    const stopConnectionMonitoring = () => {
        if (connectionCheckInterval) {
            clearInterval(connectionCheckInterval);
        }
    };

    if (!isUserOnline.value) {
        startConnectionMonitoring();
    }

    window.addEventListener("online", stopConnectionMonitoring);
    window.addEventListener("offline", startConnectionMonitoring);
}

// ---------- Request Limit Functions ----------
function loadRequestCount() {
    try {
        if (!parsedUserDetails.value) return;

        // Initialize if doesn't exist
        if (!parsedUserDetails.value.requestCount) {
            parsedUserDetails.value.requestCount = {
                count: 0,
                timestamp: Date.now(),
            };
            return;
        }

        const data = parsedUserDetails.value.requestCount;

        // Validate data structure
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

        // Check if 24 hours have passed
        const now = Date.now();
        const timeDiff = now - data.timestamp;
        const twentyFourHours = 24 * 60 * 60 * 1000;

        if (timeDiff > twentyFourHours) {
            // Reset count after 24 hours
            parsedUserDetails.value.requestCount = { count: 0, timestamp: now };
        } else {
            // Ensure count doesn't exceed limit
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
            return true; // Unlimited user - allow
        }

        if (!parsedUserDetails.value.requestCount) {
            return true; // No limit data - allow (safety)
        }

        const currentCount = parsedUserDetails.value.requestCount?.count || 0;

        if (currentCount >= FREE_REQUEST_LIMIT) {
            // Show error toast here (at submit time, not load time)
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
            return false; // Block submission
        }

        return true; // Allow submission
    } catch (error: any) {
        reportError({
            action: `checkRequestLimitBeforeSubmit`,
            message: `Error checking request limit for user ${parsedUserDetails.value.username} : ${error.message}`,
            status: getErrorStatus(error),
            userId: parsedUserDetails.value?.userId || "unknown",
            severity: "low",
        } as PlatformError);
        return true; // Allow on error (safety)
    }
}

function incrementRequestCount() {
    try {
        if (!userHasRequestLimits.value) {
            return; // No limits for unlimited users
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
            action: `incrementRequestCount`,
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
            action: `resetRequestCount`,
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

        // Clear any pending sync immediately
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
            // Store the new state temporarily
            const tempNewState = JSON.parse(JSON.stringify(newUserDetails));

            try {
                console.log("Syncing user details changes to server...");

                // Save to localStorage immediately for good UX
                localStorage.setItem(
                    "userdetails",
                    JSON.stringify(tempNewState),
                );

                // Mark as having unsynced changes
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
    () => chats.value,
    (newChats, oldChats) => {
        if (!isAuthenticated.value || !parsedUserDetails.value?.syncEnabled) {
            console.log("🔕 Sync disabled - skipping chat change detection");
            return;
        }

        if (!oldChats || oldChats.length === 0) {
            console.log("🆕 Initial chats load - no previous state to compare");
            return;
        }

        if (chatsDebounceTimer) {
            clearTimeout(chatsDebounceTimer);
        }

        const hasMeaningfulChanges = hasChatsChangedMeaningfully(
            newChats,
            oldChats,
        );

        if (hasMeaningfulChanges) {
            console.log("💾 Chat changes detected - will sync");
            syncStatus.value.hasUnsyncedChanges = true;

            // Use a reasonable debounce but ensure sync happens
            chatsDebounceTimer = setTimeout(() => {
                if (
                    syncStatus.value.hasUnsyncedChanges &&
                    !syncStatus.value.syncing
                ) {
                    console.log("🔄 Triggering sync from chat changes");
                    performSmartSync().catch((error) => {
                        reportError({
                            action: `chats-watcher`,
                            message:
                                "Sync from chat changed failed: " +
                                error.message,
                            status: getErrorStatus(error),
                            userId:
                                parsedUserDetails.value?.userId || "unknown",
                            severity: "low",
                        } as PlatformError);
                    });
                }
            }, 1500); // Increased to 1.5 seconds
        }
    },
    { deep: true, immediate: false },
);

function hasChatsChangedMeaningfully(
    newChats: Chat[],
    oldChats: Chat[],
): boolean {
    if (!oldChats || !Array.isArray(oldChats)) return true;

    if (newChats.length !== oldChats.length) return true;

    for (let i = 0; i < newChats.length; i++) {
        const newChat = newChats[i];
        const oldChat = oldChats[i];

        if (!oldChat) return true;

        if (newChat.messages.length !== oldChat.messages.length) return true;

        for (let j = 0; j < newChat.messages.length; j++) {
            const newMessage = newChat.messages[j];
            const oldMessage = oldChat.messages[j];

            if (!oldMessage) return true;

            if (
                newMessage.prompt !== oldMessage.prompt ||
                newMessage.response !== oldMessage.response
            ) {
                return true;
            }
        }
    }

    return false;
}

watch(
    () => currentChatId.value,
    (newChatId, oldChatId) => {
        if (!isAuthenticated.value) {
            return;
        }

        // Only navigate if chat ID actually changed and is valid
        if (newChatId && newChatId !== oldChatId) {
            const currentPath = router.currentRoute.value.path;
            const targetPath = `/chat/${newChatId}`;

            // Prevent redundant navigation
            if (currentPath !== targetPath) {
                console.log(`🔄 Navigating to chat: ${newChatId}`);
                router.push(targetPath).catch((err) => {
                    // Ignore navigation duplicated errors
                    if (err.name !== "NavigationDuplicated") {
                        console.error("Navigation error:", err);
                    }
                });
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

watch(
    () => isAuthenticated.value,
    async (isAuth, wasAuth) => {
        // Only act on actual auth state changes
        if (isAuth === wasAuth) return;

        const currentRoute = router.currentRoute.value;

        if (isAuth) {
            // User just logged in
            console.log("✅ User authenticated");

            // Don't navigate if already on a valid chat route
            if (currentRoute.path.startsWith("/chat/")) {
                console.log("Already on chat route, staying here");
                return;
            }

            // Navigate to new chat only if on login/home page
            if (currentRoute.path === "/") {
                console.log("Navigating to new chat");
                await router.push("/new");
            }
        } else {
            // User just logged out
            console.log("❌ User logged out");

            // Only navigate to home if not already there
            if (currentRoute.path !== "/") {
                await router.push("/");
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

                if (parsedUserDetails.value?.syncEnabled !== false) {
                    await syncFromServer();
                } else {
                    loadLocalData();
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
                loadLocalData();
            }
        } else {
            loadLocalData();
        }

        // Connection setup
        checkInternetConnection().then((isOnline) => {
            console.log(
                `Initial connection check: ${isOnline ? "Online" : "Offline"}`,
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
            if (!document.hidden && !isUserOnline.value) {
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
    isUserOnline,
    connectionStatus,
    hasActiveRequestsForCurrentChat,
    shouldHaveLimit,

    // Core functions
    copyResponse,
    shareResponse,
    loadChats,
    processLinksInUserPrompt,
    processLinksInResponse,
    cancelAllRequests,
    cancelChatRequests,
    checkInternetConnection,
    showConfirmDialog,
    apiCall,
    logout,
    updateExpandedArray,
    createNewChat,
    switchToChat,
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
    loadLocalData,
    isLocalDataEmpty,
    toggleSync,

    // Link preview functions
    fetchLinkPreview,
    loadLinkPreviewCache,
    saveLinkPreviewCache,

    // Sync functions
    syncFromServer,
    syncToServer,
    manualSync,

    // Authentication
    handleAuth,
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
