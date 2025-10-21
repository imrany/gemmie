<script lang="ts" setup>
import type { Chat } from '@/types'
import { ref, inject, computed, onMounted, onUnmounted, nextTick, type Ref } from 'vue';
import { useRouter } from 'vue-router';
import ChatDropdown from './Dropdowns/ChatDropdown.vue';
import ProfileDropdown from './Dropdowns/ProfileDropdown.vue';
import { WalletCards } from 'lucide-vue-next'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"

const globalState = inject('globalState') as {
  currentChatId: Ref<string>,
  activeChatMenu: Ref<string | null>,
  toggleChatMenu: (chatId: string, evenet: Event) => void,
  showProfileMenu: Ref<boolean>,
  handleClickOutside: () => void,
  isAuthenticated: Ref<boolean>,
  planStatus: Ref<{
    status: string, timeLeft: string, expiryDate: string, isExpired: boolean
  }>,
  syncStatus: Ref<{
    lastSync: Date | null,
    syncing: boolean,
    hasUnsyncedChanges: boolean,
    lastError: string | null,
    retryCount: number,
    maxRetries: number,
    showSyncIndicator: boolean,
    syncMessage: string,
    syncProgress: number
  }>,
  hideSidebar: () => void,
  isSidebarHidden: Ref<boolean>,
  screenWidth: Ref<number>,
};
const {
  activeChatMenu,
  toggleChatMenu,
  showProfileMenu,
  handleClickOutside,
  isAuthenticated,
  planStatus,
  syncStatus,
  hideSidebar,
  isSidebarHidden,
  screenWidth,
  currentChatId
} = globalState

const props = defineProps<{
  data: {
    chats: Chat[]
    isCollapsed?: boolean
    parsedUserDetails: {
      username: string
      email: string
      syncEnabled: boolean,
      planName?: string,
      expiryTimestamp?: number
    }
  }
  functions: {
    setShowInput: () => void
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
const hoveredChatId = ref<string | null>(null)

let timer: number | null = null

onMounted(() => {
  timer = window.setInterval(() => now.value = Date.now(), 1000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

// Computed
const showFullSidebar = computed(() => !props.data.isCollapsed || screenWidth.value < 720)
const planColor = computed(() => {
  if (planStatus.value.isExpired) return 'text-red-600 bg-red-50 dark:bg-red-900/20 dark:text-red-400'
  if (planStatus.value.status === 'no-plan') return 'text-gray-600 bg-gray-50'
  return 'text-green-600 bg-green-50 dark:bg-green-900/20 dark:text-green-400'
})

const sidebarIconClass = computed(() => {
  const classes = ['text-gray-500 dark:text-gray-400'];

  if (screenWidth.value > 720) {
    classes.push(props.data.isCollapsed ? 'pi pi-align-justify' : 'pi pi-align-left');
  } else {
    classes.push('pi pi-times text-lg font-bold');
  }

  return classes;
});

// Constants
const profileOptions = [
  { id: 'settings', label: 'Settings', action: () => router.push('/settings/general') },
  { id: 'help', label: 'Get help', action: () => window.open('mailto:imranmat254@gmail.com', '_blank') },
  {
    id: 'upgrade',
    label: props.data.parsedUserDetails?.planName ? 'Manage Plan' : 'Upgrade Plan',
    action: () => router.push('/upgrade')
  },
  { id: 'learn', label: 'Learn more', action: () => { } }
]

// Methods
function startRename(chatId: string, currentTitle: string) {
  isRenaming.value = chatId
  renameValue.value = currentTitle
  activeChatMenu.value = null

  nextTick(() => {
    const input = document.getElementById(`rename-${chatId}`) as HTMLInputElement
    input?.focus()
    input?.select()
  })
}

function openWorkplace() {
  window.open('/workplace', '_blank')
  if (screenWidth.value < 720) hideSidebar()
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
  if (router.currentRoute.value.path !== '/') router.push('/')
  if (screenWidth.value < 720) hideSidebar()
}

function handleNavAction(action: () => void) {
  if (screenWidth.value < 720) hideSidebar()
  action()
}

const handleSidebarToggle = () => {
  if (screenWidth.value > 720) {
    props.functions.toggleSidebar();
  } else {
    hideSidebar();
  }
};

const navLinks =[
  {
    label: "New Chat",
    description: "New Chat",
    icon: '<i class="pi pi-plus text-gray-500 dark:text-gray-400"></i>',
    action:()=> handleNavAction(() => router.push('/new'))
  },
  {
    label: "Chats",
    description: "Recent Chats",
    icon: '<i class="pi pi-comments text-gray-500 dark:text-gray-400"></i>',
    action:()=> handleNavAction(() => router.push('/chats'))
  }
]
</script>

<template>
  <div id="side_nav" :class="[
    // Mobile styles
    screenWidth <= 720
      ? isSidebarHidden
        ? 'w-0 opacity-0 -translate-x-full'
        : 'w-full opacity-100 translate-x-0'
      : '',

    // Desktop styles  
    screenWidth > 720
      ? props.data.isCollapsed
        ? 'w-[60px]'
        : 'w-[270px]'
      : '',

    // Base styles
    'border-r z-40 fixed top-0 left-0 bottom-0 flex flex-col',
    'bg-gray-100 dark:bg-gray-800 dark:border-gray-700',

    // Animation classes
    'transition-all duration-300 ease-in-out transform select-none'
  ]" @click="handleClickOutside">
    <!-- Scrollable area -->
    <div class="flex-1 overflow-y-auto custom-scrollbar">
      <!-- Top Header -->
      <div class="flex items-center p-3">
        <p v-if="showFullSidebar"
          class="text-gray-700 dark:text-gray-300 text-xl max-md:text-2xl font-semibold tracking-wide select-none">
          Gemmie
        </p>

        <div class="flex ml-auto gap-2 items-center justify-center">
          <div v-if="isAuthenticated && screenWidth < 720" class="relative">
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

          <button @click="handleSidebarToggle" title="Toggle Sidebar"
            class="w-8 h-8 flex items-center justify-center hover:bg-gray-100 dark:hover:bg-gray-800 rounded-full cursor-pointer transition-colors">
            <i :class="sidebarIconClass"></i>
          </button>
        </div>
      </div>

      <!-- Navigation Menu -->
      <div v-if="props.data.parsedUserDetails.username"
        class="px-3 mb-4 mt-2 max-md:text-lg flex flex-col gap-1 font-light text-sm">
       
        <div v-for="navlink in navLinks">
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger as-child>
                <button @click="navlink.action"
                  class="w-full font-normal flex items-center gap-2 h-[40px] hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg px-2 transition-colors">
                  <span v-html="navlink.icon"></span>
                  <span v-if="showFullSidebar" class="dark:text-gray-200">{{navlink.label }}</span>
                </button>
              </TooltipTrigger>
              <TooltipContent v-if="!showFullSidebar" side="right" :avoid-collisions="true">
                <p>{{ navlink.description }}</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>

        
        <div v-if="isAuthenticated">
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger as-child>
                <button v-if="screenWidth > 720" @click="openWorkplace"
                  class="w-full font-normal flex items-center gap-2 h-[40px] hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg px-2 transition-colors">
                  <i class="pi pi-pencil text-gray-500 dark:text-gray-400"></i>
                  <span v-if="showFullSidebar" class="dark:text-gray-200">Workplace</span>
                </button>
              </TooltipTrigger>
              <TooltipContent v-if="!showFullSidebar" side="right" :avoid-collisions="true">
                <p>Workplace</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>

          <!-- Sync button -->
           <TooltipProvider>
            <Tooltip>
              <TooltipTrigger as-child>
                <button @click="props.functions.manualSync"
                  :disabled="syncStatus.syncing || !props.data.parsedUserDetails.syncEnabled"
                  class="w-full font-normal flex items-center gap-2 h-[40px] hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg px-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors">
                  <i :class="[
                    syncStatus.syncing ? 'pi pi-spin pi-spinner' : 'pi pi-refresh',
                    'text-gray-500 dark:text-gray-400'
                  ]"></i>
                  <span v-if="showFullSidebar" class="dark:text-gray-200">
                    {{ syncStatus.syncing ? 'Syncing...' : 'Sync Data' }}
                  </span>
                  <div v-if="syncStatus.hasUnsyncedChanges && props.data.parsedUserDetails.syncEnabled && showFullSidebar"
                    class="ml-auto w-2 h-2 bg-orange-500 dark:bg-orange-400 rounded-full"></div>
                </button>
              </TooltipTrigger>
              <TooltipContent v-if="!showFullSidebar" side="right" :avoid-collisions="true">
                <p>Sync Data</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>
      </div>

      <!-- Recent Chats -->
      <div v-if="props.data.chats.length && props.data.parsedUserDetails.username && showFullSidebar"
        class="flex flex-col px-2 mb-2 py-4">
        <p class="text-sm text-gray-600 dark:text-gray-400 mb-2">Chats</p>
        <div class="flex flex-col gap-2">
          <div v-for="chat in props.data.chats" :key="chat.id" @mouseover="hoveredChatId = chat.id"
            @mouseleave="hoveredChatId = null" :class="[
              'w-full flex h-[32px] max-md:h-[36px] text-sm items-center rounded-lg relative transition-colors',
              chat.id === currentChatId
                ? 'bg-gray-300 dark:bg-gray-700'
                : 'hover:bg-gray-300 dark:hover:bg-gray-700'
            ]">
            <!-- Chat content area -->
            <div @click="() => handleChatClick(chat.id)"
              class="flex max-md:text-lg items-center h-full flex-grow px-2 cursor-pointer relative">
              <div v-if="isRenaming === chat.id" class="flex-grow" @click.stop>
                <input :id="`rename-${chat.id}`" v-model="renameValue" @keyup.enter="submitRename(chat.id)"
                  @keyup.escape="cancelRename" @blur="submitRename(chat.id)"
                  class="w-full px-1 py-0.5 max-md:text-lg text-xs bg-white dark:bg-gray-800 dark:text-gray-200 border border-blue-500 dark:border-blue-400 rounded focus:outline-none" />
              </div>
              <div v-else class="truncate dark:text-gray-200 font-normal relative flex-grow">
                {{ chat.title.slice(0, 25) || 'Untitled Chat' }}{{ chat.title.length > 25 ? '..' : '' }}
              </div>
            </div>

            <!-- Menu button -->
            <div v-if="showFullSidebar && (currentChatId === chat.id || hoveredChatId === chat.id)"
              @click="toggleChatMenu(chat.id, $event)"
              class="flex items-center justify-center h-full pr-1 rounded-r-lg flex-shrink cursor-pointer transition-colors">
              <div :class="[hoveredChatId === chat.id ?
                ' text-white bg-gray-500 dark:bg-gray-900'
                : 'bg-transparent',
                'rounded-md text-center p-1 w-6 h-6 flex items-center justify-center'
              ]">
                <i class="pi pi-ellipsis-h dark:text-gray-300 text-xs"></i>
              </div>
            </div>

            <ChatDropdown :data="{ activeChatMenu, chat, screenWidth }"
              :functions="{ deleteChat: props.functions.deleteChat, startRename, hideSidebar }" />
          </div>
        </div>
      </div>
    </div>

    <!-- Fixed Bottom User Profile -->
    <div :class="[
      screenWidth > 720
        ? (showFullSidebar ? 'border-t' : '')
        : isSidebarHidden ? 'none' : '',
      'border-gray-200 dark:border-gray-700 p-3 sticky bottom-0 bg-gray-100 dark:bg-gray-800'
    ]">
      <!-- Plan Status -->
      <div v-if="props.data.parsedUserDetails.username && planStatus.status === 'active' && showFullSidebar"
        class="mb-2 px-2 py-1 text-xs rounded transition-colors" :class="planColor">
        <div class="flex items-center justify-between">
          <span class="font-normal">{{ planStatus.timeLeft }}</span>
          <i class="pi pi-clock"></i>
        </div>
      </div>

      <!-- Profile -->
      <div class="flex items-center justify-between cursor-pointer mr-1" @click.stop="() => {
        if (props.data.isCollapsed && screenWidth > 720) {
          handleSidebarToggle()
        }
        showProfileMenu = !showProfileMenu
      }">
        <div class="flex items-center gap-2">
          <div
            class="w-[35px] h-[35px] flex justify-center items-center bg-gray-300 dark:bg-gray-700 rounded-full relative">
            <span class="text-sm max-md:text-lg dark:text-gray-200">
              {{ props.data.parsedUserDetails.username.toUpperCase().slice(0, 2) }}
            </span>
            <!-- Plan status indicator -->
            <div v-if="planStatus.isExpired"
              class="absolute -top-1 -right-1 w-3 h-3 bg-red-500 dark:bg-red-400 rounded-full border-2 border-white dark:border-gray-900">
            </div>
            <div v-else-if="planStatus.status === 'active'"
              class="absolute -top-1 -right-1 w-3 h-3 bg-green-500 dark:bg-green-400 rounded-full border-2 border-white dark:border-gray-900">
            </div>
          </div>
          <div v-if="showFullSidebar">
            <p class="text-base max-md:text-lg font-light dark:text-gray-200">
              {{ props.data.parsedUserDetails.username }}
            </p>
            <p v-if="props.data.parsedUserDetails.planName" class="text-xs text-gray-500 dark:text-gray-400">
              {{ props.data.parsedUserDetails.planName }}
            </p>
          </div>
        </div>
        <i v-if="showFullSidebar" :class="[
          showProfileMenu ? 'pi pi-chevron-up' : 'pi pi-chevron-down',
          'text-xs max-md:text-base dark:text-gray-300'
        ]"></i>
      </div>

      <!-- Profile Dropdown -->
      <ProfileDropdown :data="{
        planColor,
        profileOptions,
        showProfileMenu
      }" :functions="{
        handleNavAction,
        logout: props.functions.logout,
      }" />
    </div>
  </div>
</template>