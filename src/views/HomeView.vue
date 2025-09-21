<script lang="ts" setup>
import type { ComputedRef } from "vue"
import { ref, onMounted, nextTick, computed, onBeforeUnmount } from "vue"
import { marked } from "marked"
import hljs from "highlight.js"
import "highlight.js/styles/night-owl.css"
import SideNav from "../components/SideNav.vue"
import TopNav from "../components/TopNav.vue"
import type { Chat, ConfirmDialogOptions, CurrentChat, LinkPreview, Res } from "@/types"
import { toast } from 'vue-sonner'
import { destroyVideoLazyLoading, detectAndProcessVideo, initializeVideoLazyLoading, observeNewVideoContainers, pauseVideo, playEmbeddedVideo, playSocialVideo, resumeVideo, showVideoControls, stopDirectVideo, stopVideo, toggleDirectVideo, updateVideoControls } from "@/utils/videoProcessing"
import { onUpdated } from "vue"
import { API_BASE_URL, WRAPPER_URL } from "@/utils/globals"
import CreateSessView from "./CreateSessView.vue"

// ---------- State ----------
// Confirmation dialog state
const confirmDialog = ref<ConfirmDialogOptions>({
  visible: false,
  title: '',
  message: '',
  type: 'info' as 'danger' | 'warning' | 'info',
  confirmText: 'Confirm',
  cancelText: 'Cancel',
  onConfirm: () => { }
})

const authStep = ref(1)
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

const showInput = ref(false) //  removed 'let' declaration
const scrollableElem = ref<HTMLElement | null>(null)
const showCreateSession = ref(false) //  removed 'let' declaration
const copiedIndex = ref<number | null>(null) //  Track copied state
const screenWidth = ref(screen.width) //  removed 'let' declaration
const isCollapsed = ref(false) //  local state for collapse toggle
const isSidebarHidden = ref(true)
const showScrollDownButton = ref(false)

let userDetails: any = localStorage.getItem("userdetails")
let parsedUserDetails: any = userDetails ? JSON.parse(userDetails) : null

// Chat management state
const currentChatId = ref<string>('')
const chats = ref<Chat[]>([])
const isLoading = ref(false) //  removed 'let' declaration
const expanded = ref<boolean[]>([]) //  removed 'let' declaration

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

// ---------- Chat Management Functions ----------

// Generate unique chat ID
function generateChatId(): string {
  return 'chat_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9)
}

// Generate chat title from first message
function generateChatTitle(firstMessage: string): string {
  const title = firstMessage.slice(0, 50).trim()
  return title.length < firstMessage.length ? title + '...' : title
}

// Load chats from localStorage
function loadChats() {
  try {
    const stored = localStorage.getItem('chats')
    if (stored) {
      const parsedChats = JSON.parse(stored)
      if (Array.isArray(parsedChats)) {
        chats.value = parsedChats
        // Set current chat to the most recent one if none is set
        if (chats.value.length > 0 && !currentChatId.value) {
          currentChatId.value = chats.value[0].id
        }
      }
    }
    updateExpandedArray()
  } catch (error) {
    console.error('Failed to load chats:', error)
    chats.value = []
  }
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
    }
  })
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

// ---------- Link Preview Functions ----------

// Extract URLs from text using regex (removed extra pipe character)
function extractUrls(text: string): string[] {
  const urlRegex = /https?:\/\/[^\s<>"{}|\\^`[\]]+/gi
  return text.match(urlRegex) || []
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

// Enhanced video preview component with stop/continue functionality
function LinkPreviewComponent({ preview }: { preview: LinkPreview }) {
  if (preview.loading) {
    return `
      <div class="link-preview loading border border-gray-200 rounded-lg p-3 my-2 bg-gray-50 max-w-full">
        <div class="flex items-center gap-2">
          <i class="pi pi-spin pi-spinner text-gray-400 flex-shrink-0"></i>
          <span class="text-sm text-gray-500 truncate">Loading preview...</span>
        </div>
      </div>
    `
  }

  if (preview.error) {
    return `
      <div class="link-preview error border border-gray-200 rounded-lg p-3 my-2 bg-gray-50 max-w-full">
        <div class="flex items-center gap-2 min-w-0">
          <i class="pi pi-external-link text-gray-400 flex-shrink-0"></i>
          <a href="${preview.url}" target="_blank" rel="noopener noreferrer" 
             class="text-blue-600 hover:text-blue-800 text-sm font-medium truncate min-w-0 flex-1">
            ${preview.domain}
          </a>
        </div>
      </div>
    `
  }

  // Generate unique ID for this video instance
  const videoId = `video-${Math.random().toString(36).substr(2, 9)}`

  // Video preview component with stop/continue controls
  const renderVideoPreview = () => {
    if (!preview.video && !preview.embedHtml) return ''

    // For embeddable videos (YouTube, Vimeo)
    if (preview.embedHtml && (preview.videoType === 'youtube' || preview.videoType === 'vimeo')) {
      return `
        <div class="aspect-video w-full bg-black relative group overflow-hidden" id="${videoId}">
          <div class="video-embed-container object-cover w-full h-full" 
               data-embed='${preview.embedHtml.replace(/'/g, '&apos;')}'
               data-video-type="${preview.videoType}"
               data-video-id="${videoId}">
            
            <!-- Initial thumbnail state -->
            <div class="video-thumbnail w-full h-full bg-gray-900 flex items-center justify-center cursor-pointer overflow-hidden"
                 onclick="playEmbeddedVideo(this, '${videoId}')">
              ${preview.videoThumbnail || preview.previewImage ? `
                <img src="${preview.videoThumbnail || preview.previewImage}" 
                     alt="${preview.title}" class="w-full h-full object-cover">
              ` : ''}
              <div class="absolute inset-0 flex items-center justify-center bg-black bg-opacity-30 group-hover:bg-opacity-20 transition-colors">
                <div class="w-12 h-12 sm:w-16 sm:h-16 bg-red-600 hover:bg-red-700 rounded-full flex items-center justify-center flex-shrink-0 transform hover:scale-110 transition-all duration-200">
                  <svg class="w-4 h-4 sm:w-6 sm:h-6 text-white ml-0.5 sm:ml-1" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M8 5v14l11-7z"/>
                  </svg>
                </div>
              </div>
              ${preview.videoDuration ? `
                <div class="absolute bottom-2 right-2 bg-black bg-opacity-80 text-white text-xs px-2 py-1 rounded max-w-[calc(100%-1rem)] truncate">
                  ${preview.videoDuration}
                </div>
              ` : ''}
            </div>
          </div>
          
          <!-- Video controls overlay (hidden initially) -->
          <div class="video-controls absolute top-2 right-2 flex gap-2 opacity-0 transition-opacity duration-200" 
               id="${videoId}-controls">
            <button onclick="pauseVideo('${videoId}')" 
                    class="pause-btn w-8 h-8 bg-black bg-opacity-70 hover:bg-opacity-90 rounded-full flex items-center justify-center text-white transition-all"
                    title="Pause">
              <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                <path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z"/>
              </svg>
            </button>
            <button onclick="resumeVideo('${videoId}')" 
                    class="play-btn w-8 h-8 bg-black bg-opacity-70 hover:bg-opacity-90 rounded-full flex items-center justify-center text-white transition-all hidden"
                    title="Resume">
              <svg class="w-4 h-4 ml-0.5" fill="currentColor" viewBox="0 0 24 24">
                <path d="M8 5v14l11-7z"/>
              </svg>
            </button>
            <button onclick="stopVideo('${videoId}')" 
                    class="stop-btn w-8 h-8 bg-black bg-opacity-70 hover:bg-opacity-90 rounded-full flex items-center justify-center text-white transition-all"
                    title="Stop">
              <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                <path d="M6 6h12v12H6z"/>
              </svg>
            </button>
          </div>
        </div>
      `
    }

    // For direct video files
    if (preview.videoType === 'direct' && preview.video) {
      return `
        <div class="aspect-video w-full bg-black overflow-hidden relative group" id="${videoId}">
          <video 
            id="${videoId}-video"
            controls 
            preload="metadata" 
            class="w-full h-full object-contain" 
            poster="${preview.previewImage || ''}"
            onplay="showVideoControls('${videoId}')"
            onpause="updateVideoControls('${videoId}', 'paused')"
            onended="updateVideoControls('${videoId}', 'ended')">
            <source src="${preview.video}" type="video/mp4">
            <source src="${preview.video}" type="video/webm">
            Your browser does not support the video tag.
          </video>
          
          <!-- Custom controls overlay for direct videos -->
          <div class="video-controls absolute top-2 right-2 flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity duration-200" 
               id="${videoId}-controls">
            <button onclick="toggleDirectVideo('${videoId}')" 
                    class="toggle-btn w-8 h-8 bg-black bg-opacity-70 hover:bg-opacity-90 rounded-full flex items-center justify-center text-white transition-all"
                    title="Play/Pause">
              <svg class="play-icon w-4 h-4 ml-0.5" fill="currentColor" viewBox="0 0 24 24">
                <path d="M8 5v14l11-7z"/>
              </svg>
              <svg class="pause-icon w-4 h-4 hidden" fill="currentColor" viewBox="0 0 24 24">
                <path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z"/>
              </svg>
            </button>
            <button onclick="stopDirectVideo('${videoId}')" 
                    class="stop-btn w-8 h-8 bg-black bg-opacity-70 hover:bg-opacity-90 rounded-full flex items-center justify-center text-white transition-all"
                    title="Stop">
              <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                <path d="M6 6h12v12H6z"/>
              </svg>
            </button>
          </div>
        </div>
      `
    }

    // For social media videos (no stop/continue - just external link)
    if ((preview.videoType === 'twitter' || preview.videoType === 'tiktok') && preview.previewImage) {
      return `
        <div class="aspect-video w-full bg-gray-100 relative group overflow-hidden cursor-pointer"
             onclick="playSocialVideo('${preview.url}', '${preview.videoType}')">
          <img src="${preview.previewImage}" alt="${preview.title}" 
               class="w-full h-full object-cover">
          <div class="absolute inset-0 flex items-center justify-center bg-black bg-opacity-30 group-hover:bg-opacity-20 transition-colors">
            <div class="w-10 h-10 sm:w-12 sm:h-12 bg-white bg-opacity-90 hover:bg-opacity-100 rounded-full flex items-center justify-center flex-shrink-0 transform hover:scale-110 transition-all duration-200">
              <svg class="w-3 h-3 sm:w-4 sm:h-4 text-gray-800 ml-0.5" fill="currentColor" viewBox="0 0 24 24">
                <path d="M8 5v14l11-7z"/>
              </svg>
            </div>
          </div>
          <div class="absolute top-2 left-2 bg-black bg-opacity-80 text-white text-xs px-2 py-1 rounded capitalize">
            ${preview.videoType}
          </div>
        </div>
      `
    }

    return ''
  }

  const hasVideo = preview.video || preview.embedHtml
  const videoPreview = renderVideoPreview()

  return `
    <div class="link-preview border border-gray-200 rounded-lg overflow-hidden my-2 bg-white hover:shadow-md transition-shadow w-fit">
      ${hasVideo ? `
        <div class="w-full md:w-[500px]">
          ${videoPreview}
          <div class="p-3 sm:p-4 min-w-0">
            <div class="flex items-start justify-between gap-2 min-w-0">
              <div class="flex-1 min-w-0">
                <h4 class="font-medium text-gray-900 text-sm sm:text-base line-clamp-2 mb-1 break-words">
                  <i class="pi pi-play-circle text-red-600 mr-1 flex-shrink-0"></i>
                  <a href="${preview.url}" target="_blank" rel="noopener noreferrer" class="hover:text-blue-600 break-words">
                    ${preview.title}
                  </a>
                </h4>
                ${preview.description ? `
                  <p class="text-gray-600 text-xs sm:text-sm line-clamp-2 sm:line-clamp-3 mb-2 break-words leading-relaxed">${preview.description}</p>
                ` : ''}
                <div class="flex items-center gap-1 text-xs sm:text-sm text-gray-500 min-w-0">
                  <i class="pi pi-video text-red-600 flex-shrink-0"></i>
                  <span class="truncate min-w-0 flex-1">${preview.domain}</span>
                  ${preview.videoDuration ? `<span class="ml-2 flex-shrink-0 hidden xs:inline">• ${preview.videoDuration}</span>` : ''}
                </div>
                ${preview.videoDuration ? `
                  <div class="text-xs text-gray-500 mt-1 xs:hidden">
                    Duration: ${preview.videoDuration}
                  </div>
                ` : ''}
              </div>
            </div>
          </div>
        </div>
      ` : `
        <!-- Regular link preview -->
        <a href="${preview.url}" class="w-full md:w-[300px]" target="_blank" rel="noopener noreferrer" class="block">
          ${preview.previewImage ? `
            <div class="aspect-video overflow-hidden bg-gray-100">
              <img src="${preview.previewImage}" alt="${preview.title}" 
                   class="w-full h-full object-cover"
                   onerror="this.parentElement.style.display='none'">
            </div>
          ` : ''}
          <div class="p-3 sm:p-4 min-w-0">
            <div class="flex items-start justify-between gap-2 min-w-0">
              <div class="flex-1 min-w-0">
                <h4 class="font-medium text-gray-900 text-sm sm:text-base line-clamp-2 mb-1 break-words">
                  <span class="break-words">${preview.title}</span>
                </h4>
                ${preview.description ? `
                  <p class="text-gray-600 text-xs sm:text-sm line-clamp-2 sm:line-clamp-3 mb-2 break-words leading-relaxed">${preview.description}</p>
                ` : ''}
                <div class="flex items-center gap-1 text-xs sm:text-sm text-gray-500 min-w-0">
                  <i class="pi pi-external-link flex-shrink-0"></i>
                  <span class="truncate min-w-0 flex-1">${preview.domain}</span>
                </div>
              </div>
            </div>
          </div>
        </a>
      `}
    </div>
  `
}

// API helper function
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

function validateCredentials(username: string, email: string, password: string): string | null {
  // Username: 3–12 chars, no spaces, only letters, numbers, underscores, hyphens
  const usernameRegex = /^[a-zA-Z0-9_-]{3,12}$/
  if (!usernameRegex.test(username)) {
    return "Username must be 3–12 characters, no spaces, only letters, numbers, _ or -"
  }

  // Email: basic check
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  if (!emailRegex.test(email)) {
    return "Invalid email format"
  }

  // Password: at least 8 chars
  if (password.length < 8) {
    return "Password must be at least 8 characters"
  }

  return null
}

async function handleAuth(e: Event) {
  e.preventDefault()

  const { username, email, password } = authData.value

  // Custom validation
  const validationError = validateCredentials(username, email, password)
  if (validationError) {
    toast.error(validationError, { duration: 4000 })
    return
  }

  try {
    isLoading.value = true

    let response
    try {
      // Try login
      response = await apiCall('/login', {
        method: 'POST',
        body: JSON.stringify({ username, email, password })
      })

      toast.success('Welcome back!', {
        duration: 3000,
        description: `Logged in as ${response.data.username}`
      })
    } catch (loginError) {
      // Try register if login fails
      response = await apiCall('/register', {
        method: 'POST',
        body: JSON.stringify({ username, email, password })
      })

      toast.success('Account created successfully!', {
        duration: 3000,
        description: `Welcome ${response.data.username}!`
      })
    }

    // Store user details locally
    const userData = {
      user_id: response.data.user_id,
      username: response.data.username,
      email: response.data.email,
      created_at: response.data.created_at,
      sessionId: btoa(email + ':' + password + ':' + username)
    }

    localStorage.setItem('userdetails', JSON.stringify(userData))
    parsedUserDetails = userData

    // Sync data from server
    await syncFromServer(response.data)

    // Reset form
    authStep.value = 1
    authData.value = { username: '', email: '', password: '' }

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
    isLoading.value = false
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

// ---------- Authentication Functions ----------
function nextAuthStep() {
  if (authStep.value < 3) {
    authStep.value++
  }
}

function prevAuthStep() {
  if (authStep.value > 1) {
    authStep.value--
  }
}

function validateCurrentStep(): boolean {
  switch (authStep.value) {
    case 1:
      return authData.value.username.trim().length > 0
    case 2:
      return authData.value.email.trim().length > 0 &&
        /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(authData.value.email)
    case 3:
      return authData.value.password.length >= 6
    default:
      return false
  }
}

function handleStepSubmit(e: Event) {
  e.preventDefault()

  if (!validateCurrentStep()) {
    return
  }

  if (authStep.value < 3) {
    nextAuthStep()
  } else {
    // Final step - create session
    handleAuth(e)
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
    }
  })
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

// Updated isAuthenticated function
function isAuthenticated(): boolean {
  return parsedUserDetails &&
    parsedUserDetails.email &&
    parsedUserDetails.username &&
    parsedUserDetails.user_id &&
    parsedUserDetails.sessionId
}


// Auto-sync on app focus (when user comes back to tab)
let autoSyncInterval: any = null

function setupAutoSync() {
  // Clear existing interval
  if (autoSyncInterval) {
    clearInterval(autoSyncInterval)
  }

  // Auto sync every 5 minutes if authenticated and has unsynced changes
  autoSyncInterval = setInterval(() => {
    if (isAuthenticated() && syncStatus.value.hasUnsyncedChanges && !syncStatus.value.syncing) {
      syncToServer()
    }
  }, 5 * 60 * 1000) // 5 minutes

  // Sync when page becomes visible
  document.addEventListener('visibilitychange', () => {
    if (!document.hidden && isAuthenticated()) {
      // Small delay to ensure tab is fully active
      setTimeout(() => {
        syncFromServer()
      }, 1000)
    }
  })

  // Sync before page unload
  window.addEventListener('beforeunload', () => {
    if (syncStatus.value.hasUnsyncedChanges) {
      // Use sendBeacon for reliable data sending on page unload
      const syncData = {
        chats: JSON.stringify(chats.value),
        link_previews: JSON.stringify(Object.fromEntries(linkPreviewCache.value)),
        current_chat_id: currentChatId.value
      }

      navigator.sendBeacon(`${API_BASE_URL}/sync`, JSON.stringify(syncData))
    }
  })
}

// ---------- Helpers ----------
// Helper function to show confirmation dialog
function showConfirmDialog(options: ConfirmDialogOptions) {
  confirmDialog.value = {
    visible: true,
    title: options.title,
    message: options.message,
    type: options.type || 'info',
    confirmText: options.confirmText || 'Confirm',
    cancelText: options.cancelText || 'Cancel',
    onConfirm: options.onConfirm
  }
}

function copyCode(text: string, button?: HTMLElement) {
  navigator.clipboard.writeText(text)
    .then(() => {
      if (button) {
        button.innerText = 'Copied!'
        setTimeout(() => (button.innerText = 'Copy code'), 2000)
      }
    })
    .catch(err => {
      console.error('Failed to copy text: ', err)
      toast.error('Failed to copy code to clipboard', {
        duration: 3000,
        description: ''
      })
    })
}

// Enhanced marked configuration with link handling
marked.use({
  renderer: {
    link({ href, title, text }) {
      return `<a 
        href="${href}" 
        target="_blank" 
        rel="noopener noreferrer" 
        class="text-blue-600 underline hover:text-blue-800 link-with-preview"
        data-url="${href}"
      >${text}</a>`
    },
    code({ text, lang }) {
      let highlighted = lang && hljs.getLanguage(lang)
        ? hljs.highlight(text, { language: lang }).value
        : hljs.highlightAuto(text).value

      return `
        <div class="code-container relative">
          <pre><code class="hljs language-${lang || 'plaintext'}">${highlighted}</code></pre>
          <button 
            class="copy-button absolute top-2 right-2 bg-gray-700 text-white px-2 py-1 rounded text-xs hover:bg-gray-600"
            data-code="${encodeURIComponent(text)}"
          >Copy code</button>
        </div>
      `
    }
  }
})

function renderMarkdown(text?: string) {
  if (!text || typeof text !== "string") return ""
  try {
    return marked.parse(text)
  } catch (err) {
    console.error("Markdown parse error:", err)
    return text
  }
}

function isPromptTooShort(prompt: string): boolean {
  return prompt.trim().split(/\s+/).length < 30
}

//  Debounced scroll handler to improve performance
let scrollTimeout: any = null
function debouncedHandleScroll() {
  if (scrollTimeout) {
    clearTimeout(scrollTimeout)
  }

  scrollTimeout = setTimeout(() => {
    handleScroll()
  }, 16) // ~60fps
}

// Detect if prompt is just URLs (1 or more) with little/no extra text
function isJustLinks(prompt: string): boolean {
  const trimmed = prompt.trim()
  const urls = extractUrls(trimmed)

  if (urls.length === 0) return false

  // Remove all URLs from prompt
  let promptWithoutUrls = trimmed
  for (const url of urls) {
    promptWithoutUrls = promptWithoutUrls.replace(url, "").trim()
  }

  // If only short filler words remain, treat as "just links"
  return promptWithoutUrls.split(/\s+/).filter(Boolean).length <= 3
}

// Modified handleSubmit function (around line 420)
async function handleSubmit(e?: any, retryPrompt?: string) {
  e?.preventDefault?.()

  let promptValue = retryPrompt || e?.target?.prompt?.value?.trim()
  let fabricatedPrompt = promptValue
  if (!promptValue || isLoading.value) return

  if (!isAuthenticated()) {
    toast.warning('Please create a session first', {
      duration: 4000,
      description: 'You need to be logged in.'
    })
    return
  }

  // handling for link-only prompts
  if (isJustLinks(promptValue)) {
    const urls = extractUrls(promptValue)

    // Create chat if needed
    if (!currentChatId.value || !currentChat.value) {
      createNewChat(promptValue)
    }

    isLoading.value = true

    if (!retryPrompt && e?.target?.prompt) {
      e.target.prompt.value = ""
      e.target.prompt.style.height = "auto"
    }

    const tempResp: Res = { prompt: promptValue, response: "..." }
    currentChat.value?.messages.push(tempResp)
    expanded.value.push(false)
    await nextTick()
    scrollToBottom()

    try {
      let combinedResponse = `I've analyzed the link${urls.length > 1 ? "s" : ""} you shared:\n\n`

      for (const url of urls) {
        try {
          const linkPreview = await fetchLinkPreview(url)

          combinedResponse += `**${linkPreview.title || 'Untitled'}**\n`
          if (linkPreview.description) {
            combinedResponse += `Description: ${linkPreview.description}\n`
          }
          combinedResponse += `Domain: ${linkPreview.domain || new URL(url).hostname}\n`
          combinedResponse += `URL: ${url}\n\n`
        } catch (err: any) {
          combinedResponse += `⚠️ Failed to analyze: ${url} (${err.message || "Unknown error"})\n\n`
        }
      }

      // Update the response in chat
      if (currentChat.value) {
        const lastMessageIndex = currentChat.value.messages.length - 1
        currentChat.value.messages[lastMessageIndex] = {
          prompt: promptValue,
          response: combinedResponse.trim(),
          status: 200
        }
        currentChat.value.updatedAt = new Date().toISOString()
      }
    } finally {
      isLoading.value = false
      saveChats()
      observeNewVideoContainers()
      await nextTick()
      scrollToBottom()
    }

    return // ✅ Exit early for link-only prompts
  }


  // Merge with only the latest message if prompt is short
  if (isPromptTooShort(promptValue) && currentMessages.value.length > 0) {
    const lastMessage = currentMessages.value[currentMessages.value.length - 1]
    fabricatedPrompt = `${lastMessage.prompt || ''} ${lastMessage.response || ''}\nUser: ${promptValue}`
  }

  // Create new chat if none exists
  if (!currentChatId.value || !currentChat.value) {
    createNewChat(promptValue)
  }

  isLoading.value = true

  if (!retryPrompt && e?.target?.prompt) {
    e.target.prompt.value = ""
    e.target.prompt.style.height = "auto"
  }

  const tempResp: Res = { prompt: promptValue, response: "..." }

  // Add message to current chat
  if (currentChat.value) {
    currentChat.value.messages.push(tempResp)
    currentChat.value.updatedAt = new Date().toISOString()

    // Update chat title if this is the first message
    if (currentChat.value.messages.length === 1) {
      currentChat.value.title = generateChatTitle(promptValue)
    }
  }

  expanded.value.push(false)

  // Process links in user prompt
  processLinksInUserPrompt(promptValue)

  // Scroll to bottom after adding user message
  await nextTick()
  scrollToBottom()

  try {
    let response = await fetch(WRAPPER_URL, {
      method: "POST",
      body: JSON.stringify(fabricatedPrompt),
      headers: { "content-type": "application/json" }
    })

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`)
    }

    let parseRes = await response.json()

    if (currentChat.value) {
      const lastMessageIndex = currentChat.value.messages.length - 1
      currentChat.value.messages[lastMessageIndex] = {
        prompt: promptValue,
        response: parseRes.error ? parseRes.error : parseRes.response,
        status: response.status
      }
      currentChat.value.updatedAt = new Date().toISOString()
    }

    // Trigger link preview generation for the new response
    await processLinksInResponse(currentMessages.value.length - 1)

  } catch (err: any) {
    toast.error(`Failed to get AI response: ${err.message}`, {
      duration: 5000,
      description: ''
    })

    if (currentChat.value) {
      const lastMessageIndex = currentChat.value.messages.length - 1
      currentChat.value.messages[lastMessageIndex] = {
        prompt: promptValue,
        response: `⚠️ Error: ${err.message || 'Failed to get response. Please try again.'}`
      }
      currentChat.value.updatedAt = new Date().toISOString()
    }
  } finally {
    isLoading.value = false
    saveChats()

    // Observe any new video containers
    observeNewVideoContainers();

    // Scroll to bottom after response is complete
    await nextTick()
    scrollToBottom()
  }
}

// Process links in a response and generate previews
async function processLinksInResponse(index: number) {
  const messages = currentMessages.value
  if (!messages[index] || !messages[index].response || messages[index].response === "...") return

  const urls = extractUrls(messages[index].response)
  if (urls.length > 0) {
    // Start loading previews
    urls.slice(0, 3).forEach(url => {
      fetchLinkPreview(url).then(() => {
        // Trigger reactivity update
        linkPreviewCache.value = new Map(linkPreviewCache.value)
      })
    })
  }
}

// Process links in user prompts
async function processLinksInUserPrompt(prompt: string) {
  const urls = extractUrls(prompt)
  if (urls.length > 0) {
    // Start loading previews for user prompt links
    urls.slice(0, 3).forEach(url => {
      fetchLinkPreview(url).then(() => {
        // Trigger reactivity update
        linkPreviewCache.value = new Map(linkPreviewCache.value)
      })
    })
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
}

// ---------- Extra actions ----------
function copyResponse(text: string, index?: number) {
  navigator.clipboard.writeText(text)
    .then(() => {
      copiedIndex.value = index ?? null

      setTimeout(() => {
        copiedIndex.value = null
      }, 2000)
    })
    .catch(err => {
      console.error('Failed to copy text: ', err)
      toast.error('Copy Failed', {
        duration: 3000,
        description: ''
      })
    })
}

function toggleSidebar() {
  isCollapsed.value = !isCollapsed.value
  localStorage.setItem("isCollapsed", String(isCollapsed.value))
}

function shareResponse(text: string, prompt?: string) {
  if (navigator.share) {
    navigator.share({
      title: prompt && prompt.length > 200 ? `${prompt.slice(0, 200)}...\n\n` : `${prompt || "Gemmie Chat"}\n\n`,
      text
    }).then(() => {
      console.log("Share successful")
    }).catch(err => {
      console.log("Share canceled", err)
      toast.warning('Share Cancelled', {
        duration: 2000,
        description: "Cannot Share at the moment, copying instead."
      })
    })
  } else {
    copyCode(text)
    toast.info('Copied Instead', {
      duration: 3000,
    })
  }
}

function refreshResponse(prompt?: string) {
  if (prompt && !isLoading.value) {
    handleSubmit(undefined, prompt)
  }
}

// Add function to manually clear link preview cache
function clearLinkPreviewCache() {
  const cacheSize = linkPreviewCache.value.size

  if (cacheSize === 0) {
    toast.info('Link preview cache is already empty', {
      duration: 3000,
      description: ''
    })
    return
  }

  showConfirmDialog({
    visible: true,
    title: 'Clear Link Preview Cache',
    message: `Clear all link preview cache? This will remove ${cacheSize} cached preview(s) and require refetching previews for existing links.`,
    type: 'warning',
    confirmText: 'Clear Cache',
    onConfirm: () => {
      try {
        localStorage.removeItem('linkPreviews')
        linkPreviewCache.value.clear()
        confirmDialog.value.visible = false
      } catch (err) {
        console.error('Failed to clear link preview cache:', err)
        toast.error('Failed to clear link preview cache.', {
          duration: 3000,
          description: ''
        })
      }
    }
  })
}

// ---------- UI Helpers ----------
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

//  Improved scroll functions
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

let resizeTimeout: any
window.onresize = () => {
  clearTimeout(resizeTimeout)
  resizeTimeout = setTimeout(() => {
    screenWidth.value = screen.width
  }, 100)
}

function onEnter(e: KeyboardEvent) {
  if (e.key !== 'Enter' || e.shiftKey || isLoading.value) {
    return
  }

  e.preventDefault()

  const textarea = e.target as HTMLTextAreaElement
  if (textarea && textarea.value.trim()) {
    const formEvent = {
      preventDefault: () => { },
      target: { prompt: textarea }
    }
    handleSubmit(formEvent)
  }
}

onMounted(() => {
  // Load collapsed state
  const saved = localStorage.getItem("isCollapsed");
  if (saved && saved !== "null") {
    try {
      isCollapsed.value = JSON.parse(saved);
    } catch (err) {
      console.error("Error parsing isCollapsed:", err);
    }
  }

  // Load chat ID
  const savedChatId = localStorage.getItem("currentChatId");
  if (savedChatId) currentChatId.value = savedChatId;

  // Load cached previews
  loadLinkPreviewCache();

  // Load chats if logged in
  if (isAuthenticated()) {
    loadChats();

    // Setup auto-sync
    setupAutoSync()

    // Initial sync from server (delayed to avoid conflicts)
    setTimeout(() => {
      syncFromServer()
    }, 1000)

    // Pre-process links in existing messages
    currentMessages.value.forEach((item, index) => {
      [item.prompt, item.response].forEach((text) => {
        if (text && text !== "...") {
          extractUrls(text)
            .slice(0, 3)
            .forEach((url) => {
              if (!linkPreviewCache.value.has(url)) {
                fetchLinkPreview(url).then(() => {
                  // trigger reactivity
                  linkPreviewCache.value = new Map(linkPreviewCache.value);
                });
              }
            });
        }
      });
    });
  }

  // Global copy button handler
  const copyListener = (e: any) => {
    if (e.target?.classList.contains("copy-button")) {
      const code = decodeURIComponent(e.target.getAttribute("data-code"));
      copyCode(code, e.target);
    }
  };
  document.addEventListener("click", copyListener);

  // Make functions globally available
  if (typeof window !== 'undefined') {
    (window as any).playEmbeddedVideo = playEmbeddedVideo;
    (window as any).pauseVideo = pauseVideo;
    (window as any).resumeVideo = resumeVideo;
    (window as any).stopVideo = stopVideo;
    (window as any).toggleDirectVideo = toggleDirectVideo;
    (window as any).stopDirectVideo = stopDirectVideo;
    (window as any).showVideoControls = showVideoControls;
    (window as any).updateVideoControls = updateVideoControls;
    (window as any).playSocialVideo = playSocialVideo;
  }

  // Initialize video lazy loading once
  initializeVideoLazyLoading();

  nextTick(() => {
    // Set up scroll listener with proper cleanup
    if (scrollableElem.value) {
      scrollableElem.value.addEventListener("scroll", debouncedHandleScroll, { passive: true });
    }

    // Auto-focus input
    if (showInput.value || currentMessages.value.length > 0) {
      const textarea = document.getElementById("prompt") as HTMLTextAreaElement;
      textarea?.focus();
    }

    // Process link previews in responses
    currentMessages.value.forEach((msg: Res, index) => {
      if (msg.response && msg.response !== "...") {
        processLinksInResponse(index);
      }
    });

    // Observe existing video containers after processing links
    observeNewVideoContainers();

    // Initial scroll to bottom with delay to ensure content is rendered
    setTimeout(() => {
      scrollToBottom();
    }, 100);
  });

  // Store cleanup function for onBeforeUnmount
  onBeforeUnmount(() => {
    document.removeEventListener("click", copyListener);

    // Clean up scroll listener
    if (scrollableElem.value) {
      scrollableElem.value.removeEventListener("scroll", debouncedHandleScroll);
    }

    // Clean up video lazy loading observer
    destroyVideoLazyLoading();

    // Clean up sync functions
    if (autoSyncInterval) {
      clearInterval(autoSyncInterval);
    }

    // Remove event listeners
    // if (window.syncCleanupFunctions) {
    //   window.syncCleanupFunctions.forEach((cleanup: Function) => cleanup())
    //   window.syncCleanupFunctions = []
    // }

    // Clear timeouts
    if (scrollTimeout) {
      clearTimeout(scrollTimeout);
    }
    if (resizeTimeout) {
      clearTimeout(resizeTimeout);
    }

    // Final sync if needed
    if (syncStatus.value.hasUnsyncedChanges) {
      syncToServer()
    }
  });
});

// Move onUpdated outside of onMounted
onUpdated(() => {
  // Check for new video containers after DOM updates
  observeNewVideoContainers();
});

function updateAuthData(data: Partial<{ username: string; email: string; password: string }>) {
  Object.assign(authData.value, data)
}

function setShowCreateSession(value: boolean) {
  showCreateSession.value = value
}
</script>

<template>
  <div class="flex h-[100vh]">
    <!-- Custom Confirmation Dialog -->
    <div v-if="confirmDialog.visible"
      class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg p-6 max-w-md w-full mx-4 shadow-2xl">
        <div class="flex items-center gap-3 mb-4">
          <i :class="confirmDialog.type === 'danger' ? 'pi pi-exclamation-triangle text-red-500' :
            confirmDialog.type === 'warning' ? 'pi pi-exclamation-circle text-orange-500' :
              'pi pi-info-circle text-blue-500'" class="text-2xl"></i>
          <h3 class="text-lg font-semibold text-gray-900">{{ confirmDialog.title }}</h3>
        </div>

        <p class="text-gray-700 mb-6 leading-relaxed whitespace-pre-line">{{ confirmDialog.message }}</p>

        <div class="flex gap-3 justify-end">
          <button @click="confirmDialog.visible = false"
            class="px-4 py-2 text-gray-600 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors">
            {{ confirmDialog.cancelText }}
          </button>
          <button @click="() => { confirmDialog.onConfirm(); confirmDialog.visible = false }" :class="confirmDialog.type === 'danger' ? 'bg-red-600 hover:bg-red-700' :
            confirmDialog.type === 'warning' ? 'bg-orange-600 hover:bg-orange-700' :
              'bg-blue-600 hover:bg-blue-700'" class="px-4 py-2 text-white rounded-lg transition-colors">
            {{ confirmDialog.confirmText }}
          </button>
        </div>
      </div>
    </div>

    <!-- Sidebar -->
    <SideNav v-if="isAuthenticated()" :data="{
      chats,
      currentChatId,
      parsedUserDetails,
      screenWidth,
      isCollapsed,
      syncStatus,
      isAuthenticated
    }" :functions="{
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

    <!-- Main Chat Window -->
    <div
      :class="screenWidth > 720 && isAuthenticated() ? (!isCollapsed ?
        'flex-grow flex flex-col items-center justify-center ml-[270px] font-light text-sm transition-all duration-300 ease-in-out'
        :
        'flex-grow flex flex-col items-center justify-center ml-[60px] font-light text-sm transition-all duration-300 ease-in-out'
      )
        : 'text-sm font-light flex-grow items-center justify-center flex flex-col transition-all duration-300 ease-in-out'">

      <TopNav v-if="isAuthenticated()" :data="{
        currentChat,
        parsedUserDetails,
        screenWidth,
        isCollapsed,
        isSidebarHidden,
        syncStatus,
        isAuthenticated
      }" :functions="{
        hideSidebar,
        deleteChat,
        createNewChat,
        renameChat,
        manualSync,
      }" />

      <div
        :class="(screenWidth > 720 && isAuthenticated()) ? 'h-screen flex flex-col items-center justify-center w-[85%]' : 'h-screen flex flex-col items-center justify-center'">
        <!-- Empty State -->
        <CreateSessView v-if="currentMessages.length === 0 || !isAuthenticated()" :chats="chats"
          :current-chat-id="currentChatId" :is-collapsed="isCollapsed" :parsed-user-details="parsedUserDetails"
          :screen-width="screenWidth" :sync-status="syncStatus" :is-authenticated="isAuthenticated"
          :is-loading="isLoading" :auth-step="authStep" :show-create-session="showCreateSession" :auth-data="authData"
          :current-messages="currentMessages" :validate-current-step="validateCurrentStep()"
          :set-show-input="setShowInput" :hide-sidebar="hideSidebar" :clear-all-chats="clearAllChats"
          :toggle-sidebar="toggleSidebar" :logout="logout" :create-new-chat="createNewChat"
          :switch-to-chat="switchToChat" :delete-chat="deleteChat" :rename-chat="renameChat" :manual-sync="manualSync"
          :handle-step-submit="handleStepSubmit" :prev-auth-step="prevAuthStep" :update-auth-data="updateAuthData"
          :set-show-create-session="setShowCreateSession" 
        />


        <!-- Chat Messages -->
        <div ref="scrollableElem" v-else-if="currentMessages.length !== 0 && isAuthenticated()"
          class="flex-grow no-scrollbar overflow-y-auto px-4 space-y-4 pt-[90px] pb-[120px]">
          <div v-for="(item, i) in currentMessages" :key="`chat-${i}`" class="flex flex-col gap-2">
            <!-- User Bubble -->
            <div class="flex justify-end chat-message">
              <div :class="screenWidth > 720 ? 'max-w-[70%]' : 'max-w-[95%]'"
                class="bg-gray-50 text-black p-3 rounded-2xl prose prose-sm max-w-none chat-bubble">
                <p class="text-xs opacity-80 text-right mb-1">{{ parsedUserDetails?.username || "You" }}</p>
                <div v-html="renderMarkdown(item.prompt || '')"></div>

                <!-- Link Previews Section for User Messages -->
                <div v-if="extractUrls(item.prompt || '').length > 0" class="mt-3">
                  <div v-for="url in extractUrls(item.prompt || '').slice(0, 3)" :key="`user-${i}-${url}`">
                    <div v-if="linkPreviewCache.get(url)"
                      v-html="LinkPreviewComponent({ preview: linkPreviewCache.get(url)! })"></div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Bot Bubble -->
            <div class="flex justify-start relative">
              <div :class="screenWidth > 720 ? 'max-w-[70%]' : 'max-w-[95%]'"
                class="bg-none chat-message leading-relaxed text-black p-3 rounded-2xl prose prose-sm max-w-none">

                <!-- Loading state -->
                <div v-if="item.response === '...'" class="flex items-center gap-2 text-gray-500">
                  <i class="pi pi-spin pi-spinner"></i>
                  <span>Thinking...</span>
                </div>

                <!-- Regular response with enhanced link handling -->
                <div v-else>
                  <div v-html="renderMarkdown(item.response || '')"></div>

                  <!-- Link Previews Section -->
                  <div v-if="extractUrls(item.response || '').length > 0" class="mt-3">
                    <div v-for="url in extractUrls(item.response || '').slice(0, 3)" :key="url">
                      <div v-if="linkPreviewCache.get(url)"
                        v-html="LinkPreviewComponent({ preview: linkPreviewCache.get(url)! })"></div>
                    </div>
                  </div>
                </div>

                <!-- Actions (hidden during loading) -->
                <div v-if="item.response !== '...'" class="flex gap-3 mt-2 text-gray-500 text-sm">
                  <button @click="copyResponse(item.response, i)"
                    class="flex items-center gap-1 hover:text-blue-600 transition-colors">
                    <i class="pi pi-copy"></i>
                    {{ copiedIndex === i ? 'Copied!' : 'Copy' }}
                  </button>

                  <button @click="shareResponse(item.response, item.prompt)"
                    class="flex items-center gap-1 hover:text-green-600 transition-colors">
                    <i class="pi pi-share-alt"></i> Share
                  </button>
                  <button @click="refreshResponse(item.prompt)" :disabled="isLoading"
                    class="flex items-center gap-1 hover:text-orange-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
                    <i class="pi pi-refresh"></i> Refresh
                  </button>
                  <button @click="deleteMessage(i)" :disabled="isLoading"
                    class="flex items-center gap-1 hover:text-red-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
                    <i class="pi pi-trash"></i> Delete
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Scroll to Bottom Button -->
        <button v-if="showScrollDownButton && currentMessages.length !== 0 && isAuthenticated()" @click="scrollToBottom"
          class="fixed bottom-24 bg-gray-50 text-gray-500 border px-3 h-[34px] rounded-full shadow-lg hover:bg-gray-100 transition-colors z-20">
          <div class="flex gap-2 items-center justify-center w-full font-semibold h-full">
            <i class="pi pi-arrow-down text-center"></i>
            <p>Scroll To Bottom</p>
          </div>
        </button>

        <!-- Input -->
        <div v-if="(currentMessages.length !== 0 || showInput === true) && isAuthenticated()" :style="screenWidth > 720 && !isCollapsed ? 'left:270px;' :
          screenWidth > 720 && isCollapsed ? 'left:60px;' : 'left:0px;'"
          class="bg-white z-20 bottom-0 right-0 fixed pb-5 px-5">
          <div class="flex items-center justify-center w-full">
            <form @submit="handleSubmit"
              :class="screenWidth > 720 ? 'relative flex px-3 py-2 border-2 shadow rounded-2xl items-center gap-2 w-[85%]' : 'relative flex px-3 py-2 border-2 shadow rounded-2xl w-full items-center gap-2'">
              <textarea required id="prompt" name="prompt" @keydown="onEnter" @input="autoGrow" :disabled="isLoading"
                rows="1" class="flex-grow py-2 bg-white text-sm 
                      outline-none resize-none border-none
                      max-h-[200px] overflow-auto leading-relaxed
                      disabled:opacity-50 disabled:cursor-not-allowed"
                :placeholder="isLoading ? 'Please wait...' : 'Ask me a question...'"></textarea>
              <button type="submit" :disabled="isLoading" class="rounded-lg w-[26px] h-[26px] flex items-center justify-center transition-colors
                      text-white bg-blue-600 hover:bg-blue-500 disabled:cursor-not-allowed disabled:opacity-50
                      disabled:bg-gray-400 flex-shrink-0">
                <i v-if="!isLoading" class="pi pi-arrow-up text-sm"></i>
                <i v-else class="pi pi-spin pi-spinner text-sm"></i>
              </button>
            </form>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
