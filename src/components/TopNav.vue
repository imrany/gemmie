<script lang="ts" setup>
import type { Chat, CurrentChat } from '@/types'
import type { Ref } from 'vue';
import { inject } from 'vue';

const globalState= inject('globalState') as {
  isAuthenticated: Ref<boolean>,
  syncStatus: Ref<{ 
    lastSync: Date | null; 
    syncing: boolean; 
    hasUnsyncedChanges: boolean;
    showSyncIndicator: boolean;
    syncMessage: string;
    syncProgress: number;
    lastError: string | null;
    retryCount: number;
    maxRetries: number;
  }>,
}
const { isAuthenticated, syncStatus }= globalState
let props = defineProps<{
  data: {
    isCollapsed?: boolean
    parsedUserDetails: { username: string }
    currentChat: CurrentChat | undefined
    screenWidth: number
    isSidebarHidden?: boolean
    chat?:Chat,
  }
  functions: {
    hideSidebar: () => void
    manualSync: () => void
  }
}>()
</script>

<template>
  <div class="bg-white dark:bg-gray-900 h-[52px] z-30 fixed top-0 right-0 border-b border-gray-200 dark:border-gray-700 transition-all duration-300 ease-in-out" :style="props.data.screenWidth > 720 && !props.data.isCollapsed
      ? 'left:270px'
      : props.data.screenWidth > 720 && props.data.isCollapsed
        ? 'left:60px'
        : 'left:0'
    ">
    <div class="flex h-full px-4 items-center justify-between w-full">
      <!-- Brand -->
      <p v-if="props.data.currentChat&&props.data.screenWidth > 720" class="text-black dark:text-white font-medium truncate text-base select-none">
        <span v-if="props.data.currentChat.title.length>20">{{ props.data.currentChat.title.slice(0,20) }}...</span>
        <span v-else>{{ props.data.currentChat.title }}</span>
      </p>
      <p v-else class="text-black dark:text-white text-xl max-md:text-2xl font-semibold tracking-wide select-none">
        Gemmie
      </p>


      <!-- Mobile Sidebar Toggle -->
      <div v-if="props.data.screenWidth < 720" class="flex gap-2 items-center ml-auto">
        <div v-if="isAuthenticated" class="relative">
          <div v-if="syncStatus.syncing"
            class="flex items-center gap-2 bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-400 px-3 py-1.5 rounded-full text-xs border border-blue-200 dark:border-blue-800 shadow-sm animate-pulse">
            <i class="pi pi-spin pi-spinner"></i>
            <span>Syncing...</span>
          </div>

          <div v-else-if="syncStatus.hasUnsyncedChanges"
            class="flex items-center gap-2 bg-orange-50 dark:bg-orange-900/20 text-orange-700 dark:text-orange-400 px-3 py-1.5 rounded-full text-xs border border-orange-200 dark:border-orange-800 shadow-sm cursor-pointer hover:bg-orange-100 dark:hover:bg-orange-900/30 transition"
            @click="props.functions.manualSync">
            <i class="pi pi-cloud-upload"></i>
            <span>Sync pending</span>
          </div>

          <div v-else-if="syncStatus.lastSync"
            class="flex items-center gap-2 bg-green-50 dark:bg-green-900/20 text-green-700 dark:text-green-400 px-3 py-1.5 rounded-full text-xs border border-green-200 dark:border-green-800 shadow-sm">
            <i class="pi pi-check-circle"></i>
            <span>Synced</span>
          </div>
        </div>
        <button @click="props.functions.hideSidebar" title="Toggle Sidebar"
          class="w-9 h-9 flex items-center justify-center hover:bg-gray-100 dark:hover:bg-gray-800 rounded-full cursor-pointer transition-colors">
          <span v-if="props.data.isSidebarHidden" class="pi pi-bars text-lg text-gray-700 dark:text-gray-300"></span>
          <span v-else class="pi pi-times text-lg text-gray-700 dark:text-gray-300"></span>
        </button>
      </div>

      <!-- Desktop Actions -->
      <div v-else class="flex gap-3 items-center ml-auto">
        <!-- Sync Status -->
        <div v-if="isAuthenticated" class="relative">
          <div v-if="syncStatus.syncing"
            class="flex items-center gap-2 bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-400 px-3 py-1.5 rounded-full text-xs border border-blue-200 dark:border-blue-800 shadow-sm animate-pulse">
            <i class="pi pi-spin pi-spinner"></i>
            <span>Syncing...</span>
          </div>

          <div v-else-if="syncStatus.hasUnsyncedChanges"
            class="flex items-center gap-2 bg-orange-50 dark:bg-orange-900/20 text-orange-700 dark:text-orange-400 px-3 py-1.5 rounded-full text-xs border border-orange-200 dark:border-orange-800 shadow-sm cursor-pointer hover:bg-orange-100 dark:hover:bg-orange-900/30 transition"
            @click="props.functions.manualSync">
            <i class="pi pi-cloud-upload"></i>
            <span>Sync pending</span>
          </div>

          <div v-else-if="syncStatus.lastSync"
            class="flex items-center gap-2 bg-green-50 dark:bg-green-900/20 text-green-700 dark:text-green-400 px-3 py-1.5 rounded-full text-xs border border-green-200 dark:border-green-800 shadow-sm">
            <i class="pi pi-check-circle"></i>
            <span>Synced</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
