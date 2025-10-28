<script lang="ts" setup>
import type { Chat, UserDetails } from "@/types";
import type { Ref } from "vue";
import { inject } from "vue";

const {
    chats,
    parsedUserDetails,
    isCollapsed,
    clearAllChats,
    toggleSidebar,
    logout,
    createNewChat,
    switchToChat,
    deleteChat,
    renameChat,
    manualSync,
    isAuthenticated,
} = inject("globalState") as {
    chats: Ref<Chat[]>;
    parsedUserDetails: Ref<UserDetails>;
    isCollapsed: Ref<boolean>;
    isAuthenticated: Ref<boolean>;

    clearAllChats: () => void;
    toggleSidebar: () => void;
    logout: () => void;
    createNewChat: () => void;
    switchToChat: (id: string) => void;
    deleteChat: (id: string) => void;
    renameChat: (id: string, name: string) => void;
    manualSync: () => void;
};
</script>
<template>
    <div
        v-if="isAuthenticated"
        class="flex h-[100vh] bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100"
    >
        <!-- Sidebar -->
        <SideNav
            :data="{
                chats,
                parsedUserDetails,
                isCollapsed,
            }"
            :functions="{
                clearAllChats,
                toggleSidebar,
                logout,
                createNewChat,
                switchToChat,
                deleteChat,
                renameChat,
                manualSync,
            }"
        />

        <!-- Main Chat Window -->
        <slot></slot>
    </div>
</template>
