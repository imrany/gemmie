<script lang="ts" setup>
import { computed, ref, watch, onMounted, onBeforeUnmount } from "vue";
import type { Ref } from "vue";
import {
    Mic,
    MicOff,
    Pause,
    ArrowUp,
    LoaderCircle,
    X,
    Brain,
    Search,
    Library,
    Check,
    MessageSquare,
    ArrowDown,
} from "lucide-vue-next";
import PastePreview from "@/components/PastePreview.vue";
import type {
    UserDetails,
    Res,
    ContextReference,
    ModeOption,
    CurrentChat,
} from "@/types";
import { inject } from "vue";
import ReferenceBadge from "./ReferenceBadge.vue";

const {
    isLoading,
    parsedUserDetails,
    requestsRemaining,
    screenWidth,
    FREE_REQUEST_LIMIT,
    currentMessages,
    isCollapsed,
    currentChat,
} = inject("globalState") as {
    currentChat: Ref<CurrentChat>;
    isLoading: Ref<boolean>;
    parsedUserDetails: Ref<UserDetails>;
    requestsRemaining: Ref<boolean>;
    screenWidth: Ref<number>;
    FREE_REQUEST_LIMIT: Ref<number>;
    currentMessages: Ref<Res[]>;
    isCollapsed: Ref<boolean>;
};

// Props
const props = defineProps<{
    showScrollDownButton: Ref<boolean>;
    scrollButtonPosition: Ref<string>;
    selectedContexts: Ref<ContextReference[]>;
    isRecording: boolean;
    isTranscribing: boolean;
    transcribedText: string;
    microphonePermission: "granted" | "denied" | "prompt";
    inputDisabled: boolean;
    inputPlaceholderText: string;
    pastePreview: {
        content: string;
        wordCount: number;
        charCount: number;
        show: boolean;
    } | null;
    showInputModeDropdown: boolean;
    showLimitExceededBanner: boolean;
    showUpgradeBanner: boolean;
    planStatus: {
        status: string;
        isExpired: boolean;
    };
}>();

// Emits
const emit = defineEmits<{
    submit: [
        e?: Event,
        retryPrompt?: string,
        contextReferences?: ContextReference[],
    ];
    autoGrow: [e: Event];
    handlePaste: [e: ClipboardEvent];
    keydown: [e: KeyboardEvent];
    toggleVoiceRecording: [];
    clearVoiceTranscription: [];
    toggleInputModeDropdown: [];
    selectInputMode: [mode: "web-search" | "deep-search" | "light-response"];
    navigateToUpgrade: [];
    scrollToBottom: [];
}>();

// Context References State
const showContextDropdown = ref(false);
const contextSearchQuery = ref("");
const filteredMessages = ref<ContextReference[]>([]);
const dropdownPosition = ref({ top: 0, left: 0 });
const activeContextIndex = ref(-1);
const textareaRef = ref<HTMLTextAreaElement | null>(null);
const lastAtPosition = ref(-1);
const userInputText = ref(""); // Store the actual user input

// Mode options
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

// Computed current mode
const currentMode = computed(() => {
    return modeOptions[
        parsedUserDetails.value.responseMode || "light-response"
    ];
});

// Get available messages for context
const availableMessages = computed(() => {
    const messages = currentMessages.value
        .filter((msg) => msg.prompt || msg.response)
        .map((msg, index) => {
            const text = msg.prompt || msg.response || "";
            return {
                id: `msg-${index}`,
                preview: text.slice(0, 100),
                fullText: text,
            };
        })
        .reverse();

    console.log("📚 Available messages:", messages.length);
    return messages;
});

// Handle textarea input to detect @ mentions
function handleTextareaInput(e: Event) {
    const textarea = e.target as HTMLTextAreaElement;
    const value = textarea.value;
    const cursorPos = textarea.selectionStart;

    // Update user input text
    userInputText.value = value;

    console.log("📝 Input event:", { value, cursorPos });

    // Find @ symbol before cursor
    const textBeforeCursor = value.substring(0, cursorPos);
    const lastAtIndex = textBeforeCursor.lastIndexOf("@");

    console.log("🎯 Last @ index:", lastAtIndex);

    if (lastAtIndex !== -1) {
        const textAfterAt = textBeforeCursor.substring(lastAtIndex + 1);
        console.log("✨ Text after @:", textAfterAt);

        // Check if we're still in the mention (no spaces after @)
        if (!textAfterAt.includes(" ") && !textAfterAt.includes("\n")) {
            console.log("✅ SHOWING DROPDOWN!");
            contextSearchQuery.value = textAfterAt;
            lastAtPosition.value = lastAtIndex;
            filterMessages(textAfterAt);
            calculateDropdownPosition(textarea, lastAtIndex);
            showContextDropdown.value = true;
            activeContextIndex.value = 0;
        } else {
            console.log("❌ Space found, closing");
            closeContextDropdown();
        }
    } else {
        console.log("❌ No @ found");
        closeContextDropdown();
    }

    emit("autoGrow", e);
}

// Filter messages based on search query
function filterMessages(query: string) {
    if (!query) {
        filteredMessages.value = availableMessages.value.slice(0, 5);
    } else {
        filteredMessages.value = availableMessages.value
            .filter((msg) =>
                msg.fullText.toLowerCase().includes(query.toLowerCase()),
            )
            .slice(0, 5);
    }
    console.log("🔍 Filtered messages:", filteredMessages.value.length);
}

// Calculate dropdown position relative to cursor
function calculateDropdownPosition(
    textarea: HTMLTextAreaElement,
    atIndex: number,
) {
    const textBeforeAt = textarea.value.substring(0, atIndex);
    const lines = textBeforeAt.split("\n");
    const currentLine = lines.length - 1;
    const lineHeight = 24;
    const leftOffset = 10;

    dropdownPosition.value = {
        top: (currentLine + 1) * lineHeight + 50,
        left: leftOffset,
    };
}

// Select a context reference
function selectContext(context: ContextReference) {
    const textarea = textareaRef.value;
    if (!textarea) return;

    const value = textarea.value;
    const cursorPos = textarea.selectionStart;
    const textBeforeCursor = value.substring(0, cursorPos);
    const lastAtIndex = textBeforeCursor.lastIndexOf("@");

    if (lastAtIndex !== -1) {
        // Remove the @query from textarea
        const textBefore = value.substring(0, lastAtIndex);
        const textAfter = value.substring(cursorPos);

        // Add context if not already added
        if (
            !props.selectedContexts.value.find(
                (c) => c.preview === context.preview,
            )
        ) {
            // eslint-disable-next-line vue/no-mutating-props
            props.selectedContexts.value.push(context);
        }

        // Update textarea value (remove @ mention)
        textarea.value = textBefore + textAfter;
        userInputText.value = textarea.value;

        // Set cursor position
        const newCursorPos = textBefore.length;
        textarea.setSelectionRange(newCursorPos, newCursorPos);

        closeContextDropdown();
        emit("autoGrow", { target: textarea } as any);

        // Focus back on textarea
        textarea.focus();
    }
}

// Close context dropdown
function closeContextDropdown() {
    showContextDropdown.value = false;
    contextSearchQuery.value = "";
    activeContextIndex.value = -1;
    lastAtPosition.value = -1;
}

// Handle keyboard navigation in dropdown
function handleContextNavigation(e: KeyboardEvent) {
    if (!showContextDropdown.value) return;

    switch (e.key) {
        case "ArrowDown":
            e.preventDefault();
            activeContextIndex.value = Math.min(
                activeContextIndex.value + 1,
                filteredMessages.value.length - 1,
            );
            break;
        case "ArrowUp":
            e.preventDefault();
            activeContextIndex.value = Math.max(
                activeContextIndex.value - 1,
                0,
            );
            break;
        case "Enter":
            if (
                activeContextIndex.value >= 0 &&
                filteredMessages.value[activeContextIndex.value]
            ) {
                e.preventDefault();
                selectContext(filteredMessages.value[activeContextIndex.value]);
            }
            break;
        case "Escape":
            e.preventDefault();
            closeContextDropdown();
            break;
    }
}

// Handle form submit with context
function handleSubmit(e?: Event) {
    e?.preventDefault();

    const textarea = textareaRef.value;
    const hasPrompt = textarea && textarea.value.trim();
    const hasContexts = props.selectedContexts.value.length > 0;

    // Don't submit if no prompt and no contexts
    if (!hasPrompt && !hasContexts) {
        return;
    }

    // Emit with context references only if they exist
    emit(
        "submit",
        e,
        undefined,
        hasContexts ? [...props.selectedContexts.value] : undefined,
    );

    // Clear selected contexts and user input after successful emit
    userInputText.value = "";
}

// Handle textarea auto-grow
function handleAutoGrow(e: Event) {
    handleTextareaInput(e);
}

// Handle paste
function handlePaste(e: ClipboardEvent) {
    emit("handlePaste", e);
}

// Handle keydown
function handleKeydown(e: KeyboardEvent) {
    // Handle context dropdown navigation first
    if (showContextDropdown.value) {
        handleContextNavigation(e);
        return;
    }
    emit("keydown", e);
}

// Toggle voice recording
function toggleVoice() {
    emit("toggleVoiceRecording");
}

// Clear transcription
function clearTranscription() {
    emit("clearVoiceTranscription");
}

// Toggle mode dropdown
function toggleDropdown() {
    emit("toggleInputModeDropdown");
}

// Select mode
function selectMode(mode: "web-search" | "deep-search" | "light-response") {
    emit("selectInputMode", mode);
}

// Navigate to upgrade
function navigateToUpgrade() {
    emit("navigateToUpgrade");
}

// Click outside handler
function handleClickOutside(e: MouseEvent) {
    const dropdown = document.querySelector(".context-dropdown");
    const textarea = textareaRef.value;

    if (
        dropdown &&
        !dropdown.contains(e.target as Node) &&
        (!textarea || !textarea.contains(e.target as Node))
    ) {
        closeContextDropdown();
    }
}

// Lifecycle
onMounted(() => {
    document.addEventListener("click", handleClickOutside);

    // Add delay to ensure DOM is ready
    setTimeout(() => {
        textareaRef.value = document.getElementById(
            "prompt",
        ) as HTMLTextAreaElement;
        console.log("✅ Textarea mounted:", !!textareaRef.value);
        console.log("📚 Messages available:", currentMessages.value.length);
    }, 100);
});

onBeforeUnmount(() => {
    document.removeEventListener("click", handleClickOutside);
});

// Watch for transcript changes
watch(
    () => props.transcribedText,
    () => {
        if (textareaRef.value) {
            emit("autoGrow", { target: textareaRef.value } as any);
        }
    },
);

// Watch for input disabled state
watch(
    () => props.inputDisabled,
    (disabled) => {
        if (disabled && showContextDropdown.value) {
            closeContextDropdown();
        }
    },
);

// Watch dropdown state
watch(showContextDropdown, (newVal) => {
    console.log("🎬 Dropdown state:", newVal);
});
</script>

<template>
    <div
        :style="
            screenWidth > 720
                ? isCollapsed
                    ? 'left:60px;'
                    : 'left:270px;'
                : 'left:0px;'
        "
        class="z-20 bottom-0 right-0 fixed"
        :class="pastePreview?.show ? 'pt-2' : ''"
    >
        <div
            class="flex flex-col gap-y-4 items-center justify-center px-2 sm:px-5"
        >
            <!-- Responsive Scroll to Bottom Button -->
            <button
                v-if="showScrollDownButton && currentMessages.length !== 0"
                @click="emit('scrollToBottom')"
                :class="[
                    'bg-gray-50  dark:bg-gray-800 text-gray-500 dark:text-gray-400 border dark:border-gray-700 px-4 h-8 rounded-full shadow-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors flex items-center justify-center gap-2',
                    scrollButtonPosition,
                ]"
                :disabled="isRecording"
                :title="
                    isRecording ? 'Recording in progress' : 'Scroll to bottom'
                "
            >
                <ArrowDown
                    class="w-4 h-4"
                    :class="{ 'animate-bounce': !isRecording }"
                />
                <span class="text-sm font-medium">Scroll Down</span>
            </button>

            <form
                @submit.prevent="handleSubmit"
                class="w-full md:max-w-3xl relative flex bg-gray-50 dark:bg-gray-800 flex-col border-2 dark:border-gray-700 shadow rounded-2xl items-center"
            >
                <!-- Paste Preview -->
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
                    <div class="flex items-center justify-center w-full">
                        <!-- Mobile: Stacked Layout -->
                        <div class="flex sm:hidden w-full flex-col gap-2">
                            <div class="flex items-center gap-2">
                                <div
                                    class="w-6 h-6 sm:w-8 sm:h-8 bg-red-100 dark:bg-red-900/30 rounded-full flex items-center justify-center flex-shrink-0"
                                >
                                    <X
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
                                type="button"
                                @click="navigateToUpgrade"
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
                        <div class="hidden sm:flex w-full items-center gap-3">
                            <div
                                class="w-8 h-8 bg-red-100 dark:bg-red-900/30 rounded-full flex items-center justify-center flex-shrink-0"
                            >
                                <X
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
                                            ? "Please renew your plan to continue."
                                            : `You've used all ${FREE_REQUEST_LIMIT} free requests.`
                                    }}
                                </p>
                            </div>
                            <button
                                type="button"
                                @click="navigateToUpgrade"
                                class="bg-red-500 px-3 hover:bg-red-600 dark:bg-red-600 dark:hover:bg-red-700 text-white py-2 rounded-md text-sm font-medium transition-colors flex-shrink-0 whitespace-nowrap"
                            >
                                {{ planStatus.isExpired ? "Renew" : "Upgrade" }}
                            </button>
                        </div>
                    </div>
                </div>

                <!-- Upgrade Warning Banner -->
                <div
                    v-if="showUpgradeBanner"
                    class="py-2 sm:py-3 w-full px-2 sm:px-3"
                >
                    <div class="flex items-center justify-center w-full">
                        <!-- Mobile -->
                        <div class="flex sm:hidden w-full flex-col gap-2">
                            <div class="flex items-center gap-2">
                                <div
                                    class="w-6 h-6 bg-yellow-100 dark:bg-yellow-900/30 rounded-full flex items-center justify-center flex-shrink-0"
                                >
                                    ⚠️
                                </div>
                                <div class="min-w-0 flex-1">
                                    <h3
                                        class="text-xs font-semibold text-yellow-800 dark:text-yellow-400 leading-tight"
                                    >
                                        {{ requestsRemaining }} requests left
                                    </h3>
                                    <p
                                        class="text-xs text-yellow-600 dark:text-yellow-400 leading-tight mt-0.5"
                                    >
                                        Upgrade for unlimited
                                    </p>
                                </div>
                            </div>
                            <button
                                type="button"
                                @click="navigateToUpgrade"
                                class="w-full bg-orange-500 hover:bg-orange-600 text-white py-2 rounded-md text-xs font-medium transition-colors"
                            >
                                Upgrade
                            </button>
                        </div>

                        <!-- Desktop -->
                        <div class="hidden sm:flex w-full items-center gap-3">
                            <div
                                class="w-8 h-8 bg-yellow-100 dark:bg-yellow-900/30 rounded-full flex items-center justify-center flex-shrink-0"
                            >
                                ⚠️
                            </div>
                            <div class="min-w-0 flex-1">
                                <h3
                                    class="text-sm font-semibold text-yellow-800 dark:text-yellow-400 mb-1"
                                >
                                    {{ requestsRemaining }} requests remaining
                                </h3>
                                <p
                                    class="text-xs text-yellow-600 dark:text-yellow-400"
                                >
                                    Upgrade to continue without limits
                                </p>
                            </div>
                            <button
                                type="button"
                                @click="navigateToUpgrade"
                                class="bg-orange-500 px-3 hover:bg-orange-600 text-white py-2 rounded-md text-sm font-medium transition-colors flex-shrink-0"
                            >
                                Upgrade
                            </button>
                        </div>
                    </div>
                </div>

                <!-- Input Area -->
                <div
                    class="flex flex-col w-full bg-white dark:bg-gray-900 rounded-2xl px-2 sm:px-3 py-2 gap-1 sm:gap-2 relative"
                    :class="
                        inputDisabled
                            ? 'opacity-50 border border-t dark:border-gray-700 pointer-events-none'
                            : showUpgradeBanner
                              ? 'border border-t dark:border-gray-700'
                              : ''
                    "
                >
                    <!-- Context Dropdown -->
                    <div
                        v-if="
                            showContextDropdown && filteredMessages.length > 0
                        "
                        class="context-dropdown absolute bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg shadow-2xl max-h-60 overflow-y-auto custom-scrollbar z-[9999] w-[20rem]"
                        :style="{
                            bottom: `${dropdownPosition.top}px`,
                            left: `${dropdownPosition.left}px`,
                        }"
                        @click.stop
                    >
                        <div
                            class="px-3 py-2 text-xs font-semibold text-gray-500 dark:text-gray-400 border-b border-gray-200 dark:border-gray-700 flex items-center gap-2 sticky top-0 bg-white dark:bg-gray-800 z-10"
                        >
                            <MessageSquare class="w-4 h-4" />
                            Reference Previous Messages ({{
                                filteredMessages.length
                            }})
                        </div>
                        <button
                            v-for="(message, index) in filteredMessages"
                            :key="message.preview"
                            type="button"
                            @click="selectContext(message)"
                            :class="[
                                'w-full px-3 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors',
                                activeContextIndex === index
                                    ? 'bg-blue-50 dark:bg-blue-900/30'
                                    : '',
                            ]"
                        >
                            <div class="flex items-start gap-2">
                                <MessageSquare
                                    class="w-4 h-4 mt-0.5 flex-shrink-0 text-gray-400 dark:text-gray-500"
                                />
                                <div class="flex-1 min-w-0">
                                    <p
                                        class="text-gray-800 dark:text-gray-200 truncate"
                                    >
                                        {{ message.preview }}
                                    </p>
                                    <p
                                        class="text-xs text-gray-500 dark:text-gray-400 mt-0.5"
                                    >
                                        {{ message.fullText.length }} characters
                                    </p>
                                </div>
                            </div>
                        </button>
                    </div>

                    <!-- Selected Context Badges (INSIDE TEXTAREA AREA) -->
                    <ReferenceBadge :selected-contexts="selectedContexts" />

                    <div class="w-full items-center justify-center flex">
                        <!-- Voice Recording Indicator -->
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

                        <!-- Clear Voice Button -->
                        <button
                            v-if="transcribedText && !isRecording"
                            type="button"
                            @click="clearTranscription"
                            class="rounded-lg w-8 h-8 sm:w-9 sm:h-9 flex items-center justify-center transition-colors text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 flex-shrink-0"
                            title="Clear voice transcription"
                        >
                            <X class="w-4 h-4 sm:w-5 sm:h-5" />
                        </button>

                        <!-- Textarea -->
                        <textarea
                            required
                            id="prompt"
                            name="prompt"
                            @keydown="handleKeydown"
                            @input="handleAutoGrow"
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
                            :placeholder="
                                currentChat && currentChat.messages?.length > 0
                                    ? `Use @ to include messages as context ${inputPlaceholderText.split('...')[1].trim()}`
                                    : inputPlaceholderText
                            "
                        />
                    </div>

                    <!-- Buttons Row -->
                    <div class="flex items-center justify-between w-full gap-2">
                        <!-- Left side buttons -->
                        <div class="flex items-center gap-2">
                            <!-- Microphone Button -->
                            <button
                                type="button"
                                @click="toggleVoice"
                                :disabled="inputDisabled"
                                :class="[
                                    'rounded-lg w-8 h-8 sm:w-9 sm:h-9 flex items-center justify-center transition-all duration-200 flex-shrink-0',
                                    isRecording
                                        ? 'bg-red-500 hover:bg-red-600 text-white shadow-lg transform scale-105 animate-pulse'
                                        : 'bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-600 dark:text-gray-300',
                                    'disabled:opacity-50 disabled:cursor-not-allowed',
                                ]"
                                :title="
                                    microphonePermission === 'denied'
                                        ? 'Microphone access denied'
                                        : isRecording
                                          ? 'Stop voice input'
                                          : 'Start voice input'
                                "
                            >
                                <Mic
                                    v-if="microphonePermission === 'prompt'"
                                    class="w-4 h-4 sm:w-5 sm:h-5"
                                />
                                <Mic
                                    v-else-if="
                                        !isRecording &&
                                        microphonePermission === 'granted'
                                    "
                                    class="w-4 h-4 sm:w-5 sm:h-5"
                                />
                                <Pause
                                    v-else-if="
                                        microphonePermission === 'granted' &&
                                        isRecording
                                    "
                                    class="w-4 h-4 sm:w-5 sm:h-5"
                                />
                                <MicOff
                                    v-else-if="
                                        microphonePermission === 'denied' &&
                                        !isRecording
                                    "
                                    class="w-4 h-4 sm:w-5 sm:h-5 text-red-500 dark:text-red-400"
                                />
                            </button>

                            <!-- Mode Dropdown -->
                            <div class="relative flex-shrink-0">
                                <button
                                    type="button"
                                    @click.stop="toggleDropdown"
                                    :disabled="inputDisabled"
                                    :class="[
                                        'rounded-lg w-8 h-8 sm:w-9 sm:h-9 flex items-center justify-center transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed shadow-sm border',
                                        parsedUserDetails?.responseMode ===
                                        'web-search'
                                            ? 'border-green-300 bg-green-50 hover:bg-green-100 dark:border-green-600 dark:bg-green-900/30 text-green-700 dark:text-green-300'
                                            : parsedUserDetails?.responseMode ===
                                                'deep-search'
                                              ? 'border-orange-300 bg-orange-50 hover:bg-orange-100 dark:border-orange-600 dark:bg-orange-900/30 text-orange-700 dark:text-orange-300'
                                              : 'border-blue-300 bg-blue-50 hover:bg-blue-100 dark:border-blue-600 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300',
                                    ]"
                                    :title="currentMode.title"
                                >
                                    <component
                                        :is="currentMode.icon"
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

                                    <button
                                        v-for="(option, key) in modeOptions"
                                        :key="key"
                                        type="button"
                                        @click="selectMode(option.mode)"
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
                                                    ? 'text-green-600 dark:text-green-400'
                                                    : 'text-gray-600 dark:text-gray-400',
                                            ]"
                                            :is="option.icon"
                                        />
                                        <div class="flex-1 min-w-0">
                                            <div class="font-semibold">
                                                {{ option.label }}
                                            </div>
                                            <div class="text-xs opacity-70">
                                                {{ option.description }}
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

                        <!-- Submit Button -->
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
        <div class="bg-white dark:bg-gray-900 h-3 sm:h-5 w-full"></div>
    </div>
</template>
