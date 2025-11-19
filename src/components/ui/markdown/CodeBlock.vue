<!-- CodeBlock.vue -->
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
    <div class="relative my-4">
        <!-- Code Block -->
        <pre class="dark:border-gray-700 border">
            <code
                :class="`hljs language-${data.language} text-sm leading-relaxed`"
                v-html="data.highlighted"
            ></code>
        </pre>
        <div class="absolute top-2 right-2 left-2">
            <div class="flex w-full items-center justify-between">
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
                        class="px-3 py-1.5 h-[29px] bg-gray-900 dark:bg-white text-white dark:text-black text-xs rounded transition-colors inline-flex items-center gap-1.5 hover:bg-gray-800 dark:hover:bg-gray-100"
                        title="Preview HTML"
                    >
                        <Eye :size="14" />
                        <span>Preview</span>
                    </Button>

                    <!-- Copy Button -->
                    <Button
                        @click="copyToClipboard"
                        class="h-[29px] items-center gap-1.5 bg-gray-700 dark:bg-gray-700 hover:bg-gray-600 dark:hover:bg-gray-600 text-white dark:text-white px-3 py-1.5 rounded text-xs font-medium transition-colors"
                        :title="copied ? 'Copied!' : 'Copy code'"
                    >
                        <Check
                            v-if="copied"
                            :size="14"
                            class="text-green-400"
                        />
                        <Copy v-else :size="14" />
                        <span>{{ copied ? "Copied!" : "Copy" }}</span>
                    </Button>
                </div>
            </div>
        </div>
    </div>
</template>
