<script setup lang="ts">
import hljs from "highlight.js/lib/common";

const { content, wordCount, charCount, isClickable } = defineProps<{
    content: string;
    wordCount: number;
    charCount: number;
    isClickable?: boolean;
}>();

const preview =
    content.length > 200 ? content.substring(0, 200) + "..." : content;
// Proper HTML escaping
const escapedPreview = preview;
// .replace(/&/g, "&amp;")
// .replace(/</g, "&lt;")
// .replace(/>/g, "&gt;")
// .replace(/"/g, "&quot;")
// .replace(/'/g, "&#39;")
// .replace(/\n/g, "<br>")
// .replace(/\t/g, "&nbsp;&nbsp;&nbsp;&nbsp;");

// Generate unique ID for this component
const componentId = `paste-${Math.random().toString(36).substring(2, 9)}`;
const clickableClass = isClickable
    ? "paste-preview-clickable cursor-pointer transition-colors duration-200"
    : "";
</script>
<template>
    <div
        :class="[
            'paste-preview border border-gray-300 dark:border-gray-600 rounded-lg overflow-hidden my-2 bg-gray-100 dark:bg-gray-900 hover:shadow-md transition-all duration-300 w-[250px]',
            clickableClass,
        ]"
        :id="componentId"
        :data-paste-content="isClickable ? encodeURIComponent(content) : ''"
        :data-word-count="isClickable ? wordCount : 0"
        :data-char-count="isClickable ? charCount : 0"
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
                            :class="`language-plaintext leading-relaxed break-words whitespace-pre-wrap `"
                            v-html="escapedPreview ? hljs.highlight(escapedPreview, { language: 'plaintext' }).value : 'No paste content available'"
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
                        class="remove-paste-preview text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-gray-100 underline font-medium transition-colors duration-200"
                        type="button"
                    >
                        Remove
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>
