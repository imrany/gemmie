<script lang="ts" setup>
import {
    copyPasteContent,
    getHighlightLanguage,
} from "@/utils/previewPasteContent";
import hljs from "highlight.js";
import { ChevronLeft, Clipboard, Copy, X } from "lucide-vue-next";

defineProps<{
    data: {
        showPasteModal: boolean;
        currentPasteContent: {
            content: string;
            wordCount: number;
            charCount: number;
            type: "text" | "code" | "json" | "markdown" | "xml" | "html";
        } | null;
    };
    closePasteModal: () => void;
}>();
</script>
<template>
    <div
        v-if="data.showPasteModal"
        class="fixed inset-0 z-50 flex text-sm items-center justify-center p-4 bg-black bg-opacity-50 dark:bg-opacity-80"
        @click.self="closePasteModal"
    >
        <!-- Desktop Layout -->
        <div
            class="hidden w-full max-w-6xl h-[80vh] rounded-lg shadow-2xl overflow-hidden lg:flex bg-white dark:bg-gray-800"
        >
            <!-- Content Area -->
            <div class="flex-1 flex flex-col">
                <!-- Header -->
                <div
                    class="flex items-center justify-between px-6 py-4 border-b border-gray-200 bg-gray-100 dark:bg-gray-700 dark:border-gray-600"
                >
                    <div class="flex items-center gap-3">
                        <Clipboard
                            class="w-4 h-4 text-gray-600 dark:text-gray-400"
                        />
                        <div>
                            <h3
                                class="text-lg font-semibold text-gray-800 dark:text-gray-100"
                            >
                                Pasted
                            </h3>
                            <p class="text-sm text-gray-600 dark:text-gray-300">
                                {{ data.currentPasteContent?.wordCount }} words
                                •
                                {{ data.currentPasteContent?.charCount }}
                                characters
                            </p>
                        </div>
                    </div>
                    <div class="flex items-center gap-2">
                        <span
                            class="capitalize rounded bg-blue-100 px-2 py-1 text-xs font-medium text-blue-800"
                        >
                            {{ data.currentPasteContent?.type }}
                        </span>
                        <button
                            @click="
                                copyPasteContent(
                                    data.currentPasteContent?.content || '',
                                )
                            "
                            class="rounded px-3 py-1 text-sm text-gray-700 transition-colors bg-gray-200 hover:bg-gray-300 dark:bg-gray-600 dark:hover:bg-gray-500 dark:text-gray-200"
                        >
                            <Copy class="w-4 h-4 mr-1" />
                            <p>Copy</p>
                        </button>
                        <button
                            @click="closePasteModal"
                            class="flex items-center justify-center w-8 h-8 rounded-full transition-colors hover:bg-gray-200 dark:hover:bg-gray-600"
                        >
                            <X
                                class="w-4 h-4 text-gray-500 dark:text-gray-400"
                            />
                        </button>
                    </div>
                </div>

                <!-- Content -->
                <div class="flex-1 overflow-auto">
                    <!-- Code/JSON/XML content with syntax highlighting -->
                    <div
                        v-if="data.currentPasteContent?.type !== 'text'"
                        class="h-full bg-gray-900 dark:bg-inherit"
                    >
                        <pre class="h-full">
              <code
                :class="`hljs language-${getHighlightLanguage(
                  data.currentPasteContent?.type || 'text'
                )}`"
                v-html="
                  data.currentPasteContent?.type === 'json'
                    ? hljs.highlight(
                        JSON.stringify(
                          JSON.parse(
                            data.currentPasteContent?.content || '{}'
                          ),
                          null,
                          2
                        ),
                        { language: 'json' }
                      ).value
                    : hljs.highlight(
                        data.currentPasteContent?.content || '',
                        {
                          language: getHighlightLanguage(
                            data.currentPasteContent?.type || 'text'
                          ),
                        }
                      ).value
                "
              ></code>
            </pre>
                    </div>

                    <!-- Plain text content -->
                    <div v-else class="h-full p-6 bg-white dark:bg-gray-800">
                        <div
                            class="prose prose-sm max-w-none dark:prose-invert"
                        >
                            <pre
                                class="break-words whitespace-pre-wrap font-mono leading-relaxed text-sm text-gray-800 dark:text-gray-200"
                                >{{ data.currentPasteContent?.content }}</pre
                            >
                        </div>
                    </div>
                </div>
            </div>
            <!-- Close Button (Desktop) -->
            <button
                @click="closePasteModal"
                class="absolute top-2 right-2 w-8 h-8 rounded-full transition-colors hover:bg-gray-200 dark:hover:bg-gray-600 lg:flex items-center justify-center hidden"
            >
                <X class="w-4 h-4 text-gray-500 dark:text-gray-400" />
            </button>
        </div>

        <!-- Mobile Layout -->
        <div
            class="fixed inset-0 flex flex-col bg-white lg:hidden dark:bg-gray-800"
        >
            <!-- Mobile Header -->
            <div
                class="flex flex-shrink-0 items-center justify-between px-4 py-3 border-b border-gray-200 bg-gray-100 dark:bg-gray-700 dark:border-gray-600"
            >
                <div class="flex items-center gap-2">
                    <button
                        @click="closePasteModal"
                        class="flex items-center justify-center w-8 h-8 rounded-full transition-colors hover:bg-gray-200 dark:hover:bg-gray-600"
                    >
                        <ChevronLeft
                            class="w-5 h-5 text-gray-600 dark:text-gray-400"
                        />
                    </button>
                    <div>
                        <h3
                            class="font-semibold text-gray-800 dark:text-gray-100"
                        >
                            Pasted Content
                        </h3>
                        <p class="text-xs text-gray-600 dark:text-gray-300">
                            {{ data.currentPasteContent?.wordCount }} words •
                            {{ data.currentPasteContent?.charCount }} chars
                        </p>
                    </div>
                </div>
                <div class="flex items-center gap-2">
                    <span
                        class="capitalize rounded bg-blue-100 px-2 py-1 text-xs font-medium text-blue-800"
                    >
                        {{ data.currentPasteContent?.type }}
                    </span>
                    <button
                        @click="
                            copyPasteContent(
                                data.currentPasteContent?.content || '',
                            )
                        "
                        class="flex items-center justify-center w-8 h-8 rounded-full transition-colors hover:bg-gray-200 dark:hover:bg-gray-600"
                    >
                        <Copy
                            class="w-4 h-4 text-gray-600 dark:text-gray-400"
                        />
                    </button>
                </div>
            </div>

            <!-- Mobile Content -->
            <div class="flex-1 overflow-auto">
                <!-- Code/JSON/XML content with syntax highlighting -->
                <div
                    v-if="data.currentPasteContent?.type !== 'text'"
                    class="min-h-full bg-gray-900 dark:bg-inherit"
                >
                    <pre class="text-xs">
            <code
              :class="`hljs language-${getHighlightLanguage(
                data.currentPasteContent?.type || 'text'
              )}`"
              v-html="
                data.currentPasteContent?.type === 'json'
                  ? hljs.highlight(
                      JSON.stringify(
                        JSON.parse(
                          data.currentPasteContent?.content || '{}'
                        ),
                        null,
                        2
                      ),
                      { language: 'json' }
                    ).value
                  : hljs.highlight(
                      data.currentPasteContent?.content || '',
                      {
                        language: getHighlightLanguage(
                          data.currentPasteContent?.type || 'text'
                        ),
                      }
                    ).value
              "
            ></code>
          </pre>
                </div>

                <!-- Plain text content -->
                <div v-else class="min-h-full bg-white dark:bg-gray-800">
                    <pre
                        class="break-words whitespace-pre-wrap font-mono leading-relaxed text-sm text-gray-800 dark:text-gray-200"
                        >{{ data.currentPasteContent?.content }}</pre
                    >
                </div>
            </div>
        </div>
    </div>
</template>
