<script setup lang="ts">
import { computed, provide, ref, type ComputedRef } from 'vue';
import { toast, Toaster } from 'vue-sonner'
import 'vue-sonner/style.css'
import type { Chat, ConfirmDialogOptions, CurrentChat, LinkPreview } from './types';
import { API_BASE_URL, generateChatId, generateChatTitle, extractUrls } from './utils/globals';
import { nextTick } from 'vue';
import { detectAndProcessVideo } from './utils/videoProcessing';
import ConfirmDialog from './components/ConfirmDialog.vue';

const screenWidth = ref(screen.width) //  removed 'let' declaration
const scrollableElem = ref<HTMLElement | null>(null)
const showScrollDownButton = ref(false)
const confirmDialog = ref<ConfirmDialogOptions>({ 
  visible: false, 
  title: "",
  message: "", 
  type: undefined,
  confirmText: "",
  cancelText: "",
  onConfirm: () => {},
  onCancel: () => {}
})
const isCollapsed = ref(false) //  local state for collapse toggle
const isSidebarHidden = ref(true)
const authData = ref({
  username: '',
  email: '',
  password: ''
})
const syncStatus = ref({
  lastSync: null as Date | null,
  syncing: false,
  hasUnsyncedChanges: false
})
// Current chat computed property
const currentChat: ComputedRef<CurrentChat | undefined> = computed(() => {
  return chats.value.find(chat => chat.id === currentChatId.value)
})

// Current messages computed property
const currentMessages = computed(() => {
  return currentChat.value?.messages || []
})

// Link preview cache - now with persistence
const linkPreviewCache = ref<Map<string, LinkPreview>>(new Map())

// Load cached link previews from localStorage
function loadLinkPreviewCache() {
  try {
    const cached = localStorage.getItem('linkPreviews')
    if (cached) {
      const parsedCache = JSON.parse(cached)
      linkPreviewCache.value = new Map(Object.entries(parsedCache))
    }
  } catch (error) {
    console.error('Failed to load link preview cache:', error)
  }
}

// Save link preview cache to localStorage
function saveLinkPreviewCache() {
  try {
    const cacheObject = Object.fromEntries(linkPreviewCache.value)
    localStorage.setItem('linkPreviews', JSON.stringify(cacheObject))
  } catch (error) {
    console.error('Failed to save link preview cache:', error)
  }
}

let userDetails: any = localStorage.getItem("userdetails")
let parsedUserDetails: any = userDetails ? JSON.parse(userDetails) : null
// Updated isAuthenticated function
function isAuthenticated(): boolean {
  return parsedUserDetails &&
    parsedUserDetails.email &&
    parsedUserDetails.username &&
    parsedUserDetails.user_id &&
    parsedUserDetails.sessionId
}

const currentChatId = ref<string>('')
const chats = ref<Chat[]>([])
const isLoading = ref(false) //  removed 'let' declaration
const expanded = ref<boolean[]>([]) //  removed 'let' declaration
const showInput = ref(false) //  removed 'let' declaration

function showConfirmDialog(options: ConfirmDialogOptions) {
  confirmDialog.value = {
    visible: true,
    title: options.title,
    message: options.message,
    type: options.type || 'info',
    confirmText: options.confirmText || 'Confirm',
    cancelText: options.cancelText || 'Cancel',
    onConfirm: () => {
      options.onConfirm()
      confirmDialog.value.visible = false // Close dialog after confirm
    },
    onCancel: () => {
      confirmDialog.value.visible = false // Close dialog on cancel
    }
  }
}

async function apiCall(endpoint: string, options: RequestInit = {}) {
  try {
    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...(parsedUserDetails?.user_id ? { 'X-User-ID': parsedUserDetails.user_id } : {}),
        ...options.headers,
      },
    })

    const data = await response.json()

    if (!data.success) {
      throw new Error(data.message || 'API request failed')
    }

    return data
  } catch (error) {
    console.error('API Error:', error)
    throw error
  }
}

// Sync data from server to local storage
async function syncFromServer(serverData?: any) {
  if (!parsedUserDetails?.user_id) return

  try {
    syncStatus.value.syncing = true

    let data = serverData
    if (!data) {
      const response = await apiCall('/sync', { method: 'GET' })
      data = response.data
    }

    // Parse and merge server data
    if (data.chats && data.chats !== '[]') {
      try {
        const serverChats = JSON.parse(data.chats)
        const localChats = chats.value

        // Merge chats (server takes precedence for newer data)
        const mergedChats = mergeChats(serverChats, localChats)
        chats.value = mergedChats
        localStorage.setItem('chats', JSON.stringify(mergedChats))
      } catch (error) {
        console.error('Error parsing server chats:', error)
      }
    }

    if (data.link_previews && data.link_previews !== '{}') {
      try {
        const serverPreviews = JSON.parse(data.link_previews)
        // Merge with local previews
        const localPreviews = Object.fromEntries(linkPreviewCache.value)
        const mergedPreviews = { ...localPreviews, ...serverPreviews }

        linkPreviewCache.value = new Map(Object.entries(mergedPreviews))
        localStorage.setItem('linkPreviews', JSON.stringify(mergedPreviews))
      } catch (error) {
        console.error('Error parsing server link previews:', error)
      }
    }

    if (data.current_chat_id) {
      currentChatId.value = data.current_chat_id
      localStorage.setItem('currentChatId', data.current_chat_id)
    }

    syncStatus.value.lastSync = new Date()
    syncStatus.value.hasUnsyncedChanges = false
    updateExpandedArray()

  } catch (error: any) {
    console.error('Sync from server failed:', error)
    toast.warning('Failed to sync data from server', {
      duration: 3000,
      description: 'Using local data instead'
    })
  } finally {
    syncStatus.value.syncing = false
  }
}

// Sync local data to server
async function syncToServer() {
  if (!parsedUserDetails?.user_id) return

  try {
    syncStatus.value.syncing = true

    const syncData = {
      chats: JSON.stringify(chats.value),
      link_previews: JSON.stringify(Object.fromEntries(linkPreviewCache.value)),
      current_chat_id: currentChatId.value
    }

    await apiCall('/sync', {
      method: 'POST',
      body: JSON.stringify(syncData)
    })

    syncStatus.value.lastSync = new Date()
    syncStatus.value.hasUnsyncedChanges = false

  } catch (error: any) {
    console.error('Sync to server failed:', error)
    syncStatus.value.hasUnsyncedChanges = true
    toast.error('Failed to sync data to server', {
      duration: 3000,
      description: 'Your data is saved locally'
    })
  } finally {
    syncStatus.value.syncing = false
  }
}

// Merge chats function (server data takes precedence for conflicts)
function mergeChats(serverChats: Chat[], localChats: Chat[]): Chat[] {
  const merged = new Map<string, Chat>()

  // Add local chats first
  localChats.forEach(chat => {
    merged.set(chat.id, chat)
  })

  // Add server chats (will overwrite local if same ID and server is newer)
  serverChats.forEach(serverChat => {
    const localChat = merged.get(serverChat.id)
    if (!localChat || new Date(serverChat.updatedAt) > new Date(localChat.updatedAt)) {
      merged.set(serverChat.id, serverChat)
    }
  })

  // Sort by updatedAt (most recent first)
  return Array.from(merged.values()).sort((a, b) =>
    new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()
  )
}

// Updated saveChats function with server sync
function saveChats() {
  try {
    localStorage.setItem('chats', JSON.stringify(chats.value))
    localStorage.setItem('currentChatId', currentChatId.value)

    // Mark as having unsynced changes
    syncStatus.value.hasUnsyncedChanges = true

    // Auto-sync after a delay
    setTimeout(() => {
      if (syncStatus.value.hasUnsyncedChanges && !syncStatus.value.syncing) {
        syncToServer()
      }
    }, 2000)

  } catch (error) {
    console.error('Failed to save chats:', error)
  }
}

// Updated logout function
function logout() {
  showConfirmDialog({
    visible: true,
    title: 'Logout Confirmation',
    message: 'Are you sure you want to logout? Your data will be synced before logging out.',
    type: 'warning',
    confirmText: 'Logout',
    onConfirm: async () => {
      try {
        // Sync to server before logout if there are unsynced changes
        if (syncStatus.value.hasUnsyncedChanges) {
          toast.info('Syncing your data...', {
            duration: 2000,
            description: 'Please wait'
          })
          await syncToServer()
        }

        // Clear local storage
        localStorage.removeItem('userdetails')
        localStorage.removeItem('isCollapsed')
        localStorage.removeItem('currentChatId')
        localStorage.removeItem('chats')
        localStorage.removeItem('linkPreviews')

        // Reset state
        parsedUserDetails = null
        chats.value = []
        currentChatId.value = ''
        expanded.value = []
        showInput.value = false
        isCollapsed.value = false
        syncStatus.value = {
          lastSync: null,
          syncing: false,
          hasUnsyncedChanges: false
        }

        confirmDialog.value.visible = false

        toast.success('Logged out successfully', {
          duration: 3000,
          description: 'Your data has been synced'
        })

      } catch (err) {
        console.error('Error during logout:', err)
        toast.error('Error during logout', {
          duration: 4000,
          description: 'Data may not have been fully synced'
        })
      }
    },
    onCancel: () => {
      confirmDialog.value.visible = false // Close dialog on cancel
    }
  })
}

function setShowInput() {
  if (currentMessages.value.length !== 0) {
    return
  }
  if (!isAuthenticated()) {
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

// Update expanded array to match current messages
function updateExpandedArray() {
  expanded.value = currentMessages.value.map(() => false)
}

// Create new chat
function createNewChat(firstMessage?: string): string {
  const newChatId = generateChatId()
  const now = new Date().toISOString()

  const newChat: Chat = {
    id: newChatId,
    title: firstMessage ? generateChatTitle(firstMessage) : 'New Chat',
    messages: [],
    createdAt: now,
    updatedAt: now
  }

  // Add to beginning of chats array (most recent first)
  chats.value.unshift(newChat)
  currentChatId.value = newChatId
  updateExpandedArray()
  saveChats()

  return newChatId
}

function scrollToBottom() {
  if (!scrollableElem.value) return;

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
}

function handleScroll() {
  const elem = scrollableElem.value;
  if (!elem) return;

  // More lenient threshold - consider "at bottom" if within 50px
  const threshold = 50;
  const isAtBottom = elem.scrollTop + elem.clientHeight >= elem.scrollHeight - threshold;

  // Only show button when user has scrolled up significantly AND there's content to scroll to
  const hasScrollableContent = elem.scrollHeight > elem.clientHeight + threshold;
  showScrollDownButton.value = !isAtBottom && hasScrollableContent;
}

function hideSidebar() {
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
}

// Fixed switchToChat function
function switchToChat(chatId: string) {
  if (chats.value.find(chat => chat.id === chatId)) {
    currentChatId.value = chatId
    updateExpandedArray()
    localStorage.setItem('currentChatId', currentChatId.value)

    // Scroll to bottom after chat switch with proper timing
    nextTick(() => {
      setTimeout(() => {
        scrollToBottom()
      }, 100)
    })
  }
}

// Delete specific chat
function deleteChat(chatId: string) {
  if (isLoading.value) return

  const chatIndex = chats.value.findIndex(chat => chat.id === chatId)
  if (chatIndex === -1) return

  const chatTitle = chats.value[chatIndex].title
  const messageCount = chats.value[chatIndex].messages.length

  showConfirmDialog({
    visible: true,
    title: 'Delete Chat',
    message: `Are you sure you want to delete "${chatTitle}"?\n\nThis will permanently remove ${messageCount} message(s). This action cannot be undone.`,
    type: 'danger',
    confirmText: 'Delete',
    onConfirm: () => {
      chats.value.splice(chatIndex, 1)
      const chatToDelete = chats.value[chatIndex]

      // If we deleted the current chat, switch to another one
      if (currentChatId.value === chatId) {
        if (chats.value.length > 0) {
          currentChatId.value = chats.value[0].id
        } else {
          currentChatId.value = ''
        }
        updateExpandedArray()
      }

      // Remove link previews from cache
      if (chatToDelete.messages.length !== 0) {
        chatToDelete.messages.forEach(message => {
          const responseUrls = extractUrls(message.response || '')
          const promptUrls = extractUrls(message.prompt || '')
          const urls = [...new Set([...responseUrls, ...promptUrls])]
          if (urls.length > 0) {
            urls.forEach(url => {
              linkPreviewCache.value.delete(url)
            })
            saveLinkPreviewCache()
          }
        })
      }

      confirmDialog.value.visible = false
      toast.success('Chat deleted', {
        duration: 3000,
        description: 'Chat has been removed successfully.'
      })
      saveChats()
    },
    onCancel: () => {
      confirmDialog.value.visible = false // Close dialog on cancel
    }
  })
}

// Enhanced rename chat with success notification
function renameChat(chatId: string, newTitle: string) {
  const chat = chats.value.find(c => c.id === chatId)
  if (chat && newTitle.trim()) {
    chat.title = newTitle.trim()
    chat.updatedAt = new Date().toISOString()
    saveChats()
  }
}

// Enhanced delete specific message with custom dialog
function deleteMessage(messageIndex: number) {
  if (isLoading.value || !currentChat.value) return

  const message = currentChat.value.messages[messageIndex]
  const messageContent = message?.prompt || message?.response || 'this message'
  const preview = messageContent.slice(0, 50) + (messageContent.length > 50 ? '...' : '')

  showConfirmDialog({
    visible: true,
    title: 'Delete Message',
    message: `Are you sure you want to delete this message?\n\n"${preview}"\n\nThis action cannot be undone.`,
    type: 'danger',
    confirmText: 'Delete',
    onConfirm: () => {
      // Check if message exists before deletion
      if (!currentChat.value || messageIndex >= currentChat.value.messages.length) return

      // Get URLs before deleting the message
      const messageToDelete = currentChat.value.messages[messageIndex]
      const responseUrls = extractUrls(messageToDelete.response || '')
      const promptUrls = extractUrls(messageToDelete.prompt || '')
      const urls = [...new Set([...responseUrls, ...promptUrls])]

      currentChat.value.messages.splice(messageIndex, 1)
      expanded.value.splice(messageIndex, 1)

      // Update chat's updatedAt timestamp
      currentChat.value.updatedAt = new Date().toISOString()

      // Update title if we deleted the first message
      if (messageIndex === 0 && currentChat.value.messages.length > 0) {
        const firstMessage = currentChat.value.messages[0].prompt || currentChat.value.messages[0].response
        currentChat.value.title = generateChatTitle(firstMessage)
      } else if (currentChat.value.messages.length === 0) {
        currentChat.value.title = 'New Chat'
      }

      // Remove link previews from cache
      if (urls.length > 0) {
        urls.forEach(url => {
          linkPreviewCache.value.delete(url)
        })
        saveLinkPreviewCache()
      }

      confirmDialog.value.visible = false
      toast.success('Message deleted', {
        duration: 3000,
        description: 'Message has been removed successfully.'
      })
      saveChats()
    },
    onCancel: () => {
      confirmDialog.value.visible = false // Close dialog on cancel
    }
  })
}

// Enhanced clear all chats with custom dialog
function clearAllChats() {
  if (isLoading.value) return

  const totalChats = chats.value.length
  const totalMessages = chats.value.reduce((sum, chat) => sum + chat.messages.length, 0)

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
      chats.value = []
      currentChatId.value = ''
      expanded.value = []
      saveChats()

      toast.error(`${totalChats} chats with ${totalMessages} messages deleted`, {
        duration: 5000,
        description: ''
      })
      confirmDialog.value.visible = false
    },
    onCancel: () => {
      confirmDialog.value.visible = false // Close dialog on cancel
    }
  })
}

// Fetch link preview data with persistence
async function fetchLinkPreview(url: string): Promise<LinkPreview> {
  if (linkPreviewCache.value.has(url)) {
    return linkPreviewCache.value.get(url)!
  }

  const preview: LinkPreview = { url, loading: true }
  linkPreviewCache.value.set(url, preview)

  try {
    const lang = "en"
    const proxyUrl = `https://spindle.villebiz.com/scrape?url=${encodeURIComponent(url)}&lang=${lang}`

    const response = await fetch(proxyUrl)
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`)
    }

    const results = await response.json()
    const domain = new URL(url).hostname
    console.log("Scraped data:", results)

    // Enhanced video detection and processing
    const videoInfo = await detectAndProcessVideo(url, results)

    const updatedPreview: LinkPreview = {
      url,
      title: results.title?.slice(0, 100) || domain,
      description: results.description?.slice(0, 200) || "",
      images: results.images || [],
      previewImage: videoInfo.thumbnail || results.preview_image || results.images?.[0] || "",
      domain,
      favicon: results.favicon || `https://www.google.com/s2/favicons?domain=${domain}`,
      links: results.links || [],
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
  }

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

function toggleSidebar() {
  isCollapsed.value = !isCollapsed.value
  localStorage.setItem("isCollapsed", String(isCollapsed.value))
}

// Manual sync function
async function manualSync() {
  if (!parsedUserDetails?.user_id) {
    toast.warning('Please log in to sync data', {
      duration: 3000,
      description: ''
    })
    return
  }

  try {
    // First sync to server, then from server to get latest
    await syncToServer()
    await syncFromServer()

    toast.success('Data synced successfully', {
      duration: 3000,
      description: 'Your data is up to date across all devices'
    })
  } catch (error) {
    toast.error('Sync failed', {
      duration: 4000,
      description: 'Please check your internet connection'
    })
  }
}

const globalState ={
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
  logout,
  isLoading,
  expanded,
  showInput,
  showConfirmDialog,
  hideSidebar,
  setShowInput,
  clearAllChats,
  switchToChat,
  createNewChat,
  deleteChat,
  renameChat,
  deleteMessage,
  scrollableElem,
  showScrollDownButton,
  handleScroll,
  scrollToBottom,
  saveChats,
  linkPreviewCache,
  fetchLinkPreview,
  loadLinkPreviewCache,
  saveLinkPreviewCache,
  syncFromServer,
  syncToServer,
  currentChat,
  currentMessages,
  updateExpandedArray,
  apiCall,
  toggleSidebar,
  manualSync
}
provide("globalState", globalState)
</script>

<template>
  <Toaster position="top-right" :closeButton="true" closeButtonPosition="top-left"/>
  <ConfirmDialog v-if="confirmDialog.visible" :confirmDialog="confirmDialog" />
  <RouterView/>
</template>