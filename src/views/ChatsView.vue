<script lang="ts" setup>
import { inject, ref, computed } from "vue";
import { useRouter } from "vue-router";
import type { Ref } from "vue";
import type { Chat, UserDetails } from "@/types";
import {
    ChevronLeft,
    ChevronRight,
    Clock,
    MessageCircle,
    Plus,
    RefreshCw,
    Search,
    X,
} from "lucide-vue-next";
import OverallLayout from "@/layout/OverallLayout.vue";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

const {
    isOnline,
    chats,
    currentChatId,
    parsedUserDetails,
    screenWidth,
    isCollapsed,
} = inject("globalState") as {
    isOnline: Ref<boolean>;
    chats: Ref<Chat[]>;
    currentChatId: Ref<string>;
    parsedUserDetails: Ref<UserDetails>;
    screenWidth: Ref<number>;
    isCollapsed: Ref<boolean>;
};

const router = useRouter();
const searchQuery = ref("");

// Filter chats based on search
const filteredChats = computed(() => {
    if (!searchQuery.value.trim()) {
        return chats.value;
    }

    const query = searchQuery.value.toLowerCase();
    return chats.value.filter(
        (chat) =>
            chat.title?.toLowerCase().includes(query) ||
            chat.messages?.some(
                (message) =>
                    message.prompt?.toLowerCase().includes(query) ||
                    message.response?.toLowerCase().includes(query),
            ),
    );
});

const handleNewChat = () => {
    currentChatId.value = "";
};

const handleBack = () => {
    router.back();
};

const clearSearch = () => {
    searchQuery.value = "";
};

const handleGoToChat = (id: string) => {
    currentChatId.value = id;
};
</script>

<template>
    <OverallLayout>
        <!-- Main Content - Centered -->
        <div
            :class="[
                'flex-grow flex flex-col transition-all duration-300 ease-in-out h-screen overflow-hidden min-w-0',
                screenWidth > 720 && parsedUserDetails?.username
                    ? isCollapsed
                        ? 'ml-[60px]'
                        : 'ml-[270px]'
                    : '',
            ]"
        >
            <!-- Centered Container -->
            <div
                class="flex-1 flex overflow-hidden h-full justify-center p-3 sm:p-4 md:p-6 max-w-full"
            >
                <div
                    class="w-full px-2 flex flex-col h-full overflow-hidden md:max-w-3xl max-w-[100vw]"
                >
                    <!-- Top Header -->
                    <div
                        v-if="chats.length !== 0"
                        class="flex-shrink-0 w-full items-center justify-between mb-6 flex gap-2 min-w-0"
                    >
                        <!-- Back Button (Mobile Only) -->
                        <button
                            @click="handleBack"
                            class="md:hidden flex items-center justify-center w-10 h-10 rounded-lg bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors flex-shrink-0"
                            aria-label="Go back"
                        >
                            <ChevronLeft
                                class="w-5 h-5 text-gray-700 dark:text-gray-300"
                            />
                        </button>

                        <h1
                            class="text-2xl md:text-3xl font-bold text-gray-800 dark:text-white"
                        >
                            Chats
                        </h1>

                        <Button
                            @click="handleNewChat"
                            :disabled="!isOnline"
                            :class="[
                                'px-3 py-2 scale-90 dark:bg-gray-200 text-gray-200 bg-gray-700 text-sm dark:text-gray-700 hover:bg-gray-600 dark:hover:bg-gray-600 rounded-lg transition-colors flex items-center gap-1.5 shadow-lg hover:shadow-xl flex-shrink-0 whitespace-nowrap',
                                !isOnline ? 'cursor-not-allowed' : '',
                            ]"
                        >
                            <Plus class="w-5 h-5" />
                            <span class="hidden sm:inline"> New Chat </span>
                            <span class="sm:hidden">New</span>
                        </Button>
                    </div>

                    <!-- Search Bar -->
                    <div v-if="chats.length !== 0" class="mb-4 flex-shrink-0">
                        <div class="relative w-full">
                            <Search
                                class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400"
                            />
                            <Input
                                v-model="searchQuery"
                                placeholder="Search chats..."
                                class="pl-9 pr-9 h-10 font-medium text-gray-800 dark:text-gray-200 w-full resize-none border-none ring-[1px] ring-gray-200 dark:ring-gray-800 outline-none focus:border-none focus-visible:ring-gray-300 dark:focus-visible:ring-gray-700"
                            />
                            <button
                                v-if="searchQuery"
                                @click="clearSearch"
                                class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
                            >
                                <X class="w-4 h-4" />
                            </button>
                        </div>
                        <p
                            class="text-start text-xs sm:text-sm font-medium text-gray-500 dark:text-gray-400 mt-2"
                        >
                            {{ filteredChats.length }}
                            {{ filteredChats.length === 1 ? "chat" : "chats" }}
                        </p>
                    </div>

                    <!-- Empty State -->
                    <div
                        v-if="filteredChats.length === 0"
                        class="h-[80vh] flex items-center justify-center"
                    >
                        <div class="text-center">
                            <h3
                                class="text-base font-medium text-gray-900 dark:text-white mb-2"
                            >
                                {{
                                    searchQuery
                                        ? "No matching chats found"
                                        : "No chats yet"
                                }}
                            </h3>
                            <p
                                class="text-sm text-gray-500 dark:text-gray-400 mb-6 max-w-md mx-auto px-4"
                            >
                                {{
                                    searchQuery
                                        ? "Try adjusting your search terms to find what you're looking for."
                                        : "Start a new conversation to see your chats here."
                                }}
                            </p>
                            <Button
                                @click="
                                    searchQuery
                                        ? clearSearch()
                                        : isOnline
                                          ? handleNewChat()
                                          : null
                                "
                                :disabled="!searchQuery && !isOnline"
                                class="px-5 py-2 bg-gray-900 dark:bg-white text-white dark:text-black text-sm rounded-lg transition-colors inline-flex items-center gap-2 shadow-lg"
                            >
                                <RefreshCw class="w-4 h-4" v-if="searchQuery" />
                                <Plus class="w-4 h-5" v-else />
                                <span>{{
                                    searchQuery
                                        ? "Clear Search"
                                        : "Start New Chat"
                                }}</span>
                            </Button>
                        </div>
                    </div>

                    <!-- Chat List -->
                    <div
                        v-else
                        class="flex-1 mt-4 w-full overflow-y-auto overflow-x-hidden custom-scrollbar pr-1 sm:pr-2"
                    >
                        <div class="flex flex-col pb-4">
                            <div
                                v-for="chat in filteredChats"
                                :key="chat.id"
                                @click="() => handleGoToChat(chat.id)"
                                :class="[
                                    'border-y-[1px] hover:border-none duration-300 ease-in-out cursor-pointer transition-all hover:shadow-lg group',
                                    'border-gray-200 dark:border-gray-800 hover:bg-gray-200 dark:hover:bg-gray-800 hover:rounded-lg',
                                ]"
                            >
                                <div class="px-4 py-3">
                                    <div
                                        class="flex items-start justify-between"
                                    >
                                        <div class="flex-1 min-w-0">
                                            <div
                                                class="flex items-center gap-2 sm:gap-3 mb-2"
                                            >
                                                <p
                                                    class="text-sm font-medium text-gray-800 dark:text-gray-200 truncate"
                                                >
                                                    {{
                                                        chat.title ||
                                                        "Untitled Chat"
                                                    }}
                                                </p>
                                            </div>

                                            <div
                                                class="flex items-center gap-3 text-xs text-gray-500 dark:text-gray-400 flex-wrap"
                                            >
                                                <span
                                                    class="flex items-center gap-1"
                                                >
                                                    <MessageCircle
                                                        class="w-4 h-4"
                                                    />
                                                    {{
                                                        chat.message_count || 0
                                                    }}
                                                    {{
                                                        chat.message_count === 1
                                                            ? "message"
                                                            : "messages"
                                                    }}
                                                </span>
                                                <span
                                                    class="flex items-center gap-1"
                                                >
                                                    <Clock class="w-4 h-4" />
                                                    {{
                                                        new Date(
                                                            chat.last_message_at,
                                                        ).toLocaleDateString(
                                                            undefined,
                                                            {
                                                                month: "short",
                                                                day: "numeric",
                                                                year:
                                                                    screenWidth >
                                                                    640
                                                                        ? "numeric"
                                                                        : undefined,
                                                            },
                                                        )
                                                    }}
                                                </span>
                                            </div>
                                        </div>

                                        <div
                                            class="ml-3 sm:ml-4 opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0"
                                        >
                                            <ChevronRight class="w-5 h-5" />
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </OverallLayout>
</template>
