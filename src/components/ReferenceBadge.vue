<script setup lang="ts">
import type { ContextReference } from "@/types";
import type { Ref } from "vue";
import { unref, inject } from "vue";
import { Link, X } from "lucide-vue-next";

const { handleSrollIntoView } = inject("globalState") as {
    handleSrollIntoView: (id: string) => void;
};
const props = withDefaults(
    defineProps<{
        selectedContexts: Ref<ContextReference[]>;
        isCloseable?: boolean;
    }>(),
    {
        isCloseable: true,
    },
);

// Remove context reference
function removeContext(contextPreview: string) {
    // eslint-disable-next-line vue/no-mutating-props
    props.selectedContexts.value = props.selectedContexts.value.filter(
        (c) => c.preview !== contextPreview,
    );
}

function truncateText(text: string, maxLength: number): string {
    if (text.length <= maxLength) return text;
    return text.substring(0, maxLength) + "...";
}
</script>

<template>
    <div
        v-if="unref(props.selectedContexts).length > 0"
        :class="['flex gap-1.5 flex-wrap px-2 pt-2']"
    >
        <div
            v-for="(context, index) in unref(props.selectedContexts)"
            :key="context.preview"
            @click="handleSrollIntoView(`chat-${context.fullText}`)"
            :class="[
                'inline-flex cursor-pointer items-center gap-1.5 px-2 py-1 rounded-md text-xs font-medium border transition-colors',
                'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300',
                'dark:border-gray-600 border-gray-200',
                isCloseable
                    ? 'hover:bg-gray-100 dark:hover:bg-gray-700/30 group'
                    : '',
            ]"
            :title="context.preview"
        >
            <!-- Reference number badge (for vertical layout) or Link icon (for horizontal) -->
            <component
                :is="!isCloseable ? 'span' : Link"
                :class="
                    !isCloseable
                        ? 'flex items-center justify-center w-4 h-4 bg-blue-600 dark:bg-blue-500 text-white rounded-full text-[10px] font-bold flex-shrink-0'
                        : 'w-3 h-3 flex-shrink-0'
                "
            >
                {{ !isCloseable ? index + 1 : "" }}
            </component>

            <!-- Preview text - more compact for vertical layout -->
            <span :class="['truncate', 'max-w-[120px] sm:max-w-[150px]']">
                {{ truncateText(context.preview, 30) }}
            </span>

            <!-- Remove button (only if closeable) -->
            <button
                v-if="isCloseable"
                type="button"
                @click.stop="removeContext(context.preview)"
                class="hover:bg-blue-200 dark:hover:bg-blue-800 rounded-full p-0.5 transition-colors opacity-70 group-hover:opacity-100 flex-shrink-0"
                :title="`Remove reference ${index + 1}`"
            >
                <X class="w-3 h-3" />
            </button>
        </div>
    </div>
</template>
