<script lang="ts" setup>
import type { Chat } from '@/types'
import { ref, inject } from 'vue';
import { useRouter } from 'vue-router';
import ChatDropdown from './Dropdowns/ChatDropdown.vue';
import type { Ref } from 'vue';

const globalState = inject('globalState') as {
  activeChatMenu: Ref<string | null>,
  toggleChatMenu: (chatId: string, event: Event) => void
  showProfileMenu: Ref<boolean>,
  handleClickOutside: () => void,
  isAuthenticated: Ref<boolean>
}
const {
  activeChatMenu,
  toggleChatMenu,
  showProfileMenu,
  handleClickOutside,
  isAuthenticated
} = globalState

const props = defineProps<{
  data: {
    chats: Chat[]
    currentChatId: string
    isCollapsed?: boolean
    parsedUserDetails: {
      username: string
      email: string
      sync_enabled: boolean
    }
    screenWidth: number,
    syncStatus: any
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
const isRenaming = ref<string | null>(null) // Track which chat is being renamed
const renameValue = ref('')

const profileOptions = [
  { id: 'settings', label: 'Settings', action: () => router.push('/settings/profile') },
  { id: 'help', label: 'Get Help', action: () => { /* Add your help action */ } },
  { id: 'upgrade', label: 'Upgrade Plan', action: () => router.push('/upgrade') },
  { id: 'learn', label: 'Learn More', action: () => { /* Add your learn more action */ } }
];

function startRename(chatId: string, currentTitle: string) {
  isRenaming.value = chatId
  renameValue.value = currentTitle
  activeChatMenu.value = null

  // Focus the input after Vue updates the DOM
  setTimeout(() => {
    const input = document.getElementById(`rename-${chatId}`) as HTMLInputElement
    if (input) {
      input.focus()
      input.select()
    }
  }, 50)
}

function submitRename(chatId: string) {
  if (renameValue.value.trim()) {
    props.functions.renameChat(chatId, renameValue.value.trim())
  }
  isRenaming.value = null
  renameValue.value = ''
}

function cancelRename() {
  isRenaming.value = null
  renameValue.value = ''
}

function handleChatClick(chatId: string) {
  props.functions.switchToChat(chatId)
  props.functions.setShowInput()
  if (router.currentRoute.value.path !== '/') {
    router.push('/')
  }
  if (props.data.screenWidth < 720) {
    props.functions.hideSidebar()
  }
}
</script>

<template>
  <div id="side_nav" :class="props.data.screenWidth > 720
    ? props.data.isCollapsed
      ? 'w-[60px] bg-white z-30 fixed top-0 left-0 bottom-0 border-r flex flex-col transition-all duration-300 ease-in-out'
      : 'w-[270px] bg-white z-30 fixed top-0 left-0 bottom-0 border-r flex flex-col transition-all duration-300 ease-in-out'
    : 'none'
    " @click="handleClickOutside">

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
      <div v-if="props.data.parsedUserDetails.username" class="px-3 my-4 flex flex-col gap-1 font-light text-sm">
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

        <div v-if="isAuthenticated">
          <!-- Sync button -->
          <button title="Sync Data" @click="props.functions.manualSync"
            :disabled="props.data.syncStatus.syncing || !props.data.parsedUserDetails.sync_enabled"
            class="w-full flex items-center gap-2 h-[40px] hover:bg-gray-100 rounded-lg px-2 disabled:opacity-50 disabled:cursor-not-allowed">
            <i :class="props.data.syncStatus.syncing ? 'pi pi-spin pi-spinner' : 'pi pi-refresh'"></i>
            <span v-if="!props.data.isCollapsed || props.data.screenWidth < 720">
              {{ props.data.syncStatus.syncing ? 'Syncing...' : 'Sync Data' }}
            </span>
            <div
              v-if="props.data.syncStatus.hasUnsyncedChanges && props.data.parsedUserDetails.sync_enabled && (!props.data.isCollapsed || props.data.screenWidth < 720)"
              class="ml-auto w-2 h-2 bg-orange-500 rounded-full"></div>
          </button>

          <!-- Toggle sync -->
          <!-- <div v-if="!props.data.isCollapsed || props.data.screenWidth < 720"
            class="flex items-center justify-between mt-2 px-2">
            <span class="text-sm text-gray-700">Enable Sync</span>
            <input type="checkbox" v-model="props.data.parsedUserDetails.sync_enabled"
              class="rounded border-gray-300 text-blue-600 focus:ring-blue-500" />
          </div> -->
        </div>

      </div>

      <!-- Recent Chats -->
      <div v-if="props.data.chats.length && props.data.parsedUserDetails.username"
        class="flex flex-col px-2 mb-2 py-4 font-light">
        <p v-if="!props.data.isCollapsed || props.data.screenWidth < 720" class="text-base text-gray-600 mb-2">Chats</p>
        <div class="flex flex-col gap-2">
          <div v-for="chat in !props.data.isCollapsed ? props.data.chats : props.data.chats.slice(0, 1)" :key="chat.id"
            :class="chat.id === props.data.currentChatId
              ? 'w-full flex h-[32px] text-sm items-center bg-gray-300 rounded-lg relative'
              : 'w-full flex h-[32px] text-sm items-center hover:bg-gray-100 rounded-lg relative'">

            <!-- Chat content area -->
            <div @click="handleChatClick(chat.id)" class="flex items-center h-full flex-grow px-2 cursor-pointer">
              <i class="pi pi-comments mr-2 text-gray-500 mb-[2px]"></i>

              <!-- Chat title or rename input -->
              <div v-if="isRenaming === chat.id" class="flex-grow" @click.stop>
                <input :id="`rename-${chat.id}`" v-model="renameValue" @keyup.enter="submitRename(chat.id)"
                  @keyup.escape="cancelRename" @blur="submitRename(chat.id)"
                  class="w-full px-1 py-0.5 text-xs bg-white border border-blue-500 rounded focus:outline-none" />
              </div>
              <p v-else-if="!props.data.isCollapsed || props.data.screenWidth < 720" class="truncate">
                <span v-if="chat.title.length>20">{{ `${chat.title.slice(0,20)}..` || 'Untitled Chat' }}</span>
                <span v-else>{{ chat.title || 'Untitled Chat' }}</span>
              </p>
            </div>

            <!-- Menu button -->
            <div v-if="!props.data.isCollapsed || props.data.screenWidth < 720" @click="toggleChatMenu(chat.id, $event)"
              class="flex items-center justify-center h-full hover:bg-blue-600 hover:text-white rounded-r-lg flex-shrink px-3 cursor-pointer">
              <i class="pi pi-ellipsis-h"></i>
            </div>

            <ChatDropdown :data="{
              activeChatMenu,
              chat,
            }" :functions="{
              deleteChat: props.functions.deleteChat,
              startRename
            }" />
          </div>
        </div>
      </div>
    </div>

    <!-- Fixed Bottom User Profile -->
    <div class="border-t border-gray-200 p-3 bg-gray-100 sticky bottom-0">
      <div class="flex items-center justify-between cursor-pointer mr-1"
        @click.stop="showProfileMenu = !showProfileMenu">
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

      <!-- Profile Dropdown -->
      <transition name="fade">
        <div v-if="showProfileMenu"
          class="absolute bottom-full left-3 right-3 mb-2 bg-white border rounded-lg shadow-lg text-sm z-50"
          @click.stop>
          <p class="px-4 py-2 text-gray-500 border-b">{{ props.data.parsedUserDetails.email || 'No email' }}</p>
          <button v-for="option in profileOptions" :key="option.id" @click="option.action"
            class="w-full text-left px-4 py-2 hover:bg-gray-100">
            {{ option.label }}
          </button>
          <button @click="props.functions.logout"
            class="w-full text-left px-4 py-2 text-red-600 hover:bg-red-100 rounded-b-lg">
            Log Out
          </button>
        </div>
      </transition>
    </div>
  </div>
</template>