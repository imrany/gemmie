<script setup lang="ts">
import type { Ref } from "vue";
import { inject } from "vue";
import { Button } from "@/components/ui/button";

const { currentChatId, screenWidth } = inject("globalState") as {
    currentChatId: Ref<string>;
    screenWidth: Ref<number>;
};
defineProps<{
    fallbackChatId: string;
    showErrorSection: Ref<boolean>;
}>();
</script>

<template>
    <div
        :class="
            screenWidth > 720
                ? 'h-screen bg-inherit flex flex-col items-center justify-center w-[85%]'
                : 'bg-inherit h-screen w-full flex flex-col items-center justify-center'
        "
    >
        <div class="flex flex-col items-center gap-3 max-w-md text-center">
            <p
                class="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-gray-100"
            >
                Can't open this chat
            </p>
            <p
                class="max-sm:max-w-xs sm:text-sm text-xs text-gray-600 dark:text-gray-400"
            >
                It may have been deleted or you might not have permission to
                view it.
            </p>
            <div class="flex gap-3 mt-4">
                <Button
                    @click="
                        () => {
                            const fallbackId = fallbackChatId;
                            if (fallbackId) {
                                // eslint-disable-next-line vue/no-mutating-props
                                showErrorSection.value = false;
                                currentChatId = fallbackId;
                                $router.push(`/chat/${fallbackId}`);
                            } else {
                                $router.push('/new');
                            }
                        }
                    "
                    class="px-5 py-2 bg-gray-900 dark:bg-white text-white dark:text-black text-sm rounded-lg transition-colors inline-flex items-center gap-2 shadow-lg hover:bg-gray-800 dark:hover:bg-gray-100"
                >
                    {{
                        fallbackChatId
                            ? "Go to Recent Chat"
                            : "Start a New Chat"
                    }}
                </Button>
            </div>
        </div>
    </div>
</template>
