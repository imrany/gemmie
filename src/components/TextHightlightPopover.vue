<script setup lang="ts">
import { ref, onMounted, onUnmounted, inject, type Ref, h } from "vue";
import { toast } from "vue-sonner";
import {
    Popover,
    PopoverContent,
    PopoverTrigger,
} from "@/components/ui/popover";

// Extend Window interface for TypeScript
declare global {
    interface Window {
        stopCurrentSpeech?: () => void;
    }
}

const { isOpenTextHighlightPopover } = inject("globalState") as {
    isOpenTextHighlightPopover: Ref<boolean>;
};
const selectedText = ref("");
const triggerElement = ref<HTMLElement | null>(null);
const triggerStyle = ref({ top: "0px", left: "0px" });
const actualSide = ref("top");
const currentUtterance = ref<SpeechSynthesisUtterance | null>(null);
const toastId = ref<string | number | null>(null);
const isSpeaking = ref(false);

function handleMouseUp(e: MouseEvent) {
    const selection = window.getSelection();
    const text = selection?.toString().trim();

    if (text && text.length > 0) {
        selectedText.value = text;

        const range = selection?.getRangeAt(0);
        const rect = range?.getBoundingClientRect();

        if (rect && triggerElement.value) {
            triggerStyle.value = {
                top: rect.top + window.scrollY - 10 + "px",
                left: rect.left + window.scrollX + rect.width / 2 + "px",
            };

            isOpenTextHighlightPopover.value = true;
            // ✅ Prevent default behavior after showing popover
            e.preventDefault();
        }
    }
}

function handleContextMenu(e: MouseEvent) {
    const selection = window.getSelection();
    const hasSelection = (selection?.toString().trim().length ?? 0) > 0;

    if (hasSelection) {
        e.preventDefault(); // Prevent default context menu
        e.stopPropagation(); // Stop event from bubbling
        return false;
    }
}

function copyText() {
    navigator.clipboard.writeText(selectedText.value);
    toast.success("Copied to clipboard!");
    isOpenTextHighlightPopover.value = false;
    window.getSelection()?.removeAllRanges();
}

function stopSpeaking() {
    window.speechSynthesis.cancel();
    isSpeaking.value = false;
    currentUtterance.value = null;

    if (toastId.value) {
        toast.dismiss(toastId.value);
        toastId.value = null;
    }
}

function speakText() {
    // Stop any ongoing speech
    if (window.speechSynthesis.speaking) {
        stopSpeaking();
        return;
    }

    // Check if speech synthesis is supported
    if (!window.speechSynthesis) {
        toast.error("Text-to-speech is not supported in your browser");
        isOpenTextHighlightPopover.value = false;
        window.getSelection()?.removeAllRanges();
        return;
    }

    const utterance = new SpeechSynthesisUtterance(selectedText.value);
    utterance.rate = 0.8;
    utterance.pitch = 1;
    utterance.volume = 1;
    currentUtterance.value = utterance;

    // Create the toast component
    const SpeechToast = () =>
        h(
            "div",
            {
                class: "flex items-center gap-3 bg-white dark:bg-slate-900 border border-gray-200 dark:border-slate-700 rounded-lg shadow-lg px-4 py-3 min-w-[280px]",
            },
            [
                h(
                    "div",
                    {
                        class: "flex items-center justify-center w-8 h-8 bg-blue-500 rounded-full",
                    },
                    [
                        h(
                            "svg",
                            {
                                class: "w-4 h-4 text-white",
                                fill: "currentColor",
                                viewBox: "0 0 20 20",
                            },
                            [
                                h("path", {
                                    d: "M18 10a8 8 0 11-16 0 8 8 0 0116 0zM7 8a1 1 0 012 0v4a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v4a1 1 0 102 0V8a1 1 0 00-1-1z",
                                }),
                            ],
                        ),
                    ],
                ),
                h("div", { class: "flex-1" }, [
                    h(
                        "p",
                        {
                            class: "text-sm font-medium text-gray-900 dark:text-gray-100",
                        },
                        "Playing audio",
                    ),
                    h(
                        "p",
                        {
                            class: "text-xs text-gray-500 dark:text-gray-400 mt-0.5 line-clamp-1",
                        },
                        `${selectedText.value.substring(0, 40)}${selectedText.value.length > 40 ? "..." : ""}`,
                    ),
                ]),
                h(
                    "button",
                    {
                        class: "flex items-center justify-center w-8 h-8 text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-slate-800 rounded-full transition-colors",
                        onClick: stopSpeaking,
                    },
                    [
                        h(
                            "svg",
                            {
                                class: "w-4 h-4",
                                fill: "none",
                                stroke: "currentColor",
                                viewBox: "0 0 24 24",
                            },
                            [
                                h("path", {
                                    "stroke-linecap": "round",
                                    "stroke-linejoin": "round",
                                    "stroke-width": "2",
                                    d: "M6 18L18 6M6 6l12 12",
                                }),
                            ],
                        ),
                    ],
                ),
            ],
        );

    // Show toast with media player styling
    toastId.value = toast.custom(SpeechToast, {
        duration: Infinity,
        position: "top-center",
        unstyled: true,
    });

    // Handle speech events
    utterance.onstart = () => {
        isSpeaking.value = true;
    };

    utterance.onend = () => {
        isSpeaking.value = false;
        currentUtterance.value = null;
        if (toastId.value) {
            toast.dismiss(toastId.value);
            toastId.value = null;
        }
        console.log("Finished playing");
    };

    utterance.onerror = (event) => {
        isSpeaking.value = false;
        currentUtterance.value = null;
        if (toastId.value) {
            toast.dismiss(toastId.value);
            toastId.value = null;
        }
        console.error("Speech synthesis error:", event);
    };

    window.speechSynthesis.speak(utterance);
    isOpenTextHighlightPopover.value = false;
    window.getSelection()?.removeAllRanges();
}

function translateText() {
    window.open(
        `https://translate.google.com/?text=${encodeURIComponent(selectedText.value)}`,
        "_blank",
    );
    isOpenTextHighlightPopover.value = false;
    window.getSelection()?.removeAllRanges();
}

// Update Listen button text based on speaking state
const listenButtonText = ref("Listen");

import { watch } from "vue";
watch(isSpeaking, (newValue) => {
    listenButtonText.value = newValue ? "Stop" : "Listen";
});

onMounted(() => {
    document.removeEventListener("mouseup", handleMouseUp);
    document.removeEventListener("contextmenu", handleContextMenu, {
        capture: true,
    });
    document.removeEventListener("selectstart", () => {});

    if (window.speechSynthesis.speaking) {
        stopSpeaking();
    }
});

onUnmounted(() => {
    document.removeEventListener("mouseup", handleMouseUp);
    document.removeEventListener("contextmenu", handleContextMenu);
    // Clean up speech synthesis
    if (window.speechSynthesis.speaking) {
        stopSpeaking();
    }
});
</script>

<template>
    <div>
        <Popover v-model:open="isOpenTextHighlightPopover">
            <PopoverTrigger as-child>
                <div
                    ref="triggerElement"
                    class="fixed w-0 h-0 pointer-events-none"
                    :style="{
                        top: triggerStyle.top,
                        left: triggerStyle.left,
                        transform: 'translateX(-50%)',
                        zIndex: 50,
                    }"
                />
            </PopoverTrigger>
            <PopoverContent
                side="top"
                :avoid-collisions="true"
                :collision-padding="10"
                align="center"
                class="w-fit h-fit rounded-md shadow-md transition-all duration-200 bg-white dark:bg-slate-900 border border-gray-300 dark:border-slate-600 p-0 relative"
                :side-offset="6"
                @escape-key-down="isOpenTextHighlightPopover = false"
                @interact-outside="isOpenTextHighlightPopover = false"
            >
                <!-- Pointer -->
                <div
                    class="absolute w-2 h-2 bg-white dark:bg-slate-900 border-b border-r border-gray-300 dark:border-slate-600 rotate-45"
                    :class="{
                        '-bottom-1 left-1/2 transform -translate-x-1/2':
                            actualSide === 'top',
                        '-top-1 left-1/2 transform -translate-x-1/2':
                            actualSide === 'bottom',
                    }"
                ></div>

                <div class="flex items-center relative z-10 rounded-md">
                    <button
                        @click="copyText"
                        class="px-3 py-2 text-sm text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-slate-800 transition-colors duration-200 rounded-l-md"
                    >
                        Copy
                    </button>

                    <button
                        @click="speakText"
                        :disabled="!selectedText"
                        class="px-3 py-2 text-sm text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-slate-800 transition-colors duration-200 border-l border-gray-200 dark:border-slate-700 disabled:opacity-50 disabled:cursor-not-allowed"
                        :class="{
                            'text-red-600 dark:text-red-400 hover:text-red-700 dark:hover:text-red-300':
                                isSpeaking,
                        }"
                    >
                        {{ listenButtonText }}
                    </button>

                    <button
                        @click="translateText"
                        class="px-3 py-2 text-sm text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-slate-800 transition-colors duration-200 rounded-r-md border-l border-gray-200 dark:border-slate-700"
                    >
                        Translate
                    </button>
                </div>
            </PopoverContent>
        </Popover>
    </div>
</template>

<style scoped>
/* Prevent browser's native text selection tooltip/menu */
::selection {
    background-color: rgba(59, 130, 246, 0.3); /* Blue highlight */
}

:deep(*) {
    -webkit-user-select: text;
    user-select: text;
    -webkit-touch-callout: none; /* Disable iOS callout */
    -webkit-tap-highlight-color: transparent; /* Disable tap highlight */
}
</style>
