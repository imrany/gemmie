<script setup lang="ts">
import { detectContentType } from "@/lib/previewPasteContent";
import hljs from "highlight.js/lib/common";
import { computed } from "vue";
import { inject } from "vue";

const { openPreview, removePastePreview } = inject("globalState") as {
    openPreview: (
        code: string,
        language: string,
        content: string,
        metadata?: {
            fileSize: string;
            wordCount: number;
            charCount: number;
        },
    ) => void;
    removePastePreview: () => void;
};

const { content, isClickable } = defineProps<{
    content: string;
    isClickable?: boolean;
}>();

const wordCount = computed(() => {
    return content?.trim().split("#pastedText#")[1]?.split(/\s+/)?.length || 0;
});
const text = computed(() => {
    return content?.trim()?.split("#pastedText#")[1]?.trim() || "";
});
const charCount = computed(() => {
    return content?.trim()?.split("#pastedText#")[1]?.length || 0;
});

const preview =
    charCount.value > 200 ? text.value.substring(0, 200) + "..." : text.value;

// Generate unique ID for this component
const componentId = `paste-${Math.random().toString(36).substring(2, 9)}`;
const clickableClass = isClickable
    ? "cursor-pointer transition-colors duration-200"
    : "";

const language = computed(() => {
    let contentType:
        | "plaintext"
        | "code"
        | "json"
        | "markdown"
        | "xml"
        | "html" = "plaintext";

    if (typeof detectContentType === "function") {
        contentType = detectContentType(text.value);
    } else {
        // Simple content type detection as fallback
        if (
            text.value.trim().startsWith("{") &&
            text.value.trim().endsWith("}")
        ) {
            contentType = "json";
        } else if (
            text.value.includes("```") ||
            text.value.includes("function") ||
            text.value.includes("class")
        ) {
            contentType = "code";
        } else if (text.value.includes("#") || text.value.includes("**")) {
            contentType = "markdown";
        } else if (text.value.includes("<") && text.value.includes(">")) {
            contentType = "html";
        }
    }
    return contentType;
});

const fileSize = (text: string): string => {
    return `${(new Blob([text]).size / 1024).toFixed(2)} KB`;
};
</script>
<template>
    <div
        v-if="content && (wordCount > 100 || wordCount > 800)"
        :class="[
            'paste-preview border border-gray-300 dark:border-gray-600 rounded-lg overflow-hidden my-2 bg-gray-100 dark:bg-gray-900 hover:shadow-md transition-all duration-300 w-[250px]',
            clickableClass,
        ]"
        :id="componentId"
        @click="
            if (isClickable)
                openPreview(text, language, content, {
                    wordCount,
                    charCount,
                    fileSize: fileSize(text),
                });
        "
    >
        <div class="w-full">
            <div
                class="bg-gray-600 dark:bg-gray-800 px-3 py-1 text-white dark:text-gray-200 text-xs font-medium flex items-center gap-2 transition-colors duration-200"
            >
                <i class="pi pi-clipboard text-gray-300 dark:text-gray-400"></i>
                <span>PASTED</span>
                <span class="ml-auto text-gray-200 dark:text-gray-400"
                    >{{ charCount }} chars</span
                >
                <ExternalLink
                    v-if="isClickable"
                    class="w-4 h-4 ml-1 text-gray-300 dark:text-gray-500"
                />
            </div>
            <div class="pb-3 px-3">
                <div class="relative">
                    <pre
                        class="text-sm h-20 sm:h-24 overflow-hidden text-gray-800 dark:text-gray-200 transition-colors duration-200"
                    >
                        <code
                            :class="`language-${language} leading-relaxed break-words whitespace-pre-wrap `"
                            v-html="preview ? hljs.highlight(preview, { language }).value : 'No paste content available'"
                        ></code>
                    </pre>
                    <div
                        class="absolute bottom-0 left-0 right-0 h-14 bg-gradient-to-b from-transparent to-gray-100 dark:to-gray-900 pointer-events-none transition-colors duration-200"
                    ></div>
                </div>
                <div
                    class="flex items-center justify-between mt-2 text-xs text-gray-600 dark:text-gray-400 transition-colors duration-200"
                >
                    <span class="hidden sm:inline"
                        >{{
                            isClickable
                                ? "Click to view full content"
                                : "Large content detected"
                        }}
                    </span>
                    <span class="sm:hidden">{{
                        isClickable ? "Tap to view" : "Large content"
                    }}</span>
                    <button
                        v-if="!isClickable"
                        @click="removePastePreview"
                        class="text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-gray-100 underline font-medium transition-colors duration-200"
                        type="button"
                    >
                        Remove
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>
