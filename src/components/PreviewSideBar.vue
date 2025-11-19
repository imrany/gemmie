<script setup lang="ts">
import { ref, watch, onMounted } from "vue";
import { X, Code, Eye, Check, Copy } from "lucide-vue-next";
import { Button } from "./ui/button";
import type { Ref } from "vue";
import { inject } from "vue";
import hljs from "highlight.js/lib/common";

const {
    screenWidth,
    showPreviewSidebar,
    previewCode,
    previewLanguage,
    closePreview,
} = inject("globalState") as {
    screenWidth: Ref<number>;
    showPreviewSidebar: Ref<boolean>;
    previewCode: Ref<string>;
    previewLanguage: Ref<string>;
    closePreview: () => void;
};

const activeTab = ref<"preview" | "code">("preview");
const copied = ref(false);
const sidebarWidth = ref<number>(window.innerWidth * 0.4); // Initial width, can be adjusted
const minWidth = 300; // Minimum width
const maxWidth = window.innerWidth * 0.8; // Maximum width
const isResizing = ref(false);

// Reset to preview tab when sidebar opens
watch(showPreviewSidebar, (newVal) => {
    if (newVal) {
        activeTab.value = "preview";
    }
});

const copyToClipboard = async () => {
    try {
        await navigator.clipboard.writeText(previewCode.value);
        copied.value = true;
        setTimeout(() => {
            copied.value = false;
        }, 2000);
    } catch (err) {
        console.error("Failed to copy:", err);
    }
};

const startResize = (e: MouseEvent) => {
    console.log("startResize: " + e);
    isResizing.value = true;
    document.addEventListener("mousemove", resize);
    document.addEventListener("mouseup", stopResize);
};

const resize = (e: MouseEvent) => {
    if (!isResizing.value) return;
    let newWidth = window.innerWidth - e.clientX;
    if (newWidth > minWidth && newWidth < maxWidth) {
        sidebarWidth.value = newWidth;
    }
};

const stopResize = () => {
    isResizing.value = false;
    document.removeEventListener("mousemove", resize);
    document.removeEventListener("mouseup", stopResize);
};

onMounted(() => {
    sidebarWidth.value = window.innerWidth * 0.4;
});
</script>

<template>
    <!-- Backdrop for mobile -->
    <Transition name="backdrop" @after-leave="activeTab = 'preview'">
        <div
            v-if="showPreviewSidebar"
            class="fixed inset-0 bg-black/50 z-40 md:hidden"
            @click="closePreview"
        />
    </Transition>
    <div
        v-if="showPreviewSidebar"
        class="max-md:hidden group w-2 relative h-full cursor-col-resize -mr-1 z-30 grid place-items-center hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
        @mousedown="startResize"
    >
        <div
            class="absolute top-0 bottom-0 right-1 w-[0.5px] bg-gray-300 dark:bg-gray-600 transition-all group-hover:bg-gray-400 dark:group-hover:bg-gray-500 group-hover:w-px group-hover:translate-x-[0.5px]"
        ></div>
        <div
            class="h-6 w-2 relative rounded-full border-[0.5px] border-gray-400 dark:border-gray-500 bg-gray-200 dark:bg-gray-700 shadow transition duration-200 group-hover:border-gray-700 dark:group-hover:border-gray-400 cursor-col-resize"
        ></div>
    </div>
    <!-- Sidebar -->
    <Transition name="slide">
        <div
            v-if="showPreviewSidebar"
            class="fixed top-0 right-0 bottom-0 md:relative md:z-20 w-full max-w-full z-50 flex-shrink-0 md:flex md:items-stretch md:justify-stretch"
            :style="{
                width: screenWidth > 720 ? sidebarWidth + 'px' : '100vw',
            }"
        >
            <!-- Sidebar Content -->
            <div
                class="relative flex flex-col h-full w-full bg-gray-100 dark:bg-gray-900 overflow-hidden shadow-2xl md:shadow-none md:border-l md:border-gray-200 md:dark:border-gray-800"
            >
                <!-- Header -->
                <div
                    class="flex-shrink-0 border-b border-gray-200 dark:border-gray-800"
                >
                    <div class="flex items-center justify-between px-4 py-3">
                        <!-- Tabs -->
                        <div
                            class="flex items-center gap-1.5 bg-gray-200 dark:bg-gray-800 p-1 rounded-lg"
                        >
                            <button
                                @click="activeTab = 'preview'"
                                :class="[
                                    'px-3 py-1.5 text-xs rounded-md transition-all duration-200 inline-flex items-center gap-1.5 font-medium',
                                    activeTab === 'preview'
                                        ? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-white shadow-sm'
                                        : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200',
                                ]"
                            >
                                <Eye :size="14" />
                                <span>Preview</span>
                            </button>
                            <button
                                @click="activeTab = 'code'"
                                :class="[
                                    'px-3 py-1.5 text-xs rounded-md transition-all duration-200 inline-flex items-center gap-1.5 font-medium',
                                    activeTab === 'code'
                                        ? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-white shadow-sm'
                                        : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200',
                                ]"
                            >
                                <Code :size="14" />
                                <span>Code</span>
                            </button>
                        </div>

                        <!-- Close Button -->
                        <Button
                            size="sm"
                            variant="ghost"
                            class="h-8 w-8 p-0 hover:bg-gray-200 dark:hover:bg-gray-800"
                            @click="closePreview"
                            title="Close preview"
                        >
                            <X class="w-4 h-4" />
                        </Button>
                    </div>
                </div>

                <!-- Content Area -->
                <div class="flex-1 overflow-hidden relative">
                    <!-- Preview Tab -->
                    <Transition name="fade" mode="out-in">
                        <div
                            v-if="activeTab === 'preview'"
                            class="absolute inset-0 bg-gray-100 dark:bg-gray-800"
                            key="preview"
                        >
                            <iframe
                                v-if="previewCode"
                                :srcdoc="previewCode"
                                class="w-full h-full border-0"
                                sandbox="allow-scripts allow-forms allow-modals allow-popups allow-same-origin"
                                title="HTML Preview"
                            />
                            <div
                                v-else
                                class="flex items-center justify-center h-full text-gray-500 dark:text-gray-400"
                            >
                                <div class="text-center">
                                    <Eye
                                        class="w-12 h-12 mx-auto mb-4 opacity-30"
                                    />
                                    <p class="text-sm font-medium">
                                        No preview available
                                    </p>
                                </div>
                            </div>
                        </div>

                        <!-- Code Tab -->
                        <div
                            v-else
                            class="absolute inset-0 overflow-auto w-full bg-gray-800 custom-scrollbar"
                            key="code"
                        >
                            <div class="relative min-h-full w-full">
                                <!-- Copy Button -->
                                <div
                                    class="sticky top-0 z-10 flex justify-end p-3"
                                >
                                    <button
                                        @click="copyToClipboard"
                                        class="inline-flex items-center gap-2 bg-gray-900 dark:bg-gray-200 hover:bg-gray-700 text-white dark:text-gray-800 px-3 py-1.5 rounded-md text-xs font-medium transition-all duration-200 shadow-lg hover:shadow-xl"
                                    >
                                        <Check
                                            v-if="copied"
                                            :size="14"
                                            class="text-green-400 dark:text-green-600"
                                        />
                                        <Copy v-else :size="14" />
                                        <span>{{
                                            copied ? "Copied!" : "Copy code"
                                        }}</span>
                                    </button>
                                </div>

                                <!-- Code Display -->
                                <pre class="px-4 pb-4 text-sm"><code
                                    :class="`language-${previewLanguage} leading-relaxed text-gray-300 `"
                                    v-html="previewCode ? hljs.highlight(previewCode, { language: previewLanguage }).value : 'No code available'"
                                ></code></pre>
                            </div>
                        </div>
                    </Transition>
                </div>

                <!-- Footer -->
                <div
                    class="flex-shrink-0 border-t border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900 px-4 py-2.5"
                >
                    <div class="flex items-center justify-between text-xs">
                        <div class="flex items-center gap-2">
                            <div
                                class="w-1.5 h-1.5 rounded-full transition-all duration-300"
                                :class="
                                    activeTab === 'preview'
                                        ? 'bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.6)]'
                                        : 'bg-blue-500 shadow-[0_0_8px_rgba(59,130,246,0.6)]'
                                "
                            ></div>
                            <span
                                class="font-medium text-gray-700 dark:text-gray-300"
                            >
                                {{
                                    activeTab === "preview"
                                        ? "Live Preview"
                                        : "Source Code"
                                }}
                            </span>
                        </div>
                        <span
                            class="text-gray-500 dark:text-gray-500 hidden sm:inline"
                        >
                            {{
                                activeTab === "preview"
                                    ? "Interactive demo"
                                    : `${previewCode.split("\n").length} lines`
                            }}
                        </span>
                    </div>
                </div>
            </div>
        </div>
    </Transition>
</template>

<style scoped>
/* Slide transition for sidebar */
.slide-enter-active,
.slide-leave-active {
    transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.slide-enter-from {
    transform: translateX(100%);
}

.slide-leave-to {
    transform: translateX(-100%);
}

@media (min-width: 768px) {
    .slide-enter-from,
    .slide-leave-to {
        transform: translateX(0);
        opacity: 0;
    }

    .slide-enter-active,
    .slide-leave-active {
        transition: opacity 0.4s ease-in-out;
    }
}

/* Backdrop transition */
.backdrop-enter-active,
.backdrop-leave-active {
    transition: opacity 0.3s ease;
}

.backdrop-enter-from,
.backdrop-leave-to {
    opacity: 0;
}

/* Fade transition for tab content */
.fade-enter-active,
.fade-leave-active {
    transition: opacity 0.15s ease;
}

.fade-enter-from,
.fade-leave-to {
    opacity: 0;
}
</style>
