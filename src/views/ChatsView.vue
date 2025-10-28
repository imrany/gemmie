<script lang="ts" setup>
import { inject, ref, computed } from "vue";
import { useRouter } from "vue-router";
import type { Ref } from "vue";
import type { Chat, UserDetails } from "@/types";
import {
    ChevronLeft,
    ChevronRight,
    Clock,
    Inbox,
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
    switchToChat,
    createNewChat,
} = inject("globalState") as {
    chats: Ref<Chat[]>;
    currentChatId: Ref<string>;
    parsedUserDetails: Ref<UserDetails>;
    screenWidth: Ref<number>;
    isCollapsed: Ref<boolean>;
    switchToChat: (chatId: string) => void;
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

const handleChatClick = (chatId: string) => {
    switchToChat(chatId);
    router.push("/");
};

const handleNewChat = () => {
    createNewChat();
    router.push("/new");
};

const handleBack = () => {
    router.back();
};

const clearSearch = () => {
    searchQuery.value = "";
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
                    class="w-full flex flex-col h-full overflow-hidden md:max-w-3xl max-w-[100vw]"
                >
                    <!-- Top Header -->
                    <div
                        class="flex-shrink-0 w-full items-center justify-between mb-4 sm:mb-6 flex gap-2 min-w-0"
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
                            class="text-gray-700 dark:text-gray-300 text-xl sm:text-2xl font-semibold truncate min-w-0"
                        >
                            Your chat history
                        </p>

                        <button
                            @click="handleNewChat"
                            class="px-3 py-2 bg-blue-600 text-xs sm:text-sm hover:bg-blue-700 text-white rounded-lg transition-colors flex items-center gap-1.5 sm:gap-2 shadow-lg hover:shadow-xl flex-shrink-0 whitespace-nowrap"
                        >
                            <Plus class="w-5 h-5" />
                            <span class="font-medium hidden sm:inline">
                                New Chat
                            </span>
                            <span class="font-medium sm:hidden">New</span>
                        </button>
                    </div>

                    <!-- Search Bar -->
                    <div class="mb-4 sm:mb-6 flex-shrink-0">
                        <div class="relative">
                            <Search
                                class="w-5 h-5 absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400"
                            />
                            <input
                                v-model="searchQuery"
                                type="text"
                                placeholder="Search chats and messages..."
                                class="w-full text-sm pl-10 pr-10 py-2.5 sm:py-3 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
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
                        class="flex-1 flex items-center justify-center overflow-y-auto"
                    >
                        <div class="text-center py-8 sm:py-12 px-4">
                            <div class="text-gray-300 dark:text-gray-600 mb-4">
                                <Inbox class="w-6 h-6" />
                            </div>
                            <h3
                                class="text-base sm:text-lg font-medium text-gray-900 dark:text-white mb-2"
                            >
                                {{
                                    searchQuery
                                        ? "No matching chats found"
                                        : "No chats yet"
                                }}
                            </h3>
                            <p
                                class="text-sm sm:text-base text-gray-500 dark:text-gray-400 mb-6 max-w-md mx-auto px-4"
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
                                class="px-5 sm:px-6 py-2.5 sm:py-3 bg-blue-600 hover:bg-blue-700 text-white text-sm sm:text-base rounded-lg transition-colors inline-flex items-center gap-2 shadow-lg"
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
                        class="flex-1 mt-4 md:max-w-3xl max-w-[100vw] overflow-y-auto overflow-x-hidden custom-scrollbar pr-1 sm:pr-2"
                    >
                        <div class="flex flex-col gap-3 sm:gap-4 pb-4">
                            <div
                                v-for="chat in filteredChats"
                                :key="chat.id"
                                @click="handleChatClick(chat.id)"
                                :class="[
                                    'bg-white dark:bg-gray-800 rounded-xl border-[1px] cursor-pointer transition-all hover:shadow-lg group',
                                    currentChatId === chat.id
                                        ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                                        : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600',
                                ]"
                            >
                                <div class="p-4 sm:p-5">
                                    <div
                                        class="flex items-start justify-between"
                                    >
                                        <div class="flex-1 min-w-0">
                                            <div
                                                class="flex items-center gap-2 sm:gap-3 mb-2"
                                            >
                                                <h3
                                                    class="text-lg sm:text-xl font-semibold text-gray-900 dark:text-white truncate"
                                                >
                                                    {{
                                                        chat.title ||
                                                        "Untitled Chat"
                                                    }}
                                                </h3>
                                                <div
                                                    v-if="
                                                        currentChatId ===
                                                        chat.id
                                                    "
                                                    class="flex items-center gap-1 px-2 py-1 bg-blue-100 dark:bg-blue-800 text-blue-700 dark:text-blue-300 rounded-full text-xs font-medium flex-shrink-0"
                                                >
                                                    <div
                                                        class="w-1.5 h-1.5 sm:w-2 sm:h-2 bg-blue-500 rounded-full animate-pulse"
                                                    ></div>
                                                    <span
                                                        class="hidden sm:inline"
                                                        >Active</span
                                                    >
                                                </div>
                                            </div>

                                            <div
                                                class="flex items-center gap-3 sm:gap-4 text-xs sm:text-sm text-gray-500 dark:text-gray-400 mb-2 sm:mb-3 flex-wrap"
                                            >
                                                <span
                                                    class="flex items-center gap-1"
                                                >
                                                    <MessageCircle
                                                        class="w-4 h-4"
                                                    />
                                                    {{
                                                        chat.messages?.length ||
                                                        0
                                                    }}
                                                    {{
                                                        chat.messages
                                                            ?.length === 1
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
                                                            chat.updatedAt,
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

                                            <!-- Preview of last message -->
                                            <div
                                                v-if="
                                                    chat.messages &&
                                                    chat.messages.length > 0
                                                "
                                                class="text-gray-600 dark:text-gray-300 line-clamp-2 text-xs sm:text-sm bg-gray-50 dark:bg-gray-700/50 rounded-lg p-2.5 sm:p-3"
                                            >
                                                {{
                                                    chat.messages[
                                                        chat.messages.length - 1
                                                    ]?.response?.substring(
                                                        0,
                                                        screenWidth < 640
                                                            ? 100
                                                            : 150,
                                                    ) || "No messages yet"
                                                }}
                                                {{
                                                    chat.messages[
                                                        chat.messages.length - 1
                                                    ]?.response?.length >
                                                    (screenWidth < 640
                                                        ? 100
                                                        : 150)
                                                        ? "..."
                                                        : ""
                                                }}
                                            </div>
                                            <div
                                                v-else
                                                class="text-gray-400 dark:text-gray-500 text-xs sm:text-sm italic"
                                            >
                                                No messages yet
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
