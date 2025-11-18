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
import ProtectedPage from "@/layout/ProtectedPage.vue";

const {
    chats,
    currentChatId,
    parsedUserDetails,
    screenWidth,
    isCollapsed,
    createNewChat,
} = inject("globalState") as {
    chats: Ref<Chat[]>;
    currentChatId: Ref<string>;
    parsedUserDetails: Ref<UserDetails>;
    screenWidth: Ref<number>;
    isCollapsed: Ref<boolean>;
    createNewChat: (initialMessage?: string) => string;
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
    createNewChat();
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
    <ProtectedPage>
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
                    class="w-full px-2 flex flex-col h-full overflow-hidden md:max-w-4xl max-w-[100vw]"
                >
                    <!-- Top Header -->
                    <div
                        v-if="chats.length !== 0"
                        class="flex-shrink-0 w-full items-center justify-between mb-4 flex gap-2 min-w-0"
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

                        <p
                            class="text-gray-700 dark:text-gray-300 text-xl font-semibold truncate min-w-0"
                        >
                            Your chat history
                        </p>

                        <button
                            @click="handleNewChat"
                            class="px-3 py-2 dark:bg-white text-white bg-gray-900 text-sm dark:text-gray-800 rounded-lg transition-colors flex items-center gap-1.5 shadow-lg hover:shadow-xl flex-shrink-0 whitespace-nowrap"
                        >
                            <Plus class="w-5 h-5" />
                            <span class="hidden sm:inline"> New Chat </span>
                            <span class="sm:hidden">New</span>
                        </button>
                    </div>

                    <!-- Search Bar -->
                    <div v-if="chats.length !== 0" class="mb-4 flex-shrink-0">
                        <div class="relative">
                            <Search
                                class="w-5 h-5 absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400"
                            />
                            <input
                                v-model="searchQuery"
                                type="text"
                                placeholder="Search chats and messages..."
                                class="w-full text-base pl-10 pr-10 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-inherit text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:ring-1 focus:ring-blue-500 focus:border-transparent transition-all"
                            />
                            <button
                                v-if="searchQuery"
                                @click="clearSearch"
                                class="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
                            >
                                <X class="w-5 h-5" />
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
                            <button
                                @click="
                                    searchQuery
                                        ? clearSearch()
                                        : handleNewChat()
                                "
                                class="px-5 py-2 bg-gray-900 dark:bg-white text-white dark:text-black text-sm rounded-lg transition-colors inline-flex items-center gap-2 shadow-lg"
                            >
                                <RefreshCw class="w-4 h-4" v-if="searchQuery" />
                                <Plus class="w-4 h-5" v-else />
                                <span>{{
                                    searchQuery
                                        ? "Clear Search"
                                        : "Start New Chat"
                                }}</span>
                            </button>
                        </div>
                    </div>

                    <!-- Chat List -->
                    <div
                        v-else
                        class="flex-1 mt-4 md:max-w-4xl max-w-[100vw] overflow-y-auto overflow-x-hidden custom-scrollbar pr-1 sm:pr-2"
                    >
                        <div class="flex flex-col gap-3 sm:gap-4 pb-4">
                            <div
                                v-for="chat in filteredChats"
                                :key="chat.id"
                                @click="() => handleGoToChat(chat.id)"
                                :class="[
                                    'rounded-xl border-[1px] cursor-pointer transition-all hover:shadow-lg group',
                                    currentChatId === chat.id
                                        ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                                        : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600',
                                ]"
                            >
                                <div class="p-4">
                                    <div
                                        class="flex items-start justify-between"
                                    >
                                        <div class="flex-1 min-w-0">
                                            <div
                                                class="flex items-center gap-2 sm:gap-3 mb-2"
                                            >
                                                <p
                                                    class="text-base font-medium text-gray-900 dark:text-gray-100 truncate"
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
    </ProtectedPage>
</template>
