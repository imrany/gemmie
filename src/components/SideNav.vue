<script lang="ts" setup>
import type { Chat } from '@/types'
import { ref } from 'vue';
import { useRouter } from 'vue-router';

let props = defineProps<{
  data: {
    chats: Chat[]
    currentChatId: string
    isCollapsed?: boolean
    parsedUserDetails: {
      username: string
      email: string
    }
    screenWidth: number,
    syncStatus: any,
    isAuthenticated: () => boolean
  }
  functions: {
    setShowInput: () => void
    hideSidebar: () => void
    clearAllChats: () => void
    toggleSidebar: () => void
    logout: () => void
    createNewChat: () => void
    switchToChat: (chatId: string) => void
    deleteChat: (chatId: string) => void
    renameChat: (chatId: string, newTitle: string) => void
    manualSync: () => void
  }
}>()

const router = useRouter()
const showProfileMenu = ref(false)

function reload() {
  window.location.reload()
}

function openUpgrade() {
  window.open('https://github.com/sponsors/imrany', '_blank')
}
</script>

<template>
  <div id="side_nav" :class="props.data.screenWidth > 720
    ? props.data.isCollapsed
      ? 'w-[60px] bg-white z-30 fixed top-0 left-0 bottom-0 border-r flex flex-col transition-all duration-300 ease-in-out'
      : 'w-[270px] bg-white z-30 fixed top-0 left-0 bottom-0 border-r flex flex-col transition-all duration-300 ease-in-out'
    : 'none'
    ">
    <!-- Scrollable area -->
    <div class="flex-1 overflow-y-auto">
      <!-- Top Header -->
      <div class="flex items-center justify-between p-3">
        <p v-if="!props.data.isCollapsed" class="font-semibold text-xl text-black">
          Gemmie
        </p>
        <div class="flex gap-2 items-center ml-auto">
          <button @click="props.functions.toggleSidebar" title="Toggle Sidebar"
            class="w-[30px] h-[30px] flex items-center justify-center hover:bg-gray-100 rounded-full cursor-pointer">
            <span class="pi pi-bars text-base"></span>
          </button>
        </div>
      </div>

      <!-- New Chat & Actions -->
      <div v-if="props.data.parsedUserDetails.username"
        class="px-3 my-4 flex flex-col gap-1 font-light text-sm">
        
        <button @click="
          () => {
            props.functions.createNewChat()
            props.functions.setShowInput()
            if (router.currentRoute.value.path !== '/') {
              router.push('/')
            }
            if (props.data.screenWidth < 720) props.functions.hideSidebar()
          }
        " title="New Chat" class="w-full flex items-center gap-2 h-[40px] hover:bg-gray-100 rounded-lg px-2">
          <i class="pi pi-pencil text-gray-500 mb-[2px]"></i>
          <p v-if="!props.data.isCollapsed || props.data.screenWidth < 720">New Chat</p>
        </button>

        <button title="Sync Data" v-if="props.data.isAuthenticated()" @click="props.functions.manualSync"
          :disabled="props.data.syncStatus.syncing"
          class="w-full flex items-center gap-2 h-[40px] hover:bg-gray-100 rounded-lg px-2 disabled:opacity-50 disabled:cursor-not-allowed">
          <i :class="props.data.syncStatus.syncing ? 'pi pi-spin pi-spinner' : 'pi pi-refresh'"></i>
          <span v-if="!props.data.isCollapsed || props.data.screenWidth < 720">
            {{ props.data.syncStatus.syncing ? 'Syncing...' : 'Sync Data' }}
          </span>
          <div v-if="props.data.syncStatus.hasUnsyncedChanges && (!props.data.isCollapsed || props.data.screenWidth < 720)"
            class="ml-auto w-2 h-2 bg-orange-500 rounded-full"></div>
        </button>

        <button @click="() => { router.push('/auth/delete_account') }" title="Delete Account"
          class="w-full flex items-center gap-2 h-[40px] hover:bg-red-100 rounded-lg px-2">
          <i class="pi pi-user-minus text-gray-500 mb-[2px]"></i>
          <p v-if="!props.data.isCollapsed || props.data.screenWidth < 720">Delete Account</p>
        </button>
      </div>

      <!-- Recent Chats -->
      <div v-if="props.data.chats.length && props.data.parsedUserDetails.username"
        class="flex flex-col px-2 mb-2 py-4 font-light">
        <p v-if="!props.data.isCollapsed || props.data.screenWidth < 720"
          class="text-base text-gray-600 mb-2">Chats</p>
        <div class="flex flex-col gap-2">
          <button v-for="chat in !props.data.isCollapsed ? props.data.chats : props.data.chats.slice(0, 1)"
            :key="chat.id"
            @click="
              () => {
                props.functions.switchToChat(chat.id)
                props.functions.setShowInput()
                if (router.currentRoute.value.path !== '/') {
                  router.push('/')
                }
                if (props.data.screenWidth < 720) props.functions.hideSidebar()
              }
            "
            :class="chat.id === props.data.currentChatId 
              ? 'w-full flex h-[32px] text-sm items-center bg-gray-300 rounded-lg px-2'
              : 'w-full flex h-[32px] text-sm items-center hover:bg-gray-100 rounded-lg px-2'">
            <i class="pi pi-comments mr-2 text-gray-500 mb-[2px]"></i>
            <p v-if="!props.data.isCollapsed || props.data.screenWidth < 720" class="truncate">
              {{ chat.title || 'Untitled Chat' }}
            </p>
          </button>
        </div>
      </div>
    </div>

    <!-- Fixed Bottom User Profile -->
    <div class="border-t border-gray-200 p-3 bg-gray-100 sticky bottom-0">
      <div class="flex items-center justify-between cursor-pointer mr-1" @click="showProfileMenu = !showProfileMenu">
        <div class="flex items-center gap-2">
          <div class="w-[35px] h-[35px] flex justify-center items-center bg-gray-300 rounded-full">
            <span class="text-sm">{{ props.data.parsedUserDetails.username.toUpperCase().slice(0, 2) }}</span>
          </div>
          <p v-if="!props.data.isCollapsed || props.data.screenWidth < 720" class="text-base font-light">
            {{ props.data.parsedUserDetails.username }}
          </p>
        </div>
        <i class="pi pi-chevron-up text-xs" v-if="showProfileMenu"></i>
        <i class="pi pi-chevron-down text-xs" v-else></i>
      </div>

      <!-- Dropdown -->
      <transition name="fade">
        <div v-if="showProfileMenu"
          class="absolute bottom-14 left-3 right-3 bg-white border rounded-lg shadow-lg text-sm z-50">
          <p class="px-4 py-2 text-gray-500 border-b">{{ props.data.parsedUserDetails.email || 'No email' }}</p>
          <button @click="router.push('/settings')" class="w-full text-left px-4 py-2 hover:bg-gray-100">Settings</button>
          <button class="w-full text-left px-4 py-2 hover:bg-gray-100">Get Help</button>
          <button @click="router.push('/upgrade')" class="w-full text-left px-4 py-2 hover:bg-gray-100">Upgrade Plan</button>
          <button class="w-full text-left px-4 py-2 hover:bg-gray-100">Learn More</button>
          <button @click="props.functions.logout"
            class="w-full text-left px-4 py-2 text-red-600 hover:bg-red-100">Log Out</button>
        </div>
      </transition>
    </div>
  </div>
</template>