<script lang="ts" setup>
import type { Ref } from "vue"
import { ref, onMounted, nextTick, computed, onBeforeUnmount, inject, watch, onUnmounted } from "vue"
import { marked } from "marked"
import hljs from "highlight.js"
import "highlight.js/styles/night-owl.css"
import SideNav from "../components/SideNav.vue"
import TopNav from "../components/TopNav.vue"
import type { Chat, ConfirmDialogOptions, CurrentChat, LinkPreview, Res } from "@/types"
import { toast } from 'vue-sonner'
import { destroyVideoLazyLoading, initializeVideoLazyLoading, observeNewVideoContainers, pauseVideo, playEmbeddedVideo, playSocialVideo, resumeVideo, showVideoControls, stopDirectVideo, stopVideo, toggleDirectVideo, updateVideoControls } from "@/utils/videoProcessing"
import { onUpdated } from "vue"
import { extractUrls, generateChatTitle, copyCode, isPromptTooShort, WRAPPER_URL } from "@/utils/globals"
import CreateSessView from "./CreateSessView.vue"
import router from "@/router"
import { copyPasteContent, detectContentType } from "@/utils/previewPasteContent"
import PastePreviewModal from "@/components/Modals/PastePreviewModal.vue"

// Inject global state
const globalState = inject('globalState') as {
  handleAuth: (data: {
    username: string;
    email: string;
    password: string;
  }) => any,
  screenWidth: Ref<number>,
  confirmDialog: Ref<ConfirmDialogOptions>,
  isCollapsed: Ref<boolean>,
  isSidebarHidden: Ref<boolean>,
  authData: Ref<{ username: string; email: string; password: string; agreeToTerms: boolean; }>,
  syncStatus: Ref<{ lastSync: Date | null; syncing: boolean; hasUnsyncedChanges: boolean; }>,
  isAuthenticated: Ref<boolean>,
  parsedUserDetails: any,
  planStatus: Ref<{ status: string; timeLeft: string; expiryDate: string; isExpired: boolean; }>,
  currentChatId: Ref<string>,
  chats: Ref<Chat[]>
  logout: () => void,
  isLoading: Ref<boolean>,
  expanded: Ref<boolean[]>,
  showInput: Ref<boolean>,
  showConfirmDialog: (options: ConfirmDialogOptions) => void,
  hideSidebar: () => void,
  setShowInput: () => void,
  clearAllChats: () => void,
  switchToChat: (chatId: string) => void,
  createNewChat: (initialMessage?: string) => void,
  deleteChat: (chatId: string) => void,
  renameChat: (chatId: string, newTitle: string) => void,
  deleteMessage: (messageIndex: number) => void,
  scrollableElem: Ref<HTMLElement | null>,
  showScrollDownButton: Ref<boolean>,
  handleScroll: () => void,
  scrollToBottom: () => void,
  saveChats: () => void,
  linkPreviewCache: Ref<Map<string, LinkPreview>>,
  fetchLinkPreview: (url: string) => Promise<LinkPreview>,
  loadLinkPreviewCache: () => void,
  saveLinkPreviewCache: () => void,
  syncFromServer: (data?: any) => void,
  syncToServer: () => void,
  currentChat: Ref<CurrentChat | undefined>,
  currentMessages: Ref<Res[]>,
  linkPreview: LinkPreview,
  updateExpandedArray: () => void,
  apiCall: (endpoint: string, options: RequestInit) => any,
  manualSync: () => Promise<any>
  toggleSidebar: () => void,
  setupAutoSync: () => void,
  autoSyncInterval: any,
  isFreeUser: Ref<boolean>,
}

const {
  screenWidth,
  confirmDialog,
  isCollapsed,
  isSidebarHidden,
  authData,
  syncStatus,
  isAuthenticated,
  currentChatId,
  chats,
  planStatus,
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
  linkPreview,
  currentMessages,
  updateExpandedArray,
  apiCall,
  manualSync,
  toggleSidebar,
  setupAutoSync,
  autoSyncInterval,
  handleAuth,
  isFreeUser,
} = globalState
let parsedUserDetails: Ref<any> = globalState.parsedUserDetails

// ---------- State ----------
const authStep = ref(1)
const showCreateSession = ref(false)
const copiedIndex = ref<number | null>(null)
const requestCount = ref(0)
const FREE_REQUEST_LIMIT = 5
const now = ref(Date.now())

const isRecording = ref(false)
const isTranscribing = ref(false)
const transcribedText = ref('')
const voiceRecognition = ref<any | null>(null)
const microphonePermission = ref<'granted' | 'denied' | 'prompt'>('prompt')
const transcriptionDuration = ref(0)
let transcriptionTimer: number | null = null
let updateTimeout: number | null = null

const showPasteModal = ref(false)
const currentPasteContent = ref<{
  content: string;
  wordCount: number;
  charCount: number;
  type: 'text' | 'code' | 'json' | 'markdown' | 'xml' | 'html';
} | null>(null)

const pastePreview = ref<{
  content: string;
  wordCount: number;
  charCount: number;
  show: boolean;
} | null>(null)

// ---------- Request Limit Functions ----------
function loadRequestCount() {
  try {
    if (!parsedUserDetails.value) {
      requestCount.value = 0
      return
    }

    // Check if user should have request limits
    const shouldHaveLimit = isFreeUser.value ||
      planStatus.value.isExpired ||
      planStatus.value.status === 'no-plan' ||
      planStatus.value.status === 'expired'

    if (!shouldHaveLimit) {
      requestCount.value = 0
      try {
        localStorage.removeItem('requestCount')
      } catch (error) {
        console.error('Failed to clear request count for unlimited user:', error)
      }
      return
    }

    const saved = localStorage.getItem('requestCount')
    if (saved) {
      try {
        const data = JSON.parse(saved)
        const now = new Date().getTime()

        if (typeof data !== 'object' || typeof data.timestamp !== 'number' || typeof data.count !== 'number') {
          console.warn('Invalid request count data format, resetting')
          requestCount.value = 0
          saveRequestCount()
          return
        }

        // Check if 24 hours have passed since last timestamp
        const timeDiff = now - data.timestamp
        const twentyFourHours = 24 * 60 * 60 * 1000

        if (timeDiff > twentyFourHours) {
          // Reset count after 24 hours
          requestCount.value = 0
          saveRequestCount() // Save with new timestamp
        } else {
          requestCount.value = Math.max(0, Math.min(data.count, FREE_REQUEST_LIMIT))
        }
      } catch (parseError) {
        console.error('Failed to parse request count data:', parseError)
        requestCount.value = 0
        localStorage.removeItem('requestCount')
        saveRequestCount() // Initialize with current timestamp
      }
    } else {
      // No saved data, initialize
      requestCount.value = 0
      saveRequestCount()
    }
  } catch (error) {
    console.error('Failed to load request count:', error)
    requestCount.value = 0
  }
}

// Updated saveRequestCount function for better logic
function saveRequestCount() {
  try {
    // Save if user has limitations (free, expired, or no plan)
    const shouldHaveLimit = isFreeUser.value ||
      planStatus.value?.isExpired ||
      planStatus.value?.status === 'expired' ||
      planStatus.value?.status === 'no-plan'

    if (shouldHaveLimit) {
      const data = {
        count: Math.max(0, Math.min(requestCount.value, FREE_REQUEST_LIMIT)),
        timestamp: Date.now() // Always update timestamp when saving
      }

      localStorage.setItem('requestCount', JSON.stringify(data))
    } else {
      // Remove limits for unlimited users
      localStorage.removeItem('requestCount')
    }
  } catch (error) {
    console.error('Failed to save request count:', error)
  }
}

// Updated incrementRequestCount function
function incrementRequestCount() {
  try {
    const shouldHaveLimit = isFreeUser.value ||
      planStatus.value?.isExpired ||
      planStatus.value?.status === 'expired' ||
      planStatus.value?.status === 'no-plan'

    if (!shouldHaveLimit) {
      return // No limits for unlimited users
    }

    if (requestCount.value < FREE_REQUEST_LIMIT) {
      requestCount.value++
      saveRequestCount()
    }
  } catch (error) {
    console.error('Failed to increment request count:', error)
  }
}

// Add a function to check and reset count if needed (call this periodically)
function checkAndResetDailyCount() {
  try {
    const saved = localStorage.getItem('requestCount')
    if (!saved) return

    const data = JSON.parse(saved)
    const now = new Date().getTime()
    const timeDiff = now - data.timestamp
    const twentyFourHours = 24 * 60 * 60 * 1000

    if (timeDiff > twentyFourHours) {
      requestCount.value = 0
      saveRequestCount()

      // Optionally notify user
      toast.success('Daily request limit reset!', {
        duration: 4000,
        description: `You have ${FREE_REQUEST_LIMIT} new requests available.`
      })
    }
  } catch (error) {
    console.error('Failed to check daily reset:', error)
  }
}

function resetRequestCount() {
  try {
    requestCount.value = 0
    localStorage.removeItem('requestCount')
  } catch (error) {
    console.error('Failed to reset request count:', error)
  }
}

// Load chats from localStorage
function loadChats() {
  try {
    const stored = localStorage.getItem('chats')
    if (stored) {
      const parsedChats = JSON.parse(stored)
      if (Array.isArray(parsedChats)) {
        chats.value = parsedChats
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
    <div class="link-preview border border-gray-200 rounded-lg overflow-hidden my-2 bg-white hover:shadow-md transition-shadow w-fit max-w-full">
      ${hasVideo ? `
        <div class="w-full max-w-[500px]">
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
        <a href="${preview.url}" class="block w-full max-w-[400px]" target="_blank" rel="noopener noreferrer">
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

// ---------- Authentication Functions ----------
function nextAuthStep() {
  if (authStep.value < 4) {
    authStep.value++
  }
}

function prevAuthStep() {
  if (authStep.value > 1) {
    authStep.value--
  }
}

function validateCurrentStep(): boolean {
  try {
    switch (authStep.value) {
      case 1:
        const username = authData.value.username?.trim()
        return !!(username && username.length >= 2 && username.length <= 50)
      case 2:
        const email = authData.value.email?.trim()
        return !!(email &&
          email.length > 0 &&
          email.length <= 100 &&
          /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email))
      case 3:
        const password = authData.value.password
        return !!(password && password.length > 7 && password.length < 25)
      case 4:
        const agreeToTerms = authData.value.agreeToTerms
        return agreeToTerms
      default:
        return false
    }
  } catch (error) {
    console.error('Error validating current step:', error)
    return false
  }
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

// Debounced scroll handler to improve performance
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

// paste detection function
function detectLargePaste(text: string): boolean {
  const wordCount = text.trim().split(/\s+/).length
  const charCount = text.length
  return wordCount > 100 || charCount > 800
}

function handlePaste(e: ClipboardEvent) {
  try {
    const pastedText = e.clipboardData?.getData('text') || ''

    if (!pastedText.trim()) return

    if (detectLargePaste(pastedText)) {
      e.preventDefault()

      const wordCount = pastedText.trim().split(/\s+/).filter(word => word.length > 0).length
      const charCount = pastedText.length

      // Enhanced paste preview with proper content handling
      const processedContent = wordCount > 100 ? `#pastedText#${pastedText}` : pastedText

      pastePreview.value = {
        content: processedContent,
        wordCount,
        charCount,
        show: true
      }

      toast.info('Large content detected', {
        duration: 4000,
        description: `${wordCount} words, ${charCount} characters. Preview shown below.`
      })
    }
  } catch (error) {
    console.error('Error handling paste:', error)
    // Don't prevent default on error - let normal paste proceed
  }
}


function removePastePreview() {
  pastePreview.value = null

  // Also clear the textarea if it contains the preview content
  nextTick(() => {
    const textarea = document.getElementById("prompt") as HTMLTextAreaElement
    if (textarea && textarea.value.includes('#pastedText#')) {
      // Extract any non-pasted content
      const parts = textarea.value.split('#pastedText#')
      textarea.value = parts[0] || ''
      autoGrow({ target: textarea } as any)
    }
  })
}

// Fixed handleSubmit function
async function handleSubmit(e?: any, retryPrompt?: string) {
  e?.preventDefault?.()

  let promptValue = retryPrompt || e?.target?.prompt?.value?.trim()

  // If we have a paste preview, use that content instead
  if (pastePreview.value && pastePreview.value.show && !retryPrompt) {
    promptValue += pastePreview.value.content
    // Clear the paste preview
    pastePreview.value = null
  }

  let fabricatedPrompt = promptValue
  if (!promptValue || isLoading.value) return

  if (!isAuthenticated) {
    toast.warning('Please create a session first', {
      duration: 4000,
      description: 'You need to be logged in.'
    })
    return
  }

  // Check for daily reset before checking limits
  checkAndResetDailyCount()

  // Fixed: Check request limit for ALL users who should have limits
  const shouldHaveLimit = isFreeUser.value ||
    planStatus.value.isExpired ||
    planStatus.value.status === 'no-plan' ||
    planStatus.value.status === 'expired'

  if (shouldHaveLimit && requestCount.value >= FREE_REQUEST_LIMIT) {
    if (planStatus.value.isExpired) {
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

    // Increment request count for limited users
    incrementRequestCount()

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

  // Increment request count for limited users
  incrementRequestCount()

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

// Fixed computed properties for better logic
const isRequestLimitExceeded = computed(() => {
  const shouldHaveLimit = isFreeUser.value ||
    planStatus.value.isExpired ||
    planStatus.value.status === 'no-plan' ||
    planStatus.value.status === 'expired'

  return shouldHaveLimit && requestCount.value >= FREE_REQUEST_LIMIT
})

const shouldShowUpgradePrompt = computed(() => {
  const shouldHaveLimit = isFreeUser.value ||
    planStatus.value.isExpired ||
    planStatus.value.status === 'no-plan' ||
    planStatus.value.status === 'expired'

  return shouldHaveLimit && requestCount.value >= 3 && requestCount.value < FREE_REQUEST_LIMIT
})

const requestsRemaining = computed(() => {
  const shouldHaveLimit = isFreeUser.value ||
    planStatus.value.isExpired ||
    planStatus.value.status === 'no-plan' ||
    planStatus.value.status === 'expired'

  if (!shouldHaveLimit) return Infinity
  return Math.max(0, FREE_REQUEST_LIMIT - requestCount.value)
})

// Fixed input area template logic
const inputPlaceholderText = computed(() => {
  if (pastePreview.value && pastePreview.value.show) {
    return 'Large content ready to send...'
  }

  if (isRecording.value) {
    return screenWidth.value > 640 ? 'Speak now... (Click mic to stop)' : 'Speak now...'
  }

  if (isRequestLimitExceeded.value) {
    if (planStatus.value.isExpired) {
      return screenWidth.value > 640 ? 'Plan expired - renew to continue...' : 'Plan expired...'
    }
    return screenWidth.value > 640 ? 'Upgrade to continue chatting...' : 'Upgrade to continue...'
  }

  if (isLoading.value) {
    return 'Please wait...'
  }

  const shouldHaveLimit = isFreeUser.value ||
    planStatus.value.isExpired ||
    planStatus.value.status === 'no-plan' ||
    planStatus.value.status === 'expired'

  if (shouldHaveLimit) {
    return `Ask me a question... (${requestsRemaining.value} requests left)`
  }

  return 'Ask me a question...'
})

const inputDisabled = computed(() => {
  return isLoading.value || isRequestLimitExceeded.value
})

const showLimitExceededBanner = computed(() => {
  return isRequestLimitExceeded.value
})

const showUpgradeBanner = computed(() => {
  return shouldShowUpgradePrompt.value && !isRequestLimitExceeded.value
})

// Add connection checking before authentication
async function handleStepSubmit(e: Event) {
  e.preventDefault()

  if (!validateCurrentStep()) {
    // Show specific validation error
    let errorMsg = ''
    switch (authStep.value) {
      case 1:
        errorMsg = 'Username must be 2-50 characters'
        break
      case 2:
        errorMsg = 'Please enter a valid email address'
        break
      case 3:
        errorMsg = 'Password must be at least 7 characters'
        break
      case 4:
        errorMsg = 'Please accept the terms of service and privacy policy'
        break
    }

    if (errorMsg) {
      toast.warning(errorMsg, {
        duration: 3000,
        description: 'Please correct the error and try again'
      })
    }
    return
  }

  if (authStep.value < 4) {
    nextAuthStep()
  } else {
    // Check connection before attempting authentication
    toast.info('Connecting...', { duration: 2000 })

    try {
      // Show loading state
      isLoading.value = true

      // Final step - create session
      const response = await handleAuth(authData.value)

      // Check if response exists and has expected structure
      if (!response) {
        throw new Error('No response received from server')
      }

      if (response.error) {
        throw new Error(response.error)
      }

      if (response.data && response.success) {
        setShowCreateSession(false)

        // Reset form
        authStep.value = 1
        authData.value = { username: '', email: '', password: '', agreeToTerms: false }
        loadRequestCount() // Load request count after successful auth
        nextTick(() => {
          const textarea = document.getElementById("prompt") as HTMLTextAreaElement
          if (textarea) textarea.focus()
        })
      } else {
        throw new Error('Authentication failed - invalid response')
      }
    } catch (err: any) {
      console.error('Authentication error in handleStepSubmit:', err)

      // Show user-friendly error messages
      let errorMessage = 'Authentication Failed'
      let errorDescription = 'Please try again'

      if (err.message?.includes('timeout')) {
        errorMessage = 'Connection Timeout'
        errorDescription = 'Server took too long to respond. Please try again.'
      } else if (err.message?.includes('Failed to fetch') || err.message?.includes('NetworkError')) {
        errorMessage = 'Connection Error'
        errorDescription = 'Unable to reach server. Check your internet connection.'
      } else if (err.message?.includes('HTTP 4')) {
        errorMessage = 'Invalid Credentials'
        errorDescription = 'Please check your username, email, and password.'
      } else if (err.message?.includes('HTTP 5')) {
        errorMessage = 'Server Error'
        errorDescription = 'Server is experiencing issues. Please try again later.'
      } else if (err.message) {
        errorDescription = err.message
      }

      toast.error(errorMessage, {
        duration: 5000,
        description: errorDescription
      })

      // Don't reset the form on error - let user try again
    } finally {
      isLoading.value = false
    }
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

async function refreshResponse(oldPrompt?: string) {
  const chatIndex = chats.value.findIndex(chat => chat.id === currentChatId.value)
  if (chatIndex === -1) return

  const chat = chats.value[chatIndex]
  const msgIndex = chat.messages.findIndex(m => m.prompt === oldPrompt)
  if (msgIndex === -1) return

  const oldMessage = chat.messages[msgIndex]

  // Show placeholder while refreshing
  chat.messages[msgIndex] = {
    ...oldMessage,
    response: "refreshing...",
  }

  try {
    // Fetch new response using the same prompt
    let response = await fetch(WRAPPER_URL, {
      method: "POST",
      body: JSON.stringify(oldMessage.prompt),
      headers: { "content-type": "application/json" }
    })

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`)
    }

    let parseRes = await response.json()

    // Replace the same message with the refreshed response
    chat.messages[msgIndex] = {
      ...oldMessage,
      response: parseRes.error ? parseRes.error : parseRes.response,
      status: response.status,
    }

    chat.updatedAt = new Date().toISOString()
    saveChats()

    // Re-run link previews if needed
    await processLinksInResponse(msgIndex)

  } catch (err: any) {
    chat.messages[msgIndex] = {
      ...oldMessage,
      response: `⚠️ Failed to refresh response: ${err.message || "Unknown error"}`,
    }
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

// Function to close paste modal
function closePasteModal() {
  showPasteModal.value = false
  currentPasteContent.value = null

  // Restore body scroll
  document.body.style.overflow = 'auto'
}

// Keyboard handler for modal
function handleModalKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && showPasteModal.value) {
    closePasteModal()
  }
}

function PastePreviewComponent(content: string, wordCount: number, charCount: number, isClickable: boolean = false) {
  const preview = content.length > 200 ? content.substring(0, 200) + '...' : content

  // Proper HTML escaping
  const escapedPreview = preview
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
    .replace(/\n/g, '<br>')
    .replace(/\t/g, '&nbsp;&nbsp;&nbsp;&nbsp;')

  // Generate unique ID for this component
  const componentId = `paste-${Math.random().toString(36).substr(2, 9)}`

  // For clickable elements, add the data attributes and class to the main container
  const clickableAttributes = isClickable ?
    `data-paste-content="${encodeURIComponent(content)}" data-word-count="${wordCount}" data-char-count="${charCount}"` : ''

  const clickableClass = isClickable ? 'paste-preview-clickable cursor-pointer hover:bg-gray-200' : ''

  return `
    <div class="paste-preview border border-gray-300 rounded-lg overflow-hidden my-2 bg-gray-100 hover:shadow-md transition-all w-full ${clickableClass}" 
         id="${componentId}" ${clickableAttributes}>
      <div class="w-full">
        <div class="bg-gray-600 px-3 py-1 text-white text-xs font-medium flex items-center gap-2">
          <i class="pi pi-clipboard"></i>
          <span>PASTED CONTENT</span>
          <span class="ml-auto text-gray-200 hidden sm:inline">${wordCount} words • ${charCount} chars</span>
          <span class="ml-auto text-gray-200 sm:hidden">${charCount} chars</span>
          ${isClickable ? '<i class="pi pi-external-link ml-1 text-gray-300"></i>' : ''}
        </div>
        <div class="pb-3 px-3">
          <div class="relative">
            <div class="text-sm text-gray-800 leading-relaxed break-words whitespace-pre-wrap font-mono h-20 sm:h-24 overflow-hidden">
              ${escapedPreview}
            </div>
            <div class="absolute bottom-0 left-0 right-0 h-8 bg-gradient-to-t from-gray-100 via-gray-100/80 to-transparent pointer-events-none"></div>
          </div>
          <div class="flex items-center justify-between mt-2 text-xs text-gray-600">
            <span class="hidden sm:inline">${isClickable ? 'Click to view full content' : 'Large content detected'}</span>
            <span class="sm:hidden">${isClickable ? 'Tap to view' : 'Large content'}</span>
            ${!isClickable ? '<button class="remove-paste-preview text-gray-700 hover:text-gray-900 underline font-medium" type="button">Remove</button>' : ''}
          </div>
        </div>
      </div>
    </div>
  `
}

function handlePastePreviewClick(e: Event) {
  const target = e.target as HTMLElement

  // Check if the clicked element itself or any parent has the clickable class
  const clickableElement = target.closest('.paste-preview-clickable')

  if (clickableElement) {
    // Prevent event bubbling to avoid conflicts
    e.preventDefault()
    e.stopPropagation()

    const content = clickableElement.getAttribute('data-paste-content')
    const wordCount = clickableElement.getAttribute('data-word-count')
    const charCount = clickableElement.getAttribute('data-char-count')

    if (content && wordCount && charCount) {
      try {
        const decodedContent = decodeURIComponent(content)
        const parsedWordCount = parseInt(wordCount, 10)
        const parsedCharCount = parseInt(charCount, 10)

        openPasteModal(decodedContent, parsedWordCount, parsedCharCount)
      } catch (error) {
        console.error('Error parsing paste preview data:', error)
        toast.error('Error opening paste preview', {
          duration: 3000,
          description: 'Could not parse content data'
        })
      }
    }
  }
}

function handleRemovePastePreview(e: Event) {
  const target = e.target as HTMLElement

  if (target.classList.contains('remove-paste-preview')) {
    e.preventDefault()
    e.stopPropagation()
    removePastePreview()
  }
}

function setupPastePreviewHandlers() {
  // Remove existing listeners to avoid duplicates
  document.removeEventListener('click', handlePastePreviewClick, true)
  document.removeEventListener('click', handleRemovePastePreview, true)

  // Add event delegation with capture phase for better reliability
  document.addEventListener('click', handlePastePreviewClick, true)
  document.addEventListener('click', handleRemovePastePreview, true)

  console.log('Paste preview handlers setup complete') // Debug log
}

// Function to open paste modal
function openPasteModal(content: string, wordCount: number, charCount: number) {
  try {
    // Handle the #pastedText# prefix if present
    const actualContent = content.startsWith('#pastedText#') ? content.substring(12) : content

    // Detect content type - provide fallback if function not available
    let contentType: 'text' | 'code' | 'json' | 'markdown' | 'xml' | 'html' = 'text'

    if (typeof detectContentType === 'function') {
      contentType = detectContentType(actualContent)
    } else {
      // Simple content type detection as fallback
      if (actualContent.trim().startsWith('{') && actualContent.trim().endsWith('}')) {
        contentType = 'json'
      } else if (actualContent.includes('```') || actualContent.includes('function') || actualContent.includes('class')) {
        contentType = 'code'
      } else if (actualContent.includes('#') || actualContent.includes('**')) {
        contentType = 'markdown'
      } else if (actualContent.includes('<') && actualContent.includes('>')) {
        contentType = 'html'
      }
    }

    currentPasteContent.value = {
      content: actualContent,
      wordCount,
      charCount,
      type: contentType
    }

    showPasteModal.value = true

    // Prevent body scroll
    document.body.style.overflow = 'hidden'

    console.log('Paste modal opened successfully', { wordCount, charCount, type: contentType }) // Debug log

  } catch (error) {
    console.error('Error opening paste modal:', error)
    toast.error('Error opening preview', {
      duration: 3000,
      description: 'Could not display content preview'
    })
  }
}

function cleanupPastePreviewHandlers() {
  document.removeEventListener('click', handlePastePreviewClick, true)
  document.removeEventListener('click', handleRemovePastePreview, true)
  console.log('Paste preview handlers cleaned up') // Debug log
}


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

// Initialize Speech Recognition (add this in the onMounted hook)
function initializeSpeechRecognition() {
  const SpeechRecognition = (window as any).SpeechRecognition || (window as any).webkitSpeechRecognition
  if (!SpeechRecognition) {
    console.warn('Speech recognition not supported')
    return
  }

  const recognition = new SpeechRecognition()
  recognition.continuous = true
  recognition.interimResults = true
  recognition.lang = 'en-US'
  recognition.maxAlternatives = 1

  recognition.onresult = (event: any) => {
    let interimTranscript = ''
    for (let i = event.resultIndex; i < event.results.length; i++) {
      const transcript = event.results[i][0].transcript
      if (event.results[i].isFinal) {
        transcribedText.value += transcript + ' '
      } else {
        interimTranscript += transcript
      }
    }
    updateTextarea(interimTranscript)
  }

  recognition.onerror = (event: any) => {
    console.error('Speech recognition error:', event.error)
    isRecording.value = false
    isTranscribing.value = false
    microphonePermission.value = event.error === 'not-allowed' ? 'denied' : microphonePermission.value
    toast.error('Voice Input Error', {
      duration: 4000,
      description: event.error
    })
  }

  recognition.onend = () => {
    if (isRecording.value) {
      setTimeout(() => recognition.start(), 500)
    } else {
      isTranscribing.value = false
    }
  }

  recognition.onstart = () => {
    isTranscribing.value = true
  }

  voiceRecognition.value = recognition
}

function updateTextarea(interim: string) {
  if (updateTimeout) clearTimeout(updateTimeout)
  updateTimeout = window.setTimeout(() => {
    const textarea = document.getElementById('prompt') as HTMLTextAreaElement
    if (textarea) {
      // Only update if we're actually recording or have transcribed text
      if (isRecording.value || transcribedText.value) {
        textarea.value = transcribedText.value + interim
        autoGrow({ target: textarea } as any)
      }
    }
  }, 100)
}

// Toggle voice recording
async function toggleVoiceRecording() {
  if (!voiceRecognition.value) {
    toast.error('Speech recognition not available', {
      duration: 3000,
      description: 'Your browser may not support speech recognition.'
    })
    return
  }

  if (isRecording.value) {
    // Stop recording
    stopVoiceRecording()
  } else {
    // Start recording
    await startVoiceRecording()
  }
}

// Start voice recording
async function startVoiceRecording() {
  try {
    await navigator.mediaDevices.getUserMedia({ audio: true })
    microphonePermission.value = 'granted'
    isRecording.value = true
    transcribedText.value = ''
    transcriptionDuration.value = 0
    startTimer()

    const textarea = document.getElementById('prompt') as HTMLTextAreaElement
    if (textarea) {
      textarea.value = ''
      autoGrow({ target: textarea } as any)
    }

    voiceRecognition.value?.start()
  } catch (error) {
    microphonePermission.value = 'denied'
    isRecording.value = false
    toast.error('Microphone Access Denied', {
      duration: 5000,
      description: 'Please allow microphone access.'
    })
  }
}

// Stop voice recording
function stopVoiceRecording() {
  isRecording.value = false
  isTranscribing.value = false
  stopTimer()
  voiceRecognition.value?.stop()
}

function startTimer() {
  transcriptionTimer = window.setInterval(() => {
    transcriptionDuration.value += 1
  }, 1000)
}

function stopTimer() {
  if (transcriptionTimer) clearInterval(transcriptionTimer)
}

// Clear voice transcription
function clearVoiceTranscription() {
  transcribedText.value = ''
  const textarea = document.getElementById('prompt') as HTMLTextAreaElement
  if (textarea) {
    textarea.value = ''
    autoGrow({ target: textarea } as any)
    textarea.focus() // Refocus after clearing
  }
}

watch([isRecording, isTranscribing], ([recording, transcribing]) => {
  if (!recording && !transcribing && !transcribedText.value) {
    const textarea = document.getElementById('prompt') as HTMLTextAreaElement
    if (textarea && textarea.value && !pastePreview.value) {
      textarea.value = ''
      autoGrow({ target: textarea } as any)
    }
  }
})

watch(isAuthenticated, (newVal) => {
  if (newVal) {
    showCreateSession.value = false
  }
})

// watch for user plan changes
watch(() => ({
  isFree: isFreeUser.value,
  planName: parsedUserDetails.value?.plan_name,
  planStatus: planStatus.value.status
}), (newValue, oldValue) => {
  if (!oldValue) return // Skip initial call

  // If user upgraded from free to paid
  if (oldValue.isFree === true && newValue.isFree === false) {
    resetRequestCount()
    toast.success(`Welcome to ${newValue.planName || 'Premium'}!`, {
      duration: 5000,
      description: 'You now have unlimited requests!'
    })
  }
  // If user downgraded from paid to free (plan expired)
  else if (oldValue.isFree === false && newValue.isFree === true) {
    loadRequestCount()

    if (newValue.planStatus === 'expired') {
      toast.warning('Your plan has expired', {
        duration: Infinity,
        description: `You're now limited to ${FREE_REQUEST_LIMIT} requests per day`,
        action: {
          label: 'Upgrade',
          onClick: () => {
            router.push('/upgrade')
          }
        }
      })
    }
  }
}, { deep: true })

// Call loadRequestCount after user details are fully loaded
watch(() => parsedUserDetails.value, (newUserDetails) => {
  if (newUserDetails) {
    // Small delay to ensure all computed properties are updated
    nextTick(() => {
      setTimeout(() => {
        loadRequestCount()
      }, 100)
    })
  }
}, { immediate: true })

// planStatus to handle reactive objects properly
watch(() => ({ ...planStatus.value }), (newStatus, oldStatus) => {
  if (oldStatus && oldStatus.isExpired === false && newStatus.isExpired === true) {
    toast.error('Your plan has expired', {
      duration: Infinity,
      description: 'Please renew your plan to continue using the service.',
      action: {
        label: 'Renew Now',
        onClick: () => {
          router.push('/upgrade')
        }
      }
    })
  }
}, { deep: true })


onBeforeUnmount(() => {
  // Clean up speech recognition
  if (isRecording.value) {
    stopVoiceRecording()
  }

  // Remove keyboard listener
  document.removeEventListener('keydown', handleModalKeydown)

  // Clean up paste preview handlers (use the enhanced cleanup function)
  cleanupPastePreviewHandlers()

  // Restore body scroll if modal is open
  if (showPasteModal.value) {
    document.body.style.overflow = 'auto'
  }
})

// Consolidated onMounted hook for better organization
onMounted(() => {
  // 1. Load basic state
  const saved = localStorage.getItem("isCollapsed");
  if (saved && saved !== "null") {
    try {
      isCollapsed.value = JSON.parse(saved);
    } catch (err) {
      console.error("Error parsing isCollapsed:", err);
    }
  }

  const savedChatId = localStorage.getItem("currentChatId");
  if (savedChatId) currentChatId.value = savedChatId;

  // 2. Load cached previews
  loadLinkPreviewCache();

  // 3. Setup paste preview handlers
  setupPastePreviewHandlers();

  // 4. Make functions globally available with error boundaries
  if (typeof window !== 'undefined') {
    const safeGlobalFunction = (fn: Function, name: string) => {
      return (...args: any[]) => {
        try {
          return fn.apply(this, args);
        } catch (error) {
          console.error(`Error in ${name}:`, error);
          toast.error(`Error in ${name}`, {
            duration: 3000,
            description: 'An unexpected error occurred'
          });
        }
      };
    };

    (window as any).openPasteModal = safeGlobalFunction(openPasteModal, 'openPasteModal');
    (window as any).copyPasteContent = safeGlobalFunction(copyPasteContent, 'copyPasteContent');
    (window as any).removePastePreview = safeGlobalFunction(removePastePreview, 'removePastePreview');

    // Video functions
    (window as any).playEmbeddedVideo = safeGlobalFunction(playEmbeddedVideo, 'playEmbeddedVideo');
    (window as any).pauseVideo = safeGlobalFunction(pauseVideo, 'pauseVideo');
    (window as any).resumeVideo = safeGlobalFunction(resumeVideo, 'resumeVideo');
    (window as any).stopVideo = safeGlobalFunction(stopVideo, 'stopVideo');
    (window as any).toggleDirectVideo = safeGlobalFunction(toggleDirectVideo, 'toggleDirectVideo');
    (window as any).stopDirectVideo = safeGlobalFunction(stopDirectVideo, 'stopDirectVideo');
    (window as any).showVideoControls = safeGlobalFunction(showVideoControls, 'showVideoControls');
    (window as any).updateVideoControls = safeGlobalFunction(updateVideoControls, 'updateVideoControls');
    (window as any).playSocialVideo = safeGlobalFunction(playSocialVideo, 'playSocialVideo');
  }

  // 5. Initialize features based on authentication
  if (isAuthenticated.value) {
    // Load request count for limited users
    const shouldHaveLimit = isFreeUser.value ||
      planStatus.value.isExpired ||
      planStatus.value.status === 'no-plan' ||
      planStatus.value.status === 'expired';

    if (shouldHaveLimit) {
      loadRequestCount();
    }

    loadChats();
    setupAutoSync();

    // Initial sync from server (delayed to avoid conflicts)
    setTimeout(() => {
      syncFromServer();
    }, 1000);

    // Pre-process links in existing messages
    currentMessages.value.forEach((item, index) => {
      [item.prompt, item.response].forEach((text) => {
        if (text && text !== "...") {
          extractUrls(text)
            .slice(0, 3)
            .forEach((url) => {
              if (!linkPreviewCache.value.has(url)) {
                fetchLinkPreview(url).then(() => {
                  linkPreviewCache.value = new Map(linkPreviewCache.value);
                });
              }
            });
        }
      });
    });
  }

  // 6. Initialize speech recognition
  initializeSpeechRecognition();

  // 7. Global event listeners
  const copyListener = (e: any) => {
    if (e.target?.classList.contains("copy-button")) {
      const code = decodeURIComponent(e.target.getAttribute("data-code"));
      copyCode(code, e.target);
    }
  };
  document.addEventListener("click", copyListener);
  document.addEventListener('keydown', handleModalKeydown);

  // 8. Initialize video lazy loading
  initializeVideoLazyLoading();

  // 9. Setup periodic tasks
  const interval = setInterval(() => {
    now.value = Date.now();
  }, 1000);

  // Check for daily reset on app start
  checkAndResetDailyCount();
  const resetCheckInterval = setInterval(checkAndResetDailyCount, 5 * 60 * 1000);

  // 10. Setup DOM-dependent functionality
  nextTick(() => {
    // Clear any initial content in textarea
    const textarea = document.getElementById('prompt') as HTMLTextAreaElement;
    if (textarea && textarea.value) {
      console.log('Initial textarea content found:', textarea.value);
      textarea.value = '';
      transcribedText.value = '';
    }

    // Set up scroll listener with proper cleanup
    if (scrollableElem.value) {
      scrollableElem.value.addEventListener("scroll", debouncedHandleScroll, { passive: true });
    }

    // Auto-focus input
    if (showInput.value || currentMessages.value.length > 0) {
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

  // 11. Store cleanup functions for onBeforeUnmount
  onBeforeUnmount(() => {
    // Clean up event listeners
    document.removeEventListener("click", copyListener);
    document.removeEventListener('keydown', handleModalKeydown);
    document.removeEventListener('click', handlePastePreviewClick);
    document.removeEventListener('click', handleRemovePastePreview);

    // Clean up scroll listener
    if (scrollableElem.value) {
      scrollableElem.value.removeEventListener("scroll", debouncedHandleScroll);
    }

    // Clean up video lazy loading observer
    destroyVideoLazyLoading();

    // Clean up intervals
    clearInterval(interval);
    clearInterval(resetCheckInterval);

    if (autoSyncInterval) {
      clearInterval(autoSyncInterval);
    }

    // Clear timeouts
    if (scrollTimeout) {
      clearTimeout(scrollTimeout);
    }
    if (resizeTimeout) {
      clearTimeout(resizeTimeout);
    }
    if (transcriptionTimer) {
      clearInterval(transcriptionTimer);
    }
    if (updateTimeout) {
      clearTimeout(updateTimeout);
    }

    // Clean up speech recognition
    if (isRecording.value) {
      stopVoiceRecording();
    }

    // Restore body scroll if modal is open
    if (showPasteModal.value) {
      document.body.style.overflow = 'auto';
    }

    // Final sync if needed
    if (syncStatus.value.hasUnsyncedChanges) {
      syncToServer();
    }
  });
});
</script>

<template>
  <div class="flex h-[100vh] bg-background text-[var(--foreground)]">
    <!-- Sidebar -->
    <SideNav v-if="isAuthenticated" :data="{
      chats,
      currentChatId,
      parsedUserDetails,
      screenWidth,
      isCollapsed,
      syncStatus,
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
      :class="screenWidth > 720 && isAuthenticated ? (!isCollapsed ?
        'flex-grow flex flex-col items-center justify-center ml-[270px] font-light text-sm transition-all duration-300 ease-in-out bg-inherit'
        :
        'flex-grow flex flex-col items-center justify-center ml-[60px] font-light text-sm transition-all duration-300 ease-in-out bg-inherit'
      )
        : 'text-sm font-light flex-grow items-center justify-center flex flex-col transition-all duration-300 ease-in-out bg-inherit'">

      <div
        :class="(screenWidth > 720 && isAuthenticated) ? 'h-screen bg-inherit flex flex-col items-center justify-center w-[85%]' : 'bg-inherit h-screen w-full flex flex-col items-center justify-center'">
        <TopNav v-if="isAuthenticated" :data="{
          currentChat,
          parsedUserDetails,
          screenWidth,
          isCollapsed,
          isSidebarHidden,
          syncStatus,
          }" 
          :functions="{
            hideSidebar,
            manualSync,
          }" 
        />
        <!-- Empty State -->
        <CreateSessView v-if="!isAuthenticated" :chats="chats" :current-chat-id="currentChatId"
          :is-collapsed="isCollapsed" :parsed-user-details="parsedUserDetails" :screen-width="screenWidth"
          :sync-status="syncStatus" :is-loading="isLoading" :auth-step="authStep"
          :show-create-session="showCreateSession" :auth-data="authData" :current-messages="currentMessages"
          :validate-current-step="validateCurrentStep()" :set-show-input="setShowInput" :hide-sidebar="hideSidebar"
          :clear-all-chats="clearAllChats" :toggle-sidebar="toggleSidebar" :logout="logout"
          :create-new-chat="createNewChat" :switch-to-chat="switchToChat" :delete-chat="deleteChat"
          :rename-chat="renameChat" :manual-sync="manualSync" :handle-step-submit="handleStepSubmit"
          :prev-auth-step="prevAuthStep" :update-auth-data="updateAuthData"
          :set-show-create-session="setShowCreateSession" />

        <div v-else-if="isAuthenticated && currentMessages.length === 0">
          <!-- Loading Overlay -->
          <div v-if="isLoading"
            class="absolute top-0 left-0 w-full h-full bg-white bg-opacity-80 z-10 flex flex-col items-center justify-center">
            <i class="pi pi-spin pi-spinner text-4xl text-gray-500 mb-4"></i>
            <p class="text-gray-700">Loading...</p>
          </div>

          <div v-else class="flex flex-col md:flex-grow items-center gap-3 text-gray-600">
            <div class="rounded-full bg-gray-200 w-[60px] h-[60px] flex justify-center items-center">
              <span class="pi pi-comment text-lg"></span>
            </div>

            <p class="text-3xl text-black font-semibold">{{ parsedUserDetails?.username || 'Gemmie' }}</p>
            <div class="text-center max-w-md space-y-2">
              <p class="text-gray-600 leading-relaxed">
                Experience privacy-first conversations with advanced AI. Your data stays secure, local and synced to
                your
                all devices.
              </p>
              <div class="flex items-center justify-center gap-6 text-sm text-gray-500 mt-4">
                <div class="flex items-center gap-1">
                  <i class="pi pi-shield text-green-500"></i>
                  <span>Private</span>
                </div>
                <div class="flex items-center gap-1">
                  <i class="pi pi-database text-blue-500"></i>
                  <span>Local Stored</span>
                </div>
                <div class="flex items-center gap-1">
                  <i class="pi pi-sync text-purple-500"></i>
                  <span>Synced</span>
                </div>
              </div>
            </div>
            <div class="flex flex-col gap-3 w-full max-w-xs">
              <button v-if="!showInput" @click="setShowInput"
                class="group px-6 py-3 bg-gradient-to-r from-blue-500 to-purple-600 text-white rounded-lg hover:from-blue-600 hover:to-purple-700 transition-all duration-300 transform hover:scale-105 shadow-lg hover:shadow-xl font-medium">
                <span class="flex items-center justify-center gap-2">
                  <i class="pi pi-pencil group-hover:rotate-12 transition-transform"></i>
                  Start Writing
                </span>
              </button>
            </div>
          </div>
        </div>

        <!-- Update your chat messages container -->
        <div ref="scrollableElem" v-else-if="currentMessages.length !== 0 && isAuthenticated"
          class="flex-grow no-scrollbar overflow-y-auto px-2 sm:px-4 w-full space-y-3 sm:space-y-4 mt-[55px]" :class="isRequestLimitExceeded || shouldShowUpgradePrompt ? 'pb-[160px] sm:pb-[150px]' :
            showScrollDownButton ? 'pb-[140px] sm:pb-[120px]' :
              'pb-[110px] sm:pb-[120px]'">
          <div v-if="currentMessages.length !== 0" v-for="(item, i) in currentMessages" :key="`chat-${i}`"
            class="flex flex-col gap-2">
            <!-- User Bubble -->
            <div class="flex w-full justify-end chat-message">
              <div class="flex flex-col items-end max-w-[85%] sm:max-w-[75%] md:max-w-[70%]">
                <!-- User message bubble -->
                <div class="bg-gray-50 text-black p-3 rounded-2xl prose prose-sm max-w-none chat-bubble w-fit">
                  <p class="text-xs opacity-80 text-right mb-1">{{ parsedUserDetails?.username || "You" }}</p>
                  <div class="break-words"
                    v-html="renderMarkdown((item && item?.prompt && item?.prompt?.length > 800) ? item?.prompt?.trim().split('#pastedText#')[0] : item.prompt || '')">
                  </div>
                </div>
                <div class="flex flex-col gap-3 mt-2">
                  <!-- Link Previews for User Messages -->
                  <div v-if="extractUrls(item.prompt || '').length > 0" class="w-full">
                    <div v-for="url in extractUrls(item.prompt || '').slice(0, 3)" :key="`user-${i}-${url}`">
                      <div v-if="linkPreviewCache.get(url)"
                        v-html="LinkPreviewComponent({ preview: linkPreviewCache.get(url)! })"></div>
                    </div>
                  </div>

                  <div
                    v-if="item && item.prompt && (item?.prompt?.trim().split(/\s+/).length > 100 || item?.prompt?.length > 800)"
                    class="mb-3">
                    <div class="flex justify-center">
                      <div class="ml-auto sm:w-[80%] md:w-[70%] lg:w-[60%] xl:w-[50%]" v-html="PastePreviewComponent(
                        item?.prompt?.trim()?.split('#pastedText#')[1] || '',
                        item?.prompt?.trim().split('#pastedText#')[1]?.split(/\s+/)?.length || 0,
                        item?.prompt?.trim()?.split('#pastedText#')[1]?.length || 0, true
                      )">
                      </div>
                    </div>
                  </div>

                </div>

              </div>
            </div>

            <!-- Bot Bubble -->
            <div class="flex w-full justify-start relative">
              <div
                class="bg-none chat-message leading-relaxed text-black p-3 rounded-2xl prose prose-sm w-fit max-w-[95%] sm:max-w-[85%] md:max-w-[80%]">

                <!-- Loading state -->
                <div v-if="item.response === '...'" class="flex w-full items-center gap-2 text-gray-500">
                  <i class="pi pi-spin pi-spinner"></i>
                  <span class="text-sm">Thinking...</span>
                </div>
                <div v-else-if="item.response === 'refreshing...'" class="flex w-full items-center gap-2 text-gray-500">
                  <i class="pi pi-spin pi-spinner"></i>
                  <span class="text-sm">Refreshing...</span>
                </div>

                <!-- Regular response with enhanced link handling -->
                <div v-else>
                  <div v-html="renderMarkdown(item.response || '')"></div>

                  <!-- Link Previews Section -->
                  <div v-if="extractUrls(item.response || '').length > 0" class="mt-2 sm:mt-3">
                    <div v-for="url in extractUrls(item.response || '').slice(0, 3)" :key="url">
                      <div v-if="linkPreviewCache.get(url)"
                        v-html="LinkPreviewComponent({ preview: linkPreviewCache.get(url)! })"></div>
                    </div>
                  </div>
                </div>

                <!-- Actions - Responsive with fewer labels on mobile -->
                <div v-if="item.response !== '...' && item.response !== 'refreshing...'"
                  class="flex flex-wrap gap-2 sm:gap-3 mt-2 text-gray-500 text-sm">
                  <button @click="copyResponse(item.response, i)"
                    class="flex items-center gap-1 hover:text-blue-600 transition-colors min-h-[32px]">
                    <i class="pi pi-copy"></i>
                    <span>{{ copiedIndex === i ? 'Copied!' : 'Copy' }}</span>
                  </button>

                  <button @click="shareResponse(item.response, item.prompt)"
                    class="flex items-center gap-1 hover:text-green-600 transition-colors min-h-[32px]">
                    <i class="pi pi-share-alt"></i>
                    <span>Share</span>
                  </button>

                  <button @click="refreshResponse(item.prompt)" :disabled="isLoading"
                    class="flex items-center gap-1 hover:text-orange-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed min-h-[32px]">
                    <i class="pi pi-refresh"></i>
                    <span>Refresh</span>
                  </button>

                  <button @click="deleteMessage(i)" :disabled="isLoading"
                    class="flex items-center gap-1 hover:text-red-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed min-h-[32px]">
                    <i class="pi pi-trash"></i>
                    <span>Delete</span>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Responsive Scroll to Bottom Button -->
        <button v-if="showScrollDownButton && currentMessages.length !== 0 && isAuthenticated" @click="scrollToBottom"
          :class="isRequestLimitExceeded || shouldShowUpgradePrompt ? 'bottom-[180px]' : 'bottom-[90px]'"
          class="fixed bg-gray-50 text-gray-500 border px-5 h-[34px] rounded-full shadow-lg hover:bg-gray-100 transition-colors z-20 left-1/2 transform -translate-x-1/2">
          <div class="flex gap-1 sm:gap-2 items-center justify-center w-full font-semibold h-full">
            <i class="pi pi-arrow-down text-center"></i>
            <p class="whitespace-nowrap">Scroll Down</p>
          </div>
        </button>


        <!-- Input Area -->
        <div v-if="(currentMessages.length !== 0 || showInput === true) && isAuthenticated" :style="screenWidth > 720 && !isCollapsed ? 'left:270px;' :
          screenWidth > 720 && isCollapsed ? 'left:60px;' : 'left:0px;'"
          class="bg-white z-20 bottom-0 right-0 fixed pb-3 sm:pb-5 px-2 sm:px-5">

          <div class="flex items-center justify-center w-full">
            <form @submit="handleSubmit" :class="screenWidth > 720 ? isCollapsed ?
              'relative flex bg-gray-50 flex-col border-2 shadow rounded-2xl items-center w-[85%] max-w-6xl' :
              'relative flex bg-gray-50 flex-col border-2 shadow rounded-2xl items-center w-[85%] max-w-4xl' :
              'relative flex flex-col border-2 bg-gray-50 shadow rounded-2xl w-full max-w-full items-center'">

              <!-- Paste Preview inside form - above other content -->
              <div v-if="pastePreview && pastePreview.show" class="w-full p-3 border-b">
                <div
                  v-html="PastePreviewComponent(pastePreview.content, pastePreview.wordCount, pastePreview.charCount)">
                </div>
              </div>

              <!-- Request Limit Exceeded Banner -->
              <div v-if="showLimitExceededBanner" class="py-2 sm:py-3 w-full px-2 sm:px-3">
                <div class="flex items-center justify-center w-full">
                  <!-- Mobile: Stacked Layout -->
                  <div class="flex sm:hidden w-full flex-col gap-2">
                    <div class="flex items-center gap-2">
                      <div
                        class="w-6 h-6 sm:w-8 sm:h-8 bg-red-100 rounded-full flex items-center justify-center flex-shrink-0">
                        <i class="pi pi-ban text-red-600 text-xs sm:text-sm"></i>
                      </div>
                      <div class="min-w-0 flex-1">
                        <h3 class="text-xs sm:text-sm font-semibold text-red-800 leading-tight">
                          {{ planStatus.isExpired ? 'Plan Expired' : 'Free Requests Exhausted' }}
                        </h3>
                        <p class="text-xs text-red-600 leading-tight mt-0.5">
                          {{ planStatus.isExpired ? 'Renew your plan' : `Used all ${FREE_REQUEST_LIMIT} requests` }}
                        </p>
                      </div>
                    </div>
                    <button @click="$router.push('/upgrade')"
                      class="w-full bg-red-500 hover:bg-red-600 text-white py-2 rounded-md text-xs font-medium transition-colors">
                      {{ planStatus.isExpired ? 'Renew Plan' : 'Upgrade Plan' }}
                    </button>
                  </div>

                  <!-- Desktop: Horizontal Layout -->
                  <div class="hidden sm:flex w-full items-center gap-3">
                    <div class="w-8 h-8 bg-red-100 rounded-full flex items-center justify-center flex-shrink-0">
                      <i class="pi pi-ban text-red-600 text-sm"></i>
                    </div>
                    <div class="min-w-0 flex-1">
                      <h3 class="text-sm font-semibold text-red-800 mb-1">
                        {{ planStatus.isExpired ? 'Plan Expired' : 'Free Requests Exhausted' }}
                      </h3>
                      <p class="text-xs text-red-600">
                        {{ planStatus.isExpired ?
                          'Please renew your plan to continue using the service.' :
                          `You've used all ${FREE_REQUEST_LIMIT} free requests. Upgrade to continue chatting.` }}
                      </p>
                    </div>
                    <button @click="$router.push('/upgrade')"
                      class="bg-red-500 px-3 hover:bg-red-600 text-white py-2 rounded-md text-sm font-medium transition-colors flex-shrink-0 whitespace-nowrap">
                      {{ planStatus.isExpired ? 'Renew Plan' : 'Upgrade Plan' }}
                    </button>
                  </div>
                </div>
              </div>

              <!-- Upgrade Warning Banner -->
              <div v-if="showUpgradeBanner" class="py-2 sm:py-3 w-full px-2 sm:px-3">
                <div class="flex items-center justify-center w-full">
                  <!-- Mobile: Stacked Layout -->
                  <div class="flex sm:hidden w-full flex-col gap-2">
                    <div class="flex items-center gap-2">
                      <div
                        class="w-6 h-6 sm:w-8 sm:h-8 bg-yellow-100 rounded-full flex items-center justify-center flex-shrink-0">
                        <i class="pi pi-exclamation-triangle text-yellow-600 text-xs sm:text-sm"></i>
                      </div>
                      <div class="min-w-0 flex-1">
                        <h3 class="text-xs sm:text-sm font-semibold text-yellow-800 leading-tight">
                          {{ requestsRemaining }} requests left
                        </h3>
                        <p class="text-xs text-yellow-600 leading-tight mt-0.5">
                          Upgrade for unlimited access
                        </p>
                      </div>
                    </div>
                    <button @click="$router.push('/upgrade')"
                      class="w-full bg-orange-500 hover:bg-orange-600 text-white py-2 rounded-md text-xs font-medium transition-colors">
                      Upgrade Plan
                    </button>
                  </div>

                  <!-- Desktop: Horizontal Layout -->
                  <div class="hidden sm:flex w-full items-center gap-3">
                    <div class="w-8 h-8 bg-yellow-100 rounded-full flex items-center justify-center flex-shrink-0">
                      <i class="pi pi-exclamation-triangle text-yellow-600 text-sm"></i>
                    </div>
                    <div class="min-w-0 flex-1">
                      <h3 class="text-sm font-semibold text-yellow-800 mb-1">
                        {{ requestsRemaining }} free requests remaining
                      </h3>
                      <p class="text-xs text-yellow-600">
                        Upgrade to continue chatting without limits
                      </p>
                    </div>
                    <button @click="$router.push('/upgrade')"
                      class="bg-orange-500 px-3 hover:bg-orange-600 text-white py-2 rounded-md text-sm font-medium transition-colors flex-shrink-0 whitespace-nowrap">
                      Upgrade Plan
                    </button>
                  </div>
                </div>
              </div>

              <!-- Input Area with Voice Recording -->
              <div class="flex w-full bg-white rounded-2xl px-2 sm:px-3 py-1 sm:py-2 items-center gap-2 sm:gap-3"
                :class="inputDisabled ? 'opacity-50 border border-t pointer-events-none' :
                  showUpgradeBanner ? 'border border-t' : ''">

                <!-- Voice Recording Indicator (when active) -->
                <div v-if="isRecording || isTranscribing"
                  class="flex items-center gap-1 sm:gap-2 px-2 py-1 bg-red-50 rounded-lg border border-red-200 text-red-600 text-xs sm:text-sm flex-shrink-0">
                  <div class="flex items-center gap-1">
                    <div class="w-2 h-2 bg-red-500 rounded-full animate-pulse"></div>
                    <span class="hidden sm:inline">{{ isTranscribing ? 'Listening...' : 'Starting...' }}</span>
                    <span class="sm:hidden">{{ isTranscribing ? '🎤' : '⏳' }}</span>
                  </div>
                </div>

                <!-- Microphone Toggle Button -->
                <button type="button" @click="toggleVoiceRecording" :disabled="inputDisabled" :class="[
                  'rounded-lg w-8 h-8 sm:w-9 sm:h-9 flex items-center justify-center transition-all duration-200 flex-shrink-0',
                  isRecording
                    ? 'bg-red-500 hover:bg-red-600 text-white shadow-lg transform scale-105 animate-pulse'
                    : 'bg-gray-100 hover:bg-gray-200 text-gray-600 hover:text-gray-700',
                  'disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none'
                ]" :title="microphonePermission === 'denied'
                  ? 'Microphone access denied'
                  : isRecording
                    ? 'Stop voice input'
                    : 'Start voice input'" :aria-label="isRecording ? 'Stop voice input' : 'Start voice input'">

                  <!-- Microphone Icon -->
                  <svg v-if="microphonePermission === 'prompt'" class="w-4 h-4 sm:w-5 sm:h-5" fill="currentColor"
                    viewBox="0 0 24 24">
                    <path d="M12 2a3 3 0 0 0-3 3v6a3 3 0 0 0 6 0V5a3 3 0 0 0-3-3Z" />
                    <path d="M19 10v1a7 7 0 0 1-14 0v-1a1 1 0 0 1 2 0v1a5 5 0 0 0 10 0v-1a1 1 0 0 1 2 0Z" />
                    <path d="M12 18.5a1 1 0 0 1 1 1v1a1 1 0 0 1-2 0v-1a1 1 0 0 1 1-1Z" />
                  </svg>

                  <svg v-if="!isRecording && microphonePermission === 'granted'" class="w-4 h-4 sm:w-5 sm:h-5"
                    fill="currentColor" viewBox="0 0 24 24">
                    <path d="M12 2a3 3 0 0 0-3 3v6a3 3 0 0 0 6 0V5a3 3 0 0 0-3-3Z" />
                    <path d="M19 10v1a7 7 0 0 1-14 0v-1a1 1 0 0 1 2 0v1a5 5 0 0 0 10 0v-1a1 1 0 0 1 2 0Z" />
                    <path d="M12 18.5a1 1 0 0 1 1 1v1a1 1 0 0 1-2 0v-1a1 1 0 0 1 1-1Z" />
                  </svg>

                  <!-- Stop Icon -->
                  <svg v-else-if="microphonePermission === 'granted' && isRecording" class="w-4 h-4 sm:w-5 sm:h-5"
                    fill="currentColor" viewBox="0 0 24 24">
                    <rect x="6" y="6" width="12" height="12" rx="2" />
                  </svg>

                  <!-- Microphone Denied Icon -->
                  <svg v-else-if="microphonePermission === 'denied' && !isRecording"
                    class="w-4 h-4 sm:w-5 sm:h-5 text-red-500" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M12 2a3 3 0 0 0-3 3v6a3 3 0 0 0 6 0V5a3 3 0 0 0-3-3Z" />
                    <path d="M19 10v1a7 7 0 0 1-14 0v-1a1 1 0 0 1 2 0v1a5 5 0 0 0 10 0v-1a1 1 0 0 1 2 0Z" />
                    <path d="M12 18.5a1 1 0 0 1 1 1v1a1 1 0 0 1-2 0v-1a1 1 0 0 1 1-1Z" />
                    <line x1="4" y1="4" x2="20" y2="20" stroke="currentColor" stroke-width="2" />
                  </svg>
                </button>

                <!-- Clear Voice Button (when transcribed text exists) -->
                <button v-if="transcribedText && !isRecording" type="button" @click="clearVoiceTranscription"
                  class="rounded-lg w-6 h-6 sm:w-7 sm:h-7 flex items-center justify-center transition-colors text-gray-400 hover:text-gray-600 hover:bg-gray-50 flex-shrink-0"
                  title="Clear voice transcription">
                  <svg class="w-3 h-3 sm:w-4 sm:h-4" fill="currentColor" viewBox="0 0 24 24">
                    <path
                      d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z" />
                  </svg>
                </button>

                <!-- Textarea -->
                <textarea required id="prompt" name="prompt" @keydown="onEnter" @input="autoGrow" @paste="handlePaste"
                  :disabled="inputDisabled" rows="1" :class="[
                    'flex-grow py-3 px-1 placeholder:text-gray-500 rounded-t-2xl bg-white text-sm outline-none resize-none border-none max-h-[120px] sm:max-h-[150px] md:max-h-[200px] overflow-auto leading-relaxed w-full min-w-0',
                    'disabled:opacity-50 disabled:cursor-not-allowed',
                    isRecording ? 'bg-red-50 border border-red-100' : ''
                  ]" :placeholder="inputPlaceholderText">
                </textarea>

                <!-- Submit Button -->
                <button type="submit" :disabled="inputDisabled"
                  class="rounded-lg w-6 h-6 sm:w-7 sm:h-7 flex items-center justify-center transition-colors text-white bg-blue-500 hover:bg-blue-600 disabled:cursor-not-allowed disabled:opacity-50 disabled:bg-gray-400 flex-shrink-0">
                  <i v-if="!isLoading" class="pi pi-arrow-up text-xs sm:text-sm"></i>
                  <i v-else class="pi pi-spin pi-spinner text-xs sm:text-sm"></i>
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>
    </div>
    <PastePreviewModal :data="{
      showPasteModal,
      currentPasteContent,
    }" :closePasteModal="closePasteModal" />
  </div>
</template>
