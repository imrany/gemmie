 <script setup lang="ts">
import { computed, onMounted, onUnmounted, provide, ref, watch, type ComputedRef } from 'vue';
import { toast, Toaster } from 'vue-sonner'
import 'vue-sonner/style.css'
import type { Chat, ConfirmDialogOptions, CurrentChat, LinkPreview, UserDetails } from './types';
import { API_BASE_URL, generateChatId, generateChatTitle, extractUrls, validateCredentials, getTransaction, WRAPPER_URL, detectLargePaste, SPINDLE_URL } from './utils/globals';
import { nextTick } from 'vue';
import { detectAndProcessVideo } from './utils/videoProcessing';
import ConfirmDialog from './components/ConfirmDialog.vue';
import type { Theme } from 'vue-sonner/src/packages/types.js';
import UpdateModal from './components/Modals/UpdateModal.vue';

const isUserOnline = ref(navigator.onLine)
const connectionStatus = ref<'online' | 'offline' | 'checking'>('online')
const screenWidth = ref(screen.width)
const isDarkMode = ref(false)
const scrollableElem = ref<HTMLElement | null>(null)
const showScrollDownButton = ref(false)
const activeRequests = ref<Map<string, AbortController>>(new Map())
const requestChatMap = ref<Map<string, string>>(new Map())
const pendingResponses = ref<Map<string, { prompt: string; chatId: string }>>(new Map())
const chatDrafts = ref<Map<string, string>>(new Map())
const pastePreviews = ref<Map<string, {
  content: string;
  wordCount: number;
  charCount: number;
  show: boolean;
}>>(new Map())

const confirmDialog = ref<ConfirmDialogOptions>({
  visible: false,
  title: "",
  message: "",
  type: undefined,
  confirmText: "",
  cancelText: "",
  onConfirm: () => { },
  onCancel: () => { }
})
const isCollapsed = ref(false)
const isSidebarHidden = ref(true)
const authData = ref({
  username: '',
  email: '',
  password: '',
  agreeToTerms: false
})

const syncStatus = ref({
  lastSync: null as Date | null,
  syncing: false,
  hasUnsyncedChanges: false,
  lastError: null as string | null,
  retryCount: 0,
  maxRetries: 3,
  showSyncIndicator: false,
  syncMessage: '',
  syncProgress: 0
})

const isOpenTextHighlightPopover = ref(false)

const currentChat: ComputedRef<CurrentChat | undefined> = computed(() => {
  return chats.value.find(chat => chat.id === currentChatId.value)
})

const currentMessages = computed(() => {
  return currentChat.value?.messages || []
})

const parsedUserDetailsNullValues: UserDetails = {
  agreeToTerms: false,
  createdAt: new Date,
  email: '',
  emailSubscribed: false,
  emailVerified: false,
  userId: '',
  syncEnabled: true,
  username:"",
  theme: 'system'
}

const linkPreviewCache = ref<Map<string, LinkPreview>>(new Map())

function loadLinkPreviewCache() {
  try {
    const cached = localStorage.getItem('linkPreviews')
    if (cached) {
      const parsedCache = JSON.parse(cached)
      if (typeof parsedCache === 'object' && parsedCache !== null) {
        linkPreviewCache.value = new Map(Object.entries(parsedCache))
      }
    }
  } catch (error) {
    console.error('Failed to load link preview cache:', error)
    localStorage.removeItem('linkPreviews')
    linkPreviewCache.value.clear()
  }
}

function saveLinkPreviewCache() {
  try {
    const cacheObject = Object.fromEntries(linkPreviewCache.value)
    localStorage.setItem('linkPreviews', JSON.stringify(cacheObject))
  } catch (error) {
    console.error('Failed to save link preview cache:', error)
    if (linkPreviewCache.value.size > 100) {
      const entries = Array.from(linkPreviewCache.value.entries())
      const recent = entries.slice(-50)
      linkPreviewCache.value = new Map(recent)
      try {
        const reducedCache = Object.fromEntries(linkPreviewCache.value)
        localStorage.setItem('linkPreviews', JSON.stringify(reducedCache))
      } catch (retryError) {
        console.error('Failed to save reduced cache:', retryError)
      }
    }
  }
}

let userDetails: any = localStorage.getItem("userdetails")
const parsedUserDetails = ref<UserDetails>(
  (() => {
    try {
      return userDetails ? JSON.parse(userDetails) : null
    } catch (error) {
      console.error('Invalid user details in localStorage:', error)
      localStorage.removeItem("userdetails")
      return null
    }
  })()
)

const syncEnabled = ref(parsedUserDetails.value?.syncEnabled !== false)

const isAuthenticated = computed(() => {
  const user = parsedUserDetails.value
  if (!user || typeof user !== 'object') return false

  return !!(
    user.email &&
    user.username &&
    user.sessionId &&
    /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(user.email)
  )
})

const currentChatId = ref<string>('')
const chats = ref<Chat[]>([])
const isLoading = ref(false)
const expanded = ref<boolean[]>([])
const showInput = ref(false)
const activeChatMenu = ref<string | null>(null)
const showProfileMenu = ref(false)
const now = ref(Date.now())

const planStatus = computed(() => {
  if (!parsedUserDetails.value || !parsedUserDetails.value.expiryTimestamp) {
    return { status: 'no-plan', timeLeft: '', expiryDate: '', isExpired: false }
  }

  const expiryMs = parsedUserDetails.value.expiryTimestamp < 1e12
    ? parsedUserDetails.value.expiryTimestamp * 1000
    : parsedUserDetails.value.expiryTimestamp

  const diff = expiryMs - now.value
  const isExpired = diff <= 0

  if (isExpired) {
    return { status: 'expired', timeLeft: 'Expired', expiryDate: '', isExpired: true }
  }

  const days = Math.floor(diff / (1000 * 60 * 60 * 24))
  const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60))
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60))

  let timeLeft = ''
  if (days > 0) {
    timeLeft = `${days}d ${hours}h ${minutes}m`
  } else if (hours > 0) {
    timeLeft = `${hours}h ${minutes}m`
  } else {
    timeLeft = `${minutes}m`
  }

  const expiryDate = new Date(expiryMs).toLocaleString('en-KE', {
    weekday: 'short',
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })

  return { status: 'active', timeLeft, expiryDate, isExpired: false }
})

const FREE_REQUEST_LIMIT = 5

const userPlanStatus = computed(() => {
  if (!parsedUserDetails.value || !parsedUserDetails.value.expiryTimestamp) {
    return { status: 'no-plan', isExpired: false }
  }

  const expiryMs = parsedUserDetails.value.expiryTimestamp < 1e12
    ? parsedUserDetails.value.expiryTimestamp * 1000
    : parsedUserDetails.value.expiryTimestamp

  const diff = expiryMs - now.value
  const isExpired = diff <= 0

  if (isExpired) {
    return { status: 'expired', isExpired: true }
  }

  return { status: 'active', isExpired: false }
})

// computed property to determine if user has limits
const userHasRequestLimits = computed(() => {
  if (!parsedUserDetails.value) return true
  
  const hasFreePlan = !parsedUserDetails.value.plan ||
    parsedUserDetails.value.plan === 'free' ||
    parsedUserDetails.value.plan === '' ||
    userPlanStatus.value.status === 'no-plan'

  return hasFreePlan || userPlanStatus.value.isExpired
})

// Consolidated request limit computations
const requestLimitInfo = computed(() => {
  const hasLimits = userHasRequestLimits.value
  const currentCount = parsedUserDetails.value.requestCount?.count || 0
  
  return {
    // Core limits
    hasLimits,
    currentCount,
    limit: FREE_REQUEST_LIMIT,
    
    // Derived states
    isExceeded: hasLimits && currentCount >= FREE_REQUEST_LIMIT,
    shouldShowUpgradePrompt: hasLimits && currentCount >= 3 && currentCount < FREE_REQUEST_LIMIT,
    remaining: hasLimits ? Math.max(0, FREE_REQUEST_LIMIT - currentCount) : Infinity,
    
    // Status messages
    status: hasLimits ? 
      (currentCount >= FREE_REQUEST_LIMIT ? 'exceeded' : 
       currentCount >= 3 ? 'warning' : 'normal') : 'unlimited'
  }
})

const isRequestLimitExceeded = computed(() => requestLimitInfo.value.isExceeded)
const shouldShowUpgradePrompt = computed(() => requestLimitInfo.value.shouldShowUpgradePrompt)
const requestsRemaining = computed(() => requestLimitInfo.value.remaining)
const isFreeUser = computed(() => userHasRequestLimits.value)
const requestCount = computed(() => requestLimitInfo.value.currentCount)

function showSyncIndicator(message: string, progress: number = 0) {
  syncStatus.value.showSyncIndicator = true
  syncStatus.value.syncMessage = message
  syncStatus.value.syncProgress = progress
}

function hideSyncIndicator() {
  syncStatus.value.showSyncIndicator = false
  syncStatus.value.syncMessage = ''
  syncStatus.value.syncProgress = 0
}

function updateSyncProgress(message: string, progress: number) {
  syncStatus.value.syncMessage = message
  syncStatus.value.syncProgress = progress
}

function showConfirmDialog(options: ConfirmDialogOptions) {
  confirmDialog.value = {
    visible: true,
    title: options.title,
    message: options.message,
    type: options.type || 'info',
    confirmText: options.confirmText || 'Confirm',
    cancelText: options.cancelText || 'Cancel',
    onConfirm: () => {
      try {
        options.onConfirm()
      } catch (error) {
        console.error('Error in confirm callback:', error)
        toast.error('An error occurred while processing your request')
      }
      confirmDialog.value.visible = false
    },
    onCancel: () => {
      try {
        options.onCancel?.()
      } catch (error) {
        console.error('Error in cancel callback:', error)
      }
      confirmDialog.value.visible = false
    }
  }
}

async function apiCall(endpoint: string, options: RequestInit = {}, retryCount = 0): Promise<any> {
  if (!isUserOnline.value) {
    const isActuallyOnline = await checkInternetConnection()
    if (!isActuallyOnline) {
      throw new Error('No internet connection. Please check your network and try again.')
    }
  }

  const maxRetries = 3
  const retryDelay = Math.pow(2, retryCount) * 1000

  try {
    if (!parsedUserDetails.value?.userId && !endpoint.includes('/login') && !endpoint.includes('/register')) {
      throw new Error('User not authenticated')
    }

    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), 30000)

    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...(parsedUserDetails.value?.userId ? { 'X-User-ID': parsedUserDetails.value.userId } : {}),
        ...options.headers,
      },
      signal: controller.signal
    })

    clearTimeout(timeoutId)

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`)
    }

    const data = await response.json()

    if (!data.success) {
      throw new Error(data.message || 'API request failed')
    }

    syncStatus.value.retryCount = 0
    syncStatus.value.lastError = null

    return data
  } catch (error: any) {
    console.error(`API Error on ${endpoint}:`, error)

    if (error.name === 'AbortError') {
      throw new Error('Request timeout - please try again')
    }

    if ((error.name === 'NetworkError' || error.name === 'TypeError' || error.name === 'TimeoutError') &&
      retryCount < maxRetries) {
      console.log(`Retrying ${endpoint} in ${retryDelay}ms (attempt ${retryCount + 1}/${maxRetries})`)

      await new Promise(resolve => setTimeout(resolve, retryDelay))
      return apiCall(endpoint, options, retryCount + 1)
    }

    if (endpoint.includes('/sync')) {
      syncStatus.value.lastError = error.message
      syncStatus.value.retryCount = retryCount
    }

    throw error
  }
}

function isValidChat(chat: any): chat is Chat {
  return chat &&
    typeof chat === 'object' &&
    chat.id &&
    typeof chat.id === 'string' &&
    Array.isArray(chat.messages) &&
    chat.updatedAt &&
    typeof chat.updatedAt === 'string'
}

function isNewerChat(serverChat: Chat, localChat: Chat): boolean {
  try {
    return new Date(serverChat.updatedAt).getTime() > new Date(localChat.updatedAt).getTime()
  } catch (error) {
    return false
  }
}

function mergeChats(serverChats: Chat[], localChats: Chat[]): Chat[] {
  try {
    if (!Array.isArray(serverChats)) serverChats = []
    if (!Array.isArray(localChats)) localChats = []

    const merged = new Map<string, Chat>()

    localChats.forEach(chat => {
      if (isValidChat(chat)) {
        merged.set(chat.id, chat)
      }
    })

    serverChats.forEach(serverChat => {
      if (!isValidChat(serverChat)) return

      const localChat = merged.get(serverChat.id)
      if (!localChat || isNewerChat(serverChat, localChat)) {
        merged.set(serverChat.id, serverChat)
      }
    })

    return Array.from(merged.values()).sort((a, b) => {
      try {
        return new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()
      } catch (error) {
        return 0
      }
    })
  } catch (error) {
    console.error('Error merging chats:', error)
    return localChats || []
  }
}

function deleteChat(chatId: string) {
  if (isLoading.value || !chatId) return

  try {
    const chatIndex = chats.value.findIndex(chat => chat.id === chatId)
    if (chatIndex === -1) {
      toast.error('Chat not found')
      return
    }

    const chatToDelete = chats.value[chatIndex]
    const chatTitle = chatToDelete.title || 'Untitled Chat'
    const messageCount = chatToDelete.messages?.length || 0

    showConfirmDialog({
      visible: true,
      title: 'Delete Chat',
      message: `Are you sure you want to delete "${chatTitle}"?\n\nThis will permanently remove ${messageCount} message(s). This action cannot be undone.`,
      type: 'danger',
      confirmText: 'Delete',
      onConfirm: () => {
        try {
          // Remove link previews from cache before deleting
          if (chatToDelete.messages?.length > 0) {
            chatToDelete.messages.forEach(message => {
              try {
                const responseUrls = extractUrls(message.response || '')
                const promptUrls = extractUrls(message.prompt || '')
                const urls = [...new Set([...responseUrls, ...promptUrls])]

                urls.forEach(url => {
                  linkPreviewCache.value.delete(url)
                })
              } catch (error) {
                console.error('Error extracting URLs for cache cleanup:', error)
              }
            })
            saveLinkPreviewCache()
          }

          // Remove chat from array
          chats.value.splice(chatIndex, 1)

          // If we deleted the current chat, switch to another one
          if (currentChatId.value === chatId) {
            if (chats.value.length > 0) {
              currentChatId.value = chats.value[0].id
            } else {
              currentChatId.value = ''
            }
            updateExpandedArray()
          }

          clearCurrentDraft()
          saveChats()

          // Trigger sync after deleting chat
          if (isAuthenticated.value && parsedUserDetails.value?.syncEnabled) {
            setTimeout(() => {
              performSmartSync()
            }, 1000)
          }
        } catch (error) {
          console.error('Error deleting chat:', error)
          toast.error('Failed to delete chat')
        }
      }
    })
  } catch (error) {
    console.error('Error in deleteChat:', error)
    toast.error('Failed to delete chat')
  }
}

async function logout() {
  showConfirmDialog({
    visible: true,
    title: 'Logout Confirmation',
    message: 'Are you sure you want to logout?' + 
    (parsedUserDetails.value?.syncEnabled ? '' : ' Your unsynced data will permanently lost.'),
    type: 'warning',
    confirmText: 'Logout',
    cancelText: 'Cancel',
    onConfirm: async () => {
      try {
        isLoading.value = true

        const syncEnabled = parsedUserDetails.value?.syncEnabled
        const hasUnsyncedChanges = syncStatus.value.hasUnsyncedChanges

        if (hasUnsyncedChanges && syncEnabled && !syncStatus.value.syncing) {
          try {
            showSyncIndicator('Syncing your data before logout...', 50)
            await syncToServer()
            hideSyncIndicator()
          } catch (syncError) {
            console.error('Sync failed during logout:', syncError)
            hideSyncIndicator()
          }
        }

        const userBackup = { ...parsedUserDetails.value }
        const hasChats = chats.value.length > 0

        try {
          chats.value = []
          currentChatId.value = ''
          expanded.value = []
          showInput.value = false
          isCollapsed.value = false

          syncStatus.value = {
            lastSync: null,
            syncing: false,
            hasUnsyncedChanges: false,
            lastError: null,
            retryCount: 0,
            maxRetries: 3,
            showSyncIndicator: false,
            syncMessage: '',
            syncProgress: 0
          }

          parsedUserDetails.value = parsedUserDetailsNullValues

          let keysToRemove = [
            'userdetails', 'chatDrafts', 'pastePreviews', 'linkPreviews'
          ]
          if (syncEnabled) {
            keysToRemove.push('chats', 'currentChatId')
          } else {
            keysToRemove.push('chats', 'currentChatId')
          }

          linkPreviewCache.value.clear()
          
          keysToRemove.forEach(key => {
            try {
              localStorage.removeItem(key)
            } catch (error) {
              console.error(`Failed to remove ${key} from localStorage:`, error)
            }
          })
        } catch (stateError) {
          console.error('Error clearing application state:', stateError)
          try {
            parsedUserDetails.value = userBackup
            if (!syncEnabled) {
              loadLocalData()
            }
          } catch (restoreError) {
            console.error('Failed to restore state after logout error:', restoreError)
          }

          throw new Error('Failed to clear application state during logout')
        }

        if (syncEnabled) {
          toast.success('Logged out successfully', {
            duration: 3000,
            description: hasUnsyncedChanges
              ? 'Your data has been synced to the cloud'
              : 'Ready to log back in anytime'
          })
        } else {
          toast.success('Logged out successfully', {
            duration: 3000,
            description: hasChats
              ? 'Your chats are saved locally on this device'
              : 'Ready to start fresh when you return'
          })
        }

      } catch (error: any) {
        console.error('Critical error during logout:', error)
        toast.error('Error during logout process', {
          duration: 5000,
          description: 'Some cleanup operations may not have completed. Please refresh the page.'
        })
      } finally {
        isLoading.value = false
      }
    }
  })
}

function setShowInput() {
  if (currentMessages.value.length !== 0) {
    return
  }
  if (!isAuthenticated.value) {
    toast.warning('Please create a session first', {
      duration: 3000,
      description: 'You need to be logged in.'
    })
    return
  }
  showInput.value = true
  nextTick(() => {
    const textarea = document.getElementById("prompt") as HTMLTextAreaElement
    if (textarea) textarea.focus()
  })
}

function updateExpandedArray() {
  try {
    const messagesLength = currentMessages.value?.length || 0
    expanded.value = new Array(messagesLength).fill(false)
  } catch (error) {
    console.error('Error updating expanded array:', error)
    expanded.value = []
  }
}

function scrollToBottom(behavior: ScrollBehavior = 'smooth') {
  if (!scrollableElem.value) return;

  try {
    nextTick(() => {
      setTimeout(() => {
        if (!scrollableElem.value) return;

        const container = scrollableElem.value;
        const scrollHeight = container.scrollHeight;
        const clientHeight = container.clientHeight;

        if (scrollHeight > clientHeight) {
          container.scrollTo({
            top: scrollHeight,
            behavior
          });
        }

        setTimeout(() => {
          handleScroll();
        }, 150);
      }, 50);
    });
  } catch (error) {
    console.error('Error scrolling to bottom:', error);
  }
}

function scrollToLastMessage() {
  if (!scrollableElem.value) return
  
  nextTick(() => {
    const messages = scrollableElem.value?.querySelectorAll('.chat-message')
    if (messages && messages.length > 0) {
      const lastMessage = messages[messages.length - 2] as HTMLElement // Get user's prompt (second to last)
      if (lastMessage) {
        // Scroll so the last message pair starts at the top with some padding
        const offsetTop = lastMessage.offsetTop - 10 // 10px padding from top
        scrollableElem.value?.scrollTo({
          top: offsetTop,
          behavior: 'smooth'
        })
      }
    }
  })
}

function handleScroll() {
  try {
    console.log('scrolling')
    if (isOpenTextHighlightPopover.value) {
      isOpenTextHighlightPopover.value = false
    }

    const elem = scrollableElem.value;
    if (!elem) return;

    const scrollTop = elem.scrollTop;
    const scrollHeight = elem.scrollHeight;
    const clientHeight = elem.clientHeight;

    const currentScrollPosition = scrollTop + clientHeight;
    const totalScrollableHeight = scrollHeight;

    const threshold = 2;
    const isAtBottom = Math.abs(currentScrollPosition - totalScrollableHeight) <= threshold;

    const hasSubstantialContent = scrollHeight > clientHeight + 100;
    showScrollDownButton.value = !isAtBottom && hasSubstantialContent;

  } catch (error) {
    console.error('Error handling scroll:', error);
  }
}

function hideSidebar() {
  try {
    isSidebarHidden.value = !isSidebarHidden.value
  } catch (error) {
    console.error('Error toggling sidebar:', error)
  }
}

function toggleChatMenu(chatId: string, event: Event) {
  try {
    event.stopPropagation()
    activeChatMenu.value = activeChatMenu.value === chatId ? null : chatId
  } catch (error) {
    console.error('Error toggling chat menu:', error)
  }
}

function autoGrow(e: Event) {
  const el = e.target as HTMLTextAreaElement
  const maxHeight = 200
  el.style.height = "auto"
  if (el.scrollHeight <= maxHeight) {
    el.style.height = el.scrollHeight + "px"
    el.style.overflowY = "hidden"
  } else {
    el.style.height = maxHeight + "px"
    el.style.overflowY = "auto"
  }

  autoSaveDraft()
}

function saveChatDrafts() {
  try {
    const draftsObject = Object.fromEntries(chatDrafts.value)
    localStorage.setItem('chatDrafts', JSON.stringify(draftsObject))

    const previewsObject = Object.fromEntries(pastePreviews.value)
    localStorage.setItem('pastePreviews', JSON.stringify(previewsObject))
  } catch (error) {
    console.error('Failed to save chat drafts:', error)
  }
}

function loadChatDrafts() {
  try {
    const saved = localStorage.getItem('chatDrafts')
    if (saved) {
      const parsedDrafts = JSON.parse(saved)
      chatDrafts.value = new Map(Object.entries(parsedDrafts))
    }

    const savedPastePreviews = localStorage.getItem('pastePreviews')
    if (savedPastePreviews) {
      const parsedPreviews = JSON.parse(savedPastePreviews)
      pastePreviews.value = new Map(Object.entries(parsedPreviews))
    }

    if (currentChatId.value) {
      const currentDraft = chatDrafts.value.get(currentChatId.value) || ''
      const currentPastePreview = pastePreviews.value.get(currentChatId.value)

      let shouldFocus = false

      if (currentPastePreview && currentPastePreview.show) {
        const textarea = document.getElementById('prompt') as HTMLTextAreaElement
        if (textarea) {
          const draftWithoutPaste = currentDraft.replace(currentPastePreview.content, '')
          textarea.value = draftWithoutPaste
          autoGrow({ target: textarea } as any)
          shouldFocus = true
          console.log('📝 Focus: Paste preview detected')
        }
      } else if (currentDraft && detectLargePaste(currentDraft)) {
        const wordCount = currentDraft.trim().split(/\s+/).filter(word => word.length > 0).length
        const charCount = currentDraft.length

        pastePreviews.value.set(currentChatId.value, {
          content: currentDraft,
          wordCount,
          charCount,
          show: true
        })

        const textarea = document.getElementById('prompt') as HTMLTextAreaElement
        if (textarea) {
          textarea.value = ''
          autoGrow({ target: textarea } as any)
          shouldFocus = true
          console.log('📝 Focus: Large paste detected and converted to preview')
        }
      } else if (currentDraft.trim()) {
        const textarea = document.getElementById('prompt') as HTMLTextAreaElement
        if (textarea) {
          textarea.value = currentDraft
          autoGrow({ target: textarea } as any)
          shouldFocus = true
          console.log('📝 Focus: Draft content detected')
        }
        pastePreviews.value.delete(currentChatId.value)
      } else {
        const textarea = document.getElementById('prompt') as HTMLTextAreaElement
        if (textarea) {
          textarea.value = ''
          autoGrow({ target: textarea } as any)
          // Don't focus on empty drafts
          console.log('📝 No focus: Empty draft')
        }
        pastePreviews.value.delete(currentChatId.value)
      }

      // Focus the textarea if we have content to work with
      if (shouldFocus) {
        nextTick(() => {
          const textarea = document.getElementById('prompt') as HTMLTextAreaElement
          if (textarea) {
            // Small delay to ensure DOM is updated
            setTimeout(() => {
              textarea.focus()
              console.log('🎯 Textarea focused due to draft/preview content')
            }, 100)
          }
        })
      }
    }
  } catch (error) {
    console.error('Failed to load chat drafts:', error)
  }
}

function clearCurrentDraft() {
  if (currentChatId.value) {
    chatDrafts.value.delete(currentChatId.value)
    pastePreviews.value.delete(currentChatId.value)
    saveChatDrafts()

    const textarea = document.getElementById('prompt') as HTMLTextAreaElement
    if (textarea) {
      textarea.value = ''
      autoGrow({ target: textarea } as any)
    }
  }
}

let draftSaveTimeout: any = null

function autoSaveDraft() {
  if (draftSaveTimeout) {
    clearTimeout(draftSaveTimeout)
  }

  draftSaveTimeout = setTimeout(() => {
    if (currentChatId.value) {
      const textarea = document.getElementById('prompt') as HTMLTextAreaElement
      let currentDraft = textarea?.value || ''

      const currentPastePreview = pastePreviews.value.get(currentChatId.value)
      if (currentPastePreview?.show) {
        currentDraft += currentPastePreview.content
      }

      if (currentDraft.trim().length > 0) {
        chatDrafts.value.set(currentChatId.value, currentDraft)
        saveChatDrafts()
      } else {
        chatDrafts.value.delete(currentChatId.value)
        saveChatDrafts()
      }
    }
  }, 1000)
}

function switchToChat(chatId: string) {
  try {
    if (!chatId || typeof chatId !== 'string') {
      console.error('Invalid chat ID provided')
      return
    }

    const chatExists = chats.value.find(chat => chat.id === chatId)
    if (!chatExists) {
      toast.error('Chat not found')
      return
    }

    if (currentChatId.value) {
      const textarea = document.getElementById('prompt') as HTMLTextAreaElement
      let currentDraft = textarea?.value || ''

      const currentPastePreview = pastePreviews.value.get(currentChatId.value)
      if (currentPastePreview?.show && currentPastePreview.content) {
        currentDraft += currentPastePreview.content
      }

      if (currentDraft.trim().length === 0) {
        chatDrafts.value.delete(currentChatId.value)
        pastePreviews.value.delete(currentChatId.value)
      } else {
        chatDrafts.value.set(currentChatId.value, currentDraft)
      }
      saveChatDrafts()
    }

    currentChatId.value = chatId
    updateExpandedArray()

    try {
      localStorage.setItem('currentChatId', currentChatId.value)
    } catch (error) {
      console.error('Failed to save current chat ID:', error)
    }

    nextTick(() => {
      loadChatDrafts()
      // scrollToLastMessage()
    })
  } catch (error) {
    console.error('Error switching to chat:', error)
    toast.error('Failed to switch to chat')
  }
}

function createNewChat(firstMessage?: string): string {
  try {
    const newChatId = generateChatId()
    const now = new Date().toISOString()

    const newChat: Chat = {
      id: newChatId,
      title: firstMessage ? generateChatTitle(firstMessage) : 'New Chat',
      messages: [],
      createdAt: now,
      updatedAt: now
    }

    if (currentChatId.value) {
      pastePreviews.value.delete(currentChatId.value)
    }

    if (currentChatId.value) {
      const textarea = document.getElementById('prompt') as HTMLTextAreaElement
      const currentDraft = textarea?.value || ''
      if (currentDraft.trim()) {
        chatDrafts.value.set(currentChatId.value, currentDraft)
        saveChatDrafts()
      }
    }

    chatDrafts.value.set(newChatId, '')
    pastePreviews.value.delete(newChatId)

    chats.value.unshift(newChat)
    currentChatId.value = newChatId
    updateExpandedArray()
    saveChats()

    // Trigger sync after creating new chat
    if (isAuthenticated.value && parsedUserDetails.value?.syncEnabled) {
      setTimeout(() => {
        performSmartSync()
      }, 1000)
    }

    nextTick(() => {
      const textarea = document.getElementById('prompt') as HTMLTextAreaElement
      if (textarea) {
        textarea.value = ''
        autoGrow({ target: textarea } as any)
        textarea.focus()
      }
    })

    return newChatId
  } catch (error) {
    console.error('Error creating new chat:', error)
    toast.error('Failed to create new chat')
    return ''
  }
}

//  deleteMessage function
function deleteMessage(messageIndex: number) {
  if (isLoading.value || !currentChat.value) return

  try {
    if (messageIndex < 0 || messageIndex >= currentChat.value.messages.length) {
      toast.error('Invalid message')
      return
    }

    const message = currentChat.value.messages[messageIndex]
    const messageContent = message?.prompt || message?.response || 'this message'
    const preview = messageContent.slice(0, 50) + (messageContent.length > 50 ? '...' : '')

    showConfirmDialog({
      visible: true,
      title: 'Delete Message',
      message: `Are you sure you want to delete this message?\n"${preview}"\n\nThis action cannot be undone.`,
      type: 'danger',
      confirmText: 'Delete',
      onConfirm: () => {
        try {
          if (!currentChat.value || messageIndex >= currentChat.value.messages.length) {
            toast.error('Message no longer exists')
            return
          }

          const messageToDelete = currentChat.value.messages[messageIndex]
          const responseUrls = extractUrls(messageToDelete.response || '')
          const promptUrls = extractUrls(messageToDelete.prompt || '')
          const urls = [...new Set([...responseUrls, ...promptUrls])]

          currentChat.value.messages.splice(messageIndex, 1)
          expanded.value.splice(messageIndex, 1)

          currentChat.value.updatedAt = new Date().toISOString()

          if (messageIndex === 0 && currentChat.value.messages.length > 0) {
            const firstMessage = currentChat.value.messages[0].prompt || currentChat.value.messages[0].response
            if (firstMessage) {
              currentChat.value.title = generateChatTitle(firstMessage)
            }
          } else if (currentChat.value.messages.length === 0) {
            currentChat.value.title = 'New Chat'
          }

          if (urls.length > 0) {
            urls.forEach(url => {
              linkPreviewCache.value.delete(url)
            })
            saveLinkPreviewCache()
          }

          saveChats()

          // Trigger sync after deleting message
          if (isAuthenticated.value && parsedUserDetails.value?.syncEnabled) {
            setTimeout(() => {
              performSmartSync()
            }, 1000)
          }
        } catch (error) {
          console.error('Error deleting message:', error)
          toast.error('Failed to delete message')
        }
      }
    })
  } catch (error) {
    console.error('Error in deleteMessage:', error)
    toast.error('Failed to delete message')
  }
}

//  renameChat function
function renameChat(chatId: string, newTitle: string) {
  try {
    if (!chatId || !newTitle || typeof newTitle !== 'string') {
      toast.error('Invalid chat title')
      return
    }

    const chat = chats.value.find(c => c.id === chatId)
    if (!chat) {
      toast.error('Chat not found')
      return
    }

    const trimmedTitle = newTitle.trim()
    if (!trimmedTitle) {
      toast.error('Chat title cannot be empty')
      return
    }

    chat.title = trimmedTitle
    chat.updatedAt = new Date().toISOString()
    saveChats()

    // Trigger sync after renaming chat
    if (isAuthenticated.value && parsedUserDetails.value?.syncEnabled) {
      setTimeout(() => {
        performSmartSync()
      }, 1000)
    }
  } catch (error) {
    console.error('Error renaming chat:', error)
    toast.error('Failed to rename chat')
  }
}

function clearAllChats() {
  if (isLoading.value) return

  try {
    const totalChats = chats.value.length
    const totalMessages = chats.value.reduce((sum, chat) => sum + (chat.messages?.length || 0), 0)

    if (totalChats === 0) {
      toast.info('There are no chats to clear', {
        duration: 3000,
        description: 'Your chat list is already empty.'
      })
      return
    }

    showConfirmDialog({
      visible: true,
      title: 'Clear All Chats',
      message: `⚠️ DELETE ALL CHATS?\n\nThis will permanently delete:\n• ${totalChats} chat(s)\n• ${totalMessages} total message(s)\n\nThis action cannot be undone!`,
      type: 'danger',
      confirmText: 'Delete All',
      onConfirm: () => {
        try {
          chats.value = []
          currentChatId.value = ''
          expanded.value = []
          linkPreviewCache.value.clear()

          chatDrafts.value.clear()
          saveChatDrafts()

          const keysToRemove = ['chats', 'currentChatId', 'linkPreviews', 'chatDrafts']
          keysToRemove.forEach(key => {
            try {
              localStorage.removeItem(key)
            } catch (error) {
              console.error(`Failed to remove ${key} from localStorage:`, error)
            }
          })

          saveChats()
          toast.error(`${totalChats} chats with ${totalMessages} messages deleted`, {
            duration: 5000,
            description: 'All chat data has been cleared'
          })
        } catch (error) {
          console.error('Error clearing all chats:', error)
          toast.error('Failed to clear all chats')
        }
      }
    })
  } catch (error) {
    console.error('Error in clearAllChats:', error)
    toast.error('Failed to clear chats')
  }
}

async function fetchLinkPreview(url: string): Promise<LinkPreview> {
  try {
    new URL(url)
  } catch (error) {
    console.error('Invalid URL provided:', url)
    return {
      url,
      title: 'Invalid URL',
      domain: 'Invalid',
      loading: false,
      error: true
    }
  }

  if (linkPreviewCache.value.has(url)) {
    return linkPreviewCache.value.get(url)!
  }

  const preview: LinkPreview = { url, loading: true }
  linkPreviewCache.value.set(url, preview)

  try {
    const lang = "en"
    const proxyUrl = `${SPINDLE_URL}/scrape?url=${encodeURIComponent(url)}&lang=${lang}`

    const response = await fetch(proxyUrl, {
      signal: AbortSignal.timeout(15000)
    })

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`)
    }

    const results = await response.json()
    const domain = new URL(url).hostname

    let videoInfo: any = {}
    try {
      videoInfo = await detectAndProcessVideo(url, results)
    } catch (videoError) {
      console.error('Error processing video:', videoError)
    }

    const updatedPreview: LinkPreview = {
      url,
      title: results.title?.slice(0, 100) || domain,
      description: results.description?.slice(0, 200) || "",
      images: Array.isArray(results.images) ? results.images : [],
      previewImage: videoInfo.thumbnail || results.preview_image || results.images?.[0] || "",
      domain,
      favicon: results.favicon || `https://www.google.com/s2/favicons?domain=${domain}`,
      links: Array.isArray(results.links) ? results.links : [],
      video: videoInfo.videoUrl || results.video || "",
      videoType: videoInfo.type,
      videoDuration: videoInfo.duration,
      videoThumbnail: videoInfo.thumbnail,
      embedHtml: videoInfo.embedHtml,
      loading: false,
      error: false
    }

    linkPreviewCache.value.set(url, updatedPreview)
    saveLinkPreviewCache()
    return updatedPreview
  } catch (error) {
    console.error("Failed to fetch link preview:", error)

    const fallbackPreview: LinkPreview = {
      url,
      title: new URL(url).hostname,
      domain: new URL(url).hostname,
      loading: false,
      error: true
    }

    linkPreviewCache.value.set(url, fallbackPreview)
    saveLinkPreviewCache()
    return fallbackPreview
  }
}

function toggleSidebar() {
  try {
    isCollapsed.value = !isCollapsed.value
    localStorage.setItem("isCollapsed", String(isCollapsed.value))
  } catch (error) {
    console.error('Error toggling sidebar:', error)
  }
}

function handleClickOutside() {
  try {
    activeChatMenu.value = null
    showProfileMenu.value = false
  } catch (error) {
    console.error('Error handling click outside:', error)
  }
}

function saveChats() {
  try {
    if (!Array.isArray(chats.value)) {
      console.error('Chats is not an array, resetting to empty array')
      chats.value = []
    }

    const chatsJson = JSON.stringify(chats.value)
    const currentChatJson = JSON.stringify(currentChatId.value)

    if (chatsJson.length > 5000000) {
      toast.warning('Chat data is getting large', {
        duration: 5000,
        description: 'Consider clearing old chats to improve performance'
      })
    }

    localStorage.setItem('chats', chatsJson)
    localStorage.setItem('currentChatId', currentChatId.value)

    if (isAuthenticated.value && parsedUserDetails.value?.syncEnabled) {
      syncStatus.value.hasUnsyncedChanges = true
    }
  } catch (error) {
    console.error('Failed to save chats:', error)
    toast.error('Failed to save chat data', {
      duration: 4000,
      description: 'Your recent changes may not be saved'
    })
  }
}

async function syncFromServer(serverData?: any) {
  if (!parsedUserDetails.value?.userId) {
    console.log('❌ syncFromServer: No user ID')
    return
  }

  const shouldSync = parsedUserDetails.value?.syncEnabled !== false || serverData
  if (!shouldSync) {
    console.log('❌ syncFromServer: Sync disabled')
    return
  }

  try {
    syncStatus.value.syncing = true
    syncStatus.value.lastError = null
    showSyncIndicator('Syncing data from server...', 30)

    let data = serverData
    if (!data) {
      console.log('📡 Fetching data from server...')
      updateSyncProgress('Fetching data from server...', 50)
      const response = await apiCall('/sync', { method: 'GET' })
      data = response.data
    }

    if (!data) {
      console.warn('⚠️ No data received from server')
      return
    }

    console.log('📥 Server data received:')

    updateSyncProgress('Processing chats...', 70)

    // Process chats
    if (data.chats && data.chats !== '[]') {
      try {
        const serverChatsData = typeof data.chats === 'string' ? JSON.parse(data.chats) : data.chats

        if (Array.isArray(serverChatsData)) {
          const localChats = chats.value
          const mergedChats = mergeChats(serverChatsData, localChats)

          chats.value = mergedChats
          localStorage.setItem('chats', JSON.stringify(mergedChats))
          console.log(`✅ Synced ${mergedChats.length} chats from server`)
        }
      } catch (parseError) {
        console.error('❌ Error parsing server chats:', parseError)
      }
    } else {
      console.log('📭 No chats data from server')
    }

    updateSyncProgress('Processing link previews...', 85)

    // Process link previews
    if (data.link_previews && data.link_previews !== '{}') {
      try {
        const serverPreviewsData = typeof data.link_previews === 'string'
          ? JSON.parse(data.link_previews)
          : data.link_previews

        if (typeof serverPreviewsData === 'object' && serverPreviewsData !== null) {
          const localPreviews = Object.fromEntries(linkPreviewCache.value)
          const mergedPreviews = { ...localPreviews, ...serverPreviewsData }

          linkPreviewCache.value = new Map(Object.entries(mergedPreviews))
          localStorage.setItem('linkPreviews', JSON.stringify(mergedPreviews))
          console.log(`✅ Synced ${Object.keys(mergedPreviews).length} link previews from server`)
        }
      } catch (parseError) {
        console.error('❌ Error parsing server link previews:', parseError)
      }
    }

    // Process current chat ID
    if (data.current_chat_id && typeof data.current_chat_id === 'string') {
      const chatExists = chats.value.some(chat => chat.id === data.current_chat_id)
      if (chatExists) {
        currentChatId.value = data.current_chat_id
        localStorage.setItem('currentChatId', data.current_chat_id)
        console.log(`✅ Set current chat ID: ${data.current_chat_id}`)
      }
    }

    updateSyncProgress('Updating preferences...', 95)

    // Update user details if provided
    if (data.syncEnabled !== undefined || data.preferences || data.theme ||
      data.work_function || data.phone_number || data.plan) {

      const updatedUserDetails: UserDetails = {
        ...parsedUserDetails.value,
        preferences: data.preferences || parsedUserDetails.value.preferences,
        theme: data.theme || parsedUserDetails?.value?.theme,
        workFunction: data.work_function || parsedUserDetails.value.workFunction,
        phoneNumber: data.phone_number || parsedUserDetails.value.phoneNumber,
        plan: data.plan || parsedUserDetails.value.plan,
        planName: data.plan_name || parsedUserDetails.value.planName,
        amount: data.amount ?? parsedUserDetails.value.amount,
        duration: data.duration || parsedUserDetails.value.duration,
        price: data.price || parsedUserDetails.value.price,
        responseMode: data.response_mode || parsedUserDetails.value.responseMode,
        requestCount : data.request_count || parsedUserDetails.value.requestCount,
        expiryTimestamp: data.expiry_timestamp || parsedUserDetails.value.expiryTimestamp,
        expireDuration: data.expire_duration || parsedUserDetails.value.expireDuration,
        syncEnabled: data.sync_enabled || parsedUserDetails.value.syncEnabled,
      }
      
      parsedUserDetails.value = updatedUserDetails
      localStorage.setItem("userdetails", JSON.stringify(updatedUserDetails))

      parsedUserDetails.value.userTransactions = data.user_transactions || []
      console.log('✅ User details updated from server')
    }

    syncStatus.value.lastSync = new Date()
    syncStatus.value.hasUnsyncedChanges = false
    syncStatus.value.retryCount = 0
    updateExpandedArray()

    updateSyncProgress('Sync complete!', 100)
    setTimeout(() => {
      hideSyncIndicator()
    }, 1000)

    console.log('✅ Successfully synced data from server')

  } catch (error: any) {
    console.error('❌ Sync from server failed:', error)
    syncStatus.value.lastError = error.message
    hideSyncIndicator()

    if (!error.message.includes('NetworkError') && !error.message.includes('TypeError')) {
      toast.warning('Failed to sync data from server', {
        duration: 3000,
        description: 'Using local data instead'
      })
    }

    throw error
  } finally {
    syncStatus.value.syncing = false
  }
}

async function syncToServer() {
  if (!parsedUserDetails.value?.userId || parsedUserDetails.value.syncEnabled === false) {
    return
  }

  try {
    syncStatus.value.syncing = true
    syncStatus.value.lastError = null
    showSyncIndicator('Syncing data to server...', 20)

    if (!Array.isArray(chats.value)) {
      throw new Error('Chats data is not valid')
    }

    if (!parsedUserDetails.value.username) {
      throw new Error('User data is incomplete')
    }

    updateSyncProgress('Preparing sync data...', 40)

    const syncData = {
      chats: JSON.stringify(chats.value),
      link_previews: JSON.stringify(Object.fromEntries(linkPreviewCache.value)),
      current_chat_id: currentChatId.value || '',
      username: parsedUserDetails.value.username,
      preferences: parsedUserDetails.value.preferences || '',
      work_function: parsedUserDetails.value.workFunction || '',
      theme: parsedUserDetails?.value?.theme || 'system',
      sync_enabled: parsedUserDetails?.value?.syncEnabled,
      response_mode: parsedUserDetails.value.responseMode || 'light-response',
      request_count: {
        count: parsedUserDetails.value.requestCount?.count || 0,
        timestamp: parsedUserDetails.value.requestCount?.timestamp || Date.now()
      }
    }

    const syncDataSize = JSON.stringify(syncData).length
    if (syncDataSize > 10000000) {
      throw new Error('Sync data is too large. Please clear some old chats.')
    }

    console.log(`Syncing ${chats.value.length} chats to server (${(syncDataSize / 1024).toFixed(1)}KB)`)

    updateSyncProgress('Sending data to server...', 70)

    const response = await apiCall('/sync', {
      method: 'POST',
      body: JSON.stringify(syncData)
    })

    syncStatus.value.lastSync = new Date()
    syncStatus.value.hasUnsyncedChanges = false
    syncStatus.value.retryCount = 0

    updateSyncProgress('Sync complete!', 100)
    setTimeout(() => {
      hideSyncIndicator()
    }, 1000)

    console.log('✅ Successfully synced data to server')

    return response

  } catch (error: any) {
    console.error('❌ Sync to server failed:', error)
    syncStatus.value.lastError = error.message
    syncStatus.value.hasUnsyncedChanges = true
    hideSyncIndicator()

    syncStatus.value.retryCount = Math.min(syncStatus.value.retryCount + 1, syncStatus.value.maxRetries)

    if (error.message.includes('too large')) {
      toast.error('Data too large to sync', {
        duration: 5000,
        description: error.message
      })
    } else if (!error.message.includes('NetworkError') && 
               !error.message.includes('TypeError') && 
               !error.message.includes('already in progress') &&
               !error.message.includes('AbortError')) {
      toast.error('Failed to sync data to server', {
        duration: 3000,
        description: 'Your data is saved locally'
      })
    }

    throw error
  } finally {
    syncStatus.value.syncing = false
  }
}

async function performSmartSync() {
  if (syncStatus.value.syncing) {
    console.log('⏳ Sync already in progress, skipping...')
    return
  }

  console.log('🔄 Performing SmartSync')

  if (!isAuthenticated.value || !parsedUserDetails.value?.userId) {
    console.log('❌ Sync skipped: not authenticated or no user ID')
    return
  }

  try {
    const isLocalDataEmpty = chats.value.length === 0 ||
      (chats.value.length === 1 && chats.value[0].messages.length === 0)

    if (isLocalDataEmpty) {
      console.log('📥 Local data empty - syncing FROM server')
      try {
        syncStatus.value.syncing = true
        await syncFromServer()
        console.log('✅ Successfully synced data from server')
      } catch (error) {
        console.error('❌ Failed to sync from server:', error)
      }
    } else if (syncStatus.value.hasUnsyncedChanges) {
      console.log('📤 Has unsynced changes - syncing TO server')

      const hadUnsyncedChangesBeforeSync = syncStatus.value.hasUnsyncedChanges

      // Clear the flag BEFORE attempting sync
      syncStatus.value.hasUnsyncedChanges = false

      try {
        syncStatus.value.syncing = true
        await syncToServer()
        console.log('✅ Successfully synced changes to server')

      } catch (error: any) {
        console.error('❌ Failed to sync to server:', error)

        if (hadUnsyncedChangesBeforeSync) {
          syncStatus.value.hasUnsyncedChanges = true
          console.log('🔄 Marked changes as unsynced due to sync failure')
        }

        // Auto-retry for network errors
        if (error.message?.includes('Network') || error.message?.includes('timeout')) {
          console.log('🔄 Network error - will retry sync')
          setTimeout(() => {
            performSmartSync().catch(console.error)
          }, 5000)
        }
      }
    } else {
      console.log('🔍 No unsynced changes - checking for server updates')
      try {
        syncStatus.value.syncing = true
        await syncFromServer()
        console.log('✅ Server data is current')
      } catch (error) {
        console.error('❌ Failed to check for server updates:', error)
      }
    }
  } catch (error) {
    console.error('💥 Critical error in performSmartSync:', error)
  } finally {
    console.log('🔓 Sync completed')
    syncStatus.value.syncing = false
  }
}

async function unsecureApiCall(endpoint: string, options: RequestInit = {}, retryCount = 0): Promise<any> {
  const maxRetries = 2
  const retryDelay = Math.pow(2, retryCount) * 1000

  try {
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), 15000)

    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      signal: controller.signal
    })

    clearTimeout(timeoutId)

    if (!response.ok) {
      const errorText = await response.text().catch(() => 'Unknown error')
      throw new Error(`HTTP ${response.status}: ${response.statusText} - ${errorText}`)
    }

    const data = await response.json()

    if (!data.success) {
      throw new Error(data.message || 'API request failed')
    }

    return data

  } catch (error: any) {
    console.error(`Unsecure API Error on ${endpoint} (attempt ${retryCount + 1}):`, error)

    if (error.name === 'AbortError') {
      throw new Error('Request timeout - please try again')
    }

    if ((error.name === 'NetworkError' ||
      error.name === 'TypeError' ||
      error.message?.includes('Failed to fetch')) &&
      retryCount < maxRetries) {

      console.log(`Retrying ${endpoint} in ${retryDelay}ms (attempt ${retryCount + 1}/${maxRetries})`)

      await new Promise(resolve => setTimeout(resolve, retryDelay))
      return unsecureApiCall(endpoint, options, retryCount + 1)
    }

    throw error
  }
}

async function handleAuth(data: {
  username: string
  email: string
  password: string
  agreeToTerms: boolean
}) {
  const { username, email, password, agreeToTerms } = data

  try {
    const validationError = validateCredentials(username, email, password, agreeToTerms)
    if (validationError) {
      throw new Error(validationError)
    }

    let response
    let isLogin = false

    try {
      console.log('Attempting login...')
      response = await unsecureApiCall('/login', {
        method: 'POST',
        body: JSON.stringify({ username, email, password, agree_to_terms: agreeToTerms })
      })
      isLogin = true

    } catch (loginError: any) {
      console.log('Login failed, attempting registration...')

      try {
        response = await unsecureApiCall('/register', {
          method: 'POST',
          body: JSON.stringify({ username, email, password, agree_to_terms: agreeToTerms })
        })

        toast.success('Account created successfully!', {
          duration: 3000,
          description: `Welcome ${response.data.username}!`
        })
      } catch (registerError: any) {
        if (loginError.message?.includes('Connection') || loginError.message?.includes('Network')) {
          throw loginError
        } else {
          throw registerError
        }
      }
    }
    
    if (!response || !response.data) {
      throw new Error('Invalid response from server')
    }

    if (isLogin) {
      toast.success('Welcome back!', {
        duration: 3000,
        description: `Logged in as ${response.data.username}`
      })
    }

    const userData: UserDetails = {
      userId: response.data.user_id,
      username: response.data.username,
      email: response.data.email,
      createdAt: response.data.created_at,
      sessionId: btoa(email + ':' + password + ':' + username),
      workFunction: response.data.work_function || "",
      preferences: response.data.preferences || "",
      theme: response.data.theme || "system",
      syncEnabled: response.data.sync_enabled,
      phoneNumber: response.data.phone_number || "",
      plan: response.data.plan || "free",
      planName: response.data.plan_name || "",
      amount: response.data.amount || 0,
      duration: response.data.duration || "",
      price: response.data.price || 0,
      responseMode: response.data.response_mode || "light-response",
      expiryTimestamp: response.data.expiry_timestamp || null,
      expireDuration: response.data.expire_duration || "",
      emailVerified: response.data.email_verified || false,
      emailSubscribed: response.data.email_subscribed || true,
    }

    // ✅ Set in-memory state first
    parsedUserDetails.value = userData

    console.log(`Authentication successful for user: ${userData.username} (sync: ${userData.syncEnabled})`)

    if (userData.syncEnabled) {
      try {
        // Sync first, then save
        await performSmartSync()
        console.log('Initial smart sync completed')

        // ✅ ONLY save to localStorage AFTER successful sync
        localStorage.setItem('userdetails', JSON.stringify(userData))
        console.log('User details saved locally after successful sync')

      } catch (syncError) {
        console.error('Initial sync failed:', syncError)

        // Reset unsynced changes flag
        syncStatus.value.hasUnsyncedChanges = false

        // Load local data as fallback
        loadLocalData()

        // Still save userData to localStorage as fallback
        localStorage.setItem('userdetails', JSON.stringify(userData))

        toast.warning('Synced with server but failed to merge data', {
          duration: 4000,
          description: 'Your local data is preserved'
        })
      }
    } else {
      // If sync is disabled, safe to save immediately
      localStorage.setItem('userdetails', JSON.stringify(userData))
      loadLocalData()
      console.log('Sync disabled, loaded local data only')
    }

    return response

  } catch (error: any) {
    console.error('Authentication error:', error)
    throw error
  }
}

function loadLocalData() {
  try {
    console.log('Loading data from localStorage...')

    const storedChats = localStorage.getItem('chats')
    if (storedChats) {
      try {
        const parsedChats = JSON.parse(storedChats)
        if (Array.isArray(parsedChats)) {
          const validChats = parsedChats.filter(chat =>
            chat &&
            typeof chat === 'object' &&
            chat.id &&
            typeof chat.id === 'string' &&
            Array.isArray(chat.messages)
          )
          chats.value = validChats
          console.log(`Loaded ${validChats.length} valid chats from localStorage`)
        } else {
          console.warn('Stored chats is not an array, resetting to empty')
          chats.value = []
        }
      } catch (parseError) {
        console.error('Error parsing stored chats:', parseError)
        chats.value = []
        localStorage.removeItem('chats')
      }
    }

    const storedChatId = localStorage.getItem('currentChatId')
    if (storedChatId && typeof storedChatId === 'string') {
      const chatExists = chats.value.some(chat => chat.id === storedChatId)
      if (chatExists) {
        currentChatId.value = storedChatId
      } else {
        console.warn('Stored currentChatId does not exist in chats, clearing')
        currentChatId.value = chats.value.length > 0 ? chats.value[0].id : ''
        localStorage.setItem('currentChatId', currentChatId.value)
      }
    }

    loadLinkPreviewCache()

    updateExpandedArray()
    console.log('Successfully loaded local data')

  } catch (error) {
    console.error('Error loading local data:', error)
    toast.error('Failed to load local data', {
      duration: 4000,
      description: 'Some data may not be available'
    })
  }
}

async function manualSync() {
  if (!isUserOnline.value) {
    const isActuallyOnline = await checkInternetConnection()
    if (!isActuallyOnline) {
      toast.error('Cannot sync while offline', {
        duration: 4000,
        description: 'Please check your internet connection'
      })
      return
    }
  }

  if (!parsedUserDetails.value?.userId) {
    toast.warning('Please log in to sync data', {
      duration: 3000,
      description: 'Authentication required for syncing'
    })
    return
  }

  if (parsedUserDetails.value.syncEnabled === false) {
    toast.info('Sync is disabled', {
      duration: 3000,
      description: 'Enable sync in settings to sync your data'
    })
    return
  }

  if (syncStatus.value.syncing) {
    toast.info('Sync already in progress', {
      duration: 2000,
      description: 'Please wait for current sync to complete'
    })
    return
  }

  try {
    console.log('Starting manual smart sync...')
    showSyncIndicator('Starting manual sync...', 10)

    await performSmartSync()
    console.log('Manual smart sync completed successfully')
  } catch (error: any) {
    console.error('Manual sync failed:', error)
    toast.error('Sync failed', {
      duration: 4000,
      description: 'Please check your internet connection and try again'
    })
  }
}

async function toggleSync() {
  const targetSyncValue = !parsedUserDetails.value.syncEnabled

  // Store original state for rollback
  const originalSyncValue = parsedUserDetails.value.syncEnabled
  const originalUserDetails = { ...parsedUserDetails.value }

  try {
    // Update in-memory state
    parsedUserDetails.value.syncEnabled = targetSyncValue
    syncEnabled.value = targetSyncValue

    console.log('Attempting sync toggle:', { current: syncEnabled.value, target: targetSyncValue })

    if (targetSyncValue) {
      // ENABLING sync - sync TO server first to preserve local data
      try {
        showSyncIndicator('Uploading your local data...', 20)
        
        // Always sync TO server first when enabling sync
        await syncToServer()
        
        console.log('Local data uploaded to server successfully')
        
        // Then optionally sync FROM server to get any additional server data
        updateSyncProgress('Checking for server updates...', 70)
        await syncFromServer()
        
        hideSyncIndicator()
        
        toast.success('Sync enabled and data synchronized', {
          duration: 3000,
          description: 'Your data is now syncing across devices'
        })
      } catch (error) {
        console.error('Failed to sync when enabling:', error)
        hideSyncIndicator()
        
        // Rollback on failure
        parsedUserDetails.value.syncEnabled = originalSyncValue
        syncEnabled.value = originalSyncValue
        parsedUserDetails.value = originalUserDetails
        localStorage.setItem("userdetails", JSON.stringify(originalUserDetails))
        
        toast.error('Failed to enable sync', {
          duration: 4000,
          description: 'Could not upload your data. Please try again.'
        })
        throw error
      }
    } else {
      // DISABLING sync - just update server preference
      try {
        showSyncIndicator('Updating sync preference...', 50)
        
        // Update server to disable sync (with empty data)
        await apiCall('/sync', {
          method: 'POST',
          body: JSON.stringify({
            username: parsedUserDetails.value.username,
            sync_enabled: false,
            chats: "[]",
            link_previews: "{}",
            current_chat_id: "",
          })
        })
        
        hideSyncIndicator()
        
        // Save to localStorage after successful server update
        localStorage.setItem('userdetails', JSON.stringify(parsedUserDetails.value))
        
        toast.info('Sync disabled', {
          duration: 3000,
          description: 'Your data will only be saved locally on this device'
        })
      } catch (error) {
        console.error('Failed to disable sync on server:', error)
        hideSyncIndicator()
        
        // Even if server update fails, allow local disable
        localStorage.setItem('userdetails', JSON.stringify(parsedUserDetails.value))
        
        toast.warning('Sync disabled locally', {
          duration: 3000,
          description: 'Server update failed, but sync is disabled on this device'
        })
      }
    }

  } catch (error) {
    console.error('Failed to toggle sync:', error)

    // Rollback: Revert both in-memory AND localStorage
    parsedUserDetails.value.syncEnabled = originalSyncValue
    syncEnabled.value = originalSyncValue
    parsedUserDetails.value = originalUserDetails
    localStorage.setItem("userdetails", JSON.stringify(originalUserDetails))

    toast.error('Failed to update sync setting', {
      duration: 4000,
      description: 'Changes have been reverted. Please try again.'
    })
    throw error
  }
}

function isLocalDataEmpty(): boolean {
  try {
    if (chats.value.length === 0) {
      return true
    }

    const hasMeaningfulData = chats.value.some(chat => {
      return chat.messages.length > 0 ||
        (chat.title && chat.title !== 'New Chat' && chat.title !== '')
    })

    return !hasMeaningfulData
  } catch (error) {
    console.error('Error checking local data state:', error)
    return false
  }
}

async function checkInternetConnection(): Promise<boolean> {
  try {
    connectionStatus.value = 'checking'

    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), 5000)

    const response = await fetch(`${API_BASE_URL}/health`, {
      method: 'GET',
      signal: controller.signal,
      cache: 'no-cache'
    })

    clearTimeout(timeoutId)

    const isConnected = response.status < 400
    isUserOnline.value = isConnected
    connectionStatus.value = isConnected ? 'online' : 'offline'

    return isConnected
  } catch (error) {
    console.warn('Internet connection check failed:', error)
    isUserOnline.value = false
    connectionStatus.value = 'offline'
    return false
  }
}

function showConnectionStatus() {
  if (!isUserOnline.value) {
    toast.error('You are offline', {
      duration: 4000,
      description: 'Please check your internet connection'
    })
  } else {
    toast.success('Connection restored', {
      duration: 3000,
      description: 'You are back online'
    })
  }
}

function cancelChatRequests(chatId: string) {
  const requestsToCancel: string[] = []

  requestChatMap.value.forEach((requestChatId, requestId) => {
    if (requestChatId === chatId) {
      requestsToCancel.push(requestId)
    }
  })

  requestsToCancel.forEach(requestId => {
    const controller = activeRequests.value.get(requestId)
    if (controller) {
      controller.abort()
      activeRequests.value.delete(requestId)
      requestChatMap.value.delete(requestId)
    }
  })
}

function cancelAllRequests() {
  activeRequests.value.forEach((controller, requestId) => {
    controller.abort()
  })
  activeRequests.value.clear()
  requestChatMap.value.clear()
}

const hasActiveRequestsForCurrentChat = computed(() => {
  let hasRequests = false
  requestChatMap.value.forEach((chatId) => {
    if (chatId === currentChatId.value) {
      hasRequests = true
    }
  })
  return hasRequests
})

function setupConnectionListeners() {
  window.addEventListener('online', async () => {
    console.log('Browser reports online, verifying...')
    const isActuallyOnline = await checkInternetConnection()
    if (isActuallyOnline) {
      showConnectionStatus()

      // Auto-sync when coming back online with retry logic
      if (syncStatus.value.hasUnsyncedChanges && parsedUserDetails.value?.syncEnabled) {
        console.log('Connection restored - syncing unsaved changes')
        setTimeout(() => {
          performSmartSync().catch(error => {
            console.error('Auto-sync after connection recovery failed:', error)
            // Don't show error toast for auto-sync failures
          })
        }, 3000)
      }
    }
  })

  window.addEventListener('offline', () => {
    console.log('Browser reports offline')
    isUserOnline.value = false
    connectionStatus.value = 'offline'
    showConnectionStatus()
  })

  let connectionCheckInterval: any
  const startConnectionMonitoring = () => {
    connectionCheckInterval = setInterval(async () => {
      if (!isUserOnline.value) {
        await checkInternetConnection()
        if (isUserOnline.value) {
          showConnectionStatus()
        }
      }
    }, 30000)
  }

  const stopConnectionMonitoring = () => {
    if (connectionCheckInterval) {
      clearInterval(connectionCheckInterval)
    }
  }

  if (!isUserOnline.value) {
    startConnectionMonitoring()
  }

  window.addEventListener('online', stopConnectionMonitoring)
  window.addEventListener('offline', startConnectionMonitoring)
}

// ---------- Request Limit Functions ----------
function loadRequestCount() {
  try {
    if (!parsedUserDetails.value) return

    // Initialize if doesn't exist
    if (!parsedUserDetails.value.requestCount) {
      parsedUserDetails.value.requestCount = { count: 0, timestamp: Date.now() }
      return
    }

    const data = parsedUserDetails.value.requestCount

    // Validate data structure
    if (typeof data !== 'object' || typeof data.timestamp !== 'number' || typeof data.count !== 'number') {
      console.warn('Invalid request count data format, resetting')
      parsedUserDetails.value.requestCount = { count: 0, timestamp: Date.now() }
      return
    }

    // Check if 24 hours have passed
    const now = Date.now()
    const timeDiff = now - data.timestamp
    const twentyFourHours = 24 * 60 * 60 * 1000

    if (timeDiff > twentyFourHours) {
      // Reset count after 24 hours
      parsedUserDetails.value.requestCount = { count: 0, timestamp: now }
    } else {
      // Ensure count doesn't exceed limit
      parsedUserDetails.value.requestCount.count = Math.max(0, Math.min(data.count, FREE_REQUEST_LIMIT))
    }
  } catch (error) {
    console.error('Failed to load request count:', error)
    if (parsedUserDetails.value) {
      parsedUserDetails.value.requestCount = { count: 0, timestamp: Date.now() }
    }
  }
}

function checkRequestLimitBeforeSubmit(): boolean {
  try {
    if (!userHasRequestLimits.value) {
      return true // Unlimited user - allow
    }

    if (!parsedUserDetails.value.requestCount) {
      return true // No limit data - allow (safety)
    }

    const currentCount = parsedUserDetails.value.requestCount?.count || 0

    if (currentCount >= FREE_REQUEST_LIMIT) {
      // Show error toast here (at submit time, not load time)
      if (userPlanStatus.value.isExpired) {
        toast.error('Your plan has expired', {
          duration: 5000,
          description: 'Please renew your plan to continue using the service.'
        })
      } else {
        toast.warning('Free requests exhausted', {
          duration: 4000,
          description: 'Please upgrade to continue chatting.'
        })
      }
      return false // Block submission
    }

    return true // Allow submission
  } catch (error) {
    console.error('Error checking request limit:', error)
    return true // Allow on error (safety)
  }
}

function incrementRequestCount() {
  try {
    if (!userHasRequestLimits.value) {
      return // No limits for unlimited users
    }

    if (!parsedUserDetails.value) {
      return
    }

    if (!parsedUserDetails.value.requestCount) {
      parsedUserDetails.value.requestCount = { count: 0, timestamp: Date.now() }
    }

    const currentCount = parsedUserDetails.value.requestCount.count || 0

    if (currentCount < FREE_REQUEST_LIMIT) {
      parsedUserDetails.value.requestCount.count = currentCount + 1
    }
  } catch (error) {
    console.error('Failed to increment request count:', error)
  }
}

function resetRequestCount() {
  try {
    if (parsedUserDetails.value) {
      parsedUserDetails.value.requestCount = { count: 0, timestamp: Date.now() }
    }
  } catch (error) {
    console.error('Failed to reset request count:', error)
  }
}

let userDetailsDebounceTimer: any = null
let chatsDebounceTimer: any = null
let previousUserDetails: any = null

let userDetailsSyncTimeout: any = null
watch(() => parsedUserDetails.value, (newUserDetails) => {
  if (!isAuthenticated.value || !newUserDetails?.syncEnabled) return

  const oldUserDetails = previousUserDetails

  if (!oldUserDetails) {
    previousUserDetails = JSON.parse(JSON.stringify(newUserDetails))
    return
  }

  // Clear any pending sync immediately
  if (userDetailsSyncTimeout) {
    clearTimeout(userDetailsSyncTimeout)
  }

  const hasChanges = hasUserDetailsChangedMeaningfully(newUserDetails, oldUserDetails)

  if (!hasChanges) {
    console.log('No meaningful user details changes detected')
    previousUserDetails = JSON.parse(JSON.stringify(newUserDetails))
    return
  }

  console.log('User details changed meaningfully')

  userDetailsSyncTimeout = setTimeout(async () => {
    // Store the new state temporarily
    const tempNewState = JSON.parse(JSON.stringify(newUserDetails))

    try {
      console.log('Syncing user details changes to server...')

      // Save to localStorage immediately for good UX
      localStorage.setItem("userdetails", JSON.stringify(tempNewState))

      // Mark as having unsynced changes
      syncStatus.value.hasUnsyncedChanges = true

      // Attempt sync
      await performSmartSync()

      console.log('User details synced successfully')
      previousUserDetails = tempNewState

    } catch (error: any) {
      console.error('Sync after user details change failed:', error)

      // On sync failure, we keep the local changes
      previousUserDetails = tempNewState

      // Don't show error for lock conflicts - these are temporary
      if (!error.message?.includes('already in progress')) {
        toast.warning('Failed to sync user details', {
          duration: 4000,
          description: 'Changes saved locally. Will retry automatically.'
        })
      } else {
        console.log('Sync conflict - changes saved locally, will retry')
      }
    }
  }, 1500) 
}, { deep: true, immediate: false })

function hasUserDetailsChangedMeaningfully(newDetails: any, oldDetails: any): boolean {
  if (!oldDetails || !newDetails) return false

  // Ignore timestamp-only changes
  const keysToCheck = ['preferences', 
    'theme', 'workFunction', 'phoneNumber', 'syncEnabled', 'responseMode', 
    'requestCount'
  ]

  return keysToCheck.some(key => {
    const oldValue = oldDetails[key]
    const newValue = newDetails[key]

    // Handle undefined/null comparisons properly
    if (oldValue === undefined || oldValue === null) {
      return newValue !== undefined && newValue !== null
    }

    return JSON.stringify(newValue) !== JSON.stringify(oldValue)
  })
}

watch(() => chats.value, (newChats, oldChats) => {
  if (!isAuthenticated.value || !parsedUserDetails.value?.syncEnabled) {
    console.log('🔕 Sync disabled - skipping chat change detection');
    return;
  }

  if (!oldChats || oldChats.length === 0) {
    console.log('🆕 Initial chats load - no previous state to compare');
    return;
  }

  if (chatsDebounceTimer) {
    clearTimeout(chatsDebounceTimer);
  }

  const hasMeaningfulChanges = hasChatsChangedMeaningfully(newChats, oldChats);

  if (hasMeaningfulChanges) {
    console.log('💾 Chat changes detected - will sync');
    syncStatus.value.hasUnsyncedChanges = true;

    // Use a reasonable debounce but ensure sync happens
    chatsDebounceTimer = setTimeout(() => {
      if (syncStatus.value.hasUnsyncedChanges && !syncStatus.value.syncing) {
        console.log('🔄 Triggering sync from chat changes');
        performSmartSync().catch(error => {
          console.error('Sync from chat changes failed:', error);
        });
      }
    }, 1500); // Increased to 1.5 seconds
  }
}, { deep: true, immediate: false });

function hasChatsChangedMeaningfully(newChats: Chat[], oldChats: Chat[]): boolean {
  if (!oldChats || !Array.isArray(oldChats)) return true

  if (newChats.length !== oldChats.length) return true

  for (let i = 0; i < newChats.length; i++) {
    const newChat = newChats[i]
    const oldChat = oldChats[i]

    if (!oldChat) return true

    if (newChat.messages.length !== oldChat.messages.length) return true

    for (let j = 0; j < newChat.messages.length; j++) {
      const newMessage = newChat.messages[j]
      const oldMessage = oldChat.messages[j]

      if (!oldMessage) return true

      if (newMessage.prompt !== oldMessage.prompt ||
        newMessage.response !== oldMessage.response) {
        return true
      }
    }
  }

  return false
}

// Add this watch function to sync currentChatId changes
watch(() => currentChatId.value, (newChatId, oldChatId) => {
  if (!isAuthenticated.value || !parsedUserDetails.value?.syncEnabled) {
    return
  }

  if (newChatId && newChatId !== oldChatId) {
    console.log('🔄 Current chat ID changed, marking for sync')
    
    // Mark as having unsynced changes
    syncStatus.value.hasUnsyncedChanges = true
    
    // Debounced sync
    if (chatsDebounceTimer) {
      clearTimeout(chatsDebounceTimer)
    }
    
    chatsDebounceTimer = setTimeout(() => {
      if (syncStatus.value.hasUnsyncedChanges && !syncStatus.value.syncing) {
        console.log('🔄 Triggering sync from current chat ID change')
        performSmartSync().catch(error => {
          console.error('Sync from current chat ID change failed:', error)
        })
      }
    }, 500) // 0.5 second debounce
  }
}, { immediate: false, deep: true })

function applyTheme(theme: Theme) {
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
  
  if (theme === 'dark' || (theme === 'system' && prefersDark)) {
    isDarkMode.value = true
    document.documentElement.classList.add('dark')
  } else {
    isDarkMode.value = false
    document.documentElement.classList.remove('dark')
  }
}

function toggleTheme(newTheme?: Theme) {
  if (newTheme && ['light', 'dark', 'system'].includes(newTheme)) {
    parsedUserDetails.value.theme = newTheme
  } else {
    if (parsedUserDetails.value.theme === 'system') {
      parsedUserDetails.value.theme = 'light'
    } else if (parsedUserDetails.value?.theme === 'light') {
      parsedUserDetails.value.theme = 'dark'
    } else {
      parsedUserDetails.value.theme = 'system'
    }
  }

  // Apply the theme
  applyTheme(parsedUserDetails?.value?.theme || 'system')
}

// Update the watch for parsedUserDetails to handle theme changes:
watch(() => parsedUserDetails.value?.theme, (newTheme) => {
  if (newTheme && ['light', 'dark', 'system'].includes(newTheme)) {
    applyTheme(newTheme)
  }
}, { immediate: true })

let handleResize: (() => void) | null = null
let systemThemeListener: ((e: MediaQueryListEvent) => void) | null = null
let darkModeQuery: MediaQueryList | null = null

onMounted(async () => {
  try {
    console.log('App mounting...')
    
    // Theme setup
    const savedTheme = parsedUserDetails.value?.theme || 'system'

    // Apply theme immediately
    applyTheme(savedTheme)

    // System theme listener
    systemThemeListener = (e: MediaQueryListEvent) => {
      const currentTheme = parsedUserDetails?.value?.theme
      if (currentTheme === 'system' || !currentTheme) {
        isDarkMode.value = e.matches
        if (e.matches) {
          document.documentElement.classList.add('dark')
        } else {
          document.documentElement.classList.remove('dark')
        }
      }
    }

    darkModeQuery = window.matchMedia('(prefers-color-scheme: dark)')
    darkModeQuery.addEventListener('change', systemThemeListener)

    // Collapsed state
    try {
      const storedIsCollapsed = localStorage.getItem("isCollapsed")
      if (storedIsCollapsed !== null) {
        isCollapsed.value = storedIsCollapsed === "true"
      }
    } catch (error) {
      console.error('Error loading collapsed state:', error)
    }

    window.addEventListener('scroll', handleScroll, { passive: true })
    // Resize handler
    screenWidth.value = window.innerWidth
    handleResize = () => {
      try {
        screenWidth.value = window.innerWidth
      } catch (error) {
        console.error('Error handling resize:', error)
      }
    }
    window.addEventListener('resize', handleResize)

    // Authentication and data loading
    if (isAuthenticated.value) {
      console.log('User is authenticated, initializing...')

      try {
        const localExt = localStorage.getItem("external_reference")
        if (localExt) {
          try {
            const ext = JSON.parse(localExt)
            Promise.race([
              getTransaction(ext),
              new Promise((_, reject) => setTimeout(() => reject(new Error('Timeout')), 5000))
            ]).catch(extError => {
              console.error('Error processing external reference:', extError)
              localStorage.removeItem("external_reference")
            })
          } catch (extError) {
            console.error('Error processing external reference:', extError)
            localStorage.removeItem("external_reference")
          }
        }

        if (parsedUserDetails.value?.syncEnabled !== false) {
          await syncFromServer()
        } else {
          loadLocalData()
        }
      } catch (syncError) {
        console.error('Error during initial sync:', syncError)
        toast.warning('Failed to sync initial data', {
          duration: 3000,
          description: 'Loading local data instead'
        })
        loadLocalData()
      }
    } else {
      loadLocalData()
    }

    // Chat drafts
    if (currentChatId.value) {
      loadChatDrafts()
    }

    // Connection setup
    checkInternetConnection().then(isOnline => {
      console.log(`Initial connection check: ${isOnline ? 'Online' : 'Offline'}`)
    })

    setupConnectionListeners()

    if (parsedUserDetails.value) {
      loadRequestCount()
      previousUserDetails = JSON.parse(JSON.stringify(parsedUserDetails.value))
    }

    // Visibility change listener
    document.addEventListener('visibilitychange', () => {
      if (!document.hidden && !isUserOnline.value) {
        setTimeout(() => {
          checkInternetConnection()
        }, 1000)
      }
    })

    console.log('App mounted successfully')

  } catch (error) {
    console.error('Critical error during app mounting:', error)
    toast.error('Failed to initialize application', {
      duration: 5000,
      description: 'Some features may not work correctly'
    })
  }
})

onUnmounted(() => {
  // Clean up event listeners
  if (handleResize) {
    window.removeEventListener('resize', handleResize)
  }
  
  if (darkModeQuery && systemThemeListener) {
    darkModeQuery.removeEventListener('change', systemThemeListener)
  }
  
  window.removeEventListener('online', setupConnectionListeners)
  window.removeEventListener('offline', setupConnectionListeners)
  
  document.removeEventListener('visibilitychange', () => {})

  // Clean up intervals and timeouts
  const intervals = Object.values(globalThis).filter(
    value => value && typeof value === 'object' && 'refresh' in value
  )
  intervals.forEach(interval => clearInterval(interval as any))
  
  if (draftSaveTimeout) {
    clearTimeout(draftSaveTimeout)
  }
  
  if (userDetailsSyncTimeout) {
    clearTimeout(userDetailsSyncTimeout)
  }
  
  if (chatsDebounceTimer) {
    clearTimeout(chatsDebounceTimer)
  }
  
  if (userDetailsDebounceTimer) {
    clearTimeout(userDetailsDebounceTimer)
  }

  window.removeEventListener('scroll', handleScroll)
})

// Global state object with all functions and reactive references
const globalState = {
  // Reactive references
  isOpenTextHighlightPopover,
  FREE_REQUEST_LIMIT,
  requestCount,
  userDetailsDebounceTimer,
  chatsDebounceTimer,
  activeRequests,
  requestChatMap,
  pendingResponses,
  chatDrafts,
  pastePreviews,
  screenWidth,
  confirmDialog,
  isCollapsed,
  isSidebarHidden,
  authData,
  syncStatus,
  isAuthenticated,
  parsedUserDetails,
  currentChatId,
  requestsRemaining,
  shouldShowUpgradePrompt,
  isRequestLimitExceeded,
  chats,
  isLoading,
  expanded,
  showInput,
  isFreeUser,
  scrollableElem,
  showScrollDownButton,
  linkPreviewCache,
  currentChat,
  currentMessages,
  activeChatMenu,
  showProfileMenu,
  planStatus,
  isDarkMode,
  isUserOnline,
  connectionStatus,
  hasActiveRequestsForCurrentChat,

  // Core functions
  cancelAllRequests,
  cancelChatRequests,
  checkInternetConnection,
  showConfirmDialog,
  apiCall,
  logout,
  setShowInput,
  updateExpandedArray,
  createNewChat,
  switchToChat,
  deleteChat,
  renameChat,
  deleteMessage,
  clearAllChats,
  loadChatDrafts,
  clearCurrentDraft,
  saveChatDrafts,
  autoSaveDraft,
  resetRequestCount,
  incrementRequestCount,
  loadRequestCount,
  checkRequestLimitBeforeSubmit,

  // UI functions
  toggleTheme,
  scrollToBottom,
  handleScroll,
  hideSidebar,
  toggleSidebar,
  toggleChatMenu,
  handleClickOutside,
  autoGrow,
  performSmartSync,
  scrollToLastMessage,

  // Sync UI functions
  showSyncIndicator,
  hideSyncIndicator,
  updateSyncProgress,

  // Data persistence functions
  saveChats,
  loadLocalData,
  isLocalDataEmpty,
  toggleSync,

  // Link preview functions
  fetchLinkPreview,
  loadLinkPreviewCache,
  saveLinkPreviewCache,

  // Sync functions
  syncFromServer,
  syncToServer,
  manualSync,

  // Authentication
  handleAuth,
}

// Provide global state to child components
provide("globalState", globalState)
</script>

<template>
  <div @click="handleClickOutside">
    <Toaster position="top-right" :closeButton="true" closeButtonPosition="top-left" :theme="parsedUserDetails?parsedUserDetails.theme : 'system'" />
    <ConfirmDialog v-if="confirmDialog.visible" :confirmDialog="confirmDialog" />
    <UpdateModal/>
    <RouterView />
  </div>
</template>
