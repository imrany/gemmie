<script lang="ts" setup>
import type { Chat } from "@/types";
import { AlignJustify } from "lucide-vue-next";
import type { Ref } from "vue";
import { inject } from "vue";
import { Button } from "./ui/button";

const {
    hideSidebar,
    screenWidth,
    isCollapsed,
    currentChat,
    isLoading,
    shareChat,
} = inject("globalState") as {
    shareChat: (chatId: string) => Promise<void>;
    isLoading: Ref<boolean>;
    isCollapsed: Ref<boolean>;
    currentChat: Ref<Chat | undefined>;

    screenWidth: Ref<number>;
    hideSidebar: () => void;
};
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
                <!-- :disabled="isLoading || !currentChat" -->
                <Button
                    @click="shareChat(currentChat?.id || '')"
                    variant="outline"
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
</template>
