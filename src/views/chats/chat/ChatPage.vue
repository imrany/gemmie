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
import TopNav from "@/components/TopNav.vue";
import type {
    ApiResponse,
    Chat,
    ConfirmDialogOptions,
    ContextReference,
    LinkPreview,
    Message,
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
} from "@/lib/videoProcessing";
import { onUpdated } from "vue";
import {
    extractUrls,
    generateChatTitle,
    copyCode,
    isPromptTooShort,
    SPINDLE_URL,
} from "@/lib/globals";
import router from "@/router";
import { copyPasteContent } from "@/lib/previewPasteContent";
import { useRoute } from "vue-router";
import TextHightlightPopover from "@/components/TextHightlightPopover.vue";
import {
    ClipboardList,
    Trash,
    RotateCw,
    Code,
    Pencil,
    BookText,
    HeartPulse,
    Globe,
    LoaderCircle,
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
import EmptyChatView from "@/views/EmptyChatView.vue";
import PastePreview from "@/components/PastePreview.vue";
import { useVoiceRecord } from "@/composables/useVoiceRecord";
import { usePagination } from "@/composables/usePagination";
import { useMessage } from "@/composables/useMessage";
import OverallLayout from "@/layout/OverallLayout.vue";
import InputArea from "@/components/InputArea.vue";
import ReferenceBadge from "@/components/ReferenceBadge.vue";
import ChatErrorSection from "@/components/ChatErrorSection.vue";

// Inject global state
const {
    isOnline,
    showPreviewSidebar,
    handlePaste,
    isChatLoading,
    copiedIndex,
    shouldHaveLimit,
    chatDrafts,
    screenWidth,
    isCollapsed,
    syncStatus,
    isAuthenticated,
    currentChatId,
    currentChat,
    pastePreviews,
    chats,
    planStatus,
    userDetailsDebounceTimer,
    chatsDebounceTimer,
    cancelAllRequests,
    checkRequestLimitBeforeSubmit,
    createNewChat,
    loadChat,
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
    checkInternetConnection,
    activeRequests,
    requestChatMap,
    resetRequestCount,
    incrementRequestCount,
    loadRequestCount,
    FREE_REQUEST_LIMIT,
    requestsRemaining,
    shouldShowUpgradePrompt,
    isRequestLimitExceeded,
    parsedUserDetails,
    apiCall,

    copyResponse,
    fallbackChatId,
    showErrorSection,
    processLinksInUserPrompt,
    processLinksInResponse,
} = inject("globalState") as {
    isOnline: Ref<boolean>;
    showPreviewSidebar: Ref<string | null>;
    isChatLoading: Ref<boolean>;
    fallbackChatId: Ref<string>;
    showErrorSection: Ref<boolean>;
    handlePaste: (event: ClipboardEvent) => void;
    copyResponse: (text: string, index?: number) => void;
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
    expanded: Ref<boolean[]>;
    scrollToLastMessage: () => void;
    showConfirmDialog: (options: ConfirmDialogOptions) => void;
    clearAllChats: () => void;
    switchToChat: (chatId: string) => boolean;
    createNewChat: (initialMessage?: string) => Promise<string>;
    deleteChat: (chatId: string) => void;
    loadChatDrafts: () => void;
    loadChat: (chatId?: string) => Promise<[boolean, string]>;
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
    currentChat: Ref<Chat | undefined>;
    currentMessages: Ref<Message[]>;
    linkPreview: LinkPreview;
    updateExpandedArray: () => void;
    manualSync: () => Promise<any>;
    toggleSidebar: () => void;
    isFreeUser: Ref<boolean>;
    FREE_REQUEST_LIMIT: number;
    isDarkMode: Ref<boolean>;
    hasActiveRequestsForCurrentChat: Ref<boolean>;
    connectionStatus: Ref<string>;
    checkInternetConnection: () => Promise<boolean>;
    autoGrow: (e: Event) => void;
    showSyncIndicator: (message: string, progress?: number) => void;
    hideSyncIndicator: () => void;
    updateSyncProgress: (message: string, progress: number) => void;
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
    apiCall: <T>(
        endpoint: string,
        options: RequestInit,
    ) => Promise<ApiResponse<T>>;
};

const route = useRoute();
// ---------- State ----------
const now = ref(Date.now());
const showInputModeDropdown = ref(false);
const isLoading = ref(false);

const isRecording = ref(false);
const selectedContexts = ref<ContextReference[]>([]);
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
const pastePreview = computed(() => {
    return pastePreviews.value.get(currentChatId.value) || null;
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

const chatContainerWidth = computed(() => {
    if (screenWidth.value > 720) {
        return showPreviewSidebar.value
            ? "w-full ease-in-out transition-width duration-300"
            : "w-[85%] ease-in-out transition-width duration-300";
    } else {
        return "w-full";
    }
});

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

function clearContextReferences() {
    selectedContexts.value = [];
}

// handleSubmit function
async function handleSubmit(
    e?: any,
    retryPrompt?: string,
    forceMode?: "web-search" | "deep-search" | "light-response",
) {
    e?.preventDefault?.();

    // Create context references from selected contexts
    const effectiveContextReferences =
        selectedContexts.value.length > 0 ? selectedContexts.value : undefined;

    // Stop voice recording immediately
    if (isRecording.value || isTranscribing.value) {
        stopVoiceRecording(true);
    }

    // Check internet connection
    if (!isOnline.value) {
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

    // Handle paste preview content
    const currentPastePreview = pastePreviews.value.get(currentChatId.value);
    const hasPastePreview =
        currentPastePreview && currentPastePreview.show && !retryPrompt;

    if (hasPastePreview) {
        promptValue += currentPastePreview.content;
        pastePreviews.value.delete(currentChatId.value);
    }

    // Validate prompt
    if (!promptValue || isLoading.value) return;

    if (!isAuthenticated.value) {
        toast.warning("Please create a session first", {
            duration: 4000,
            description: "You need to be logged in.",
        });
        return;
    }

    // Load and check request limits
    loadRequestCount();

    // Ensure we have a valid chat
    let submissionChatId = currentChatId.value;
    let submissionChat = chats.value.find(
        (chat) => chat.id === submissionChatId,
    );

    if (
        submissionChat &&
        (submissionChat?.messages === null ||
            submissionChat?.messages === undefined ||
            !submissionChat?.messages)
    ) {
        submissionChat.messages = [];
    }

    // Create new chat if needed
    if (!submissionChatId || !submissionChat) {
        submissionChatId = await createNewChat(promptValue);
        if (!submissionChatId) {
            toast.error("Failed to create chat", {
                duration: 3000,
                description: "Please try again",
            });
            return;
        }

        // Wait for Vue to update the reactive array
        await nextTick();

        const [success, message] = await loadChat(submissionChatId);

        // Verify chat exists
        if (!success) {
            console.error(message);
            toast.error(message, {
                duration: 3000,
                description: "Please refresh and try again",
            });
            return;
        }
    }

    // Generate unique request ID and setup abort controller
    const requestId = `req_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`;
    const abortController = new AbortController();

    activeRequests.value.set(requestId, abortController);
    requestChatMap.value.set(requestId, submissionChatId);

    // Determine response mode
    let responseMode =
        forceMode || parsedUserDetails?.value.responseMode || "light-response";

    const isSearchMode =
        responseMode === "web-search" || responseMode === "deep-search";

    // Handle link-only prompts specially
    if (isJustLinks(promptValue) && !isSearchMode) {
        return handleLinkOnlyRequest(
            promptValue,
            submissionChatId,
            requestId,
            abortController,
            effectiveContextReferences?.map((ref) => ref.preview) || [],
        );
    }

    // Override to light-response if pasted content detected
    if (hasPastePreview) {
        responseMode = "light-response";
        console.log("Pasted content detected - using light-response mode");
    }

    // Build fabricated prompt with context
    let fabricatedPrompt = promptValue;

    // Add explicit context references if provided
    if (effectiveContextReferences && effectiveContextReferences.length > 0) {
        let contextInfo = "";

        if (isSearchMode) {
            // For search modes, integrate context differently
            effectiveContextReferences.forEach((ctx) => {
                contextInfo += `${ctx.fullText} `;
            });
            fabricatedPrompt = `${promptValue} ${contextInfo}`;
        } else {
            // For light-response, put context before the question
            contextInfo += "\n\n--- Context from previous messages ---\n";

            effectiveContextReferences.forEach((ctx, idx) => {
                contextInfo += `\n[Reference ${idx + 1}]:\n${ctx.fullText}\n`;
            });
            contextInfo += "\n--- End of context ---\n\n";
            fabricatedPrompt =
                contextInfo + `User's current question: ${promptValue}`;
        }

        console.log(
            `Added ${effectiveContextReferences.length} explicit context reference(s) for ${responseMode} mode`,
        );
    }

    isLoading.value = true;
    scrollToLastMessage();

    // Create temporary message
    const tempResp: Message = {
        id: `temp_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`,
        chat_id: submissionChatId,
        created_at: new Date().toISOString(),
        prompt: promptValue,
        response: responseMode ? `${responseMode}...` : "...",
        references: effectiveContextReferences?.map((ref) => ref.preview) || [],
        model: "gemini-pro",
    };

    // Store original selectedContexts for rollback on error
    const originalSelectedContexts = [...selectedContexts.value];
    let tempMessageIndex = -1;

    try {
        // Verify submissionChat still exists (defensive check)
        if (!submissionChat || !submissionChat.messages) {
            console.error("Invalid chat state - attempting recovery");

            // Try to recover by finding the chat again
            submissionChat = chats.value.find(
                (chat) => chat.id === submissionChatId,
            );

            // If still invalid, abort
            if (!submissionChat) {
                isLoading.value = false;
                toast.error("Chat initialization error", {
                    duration: 3000,
                    description: "Please refresh and try again",
                    action: {
                        text: "Refresh",
                        onClick: () => location.reload(),
                    },
                });
                return;
            }

            // Initialize messages array if missing
            if (
                submissionChat.messages === null ||
                submissionChat.messages === undefined ||
                !submissionChat.messages
            ) {
                submissionChat.messages = [];
            }
        }

        // Add temporary message to chat
        submissionChat.messages.push(tempResp);
        tempMessageIndex = submissionChat.messages.length - 1;
        submissionChat.last_message_at = new Date().toISOString();

        // Update chat title if first message
        if (submissionChat.messages.length === 1) {
            submissionChat.title = generateChatTitle(promptValue);
        }

        updateExpandedArray();
        await processLinksInUserPrompt(tempMessageIndex);

        let searchResults = "";
        // Make API request based on mode
        if (isSearchMode) {
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

            const response = await fetch(
                `${SPINDLE_URL}/search?${searchParams}`,
                {
                    method: "GET",
                    signal: abortController.signal,
                    headers: {
                        "Content-Type": "application/json",
                    },
                },
            );

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();

            if (data && data.error) {
                throw new Error(`API error: ${data.error}`);
            }

            const hasResults = data.results || data.json;

            if (!hasResults) {
                throw new Error("No search results found for your query.");
            }

            const searchResultsSummary = formatSearchResults(
                data,
                responseMode,
                tempMessageIndex,
            );
            searchResults = searchResultsSummary;
            // Append search results to the fabricated prompt for the main API call
            // fabricatedPrompt = `${promptValue}\n\n--- Search Results ---\n${searchResultsSummary}\n\nProvide detailed information based on the search results above. Maintain the same format.`;
        }

        const response = await apiCall<Message>(
            `/chats/${submissionChatId}/messages`,
            {
                method: "POST",
                signal: abortController.signal,
                body: JSON.stringify({
                    prompt: fabricatedPrompt,
                    response: searchResults,
                    references:
                        effectiveContextReferences?.map((ref) => ref.preview) ||
                        [],
                }),
            },
        );

        // Check if request was aborted
        if (abortController.signal.aborted) {
            console.log(`Request ${requestId} was aborted`);
            removeTemporaryMessage(submissionChatId, tempMessageIndex);
            selectedContexts.value = originalSelectedContexts;
            return;
        }

        // Handle API errors
        if (!response.success) {
            const errorText = response.message;
            throw new Error(errorText);
        }

        const parseRes = response.data;

        // Clear draft for current chat
        clearCurrentDraft();

        // Update the response in chat (replace temporary message)
        const updatedTargetChat = chats.value.find(
            (chat) => chat.id === submissionChatId,
        );

        if (updatedTargetChat && tempMessageIndex >= 0 && parseRes) {
            const updatedMessage: Message = {
                id: parseRes.id,
                chat_id: submissionChatId,
                created_at: parseRes.created_at,
                prompt: parseRes.prompt,
                response: parseRes.response,
                references: parseRes.references,
                model: parseRes.model,
            };

            updatedTargetChat.messages[tempMessageIndex] = updatedMessage;
            updatedTargetChat.last_message_at = parseRes.created_at;

            await processLinksInResponse(tempMessageIndex);

            // SUCCESS - Increment request count
            incrementRequestCount();

            // Clear input field
            if (!retryPrompt && e?.target?.prompt) {
                e.target.prompt.value = "";
                e.target.prompt.style.height = "auto";
            }

            // Clear voice transcription
            if (transcribedText.value) {
                transcribedText.value = "";
            }

            // Show context feedback ONLY on success
            if (
                effectiveContextReferences &&
                effectiveContextReferences.length > 0
            ) {
                toast.success(
                    `Used ${effectiveContextReferences.length} reference(s)`,
                    {
                        duration: 2000,
                        description: "Previous messages added as context",
                    },
                );
            }

            // Clear selected contexts ONLY on success
            selectedContexts.value = [];

            // Show success notification if user switched away
            if (currentChatId.value !== submissionChatId) {
                const targetChat = chats.value.find(
                    (chat) => chat.id === submissionChatId,
                );
                toast.success("Response received", {
                    duration: 3000,
                    description: `Switch to "${targetChat?.title || "chat"}" to view the response`,
                });
            }
        }

        currentChatId.value = submissionChatId;
    } catch (err: any) {
        // Handle abort
        if (err.name === "AbortError") {
            console.log(`Request ${requestId} was aborted`);
            removeTemporaryMessage(submissionChatId, tempMessageIndex);
            selectedContexts.value = originalSelectedContexts;
            return;
        }

        console.error("AI Response Error:", err);

        // Remove temporary message on error
        removeTemporaryMessage(submissionChatId, tempMessageIndex);

        // Restore original context selection on error
        selectedContexts.value = originalSelectedContexts;

        // More specific error messages
        let errorMessage = err.message;
        let description = "Please try again or check your connection";

        if (err.message.includes("Failed to fetch")) {
            errorMessage = "Network Error";
            description = "Please check your internet connection";
        } else if (err.message.includes("timeout")) {
            errorMessage = "Request Timeout";
            description = "The request took too long. Please try again";
        } else if (err.message.includes("Invalid chat state")) {
            errorMessage = "Chat Error";
            description = "Please refresh the page and try again";
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
    contextReferenceIds: string[],
) {
    const urls = extractUrls(promptValue);

    isLoading.value = true;
    scrollToLastMessage();

    // Store temporary message reference
    let tempMessageIndex = -1;
    const tempResp: Message = {
        id: `temp_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`,
        chat_id: chatId,
        created_at: new Date().toISOString(),
        prompt: promptValue,
        response: "...",
        references: contextReferenceIds,
        model: "gemini-pro",
    };
    const targetChat = chats.value.find((chat) => chat.id === chatId);

    if (targetChat) {
        targetChat.messages.push(tempResp);
        tempMessageIndex = targetChat.messages.length - 1;
        targetChat.last_message_at = new Date().toISOString();
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

        // Save to backend
        const response = await apiCall<Message>(`/chats/${chatId}/messages`, {
            method: "POST",
            signal: abortController.signal,
            body: JSON.stringify({
                prompt: promptValue,
                response: combinedResponse.trim(),
                references: contextReferenceIds,
            }),
        });

        // Check if request was aborted
        if (abortController.signal.aborted) {
            console.log(`Link request ${requestId} was aborted`);
            removeTemporaryMessage(chatId, tempMessageIndex);
            return;
        }

        // Handle API errors
        if (!response.success) {
            throw new Error(response.message);
        }

        const parseRes = response.data;

        // Update the response in chat (replace temporary message)
        const updatedTargetChat = chats.value.find(
            (chat) => chat.id === parseRes?.chat_id,
        );

        if (updatedTargetChat && tempMessageIndex >= 0 && parseRes) {
            const updatedMessage: Message = {
                id: parseRes.id,
                chat_id: parseRes.chat_id,
                created_at: parseRes.created_at,
                prompt: parseRes.prompt,
                response: parseRes.response,
                references: parseRes.references || [],
                model: parseRes.model || "gemini-pro",
            };

            updatedTargetChat.messages[tempMessageIndex] = updatedMessage;
            updatedTargetChat.last_message_at = parseRes.created_at;

            // ONLY INCREMENT ON SUCCESS for link-only prompts
            incrementRequestCount();

            // Show notification if user switched away
            if (currentChatId.value !== chatId) {
                toast.success("Links analyzed", {
                    duration: 3000,
                    description: `Switch to "${targetChat?.title || "chat"}" to view the analysis`,
                });
            }
        }
    } catch (err: any) {
        // Handle abort
        if (err.name === "AbortError") {
            console.log(`Link request ${requestId} was aborted`);
            removeTemporaryMessage(chatId, tempMessageIndex);
            return;
        }

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
    }
}

async function refreshResponse(
    oldPrompt?: string,
    originalReferences?: string[],
) {
    if (!isOnline.value) {
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

    // Detect original mode from response content
    let originalMode: "web-search" | "deep-search" | "light-response" | string =
        "light-response";

    if (isDeepSearchResult(oldMessage.response || "")) {
        originalMode = "deep-search";
    } else if (
        oldMessage.response?.includes("light-search") ||
        oldMessage.response?.includes("### 1.") ||
        oldMessage.response?.includes("**Source:**")
    ) {
        originalMode = "web-search";
    } else {
        originalMode =
            parsedUserDetails?.value?.responseMode || "light-response";
    }

    const isSearchMode =
        originalMode === "web-search" || originalMode === "deep-search";

    console.log(`Refreshing with detected mode: ${originalMode}`);

    // Reconstruct context references if they exist
    let contextReferences: ContextReference[] | undefined;

    if (originalReferences && originalReferences.length > 0) {
        contextReferences = originalReferences.map((previewText) => {
            // The preview is a truncated prompt, find the matching message
            const messageIndex = currentMessages.value.findIndex(
                (m) =>
                    m.prompt &&
                    m.prompt.startsWith(previewText.replace("...", "").trim()),
            );

            if (messageIndex >= 0) {
                const refMessage = currentMessages.value[messageIndex];

                return {
                    preview: previewText,
                    fullText:
                        refMessage.response || refMessage.prompt || previewText,
                };
            }

            // Fallback if message not found
            return {
                preview: previewText,
                fullText: previewText,
            };
        });

        console.log(
            `Refreshing with ${contextReferences.length} context reference(s)`,
        );
    }

    // Build fabricated prompt with context
    let fabricatedPrompt = oldPrompt;

    // Add context if available
    if (contextReferences && contextReferences.length > 0) {
        let contextInfo = "\n\n--- Context from previous messages ---\n";
        contextReferences.forEach((ctx, idx) => {
            contextInfo += `\n[Reference ${idx + 1}]:\n${ctx.fullText}\n`;
        });
        contextInfo += "\n--- End of context ---\n\n";

        if (isSearchMode) {
            fabricatedPrompt = `${oldPrompt}\n\nRelevant context:\n${contextInfo}`;
        } else {
            fabricatedPrompt =
                contextInfo + `User's current question: ${oldPrompt}`;
        }
    }
    // Add implicit context if prompt is short and no explicit context
    else if (
        oldPrompt &&
        isPromptTooShort(oldPrompt) &&
        currentMessages.value.length > 1 &&
        !isSearchMode
    ) {
        const lastMessage = currentMessages.value[msgIndex - 1];
        fabricatedPrompt = `Previous: ${lastMessage.prompt || ""} ${lastMessage.response || ""}\n\nCurrent: ${oldPrompt}`;
    }

    // Check request limits for refresh too
    if (!checkRequestLimitBeforeSubmit()) {
        return;
    }

    // Show placeholder while refreshing
    chat.messages[msgIndex] = {
        ...oldMessage,
        response: originalMode ? `${originalMode}...` : "Refreshing...",
        references: originalReferences || [], // PRESERVE REFERENCES
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

            // Save to backend
            const response = await apiCall<Message>(
                `/chats/${currentChatId.value}/messages`,
                {
                    method: "POST",
                    body: JSON.stringify({
                        prompt: oldPrompt,
                        response: combinedResponse.trim(),
                        references: originalReferences || [],
                    }),
                },
            );

            // Handle API errors
            if (!response.success) {
                throw new Error(response.message);
            }

            const parseRes = response.data;

            if (parseRes) {
                // Replace the same message with the refreshed response
                const updatedMessage: Message = {
                    id: parseRes.id,
                    chat_id: parseRes.chat_id,
                    created_at: parseRes.created_at,
                    prompt: parseRes.prompt,
                    response: parseRes.response,
                    references: parseRes.references || [],
                    model: parseRes.model || "gemini-pro",
                };

                chat.messages[msgIndex] = updatedMessage;
                chat.updated_at = new Date().toISOString();
                chat.last_message_at = parseRes.created_at;

                await processLinksInResponse(msgIndex);
                incrementRequestCount();
            }
        } catch (err: any) {
            console.error("Link refresh error:", err);
            toast.error(`Failed to refresh link analysis: ${err.message}`);

            // Restore original message on error
            chat.messages[msgIndex] = oldMessage;
        } finally {
            isLoading.value = false;
            observeNewVideoContainers();
        }

        return;
    }

    try {
        let finalResponse = "";

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

            const searchResponse = await fetch(
                `${SPINDLE_URL}/search?${searchParams}`,
                {
                    method: "GET",
                    headers: {
                        "Content-Type": "application/json",
                    },
                },
            );

            if (!searchResponse.ok) {
                throw new Error(
                    `HTTP ${searchResponse.status}: ${searchResponse.statusText}`,
                );
            }

            const searchData = await searchResponse.json();

            if (searchData && searchData.error) {
                throw new Error(`API error: ${searchData.error}`);
            }

            const hasResults = searchData.results || searchData.json;

            if (!hasResults) {
                throw new Error("No search results found for your query.");
            }

            const searchResultsSummary = formatSearchResults(
                searchData,
                originalMode,
                msgIndex,
            );

            finalResponse = searchResultsSummary;
            // Append search results to the fabricated prompt
            // fabricatedPrompt = `${oldPrompt}\n\n--- Search Results ---\n${searchResultsSummary}\n\nProvide detailed information based on the search results above. Maintain the same format.`;
        }

        // Make the main API call to save the message
        const response = await apiCall<Message>(
            `/chats/${currentChatId.value}/messages`,
            {
                method: "POST",
                body: JSON.stringify({
                    prompt: oldPrompt,
                    response: finalResponse,
                    references: originalReferences || [],
                }),
            },
        );

        // Handle API errors
        if (!response.success) {
            throw new Error(response.message);
        }

        const parseRes = response.data;

        // Process response based on mode
        if (parseRes) {
            const updatedMessage: Message = {
                id: parseRes.id,
                chat_id: parseRes.chat_id,
                created_at: parseRes.created_at,
                prompt: parseRes.prompt,
                response: parseRes.response,
                references: parseRes.references || [],
                model: parseRes.model || "gemini-pro",
            };

            // Replace the same message with the refreshed response
            chat.messages[msgIndex] = updatedMessage;
            chat.updated_at = new Date().toISOString();
            chat.last_message_at = parseRes.created_at;

            await processLinksInResponse(msgIndex);
            incrementRequestCount();
        }
    } catch (err: any) {
        console.error("Refresh error:", err);
        toast.error(`Failed to refresh response: ${err.message}`);

        // Restore original message on error
        chat.messages[msgIndex] = oldMessage;
    } finally {
        isLoading.value = false;
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
    return (
        isLoading.value || isRequestLimitExceeded.value || isChatLoading.value
    );
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
        return "pb-[240px] sm:pb-[240px]";
    } else if (isRequestLimitExceeded.value || shouldShowUpgradePrompt.value) {
        return "pb-[240px] sm:pb-[240px]";
    } else if (pastePreview.value?.show) {
        return "pb-[220px] sm:pb-[220px]";
    } else if (showScrollDownButton.value) {
        return "pb-[210px] sm:pb-[190px]";
    } else {
        // Base padding when nothing special is showing
        return "pb-[190px] sm:pb-[195px]";
    }
});

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

onUpdated(() => {
    // Check for new video containers after DOM updates
    observeNewVideoContainers();
});

// Watch for chat switches to manage requests
watch(currentChatId, (newChatId, oldChatId) => {
    loadChatDrafts();

    if (oldChatId && newChatId !== oldChatId) {
        showErrorSection.value = false;
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
    document.removeEventListener("click", handleClickOutside);

    // Clear debounce timers
    if (chatsDebounceTimer) {
        clearTimeout(chatsDebounceTimer);
    }
    if (userDetailsDebounceTimer) {
        clearTimeout(userDetailsDebounceTimer);
    }
});

onMounted(async () => {
    const path = router.currentRoute.value.path;
    const routeChatId = route.params.id;
    const chatId = Array.isArray(routeChatId) ? routeChatId[0] : routeChatId;

    // Wait for authentication and chats to load
    await nextTick();

    if (!isAuthenticated.value) {
        router.push(`/?from=chat/${chatId}`);
        return;
    }

    // Handle /new route - create new chat immediately
    if (path === "/new") {
        currentChatId.value = "";
        showErrorSection.value = false;
        return;
    }

    // For specific chat routes, set currentChatId immediately
    // This prevents the app from thinking there's no active chat
    if (chatId) {
        currentChatId.value = chatId;
    }

    // Load cached data and setup handlers
    loadLinkPreviewCache();

    // Setup global functions
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

        setupGlobalFunction("copyPasteContent", copyPasteContent);
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

    // Setup event listeners
    document.addEventListener("click", handleClickOutside);

    const copyListener = (e: any) => {
        if (e.target?.classList.contains("copy-button")) {
            const code = decodeURIComponent(e.target.getAttribute("data-code"));
            copyCode(code, e.target);
        }
    };
    document.addEventListener("click", copyListener);

    // Initialize core features
    initializeSpeechRecognition();
    initializeVideoLazyLoading();

    // Setup periodic tasks
    const interval = setInterval(() => {
        now.value = Date.now();
    }, 1000);

    const resetCheckInterval = setInterval(loadRequestCount, 5 * 60 * 1000);

    // Initialize authentication-dependent features
    if (isAuthenticated.value) {
        const shouldHaveLimit =
            isFreeUser.value ||
            planStatus.value.isExpired ||
            planStatus.value.status === "no-plan" ||
            planStatus.value.status === "expired";

        if (shouldHaveLimit) {
            loadRequestCount();
        }

        // Pre-process links in existing messages
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
    }

    // DOM-dependent functionality
    nextTick(() => {
        const textarea = document.getElementById(
            "prompt",
        ) as HTMLTextAreaElement;

        // Setup scroll handling
        if (scrollableElem.value) {
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

        // Process link previews in responses
        currentMessages.value.forEach((msg: Message, index) => {
            if (
                msg.response &&
                msg.response !== "..." &&
                !msg.response.endsWith("...")
            ) {
                processLinksInResponse(index);
            }
        });

        // Setup for scroll and video observation
        setTimeout(() => {
            scrollToLastMessage();
            observeNewVideoContainers();
            handleScroll();
        }, 300);
    });

    // Store cleanup references
    onBeforeUnmount(() => {
        // Clean up event listeners
        document.removeEventListener("click", copyListener);
        document.removeEventListener("click", handleClickOutside);

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
        const timers = [scrollTimeout, transcriptionTimer, updateTimeout];
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

    if (currentChatId.value) {
        currentChatId.value = "";
    }
});
</script>

<template>
    <OverallLayout>
        <!-- Main Chat Window -->
        <div
            :class="[
                screenWidth > 720
                    ? !isCollapsed
                        ? 'flex-grow flex flex-col items-center justify-center ml-[270px] font-light text-sm transition-all duration-300 ease-in-out bg-inherit'
                        : 'flex-grow flex flex-col items-center justify-center ml-[60px] font-light text-sm transition-all duration-300 ease-in-out bg-inherit'
                    : 'text-sm font-light flex-grow items-center justify-center flex flex-col transition-all duration-300 ease-in-out bg-inherit',
            ]"
        >
            <ChatErrorSection
                v-if="showErrorSection"
                :fallback-chat-id="fallbackChatId"
                :show-error-section="ref(showErrorSection)"
            />

            <div class="relative w-full" v-else>
                <div
                    :class="[
                        'w-full h-screen flex flex-col justify-center bg-inherit',
                        screenWidth > 720 && !isCollapsed
                            ? 'left-[270px]'
                            : screenWidth > 720 && isCollapsed
                              ? 'left-[60px]'
                              : 'left-[0]',
                    ]"
                >
                    <TopNav />
                    <!-- Chat Messages Container -->
                    <div
                        :class="[
                            'flex flex-col w-full items-center h-full pt-[70px] px-2 ',
                        ]"
                    >
                        <!-- Empty State -->
                        <EmptyChatView
                            :class="[chatContainerWidth]"
                            v-if="currentMessages.length === 0"
                            :suggestionPrompts="suggestionPrompts"
                            :selectSuggestion="selectSuggestion"
                        />

                        <div
                            v-else
                            ref="scrollableElem"
                            :class="[
                                'relative md:max-w-3xl min-h-[calc(100vh-100px)] max-w-[100vw] flex-grow no-scrollbar overflow-y-auto space-y-3 sm:space-y-4  scroll-container',
                                scrollContainerPadding,
                                chatContainerWidth,
                            ]"
                        >
                            <div class="flex flex-col gap-1">
                                <div
                                    v-for="(item, i) in currentMessages"
                                    :key="`chat-${i}`"
                                    :id="`chat-${item.response}`"
                                    class="flex flex-col gap-1"
                                >
                                    <!-- User Bubble -->
                                    <div class="flex w-full chat-message">
                                        <div class="flex flex-col w-full">
                                            <div
                                                class="flex justify-start mb-3"
                                            >
                                                <PastePreview
                                                    :content="item?.prompt"
                                                    :is-clickable="true"
                                                />
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
                                                    <!-- Selected Context Badges  -->
                                                    <ReferenceBadge
                                                        :is-closeable="false"
                                                        :selected-contexts="
                                                            ref(
                                                                item.references &&
                                                                    item
                                                                        .references
                                                                        .length >
                                                                        0
                                                                    ? item.references.map(
                                                                          (
                                                                              previewText,
                                                                          ) => {
                                                                              const messageIndex =
                                                                                  currentMessages.findIndex(
                                                                                      (
                                                                                          m,
                                                                                      ) =>
                                                                                          m.prompt &&
                                                                                          m.prompt.startsWith(
                                                                                              previewText
                                                                                                  .replace(
                                                                                                      '...',
                                                                                                      '',
                                                                                                  )
                                                                                                  .trim(),
                                                                                          ),
                                                                                  );

                                                                              if (
                                                                                  messageIndex >=
                                                                                  0
                                                                              ) {
                                                                                  const refMessage =
                                                                                      currentMessages[
                                                                                          messageIndex
                                                                                      ];
                                                                                  return {
                                                                                      preview:
                                                                                          previewText,
                                                                                      fullText:
                                                                                          refMessage.response ||
                                                                                          refMessage.prompt ||
                                                                                          previewText,
                                                                                  };
                                                                              }

                                                                              return {
                                                                                  preview:
                                                                                      previewText,
                                                                                  fullText:
                                                                                      previewText,
                                                                              };
                                                                          },
                                                                      )
                                                                    : [],
                                                            )
                                                        "
                                                    />
                                                    <MarkdownRenderer
                                                        class="break-words text-base leading-relaxed"
                                                        :content="
                                                            item &&
                                                            item?.prompt &&
                                                            item?.prompt
                                                                ?.length > 800
                                                                ? item?.prompt
                                                                      ?.trim()
                                                                      .split(
                                                                          '#pastedText#',
                                                                      )[0]
                                                                : item.prompt ||
                                                                  ''
                                                        "
                                                        :maxHeight="150"
                                                        :minLength="400"
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
                                                v-if="
                                                    isLoadingState(
                                                        item.response,
                                                    )
                                                "
                                                class="flex w-full rounded-lg bg-gray-50 dark:bg-gray-800 p-2 items-center animate-pulse gap-2 text-gray-500 dark:text-gray-400"
                                            >
                                                <LoaderCircle
                                                    class="w-4 h-4 animate-spin"
                                                />
                                                <span class="text-sm">{{
                                                    getLoadingMessage(
                                                        item.response,
                                                    )
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
                                                        :collapsible="false"
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
                                                        :collapsible="false"
                                                        class="break-words overflow-x-hidden"
                                                        :content="
                                                            item.response || ''
                                                        "
                                                    />
                                                </template>

                                                <!-- Link Previews Section -->
                                                <div
                                                    v-if="
                                                        !isDeepSearchResult(
                                                            item.response,
                                                        ) &&
                                                        extractUrls(
                                                            item.response || '',
                                                        ).length > 0
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
                                                v-if="
                                                    !isLoadingState(
                                                        item.response,
                                                    )
                                                "
                                                class="flex flex-wrap items-center justify-between gap-2 sm:gap-3 mt-3 text-gray-500 dark:text-gray-400 text-sm"
                                            >
                                                <!-- Left side: Navigation for deep search -->
                                                <div
                                                    v-if="
                                                        isDeepSearchResult(
                                                            item.response,
                                                        ) &&
                                                        getPagination(i)
                                                            .totalPages > 1
                                                    "
                                                    class="flex mr-auto items-center gap-2"
                                                >
                                                    <Pagination
                                                        :items-per-page="1"
                                                        :total="
                                                            getPagination(i)
                                                                .totalPages
                                                        "
                                                        :default-page="
                                                            getPagination(i)
                                                                .currentPage + 1
                                                        "
                                                        @update:page="
                                                            (newPage) =>
                                                                goToPage(
                                                                    i,
                                                                    newPage - 1,
                                                                )
                                                        "
                                                    >
                                                        <PaginationContent
                                                            v-slot="{ items }"
                                                        >
                                                            <PaginationPrevious
                                                                @click="
                                                                    prevResult(
                                                                        i,
                                                                    )
                                                                "
                                                                :disabled="
                                                                    getPagination(
                                                                        i,
                                                                    )
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
                                                                        getPagination(
                                                                            i,
                                                                        )
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
                                                                @click="
                                                                    nextResult(
                                                                        i,
                                                                    )
                                                                "
                                                                :disabled="
                                                                    getPagination(
                                                                        i,
                                                                    )
                                                                        .currentPage >=
                                                                    getPagination(
                                                                        i,
                                                                    )
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
                                                        v-if="
                                                            currentChat &&
                                                            !currentChat.is_read_only
                                                        "
                                                        @click="
                                                            refreshResponse(
                                                                item.prompt,
                                                                item.references,
                                                            )
                                                        "
                                                        :disabled="
                                                            isLoading ||
                                                            !isOnline
                                                        "
                                                        class="flex items-center gap-1 hover:text-orange-600 dark:hover:text-orange-400 transition-colors disabled:opacity-50 disabled:cursor-not-allowed min-h-[32px]"
                                                    >
                                                        <RotateCw
                                                            class="w-4 h-4"
                                                        />
                                                        <span>Retry</span>
                                                    </button>

                                                    <button
                                                        v-if="
                                                            currentChat &&
                                                            !currentChat.is_read_only
                                                        "
                                                        @click="
                                                            deleteMessage(i)
                                                        "
                                                        :disabled="
                                                            isLoading ||
                                                            !isOnline
                                                        "
                                                        class="flex items-center gap-1 hover:text-red-600 dark:hover:text-red-400 transition-colors disabled:opacity-50 disabled:cursor-not-allowed min-h-[32px]"
                                                    >
                                                        <Trash
                                                            class="w-4 h-4"
                                                        />
                                                        <span>Delete</span>
                                                    </button>
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- Input Area -->
                    <InputArea
                        :class="['px-2']"
                        :chat-container-width="chatContainerWidth"
                        :is-loading="isLoading"
                        :show-input="
                            !(currentChat && currentChat?.is_read_only)
                        "
                        :is-recording="isRecording"
                        :is-transcribing="isTranscribing"
                        :transcribed-text="transcribedText"
                        :microphone-permission="microphonePermission"
                        :input-disabled="
                            (currentChat &&
                                currentChat?.is_read_only &&
                                !isOnline) ||
                            inputDisabled ||
                            !isOnline
                        "
                        :input-placeholder-text="inputPlaceholderText"
                        :paste-preview="
                            pastePreview
                                ? {
                                      show: pastePreview.show,
                                      content: pastePreview.content,
                                  }
                                : null
                        "
                        :show-input-mode-dropdown="showInputModeDropdown"
                        :show-limit-exceeded-banner="showLimitExceededBanner"
                        :show-upgrade-banner="showUpgradeBanner"
                        :plan-status="planStatus"
                        :FREE_REQUEST_LIMIT="FREE_REQUEST_LIMIT"
                        :selected-contexts="ref(selectedContexts)"
                        :show-scroll-down-button="ref(showScrollDownButton)"
                        :scroll-button-position="ref(scrollButtonPosition)"
                        @scroll-to-bottom="scrollToBottom"
                        @submit="handleSubmit"
                        @auto-grow="autoGrow"
                        @handle-paste="handlePaste"
                        @keydown="onEnter"
                        @toggle-voice-recording="toggleVoiceRecording"
                        @clear-voice-transcription="clearVoiceTranscription"
                        @toggle-input-mode-dropdown="
                            showInputModeDropdown = !showInputModeDropdown
                        "
                        @select-input-mode="selectInputMode"
                        @navigate-to-upgrade="router.push('/upgrade')"
                        @remove-context="
                            (index: number) => selectedContexts.splice(index, 1)
                        "
                        @clear-all-contexts="clearContextReferences"
                    />
                </div>
            </div>
        </div>
        <TextHightlightPopover />
    </OverallLayout>
</template>
