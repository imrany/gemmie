<script lang="ts" setup>
import type { Chat } from "@/types";
import { AlignJustify } from "lucide-vue-next";
import type { Ref } from "vue";
import { inject, ref } from "vue";
import { Button } from "@/components/ui/button";
import ShareChatDialog from "./ShareChatDialog.vue";
import { toast } from "vue-sonner/src/packages/state.js";

const {
    hideSidebar,
    screenWidth,
    currentChat,
    isLoading,
    updateChat,
    shareResponse,
    copyResponse,
} = inject("globalState") as {
    isLoading: Ref<boolean>;
    copyResponse: (text: string, copiedIndex?: number) => void;
    shareResponse: (title: string, body: string) => void;
    currentChat: Ref<Chat | undefined>;
    updateChat: (
        chatId: string,
        update: Record<string, any>,
    ) => Promise<Chat | undefined>;
    screenWidth: Ref<number>;
    hideSidebar: () => void;
};

const isShowShareChat = ref(false);
const isShareLoading = ref(false);

async function handleShareChat(isPrivate: boolean) {
    isShareLoading.value = true;
    if (!currentChat.value) {
        toast.error("Invalid chat ID");
        isShareLoading.value = false;
        return;
    }

    try {
        const chat = await updateChat(currentChat.value.id, {
            is_private: isPrivate,
        });

        if (!chat?.is_private) {
            const shareLink = `${window.location.origin}/chat/${chat?.id || ""}`;
            shareResponse(chat?.title || "", shareLink);
        }
    } catch (error) {
        toast.error("Failed to share chat");
    } finally {
        isShareLoading.value = false;
    }
}

function handleCopyLink(link: string) {
    copyResponse(link);
}

function formatTitle(title: string, maxLength: number = 30): string {
    if (title.length <= maxLength) return title;

    if (title.includes("#pastedText#")) {
        return title.split("#pastedText#")[1].slice(0, maxLength) + "...";
    }
    return title.slice(0, maxLength) + "...";
}
</script>

<template>
    <div
        class="bg-white dark:bg-gray-900 h-[52px] w-full z-30 absolute top-0 right-0 transition-all duration-300 ease-in-out"
    >
        <div
            class="flex h-full px-3 sm:px-4 items-center justify-between w-full"
        >
            <!-- Title/Brand -->
            <div class="flex-1 min-w-0 mr-3">
                <p
                    v-if="currentChat && screenWidth > 720"
                    class="text-gray-600 dark:text-gray-400 font-medium truncate text-sm select-none"
                    :title="currentChat.title"
                >
                    {{ formatTitle(currentChat.title, 40) }}
                </p>
                <p
                    v-else-if="screenWidth <= 720"
                    class="text-gray-700 dark:text-gray-300 text-xl font-semibold tracking-wide select-none"
                >
                    Gemmie
                </p>
            </div>

            <!-- Actions -->
            <div class="flex gap-2 sm:gap-3 items-center flex-shrink-0">
                <!-- Share Button -->
                <Button
                    v-if="
                        currentChat &&
                        currentChat.messages &&
                        currentChat.messages.length > 0 &&
                        !currentChat.is_read_only
                    "
                    @click="isShowShareChat = true"
                    variant="outline"
                    :disabled="isLoading || !currentChat"
                    size="sm"
                    class="text-xs hover:border-gray-300 dark:hover:border-gray-600 bg-white dark:bg-gray-900 border-gray-200 dark:border-gray-700"
                >
                    Share
                </Button>

                <!-- Read Only Badge -->
                <div
                    v-if="currentChat && currentChat.is_read_only"
                    class="text-xs text-gray-600 dark:text-gray-400 bg-gray-100 dark:bg-gray-800 rounded-md px-2 sm:px-3 py-1 whitespace-nowrap"
                >
                    Read Only
                </div>

                <!-- Mobile Sidebar Toggle -->
                <button
                    v-if="screenWidth < 720"
                    @click="hideSidebar"
                    title="Toggle Sidebar"
                    class="w-9 h-9 flex items-center justify-center hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg cursor-pointer transition-colors"
                >
                    <AlignJustify
                        class="w-5 h-5 text-gray-700 dark:text-gray-300"
                    />
                </button>
            </div>
        </div>
    </div>

    <!-- Share Dialog -->
    <ShareChatDialog
        :show="isShowShareChat"
        :current-chat="currentChat"
        :is-loading="isShareLoading"
        @close="isShowShareChat = false"
        @share="handleShareChat"
        @copy="handleCopyLink"
    />
</template>
