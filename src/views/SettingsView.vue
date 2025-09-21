<script lang="ts" setup>
import type { ConfirmDialogOptions } from '@/types';
import { ref } from 'vue';

const confirmDialog = ref<ConfirmDialogOptions>({
  visible: false,
  title: '',
  message: '',
  type: 'info' as 'danger' | 'warning' | 'info',
  confirmText: 'Confirm',
  cancelText: 'Cancel',
  onConfirm: () => { }
})
</script>

<template>
    <div class="flex h-[100vh]">
        <!-- Custom Confirmation Dialog -->
        <ConfirmDialog v-if="confirmDialog.visible" :data="confirmDialog" />

        <!-- Sidebar -->
        <SideNav v-if="isAuthenticated()" :data="{
            chats,
            currentChatId,
            parsedUserDetails,
            screenWidth,
            isCollapsed,
            syncStatus,
            isAuthenticated
        }" 
        :functions="{
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
                : 'text-sm font-light flex-grow items-center justify-center flex flex-col transition-all duration-300 ease-in-out'">
            >
        </div>
    </div>
</template>