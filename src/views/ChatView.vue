<script lang="ts" setup>
import type { Ref } from "vue";
import {
    ref,
    onMounted,
    nextTick,
    computed,
    onBeforeUnmount,
    inject,
    watch,
    onUnmounted,
} from "vue";
import TopNav from "../components/TopNav.vue";
import type {
    Chat,
    ConfirmDialogOptions,
    CurrentChat,
    LinkPreview,
    Res,
    UserDetails,
} from "@/types";
import { toast } from "vue-sonner";
import {
    destroyVideoLazyLoading,
    initializeVideoLazyLoading,
    observeNewVideoContainers,
    pauseVideo,
    playEmbeddedVideo,
    playSocialVideo,
    resumeVideo,
    showVideoControls,
    stopDirectVideo,
    stopVideo,
    toggleDirectVideo,
    updateVideoControls,
} from "@/utils/videoProcessing";
import { onUpdated } from "vue";
import {
    extractUrls,
    generateChatTitle,
    copyCode,
    isPromptTooShort,
    WRAPPER_URL,
    SPINDLE_URL,
} from "@/utils/globals";
import router from "@/router";
import { copyPasteContent } from "@/utils/previewPasteContent";
import PastePreviewModal from "@/components/Modals/PastePreviewModal.vue";
import { useRoute } from "vue-router";
import TextHightlightPopover from "@/components/TextHightlightPopover.vue";
import {
    Brain,
    ClipboardList,
    Library,
    Search,
    Trash,
    Mic,
    MicOff,
    Pause,
    ArrowUp,
    RotateCw,
    Ban,
    X,
    CircleAlert,
    TriangleAlert,
    ArrowDown,
    Share,
    Code,
    Pencil,
    BookText,
    HeartPulse,
    Globe,
    LoaderCircle,
    Check,
} from "lucide-vue-next";
import {
    Pagination,
    PaginationContent,
    PaginationEllipsis,
    PaginationItem,
    PaginationNext,
    PaginationPrevious,
} from "@/components/ui/pagination";
import MarkdownRenderer from "@/components/ui/markdown/MarkdownRenderer.vue";
import LinkPreviewComponent from "@/components/LinkPreview.vue";
import EmptyChatView from "./EmptyChatView.vue";
import type { FunctionalComponent } from "vue";
import PastePreview from "@/components/PastePreview.vue";
import { useHandlePaste } from "@/composables/useHandlePaste";
import { useVoiceRecord } from "@/composables/useVoiceRecord";
import { usePagination } from "@/composables/usePagination";
import { useMessage } from "@/composables/useMessage";
import ProtectedPage from "@/layout/ProtectedPage.vue";

type ModeOption = {
    mode: "light-response" | "web-search" | "deep-search";
    label: string;
    description: string;
    icon: FunctionalComponent<any>;
    title: string;
};

// Inject global state
const {
    copiedIndex,
    shouldHaveLimit,
    chatDrafts,
    screenWidth,
    isCollapsed,
    syncStatus,
    isAuthenticated,
    currentChatId,
    pastePreviews,
    chats,
    planStatus,
    userDetailsDebounceTimer,
    chatsDebounceTimer,
    isLoading,
    cancelAllRequests,
    checkRequestLimitBeforeSubmit,
    createNewChat,
    loadChatDrafts,
    saveChatDrafts,
    clearCurrentDraft,
    deleteMessage,
    scrollToLastMessage,
    scrollableElem,
    showScrollDownButton,
    handleScroll,
    scrollToBottom,
    saveChats,
    linkPreviewCache,
    fetchLinkPreview,
    loadLinkPreviewCache,
    saveLinkPreviewCache,
    syncToServer,
    currentMessages,
    updateExpandedArray,
    autoGrow,
    isFreeUser,
    isUserOnline,
    checkInternetConnection,
    activeRequests,
    requestChatMap,
    performSmartSync,
    resetRequestCount,
    incrementRequestCount,
    loadRequestCount,
    FREE_REQUEST_LIMIT,
    requestsRemaining,
    shouldShowUpgradePrompt,
    isRequestLimitExceeded,
    parsedUserDetails,

    copyResponse,
    shareResponse,
    loadChats,
    processLinksInUserPrompt,
    processLinksInResponse,
} = inject("globalState") as {
    copyResponse: (text: string, index?: number) => void;
    shareResponse: (text: string, prompt?: string) => void;
    loadChats: () => void;
    processLinksInUserPrompt: (index: number) => Promise<void>;
    processLinksInResponse: (index: number) => Promise<void>;

    copiedIndex: Ref<number | null>;
    shouldHaveLimit: boolean;
    chatDrafts: Ref<Map<string, string>>;
    userDetailsDebounceTimer: any;
    chatsDebounceTimer: any;
    screenWidth: Ref<number>;
    confirmDialog: Ref<ConfirmDialogOptions>;
    isCollapsed: Ref<boolean>;
    syncStatus: Ref<{
        lastSync: Date | null;
        syncing: boolean;
        hasUnsyncedChanges: boolean;
        showSyncIndicator: boolean;
        syncMessage: string;
        syncProgress: number;
        lastError: string | null;
        retryCount: number;
        maxRetries: number;
    }>;
    isAuthenticated: Ref<boolean>;
    parsedUserDetails: Ref<UserDetails>;
    planStatus: Ref<{
        status: string;
        timeLeft: string;
        expiryDate: string;
        isExpired: boolean;
    }>;
    currentChatId: Ref<string>;
    pastePreviews: Ref<
        Map<
            string,
            {
                content: string;
                wordCount: number;
                charCount: number;
                show: boolean;
            }
        >
    >;
    chats: Ref<Chat[]>;
    logout: () => void;
    isLoading: Ref<boolean>;
    expanded: Ref<boolean[]>;
    scrollToLastMessage: () => void;
    showConfirmDialog: (options: ConfirmDialogOptions) => void;
    clearAllChats: () => void;
    switchToChat: (chatId: string) => boolean;
    createNewChat: (initialMessage?: string) => string;
    deleteChat: (chatId: string) => void;
    loadChatDrafts: () => void;
    saveChatDrafts: () => void;
    renameChat: (chatId: string, newTitle: string) => void;
    deleteMessage: (messageIndex: number) => void;
    scrollableElem: Ref<HTMLElement | null>;
    showScrollDownButton: Ref<boolean>;
    handleScroll: () => void;
    scrollToBottom: () => void;
    cancelAllRequests: () => void;
    cancelChatRequests: (chatId: string) => void;
    saveChats: () => void;
    clearCurrentDraft: () => void;
    linkPreviewCache: Ref<Map<string, LinkPreview>>;
    fetchLinkPreview: (url: string) => Promise<LinkPreview>;
    loadLinkPreviewCache: () => void;
    saveLinkPreviewCache: () => void;
    syncFromServer: (data?: any) => void;
    syncToServer: () => void;
    currentChat: Ref<CurrentChat | undefined>;
    currentMessages: Ref<Res[]>;
    linkPreview: LinkPreview;
    updateExpandedArray: () => void;
    apiCall: (endpoint: string, options: RequestInit) => any;
    manualSync: () => Promise<any>;
    toggleSidebar: () => void;
    isFreeUser: Ref<boolean>;
    FREE_REQUEST_LIMIT: number;
    isDarkMode: Ref<boolean>;
    hasActiveRequestsForCurrentChat: Ref<boolean>;
    isUserOnline: Ref<boolean>;
    connectionStatus: Ref<string>;
    checkInternetConnection: () => Promise<boolean>;
    autoGrow: (e: Event) => void;
    showSyncIndicator: (message: string, progress?: number) => void;
    hideSyncIndicator: () => void;
    updateSyncProgress: (message: string, progress: number) => void;
    performSmartSync: () => Promise<void>;
    activeRequests: Ref<Map<string, AbortController>>;
    requestChatMap: Ref<Map<string, string>>;
    pendingResponses: Ref<Map<string, { prompt: string; chatId: string }>>;
    requestCount: Ref<number>;
    resetRequestCount: () => void;
    incrementRequestCount: () => void;
    loadRequestCount: () => void;
    checkRequestLimitBeforeSubmit: () => boolean;
    requestsRemaining: Ref<boolean>;
    shouldShowUpgradePrompt: Ref<boolean>;
    isRequestLimitExceeded: Ref<boolean>;
};

const route = useRoute();
// ---------- State ----------
const now = ref(Date.now());
const showInputModeDropdown = ref(false);

const isRecording = ref(false);
const isTranscribing = ref(false);
const transcribedText = ref("");
const voiceRecognition = ref<any | null>(null);
const microphonePermission = ref<"granted" | "denied" | "prompt">("prompt");
const transcriptionDuration = ref(0);
let transcriptionTimer: number | null = null;
let updateTimeout: number | null = null;
const {
    stopVoiceRecording,
    clearVoiceTranscription,
    toggleVoiceRecording,
    initializeSpeechRecognition,
} = useVoiceRecord({
    voiceRecognition,
    isRecording,
    isTranscribing,
    transcribedText,
    microphonePermission,
    autoGrow,
    transcriptionDuration,
    updateTimeout,
    transcriptionTimer,
});

const showSuggestionsDropup = ref(false);

const showPasteModal = ref(false);
const pastePreview = computed(() => {
    return pastePreviews.value.get(currentChatId.value) || null;
});
const currentPasteContent = ref<{
    content: string;
    wordCount: number;
    charCount: number;
    type: "text" | "code" | "json" | "markdown" | "xml" | "html";
} | null>(null);
const {
    openPasteModal,
    handlePastePreviewClick,
    closePasteModal,
    cleanupPastePreviewHandlers,
    setupPastePreviewHandlers,
    handleRemovePastePreview,
    removePastePreview,
    handlePaste,
} = useHandlePaste({
    currentChatId,
    pastePreviews,
    chatDrafts,
    saveChatDrafts,
    autoGrow,
    currentPasteContent,
    showPasteModal,
});

const deepSearchPagination = ref<
    Map<string, Map<number, { currentPage: number; totalPages: number }>>
>(new Map());

const isLoadingState = (response: string): boolean => {
    return (
        response.endsWith("...") ||
        response === "..." ||
        response.includes("web-search...") ||
        response.includes("light-search...") ||
        response.includes("deep-search...") ||
        response.includes("light-response...") ||
        response === "refreshing..."
    );
};

const { prevResult, goToPage, nextResult, getPagination } = usePagination({
    currentChatId,
    currentMessages,
    isDeepSearchResult,
    deepSearchPagination,
    scrollToLastMessage,
});

const { getLoadingMessage, formatSearchResults, removeTemporaryMessage } =
    useMessage({
        chats,
        currentChatId,
        deepSearchPagination,
        updateExpandedArray,
        saveChats,
    });

const suggestionPrompts = [
    {
        icon: Pencil,
        title: "Write",
        prompt: "Write a short story about a time traveler who accidentally changes history",
    },
    {
        icon: Code,
        title: "Code",
        prompt: "Help me debug a JavaScript function that's not working as expected",
    },
    {
        icon: BookText,
        title: "Learn",
        prompt: "Explain quantum computing in simple terms",
    },
    {
        icon: HeartPulse,
        title: "Health",
        prompt: "Get me daily healthy routines",
    },
    {
        icon: Globe,
        title: "Events",
        prompt: "What are the latest global events?",
    },
];

// Handle suggestion selection
function selectSuggestion(prompt: string) {
    showSuggestionsDropup.value = false;

    nextTick(() => {
        const textarea = document.getElementById(
            "prompt",
        ) as HTMLTextAreaElement;
        if (textarea) {
            textarea.value = prompt;
            autoGrow({ target: textarea } as any);
            textarea.focus();
        }
    });
}

// Debounced scroll handler to improve performance
let scrollTimeout: any = null;
function debouncedHandleScroll() {
    if (scrollTimeout) {
        clearTimeout(scrollTimeout);
    }

    scrollTimeout = setTimeout(() => {
        handleScroll();
        scrollTimeout = null;
    }, 100); // Increased for better performance
}

// Detect if prompt is just URLs (1 or more) with little/no extra text
function isJustLinks(prompt: string): boolean {
    const trimmed = prompt.trim();
    const urls = extractUrls(trimmed);

    if (urls.length === 0) return false;

    // Remove all URLs from prompt
    let promptWithoutUrls = trimmed;
    for (const url of urls) {
        promptWithoutUrls = promptWithoutUrls.replace(url, "").trim();
    }

    // If only short filler words remain, treat as "just links"
    return promptWithoutUrls.split(/\s+/).filter(Boolean).length <= 3;
}

// handleSubmit function
async function handleSubmit(e?: any, retryPrompt?: string) {
    e?.preventDefault?.();

    // Stop voice recording immediately when submitting
    if (isRecording.value || isTranscribing.value) {
        stopVoiceRecording(true);
    }

    // Use the global connection check
    if (!isUserOnline.value) {
        const isActuallyOnline = await checkInternetConnection();
        if (!isActuallyOnline) {
            toast.error("You are offline", {
                duration: 4000,
                description:
                    "Please check your internet connection and try again",
            });
            return;
        }
    }

    let promptValue = retryPrompt || e?.target?.prompt?.value?.trim();

    // Check if we have paste preview content
    const currentPastePreview = pastePreviews.value.get(currentChatId.value);
    const hasPastePreview =
        currentPastePreview && currentPastePreview.show && !retryPrompt;

    if (hasPastePreview) {
        promptValue += currentPastePreview.content;
        pastePreviews.value.delete(currentChatId.value);
    }

    let fabricatedPrompt = promptValue;
    if (!promptValue || isLoading.value) return;

    if (!isAuthenticated.value) {
        toast.warning("Please create a session first", {
            duration: 4000,
            description: "You need to be logged in.",
        });
        return;
    }

    // Check request limits
    loadRequestCount();

    // Clear draft for current chat
    clearCurrentDraft();

    // Create new chat if none exists
    let submissionChatId = currentChatId.value;
    const submissionChat = chats.value.find(
        (chat) => chat.id === submissionChatId,
    );

    if (!submissionChatId || !submissionChat) {
        const newChatId = createNewChat(promptValue);
        if (!newChatId) return;
        currentChatId.value = newChatId;
        submissionChatId = newChatId;
    }

    // Generate unique request ID
    const requestId = `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    const abortController = new AbortController();

    // Track the active request using global state
    activeRequests.value.set(requestId, abortController);
    requestChatMap.value.set(requestId, submissionChatId);

    // Handle link-only prompts
    if (isJustLinks(promptValue)) {
        return handleLinkOnlyRequest(
            promptValue,
            submissionChatId,
            requestId,
            abortController,
        );
    }

    // Determine response mode - use light-response for pasted content, otherwise user preference
    let responseMode =
        parsedUserDetails?.value.responseMode || "light-response";

    // Override to light-response if pasted content is detected
    if (hasPastePreview) {
        responseMode = "light-response";
        console.log("Pasted content detected - using light-response mode");
    }

    const isSearchMode =
        responseMode === "web-search" || responseMode === "deep-search";

    // Merge with context for short prompts in non-search modes
    if (
        responseMode === "light-response" &&
        isPromptTooShort(promptValue) &&
        currentMessages.value.length > 0
    ) {
        const lastMessage =
            currentMessages.value[currentMessages.value.length - 1];
        fabricatedPrompt = `${lastMessage.prompt || ""} ${lastMessage.response || ""}\nUser: ${promptValue}`;
    }

    // Clear input field
    if (!retryPrompt && e?.target?.prompt) {
        e.target.prompt.value = "";
        e.target.prompt.style.height = "auto";
    }

    // Clear voice transcription
    if (transcribedText.value) {
        transcribedText.value = "";
    }

    isLoading.value = true;
    scrollToLastMessage();

    // Store temporary message reference for potential removal
    let tempMessageIndex = -1;
    const tempResp: Res = {
        prompt: promptValue,
        response: responseMode ? `${responseMode}...` : "...",
    };

    try {
        // Add message to submission chat (temporarily)
        const targetChat = chats.value.find(
            (chat) => chat.id === submissionChatId,
        );
        if (targetChat) {
            targetChat.messages.push(tempResp);
            tempMessageIndex = targetChat.messages.length - 1;
            targetChat.updatedAt = new Date().toISOString();

            // Update chat title if first message
            if (targetChat.messages.length === 1) {
                targetChat.title = generateChatTitle(promptValue);
            }
        }

        updateExpandedArray();
        processLinksInUserPrompt(promptValue);

        let response: Response;
        let parseRes: any;

        if (isSearchMode) {
            // Enhanced search request with proper parameters
            const searchParams = new URLSearchParams({
                query: encodeURIComponent(fabricatedPrompt),
                mode:
                    responseMode === "web-search"
                        ? "light-search"
                        : "deep-search",
                max_results: responseMode === "deep-search" ? "5" : "5",
                content_depth: responseMode === "deep-search" ? "2" : "1",
            });

            console.log(`Making ${responseMode} request`);

            response = await fetch(`${SPINDLE_URL}/search?${searchParams}`, {
                method: "GET",
                signal: abortController.signal,
                headers: {
                    "Content-Type": "application/json",
                },
            });
        } else {
            // Standard light-response mode
            console.log("Making light-response request");

            response = await fetch(WRAPPER_URL, {
                method: "POST",
                body: JSON.stringify(fabricatedPrompt),
                headers: {
                    "Content-Type": "application/json",
                },
                signal: abortController.signal,
            });
        }

        // Check if request was aborted
        if (abortController.signal.aborted) {
            console.log(`Request ${requestId} was aborted`);
            // Remove the temporary message if request was aborted
            removeTemporaryMessage(submissionChatId, tempMessageIndex);
            return;
        }

        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(
                `HTTP ${response.status}: ${errorText || response.statusText}`,
            );
        }

        parseRes = await response.json();

        // Enhanced response processing for search modes
        let finalResponse = parseRes.error ? parseRes.error : parseRes.response;

        if (isSearchMode) {
            const hasResults = parseRes.results || parseRes.json;
            if (hasResults) {
                finalResponse = formatSearchResults(
                    parseRes,
                    responseMode,
                    tempMessageIndex,
                );
            } else {
                finalResponse = "No search results found for your query.";
            }
        }

        // Update the response in chat (replace the temporary message)
        const updatedTargetChat = chats.value.find(
            (chat) => chat.id === submissionChatId,
        );
        if (updatedTargetChat && tempMessageIndex >= 0) {
            const updatedMessage = {
                prompt: promptValue,
                response: finalResponse,
                status: response.status,
            };
            updatedTargetChat.messages[tempMessageIndex] = updatedMessage;
            updatedTargetChat.updatedAt = new Date().toISOString();

            // Process links in the response
            await processLinksInResponse(tempMessageIndex);
        }

        // Increment request count on success
        incrementRequestCount();

        // Show success notification if user switched away
        if (currentChatId.value !== submissionChatId) {
            toast.success("Response received", {
                duration: 3000,
                description: `Switch to "${targetChat?.title || "chat"}" to view the response`,
            });
        }
    } catch (err: any) {
        // Don't show error if request was intentionally aborted
        if (err.name === "AbortError") {
            console.log(`Request ${requestId} was aborted`);
            removeTemporaryMessage(submissionChatId, tempMessageIndex);
            return;
        }

        console.error("AI Response Error:", err);

        // Remove the temporary message on error
        removeTemporaryMessage(submissionChatId, tempMessageIndex);

        // More specific error messages
        let errorMessage = err.message;
        let description = "Please try again or check your connection";

        if (err.message.includes("Failed to fetch")) {
            errorMessage = "Network Error";
            description = "Please check your internet connection";
        } else if (err.message.includes("timeout")) {
            errorMessage = "Request Timeout";
            description = "The request took too long. Please try again";
        }

        toast.error(`Failed to get AI response: ${errorMessage}`, {
            duration: 5000,
            description,
        });

        // Restore draft if request failed
        if (submissionChatId && promptValue.trim()) {
            chatDrafts.value.set(submissionChatId, promptValue);
            saveChatDrafts();
        }
    } finally {
        // Clean up request tracking
        activeRequests.value.delete(requestId);
        requestChatMap.value.delete(requestId);

        isLoading.value = false;
        saveChats();

        // Trigger background sync if needed
        setTimeout(() => {
            performSmartSync().catch((error) => {
                console.error("Background sync failed:", error);
            });
        }, 500);

        // Observe new video containers
        observeNewVideoContainers();
    }
}

// Function to render a single deep search result
function renderDeepSearchResult(data: any, currentPage: number) {
    const { results } = data;
    const result = results[currentPage];

    if (!result) return "No result available";

    const title = result.title || "No Title";
    const url = result.url || "#";
    const markdownContent = result.markdown_content || "";
    const depth = result.depth || 0;
    const source: string = result.source || "Unknown";

    let formatted = `### ${currentPage + 1}. ${title}\n\n`;
    formatted += `**URL:** [${url}](${url}) \n\n`;
    formatted += `> **Source:** ${source.startsWith("https://") ? `[${source}](${source})` : source.length > 60 ? source.slice(0, 60) + "..." : source} • **Depth:** ${depth}  \n`;

    if (markdownContent) {
        formatted += `${markdownContent} \n\n`;
    } else if (result.content) {
        formatted += `${result.content.substring(0, 500)}... \n\n`;
    }

    return formatted;
}

async function handleLinkOnlyRequest(
    promptValue: string,
    chatId: string,
    requestId: string,
    abortController: AbortController,
) {
    const urls = extractUrls(promptValue);

    isLoading.value = true;
    scrollToLastMessage();

    // Store temporary message reference
    let tempMessageIndex = -1;
    const tempResp: Res = { prompt: promptValue, response: "..." };
    const targetChat = chats.value.find((chat) => chat.id === chatId);

    if (targetChat) {
        targetChat.messages.push(tempResp);
        tempMessageIndex = targetChat.messages.length - 1;
        targetChat.updatedAt = new Date().toISOString();
    }

    try {
        let combinedResponse = `I've analyzed the link${urls.length > 1 ? "s" : ""} you shared:  \n\n`;

        for (const url of urls) {
            if (abortController.signal.aborted) {
                console.log(`Link request ${requestId} was aborted`);
                removeTemporaryMessage(chatId, tempMessageIndex);
                return;
            }

            try {
                const linkPreview = await fetchLinkPreview(url);

                // Use proper markdown with double spaces for line breaks
                combinedResponse += `### ${linkPreview.title || "Untitled"}  \n\n`;

                if (linkPreview.description) {
                    combinedResponse += `${linkPreview.description}  \n\n`;
                }

                combinedResponse += `**Source:** ${linkPreview.domain || new URL(url).hostname}  \n`;
                combinedResponse += `**Url:** [${url}](${url})  \n\n`;

                if (urls.length > 1) {
                    combinedResponse += `  \n\n`;
                }
            } catch (err: any) {
                combinedResponse += `### ⚠️ Error  \n\n`;
                combinedResponse += `Failed to analyze: [${url}](${url})  \n\n`;
                combinedResponse += `> ${err.message || "Unknown error occurred"}  \n\n`;

                if (urls.length > 1) {
                    combinedResponse += `  \n\n`;
                }
            }
        }

        // Add summary footer for multiple links
        if (urls.length > 1) {
            combinedResponse += `> ✨ *Analyzed ${urls.length} links* \n`;
        }

        // Update the response in chat
        const updatedTargetChat = chats.value.find(
            (chat) => chat.id === chatId,
        );
        if (updatedTargetChat && tempMessageIndex >= 0) {
            updatedTargetChat.messages[tempMessageIndex] = {
                prompt: promptValue,
                response: combinedResponse.trim(),
                status: 200,
            };
            updatedTargetChat.updatedAt = new Date().toISOString();
        }

        // ✅ ONLY INCREMENT ON SUCCESS for link-only prompts
        incrementRequestCount();

        // Show notification if user switched away
        if (currentChatId.value !== chatId) {
            toast.success("Links analyzed", {
                duration: 3000,
                description: `Switch to "${targetChat?.title || "chat"}" to view the analysis`,
            });
        }
    } catch (err: any) {
        console.error("Link analysis error:", err);

        // Remove temporary message on error
        removeTemporaryMessage(chatId, tempMessageIndex);

        toast.error(`Failed to analyze links: ${err.message}`, {
            duration: 5000,
            description: "Please try again",
        });
    } finally {
        activeRequests.value.delete(requestId);
        requestChatMap.value.delete(requestId);
        isLoading.value = false;
        saveChats();
    }
}

async function refreshResponse(oldPrompt?: string) {
    if (!isUserOnline.value) {
        const isActuallyOnline = await checkInternetConnection();
        if (!isActuallyOnline) {
            toast.error("You are offline", {
                duration: 4000,
                description:
                    "Please check your internet connection and try again",
            });
            return;
        }
    }

    const chatIndex = chats.value.findIndex(
        (chat) => chat.id === currentChatId.value,
    );
    if (chatIndex === -1) return;

    const chat = chats.value[chatIndex];
    const msgIndex = chat.messages.findIndex((m) => m.prompt === oldPrompt);
    if (msgIndex === -1) return;

    const oldMessage = chat.messages[msgIndex];

    // Get the original response mode from message metadata or use current
    const originalMode =
        parsedUserDetails?.value?.responseMode || "light-response";
    const isSearchMode =
        originalMode === "web-search" || originalMode === "deep-search";

    let fabricatedPrompt = oldPrompt;
    if (
        originalMode === "light-response" &&
        oldPrompt &&
        isPromptTooShort(oldPrompt) &&
        currentMessages.value.length > 1
    ) {
        const lastMessage = currentMessages.value[msgIndex - 1];
        fabricatedPrompt = `${lastMessage.prompt || ""} ${lastMessage.response || ""}\nUser: ${oldPrompt}`;
    }

    // Check request limits for refresh too
    if (!checkRequestLimitBeforeSubmit()) {
        return;
    }

    // Show placeholder while refreshing
    chat.messages[msgIndex] = {
        ...oldMessage,
        response: originalMode ? `${originalMode}...` : "Refreshing...",
    };

    isLoading.value = true;
    scrollToLastMessage();

    // Clean up link previews if needed
    const responseUrls = extractUrls(oldMessage.response || "");
    const promptUrls = extractUrls(oldMessage.prompt || "");
    const urls = [...new Set([...responseUrls, ...promptUrls])];

    if (urls.length > 0) {
        urls.forEach((url) => {
            linkPreviewCache.value.delete(url);
        });
        saveLinkPreviewCache();
    }

    // Handle link-only prompts
    if (oldPrompt && isJustLinks(oldPrompt)) {
        const urls = extractUrls(oldPrompt);

        try {
            let combinedResponse = `I've analyzed the link${urls.length > 1 ? "s" : ""} you shared:  \n\n`;

            for (const url of urls) {
                try {
                    const linkPreview = await fetchLinkPreview(url);

                    // Use proper markdown with double spaces for line breaks
                    combinedResponse += `### ${linkPreview.title || "Untitled"}  \n\n`;

                    if (linkPreview.description) {
                        combinedResponse += `${linkPreview.description}  \n\n`;
                    }

                    combinedResponse += `**Source:** ${linkPreview.domain || new URL(url).hostname}  \n`;
                    combinedResponse += `**Url:** [${url}](${url})  \n\n`;

                    if (urls.length > 1) {
                        combinedResponse += ` \n\n`;
                    }
                } catch (err: any) {
                    combinedResponse += `### ⚠️ Error  \n\n`;
                    combinedResponse += `Failed to analyze: [${url}](${url})  \n\n`;
                    combinedResponse += `> ${err.message || "Unknown error occurred"}  \n\n`;

                    if (urls.length > 1) {
                        combinedResponse += ` \n\n`;
                    }
                }
            }

            // Replace the same message with the refreshed response
            chat.messages[msgIndex] = {
                ...oldMessage,
                response: combinedResponse.trim(),
                status: 200,
            };

            chat.updatedAt = new Date().toISOString();
            saveChats();

            // Re-run link previews if needed
            await processLinksInResponse(msgIndex);

            incrementRequestCount();
        } finally {
            observeNewVideoContainers();
        }

        return;
    }

    try {
        let response: Response;
        let parseRes: any;

        if (isSearchMode) {
            // Refresh search request with same parameters
            const searchParams = new URLSearchParams({
                query: encodeURIComponent(fabricatedPrompt || ""),
                mode:
                    originalMode === "web-search"
                        ? "light-search"
                        : "deep-search",
                max_results: originalMode === "deep-search" ? "5" : "5",
                content_depth: originalMode === "deep-search" ? "2" : "1",
            });

            console.log(`Refreshing ${originalMode} request`);

            response = await fetch(`${SPINDLE_URL}/search?${searchParams}`, {
                method: "GET",
                headers: {
                    "Content-Type": "application/json",
                },
            });
        } else {
            // Standard light-response mode refresh
            response = await fetch(WRAPPER_URL, {
                method: "POST",
                body: JSON.stringify(fabricatedPrompt),
                headers: {
                    "Content-Type": "application/json",
                },
            });
        }

        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }

        parseRes = await response.json();

        let finalResponse = parseRes.error ? parseRes.error : parseRes.response;

        if (isSearchMode) {
            // Check for results in both locations (results or json)
            const hasResults = parseRes.results || parseRes.json;
            if (hasResults) {
                finalResponse = formatSearchResults(
                    parseRes,
                    originalMode,
                    msgIndex,
                );
            } else {
                finalResponse = "No search results found for your query.";
            }
        }

        // Replace the same message with the refreshed response
        chat.messages[msgIndex] = {
            ...oldMessage,
            response: finalResponse,
            status: response.status,
        };

        chat.updatedAt = new Date().toISOString();
        saveChats();

        // Re-run link previews if needed
        await processLinksInResponse(msgIndex);

        incrementRequestCount();
    } catch (err: any) {
        console.error("Refresh error:", err);
        toast.error(`Failed to refresh response: ${err.message}`);

        // Restore original message on error
        chat.messages[msgIndex] = oldMessage;
    } finally {
        isLoading.value = false;
        saveChats();
        observeNewVideoContainers();
    }
}

// Helper to check if response is deep search result
function isDeepSearchResult(response: string): boolean {
    if (!response || typeof response !== "string") return false;

    try {
        if (response.startsWith("{") && response.includes('"mode"')) {
            const parsed = JSON.parse(response);
            return parsed.mode === "deep-search";
        }
    } catch (e) {
        return false;
    }

    return false;
}

// input area template logic
const inputPlaceholderText = computed(() => {
    if (pastePreview.value && pastePreview.value.show) {
        return "Large content ready to send...";
    }

    if (isRecording.value) {
        return screenWidth.value > 640
            ? "Speak now... (Click mic to stop)"
            : "Speak now...";
    }

    if (isRequestLimitExceeded.value) {
        if (planStatus.value.isExpired) {
            return screenWidth.value > 640
                ? "Plan expired - renew to continue..."
                : "Plan expired...";
        }
        return screenWidth.value > 640
            ? "Upgrade to continue chatting..."
            : "Upgrade to continue...";
    }

    if (isLoading.value) {
        return "Please wait...";
    }

    if (shouldHaveLimit) {
        return `Ask me a question... (${requestsRemaining.value} requests left)`;
    }

    return "Ask me a question...";
});

const inputDisabled = computed(() => {
    return isLoading.value || isRequestLimitExceeded.value;
});

const showLimitExceededBanner = computed(() => {
    return isRequestLimitExceeded.value;
});

const showUpgradeBanner = computed(() => {
    return shouldShowUpgradePrompt.value && !isRequestLimitExceeded.value;
});

const scrollButtonPosition = computed(() => {
    // Base positions
    const basePosition = "bottom-[130px] sm:bottom-[140px]";
    const withScrollButton = "bottom-[130px] sm:bottom-[140px]";
    const withBanners = "bottom-[210px] sm:bottom-[220px]";
    const withPastePreview = "bottom-[300px] sm:bottom-[350px]";
    const withPasteAndBanners = "bottom-[400px] sm:bottom-[420px]";

    // Priority order: paste + banners > banners > paste > scroll button > base
    if (
        (isRequestLimitExceeded.value || shouldShowUpgradePrompt.value) &&
        pastePreview.value?.show
    ) {
        return withPasteAndBanners;
    } else if (isRequestLimitExceeded.value || shouldShowUpgradePrompt.value) {
        return withBanners;
    } else if (pastePreview.value?.show) {
        return withPastePreview;
    } else if (showScrollDownButton.value) {
        return withScrollButton;
    } else {
        return basePosition;
    }
});

const scrollContainerPadding = computed(() => {
    // When loading (during handleSubmit or refreshResponse), use full viewport padding
    if (isLoading.value) {
        return "pb-[calc(100vh-100px)]";
    }

    // After loading completes, calculate appropriate padding based on UI state
    if (
        (isRequestLimitExceeded.value || shouldShowUpgradePrompt.value) &&
        pastePreview.value?.show
    ) {
        return "pb-[200px] sm:pb-[190px]";
    } else if (isRequestLimitExceeded.value || shouldShowUpgradePrompt.value) {
        return "pb-[190px] sm:pb-[200px]";
    } else if (pastePreview.value?.show) {
        return "pb-[150px] sm:pb-[140px]";
    } else if (showScrollDownButton.value) {
        return "pb-[140px] sm:pb-[120px]";
    } else {
        // Base padding when nothing special is showing
        return "pb-[110px] sm:pb-[120px]";
    }
});

let resizeTimeout: any;
window.onresize = () => {
    clearTimeout(resizeTimeout);
    resizeTimeout = setTimeout(() => {
        screenWidth.value = screen.width;
    }, 100);
};

function onEnter(e: KeyboardEvent) {
    if (e.key !== "Enter" || e.shiftKey || isLoading.value) {
        return;
    }

    e.preventDefault();

    const textarea = e.target as HTMLTextAreaElement;
    if (textarea && textarea.value.trim()) {
        const formEvent = {
            preventDefault: () => {},
            target: { prompt: textarea },
        };
        handleSubmit(formEvent);
    }
}

// Keyboard handler for modal
function handleModalKeydown(e: KeyboardEvent) {
    if (e.key === "Escape" && showPasteModal.value) {
        closePasteModal();
    }
}

// Select input mode and handle special actions
async function selectInputMode(
    mode: "web-search" | "deep-search" | "light-response",
) {
    // Store original value for rollback
    const originalMode = parsedUserDetails.value.responseMode;

    // Don't do anything if same mode
    if (originalMode === mode) {
        showInputModeDropdown.value = false;
        return;
    }

    try {
        // Update in-memory state - the watch will handle syncing
        parsedUserDetails.value.responseMode = mode;
        showInputModeDropdown.value = false;
    } catch (error) {
        console.error("Error selecting input mode:", error);
        parsedUserDetails.value.responseMode = originalMode;

        toast.error("Failed to change mode", {
            duration: 3000,
            description: "An error occurred",
        });
    }
}

// Close dropdown when clicking outside
const handleClickOutside = (e: MouseEvent) => {
    const dropdown = document.querySelector(".relative .absolute");
    const button = document.querySelector(".relative button");

    if (
        dropdown &&
        !dropdown.contains(e.target as Node) &&
        button &&
        !button.contains(e.target as Node)
    ) {
        showInputModeDropdown.value = false;
    }

    // Close suggestions dropup
    const suggestionsDropup = document.querySelector(".suggestions-dropup");
    const suggestionsButton = document.querySelector(".suggestions-button");

    if (
        suggestionsDropup &&
        !suggestionsDropup.contains(e.target as Node) &&
        suggestionsButton &&
        !suggestionsButton.contains(e.target as Node)
    ) {
        showSuggestionsDropup.value = false;
    }
};

const modeOptions: Record<string, ModeOption> = {
    "light-response": {
        mode: "light-response",
        label: "Quick Response",
        description: "Fast & concise",
        icon: Brain,
        title: "Quick Response - Click to change mode",
    },
    "web-search": {
        mode: "web-search",
        label: "Web Search",
        description: "Include web results",
        icon: Search,
        title: "Web Search - Click to change mode",
    },
    "deep-search": {
        mode: "deep-search",
        label: "Deep Search",
        description: "Detailed analysis",
        icon: Library,
        title: "Deep Search - Click to change mode",
    },
};

onUpdated(() => {
    // Check for new video containers after DOM updates
    observeNewVideoContainers();
});

// Watch for chat switches to manage requests
watch(currentChatId, (newChatId, oldChatId) => {
    loadChatDrafts();

    if (oldChatId && newChatId !== oldChatId) {
        // Clear paste preview when switching chats
        // pastePreviews.value.delete(oldChatId)

        // Cancel ongoing requests for the old chat (optional - remove if you want them to continue)
        // cancelChatRequests(oldChatId)

        // User switched chats - stop any active recording
        if (isRecording.value || isTranscribing.value) {
            stopVoiceRecording(true);
            toast.info("Voice recording stopped", {
                duration: 2000,
                description: "Switched to different chat",
            });
        }
    }
});

watch([isRecording, isTranscribing], ([recording, transcribing]) => {
    if (!recording && !transcribing && !transcribedText.value) {
        const textarea = document.getElementById(
            "prompt",
        ) as HTMLTextAreaElement;
        if (textarea && textarea.value && !pastePreview.value) {
            textarea.value = "";
            autoGrow({ target: textarea } as any);
        }
    }
});

// watch for user plan changes
watch(
    () => ({
        isFree: isFreeUser.value,
        planName: parsedUserDetails.value?.planName,
        planStatus: planStatus.value.status,
    }),
    (newValue, oldValue) => {
        if (!oldValue) return; // Skip initial call

        // If user upgraded from free to paid
        if (oldValue.isFree === true && newValue.isFree === false) {
            resetRequestCount();
            toast.success(`Welcome to ${newValue.planName || "Premium"}!`, {
                duration: 5000,
                description: "You now have unlimited requests!",
            });
        }
        // If user downgraded from paid to free (plan expired)
        else if (oldValue.isFree === false && newValue.isFree === true) {
            loadRequestCount();

            if (newValue.planStatus === "expired") {
                toast.warning("Your plan has expired", {
                    duration: Infinity,
                    description: `You're now limited to ${FREE_REQUEST_LIMIT} requests per day`,
                    action: {
                        label: "Upgrade",
                        onClick: () => {
                            router.push("/upgrade");
                        },
                    },
                });
            }
        }
    },
    { deep: true },
);

// Call loadRequestCount after user details are fully loaded
watch(
    () => parsedUserDetails.value,
    (newUserDetails) => {
        if (newUserDetails) {
            // Small delay to ensure all computed properties are updated
            nextTick(() => {
                setTimeout(() => {
                    loadRequestCount();
                }, 100);
            });
        }
    },
    { immediate: true },
);

// planStatus to handle reactive objects properly
watch(
    () => ({ ...planStatus.value }),
    (newStatus, oldStatus) => {
        if (
            oldStatus &&
            oldStatus.isExpired === false &&
            newStatus.isExpired === true
        ) {
            toast.error("Your plan has expired", {
                duration: Infinity,
                description:
                    "Please renew your plan to continue using the service.",
                action: {
                    label: "Renew Now",
                    onClick: () => {
                        router.push("/upgrade");
                    },
                },
            });
        }
    },
    { deep: true },
);

watch(
    [() => currentMessages.value.length, () => chats.value],
    () => {
        nextTick(() => {
            setTimeout(() => {
                handleScroll(); // Recalculate scroll position after content changes
            }, 100);
        });
    },
    { deep: true },
);

watch(
    () => route.params.id,
    (newId, oldId) => {
        const chatId = Array.isArray(newId) ? newId[0] : newId;
        const oldChatId = Array.isArray(oldId) ? oldId[0] : oldId;

        // Skip if no change or if it's the same as current
        if (!chatId || chatId === oldChatId || chatId === currentChatId.value) {
            return;
        }

        console.log(`🔄 Route changed: ${oldChatId} → ${chatId}`);

        // Verify chat exists before syncing
        const chatExists = chats.value.find((chat) => chat.id === chatId);

        if (chatExists) {
            currentChatId.value = chatId;
            updateExpandedArray();
            nextTick(() => {
                loadChatDrafts();
            });
            console.log(`✅ Synced to chat: ${chatId}`);
        } else {
            console.warn(`⚠️ Chat ${chatId} not found in route watch`);
            // Don't create new chat here - let navigation handle it
        }
    },
    { immediate: true },
);

onBeforeUnmount(() => {
    if (transcriptionTimer) clearInterval(transcriptionTimer);
    if (updateTimeout) clearTimeout(updateTimeout);

    // Clean up all active requests
    cancelAllRequests();

    // Clean up speech recognition
    if (isRecording.value || isTranscribing.value) {
        stopVoiceRecording(false); // Don't clear text during unmount
    }

    // Remove keyboard listener
    document.removeEventListener("keydown", handleModalKeydown);
    document.removeEventListener("click", handleClickOutside);

    // Clean up paste preview handlers (use the enhanced cleanup function)
    cleanupPastePreviewHandlers();

    // Restore body scroll if modal is open
    if (showPasteModal.value) {
        document.body.style.overflow = "auto";
    }

    // Clear debounce timers
    if (chatsDebounceTimer) {
        clearTimeout(chatsDebounceTimer);
    }
    if (userDetailsDebounceTimer) {
        clearTimeout(userDetailsDebounceTimer);
    }
});

// Consolidated onMounted hook for better organization
onMounted(async () => {
    const path = router.currentRoute.value.path;
    const routeChatId = route.params.id;
    const chatId = Array.isArray(routeChatId) ? routeChatId[0] : routeChatId;

    // Wait for authentication and chats to load
    await nextTick();

    if (!isAuthenticated.value) {
        router.push("/");
        return;
    }

    // Handle different route scenarios
    if (path === "/new") {
        // Create new chat and let it handle routing
        createNewChat();
        return;
    }

    if (chatId) {
        // Check if chat exists
        const chatExists = chats.value.find((chat) => chat.id === chatId);

        if (chatExists) {
            // Sync with existing chat
            currentChatId.value = chatId;
            updateExpandedArray();
            nextTick(() => {
                loadChatDrafts();
            });
            console.log(`🔄 Synced to existing chat: ${chatId}`);
        } else {
            // Chat doesn't exist, redirect to new chat
            console.warn(`⚠️ Chat ${chatId} not found, creating new`);
            router.push("/new");
        }
    } else if (path === "/") {
        // Root path with no chat - create new one
        createNewChat();
    }

    // 2. Load cached data and setup handlers
    loadLinkPreviewCache();
    setupPastePreviewHandlers();

    // 3. Setup global functions with unified approach
    if (typeof window !== "undefined") {
        const setupGlobalFunction = (name: string, fn: Function) => {
            (window as any)[name] = (...args: any[]) => {
                try {
                    return fn(...args);
                } catch (error) {
                    console.error(`Error in ${name}:`, error);
                    toast.error(`Error in ${name}`, {
                        duration: 3000,
                        description: "An unexpected error occurred",
                    });
                }
            };
        };

        setupGlobalFunction("openPasteModal", openPasteModal);
        setupGlobalFunction("copyPasteContent", copyPasteContent);
        setupGlobalFunction("removePastePreview", removePastePreview);
        setupGlobalFunction("playEmbeddedVideo", playEmbeddedVideo);
        setupGlobalFunction("pauseVideo", pauseVideo);
        setupGlobalFunction("resumeVideo", resumeVideo);
        setupGlobalFunction("stopVideo", stopVideo);
        setupGlobalFunction("toggleDirectVideo", toggleDirectVideo);
        setupGlobalFunction("stopDirectVideo", stopDirectVideo);
        setupGlobalFunction("showVideoControls", showVideoControls);
        setupGlobalFunction("updateVideoControls", updateVideoControls);
        setupGlobalFunction("playSocialVideo", playSocialVideo);
        setupGlobalFunction("scrollToLastMessage", scrollToLastMessage);
        setupGlobalFunction("nextResult", nextResult);
        setupGlobalFunction("prevResult", prevResult);
        setupGlobalFunction("goToPage", goToPage);
    }

    // 4. Setup event listeners once
    document.addEventListener("click", handleClickOutside);

    const copyListener = (e: any) => {
        if (e.target?.classList.contains("copy-button")) {
            const code = decodeURIComponent(e.target.getAttribute("data-code"));
            copyCode(code, e.target);
        }
    };
    document.addEventListener("click", copyListener);
    document.addEventListener("keydown", handleModalKeydown);

    // 5. Initialize core features
    initializeSpeechRecognition();
    initializeVideoLazyLoading();

    // 6. Setup periodic tasks
    const interval = setInterval(() => {
        now.value = Date.now();
    }, 1000);

    const resetCheckInterval = setInterval(loadRequestCount, 5 * 60 * 1000);

    // 7. Initialize authentication-dependent features
    if (isAuthenticated.value) {
        // Load request count for limited users
        const shouldHaveLimit =
            isFreeUser.value ||
            planStatus.value.isExpired ||
            planStatus.value.status === "no-plan" ||
            planStatus.value.status === "expired";

        if (shouldHaveLimit) {
            loadRequestCount();
        }

        loadChats();

        // Initial sync from server
        setTimeout(() => {
            performSmartSync();
        }, 1000);

        // Pre-process links in existing messages - optimized to avoid duplicates
        const processedUrls = new Set();
        currentMessages.value.forEach((item) => {
            [item.prompt, item.response].forEach((text) => {
                if (text && text !== "...") {
                    extractUrls(text)
                        .slice(0, 3)
                        .forEach((url) => {
                            if (
                                !linkPreviewCache.value.has(url) &&
                                !processedUrls.has(url)
                            ) {
                                processedUrls.add(url);
                                fetchLinkPreview(url).then(() => {
                                    linkPreviewCache.value = new Map(
                                        linkPreviewCache.value,
                                    );
                                });
                            }
                        });
                }
            });
        });
    } else {
        loadChats();
    }

    // 8. DOM-dependent functionality - consolidated timing
    nextTick(() => {
        // Clear any initial content in textarea
        const textarea = document.getElementById(
            "prompt",
        ) as HTMLTextAreaElement;
        if (textarea && textarea.value) {
            console.log("Initial textarea content found:", textarea.value);
            textarea.value = "";
            transcribedText.value = "";
        }

        // Setup scroll handling once
        if (scrollableElem.value) {
            // Remove duplicate scroll listener setup
            scrollableElem.value.addEventListener(
                "scroll",
                debouncedHandleScroll,
                { passive: true },
            );
        }

        // Auto-focus input only when appropriate
        if (currentMessages.value.length === 0) {
            textarea?.focus();
        }

        // Process link previews in responses - avoid duplicate processing
        currentMessages.value.forEach((msg: Res, index) => {
            if (
                msg.response &&
                msg.response !== "..." &&
                !msg.response.endsWith("...")
            ) {
                processLinksInResponse(index);
            }
        });

        // Single delayed setup for scroll and video observation
        setTimeout(() => {
            scrollToLastMessage();
            observeNewVideoContainers();

            // Initial scroll state calculation
            handleScroll();
        }, 300); // Single delay instead of multiple
    });

    // 9. Store cleanup references
    onBeforeUnmount(() => {
        // Clean up event listeners
        document.removeEventListener("click", copyListener);
        document.removeEventListener("keydown", handleModalKeydown);
        document.removeEventListener("click", handleClickOutside);
        document.removeEventListener("click", handlePastePreviewClick);
        document.removeEventListener("click", handleRemovePastePreview);

        // Clean up scroll listener
        if (scrollableElem.value) {
            scrollableElem.value.removeEventListener(
                "scroll",
                debouncedHandleScroll,
            );
        }

        // Clean up video lazy loading
        destroyVideoLazyLoading();

        // Clean up intervals
        clearInterval(interval);
        clearInterval(resetCheckInterval);

        // Clear timeouts
        const timers = [
            scrollTimeout,
            resizeTimeout,
            transcriptionTimer,
            updateTimeout,
        ];
        timers.forEach((timer) => {
            if (timer) {
                clearTimeout(timer);
                clearInterval(timer);
            }
        });

        // Clean up speech recognition
        if (isRecording.value) {
            stopVoiceRecording();
        }

        // Restore body scroll if modal is open
        if (showPasteModal.value) {
            document.body.style.overflow = "auto";
        }

        // Final sync if needed
        if (syncStatus.value.hasUnsyncedChanges) {
            syncToServer();
        }
    });
});

onUnmounted(() => {
    // Final cleanup for voice recording
    if (voiceRecognition.value) {
        voiceRecognition.value.abort();
    }

    if (transcriptionTimer) {
        clearInterval(transcriptionTimer);
    }

    if (updateTimeout) {
        clearTimeout(updateTimeout);
    }
});

onUnmounted(() => {
    if (currentChatId.value) {
        currentChatId.value = "";
    }
});
</script>

<template>
    <ProtectedPage>
        <!-- Main Chat Window -->
        <div
            :class="
                screenWidth > 720
                    ? !isCollapsed
                        ? 'flex-grow flex flex-col items-center justify-center ml-[270px] font-light text-sm transition-all duration-300 ease-in-out bg-inherit'
                        : 'flex-grow flex flex-col items-center justify-center ml-[60px] font-light text-sm transition-all duration-300 ease-in-out bg-inherit'
                    : 'text-sm font-light flex-grow items-center justify-center flex flex-col transition-all duration-300 ease-in-out bg-inherit'
            "
        >
            <div
                :class="
                    screenWidth > 720
                        ? 'h-screen bg-inherit flex flex-col items-center justify-center w-[85%]'
                        : 'bg-inherit h-screen w-full flex flex-col items-center justify-center'
                "
            >
                <TopNav />
                <!-- Empty State -->
                <EmptyChatView
                    v-if="currentMessages.length === 0"
                    :suggestionPrompts="suggestionPrompts"
                    :selectSuggestion="selectSuggestion"
                />

                <!-- Chat Messages Container -->
                <div
                    ref="scrollableElem"
                    v-else
                    class="relative md:max-w-3xl min-h-[calc(100vh-200px)] max-w-[100vw] flex-grow no-scrollbar overflow-y-auto px-2 w-full space-y-3 sm:space-y-4 mt-[55px] pt-8 scroll-container"
                    :class="scrollContainerPadding"
                >
                    <div class="flex flex-col gap-1">
                        <div
                            v-for="(item, i) in currentMessages"
                            :key="`chat-${i}`"
                            class="flex flex-col gap-1"
                        >
                            <!-- User Bubble -->
                            <div class="flex w-full chat-message">
                                <div class="flex flex-col w-full">
                                    <div class="flex flex-col gap-">
                                        <div
                                            v-if="
                                                item &&
                                                item.prompt &&
                                                (item?.prompt
                                                    ?.trim()
                                                    .split(/\s+/).length >
                                                    100 ||
                                                    item?.prompt?.length > 800)
                                            "
                                            class="mb-3"
                                        >
                                            <div class="flex justify-start">
                                                <PastePreview
                                                    :content="
                                                        item?.prompt
                                                            ?.trim()
                                                            ?.split(
                                                                '#pastedText#',
                                                            )[1] || ''
                                                    "
                                                    :char-count="
                                                        item?.prompt
                                                            ?.trim()
                                                            ?.split(
                                                                '#pastedText#',
                                                            )[1]?.length || 0
                                                    "
                                                    :word-count="
                                                        item?.prompt
                                                            ?.trim()
                                                            .split(
                                                                '#pastedText#',
                                                            )[1]
                                                            ?.split(/\s+/)
                                                            ?.length || 0
                                                    "
                                                    :is-clickable="true"
                                                    class="w-[70%] sm:w-[50%] lg:w-[40%] xl:w-[30%]"
                                                />
                                            </div>
                                        </div>
                                    </div>

                                    <!-- User message bubble -->
                                    <div
                                        class="flex mt-[2px] items-start gap-2 font-medium bg-gray-100 dark:bg-gray-800 text-black dark:text-gray-100 px-4 rounded-2xl prose prose-sm dark:prose-invert chat-bubble w-fit max-w-full"
                                    >
                                        <!-- Avatar container -->
                                        <div class="flex-shrink-0 py-3">
                                            <div
                                                class="flex items-center justify-center w-7 h-7 text-gray-100 dark:text-gray-800 bg-gray-700 dark:bg-gray-200 rounded-full text-sm font-semibold"
                                            >
                                                {{
                                                    parsedUserDetails.username
                                                        .toUpperCase()
                                                        .slice(0, 2)
                                                }}
                                            </div>
                                        </div>

                                        <!-- Message content container -->
                                        <div class="flex-1 min-w-0">
                                            <MarkdownRenderer
                                                class="break-words text-base leading-relaxed"
                                                :content="
                                                    item &&
                                                    item?.prompt &&
                                                    item?.prompt?.length > 800
                                                        ? item?.prompt
                                                              ?.trim()
                                                              .split(
                                                                  '#pastedText#',
                                                              )[0]
                                                        : item.prompt || ''
                                                "
                                            />
                                        </div>
                                    </div>
                                </div>
                            </div>

                            <!-- Bot Bubble -->
                            <div
                                class="flex w-full md:max-w-3xl max-w-full relative pb-[20px]"
                            >
                                <div
                                    class="bg-none max-w-full w-full chat-message leading-relaxed text-black dark:text-gray-100 p-1 rounded-2xl prose prose-sm dark:prose-invert"
                                >
                                    <!-- Loading state -->
                                    <div
                                        v-if="isLoadingState(item.response)"
                                        class="flex w-full rounded-lg bg-gray-50 dark:bg-gray-800 p-2 items-center animate-pulse gap-2 text-gray-500 dark:text-gray-400"
                                    >
                                        <LoaderCircle
                                            class="w-4 h-4 animate-spin"
                                        />
                                        <span class="text-sm">{{
                                            getLoadingMessage(item.response)
                                        }}</span>
                                    </div>

                                    <!-- Regular response with enhanced link handling -->
                                    <div v-else>
                                        <!-- Check if it's a deep search result -->
                                        <template
                                            v-if="
                                                isDeepSearchResult(
                                                    item.response,
                                                )
                                            "
                                        >
                                            <MarkdownRenderer
                                                class="break-words overflow-x-hidden"
                                                :content="
                                                    renderDeepSearchResult(
                                                        JSON.parse(
                                                            item.response,
                                                        ),
                                                        getPagination(i)
                                                            .currentPage,
                                                    )
                                                "
                                            />
                                        </template>

                                        <!-- Regular response -->
                                        <template v-else>
                                            <MarkdownRenderer
                                                class="break-words overflow-x-hidden"
                                                :content="item.response || ''"
                                            />
                                        </template>

                                        <!-- Link Previews Section -->
                                        <div
                                            v-if="
                                                !isDeepSearchResult(
                                                    item.response,
                                                ) &&
                                                extractUrls(item.response || '')
                                                    .length > 0
                                            "
                                            class="mt-2 sm:mt-3"
                                        >
                                            <div
                                                v-for="url in extractUrls(
                                                    item.response || '',
                                                ).slice(0, 3)"
                                                :key="url"
                                            >
                                                <LinkPreviewComponent
                                                    v-if="
                                                        linkPreviewCache.get(
                                                            url,
                                                        )
                                                    "
                                                    :preview="
                                                        linkPreviewCache.get(
                                                            url,
                                                        )!
                                                    "
                                                />
                                            </div>
                                        </div>
                                    </div>

                                    <!-- Actions - Responsive with fewer labels on mobile -->
                                    <div
                                        v-if="!isLoadingState(item.response)"
                                        class="flex flex-wrap items-center justify-between gap-2 sm:gap-3 mt-3 text-gray-500 dark:text-gray-400 text-sm"
                                    >
                                        <!-- Left side: Navigation for deep search -->
                                        <div
                                            v-if="
                                                isDeepSearchResult(
                                                    item.response,
                                                ) &&
                                                getPagination(i).totalPages > 1
                                            "
                                            class="flex mr-auto items-center gap-2"
                                        >
                                            <Pagination
                                                :items-per-page="1"
                                                :total="
                                                    getPagination(i).totalPages
                                                "
                                                :default-page="
                                                    getPagination(i)
                                                        .currentPage + 1
                                                "
                                                @update:page="
                                                    (newPage) =>
                                                        goToPage(i, newPage - 1)
                                                "
                                            >
                                                <PaginationContent
                                                    v-slot="{ items }"
                                                >
                                                    <PaginationPrevious
                                                        @click="prevResult(i)"
                                                        :disabled="
                                                            getPagination(i)
                                                                .currentPage ===
                                                            0
                                                        "
                                                    />

                                                    <template
                                                        v-for="(
                                                            paginationItem,
                                                            index
                                                        ) in items"
                                                        :key="index"
                                                    >
                                                        <PaginationItem
                                                            v-if="
                                                                paginationItem.type ===
                                                                'page'
                                                            "
                                                            class="bg-white hover:dark:bg-gray-700 dark:bg-gray-900"
                                                            :value="
                                                                paginationItem.value
                                                            "
                                                            :is-active="
                                                                paginationItem.value ===
                                                                getPagination(i)
                                                                    .currentPage +
                                                                    1
                                                            "
                                                            @click="
                                                                goToPage(
                                                                    i,
                                                                    paginationItem.value -
                                                                        1,
                                                                )
                                                            "
                                                        >
                                                            {{
                                                                paginationItem.value
                                                            }}
                                                        </PaginationItem>
                                                        <PaginationEllipsis
                                                            v-else-if="
                                                                paginationItem.type ===
                                                                'ellipsis'
                                                            "
                                                            :key="`ellipsis-${index}`"
                                                        />
                                                    </template>

                                                    <PaginationNext
                                                        class="dark:bg-gray-900 hover:dark:bg-gray-700"
                                                        @click="nextResult(i)"
                                                        :disabled="
                                                            getPagination(i)
                                                                .currentPage >=
                                                            getPagination(i)
                                                                .totalPages -
                                                                1
                                                        "
                                                    />
                                                </PaginationContent>
                                            </Pagination>
                                        </div>

                                        <!-- Right side: Regular actions -->
                                        <div
                                            class="flex flex-wrap ml-auto gap-2 sm:gap-3"
                                        >
                                            <button
                                                @click="
                                                    copyResponse(
                                                        item.response,
                                                        i,
                                                    )
                                                "
                                                class="flex items-center gap-1 hover:text-blue-600 dark:hover:text-blue-400 transition-colors min-h-[32px]"
                                            >
                                                <ClipboardList
                                                    class="w-4 h-4"
                                                />
                                                <span>{{
                                                    copiedIndex === i
                                                        ? "Copied!"
                                                        : "Copy"
                                                }}</span>
                                            </button>

                                            <button
                                                @click="
                                                    shareResponse(
                                                        item.response,
                                                        item.prompt,
                                                    )
                                                "
                                                class="flex items-center gap-1 hover:text-green-600 dark:hover:text-green-400 transition-colors min-h-[32px]"
                                            >
                                                <Share class="w-4 h-4" />
                                                <span>Share</span>
                                            </button>

                                            <button
                                                @click="
                                                    refreshResponse(item.prompt)
                                                "
                                                :disabled="isLoading"
                                                class="flex items-center gap-1 hover:text-orange-600 dark:hover:text-orange-400 transition-colors disabled:opacity-50 disabled:cursor-not-allowed min-h-[32px]"
                                            >
                                                <RotateCw class="w-4 h-4" />
                                                <span>Retry</span>
                                            </button>

                                            <button
                                                @click="deleteMessage(i)"
                                                :disabled="isLoading"
                                                class="flex items-center gap-1 hover:text-red-600 dark:hover:text-red-400 transition-colors disabled:opacity-50 disabled:cursor-not-allowed min-h-[32px]"
                                            >
                                                <Trash class="w-4 h-4" />
                                                <span>Delete</span>
                                            </button>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Responsive Scroll to Bottom Button -->
                <button
                    v-if="showScrollDownButton && currentMessages.length !== 0"
                    @click="scrollToBottom()"
                    :class="[
                        'absolute bg-gray-50 dark:bg-gray-800 text-gray-500 dark:text-gray-400 border dark:border-gray-700 px-4 h-8 rounded-full shadow-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors z-20 flex items-center justify-center gap-2',
                        scrollButtonPosition,
                    ]"
                    :disabled="isRecording"
                    :title="
                        isRecording
                            ? 'Recording in progress'
                            : 'Scroll to bottom'
                    "
                >
                    <ArrowDown
                        class="w-4 h-4"
                        :class="{ 'animate-bounce': !isRecording }"
                    />
                    <span class="text-sm font-medium">Scroll Down</span>
                </button>

                <!-- Input Area -->
                <div
                    :style="
                        screenWidth > 720 && !isCollapsed
                            ? 'left:270px;'
                            : screenWidth > 720 && isCollapsed
                              ? 'left:60px;'
                              : 'left:0px;'
                    "
                    class="bg-white dark:bg-gray-900 z-20 bottom-0 right-0 fixed"
                    :class="pastePreview?.show ? 'pt-2' : ''"
                >
                    <div
                        class="flex items-center justify-center pb-3 sm:pb-5 px-2 sm:px-5"
                    >
                        <form
                            @submit="handleSubmit"
                            class="w-full md:max-w-3xl relative flex bg-gray-50 dark:bg-gray-800 flex-col border-2 dark:border-gray-700 shadow rounded-2xl items-center"
                        >
                            <!-- Paste Preview inside form - above other content -->
                            <div
                                v-if="pastePreview && pastePreview.show"
                                class="w-full p-3 border-b dark:border-gray-700"
                            >
                                <PastePreview
                                    :content="pastePreview.content"
                                    :char-count="pastePreview.charCount"
                                    :word-count="pastePreview.wordCount"
                                />
                            </div>

                            <!-- Request Limit Exceeded Banner -->
                            <div
                                v-if="showLimitExceededBanner"
                                class="py-2 sm:py-3 w-full px-2 sm:px-3"
                            >
                                <div
                                    class="flex items-center justify-center w-full"
                                >
                                    <!-- Mobile: Stacked Layout -->
                                    <div
                                        class="flex sm:hidden w-full flex-col gap-2"
                                    >
                                        <div class="flex items-center gap-2">
                                            <div
                                                class="w-6 h-6 sm:w-8 sm:h-8 bg-red-100 dark:bg-red-900/30 rounded-full flex items-center justify-center flex-shrink-0"
                                            >
                                                <Ban
                                                    class="w-4 h-4 sm:w-5 sm:h-5 text-red-600 dark:text-red-400"
                                                />
                                            </div>
                                            <div class="min-w-0 flex-1">
                                                <h3
                                                    class="text-xs sm:text-sm font-semibold text-red-800 dark:text-red-400 leading-tight"
                                                >
                                                    {{
                                                        planStatus.isExpired
                                                            ? "Plan Expired"
                                                            : "Free Requests Exhausted"
                                                    }}
                                                </h3>
                                                <p
                                                    class="text-xs text-red-600 dark:text-red-400 leading-tight mt-0.5"
                                                >
                                                    {{
                                                        planStatus.isExpired
                                                            ? "Renew your plan"
                                                            : `Used all ${FREE_REQUEST_LIMIT} requests`
                                                    }}
                                                </p>
                                            </div>
                                        </div>
                                        <button
                                            @click="$router.push('/upgrade')"
                                            class="w-full bg-red-500 hover:bg-red-600 dark:bg-red-600 dark:hover:bg-red-700 text-white py-2 rounded-md text-xs font-medium transition-colors"
                                        >
                                            {{
                                                planStatus.isExpired
                                                    ? "Renew Plan"
                                                    : "Upgrade Plan"
                                            }}
                                        </button>
                                    </div>

                                    <!-- Desktop: Horizontal Layout -->
                                    <div
                                        class="hidden sm:flex w-full items-center gap-3"
                                    >
                                        <div
                                            class="w-8 h-8 bg-red-100 dark:bg-red-900/30 rounded-full flex items-center justify-center flex-shrink-0"
                                        >
                                            <Ban
                                                class="w-4 h-4 sm:w-5 sm:h-5 text-red-600 dark:text-red-400"
                                            />
                                        </div>
                                        <div class="min-w-0 flex-1">
                                            <h3
                                                class="text-sm font-semibold text-red-800 dark:text-red-400 mb-1"
                                            >
                                                {{
                                                    planStatus.isExpired
                                                        ? "Plan Expired"
                                                        : "Free Requests Exhausted"
                                                }}
                                            </h3>
                                            <p
                                                class="text-xs text-red-600 dark:text-red-400"
                                            >
                                                {{
                                                    planStatus.isExpired
                                                        ? "Please renew your plan to continue using the service."
                                                        : `You've used all ${FREE_REQUEST_LIMIT} free requests. Upgrade to continue chatting.`
                                                }}
                                            </p>
                                        </div>
                                        <button
                                            @click="$router.push('/upgrade')"
                                            class="bg-red-500 px-3 hover:bg-red-600 dark:bg-red-600 dark:hover:bg-red-700 text-white py-2 rounded-md text-sm font-medium transition-colors flex-shrink-0 whitespace-nowrap"
                                        >
                                            {{
                                                planStatus.isExpired
                                                    ? "Renew Plan"
                                                    : "Upgrade Plan"
                                            }}
                                        </button>
                                    </div>
                                </div>
                            </div>

                            <!-- Upgrade Warning Banner -->
                            <div
                                v-if="showUpgradeBanner"
                                class="py-2 sm:py-3 w-full px-2 sm:px-3"
                            >
                                <div
                                    class="flex items-center justify-center w-full"
                                >
                                    <!-- Mobile: Stacked Layout -->
                                    <div
                                        class="flex sm:hidden w-full flex-col gap-2"
                                    >
                                        <div class="flex items-center gap-2">
                                            <div
                                                class="w-6 h-6 sm:w-8 sm:h-8 bg-yellow-100 dark:bg-yellow-900/30 rounded-full flex items-center justify-center flex-shrink-0"
                                            >
                                                <TriangleAlert
                                                    class="w-4 h-4 sm:w-5 sm:h-5 text-yellow-600 dark:text-yellow-400"
                                                />
                                            </div>
                                            <div class="min-w-0 flex-1">
                                                <h3
                                                    class="text-xs sm:text-sm font-semibold text-yellow-800 dark:text-yellow-400 leading-tight"
                                                >
                                                    {{ requestsRemaining }}
                                                    requests left
                                                </h3>
                                                <p
                                                    class="text-xs text-yellow-600 dark:text-yellow-400 leading-tight mt-0.5"
                                                >
                                                    Upgrade for unlimited access
                                                </p>
                                            </div>
                                        </div>
                                        <button
                                            @click="$router.push('/upgrade')"
                                            class="w-full bg-orange-500 hover:bg-orange-600 dark:bg-orange-600 dark:hover:bg-orange-700 text-white py-2 rounded-md text-xs font-medium transition-colors"
                                        >
                                            Upgrade Plan
                                        </button>
                                    </div>

                                    <!-- Desktop: Horizontal Layout -->
                                    <div
                                        class="hidden sm:flex w-full items-center gap-3"
                                    >
                                        <div
                                            class="w-8 h-8 bg-yellow-100 dark:bg-yellow-900/30 rounded-full flex items-center justify-center flex-shrink-0"
                                        >
                                            <CircleAlert
                                                class="w-4 h-4 sm:w-5 sm:h-5 text-yellow-600 dark:text-yellow-400"
                                            />
                                        </div>
                                        <div class="min-w-0 flex-1">
                                            <h3
                                                class="text-sm font-semibold text-yellow-800 dark:text-yellow-400 mb-1"
                                            >
                                                {{ requestsRemaining }} free
                                                requests remaining
                                            </h3>
                                            <p
                                                class="text-xs text-yellow-600 dark:text-yellow-400"
                                            >
                                                Upgrade to continue chatting
                                                without limits
                                            </p>
                                        </div>
                                        <button
                                            @click="$router.push('/upgrade')"
                                            class="bg-orange-500 px-3 hover:bg-orange-600 dark:bg-orange-600 dark:hover:bg-orange-700 text-white py-2 rounded-md text-sm font-medium transition-colors flex-shrink-0 whitespace-nowrap"
                                        >
                                            Upgrade Plan
                                        </button>
                                    </div>
                                </div>
                            </div>

                            <!-- Input Area with Voice Recording -->
                            <div
                                class="flex flex-col w-full bg-white dark:bg-gray-900 rounded-2xl px-2 sm:px-3 py-2 gap-1 sm:gap-2"
                                :class="
                                    inputDisabled
                                        ? 'opacity-50 border border-t dark:border-gray-700 pointer-events-none'
                                        : showUpgradeBanner
                                          ? 'border border-t dark:border-gray-700'
                                          : ''
                                "
                            >
                                <div
                                    class="w-full items-center justify-center flex"
                                >
                                    <!-- Voice Recording Indicator (when active) - Now aligned horizontally -->
                                    <div
                                        v-if="isRecording || isTranscribing"
                                        class="flex items-center gap-1 sm:gap-2 px-2 py-1 bg-red-50 dark:bg-red-900/30 rounded-lg border border-red-200 dark:border-red-800 text-red-600 dark:text-red-400 text-xs sm:text-sm flex-shrink-0 h-fit"
                                    >
                                        <div class="flex items-center gap-1">
                                            <div
                                                class="w-2 h-2 bg-red-500 dark:bg-red-400 rounded-full animate-pulse"
                                            ></div>
                                            <span class="hidden sm:inline">{{
                                                isTranscribing
                                                    ? "Listening..."
                                                    : "Starting..."
                                            }}</span>
                                            <span class="sm:hidden">{{
                                                isTranscribing ? "🎤" : "⏳"
                                            }}</span>
                                        </div>
                                    </div>

                                    <!-- Clear Voice Button (when transcribed text exists) -->
                                    <button
                                        v-if="transcribedText && !isRecording"
                                        type="button"
                                        @click="clearVoiceTranscription"
                                        class="rounded-lg w-8 h-8 sm:w-9 sm:h-9 flex items-center justify-center transition-colors text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 flex-shrink-0"
                                        title="Clear voice transcription"
                                    >
                                        <X class="w-4 h-4 sm:w-5 sm:h-5" />
                                    </button>

                                    <!-- Textarea - Now takes remaining space alongside the indicator -->
                                    <textarea
                                        required
                                        id="prompt"
                                        name="prompt"
                                        @keydown="onEnter"
                                        @input="autoGrow"
                                        @paste="handlePaste"
                                        :disabled="inputDisabled"
                                        rows="1"
                                        :class="[
                                            'flex-grow py-3 px-3 placeholder:text-gray-500 dark:placeholder:text-gray-400 rounded-xl bg-white dark:bg-gray-900 dark:text-gray-100 text-sm outline-none resize-none max-h-[120px] sm:max-h-[150px] md:max-h-[200px] overflow-auto leading-relaxed min-w-0',
                                            'disabled:opacity-50 disabled:cursor-not-allowed',
                                            isRecording
                                                ? 'bg-red-50 border-red-200 dark:border-red-800'
                                                : 'focus:border-blue-500 dark:focus:border-blue-400',
                                        ]"
                                        :placeholder="inputPlaceholderText"
                                    >
                                    </textarea>
                                </div>

                                <!-- Buttons Row - Below textarea -->
                                <div
                                    class="flex items-center justify-between w-full gap-2"
                                >
                                    <!-- Left side buttons -->
                                    <div class="flex items-center gap-2">
                                        <!-- Microphone Toggle Button -->
                                        <button
                                            type="button"
                                            @click="toggleVoiceRecording"
                                            :disabled="inputDisabled"
                                            :class="[
                                                'rounded-lg w-8 h-8 sm:w-9 sm:h-9 flex items-center justify-center transition-all duration-200 flex-shrink-0',
                                                isRecording
                                                    ? 'bg-red-500 hover:bg-red-600 text-white shadow-lg transform scale-105 animate-pulse'
                                                    : 'bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-600 dark:text-gray-300 hover:text-gray-700 dark:hover:text-gray-200',
                                                'disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none',
                                            ]"
                                            :title="
                                                microphonePermission ===
                                                'denied'
                                                    ? 'Microphone access denied'
                                                    : isRecording
                                                      ? 'Stop voice input'
                                                      : 'Start voice input'
                                            "
                                        >
                                            <!-- Microphone Icon -->
                                            <Mic
                                                v-if="
                                                    microphonePermission ===
                                                    'prompt'
                                                "
                                                class="w-4 h-4 sm:w-5 sm:h-5"
                                            />

                                            <Mic
                                                v-else-if="
                                                    !isRecording &&
                                                    microphonePermission ===
                                                        'granted'
                                                "
                                                class="w-4 h-4 sm:w-5 sm:h-5"
                                            />

                                            <!-- Stop Icon -->
                                            <Pause
                                                v-else-if="
                                                    microphonePermission ===
                                                        'granted' && isRecording
                                                "
                                                class="w-4 h-4 sm:w-5 sm:h-5"
                                            />

                                            <!-- Microphone Denied Icon -->
                                            <MicOff
                                                v-else-if="
                                                    microphonePermission ===
                                                        'denied' && !isRecording
                                                "
                                                class="w-4 h-4 sm:w-5 sm:h-5 text-red-500 dark:text-red-400"
                                            />
                                        </button>

                                        <!-- Mode Dropdown Container -->
                                        <div class="relative flex-shrink-0">
                                            <!-- Dropdown Button - Shows current mode -->
                                            <button
                                                type="button"
                                                @click.stop="
                                                    showInputModeDropdown =
                                                        !showInputModeDropdown
                                                "
                                                :disabled="inputDisabled"
                                                :class="[
                                                    'rounded-lg w-8 h-8 sm:w-9 sm:h-9 flex items-center justify-center transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed shadow-sm border',
                                                    parsedUserDetails?.responseMode ===
                                                    'web-search'
                                                        ? 'border-green-300 bg-green-50 hover:bg-green-100 dark:border-green-600 dark:bg-green-900/30 dark:hover:bg-green-900/50 text-green-700 dark:text-green-300'
                                                        : parsedUserDetails?.responseMode ===
                                                            'deep-search'
                                                          ? 'border-orange-300 bg-orange-50 hover:bg-orange-100 dark:border-orange-600 dark:bg-orange-900/30 dark:hover:bg-orange-900/50 text-orange-700 dark:text-orange-300'
                                                          : 'border-blue-300 bg-blue-50 hover:bg-blue-100 dark:border-blue-600 dark:bg-blue-900/30 dark:hover:bg-blue-900/50 text-blue-700 dark:text-blue-300',
                                                ]"
                                                :title="
                                                    modeOptions[
                                                        parsedUserDetails.responseMode ||
                                                            ''
                                                    ].title
                                                "
                                            >
                                                <!-- Dynamic icon based on selected mode -->
                                                <component
                                                    :is="
                                                        modeOptions[
                                                            parsedUserDetails.responseMode ||
                                                                ''
                                                        ].icon
                                                    "
                                                    class="w-4 h-4 sm:w-5 sm:h-5"
                                                />
                                            </button>

                                            <!-- Dropdown Menu -->
                                            <div
                                                v-show="showInputModeDropdown"
                                                class="absolute bottom-12 left-0 bg-white dark:bg-gray-800 border-[1px] border-gray-300 dark:border-gray-600 rounded-lg shadow-2xl pt-2 z-[100] w-[220px] sm:w-[240px]"
                                                @click.stop
                                            >
                                                <div
                                                    class="px-2 py-1 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide border-b border-gray-200 dark:border-gray-700"
                                                >
                                                    Response Mode
                                                </div>

                                                <!-- Mode Options -->
                                                <button
                                                    v-for="(
                                                        option, key
                                                    ) in modeOptions"
                                                    :key="key"
                                                    type="button"
                                                    @click="
                                                        selectInputMode(
                                                            option.mode,
                                                        )
                                                    "
                                                    :class="[
                                                        'w-full px-3 py-2.5 text-left text-sm flex items-center gap-3 transition-colors',
                                                        parsedUserDetails?.responseMode ===
                                                        option.mode
                                                            ? 'bg-green-100 dark:bg-green-900/40 text-green-700 dark:text-green-300 border-r-2 border-green-500'
                                                            : 'hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-800 dark:text-gray-200',
                                                    ]"
                                                >
                                                    <component
                                                        :class="[
                                                            'w-5 h-5',
                                                            parsedUserDetails?.responseMode ===
                                                            option.mode
                                                                ? ' text-green-600 dark:text-green-400'
                                                                : ' text-gray-600 dark:text-gray-400',
                                                        ]"
                                                        :is="option.icon"
                                                    />
                                                    <div class="flex-1 min-w-0">
                                                        <div
                                                            class="font-semibold"
                                                        >
                                                            {{ option.label }}
                                                        </div>
                                                        <div
                                                            class="text-xs opacity-70"
                                                        >
                                                            {{
                                                                option.description
                                                            }}
                                                        </div>
                                                    </div>
                                                    <Check
                                                        v-if="
                                                            parsedUserDetails?.responseMode ===
                                                            option.mode
                                                        "
                                                        class="w-5 h-5 text-green-600 dark:text-green-400"
                                                    />
                                                </button>
                                            </div>
                                        </div>
                                    </div>

                                    <!-- Submit Button - Right side -->
                                    <button
                                        type="submit"
                                        :disabled="inputDisabled"
                                        class="rounded-lg w-8 h-8 sm:w-9 sm:h-9 flex items-center justify-center transition-colors text-white bg-blue-500 hover:bg-blue-600 dark:bg-blue-600 dark:hover:bg-blue-800 disabled:cursor-not-allowed disabled:opacity-50 disabled:bg-gray-400 flex-shrink-0 shadow-sm"
                                    >
                                        <ArrowUp
                                            v-if="!isLoading"
                                            class="w-4 h-4 sm:w-5 sm:h-5"
                                        />

                                        <LoaderCircle
                                            v-else
                                            class="w-4 h-4 sm:w-5 sm:h-5 animate-spin"
                                        />
                                    </button>
                                </div>
                            </div>
                        </form>
                    </div>
                </div>
            </div>
        </div>
        <TextHightlightPopover />
        <PastePreviewModal
            :data="{
                showPasteModal,
                currentPasteContent,
            }"
            :closePasteModal="closePasteModal"
        />
    </ProtectedPage>
</template>
