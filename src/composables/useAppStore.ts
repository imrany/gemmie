// store.ts - Global state management
import { ref, computed, nextTick } from 'vue'
import { toast } from 'vue-sonner'
import { marked } from 'marked'
import hljs from 'highlight.js'
import type { Chat, ConfirmDialogOptions, CurrentChat, LinkPreview, Res } from '@/types'
import { 
  destroyVideoLazyLoading, 
  detectAndProcessVideo, 
  initializeVideoLazyLoading, 
  observeNewVideoContainers, 
  pauseVideo, 
  playEmbeddedVideo, 
  playSocialVideo, 
  resumeVideo, 
  showVideoControls, 
  stopDirectVideo, 
  stopVideo, 
  toggleDirectVideo, 
  updateVideoControls 
} from '@/utils/videoProcessing'
import { API_BASE_URL, WRAPPER_URL } from '@/utils/globals'

// ============ GLOBAL STATE ============
export const globalState = {
  // Confirmation dialog state
  confirmDialog: ref<ConfirmDialogOptions>({
    visible: false,
    title: '',
    message: '',
    type: 'info' as 'danger' | 'warning' | 'info',
    confirmText: 'Confirm',
    cancelText: 'Cancel',
    onConfirm: () => { }
  }),

  // Auth state
  authStep: ref(1),
  authData: ref({
    username: '',
    email: '',
    password: ''
  }),

  // Sync state
  syncStatus: ref({
    lastSync: null as Date | null,
    syncing: false,
    hasUnsyncedChanges: false
  }),

  // UI state
  showInput: ref(false),
  scrollableElem: ref<HTMLElement | null>(null),
  showCreateSession: ref(false),
  copiedIndex: ref<number | null>(null),
  screenWidth: ref(typeof window !== 'undefined' ? window.screen.width : 1024),
  isCollapsed: ref(false),
  isSidebarHidden: ref(true),
  showScrollDownButton: ref(false),

  // User state
  userDetails: typeof window !== 'undefined' ? localStorage.getItem("userdetails") : null,
  parsedUserDetails: null as any,

  // Chat state
  currentChatId: ref<string>(''),
  chats: ref<Chat[]>([]),
  isLoading: ref(false),
  expanded: ref<boolean[]>([]),

  // Link preview cache
  linkPreviewCache: ref<Map<string, LinkPreview>>(new Map()),

  // Computed properties
  currentChat: computed(() => {
    return globalState.chats.value.find(chat => chat.id === globalState.currentChatId.value)
  }),

  currentMessages: computed(() => {
    return globalState.currentChat.value?.messages || []
  })
}

// Initialize parsed user details
if (globalState.userDetails) {
  try {
    globalState.parsedUserDetails = JSON.parse(globalState.userDetails)
  } catch (error) {
    console.error('Error parsing user details:', error)
    globalState.parsedUserDetails = null
  }
}

// ============ GLOBAL FUNCTIONS ============
export const globalFunctions = {
  // ---------- Helper Functions ----------
  showConfirmDialog(options: ConfirmDialogOptions) {
    globalState.confirmDialog.value = {
      visible: true,
      title: options.title,
      message: options.message,
      type: options.type || 'info',
      confirmText: options.confirmText || 'Confirm',
      cancelText: options.cancelText || 'Cancel',
      onConfirm: options.onConfirm
    }
  },

  generateChatId(): string {
    return 'chat_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9)
  },

  generateChatTitle(firstMessage: string): string {
    const title = firstMessage.slice(0, 50).trim()
    return title.length < firstMessage.length ? title + '...' : title
  },

  extractUrls(text: string): string[] {
    const urlRegex = /https?:\/\/[^\s<>"{}|\\^`[\]]+/gi
    return text.match(urlRegex) || []
  },

  isAuthenticated(): boolean {
    return globalState.parsedUserDetails &&
      globalState.parsedUserDetails.email &&
      globalState.parsedUserDetails.username &&
      globalState.parsedUserDetails.user_id &&
      globalState.parsedUserDetails.sessionId
  },

  // ---------- Link Preview Functions ----------
  loadLinkPreviewCache() {
    try {
      const cached = localStorage.getItem('linkPreviews')
      if (cached) {
        const parsedCache = JSON.parse(cached)
        globalState.linkPreviewCache.value = new Map(Object.entries(parsedCache))
      }
    } catch (error) {
      console.error('Failed to load link preview cache:', error)
    }
  },

  saveLinkPreviewCache() {
    try {
      const cacheObject = Object.fromEntries(globalState.linkPreviewCache.value)
      localStorage.setItem('linkPreviews', JSON.stringify(cacheObject))
    } catch (error) {
      console.error('Failed to save link preview cache:', error)
    }
  },

  async fetchLinkPreview(url: string): Promise<LinkPreview> {
    if (globalState.linkPreviewCache.value.has(url)) {
      return globalState.linkPreviewCache.value.get(url)!
    }

    const preview: LinkPreview = { url, loading: true }
    globalState.linkPreviewCache.value.set(url, preview)

    try {
      const lang = "en"
      const proxyUrl = `https://spindle.villebiz.com/scrape?url=${encodeURIComponent(url)}&lang=${lang}`

      const response = await fetch(proxyUrl)
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`)
      }

      const results = await response.json()
      const domain = new URL(url).hostname

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

      globalState.linkPreviewCache.value.set(url, updatedPreview)
      globalFunctions.saveLinkPreviewCache()
      return updatedPreview
    } catch (error) {
      console.error("Failed to fetch link preview:", error)
    }

    const fallbackPreview: LinkPreview = {
      url,
      title: new URL(url).hostname,
      domain: new URL(url).hostname,
      loading: false,
      error: true
    }

    globalState.linkPreviewCache.value.set(url, fallbackPreview)
    globalFunctions.saveLinkPreviewCache()
    return fallbackPreview
  },

  // ---------- Chat Management Functions ----------
  loadChats() {
    try {
      const stored = localStorage.getItem('chats')
      if (stored) {
        const parsedChats = JSON.parse(stored)
        if (Array.isArray(parsedChats)) {
          globalState.chats.value = parsedChats
          if (globalState.chats.value.length > 0 && !globalState.currentChatId.value) {
            globalState.currentChatId.value = globalState.chats.value[0].id
          }
        }
      }
      globalFunctions.updateExpandedArray()
    } catch (error) {
      console.error('Failed to load chats:', error)
      globalState.chats.value = []
    }
  },

  updateExpandedArray() {
    globalState.expanded.value = globalState.currentMessages.value.map(() => false)
  },

  createNewChat(firstMessage?: string): string {
    const newChatId = globalFunctions.generateChatId()
    const now = new Date().toISOString()

    const newChat: Chat = {
      id: newChatId,
      title: firstMessage ? globalFunctions.generateChatTitle(firstMessage) : 'New Chat',
      messages: [],
      createdAt: now,
      updatedAt: now
    }

    globalState.chats.value.unshift(newChat)
    globalState.currentChatId.value = newChatId
    globalFunctions.updateExpandedArray()
    globalFunctions.saveChats()

    return newChatId
  },

  switchToChat(chatId: string) {
    if (globalState.chats.value.find(chat => chat.id === chatId)) {
      globalState.currentChatId.value = chatId
      globalFunctions.updateExpandedArray()
      localStorage.setItem('currentChatId', globalState.currentChatId.value)

      nextTick(() => {
        setTimeout(() => {
          globalFunctions.scrollToBottom()
        }, 100)
      })
    }
  },

  deleteChat(chatId: string) {
    if (globalState.isLoading.value) return

    const chatIndex = globalState.chats.value.findIndex(chat => chat.id === chatId)
    if (chatIndex === -1) return

    const chatTitle = globalState.chats.value[chatIndex].title
    const messageCount = globalState.chats.value[chatIndex].messages.length

    globalFunctions.showConfirmDialog({
      visible: true,
      title: 'Delete Chat',
      message: `Are you sure you want to delete "${chatTitle}"?\n\nThis will permanently remove ${messageCount} message(s). This action cannot be undone.`,
      type: 'danger',
      confirmText: 'Delete',
      onConfirm: () => {
        const chatToDelete = globalState.chats.value[chatIndex]
        globalState.chats.value.splice(chatIndex, 1)

        if (globalState.currentChatId.value === chatId) {
          if (globalState.chats.value.length > 0) {
            globalState.currentChatId.value = globalState.chats.value[0].id
          } else {
            globalState.currentChatId.value = ''
          }
          globalFunctions.updateExpandedArray()
        }

        // Remove link previews from cache
        if (chatToDelete.messages.length !== 0) {
          chatToDelete.messages.forEach(message => {
            const responseUrls = globalFunctions.extractUrls(message.response || '')
            const promptUrls = globalFunctions.extractUrls(message.prompt || '')
            const urls = [...new Set([...responseUrls, ...promptUrls])]
            if (urls.length > 0) {
              urls.forEach(url => {
                globalState.linkPreviewCache.value.delete(url)
              })
              globalFunctions.saveLinkPreviewCache()
            }
          })
        }

        globalState.confirmDialog.value.visible = false
        toast.success('Chat deleted', {
          duration: 3000,
          description: 'Chat has been removed successfully.'
        })
        globalFunctions.saveChats()
      }
    })
  },

  deleteMessage(messageIndex: number) {
    if (globalState.isLoading.value || !globalState.currentChat.value) return

    const message = globalState.currentChat.value.messages[messageIndex]
    const messageContent = message?.prompt || message?.response || 'this message'
    const preview = messageContent.slice(0, 50) + (messageContent.length > 50 ? '...' : '')

    globalFunctions.showConfirmDialog({
      visible: true,
      title: 'Delete Message',
      message: `Are you sure you want to delete this message?\n\n"${preview}"\n\nThis action cannot be undone.`,
      type: 'danger',
      confirmText: 'Delete',
      onConfirm: () => {
        if (!globalState.currentChat.value || messageIndex >= globalState.currentChat.value.messages.length) return

        const messageToDelete = globalState.currentChat.value.messages[messageIndex]
        const responseUrls = globalFunctions.extractUrls(messageToDelete.response || '')
        const promptUrls = globalFunctions.extractUrls(messageToDelete.prompt || '')
        const urls = [...new Set([...responseUrls, ...promptUrls])]

        globalState.currentChat.value.messages.splice(messageIndex, 1)
        globalState.expanded.value.splice(messageIndex, 1)
        globalState.currentChat.value.updatedAt = new Date().toISOString()

        if (messageIndex === 0 && globalState.currentChat.value.messages.length > 0) {
          const firstMessage = globalState.currentChat.value.messages[0].prompt || globalState.currentChat.value.messages[0].response
          globalState.currentChat.value.title = globalFunctions.generateChatTitle(firstMessage)
        } else if (globalState.currentChat.value.messages.length === 0) {
          globalState.currentChat.value.title = 'New Chat'
        }

        if (urls.length > 0) {
          urls.forEach(url => {
            globalState.linkPreviewCache.value.delete(url)
          })
          globalFunctions.saveLinkPreviewCache()
        }

        globalState.confirmDialog.value.visible = false
        toast.success('Message deleted', {
          duration: 3000,
          description: 'Message has been removed successfully.'
        })
        globalFunctions.saveChats()
      }
    })
  },

  clearAllChats() {
    if (globalState.isLoading.value) return

    const totalChats = globalState.chats.value.length
    const totalMessages = globalState.chats.value.reduce((sum, chat) => sum + chat.messages.length, 0)

    if (totalChats === 0) {
      toast.info('There are no chats to clear', {
        duration: 3000,
        description: 'Your chat list is already empty.'
      })
      return
    }

    globalFunctions.showConfirmDialog({
      visible: true,
      title: 'Clear All Chats',
      message: `⚠️ DELETE ALL CHATS?\n\nThis will permanently delete:\n• ${totalChats} chat(s)\n• ${totalMessages} total message(s)\n\nThis action cannot be undone!`,
      type: 'danger',
      confirmText: 'Delete All',
      onConfirm: () => {
        globalState.chats.value = []
        globalState.currentChatId.value = ''
        globalState.expanded.value = []
        globalFunctions.saveChats()

        toast.error(`${totalChats} chats with ${totalMessages} messages deleted`, {
          duration: 5000,
          description: ''
        })
        globalState.confirmDialog.value.visible = false
      }
    })
  },

  renameChat(chatId: string, newTitle: string) {
    const chat = globalState.chats.value.find(c => c.id === chatId)
    if (chat && newTitle.trim()) {
      chat.title = newTitle.trim()
      chat.updatedAt = new Date().toISOString()
      globalFunctions.saveChats()
    }
  },

  saveChats() {
    try {
      localStorage.setItem('chats', JSON.stringify(globalState.chats.value))
      localStorage.setItem('currentChatId', globalState.currentChatId.value)

      globalState.syncStatus.value.hasUnsyncedChanges = true

      setTimeout(() => {
        if (globalState.syncStatus.value.hasUnsyncedChanges && !globalState.syncStatus.value.syncing) {
          globalFunctions.syncToServer()
        }
      }, 2000)

    } catch (error) {
      console.error('Failed to save chats:', error)
    }
  },

  // ---------- API Functions ----------
  async apiCall(endpoint: string, options: RequestInit = {}) {
    try {
      const response = await fetch(`${API_BASE_URL}${endpoint}`, {
        ...options,
        headers: {
          'Content-Type': 'application/json',
          ...(globalState.parsedUserDetails?.user_id ? { 'X-User-ID': globalState.parsedUserDetails.user_id } : {}),
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
  },

  // ---------- Authentication Functions ----------
  validateCredentials(username: string, email: string, password: string): string | null {
    const usernameRegex = /^[a-zA-Z0-9_-]{3,12}$/
    if (!usernameRegex.test(username)) {
      return "Username must be 3–12 characters, no spaces, only letters, numbers, _ or -"
    }

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
    if (!emailRegex.test(email)) {
      return "Invalid email format"
    }

    if (password.length < 8) {
      return "Password must be at least 8 characters"
    }

    return null
  },

  async handleAuth(e: Event) {
    e.preventDefault()

    const { username, email, password } = globalState.authData.value

    const validationError = globalFunctions.validateCredentials(username, email, password)
    if (validationError) {
      toast.error(validationError, { duration: 4000 })
      return
    }

    try {
      globalState.isLoading.value = true

      let response
      try {
        response = await globalFunctions.apiCall('/login', {
          method: 'POST',
          body: JSON.stringify({ username, email, password })
        })

        toast.success('Welcome back!', {
          duration: 3000,
          description: `Logged in as ${response.data.username}`
        })
      } catch (loginError) {
        response = await globalFunctions.apiCall('/register', {
          method: 'POST',
          body: JSON.stringify({ username, email, password })
        })

        toast.success('Account created successfully!', {
          duration: 3000,
          description: `Welcome ${response.data.username}!`
        })
      }

      const userData = {
        user_id: response.data.user_id,
        username: response.data.username,
        email: response.data.email,
        created_at: response.data.created_at,
        sessionId: btoa(email + ':' + password + ':' + username)
      }

      localStorage.setItem('userdetails', JSON.stringify(userData))
      globalState.parsedUserDetails = userData

      await globalFunctions.syncFromServer(response.data)

      globalState.authStep.value = 1
      globalState.authData.value = { username: '', email: '', password: '' }

      nextTick(() => {
        const textarea = document.getElementById("prompt") as HTMLTextAreaElement
        if (textarea) textarea.focus()
      })

    } catch (error: any) {
      console.error('Authentication error:', error)
      toast.error('Authentication failed', {
        duration: 4000,
        description: error.message || 'Please check your credentials and try again.'
      })
    } finally {
      globalState.isLoading.value = false
    }
  },

  logout() {
    globalFunctions.showConfirmDialog({
      visible: true,
      title: 'Logout Confirmation',
      message: 'Are you sure you want to logout? Your data will be synced before logging out.',
      type: 'warning',
      confirmText: 'Logout',
      onConfirm: async () => {
        try {
          if (globalState.syncStatus.value.hasUnsyncedChanges) {
            toast.info('Syncing your data...', {
              duration: 2000,
              description: 'Please wait'
            })
            await globalFunctions.syncToServer()
          }

          localStorage.removeItem('userdetails')
          localStorage.removeItem('isCollapsed')
          localStorage.removeItem('currentChatId')
          localStorage.removeItem('chats')
          localStorage.removeItem('linkPreviews')

          globalState.parsedUserDetails = null
          globalState.chats.value = []
          globalState.currentChatId.value = ''
          globalState.expanded.value = []
          globalState.showInput.value = false
          globalState.isCollapsed.value = false
          globalState.syncStatus.value = {
            lastSync: null,
            syncing: false,
            hasUnsyncedChanges: false
          }

          globalState.confirmDialog.value.visible = false

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
      }
    })
  },

  // ---------- Sync Functions ----------
  async syncFromServer(serverData?: any) {
    if (!globalState.parsedUserDetails?.user_id) return

    try {
      globalState.syncStatus.value.syncing = true

      let data = serverData
      if (!data) {
        const response = await globalFunctions.apiCall('/sync', { method: 'GET' })
        data = response.data
      }

      if (data.chats && data.chats !== '[]') {
        try {
          const serverChats = JSON.parse(data.chats)
          const localChats = globalState.chats.value
          const mergedChats = globalFunctions.mergeChats(serverChats, localChats)
          globalState.chats.value = mergedChats
          localStorage.setItem('chats', JSON.stringify(mergedChats))
        } catch (error) {
          console.error('Error parsing server chats:', error)
        }
      }

      if (data.link_previews && data.link_previews !== '{}') {
        try {
          const serverPreviews = JSON.parse(data.link_previews)
          const localPreviews = Object.fromEntries(globalState.linkPreviewCache.value)
          const mergedPreviews = { ...localPreviews, ...serverPreviews }

          globalState.linkPreviewCache.value = new Map(Object.entries(mergedPreviews))
          localStorage.setItem('linkPreviews', JSON.stringify(mergedPreviews))
        } catch (error) {
          console.error('Error parsing server link previews:', error)
        }
      }

      if (data.current_chat_id) {
        globalState.currentChatId.value = data.current_chat_id
        localStorage.setItem('currentChatId', data.current_chat_id)
      }

      globalState.syncStatus.value.lastSync = new Date()
      globalState.syncStatus.value.hasUnsyncedChanges = false
      globalFunctions.updateExpandedArray()

    } catch (error: any) {
      console.error('Sync from server failed:', error)
      toast.warning('Failed to sync data from server', {
        duration: 3000,
        description: 'Using local data instead'
      })
    } finally {
      globalState.syncStatus.value.syncing = false
    }
  },

  async syncToServer() {
    if (!globalState.parsedUserDetails?.user_id) return

    try {
      globalState.syncStatus.value.syncing = true

      const syncData = {
        chats: JSON.stringify(globalState.chats.value),
        link_previews: JSON.stringify(Object.fromEntries(globalState.linkPreviewCache.value)),
        current_chat_id: globalState.currentChatId.value
      }

      await globalFunctions.apiCall('/sync', {
        method: 'POST',
        body: JSON.stringify(syncData)
      })

      globalState.syncStatus.value.lastSync = new Date()
      globalState.syncStatus.value.hasUnsyncedChanges = false

    } catch (error: any) {
      console.error('Sync to server failed:', error)
      globalState.syncStatus.value.hasUnsyncedChanges = true
      toast.error('Failed to sync data to server', {
        duration: 3000,
        description: 'Your data is saved locally'
      })
    } finally {
      globalState.syncStatus.value.syncing = false
    }
  },

  mergeChats(serverChats: Chat[], localChats: Chat[]): Chat[] {
    const merged = new Map<string, Chat>()

    localChats.forEach(chat => {
      merged.set(chat.id, chat)
    })

    serverChats.forEach(serverChat => {
      const localChat = merged.get(serverChat.id)
      if (!localChat || new Date(serverChat.updatedAt) > new Date(localChat.updatedAt)) {
        merged.set(serverChat.id, serverChat)
      }
    })

    return Array.from(merged.values()).sort((a, b) =>
      new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()
    )
  },

  async manualSync() {
    if (!globalFunctions.isAuthenticated()) {
      toast.warning('Please log in to sync data', {
        duration: 3000,
        description: ''
      })
      return
    }

    try {
      await globalFunctions.syncToServer()
      await globalFunctions.syncFromServer()

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
  },

  // ---------- UI Functions ----------
  setShowInput() {
    if (globalState.currentMessages.value.length !== 0) {
      return
    }
    if (!globalFunctions.isAuthenticated()) {
      toast.warning('Please create a session first', {
        duration: 3000,
        description: 'You need to be logged in.'
      })
      return
    }
    globalState.showInput.value = true
    nextTick(() => {
      const textarea = document.getElementById("prompt") as HTMLTextAreaElement
      if (textarea) textarea.focus()
    })
  },

  scrollToBottom() {
    if (!globalState.scrollableElem.value) return

    requestAnimationFrame(() => {
      if (globalState.scrollableElem.value) {
        globalState.scrollableElem.value.scrollTo({
          top: globalState.scrollableElem.value.scrollHeight,
          behavior: 'smooth'
        })
      }
    })

    setTimeout(() => {
      globalFunctions.handleScroll()
    }, 100)
  },

  handleScroll() {
    const elem = globalState.scrollableElem.value
    if (!elem) return

    const threshold = 50
    const isAtBottom = elem.scrollTop + elem.clientHeight >= elem.scrollHeight - threshold
    const hasScrollableContent = elem.scrollHeight > elem.clientHeight + threshold
    globalState.showScrollDownButton.value = !isAtBottom && hasScrollableContent
  },

  toggleSidebar() {
    globalState.isCollapsed.value = !globalState.isCollapsed.value
    localStorage.setItem("isCollapsed", String(globalState.isCollapsed.value))
  },

  hideSidebar() {
    const sideNav = document.getElementById("side_nav")
    if (sideNav) {
      if (sideNav.classList.contains("none")) {
        sideNav.classList.add("w-full", "bg-white", "z-30", "fixed", "top-0", "left-0", "bottom-0", "border-r-[1px]", "flex", "flex-col", "transition-all", "duration-300", "ease-in-out")
      } else {
        sideNav.classList.remove("w-full", "bg-white", "z-30", "fixed", "top-0", "left-0", "bottom-0", "border-r-[1px]", "flex", "flex-col", "transition-all", "duration-300", "ease-in-out")
      }
      sideNav.classList.toggle("none")
      globalState.isSidebarHidden.value = !globalState.isSidebarHidden.value
    }
  },

  // ---------- Auth Step Functions ----------
  nextAuthStep() {
    if (globalState.authStep.value < 3) {
      globalState.authStep.value++
    }
  },

  prevAuthStep() {
    if (globalState.authStep.value > 1) {
      globalState.authStep.value--
    }
  },

  validateCurrentStep(): boolean {
    switch (globalState.authStep.value) {
      case 1:
        return globalState.authData.value.username.trim().length > 0
      case 2:
        return globalState.authData.value.email.trim().length > 0 &&
          /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(globalState.authData.value.email)
      case 3:
        return globalState.authData.value.password.length >= 6
      default:
        return false
    }
  },

  handleStepSubmit(e: Event) {
    e.preventDefault()

    if (!globalFunctions.validateCurrentStep()) {
      return
    }

    if (globalState.authStep.value < 3) {
      globalFunctions.nextAuthStep()
    } else {
      globalFunctions.handleAuth(e)
    }
  },

  updateAuthData(data: Partial<{ username: string; email: string; password: string }>) {
    Object.assign(globalState.authData.value, data)
  },

  setShowCreateSession(value: boolean) {
    globalState.showCreateSession.value = value
  }
}

// ============ INITIALIZATION ============
export function initializeGlobalState() {
  // Setup marked configuration
  marked.use({
    renderer: {
      link({ href, title, text }) {
        return `<a 
          href="${href}" 
          target="_blank" 
          rel="noopener noreferrer" 
          class="text-blue-600 underline hover:text-blue