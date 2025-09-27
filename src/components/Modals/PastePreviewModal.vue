<script lang="ts" setup>
import {  copyPasteContent, getHighlightLanguage } from '@/utils/previewPasteContent';
import hljs from 'highlight.js';

defineProps<{
    data :{
        showPasteModal: boolean,
        currentPasteContent:{
          content: string;
          wordCount: number;
          charCount: number;
          type: 'text' | 'code' | 'json' | 'markdown' | 'xml' | 'html';
          } | null,
    }
  closePasteModal: ()=>void,
}>()

</script>
<template>
    <div v-if="data.showPasteModal" class="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4"
        @click.self="closePasteModal">

        <!-- Desktop Layout -->
        <div class="hidden lg:flex w-full max-w-6xl h-[80vh] bg-white rounded-lg shadow-2xl overflow-hidden">
            <!-- Content Area -->
            <div class="flex-1 flex flex-col">
                <!-- Header -->
                <div class="bg-gray-100 px-6 py-4 border-b border-gray-200 flex items-center justify-between">
                    <div class="flex items-center gap-3">
                        <i class="pi pi-clipboard text-gray-600"></i>
                        <div>
                            <h3 class="text-lg font-semibold text-gray-800">Pasted Content</h3>
                            <p class="text-sm text-gray-600">
                                {{ data.currentPasteContent?.wordCount }} words • {{ data.currentPasteContent?.charCount }}
                                characters
                            </p>
                        </div>
                    </div>
                    <div class="flex items-center gap-2">
                        <span class="px-2 py-1 bg-blue-100 text-blue-800 text-xs font-medium rounded capitalize">
                            {{ data.currentPasteContent?.type }}
                        </span>
                        <button @click="copyPasteContent(data.currentPasteContent?.content || '')"
                            class="px-3 py-1 bg-gray-200 hover:bg-gray-300 text-gray-700 text-sm rounded transition-colors">
                            <i class="pi pi-copy mr-1"></i>Copy
                        </button>
                        <button @click="closePasteModal"
                            class="w-8 h-8 rounded-full hover:bg-gray-200 flex items-center justify-center transition-colors">
                            <i class="pi pi-times text-gray-500"></i>
                        </button>
                    </div>
                </div>

                <!-- Content -->
                <div class="flex-1 overflow-hidden">
                    <!-- Code/JSON/XML content with syntax highlighting -->
                    <div v-if="data.currentPasteContent?.type !== 'text'" class="h-full overflow-auto bg-gray-900">
                        <pre
                            class="h-full"><code :class="`hljs language-${getHighlightLanguage(data.currentPasteContent?.type || 'text')}`" 
                                    v-html="data.currentPasteContent?.type === 'json' ?
                                        hljs.highlight(JSON.stringify(JSON.parse(data.currentPasteContent?.content || '{}'), null, 2), { language: 'json' }).value :
                                        hljs.highlight(data.currentPasteContent?.content || '', { language: getHighlightLanguage(data.currentPasteContent?.type || 'text') }).value"></code></pre>
                    </div>

                    <!-- Plain text content -->
                    <div v-else class="h-full p-6 overflow-auto bg-white">
                        <div class="prose prose-sm max-w-none">
                            <pre
                                class="whitespace-pre-wrap break-words text-sm font-mono leading-relaxed text-gray-800">{{ data.currentPasteContent?.content }}</pre>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Mobile Layout -->
        <div class="lg:hidden fixed inset-0 bg-white flex flex-col">
            <!-- Mobile Header -->
            <div class="bg-gray-100 px-4 py-3 border-b border-gray-200 flex items-center justify-between flex-shrink-0">
                <div class="flex items-center gap-2">
                    <button @click="closePasteModal"
                        class="w-8 h-8 rounded-full hover:bg-gray-200 flex items-center justify-center transition-colors">
                        <i class="pi pi-arrow-left text-gray-600"></i>
                    </button>
                    <div>
                        <h3 class="font-semibold text-gray-800">Pasted Content</h3>
                        <p class="text-xs text-gray-600">
                            {{ data.currentPasteContent?.wordCount }} words • {{ data.currentPasteContent?.charCount }} chars
                        </p>
                    </div>
                </div>
                <div class="flex items-center gap-2">
                    <span class="px-2 py-1 bg-blue-100 text-blue-800 text-xs font-medium rounded capitalize">
                        {{ data.currentPasteContent?.type }}
                    </span>
                    <button @click="copyPasteContent(data.currentPasteContent?.content || '')"
                        class="w-8 h-8 rounded-full hover:bg-gray-200 flex items-center justify-center transition-colors">
                        <i class="pi pi-copy text-gray-600"></i>
                    </button>
                </div>
            </div>

            <!-- Mobile Content -->
            <div class="flex-1 overflow-hidden">
                <!-- Code/JSON/XML content with syntax highlighting -->
                <div v-if="data.currentPasteContent?.type !== 'text'" class="h-full overflow-auto bg-gray-900">
                    <pre
                        class="text-xs p-4"><code :class="`hljs language-${getHighlightLanguage(data.currentPasteContent?.type || 'text')}`"
                                      v-html="data.currentPasteContent?.type === 'json' ?
                                        hljs.highlight(JSON.stringify(JSON.parse(data.currentPasteContent?.content || '{}'), null, 2), { language: 'json' }).value :
                                        hljs.highlight(data.currentPasteContent?.content || '', { language: getHighlightLanguage(data.currentPasteContent?.type || 'text') }).value"></code></pre>
                </div>

                <!-- Plain text content -->
                <div v-else class="h-full p-4 overflow-auto bg-white">
                    <pre
                        class="whitespace-pre-wrap break-words text-sm font-mono leading-relaxed text-gray-800">{{ data.currentPasteContent?.content }}</pre>
                </div>
            </div>
        </div>
    </div>
</template>