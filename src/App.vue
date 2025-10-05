<script setup lang="ts">
import { computed, onMounted, onUnmounted, provide, ref, type ComputedRef } from 'vue';
import { toast, Toaster } from 'vue-sonner'
import 'vue-sonner/style.css'
import type { Chat, ConfirmDialogOptions, CurrentChat, LinkPreview } from './types';
import { API_BASE_URL, generateChatId, generateChatTitle, extractUrls, validateCredentials, getTransaction } from './utils/globals';
import { nextTick } from 'vue';
import { detectAndProcessVideo } from './utils/videoProcessing';
import ConfirmDialog from './components/ConfirmDialog.vue';
import type { Theme } from 'vue-sonner/src/packages/types.js';

const screenWidth = ref(screen.width)
const isDarkMode = ref(false)
const currentTheme = ref<Theme | any>("system")
const scrollableElem = ref<HTMLElement | null>(null)
const showScrollDownButton = ref(false)
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

// Enhanced sync status with retry mechanism
const syncStatus = ref({
  lastSync: null as Date | null,
  syncing: false,
  hasUnsyncedChanges: false,
  lastError: null as string | null,
  retryCount: 0,
  maxRetries: 3
})

// Current chat computed property
const currentChat: ComputedRef<CurrentChat | undefined> = computed(() => {
  return chats.value.find(chat => chat.id === currentChatId.value)
})

// Current messages computed property
const currentMessages = computed(() => {
  return currentChat.value?.messages || []
})

// Link preview cache with better error handling
const linkPreviewCache = ref<Map<string, LinkPreview>>(new Map())

// Enhanced load cached link previews with error handling
function loadLinkPreviewCache() {
  try {
    const cached = localStorage.getItem('linkPreviews')
    if (cached) {
      const parsedCache = JSON.parse(cached)
      // Validate cache structure
      if (typeof parsedCache === 'object' && parsedCache !== null) {
        linkPreviewCache.value = new Map(Object.entries(parsedCache))
      }
    }
  } catch (error) {
    console.error('Failed to load link preview cache:', error)
    // Clear corrupted cache
    localStorage.removeItem('linkPreviews')
    linkPreviewCache.value.clear()
    toast.warning('Link preview cache was corrupted and has been reset')
  }
}

// Enhanced save link preview cache with error handling
function saveLinkPreviewCache() {
  try {
    const cacheObject = Object.fromEntries(linkPreviewCache.value)
    localStorage.setItem('linkPreviews', JSON.stringify(cacheObject))
  } catch (error) {
    console.error('Failed to save link preview cache:', error)
    // Try to free up space by clearing old previews
    if (linkPreviewCache.value.size > 100) {
      const entries = Array.from(linkPreviewCache.value.entries())
      const recent = entries.slice(-50) // Keep only 50 most recent
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
// Initialize user details reactively with validation
const parsedUserDetails = ref<any>(
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

// Reactive authentication state with validation
const isAuthenticated = computed(() => {
  const user = parsedUserDetails.value
  if (!user || typeof user !== 'object') return false

  return !!(
    user.email &&
    user.username &&
    user.user_id &&
    user.sessionId &&
    // Validate email format
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
  if (!parsedUserDetails.value || !parsedUserDetails.value.expiry_timestamp) {
    return { status: 'no-plan', timeLeft: '', expiryDate: '', isExpired: false }
  }

  const expiryMs = parsedUserDetails.value.expiry_timestamp < 1e12
    ? parsedUserDetails.value.expiry_timestamp * 1000
    : parsedUserDetails.value.expiry_timestamp

  const diff = expiryMs - now.value
  const isExpired = diff <= 0

  if (isExpired) {
    return { status: 'expired', timeLeft: 'Expired', expiryDate: '', isExpired: true }
  }

  const days = Math.floor(diff / (1000 * 60 * 60 * 24))
  const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60))
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60))
  const seconds = Math.floor((diff % (1000 * 60)) / 1000)

  let timeLeft = ''
  if (days > 0) {
    timeLeft = `${days}d ${hours}h ${minutes}m`
  } else if (hours > 0) {
    timeLeft = `${hours}h ${minutes}m ${seconds}s`
  } else {
    timeLeft = `${minutes}m ${seconds}s`
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

const isFreeUser = computed(() => {
  try {
    if (!parsedUserDetails.value) {
      return true // Not authenticated, consider as free user
    }

    // Check if user has no plan or free plan
    const hasFreePlan = !parsedUserDetails.value.plan ||
      parsedUserDetails.value.plan === 'free' ||
      parsedUserDetails.value.plan === '' ||
      planStatus.value.status === 'no-plan'

    // Check if user's plan has expired
    const planExpired = planStatus.value.isExpired

    // User is considered "free" if they have a free plan OR their paid plan has expired
    return hasFreePlan || planExpired

  } catch (error) {
    console.error('Error checking user plan:', error)
    return true // Default to free user on error
  }
})

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

// Enhanced API call with better error handling and retry logic
async function apiCall(endpoint: string, options: RequestInit = {}, retryCount = 0): Promise<any> {
  const maxRetries = 3
  const retryDelay = Math.pow(2, retryCount) * 1000 // Exponential backoff

  try {
    // Validate user authentication for protected endpoints
    if (!parsedUserDetails.value?.user_id && !endpoint.includes('/login') && !endpoint.includes('/register')) {
      throw new Error('User not authenticated')
    }

    // Create AbortController for timeout
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), 30000)

    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...(parsedUserDetails.value?.user_id ? { 'X-User-ID': parsedUserDetails.value.user_id } : {}),
        ...options.headers,
      },
      // Add timeout
      signal: controller.signal // 30 second timeout
    })

    clearTimeout(timeoutId)

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`)
    }

    const data = await response.json()

    if (!data.success) {
      throw new Error(data.message || 'API request failed')
    }

    // Reset retry count on success
    syncStatus.value.retryCount = 0
    syncStatus.value.lastError = null

    return data
  } catch (error: any) {
    console.error(`API Error on ${endpoint}:`, error)

    // Handle abort/timeout errors
    if (error.name === 'AbortError') {
      throw new Error('Request timeout - please try again')
    }

    // Handle network errors with retry logic
    if ((error.name === 'NetworkError' || error.name === 'TypeError' || error.name === 'TimeoutError') &&
      retryCount < maxRetries) {
      console.log(`Retrying ${endpoint} in ${retryDelay}ms (attempt ${retryCount + 1}/${maxRetries})`)

      await new Promise(resolve => setTimeout(resolve, retryDelay))
      return apiCall(endpoint, options, retryCount + 1)
    }

    // Update sync status on persistent errors
    if (endpoint.includes('/sync')) {
      syncStatus.value.lastError = error.message
      syncStatus.value.retryCount = retryCount
    }

    throw error
  }
}

// Enhanced merge chats function with better validation
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

    // Validate and add local chats first
    localChats.forEach(chat => {
      if (isValidChat(chat)) {
        merged.set(chat.id, chat)
      }
    })

    // Validate and add server chats (will overwrite local if same ID and server is newer)
    serverChats.forEach(serverChat => {
      if (!isValidChat(serverChat)) return

      const localChat = merged.get(serverChat.id)
      if (!localChat || isNewerChat(serverChat, localChat)) {
        merged.set(serverChat.id, serverChat)
      }
    })

    // Sort by updatedAt (most recent first) with error handling
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

// Enhanced logout function with better error handling
async function logout() {
  showConfirmDialog({
    visible: true,
    title: 'Logout Confirmation',
    message: 'Are you sure you want to logout? Your data will be synced before logging out.',
    type: 'warning',
    confirmText: 'Logout',
    onConfirm: async () => {
      try {
        const syncEnabled = parsedUserDetails.value?.sync_enabled

        // Attempt to sync if enabled and has changes
        if (syncStatus.value.hasUnsyncedChanges && syncEnabled && !syncStatus.value.syncing) {
          toast.info('Syncing your data...', { duration: 2000 })
          try {
            await syncToServer()
            toast.success('Data synced successfully')
          } catch (syncError) {
            console.error('Sync failed during logout:', syncError)
            toast.warning('Failed to sync data, but logout will continue')
          }
        }

        // Clear application state
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
          maxRetries: 3
        }

        // Clear storage based on sync preference
        if (syncEnabled) {
          const keysToRemove = ['chats', 'currentChatId', 'linkPreviews']
          keysToRemove.forEach(key => {
            try {
              localStorage.removeItem(key)
            } catch (error) {
              console.error(`Failed to remove ${key} from localStorage:`, error)
            }
          })
          linkPreviewCache.value.clear()
        }

        // Always remove user details
        localStorage.removeItem('userdetails')
        parsedUserDetails.value = null

        toast.success('Logged out successfully', {
          duration: 3000,
          description: syncEnabled ? 'Your data has been synced' : 'Your data was stored locally'
        })
      } catch (error) {
        console.error('Error during logout:', error)
        toast.error('Error during logout', {
          duration: 4000,
          description: 'Some cleanup operations may not have completed'
        })
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

// Enhanced update expanded array with validation
function updateExpandedArray() {
  try {
    const messagesLength = currentMessages.value?.length || 0
    expanded.value = new Array(messagesLength).fill(false)
  } catch (error) {
    console.error('Error updating expanded array:', error)
    expanded.value = []
  }
}

// Enhanced create new chat with better error handling
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

    // Validate chat data
    if (!newChat.id || !newChat.title) {
      throw new Error('Invalid chat data generated')
    }

    // Add to beginning of chats array (most recent first)
    chats.value.unshift(newChat)
    currentChatId.value = newChatId
    updateExpandedArray()
    saveChats()

    return newChatId
  } catch (error) {
    console.error('Error creating new chat:', error)
    toast.error('Failed to create new chat')
    return ''
  }
}

function scrollToBottom() {
  if (!scrollableElem.value) return;

  try {
    // Use requestAnimationFrame for smooth scrolling
    requestAnimationFrame(() => {
      if (scrollableElem.value) {
        scrollableElem.value.scrollTo({
          top: scrollableElem.value.scrollHeight,
          behavior: 'smooth'
        });
      }
    });

    // Update button visibility after scrolling
    setTimeout(() => {
      handleScroll();
    }, 100);
  } catch (error) {
    console.error('Error scrolling to bottom:', error)
  }
}

function handleScroll() {
  try {
    const elem = scrollableElem.value;
    if (!elem) return;

    // More lenient threshold - consider "at bottom" if within 50px
    const threshold = 50;
    const isAtBottom = elem.scrollTop + elem.clientHeight >= elem.scrollHeight - threshold;

    // Only show button when user has scrolled up significantly AND there's content to scroll to
    const hasScrollableContent = elem.scrollHeight > elem.clientHeight + threshold;
    showScrollDownButton.value = !isAtBottom && hasScrollableContent;
  } catch (error) {
    console.error('Error handling scroll:', error)
  }
}

function hideSidebar() {
  try {
    const sideNav = document.getElementById("side_nav")
    if (sideNav) {
      if (sideNav.classList.contains("none")) {
        sideNav.classList.add("w-full", "bg-white", "z-30", "fixed", "top-0", "left-0", "bottom-0", "border-r-[1px]", "flex", "flex-col", "transition-all", "duration-300", "ease-in-out")
      } else {
        sideNav.classList.remove("w-full", "bg-white", "z-30", "fixed", "top-0", "left-0", "bottom-0", "border-r-[1px]", "flex", "flex-col", "transition-all", "duration-300", "ease-in-out")
      }
      sideNav.classList.toggle("none")
      isSidebarHidden.value = !isSidebarHidden.value
    }
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

// Enhanced switch to chat function
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

    currentChatId.value = chatId
    updateExpandedArray()

    try {
      localStorage.setItem('currentChatId', currentChatId.value)
    } catch (error) {
      console.error('Failed to save current chat ID:', error)
    }

    // Scroll to bottom after chat switch with proper timing
    nextTick(() => {
      setTimeout(() => {
        scrollToBottom()
      }, 100)
    })
  } catch (error) {
    console.error('Error switching to chat:', error)
    toast.error('Failed to switch to chat')
  }
}

// Enhanced delete chat function
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

          saveChats()
          toast.success('Chat deleted', {
            duration: 3000,
            description: 'Chat has been removed successfully.'
          })
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

// Enhanced rename chat function
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

    toast.success('Chat renamed successfully')
  } catch (error) {
    console.error('Error renaming chat:', error)
    toast.error('Failed to rename chat')
  }
}

// Enhanced delete message function
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

          // Get URLs before deleting the message for cache cleanup
          const messageToDelete = currentChat.value.messages[messageIndex]
          const responseUrls = extractUrls(messageToDelete.response || '')
          const promptUrls = extractUrls(messageToDelete.prompt || '')
          const urls = [...new Set([...responseUrls, ...promptUrls])]

          // Remove message and corresponding expanded state
          currentChat.value.messages.splice(messageIndex, 1)
          expanded.value.splice(messageIndex, 1)

          // Update chat's timestamp
          currentChat.value.updatedAt = new Date().toISOString()

          // Update title if we deleted the first message
          if (messageIndex === 0 && currentChat.value.messages.length > 0) {
            const firstMessage = currentChat.value.messages[0].prompt || currentChat.value.messages[0].response
            if (firstMessage) {
              currentChat.value.title = generateChatTitle(firstMessage)
            }
          } else if (currentChat.value.messages.length === 0) {
            currentChat.value.title = 'New Chat'
          }

          // Clean up link previews
          if (urls.length > 0) {
            urls.forEach(url => {
              linkPreviewCache.value.delete(url)
            })
            saveLinkPreviewCache()
          }

          saveChats()
          toast.success('Message deleted', {
            duration: 3000,
            description: 'Message has been removed successfully.'
          })
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

// Enhanced clear all chats function
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
          // Clear all data
          chats.value = []
          currentChatId.value = ''
          expanded.value = []
          linkPreviewCache.value.clear()

          // Clear storage
          const keysToRemove = ['chats', 'currentChatId', 'linkPreviews']
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

// Enhanced fetch link preview with better error handling
async function fetchLinkPreview(url: string): Promise<LinkPreview> {
  // Validate URL
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
    const proxyUrl = `https://spindle.villebiz.com/scrape?url=${encodeURIComponent(url)}&lang=${lang}`

    const response = await fetch(proxyUrl, {
      signal: AbortSignal.timeout(15000) // 15 second timeout for link previews
    })

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`)
    }

    const results = await response.json()
    const domain = new URL(url).hostname

    // Enhanced video detection and processing with error handling
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

    // Fallback preview if failed
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

// Close menus when clicking outside
function handleClickOutside() {
  try {
    activeChatMenu.value = null
    showProfileMenu.value = false
  } catch (error) {
    console.error('Error handling click outside:', error)
  }
}

// Enhanced saveChats function with better error handling
function saveChats() {
  try {
    // Validate chats data before saving
    if (!Array.isArray(chats.value)) {
      console.error('Chats is not an array, resetting to empty array')
      chats.value = []
    }

    // Save to localStorage with size validation
    const chatsJson = JSON.stringify(chats.value)
    const currentChatJson = JSON.stringify(currentChatId.value)

    // Check if data is too large (localStorage typically has ~5-10MB limit)
    if (chatsJson.length > 5000000) { // ~5MB
      toast.warning('Chat data is getting large', {
        duration: 5000,
        description: 'Consider clearing old chats to improve performance'
      })
    }

    localStorage.setItem('chats', chatsJson)
    localStorage.setItem('currentChatId', currentChatId.value)

    // Only sync to server if user is authenticated and sync is enabled
    const shouldSync = isAuthenticated.value && parsedUserDetails.value?.sync_enabled

    if (shouldSync) {
      syncStatus.value.hasUnsyncedChanges = true

      // Auto-sync after short delay with debouncing
      setTimeout(() => {
        if (syncStatus.value.hasUnsyncedChanges && !syncStatus.value.syncing) {
          syncToServer().catch(error => {
            console.error('Auto-sync failed:', error)
          })
        }
      }, 2000)
    }
  } catch (error) {
    console.error('Failed to save chats:', error)
    toast.error('Failed to save chat data', {
      duration: 4000,
      description: 'Your recent changes may not be saved'
    })
  }
}

// Enhanced sync from server with comprehensive error handling
let syncLock = false

async function syncFromServer(serverData?: any) {
  if (syncLock) {
    console.log('Sync already in progress, skipping')
    return
  }

  // Only sync from server if sync is enabled or if it's initial data during auth
  const shouldSync = parsedUserDetails.value?.sync_enabled !== false || serverData

  if (!parsedUserDetails.value?.user_id || !shouldSync) {
    return
  }

  syncLock = true

  try {
    syncStatus.value.syncing = true
    syncStatus.value.lastError = null

    let data = serverData
    if (!data) {
      console.log('Fetching data from server...')
      const response = await apiCall('/sync', { method: 'GET' })
      data = response.data
    }

    if (!data) {
      throw new Error('No data received from server')
    }

    // Parse and validate server chats data
    if (data.chats && data.chats !== '[]') {
      try {
        const serverChatsData = typeof data.chats === 'string' ? JSON.parse(data.chats) : data.chats

        if (Array.isArray(serverChatsData)) {
          const localChats = chats.value
          const mergedChats = mergeChats(serverChatsData, localChats)

          chats.value = mergedChats
          localStorage.setItem('chats', JSON.stringify(mergedChats))
          console.log(`Synced ${mergedChats.length} chats from server`)
        } else {
          console.warn('Server chats data is not an array')
        }
      } catch (parseError) {
        console.error('Error parsing server chats:', parseError)
        toast.warning('Failed to parse server chat data', {
          duration: 3000,
          description: 'Using local data instead'
        })
      }
    }

    // Parse and validate server link previews
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
          console.log(`Synced ${Object.keys(mergedPreviews).length} link previews from server`)
        }
      } catch (parseError) {
        console.error('Error parsing server link previews:', parseError)
        toast.warning('Failed to parse server link preview data')
      }
    }

    // Update current chat ID if provided
    if (data.current_chat_id && typeof data.current_chat_id === 'string') {
      // Validate that the chat ID exists in our chats
      const chatExists = chats.value.some(chat => chat.id === data.current_chat_id)
      if (chatExists) {
        currentChatId.value = data.current_chat_id
        localStorage.setItem('currentChatId', data.current_chat_id)
      } else {
        console.warn('Server provided current_chat_id that does not exist in chats')
      }
    }

    // Update user preferences and settings
    if (data.sync_enabled !== undefined || data.preferences || data.theme ||
      data.work_function || data.phone_number || data.plan) {
      try {
        parsedUserDetails.value = {
          ...parsedUserDetails.value,
          preferences: data.preferences || parsedUserDetails.value?.preferences,
          theme: data.theme || parsedUserDetails.value?.theme,
          workFunction: data.work_function || parsedUserDetails.value?.workFunction,
          phone_number: data.phone_number || parsedUserDetails.value?.phone_number,
          plan: data.plan || parsedUserDetails.value?.plan,
          plan_name: data.plan_name || parsedUserDetails.value?.plan_name,
          amount: data.amount ?? parsedUserDetails.value?.amount,
          duration: data.duration || parsedUserDetails.value?.duration,
          price: data.price || parsedUserDetails.value?.price,
          expiry_timestamp: data.expiry_timestamp || parsedUserDetails.value?.expiry_timestamp,
          expire_duration: data.expire_duration || parsedUserDetails.value?.expire_duration,
          sync_enabled: data.sync_enabled !== undefined ? data.sync_enabled : parsedUserDetails.value?.sync_enabled
        }
        localStorage.setItem("userdetails", JSON.stringify(parsedUserDetails.value))
      } catch (updateError) {
        console.error('Error updating user details:', updateError)
      }
    }

    // Update sync status
    syncStatus.value.lastSync = new Date()
    syncStatus.value.hasUnsyncedChanges = false
    syncStatus.value.retryCount = 0
    updateExpandedArray()

    console.log('Successfully synced data from server')

  } catch (error: any) {
    console.error('Sync from server failed:', error)
    syncStatus.value.lastError = error.message

    // Only show toast for non-network errors or after multiple retries
    if (!error.message.includes('NetworkError') && !error.message.includes('TypeError')) {
      toast.warning('Failed to sync data from server', {
        duration: 3000,
        description: 'Using local data instead'
      })
    }

    throw error
  } finally {
    syncStatus.value.syncing = false
    syncLock = false
  }
}

// Enhanced sync to server with comprehensive error handling and validation
async function syncToServer() {
  if (syncLock) {
    console.log('Sync already in progress, skipping')
    return
  }

  // Only sync if user has sync enabled
  if (!parsedUserDetails.value?.user_id || parsedUserDetails.value.sync_enabled === false) {
    console.log('Sync to server skipped - user not authenticated or sync disabled')
    return
  }

  syncLock = true

  try {
    syncStatus.value.syncing = true
    syncStatus.value.lastError = null

    // Validate data before syncing
    if (!Array.isArray(chats.value)) {
      throw new Error('Chats data is not valid')
    }

    if (!parsedUserDetails.value.username) {
      throw new Error('User data is incomplete')
    }

    // Prepare sync data with validation
    const syncData = {
      chats: JSON.stringify(chats.value),
      link_previews: JSON.stringify(Object.fromEntries(linkPreviewCache.value)),
      current_chat_id: currentChatId.value || '',
      username: parsedUserDetails.value.username,
      preferences: parsedUserDetails.value.preferences || '',
      work_function: parsedUserDetails.value.workFunction || '',
      theme: parsedUserDetails.value.theme || 'system',
      sync_enabled: parsedUserDetails.value.sync_enabled
    }

    // Validate sync data size (prevent sending too much data)
    const syncDataSize = JSON.stringify(syncData).length
    if (syncDataSize > 10000000) { // ~10MB limit
      throw new Error('Sync data is too large. Please clear some old chats.')
    }

    console.log(`Syncing ${chats.value.length} chats to server (${(syncDataSize / 1024).toFixed(1)}KB)`)

    const response = await apiCall('/sync', {
      method: 'POST',
      body: JSON.stringify(syncData)
    })

    syncStatus.value.lastSync = new Date()
    syncStatus.value.hasUnsyncedChanges = false
    syncStatus.value.retryCount = 0

    console.log('Successfully synced data to server')

    return response

  } catch (error: any) {
    console.error('Sync to server failed:', error)
    syncStatus.value.lastError = error.message
    syncStatus.value.hasUnsyncedChanges = true

    // Increment retry count for exponential backoff
    syncStatus.value.retryCount = Math.min(syncStatus.value.retryCount + 1, syncStatus.value.maxRetries)

    // Only show error toast for certain types of errors
    if (error.message.includes('too large')) {
      toast.error('Data too large to sync', {
        duration: 5000,
        description: error.message
      })
    } else if (!error.message.includes('NetworkError') && !error.message.includes('TypeError')) {
      toast.error('Failed to sync data to server', {
        duration: 3000,
        description: 'Your data is saved locally'
      })
    }

    throw error
  } finally {
    syncStatus.value.syncing = false
    syncLock = false
  }
}

// Enhanced unsecureApiCall with better error handling and retry logic
async function unsecureApiCall(endpoint: string, options: RequestInit = {}, retryCount = 0): Promise<any> {
  const maxRetries = 2
  const retryDelay = Math.pow(2, retryCount) * 1000 // Exponential backoff

  try {
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), 15000) // 15 second timeout

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

    // Handle different types of errors
    if (error.name === 'AbortError') {
      throw new Error('Request timeout - please try again')
    }

    // Retry for network errors
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

// Enhanced handleAuth with better error handling
async function handleAuth(data: {
  username: string
  email: string
  password: string
  agreeToTerms: boolean
}) {
  const { username, email, password, agreeToTerms } = data

  try {
    // Custom validation
    const validationError = validateCredentials(username, email, password, agreeToTerms)
    if (validationError) {
      throw new Error(validationError)
    }

    let response
    let isLogin = false

    try {
      // Try login first
      console.log('Attempting login...')
      response = await unsecureApiCall('/login', {
        method: 'POST',
        body: JSON.stringify({ username, email, password, agree_to_terms: agreeToTerms })
      })
      isLogin = true

      toast.success('Welcome back!', {
        duration: 3000,
        description: `Logged in as ${response.data.username}`
      })
    } catch (loginError: any) {
      console.log('Login failed, attempting registration...')

      // Try register if login fails
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
        // If both fail, throw the more specific error
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

    // Validate response data
    if (!response.data.user_id || !response.data.username || !response.data.email) {
      throw new Error('Incomplete user data received from server')
    }

    // Store user details locally with comprehensive data
    const userData = {
      user_id: response.data.user_id,
      username: response.data.username,
      email: response.data.email,
      created_at: response.data.created_at,
      sessionId: btoa(email + ':' + password + ':' + username),
      workFunction: response.data.work_function || "",
      preferences: response.data.preferences || "",
      theme: response.data.theme || "system",
      sync_enabled: response.data.sync_enabled !== false, // Default to true if not specified
      phone_number: response.data.phone_number || "",
      plan: response.data.plan || "free",
      plan_name: response.data.plan_name || "",
      amount: response.data.amount || 0,
      duration: response.data.duration || "",
      price: response.data.price || 0,
      expiry_timestamp: response.data.expiry_timestamp || null,
      expire_duration: response.data.expire_duration || ""
    }

    localStorage.setItem('userdetails', JSON.stringify(userData))
    parsedUserDetails.value = userData

    console.log(`Authentication successful for user: ${userData.username} (sync: ${userData.sync_enabled})`)

    // Only sync data from server if sync is enabled
    if (userData.sync_enabled) {
      try {
        await syncFromServer(response.data)
        console.log('Initial data sync completed')
      } catch (syncError) {
        console.error('Initial sync failed:', syncError)
        toast.warning('Failed to sync data from server', {
          duration: 3000,
          description: 'Using local data instead'
        })
        loadLocalData()
      }
    } else {
      // Just load local data if sync is disabled
      loadLocalData()
      console.log('Sync disabled, loaded local data only')
    }

    return response

  } catch (error: any) {
    console.error('Authentication error:', error)

    // Don't show toast here - let the calling function handle it
    throw error
  }
}

// Enhanced function to load data from localStorage with validation
function loadLocalData() {
  try {
    console.log('Loading data from localStorage...')

    // Load chats with validation
    const storedChats = localStorage.getItem('chats')
    if (storedChats) {
      try {
        const parsedChats = JSON.parse(storedChats)
        if (Array.isArray(parsedChats)) {
          // Validate each chat object
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
        localStorage.removeItem('chats') // Remove corrupted data
      }
    }

    // Load current chat ID with validation
    const storedChatId = localStorage.getItem('currentChatId')
    if (storedChatId && typeof storedChatId === 'string') {
      // Verify the chat ID exists in our chats
      const chatExists = chats.value.some(chat => chat.id === storedChatId)
      if (chatExists) {
        currentChatId.value = storedChatId
      } else {
        console.warn('Stored currentChatId does not exist in chats, clearing')
        currentChatId.value = chats.value.length > 0 ? chats.value[0].id : ''
        localStorage.setItem('currentChatId', currentChatId.value)
      }
    }

    // Load link previews
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

// Enhanced manual sync with comprehensive status updates
async function manualSync() {
  if (!parsedUserDetails.value?.user_id) {
    toast.warning('Please log in to sync data', {
      duration: 3000,
      description: 'Authentication required for syncing'
    })
    return
  }

  if (parsedUserDetails.value.sync_enabled === false) {
    toast.info('Sync is disabled', {
      duration: 3000,
      description: 'Enable auto-sync in settings to sync your data'
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
    console.log('Starting manual sync...')
    // First sync to server (upload local changes)
    if (syncStatus.value.hasUnsyncedChanges) {
      console.log('Syncing local changes to server...')
      await syncToServer()
    }

    // Then sync from server (download latest data)
    console.log('Syncing latest data from server...')
    await syncFromServer()

    toast.success('Data synced successfully', {
      duration: 3000,
      description: 'Your data is up to date across all devices'
    })
    console.log('Manual sync completed successfully')

  } catch (error: any) {
    console.error('Manual sync failed:', error)
    toast.error('Sync failed', {
      duration: 4000,
      description: 'Please check your internet connection and try again'
    })
  }
}

// Enhanced auto-sync setup with better error handling and cleanup
let autoSyncInterval: any = null
let visibilityListener: (() => void) | null = null
let beforeUnloadListener: (() => void) | null = null

function setupAutoSync() {
  try {
    console.log('Setting up auto-sync...')

    // Clear existing interval and listeners
    if (autoSyncInterval) {
      clearInterval(autoSyncInterval)
      autoSyncInterval = null
    }

    if (visibilityListener) {
      document.removeEventListener('visibilitychange', visibilityListener)
    }

    if (beforeUnloadListener) {
      window.removeEventListener('beforeunload', beforeUnloadListener)
    }

    // Auto sync every 5 minutes if authenticated, sync enabled, and has unsynced changes
    autoSyncInterval = setInterval(async () => {
      if (isAuthenticated.value &&
        parsedUserDetails.value?.sync_enabled !== false &&
        syncStatus.value.hasUnsyncedChanges &&
        !syncStatus.value.syncing) {
        try {
          console.log('Auto-sync: Syncing unsynced changes...')
          await syncToServer()
        } catch (error) {
          console.error('Auto-sync failed:', error)
          // Don't show toast for auto-sync failures to avoid spam
        }
      }
    }, 5 * 60 * 1000) // 5 minutes

    // Sync when page becomes visible (only if sync enabled)
    visibilityListener = async () => {
      if (!document.hidden &&
        isAuthenticated.value &&
        parsedUserDetails.value?.sync_enabled !== false &&
        !syncStatus.value.syncing) {
        // Small delay to ensure tab is fully active
        setTimeout(async () => {
          try {
            console.log('Tab visibility: Syncing from server...')
            await syncFromServer()
          } catch (error) {
            console.error('Visibility sync failed:', error)
          }
        }, 1000)
      }
    }
    document.addEventListener('visibilitychange', visibilityListener)

    // Sync before page unload (only if sync enabled and has changes)
    beforeUnloadListener = () => {
      if (syncStatus.value.hasUnsyncedChanges &&
        parsedUserDetails.value?.sync_enabled !== false &&
        navigator.sendBeacon) {
        try {
          const syncData = {
            chats: JSON.stringify(chats.value),
            link_previews: JSON.stringify(Object.fromEntries(linkPreviewCache.value)),
            current_chat_id: currentChatId.value,
            username: parsedUserDetails.value.username,
            preferences: parsedUserDetails.value.preferences || '',
            work_function: parsedUserDetails.value.workFunction || '',
            theme: parsedUserDetails.value.theme || 'system',
            sync_enabled: parsedUserDetails.value.sync_enabled
          }

          console.log('Page unload: Sending data via beacon...')

          const formData = new FormData()
          Object.entries(syncData).forEach(([key, value]) => {
            formData.append(key, value.toString())
          })

          navigator.sendBeacon(`${API_BASE_URL}/sync`, formData)
        } catch (error) {
          console.error('Failed to send beacon:', error)
        }
      }
    }
    window.addEventListener('beforeunload', beforeUnloadListener)

    console.log('Auto-sync setup completed')

  } catch (error) {
    console.error('Error setting up auto-sync:', error)
  }
}

// Cleanup function for auto-sync
function cleanupAutoSync() {
  try {
    if (autoSyncInterval) {
      clearInterval(autoSyncInterval)
      autoSyncInterval = null
    }

    if (visibilityListener) {
      document.removeEventListener('visibilitychange', visibilityListener)
      visibilityListener = null
    }

    if (beforeUnloadListener) {
      window.removeEventListener('beforeunload', beforeUnloadListener)
      beforeUnloadListener = null
    }

    syncLock = false
    console.log('Auto-sync cleanup completed')
  } catch (error) {
    console.error('Error cleaning up auto-sync:', error)
  }
}

function toggleTheme(newTheme?: Theme) {
  // If a specific theme is provided, use it
  if (newTheme && ['light', 'dark', 'system'].includes(newTheme)) {
    currentTheme.value = newTheme
  } else {
    // Cycle through themes: system -> light -> dark -> system
    if (currentTheme.value === 'system') {
      currentTheme.value = 'light'
    } else if (currentTheme.value === 'light') {
      currentTheme.value = 'dark'
    } else {
      currentTheme.value = 'system'
    }
  }

  // Save the theme preference
  localStorage.setItem('theme', currentTheme.value)

  // Apply the theme
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
  if (currentTheme.value === 'dark' || (currentTheme.value === 'system' && prefersDark)) {
    isDarkMode.value = true
    document.documentElement.classList.add('dark')
  } else {
    isDarkMode.value = false
    document.documentElement.classList.remove('dark')
  }

  // Show theme change notification
  const themeLabel = currentTheme.value === 'system'
    ? `System (${prefersDark ? 'Dark' : 'Light'})`
    : currentTheme.value.charAt(0).toUpperCase() + currentTheme.value.slice(1)

  toast.info(`Theme: ${themeLabel}`, { duration: 1500 })
}

// Enhanced onMounted with comprehensive error handling
onMounted(async () => {
  try {
    console.log('App mounting...')
    // Initialize theme
    const savedTheme = localStorage.getItem('theme') || 'system'
    currentTheme.value = savedTheme

    // Apply the theme
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
    if (currentTheme.value === 'dark' || (currentTheme.value === 'system' && prefersDark)) {
      isDarkMode.value = true
      document.documentElement.classList.add('dark')
    } else {
      isDarkMode.value = false
      document.documentElement.classList.remove('dark')
    }

    // Listen for system theme changes when in system mode
    const systemThemeListener: any = (e: MediaQueryListEvent) => {
      const currentTheme = localStorage.getItem('theme')
      if (currentTheme === 'system' || !currentTheme) {
        isDarkMode.value = e.matches
        if (e.matches) {
          document.documentElement.classList.add('dark')
        } else {
          document.documentElement.classList.remove('dark')
        }
      }
    }

    const darkModeQuery = window.matchMedia('(prefers-color-scheme: dark)')
    darkModeQuery.addEventListener('change', systemThemeListener)

    // Load initial state from localStorage with validation
    try {
      const storedIsCollapsed = localStorage.getItem("isCollapsed")
      if (storedIsCollapsed !== null) {
        isCollapsed.value = storedIsCollapsed === "true"
      }
    } catch (error) {
      console.error('Error loading collapsed state:', error)
    }

    // Handle authenticated user initialization
    if (isAuthenticated.value) {
      console.log('User is authenticated, initializing...')

      try {
        // Handle external reference if present
        const localExt = localStorage.getItem("external_reference")
        if (localExt) {
          try {
            const ext = JSON.parse(localExt)
            // Add timeout to prevent blocking initialization
            Promise.race([
              getTransaction(ext),
              new Promise((_, reject) => setTimeout(() => reject(new Error('Timeout')), 5000))
            ]).catch(extError => {
              console.error('Error processing external reference:', extError)
              localStorage.removeItem("external_reference")
            })
          } catch (extError) {
            console.error('Error processing external reference:', extError)
            localStorage.removeItem("external_reference") // Remove invalid data
          }
        }

        // Sync from server if sync is enabled
        if (parsedUserDetails.value?.sync_enabled !== false) {
          await syncFromServer()
        } else {
          loadLocalData()
        }

        // Setup auto-sync for authenticated users
        setupAutoSync()

      } catch (syncError) {
        console.error('Error during initial sync:', syncError)
        toast.warning('Failed to sync initial data', {
          duration: 3000,
          description: 'Loading local data instead'
        })
        loadLocalData()
      }
    } else {
      // Load local data for unauthenticated users
      loadLocalData()
    }

    // Setup responsive design handling
    try {
      screenWidth.value = window.innerWidth

      const handleResize = () => {
        try {
          screenWidth.value = window.innerWidth
        } catch (error) {
          console.error('Error handling resize:', error)
        }
      }

      window.addEventListener('resize', handleResize)

      // Cleanup on unmount
      onUnmounted(() => {
        window.removeEventListener('resize', handleResize)
        cleanupAutoSync()
      })
    } catch (error) {
      console.error('Error setting up resize listener:', error)
    }

    onUnmounted(() => {
      darkModeQuery.removeEventListener('change', systemThemeListener)
      cleanupAutoSync()
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

// Global state object with all functions and reactive references
const globalState = {
  // Reactive references
  screenWidth,
  confirmDialog,
  isCollapsed,
  isSidebarHidden,
  authData,
  syncStatus,
  isAuthenticated,
  parsedUserDetails,
  currentChatId,
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
  currentTheme,

  // Core functions
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

  // UI functions
  toggleTheme,
  scrollToBottom,
  handleScroll,
  hideSidebar,
  toggleSidebar,
  toggleChatMenu,
  handleClickOutside,

  // Data persistence functions
  saveChats,
  loadLocalData,

  // Link preview functions
  fetchLinkPreview,
  loadLinkPreviewCache,
  saveLinkPreviewCache,

  // Sync functions
  syncFromServer,
  syncToServer,
  manualSync,
  setupAutoSync,
  cleanupAutoSync,

  // Authentication
  handleAuth,

  // Legacy compatibility
  autoSyncInterval // Keep for backward compatibility
}

// Provide global state to child components
provide("globalState", globalState)
</script>

<template>
  <div @click="handleClickOutside">
    <Toaster position="top-right" :closeButton="true" closeButtonPosition="top-left" :theme="currentTheme" />
    <ConfirmDialog v-if="confirmDialog.visible" :confirmDialog="confirmDialog" />
    <RouterView />
  </div>
</template>