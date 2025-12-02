<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { Eye, Copy, Check, ChevronDown, ChevronUp } from "lucide-vue-next";
import { Button } from "../button";
import { useRoute } from "vue-router";

interface Props {
    data: {
        id: string;
        language: string;
        code: string;
        highlighted: string;
    };
    isPreviewable?: boolean;
    maxLines?: number; // Maximum lines to show before collapsing
}

const props = withDefaults(defineProps<Props>(), {
    maxLines: 15,
});

const emit = defineEmits<{
    preview: [];
}>();

const route = useRoute();
const copied = ref(false);
const isExpanded = ref(false);
const shouldShowToggle = ref(false);
const codeLines = ref(0);

// Calculate if code is long enough to need collapse
onMounted(() => {
    const lines = props.data.code.split("\n").length;
    codeLines.value = lines;
    shouldShowToggle.value = lines > props.maxLines;
});

const maxHeight = computed(() => {
    if (!shouldShowToggle.value || isExpanded.value) {
        return "none";
    }
    // Approximate: 1.5rem per line (24px)
    return `${props.maxLines * 1.5}rem`;
});

const copyToClipboard = async () => {
    try {
        await navigator.clipboard.writeText(props.data.code);
        copied.value = true;
        setTimeout(() => {
            copied.value = false;
        }, 2000);
    } catch (err) {
        console.error("Failed to copy:", err);
    }
};

const handlePreview = () => {
    emit("preview");
};

const toggleExpand = () => {
    isExpanded.value = !isExpanded.value;
};
</script>

<template>
    <div
        class="relative my-4 rounded-lg overflow-hidden bg-gray-800 dark:bg-gray-800 border border-gray-200 dark:border-gray-700"
    >
        <!-- Header Bar -->
        <div class="flex items-center justify-between px-4 pt-2 bg-none">
            <span
                class="text-xs font-mono text-gray-300 dark:text-gray-400 capitalize"
            >
                {{ data.language }}
                <span v-if="shouldShowToggle" class="text-gray-500 ml-2">
                    ({{ codeLines }} lines)
                </span>
            </span>
            <div
                class="flex items-center gap-2"
                v-if="route.path.startsWith('/chat')"
            >
                <!-- Preview Button (only for HTML) -->
                <Button
                    v-if="isPreviewable"
                    @click="handlePreview"
                    title="Preview HTML"
                    class="h-7 px-3 bg-white hover:bg-gray-100 text-gray-900 text-xs rounded inline-flex items-center gap-1.5 transition-colors"
                >
                    <Eye :size="14" />
                    <span>Preview</span>
                </Button>
                <!-- Copy Button -->
                <Button
                    @click="copyToClipboard"
                    :title="copied ? 'Copied!' : 'Copy code'"
                    class="h-7 px-3 bg-gray-700 dark:bg-gray-700 text-gray-100 dark:text-gray-100 border border-gray-600 dark:border-gray-600 hover:bg-gray-600 dark:hover:bg-gray-600 text-xs rounded inline-flex items-center gap-1.5 transition-colors"
                >
                    <Check v-if="copied" :size="14" class="text-green-400" />
                    <Copy v-else :size="14" />
                    <span>{{ copied ? "Copied!" : "Copy" }}</span>
                </Button>
            </div>
        </div>

        <!-- Code Block Container with Fade Effect -->
        <div class="relative">
            <!-- Code Block -->
            <div
                :class="[
                    'code-container transition-all duration-300 ease-in-out',
                    !isExpanded && shouldShowToggle ? 'code-collapsed' : '',
                ]"
                :style="{
                    maxHeight: maxHeight,
                }"
            >
                <pre
                    class="!m-0 !rounded-none bg-gray-800 dark:bg-gray-800 overflow-x-auto custom-scrollbar"
                >
                    <code
                        :class="`hljs language-${data.language} text-sm leading-relaxed break-words whitespace-pre-wrap`"
                        v-html="data.highlighted"
                    ></code>
                </pre>
            </div>

            <!-- Fade Overlay (only when collapsed) -->
            <div
                v-if="!isExpanded && shouldShowToggle"
                class="absolute bottom-0 left-0 right-0 h-24 bg-gradient-to-t from-gray-800 to-transparent pointer-events-none"
            ></div>
        </div>

        <!-- Show More/Less Button -->
        <div v-if="shouldShowToggle" class="bg-gray-800">
            <Button
                @click="toggleExpand"
                class="w-full h-9 bg-transparent hover:bg-gray-800/50 dark:bg-transparent dark:hover:bg-gray-800/50 text-gray-300 dark:text-gray-400 hover:text-gray-100 text-xs font-medium rounded-none inline-flex items-center justify-center gap-2 transition-colors"
            >
                <component
                    :is="isExpanded ? ChevronUp : ChevronDown"
                    :size="16"
                />
                <span>{{ isExpanded ? "Show Less" : "Show More" }}</span>
            </Button>
        </div>
    </div>
</template>

<style scoped>
/* Code container with smooth transitions */
.code-container {
    position: relative;
    overflow: hidden;
}

.code-collapsed {
    overflow: hidden;
}

/* Ensure code block styling */
pre {
    padding: 0 1rem;
    margin: 0;
}

pre code {
    display: block;
    overflow-x: auto;
}

/* Override any conflicting hljs styles */
:deep(.hljs) {
    background: transparent !important;
    padding: 0 !important;
}
</style>
