<script lang="ts" setup>
import type { ComputedRef } from "vue"
import { ref, onMounted, nextTick, computed } from "vue"
import { marked } from "marked"
import hljs from "highlight.js"
import "highlight.js/styles/night-owl.css"
import SideNav from "../components/SideNav.vue"
import TopNav from "../components/TopNav.vue"
import type { Chat, ConfirmDialogOptions, CurrentChat, LinkPreview, Res } from "@/types"
import { toast } from 'vue-sonner'

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
let showInput = ref(false)
// Track copied state for each response by index
const copiedIndex = ref<number | null>(null)
let screenWidth = ref(screen.width)
// local state for collapse toggle
const isCollapsed = ref(false)
const isSidebarHidden = ref(true)

let userDetails: any = localStorage.getItem("userdetails")
let parsedUserDetails: any = userDetails ? JSON.parse(userDetails) : null

// Chat management state
const currentChatId = ref<string>('')
const chats = ref<Chat[]>([])
let isLoading = ref(false)
let expanded = ref<boolean[]>([])

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

// Save chats to localStorage
function saveChats() {
  try {
    localStorage.setItem('chats', JSON.stringify(chats.value))
    localStorage.setItem('currentChatId', currentChatId.value)
  } catch (error) {
    console.error('Failed to save chats:', error)
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

// Switch to specific chat
function switchToChat(chatId: string) {
  if (chats.value.find(chat => chat.id === chatId)) {
    currentChatId.value = chatId
    updateExpandedArray()
    localStorage.setItem('currentChatId', currentChatId.value)

    nextTick(() => {
      scrollToBottom()
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

      // If we deleted the current chat, switch to another one
      if (currentChatId.value === chatId) {
        if (chats.value.length > 0) {
          currentChatId.value = chats.value[0].id
        } else {
          currentChatId.value = ''
        }
        updateExpandedArray()
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
      currentChat.value!.messages.splice(messageIndex, 1)
      expanded.value.splice(messageIndex, 1)

      // Update chat's updatedAt timestamp
      currentChat.value!.updatedAt = new Date().toISOString()

      // Update title if we deleted the first message
      if (messageIndex === 0 && currentChat.value!.messages.length > 0) {
        const firstMessage = currentChat.value!.messages[0].prompt || currentChat.value!.messages[0].response
        currentChat.value!.title = generateChatTitle(firstMessage)
      } else if (currentChat.value!.messages.length === 0) {
        currentChat.value!.title = 'New Chat'
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


// Clear all chats
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
    }
  })
}

// Enhanced rename chat with success notification
function renameChat(chatId: string, newTitle: string) {
  const chat = chats.value.find(c => c.id === chatId)
  if (chat && newTitle.trim()) {
    const oldTitle = chat.title
    chat.title = newTitle.trim()
    chat.updatedAt = new Date().toISOString()
    saveChats()
  }
}

// ---------- Link Preview Functions ----------

// Extract URLs from text using regex
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
    // Using a CORS proxy service for demonstration
    // In production, you'd want your own backend endpoint
    const proxyUrl = `https://api.allorigins.win/get?url=${encodeURIComponent(url)}`

    const response = await fetch(proxyUrl)
    const data = await response.json()

    if (data.contents) {
      const parser = new DOMParser()
      const doc = parser.parseFromString(data.contents, 'text/html')

      // Extract meta tags
      const title = doc.querySelector('meta[property="og:title"]')?.getAttribute('content') ||
        doc.querySelector('title')?.textContent ||
        'No title'

      const description = doc.querySelector('meta[property="og:description"]')?.getAttribute('content') ||
        doc.querySelector('meta[name="description"]')?.getAttribute('content') ||
        ''

      const image = doc.querySelector('meta[property="og:image"]')?.getAttribute('content') ||
        doc.querySelector('meta[name="twitter:image"]')?.getAttribute('content') ||
        ''

      const domain = new URL(url).hostname

      const updatedPreview: LinkPreview = {
        url,
        title: title.slice(0, 100), // Limit title length
        description: description.slice(0, 200), // Limit description length
        image,
        domain,
        loading: false,
        error: false
      }

      linkPreviewCache.value.set(url, updatedPreview)
      // Save to localStorage after successful fetch
      saveLinkPreviewCache()
      return updatedPreview
    }
  } catch (error) {
    console.error('Failed to fetch link preview:', error)
  }

  // Fallback preview
  const fallbackPreview: LinkPreview = {
    url,
    title: new URL(url).hostname,
    domain: new URL(url).hostname,
    loading: false,
    error: true
  }

  linkPreviewCache.value.set(url, fallbackPreview)
  // Save even error states to avoid repeated failures
  saveLinkPreviewCache()
  return fallbackPreview
}

// Component for rendering link previews
function LinkPreviewComponent({ preview }: { preview: LinkPreview }) {
  if (preview.loading) {
    return `
      <div class="link-preview loading border border-gray-200 rounded-lg p-3 my-2 bg-gray-50">
        <div class="flex items-center gap-2">
          <i class="pi pi-spin pi-spinner text-gray-400"></i>
          <span class="text-sm text-gray-500">Loading preview...</span>
        </div>
      </div>
    `
  }

  if (preview.error) {
    return `
      <div class="link-preview error border border-gray-200 rounded-lg p-3 my-2 bg-gray-50">
        <div class="flex items-center gap-2">
          <i class="pi pi-external-link text-gray-400"></i>
          <a href="${preview.url}" target="_blank" rel="noopener noreferrer" 
             class="text-blue-600 hover:text-blue-800 text-sm font-medium">
            ${preview.domain}
          </a>
        </div>
      </div>
    `
  }

  return `
    <div class="link-preview border border-gray-200 rounded-lg overflow-hidden my-2 bg-white hover:shadow-md transition-shadow">
      <a href="${preview.url}" target="_blank" rel="noopener noreferrer" class="block">
        ${preview.image ? `
          <div class="aspect-video w-full overflow-hidden bg-gray-100">
            <img src="${preview.image}" alt="${preview.title}" 
                 class="w-full h-full object-cover"
                 onerror="this.parentElement.style.display='none'">
          </div>
        ` : ''}
        <div class="p-3">
          <div class="flex items-start justify-between gap-2">
            <div class="flex-1 min-w-0">
              <h4 class="font-medium text-gray-900 text-sm line-clamp-2 mb-1">${preview.title}</h4>
              ${preview.description ? `
                <p class="text-gray-600 text-xs line-clamp-2 mb-2">${preview.description}</p>
              ` : ''}
              <div class="flex items-center gap-1 text-xs text-gray-500">
                <i class="pi pi-external-link"></i>
                <span>${preview.domain}</span>
              </div>
            </div>
          </div>
        </div>
      </a>
    </div>
  `
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

// Update the existing handleAuth function to work with the new data structure
function handleAuth(e: Event) {
  e.preventDefault()

  const { username, email, password } = authData.value
  const createdAt = new Date().toISOString()

  if (!email || !password || !username) {
    toast.error('Please fill in all required fields', {
      duration: 4000,
      description: ''
    })
    return
  }

  const userData = {
    email,
    username,
    createdAt,
    sessionId: btoa(email + ':' + password + ':' + username)
  }

  try {
    localStorage.setItem('userdetails', JSON.stringify(userData))
    parsedUserDetails = userData

    // Load existing chats after authentication
    loadChats()
    window.location.reload()

    nextTick(() => {
      const textarea = document.getElementById("prompt") as HTMLTextAreaElement
      if (textarea) textarea.focus()
    })
  } catch (err) {
    console.error('Failed to save user data:', err)
    toast.error('Failed to create session. Please try again.', {
      duration: 4000,
      description: ''
    })
  }
}

function logout() {
  showConfirmDialog({
    visible: true,
    title: 'Logout Confirmation',
    message: 'Are you sure you want to logout? This will clear your session on this device.',
    type: 'warning',
    confirmText: 'Logout',
    onConfirm: () => {
      try {
        localStorage.removeItem('userdetails')
        localStorage.removeItem('isCollapsed')
        localStorage.removeItem('currentChatId')
        // Keep chats and link previews cached even after logout

        parsedUserDetails = null
        chats.value = []
        currentChatId.value = ''
        expanded.value = []
        showInput.value = false
        isCollapsed.value = false

      } catch (err) {
        console.error('Error during logout:', err)
        toast.error('Error during logout. Please try again.', {
          duration: 4000,
          description: ''
        })
      }
    }
  })
}


function isAuthenticated(): boolean {
  return parsedUserDetails && parsedUserDetails.email && parsedUserDetails.username && parsedUserDetails.sessionId
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

// Enhanced submit with better notifications
async function handleSubmit(e?: any, retryPrompt?: string) {
  e?.preventDefault?.()

  let promptValue = retryPrompt || e?.target?.prompt?.value?.trim()
  if (!promptValue || isLoading.value) return

  if (!isAuthenticated()) {
    toast.warning('Please create a session first', {
      duration: 4000,
      description: 'You need to be logged in.'
    })
    return
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

  await nextTick()
  scrollToBottom()

  try {
    let url = `https://wrapper.villebiz.com/v1/genai`
    let response = await fetch(url, {
      method: "POST",
      body: JSON.stringify(promptValue),
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

function scrollToBottom() {
  const elem = document.getElementById("scrollableElem")
  if (elem) {
    elem.scrollIntoView({ behavior: "smooth", block: "end" })
  }
}

function hideSidebar() {
  const sideNav = document.getElementById("side_nav")
  if (sideNav) {
    if (sideNav.classList.contains("none")) {
      sideNav.classList.add("w-full", "bg-white", "z-20", "fixed", "top-0", "left-0", "bottom-0", "border-r-[1px]", "flex", "flex-col", "transition-all", "duration-300", "ease-in-out")
    } else {
      sideNav.classList.remove("w-full", "bg-white", "z-20", "fixed", "top-0", "left-0", "bottom-0", "border-r-[1px]", "flex", "flex-col", "transition-all", "duration-300", "ease-in-out")
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
  // Load existing state
  const saved = localStorage.getItem("isCollapsed")
  if (saved && saved !== 'null') {
    try {
      isCollapsed.value = JSON.parse(saved)
    } catch (err) {
      console.error('Error parsing isCollapsed:', err)
    }
  }

  // Load current chat ID
  const savedChatId = localStorage.getItem('currentChatId')
  if (savedChatId) {
    currentChatId.value = savedChatId
  }

  // Load cached link previews
  loadLinkPreviewCache()

  // Load chats if authenticated
  if (isAuthenticated()) {
    loadChats()

    // Pre-process existing chat links on page load
    if (currentMessages.value.length > 0) {
      currentMessages.value.forEach((item, index) => {
        // Process links in prompts
        if (item.prompt) {
          const promptUrls = extractUrls(item.prompt)
          promptUrls.slice(0, 3).forEach(url => {
            if (!linkPreviewCache.value.has(url)) {
              fetchLinkPreview(url).then(() => {
                linkPreviewCache.value = new Map(linkPreviewCache.value)
              })
            }
          })
        }

        // Process links in responses
        if (item.response && item.response !== "...") {
          const responseUrls = extractUrls(item.response)
          responseUrls.slice(0, 3).forEach(url => {
            if (!linkPreviewCache.value.has(url)) {
              fetchLinkPreview(url).then(() => {
                linkPreviewCache.value = new Map(linkPreviewCache.value)
              })
            }
          })
        }
      })
    }
  }

  scrollToBottom()

  document.addEventListener("click", (e: any) => {
    if (e.target && e.target.classList.contains("copy-button")) {
      const code = decodeURIComponent(e.target.getAttribute("data-code"))
      copyCode(code, e.target)
    }
  })

  if (showInput.value || currentMessages.value.length > 0) {
    nextTick(() => {
      const textarea = document.getElementById("prompt") as HTMLTextAreaElement
      if (textarea) textarea.focus()
    })
  }
})
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
      isCollapsed
    }" :functions="{
      setShowInput,
      hideSidebar,
      clearAllChats,
      toggleSidebar,
      logout,
      createNewChat,
      switchToChat,
      deleteChat,
      renameChat
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
        isSidebarHidden
      }" :functions="{
        hideSidebar,
        deleteChat,
        createNewChat,
        renameChat
      }" />

      <div :class="(screenWidth > 720 && isAuthenticated()) ? 'h-screen flex flex-col items-center justify-center w-[85%]' : 'h-screen flex flex-col items-center justify-center'">
        <!-- Empty State -->
        <div v-if="currentMessages.length === 0 || !isAuthenticated()"
          class="flex flex-col items-center justify-center h-[90vh]">
            <div class="max-md:flex-col flex gap-10 items-center justify-center h-full w-full max-md:px-5">
              <div class="flex flex-col md:flex-grow items-center gap-3 text-gray-600">
                <div class="rounded-full bg-gray-200 w-[60px] h-[60px] flex justify-center items-center">
                  <span class="pi pi-comment text-lg"></span>
                </div>
                <p class="text-3xl font-semibold">{{ parsedUserDetails?.username || 'Gemmie' }}</p>
                <div class="text-center text-base md:max-w-[400px]">
                  <p>Your private AI assistant.</p>
                  <p class="text-sm text-gray-400">
                    We focus on privacy and security. Your data never leaves your device.
                    All your chats are stored locally in your browser.
                    Therefore, please make sure to back up your chats if you clear your browser data or switch devices.
                  </p>
                </div>
                <button v-if="isAuthenticated()" @click="setShowInput"
                  class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-500 transition-colors">
                  Write a prompt
                </button>
              </div>

              <div v-if="!isAuthenticated()" class="flex-grow text-sm  md:px-4 px-1 relative overflow-hidden">
                <!-- Progress indicator -->
                <div class="flex justify-center mb-6">
                  <div class="flex items-center space-x-2">
                    <div v-for="step in 3" :key="step" :class="step <= authStep ? 'bg-blue-600' : 'bg-gray-300'"
                      class="w-3 h-3 rounded-full transition-colors duration-300">
                    </div>
                  </div>
                </div>

                <!-- Multi-step form container -->
                <div class="relative h-80">
                  <!-- Step 1: Username -->
                  <div :class="authStep === 1 ? 'translate-x-0 opacity-100' :
                    authStep > 1 ? '-translate-x-full opacity-0' : 'translate-x-full opacity-0'"
                    class="absolute inset-0 transition-all duration-500 ease-in-out transform">
                    <div class="text-center mb-6">
                      <h2 class="text-xl font-semibold text-gray-900 mb-2">Welcome!</h2>
                      <p class="text-gray-600">Let's start by creating your username</p>
                    </div>

                    <form @submit="handleStepSubmit" class="space-y-4">
                      <div>
                        <label class="block text-sm font-medium text-gray-700 mb-2">
                          Choose a username
                        </label>
                        <input v-model="authData.username" required type="text" placeholder="johndoe"
                          class="border border-gray-300 rounded-lg px-4 py-3 w-full text-base focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
                          :class="authData.username && !validateCurrentStep() ? 'border-red-300' : ''" />
                        <p class="text-xs text-gray-500 mt-1">This will be your display name</p>
                      </div>

                      <button type="submit" :disabled="!validateCurrentStep()"
                        class="w-full bg-blue-600 text-white rounded-lg px-4 py-3 font-medium hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed transition-colors duration-200">
                        Continue
                      </button>
                    </form>
                  </div>

                  <!-- Step 2: Email -->
                  <div :class="authStep === 2 ? 'translate-x-0 opacity-100' :
                    authStep > 2 ? '-translate-x-full opacity-0' : 'translate-x-full opacity-0'"
                    class="absolute inset-0 transition-all duration-500 ease-in-out transform">
                    <div class="text-center mb-6">
                      <h2 class="text-xl font-semibold text-gray-900 mb-2">Hi {{ authData.username }}!</h2>
                      <p class="text-gray-600">What's your email address?</p>
                    </div>

                    <form @submit="handleStepSubmit" class="space-y-4">
                      <div>
                        <label class="block text-sm font-medium text-gray-700 mb-2">
                          Email address
                        </label>
                        <input v-model="authData.email" required type="email" placeholder="johndoe@example.com"
                          class="border border-gray-300 rounded-lg px-4 py-3 w-full text-base focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
                          :class="authData.email && !validateCurrentStep() ? 'border-red-300' : ''" />
                        <p class="text-xs text-gray-500 mt-1">Used for session identification only</p>
                      </div>

                      <div class="flex gap-3">
                        <button type="button" @click="prevAuthStep"
                          class="flex-1 bg-gray-100 text-gray-700 rounded-lg px-4 py-3 font-medium hover:bg-gray-200 transition-colors duration-200">
                          Back
                        </button>
                        <button type="submit" :disabled="!validateCurrentStep()"
                          class="flex-1 bg-blue-600 text-white rounded-lg px-4 py-3 font-medium hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed transition-colors duration-200">
                          Continue
                        </button>
                      </div>
                    </form>
                  </div>

                  <!-- Step 3: Password -->
                  <div :class="authStep === 3 ? 'translate-x-0 opacity-100' : 'translate-x-full opacity-0'"
                    class="absolute inset-0 transition-all duration-500 ease-in-out transform">
                    <div class="text-center mb-6">
                      <h2 class="text-xl font-semibold text-gray-900 mb-2">Almost there!</h2>
                      <p class="text-gray-600">Create a secure password</p>
                    </div>

                    <form @submit="handleStepSubmit" class="space-y-4">
                      <div>
                        <label class="block text-sm font-medium text-gray-700 mb-2">
                          Password
                        </label>
                        <input v-model="authData.password" required type="password" placeholder="Enter a secure password"
                          minlength="6"
                          class="border border-gray-300 rounded-lg px-4 py-3 w-full text-base focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
                          :class="authData.password && !validateCurrentStep() ? 'border-red-300' : ''" />
                        <div class="mt-2">
                          <div class="flex items-center gap-2 text-xs">
                            <div :class="authData.password.length >= 6 ? 'text-green-600' : 'text-gray-400'"
                              class="flex items-center gap-1">
                              <i :class="authData.password.length >= 6 ? 'pi pi-check' : 'pi pi-circle'"
                                class="text-xs"></i>
                              <span>At least 6 characters</span>
                            </div>
                          </div>
                        </div>
                      </div>

                      <div class="flex gap-3">
                        <button type="button" @click="prevAuthStep"
                          class="flex-1 bg-gray-100 text-gray-700 rounded-lg px-4 py-3 font-medium hover:bg-gray-200 transition-colors duration-200">
                          Back
                        </button>
                        <button type="submit" :disabled="!validateCurrentStep()"
                          class="flex-1 bg-blue-600 text-white rounded-lg px-4 py-3 font-medium hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed transition-colors duration-200">
                          <i class="pi pi-check mr-2"></i>
                          Create Session
                        </button>
                      </div>
                    </form>
                  </div>
                </div>

                <!-- Footer note -->
                <div class="mt-6 text-center">
                  <p class="text-xs text-gray-400 leading-relaxed">
                    Your credentials are only stored locally on your device for session management.
                    <br>All data stays private and secure.
                  </p>
                </div>
              </div>
            </div>

            <div>
              <p class="text-sm mt-2 text-gray-400">Gemmie can make mistakes. Check important info.</p>
            </div>
        </div>

        <!-- Chat Messages -->
        <div v-else-if="currentMessages.length !== 0 && isAuthenticated()"
          class="flex-grow no-scrollbar overflow-y-auto px-4 space-y-4 pt-[90px] pb-[120px]">
          <div v-for="(item, i) in currentMessages" :key="`chat-${i}`" class="flex flex-col gap-2">
            <!-- User Bubble -->
            <div class="flex justify-end chat-message ">
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
          <div id="scrollableElem"></div>
        </div>

        <!-- Input -->
        <div v-if="(currentMessages.length !== 0 || showInput === true) && isAuthenticated()" :style="screenWidth > 720 && !isCollapsed ? 'left:270px;' :
          screenWidth > 720 && isCollapsed ? 'left:60px;' : 'left:0px;'" class="bg-white z-20 bottom-0 right-0 fixed pb-5 px-5">
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