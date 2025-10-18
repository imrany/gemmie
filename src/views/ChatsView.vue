<script lang="ts" setup>
import { inject, ref, computed } from 'vue';
import { useRouter } from 'vue-router';
import SideNav from '@/components/SideNav.vue';
import type { Ref } from 'vue';
import type { Chat, UserDetails } from '@/types';

const {
    chats,
    currentChatId,
    parsedUserDetails,
    screenWidth,
    isCollapsed,
    switchToChat,
    createNewChat,
    deleteChat,
    setShowInput,
    hideSidebar,
    clearAllChats,
    toggleSidebar,
    logout,
    renameChat,
    manualSync
} = inject('globalState') as {
    chats: Ref<Chat[]>,
    currentChatId: Ref<string>,
    parsedUserDetails: Ref<UserDetails>,
    screenWidth: Ref<number>,
    isCollapsed: Ref<boolean>,
    switchToChat: (chatId: string) => void,
    createNewChat: (initialMessage?: string) => string,
    deleteChat: (chatId: string) => void,
    setShowInput: () => void,
    hideSidebar: () => void,
    clearAllChats: () => void,
    toggleSidebar: () => void,
    logout: () => void,
    renameChat: (chatId: string, newTitle: string) => void,
    manualSync: () => void
};

const router = useRouter();
const searchQuery = ref('');

// Filter chats based on search
const filteredChats = computed(() => {
    if (!searchQuery.value.trim()) {
        return chats.value;
    }

    const query = searchQuery.value.toLowerCase();
    return chats.value.filter(chat =>
        chat.title?.toLowerCase().includes(query) ||
        chat.messages?.some(message =>
            message.prompt?.toLowerCase().includes(query) ||
            message.response?.toLowerCase().includes(query)
        )
    );
});

const handleChatClick = (chatId: string) => {
    switchToChat(chatId);
    setShowInput();
    if (screenWidth.value < 720) hideSidebar();
    router.push('/');
};

const handleNewChat = () => {
    createNewChat();
    if (screenWidth.value < 720) hideSidebar();
    router.push('/new');
};

const clearSearch = () => {
    searchQuery.value = '';
};
</script>

<template>
    <div class="flex min-h-screen bg-white dark:bg-gray-900">
        <!-- Sidebar -->
        <SideNav v-if="parsedUserDetails?.username" :data="{
            chats,
            parsedUserDetails,
            isCollapsed,
        }" :functions="{
            setShowInput,
            clearAllChats,
            toggleSidebar,
            logout,
            createNewChat,
            switchToChat,
            deleteChat,
            renameChat,
            manualSync
        }" />

        <!-- Main Content - Centered -->
        <div :class="[
            'flex-grow flex flex-col transition-all duration-300 ease-in-out',
            screenWidth > 720 && parsedUserDetails?.username
                ? (isCollapsed ? 'ml-[60px]' : 'ml-[270px]')
                : ''
        ]">

            <!-- Centered Container -->
            <div class="flex-1 flex items-center justify-center p-4 md:p-6">
                <div class="w-full max-w-4xl flex flex-col max-h-full">
                    <!-- Top-->
                    <div class="flex w-full items-center justify-between  mb-6">
                        <p class="text-gray-700 dark:text-gray-300 text-2xl max-md:text-2xl font-semibold">
                            Your chat history
                        </p>
                        <button @click="handleNewChat"
                            class="px-3 py-2 bg-blue-600 text-sm hover:bg-blue-700 text-white rounded-lg transition-colors flex items-center gap-2 shadow-lg hover:shadow-xl">
                            <i class="pi pi-plus"></i>
                            <span class="font-medium">New Chat</span>
                        </button>
                    </div>

                    <!-- Search Bar -->
                    <div class="mb-6">
                        <div class="relative">
                            <i
                                class="pi pi-search absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400"></i>
                            <input v-model="searchQuery" type="text" placeholder="Search chats and messages..."
                                class="w-full text-sm pl-10 pr-10 py-3 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all" />
                            <button v-if="searchQuery" @click="clearSearch"
                                class="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors">
                                <i class="pi pi-times"></i>
                            </button>
                        </div>
                        <p class="text-start text-sm font-medium text-gray-500 dark:text-gray-400 mt-2">
                            {{ filteredChats.length }} chats
                        </p>
                    </div>

                    <!-- Empty State -->
                    <div v-if="filteredChats.length === 0" class="text-center py-12">
                        <div class="text-gray-300 dark:text-gray-600 mb-4">
                            <i class="pi pi-inbox text-6xl"></i>
                        </div>
                        <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-2">
                            {{ searchQuery ? 'No matching chats found' : 'No chats yet' }}
                        </h3>
                        <p class="text-gray-500 dark:text-gray-400 mb-6 max-w-md mx-auto">
                            {{ searchQuery
                                ? 'Try adjusting your search terms to find what you\'re looking for.'
                                : 'Start a new conversation to see your chats here.'
                            }}
                        </p>
                        <button @click="searchQuery ? clearSearch() : handleNewChat()"
                            class="px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors inline-flex items-center gap-2 shadow-lg">
                            <i :class="searchQuery ? 'pi pi-refresh' : 'pi pi-plus'"></i>
                            <span>{{ searchQuery ? 'Clear Search' : 'Start New Chat' }}</span>
                        </button>
                    </div>

                    <!-- Chat List -->
                    <div v-else class="grid gap-4 overflow-auto custom-scrollbar">
                        <div v-for="chat in filteredChats" :key="chat.id" @click="handleChatClick(chat.id)" :class="[
                            'bg-white dark:bg-gray-800 rounded-xl border-[1px] cursor-pointer transition-all hover:shadow-lg group',
                            currentChatId === chat.id
                                ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                                : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'
                        ]">
                            <div class="p-5">
                                <div class="flex items-start justify-between">
                                    <div class="flex-1 min-w-0">
                                        <div class="flex items-center gap-3 mb-2">
                                            <h3 class="text-xl font-semibold text-gray-900 dark:text-white truncate">
                                                {{ chat.title || 'Untitled Chat' }}
                                            </h3>
                                            <div v-if="currentChatId === chat.id"
                                                class="flex items-center gap-1 px-2 py-1 bg-blue-100 dark:bg-blue-800 text-blue-700 dark:text-blue-300 rounded-full text-xs font-medium">
                                                <div class="w-2 h-2 bg-blue-500 rounded-full animate-pulse"></div>
                                                <span>Active</span>
                                            </div>
                                        </div>

                                        <div
                                            class="flex items-center gap-4 text-sm text-gray-500 dark:text-gray-400 mb-3">
                                            <span class="flex items-center gap-1">
                                                <i class="pi pi-comments"></i>
                                                {{ chat.messages?.length || 0 }} messages
                                            </span>
                                            <span class="flex items-center gap-1">
                                                <i class="pi pi-clock"></i>
                                                {{ new Date(chat.updatedAt).toLocaleDateString() }}
                                            </span>
                                        </div>

                                        <!-- Preview of last message -->
                                        <div v-if="chat.messages && chat.messages.length > 0"
                                            class="text-gray-600 dark:text-gray-300 line-clamp-2 text-sm bg-gray-50 dark:bg-gray-700/50 rounded-lg p-3">
                                            {{ chat.messages[chat.messages.length - 1]?.response?.substring(0, 150)
                                                || 'No messages yet' }}
                                            {{ chat.messages[chat.messages.length - 1]?.response?.length > 150 ? '...' :
                                                ''
                                            }}
                                        </div>
                                        <div v-else class="text-gray-400 dark:text-gray-500 text-sm italic">
                                            No messages yet
                                        </div>
                                    </div>

                                    <div class="ml-4 opacity-0 group-hover:opacity-100 transition-opacity">
                                        <i class="pi pi-chevron-right text-gray-400 text-lg"></i>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>