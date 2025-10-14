<script lang="ts" setup>
import type { Chat } from '@/types'
import { ref, inject, computed, onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import ChatDropdown from './Dropdowns/ChatDropdown.vue';
import type { Ref } from 'vue';

const globalState = inject('globalState') as {
  activeChatMenu: Ref<string | null>,
  toggleChatMenu: (chatId: string, event: Event) => void
  showProfileMenu: Ref<boolean>,
  handleClickOutside: () => void,
  isAuthenticated: Ref<boolean>,
  planStatus: Ref<{ status: string; timeLeft: string; expiryDate: string; isExpired: boolean; }>,
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
const {
  activeChatMenu,
  toggleChatMenu,
  showProfileMenu,
  handleClickOutside,
  isAuthenticated,
  planStatus,
  syncStatus,
} = globalState

const props = defineProps<{
  data: {
    chats: Chat[]
    currentChatId: string
    isCollapsed?: boolean
    parsedUserDetails: {
      username: string
      email: string
      sync_enabled: boolean,
      plan_name: string,
      expiry_timestamp?: number
    }
    screenWidth: number,
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
const isRenaming = ref<string | null>(null)
const renameValue = ref('')
const now = ref(Date.now())

// Timer for real-time updates
let timer: number | null = null

onMounted(() => {
  timer = window.setInterval(() => {
    now.value = Date.now()
  }, 1000)
})

onUnmounted(() => {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
})

const planColor = computed(() => {
  if (planStatus.value.isExpired) return 'text-red-600 bg-red-50'
  if (planStatus.value.status === 'no-plan') return 'text-gray-600 bg-gray-50'
  return 'text-green-600 bg-green-50'
})

const profileOptions = [
  { id: 'settings', label: 'Settings', action: () => router.push('/settings/profile') },
  {
    id: 'help', label: 'Get help', action: () => {
      window.open('mailto:imranmat254@gmail.com', '_blank')
    }
  },
  {
    id: 'upgrade',
    label: props.data.parsedUserDetails?.plan_name ? 'Manage Plan' : 'Upgrade Plan',
    action: () => router.push('/upgrade')
  },
  { id: 'learn', label: 'Learn more', action: () => { /* Add your learn more action */ } }
];

function startRename(chatId: string, currentTitle: string) {
  isRenaming.value = chatId
  renameValue.value = currentTitle
  activeChatMenu.value = null

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
  <div id="side_nav" :class="[
    props.data.screenWidth > 720
      ? props.data.isCollapsed
        ? 'w-[60px] z-30 fixed top-0 left-0 bottom-0 border-r flex flex-col transition-all duration-300 ease-in-out'
        : 'w-[270px] z-30 fixed top-0 left-0 bottom-0 border-r flex flex-col transition-all duration-300 ease-in-out'
      : 'none',
    'bg-white dark:bg-gray-900 dark:border-gray-700'
  ]" @click="handleClickOutside">

    <!-- Scrollable area -->
    <div class="flex-1 overflow-y-auto custom-scrollbar">
      <!-- Top Header -->
      <div @click="() => {
        if (props.data.screenWidth < 720) props.functions.hideSidebar()
      }" class="flex items-center justify-between p-3">
        <p v-if="!props.data.isCollapsed" class="font-semibold text-xl text-black dark:text-white">
          Gemmie
        </p>
        <div class="flex gap-2 items-center ml-auto">
          <button @click="props.functions.toggleSidebar" title="Toggle Sidebar"
            class="w-8 h-8 flex items-center justify-center hover:bg-gray-100 dark:hover:bg-gray-800 rounded-full cursor-pointer transition-colors">
            <svg v-if="props.data.isCollapsed" width="20" height="20" viewBox="0 0 20 20" fill="currentColor" xmlns="http://www.w3.org/2000/svg"
              class="text-gray-500 dark:text-gray-400 transition-colors" aria-hidden="true">
              <path
                d="M16.5 4C17.3284 4 18 4.67157 18 5.5V14.5C18 15.3284 17.3284 16 16.5 16H3.5C2.67157 16 2 15.3284 2 14.5V5.5C2 4.67157 2.67157 4 3.5 4H16.5ZM7 15H16.5C16.7761 15 17 14.7761 17 14.5V5.5C17 5.22386 16.7761 5 16.5 5H7V15ZM3.5 5C3.22386 5 3 5.22386 3 5.5V14.5C3 14.7761 3.22386 15 3.5 15H6V5H3.5Z">
              </path>
            </svg>

            <svg v-else width="20" height="20" viewBox="0 0 20 20" fill="currentColor"
              xmlns="http://www.w3.org/2000/svg" class="text-gray-500 dark:text-gray-400 transition-colors"
              aria-hidden="true">
              <path
                d="M3.5 3C3.77614 3 4 3.22386 4 3.5V16.5L3.99023 16.6006C3.94371 16.8286 3.74171 17 3.5 17C3.25829 17 3.05629 16.8286 3.00977 16.6006L3 16.5V3.5C3 3.22386 3.22386 3 3.5 3ZM11.2471 5.06836C11.4476 4.95058 11.7104 4.98547 11.8721 5.16504C12.0338 5.34471 12.0407 5.60979 11.9023 5.79688L11.835 5.87207L7.80371 9.5H16.5C16.7761 9.5 17 9.72386 17 10C17 10.2761 16.7761 10.5 16.5 10.5H7.80371L11.835 14.1279C12.0402 14.3127 12.0568 14.6297 11.8721 14.835C11.6873 15.0402 11.3703 15.0568 11.165 14.8721L6.16504 10.3721L6.09473 10.2939C6.03333 10.2093 6 10.1063 6 10C6 9.85828 6.05972 9.72275 6.16504 9.62793L11.165 5.12793L11.2471 5.06836Z">
              </path>
            </svg>
            <!-- <svg  width="20" height="20" viewBox="0 0 20 20" fill="currentColor"
              xmlns="http://www.w3.org/2000/svg" class="text-gray-500 dark:text-gray-400 transition-colors"
              aria-hidden="true">
              <path
                d="M16.5 3C16.7761 3 17 3.22386 17 3.5V16.5L16.9902 16.6006C16.9437 16.8286 16.7417 17 16.5 17C16.2583 17 16.0563 16.8286 16.0098 16.6006L16 16.5V3.5C16 3.22386 16.2239 3 16.5 3ZM8.12793 5.16504C8.28958 4.98547 8.5524 4.95058 8.75293 5.06836L8.83496 5.12793L13.835 9.62793C13.9403 9.72275 14 9.85828 14 10C14 10.1063 13.9667 10.2093 13.9053 10.2939L13.835 10.3721L8.83496 14.8721C8.62972 15.0568 8.31267 15.0402 8.12793 14.835C7.94322 14.6297 7.95984 14.3127 8.16504 14.1279L12.1963 10.5H3.5C3.22386 10.5 3 10.2761 3 10C3 9.72386 3.22386 9.5 3.5 9.5H12.1963L8.16504 5.87207L8.09766 5.79688C7.95931 5.60979 7.96622 5.34471 8.12793 5.16504Z">
              </path>
            </svg> -->
          </button>
        </div>
      </div>


      <!-- New Chat & Actions -->
      <div v-if="props.data.parsedUserDetails.username"
        class="px-3 mb-4 mt-2 max-md:text-lg flex flex-col gap-1 font-light text-sm">
        <button @click="
          () => {
            router.push('/new')
            if (props.data.screenWidth < 720) props.functions.hideSidebar()
          }
        " title="New Chat"
          class="w-full flex items-center gap-2 h-[40px] hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg px-2 transition-colors">
          <i class="pi pi-plus text-gray-500 dark:text-gray-400 mb-[2px]"></i>
          <p v-if="!props.data.isCollapsed || props.data.screenWidth < 720" class="dark:text-gray-200">New Chat</p>
        </button>

        <div v-if="isAuthenticated">
          <!-- Sync button -->
          <button title="Sync Data" @click="props.functions.manualSync"
            :disabled="syncStatus.syncing || !props.data.parsedUserDetails.sync_enabled"
            class="w-full flex items-center gap-2 h-[40px] hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg px-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors">
            <i :class="[
              syncStatus.syncing ? 'pi pi-spin pi-spinner' : 'pi pi-refresh',
              'dark:text-gray-400'
            ]"></i>
            <span v-if="!props.data.isCollapsed || props.data.screenWidth < 720" class="dark:text-gray-200">
              {{ syncStatus.syncing ? 'Syncing...' : 'Sync Data' }}
            </span>
            <div
              v-if="syncStatus.hasUnsyncedChanges && props.data.parsedUserDetails.sync_enabled && (!props.data.isCollapsed || props.data.screenWidth < 720)"
              class="ml-auto w-2 h-2 bg-orange-500 dark:bg-orange-400 rounded-full"></div>
          </button>

          <button v-if="props.data.screenWidth > 720" @click="
            () => {
              if (router.currentRoute.value.path !== '/editor') {
                router.push('/editor')
              }
              if (props.data.screenWidth < 720) props.functions.hideSidebar()
            }
          " title="Open Editor"
            class="w-full flex items-center gap-2 h-[40px] hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg px-2 transition-colors">
            <i class="pi pi-pencil text-gray-500 dark:text-gray-400 mb-[2px]"></i>
            <p v-if="!props.data.isCollapsed || props.data.screenWidth < 720" class="dark:text-gray-200">Editor</p>
          </button>
        </div>
      </div>

      <!-- Recent Chats -->
      <div v-if="props.data.chats.length && props.data.parsedUserDetails.username"
        class="flex flex-col px-2 mb-2 py-4 font-light">
        <p v-if="!props.data.isCollapsed || props.data.screenWidth < 720"
          class="text-base text-gray-600 dark:text-gray-400 mb-2">Chats</p>
        <div class="flex flex-col gap-2">
          <div v-for="chat in !props.data.isCollapsed ? props.data.chats : props.data.chats.slice(0, 1)" :key="chat.id"
            :class="chat.id === props.data.currentChatId
              ? 'w-full flex h-[32px] max-md:h-[36px] text-sm items-center bg-gray-300 dark:bg-gray-700 rounded-lg relative transition-colors'
              : 'w-full flex h-[32px] max-md:h-[36px] text-sm items-center hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg relative transition-colors'">

            <!-- Chat content area -->
            <div @click="() => {
              if (props.data.isCollapsed) {
                props.functions.toggleSidebar()
                return
              }
              handleChatClick(chat.id)
            }" class="flex max-md:text-lg items-center h-full flex-grow px-2 cursor-pointer">
              <i class="pi pi-comments mr-2 text-gray-500 dark:text-gray-400 mb-[2px]"></i>

              <!-- Chat title or rename input -->
              <div v-if="isRenaming === chat.id" class="flex-grow" @click.stop>
                <input :id="`rename-${chat.id}`" v-model="renameValue" @keyup.enter="submitRename(chat.id)"
                  @keyup.escape="cancelRename" @blur="submitRename(chat.id)"
                  class="w-full px-1 py-0.5 max-md:text-lg text-xs bg-white dark:bg-gray-800 dark:text-gray-200 border border-blue-500 dark:border-blue-400 rounded focus:outline-none" />
              </div>
              <p v-else-if="!props.data.isCollapsed || props.data.screenWidth < 720"
                class="truncate dark:text-gray-200">
                <span v-if="chat.title.length > 20">{{ `${chat.title.slice(0, 20)}..` || 'Untitled Chat' }}</span>
                <span v-else>{{ chat.title || 'Untitled Chat' }}</span>
              </p>
            </div>

            <!-- Menu button -->
            <div v-if="!props.data.isCollapsed || props.data.screenWidth < 720" @click="toggleChatMenu(chat.id, $event)"
              class="flex items-center justify-center h-full hover:bg-blue-600 hover:text-white dark:hover:bg-blue-500 rounded-r-lg flex-shrink px-3 cursor-pointer transition-colors">
              <i class="pi pi-ellipsis-h dark:text-gray-300"></i>
            </div>

            <ChatDropdown :data="{
              activeChatMenu,
              chat,
              screenWidth: props.data.screenWidth,
            }" :functions="{
              deleteChat: props.functions.deleteChat,
              startRename,
              hideSidebar: props.functions.hideSidebar
            }" />
          </div>
        </div>
      </div>
    </div>

    <!-- Fixed Bottom User Profile -->
    <div :class="[
      !props.data.isCollapsed ? 'border-t' : '',
      'border-gray-200 dark:border-gray-700 p-3 sticky bottom-0 bg-white dark:bg-gray-900'
    ]">
      <!-- Plan Status Quick View (Above Profile) -->
      <div
        v-if="props.data.parsedUserDetails.username && planStatus.status === 'active' && (!props.data.isCollapsed || props.data.screenWidth < 720)"
        class="mb-2 px-2 py-1 text-xs rounded transition-colors" :class="planStatus.isExpired
          ? 'text-red-600 bg-red-50 dark:bg-red-900/20 dark:text-red-400'
          : 'text-green-600 bg-green-50 dark:bg-green-900/20 dark:text-green-400'">
        <div class="flex items-center justify-between">
          <span class="font-medium">{{ planStatus.timeLeft }}</span>
          <i class="pi pi-clock"></i>
        </div>
      </div>

      <div class="flex items-center justify-between cursor-pointer mr-1" @click.stop="() => {
        if (!props.data.isCollapsed || props.data.screenWidth < 720) {
          showProfileMenu = !showProfileMenu
        }
      }">
        <div class="flex items-center gap-2">
          <div
            class="w-[35px] h-[35px] flex justify-center items-center bg-gray-300 dark:bg-gray-700 rounded-full relative">
            <span @click="() => {
              if (props.data.isCollapsed) {
                props.functions.toggleSidebar()
              }
            }" class="text-sm max-md:text-lg dark:text-gray-200">{{
              props.data.parsedUserDetails.username.toUpperCase().slice(0, 2) }}</span>
            <!-- Plan status indicator -->
            <div v-if="planStatus.isExpired"
              class="absolute -top-1 -right-1 w-3 h-3 bg-red-500 dark:bg-red-400 rounded-full border-2 border-white dark:border-gray-900">
            </div>
            <div v-else-if="planStatus.status === 'active'"
              class="absolute -top-1 -right-1 w-3 h-3 bg-green-500 dark:bg-green-400 rounded-full border-2 border-white dark:border-gray-900">
            </div>
          </div>
          <div v-if="!props.data.isCollapsed || props.data.screenWidth < 720">
            <p class="text-base max-md:text-lg font-light dark:text-gray-200">
              {{ props.data.parsedUserDetails.username }}
            </p>
            <p v-if="props.data.parsedUserDetails.plan_name" class="text-xs text-gray-500 dark:text-gray-400">
              {{ props.data.parsedUserDetails.plan_name }}
            </p>
          </div>
        </div>
        <i class="pi pi-chevron-up text-xs max-md:text-base dark:text-gray-300"
          v-if="showProfileMenu && (!props.data.isCollapsed || props.data.screenWidth < 720)"></i>
        <i class="pi pi-chevron-down text-xs max-md:text-base dark:text-gray-300"
          v-else-if="!props.data.isCollapsed || props.data.screenWidth < 720"></i>
      </div>

      <!-- Profile Dropdown -->
      <transition name="fade">
        <div v-if="showProfileMenu"
          class="absolute max-w-[245px] max-md:text-base bottom-full left-3 right-3 mb-2 bg-white dark:bg-gray-800 border dark:border-gray-700 rounded-lg shadow-lg text-sm z-50"
          @click.stop>
          <div class="px-4 py-2 border-b dark:border-gray-700">
            <p class="text-gray-500 dark:text-gray-400"
              v-if="props.data.parsedUserDetails.email.split('@')[0].length < 12">
              {{ props.data.parsedUserDetails.email || 'No email' }}
            </p>
            <p v-else class="text-gray-500 dark:text-gray-400">
              {{
                `${props.data.parsedUserDetails.email.split('@')[0].slice(0,
                  12)}...@${props.data.parsedUserDetails.email.split('@')[1]}`
                || 'No email' }}
            </p>
            <!-- Plan info in dropdown -->
            <div v-if="props.data.parsedUserDetails.plan_name" class="mt-1 text-xs"
              :class="planStatus.isExpired ? 'text-red-600 dark:text-red-400' : 'text-green-600 dark:text-green-400'">
              {{ props.data.parsedUserDetails.plan_name }}
              <span v-if="planStatus.status === 'active'">- {{ planStatus.timeLeft }}</span>
              <span v-else-if="planStatus.isExpired">- Expired</span>
            </div>
          </div>
          <button v-for="option in profileOptions" :key="option.id" @click="() => {
            option.action()
            if (props.data.screenWidth < 720) props.functions.hideSidebar()
          }"
            class="w-full text-left px-4 py-2 hover:bg-gray-100 dark:hover:bg-gray-700 dark:text-gray-200 transition-colors">
            {{ option.label }}
          </button>
          <button @click="() => {
            props.functions.logout();
            if (props.data.screenWidth < 720) props.functions.hideSidebar()
          }"
            class="w-full text-left px-4 py-2 text-red-600 dark:text-red-400 hover:bg-red-100 dark:hover:bg-red-900/20 rounded-b-lg transition-colors">
            Log Out
          </button>
        </div>
      </transition>
    </div>
  </div>
</template>