<script lang="ts" setup>
import type { Chat, ConfirmDialogOptions } from '@/types';
import { inject, ref, type Ref } from 'vue';

// Global state and functions would typically be imported from a store or context
const globalState = inject('globalState') as {
    screenWidth: Ref<number>,
    confirmDialog: Ref<ConfirmDialogOptions>,
    isCollapsed: Ref<boolean>,
    isSidebarHidden: Ref<boolean>,
    authData: Ref<{ username: string; email: string; password: string; }>,
    syncStatus: Ref<{ lastSync: Date | null; syncing: boolean; hasUnsyncedChanges: boolean; }>,
    isAuthenticated: () => boolean,
    parsedUserDetails: any,
    currentChatId: Ref<string>,
    chats: Ref<Chat[]>
    logout: () => void,
    isLoading: Ref<boolean>,
    expanded: Ref<boolean[]>,
    showInput: Ref<boolean>,
    showConfirmDialog: (options: ConfirmDialogOptions) => void,
    hideSidebar: () => void,
    setShowInput: () => void,
    clearAllChats: () => void,
    switchToChat: (chatId: string) => void,
    createNewChat: (initialMessage?: string) => void,
    deleteChat: (chatId: string) => void,
    renameChat: (chatId: string, newTitle: string) => void,
    deleteMessage: (messageIndex: number) => void,
    scrollableElem: Ref<HTMLElement | null>,
    showScrollDownButton: Ref<boolean>,
    handleScroll: () => void,
    scrollToBottom: () => void,
    saveChats: () => void,
    loadLinkPreviewCache: () => void,
    saveLinkPreviewCache: () => void,
    syncFromServer: (data?: any) => void,
    syncToServer: () => void,
    updateExpandedArray: () => void,
    apiCall: (endpoint: string, options: RequestInit) => any,
    toggleSidebar: () => void,
    manualSync: () => Promise<any>
}

const {
    screenWidth,
    confirmDialog,
    isCollapsed,
    isSidebarHidden,
    authData,
    syncStatus,
    isAuthenticated,
    currentChatId,
    chats,
    logout,
    isLoading,
    expanded,
    showInput,
    showConfirmDialog,
    hideSidebar,
    setShowInput,
    clearAllChats,
    switchToChat,
    createNewChat,
    deleteChat,
    renameChat,
    deleteMessage,
    scrollableElem,
    showScrollDownButton,
    handleScroll,
    scrollToBottom,
    saveChats,
    loadLinkPreviewCache,
    saveLinkPreviewCache,
    syncFromServer,
    syncToServer,
    updateExpandedArray,
    apiCall,
    toggleSidebar,
    manualSync
} = globalState
let parsedUserDetails = globalState.parsedUserDetails


</script>

<template>
    <div class="flex h-[100vh]">
        <!-- Sidebar -->
        <SideNav v-if="isAuthenticated()" :data="{
            chats,
            currentChatId,
            parsedUserDetails,
            screenWidth,
            isCollapsed,
            syncStatus,
            isAuthenticated
        }" :functions="{
            setShowInput,
            hideSidebar,
            clearAllChats,
            toggleSidebar,
            logout,
            createNewChat,
            switchToChat,
            deleteChat,
            renameChat,
            manualSync,
        }" />

        <!-- Main Window -->
        <div
            :class="screenWidth > 720 && isAuthenticated() ? (!isCollapsed ?
                'flex-grow flex flex-col items-center justify-center ml-[270px] font-light text-sm transition-all duration-300 ease-in-out'
                :
                'flex-grow flex flex-col items-center justify-center ml-[60px] font-light text-sm transition-all duration-300 ease-in-out'
            )
                : 'text-sm font-light flex-grow items-center justify-center flex flex-col transition-all duration-300 ease-in-out'"
        >
            Settings    
        </div>
    </div>
</template>