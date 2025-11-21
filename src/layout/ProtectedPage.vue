<script lang="ts" setup>
import PreviewSideBar from "@/components/PreviewSideBar.vue";
import type { Chat, UserDetails } from "@/types";
import type { Ref } from "vue";
import { inject, watch } from "vue";
import { toast } from "vue-sonner";

const {
    chats,
    parsedUserDetails,
    clearAllChats,
    toggleSidebar,
    logout,
    createNewChat,
    deleteChat,
    renameChat,
    manualSync,
    isAuthenticated,
    isOnline,
} = inject("globalState") as {
    isOnline: Ref<boolean>;
    chats: Ref<Chat[]>;
    parsedUserDetails: Ref<UserDetails>;
    isAuthenticated: Ref<boolean>;

    clearAllChats: () => void;
    toggleSidebar: () => void;
    logout: () => void;
    createNewChat: () => void;
    deleteChat: (id: string) => void;
    renameChat: (id: string, name: string) => Promise<string>;
    manualSync: () => void;
};

watch(
    isOnline,
    (newIsOnline, oldIsOnline) => {
        if (!newIsOnline && oldIsOnline) {
            toast.error("You are offline", {
                duration: 5000,
                description: "Please check your internet connection",
            });
        } else if (newIsOnline && !oldIsOnline) {
            toast.success("Connection restored", {
                duration: 3000,
                description: "You are back online",
            });
        }
    },
    {
        immediate: false,
    },
);
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
            }"
            :functions="{
                clearAllChats,
                toggleSidebar,
                logout,
                createNewChat,
                deleteChat,
                renameChat,
                manualSync,
            }"
        />

        <!-- Main Chat Window -->
        <slot></slot>

        <!-- preview  -->
        <PreviewSideBar />
    </div>
</template>
