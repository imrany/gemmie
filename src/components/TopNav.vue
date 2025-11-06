<script lang="ts" setup>
import type { Chat } from "@/types";
import { AlignJustify, Check, Copy, Globe, Lock } from "lucide-vue-next";
import type { Ref } from "vue";
import { inject, watch } from "vue";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import DialogBox from "./Dialog/DialogBox.vue";
import { ref } from "vue";
import { toast } from "vue-sonner/src/packages/state.js";

const {
    hideSidebar,
    screenWidth,
    isCollapsed,
    currentChat,
    isLoading,
    updateChat,
    shareResponse,
    copyResponse,
} = inject("globalState") as {
    isLoading: Ref<boolean>;
    copyResponse: (text: string, copiedIndex?: number) => void;
    shareResponse: (title: string, body: string) => void;
    isCollapsed: Ref<boolean>;
    currentChat: Ref<Chat | undefined>;
    updateChat: (
        chatId: string,
        update: Record<string, any>,
    ) => Promise<Chat | undefined>;

    screenWidth: Ref<number>;
    hideSidebar: () => void;
};

const isShowShareChat = ref(false);
const shareLink = ref<string | null>(null);
const isCopied = ref(false);

async function shareChat(is_private: boolean) {
    if (!currentChat.value) {
        toast.error("Invalid chat ID");
        return;
    }

    try {
        const chat = await updateChat(currentChat.value.id, { is_private });

        if (!chat?.is_private) {
            shareLink.value = `${window.location.origin}/chat/${chat?.id || ""}`;
            shareResponse(chat?.title || "", shareLink.value);
        }
    } catch (error) {
        toast.error("Failed to share chat");
    }
}

function copyLink() {
    copyResponse(shareLink?.value?.toString() || "");
    isCopied.value = true;
    window.setTimeout(() => {
        isCopied.value = false;
    }, 2000);
}

watch(
    () => currentChat.value?.is_private,
    (newValue) => {
        if (!newValue) {
            shareLink.value = `${window.location.origin}/chat/${currentChat.value?.id || ""}`;
        }
    },
    { immediate: true, deep: true },
);
</script>

<template>
    <div
        class="bg-white dark:bg-gray-900 h-[52px] z-30 fixed top-0 right-0 transition-all duration-300 ease-in-out"
        :style="
            screenWidth > 720 && !isCollapsed
                ? 'left:270px'
                : screenWidth > 720 && isCollapsed
                  ? 'left:60px'
                  : 'left:0'
        "
    >
        <div class="flex h-full px-4 items-center justify-between w-full">
            <!-- Brand -->
            <p
                v-if="currentChat && screenWidth > 720"
                class="text-gray-600 dark:text-gray-400 font-medium truncate text-sm select-none"
            >
                <span v-if="currentChat.title.length > 30">{{
                    currentChat.title.includes("#pastedText#")
                        ? currentChat.title
                              .split("#pastedText#")[1]
                              .slice(0, 30) + "..."
                        : currentChat.title.slice(0, 30) + "..."
                }}</span>
                <span v-else>{{ currentChat.title }}</span>
            </p>
            <p
                v-else-if="screenWidth <= 720"
                class="text-gray-700 dark:text-gray-300 text-xl max-md:text-2xl font-semibold tracking-wide select-none"
            >
                Gemmie
            </p>

            <div class="flex gap-3 items-center ml-auto">
                <Button
                    v-if="
                        currentChat &&
                        currentChat.messages &&
                        currentChat.messages.length > 0
                    "
                    @click="isShowShareChat = true"
                    variant="outline"
                    :disabled="isLoading || !currentChat"
                    class="hover:border-white text-xs hover:dark:border-gray-900 bg-white dark:bg-gray-900 border-gray-200 dark:border-gray-700"
                    :loading="isLoading"
                >
                    Share
                </Button>

                <!-- Mobile Sidebar Toggle -->
                <div
                    v-if="screenWidth < 720"
                    class="flex gap-2 items-center ml-auto"
                >
                    <button
                        @click="hideSidebar"
                        title="Toggle Sidebar"
                        class="w-9 h-9 flex items-center justify-center text-lg hover:bg-gray-100 dark:hover:bg-gray-800 rounded-full cursor-pointer transition-colors"
                    >
                        <AlignJustify
                            class="w-5 h-5 text-gray-700 dark:text-gray-300"
                        />
                    </button>
                </div>
            </div>
        </div>
    </div>
    <DialogBox
        name="Share Chat"
        :show="isShowShareChat"
        :closeModal="() => (isShowShareChat = false)"
    >
        <div class="flex flex-col gap-5">
            <div>
                <p class="text-2xl font-bold">
                    {{ currentChat?.is_private ? "Share chat" : "Chat shared" }}
                </p>
                <p class="text-gray-700 dark:text-gray-300">
                    {{
                        currentChat?.is_private
                            ? "Share this chat with your friends!"
                            : "Anyone who has this link will be able to view this."
                    }}
                </p>
            </div>
            <div class="rounded-lg border border-gray-300 dark:border-gray-700">
                <button
                    @click="shareChat(true)"
                    class="flex w-full gap-4 py-3 px-6 items-center border-b border-gray-300 dark:border-gray-700"
                >
                    <Lock class="w-5 h-5" />
                    <div class="flex items-start flex-col gap-1">
                        <p
                            class="font-semibold text-gray-700 dark:text-gray-300"
                        >
                            Private
                        </p>
                        <p>Only you can see this chat.</p>
                    </div>
                    <Check
                        class="ml-auto w-5 h-5 text-green-500"
                        v-if="currentChat?.is_private"
                    />
                </button>
                <button
                    @click="shareChat(false)"
                    class="flex w-full gap-4 py-3 px-6 items-center"
                >
                    <Globe class="w-5 h-5" />
                    <div class="flex items-start flex-col gap-1">
                        <p
                            class="font-semibold text-gray-700 dark:text-gray-300"
                        >
                            Public
                        </p>
                        <p>Anyone can see this chat.</p>
                    </div>
                    <Check
                        class="ml-auto w-5 h-5 text-green-500"
                        v-if="!currentChat?.is_private"
                    />
                </button>
            </div>
            <div
                v-if="!currentChat?.is_private"
                class="flex gap-2 my-3 items-center"
            >
                <Input
                    type="text"
                    readonly
                    :defaultValue="shareLink || ''"
                    class="border border-gray-300 bg-inherit dark:border-gray-700 rounded-md px-3 py-2 w-full"
                />
                <Button type="button" @click="copyLink" size="sm" class="px-3">
                    <span class="sr-only">Copy</span>
                    <Copy class="w-4 h-4" v-if="!isCopied" />
                    <Check class="w-4 h-4" v-else />
                </Button>
            </div>
            <div class="flex flex-col gap-4 w-full items-center" v-else>
                <p class="text-gray-600 dark:text-gray-400 text-xs">
                    Don’t share personal information or third-party content
                    without permission.
                </p>
                <Button
                    type="button"
                    @click="shareChat(false)"
                    size="sm"
                    class="px-3 ml-auto"
                >
                    <span class="">Create share link</span>
                </Button>
            </div>
        </div>
    </DialogBox>
</template>
