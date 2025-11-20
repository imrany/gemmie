<script setup lang="ts">
import { ref } from "vue";
import { Eye, Copy, Check } from "lucide-vue-next";
import { Button } from "../button";

interface Props {
    data: {
        id: string;
        language: string;
        code: string;
        highlighted: string;
    };
    isPreviewable?: boolean;
}

const props = defineProps<Props>();
const emit = defineEmits<{
    preview: [];
}>();

const copied = ref(false);

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
</script>

<template>
    <div
        class="relative my-4 rounded-lg overflow-hidden bg-gray-800 dark:bg-gray-900 border border-gray-200 dark:border-gray-700"
    >
        <!-- Header Bar -->
        <div class="flex items-center justify-between px-4 pt-2 bg-none">
            <span
                class="text-xs font-mono text-gray-300 dark:text-gray-400 capitalize"
            >
                {{ data.language }}
            </span>
            <div class="flex items-center gap-2">
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

        <!-- Code Block -->
        <pre
            class="!m-0 !rounded-none bg-gray-800 dark:bg-gray-900 overflow-x-auto custom-scrollbar"
        >
            <code
                :class="`hljs language-${data.language} text-sm leading-relaxed`"
                v-html="data.highlighted"
            ></code>
        </pre>
    </div>
</template>

<style scoped>
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
