<script lang="ts" setup>
import type { Ref } from "vue"
import { ref, onMounted, nextTick, computed, onBeforeUnmount, inject, watch, onUnmounted } from "vue"
import SideNav from "../components/SideNav.vue"
import TopNav from "../components/TopNav.vue"
import type { Chat, ConfirmDialogOptions, CurrentChat, LinkPreview, Res, UserDetails } from "@/types"
import { toast } from 'vue-sonner'
import { destroyVideoLazyLoading, initializeVideoLazyLoading, observeNewVideoContainers, pauseVideo, playEmbeddedVideo, playSocialVideo, resumeVideo, showVideoControls, stopDirectVideo, stopVideo, toggleDirectVideo, updateVideoControls } from "@/utils/videoProcessing"
import { onUpdated } from "vue"
import { extractUrls, generateChatTitle, copyCode, isPromptTooShort, WRAPPER_URL, detectLargePaste, SPINDLE_URL } from "@/utils/globals"
import CreateSessView from "./CreateSessView.vue"
import router from "@/router"
import { copyPasteContent, detectContentType } from "@/utils/previewPasteContent"
import PastePreviewModal from "@/components/Modals/PastePreviewModal.vue"
import { useRoute } from "vue-router"
import { renderMarkdown } from "@/utils/markdownSupport"

type ModeOption = {
  mode: 'light-response' | 'web-search' | 'deep-search',
  label: string,
  description: string,
  icon: string,
  title: string
}

// Inject global state
const globalState = inject('globalState') as {
  handleAuth: (data: {
    username: string;
    email: string;
    password: string;
  }) => any,
  chatDrafts: Ref<Map<string, string>>,
  userDetailsDebounceTimer: any,
  chatsDebounceTimer: any,
  screenWidth: Ref<number>,
  confirmDialog: Ref<ConfirmDialogOptions>,
  isCollapsed: Ref<boolean>,
  authData: Ref<{ username: string; email: string; password: string; agreeToTerms: boolean; }>,
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
  isAuthenticated: Ref<boolean>,
  parsedUserDetails: Ref<UserDetails>,
  planStatus: Ref<{ status: string; timeLeft: string; expiryDate: string; isExpired: boolean; }>,
  currentChatId: Ref<string>,
  pastePreviews: Ref<Map<string, {
    content: string;
    wordCount: number;
    charCount: number;
    show: boolean;
  }>>,
  chats: Ref<Chat[]>
  logout: () => void,
  isLoading: Ref<boolean>,
  expanded: Ref<boolean[]>,
  showInput: Ref<boolean>,
  showConfirmDialog: (options: ConfirmDialogOptions) => void,
  setShowInput: () => void,
  clearAllChats: () => void,
  switchToChat: (chatId: string) => void,
  createNewChat: (initialMessage?: string) => string,
  deleteChat: (chatId: string) => void,
  loadChatDrafts: () => void,
  saveChatDrafts: () => void,
  renameChat: (chatId: string, newTitle: string) => void,
  deleteMessage: (messageIndex: number) => void,
  scrollableElem: Ref<HTMLElement | null>,
  showScrollDownButton: Ref<boolean>,
  handleScroll: () => void,
  scrollToBottom: () => void,
  cancelAllRequests: () => void,
  cancelChatRequests: (chatId: string) => void,
  saveChats: () => void,
  clearCurrentDraft: () => void,
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
  isFreeUser: Ref<boolean>,
  FREE_REQUEST_LIMIT: number,
  isDarkMode: Ref<boolean>,
  hasActiveRequestsForCurrentChat: Ref<boolean>,
  isUserOnline: Ref<boolean>,
  connectionStatus: Ref<string>,
  checkInternetConnection: () => Promise<boolean>,
  autoGrow: (e: Event) => void,
  showSyncIndicator: (message: string, progress?: number) => void,
  hideSyncIndicator: () => void,
  updateSyncProgress: (message: string, progress: number) => void,
  performSmartSync: () => Promise<void>,
  activeRequests: Ref<Map<string, AbortController>>,
  requestChatMap: Ref<Map<string, string>>,
  pendingResponses: Ref<Map<string, { prompt: string; chatId: string }>>,
  requestCount: Ref<number>,
  resetRequestCount: () => void,
  incrementRequestCount: () => void,
  loadRequestCount: () => void,
  checkRequestLimitBeforeSubmit: () => boolean,
  requestsRemaining: Ref<boolean>,
  shouldShowUpgradePrompt: Ref<boolean>,
  isRequestLimitExceeded: Ref<boolean>,
}

const {
  chatDrafts,
  screenWidth,
  confirmDialog,
  isCollapsed,
  authData,
  syncStatus,
  isAuthenticated,
  currentChatId,
  pastePreviews,
  chats,
  planStatus,
  userDetailsDebounceTimer,
  chatsDebounceTimer,
  logout,
  isLoading,
  expanded,
  showInput,
  hasActiveRequestsForCurrentChat,
  showConfirmDialog,
  cancelAllRequests,
  cancelChatRequests,
  checkRequestLimitBeforeSubmit,
  setShowInput,
  clearAllChats,
  switchToChat,
  createNewChat,
  deleteChat,
  loadChatDrafts,
  saveChatDrafts,
  clearCurrentDraft,
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
  isDarkMode,
  updateExpandedArray,
  apiCall,
  manualSync,
  toggleSidebar,
  autoGrow,
  handleAuth,
  isFreeUser,
  isUserOnline,
  connectionStatus,
  checkInternetConnection,
  activeRequests,
  requestChatMap,
  pendingResponses,
  performSmartSync,
  requestCount,
  resetRequestCount,
  incrementRequestCount,
  loadRequestCount,
  FREE_REQUEST_LIMIT,
  requestsRemaining,
  shouldShowUpgradePrompt,
  isRequestLimitExceeded,
  parsedUserDetails,
} = globalState

const route = useRoute()
// ---------- State ----------
const authStep = ref(1)
const showCreateSession = ref(false)
const copiedIndex = ref<number | null>(null)
const now = ref(Date.now())
const showInputModeDropdown = ref(false)

const isRecording = ref(false)
const isTranscribing = ref(false)
const transcribedText = ref('')
const voiceRecognition = ref<any | null>(null)
const microphonePermission = ref<'granted' | 'denied' | 'prompt'>('prompt')
const transcriptionDuration = ref(0)
let transcriptionTimer: number | null = null
let updateTimeout: number | null = null

const showSuggestionsDropup = ref(false)

const showPasteModal = ref(false)
const pastePreview = computed(() => {
  return pastePreviews.value.get(currentChatId.value) || null
})
const currentPasteContent = ref<{
  content: string;
  wordCount: number;
  charCount: number;
  type: 'text' | 'code' | 'json' | 'markdown' | 'xml' | 'html';
} | null>(null)

const isLoadingState = (response: string): boolean => {
  return response.endsWith('...') || 
         response === '...' ||
         response.includes('web-search...') || response.includes('light-search...') ||
         response.includes('deep-search...') ||
         response.includes('light-response...') ||
         response === 'refreshing...'
}

const getLoadingMessage = (response: string): string => {
  if (response === 'web-search...' || response === 'light-search...') return 'Web searching...'
  if (response === 'deep-search...') return 'Deep searching...'
  if (response === 'light-response...') return 'Thinking...'
  if (response === 'refreshing...') return 'Refreshing...'
  return 'Thinking...'
}

const suggestionPrompts = [
  {
    icon: 'pi pi-pencil',
    title: 'Write',
    prompt: 'Write a short story about a time traveler who accidentally changes history',
  },
  {
    icon: 'pi pi-code',
    title: 'Code',
    prompt: 'Help me debug a JavaScript function that\'s not working as expected',
  },
  {
    icon: 'pi pi-book',
    title: 'Learn',
    prompt: 'Explain quantum computing in simple terms',
  },
  {
    icon: 'pi pi-heart',
    title: 'Health',
    prompt: 'Get me daily healthy routines',
  },
  {
    icon: 'pi pi-globe',
    title: 'Events',
    prompt: 'What are the latest global events?',
  }
]

// Handle suggestion selection
function selectSuggestion(prompt: string) {
  showSuggestionsDropup.value = false
  setShowInput()

  nextTick(() => {
    const textarea = document.getElementById('prompt') as HTMLTextAreaElement
    if (textarea) {
      textarea.value = prompt
      autoGrow({ target: textarea } as any)
      textarea.focus()
    }
  })
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
      <div class="link-preview loading border border-gray-200 dark:border-gray-700 rounded-lg p-3 my-2 bg-gray-50 dark:bg-gray-800 max-w-full transition-colors duration-200">
        <div class="flex items-center gap-2">
          <i class="pi pi-spin pi-spinner text-gray-400 dark:text-gray-500 flex-shrink-0"></i>
          <span class="text-sm text-gray-500 dark:text-gray-400 truncate">Loading preview...</span>
        </div>
      </div>
    `
  }

  if (preview.error) {
    return `
      <div class="link-preview error border border-gray-200 dark:border-gray-700 rounded-lg p-3 my-2 bg-gray-50 dark:bg-gray-800 max-w-full transition-colors duration-200">
        <div class="flex items-center gap-2 min-w-0">
          <i class="pi pi-external-link text-gray-400 dark:text-gray-500 flex-shrink-0"></i>
          <a href="${preview.url}" target="_blank" rel="noopener noreferrer" 
             class="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 text-sm font-medium truncate min-w-0 flex-1 transition-colors duration-200">
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
        <div class="aspect-video w-full bg-black dark:bg-gray-900 relative group overflow-hidden" id="${videoId}">
          <div class="video-embed-container object-cover w-full h-full" 
               data-embed='${preview.embedHtml.replace(/'/g, '&apos;')}'
               data-video-type="${preview.videoType}"
               data-video-id="${videoId}">
            
            <!-- Initial thumbnail state -->
            <div class="video-thumbnail w-full h-full bg-gray-900 dark:bg-gray-800 flex items-center justify-center cursor-pointer overflow-hidden"
                 onclick="playEmbeddedVideo(this, '${videoId}')">
              ${preview.videoThumbnail || preview.previewImage ? `
                <img src="${preview.videoThumbnail || preview.previewImage}" 
                     alt="${preview.title}" class="w-full h-full object-cover">
              ` : ''}
              <div class="absolute inset-0 flex items-center justify-center bg-black bg-opacity-30 dark:bg-opacity-50 group-hover:bg-opacity-20 dark:group-hover:bg-opacity-40 transition-colors duration-200">
                <div class="w-12 h-12 sm:w-16 sm:h-16 bg-red-600 hover:bg-red-700 dark:bg-red-500 dark:hover:bg-red-600 rounded-full flex items-center justify-center flex-shrink-0 transform hover:scale-110 transition-all duration-200">
                  <svg class="w-4 h-4 sm:w-6 sm:h-6 text-white ml-0.5 sm:ml-1" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M8 5v14l11-7z"/>
                  </svg>
                </div>
              </div>
              ${preview.videoDuration ? `
                <div class="absolute bottom-2 right-2 bg-black bg-opacity-80 dark:bg-opacity-90 text-white text-xs px-2 py-1 rounded max-w-[calc(100%-1rem)] truncate">
                  ${preview.videoDuration}
                </div>
              ` : ''}
            </div>
          </div>
          
          <!-- Video controls overlay (hidden initially) -->
          <div class="video-controls absolute top-2 right-2 flex gap-2 opacity-0 transition-opacity duration-200" 
               id="${videoId}-controls">
            <button onclick="pauseVideo('${videoId}')" 
                    class="pause-btn w-8 h-8 bg-black bg-opacity-70 dark:bg-opacity-80 hover:bg-opacity-90 rounded-full flex items-center justify-center text-white transition-all duration-200"
                    title="Pause">
              <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                <path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z"/>
              </svg>
            </button>
            <button onclick="resumeVideo('${videoId}')" 
                    class="play-btn w-8 h-8 bg-black bg-opacity-70 dark:bg-opacity-80 hover:bg-opacity-90 rounded-full flex items-center justify-center text-white transition-all duration-200 hidden"
                    title="Resume">
              <svg class="w-4 h-4 ml-0.5" fill="currentColor" viewBox="0 0 24 24">
                <path d="M8 5v14l11-7z"/>
              </svg>
            </button>
            <button onclick="stopVideo('${videoId}')" 
                    class="stop-btn w-8 h-8 bg-black bg-opacity-70 dark:bg-opacity-80 hover:bg-opacity-90 rounded-full flex items-center justify-center text-white transition-all duration-200"
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
        <div class="aspect-video w-full bg-black dark:bg-gray-900 overflow-hidden relative group" id="${videoId}">
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
                    class="toggle-btn w-8 h-8 bg-black bg-opacity-70 dark:bg-opacity-80 hover:bg-opacity-90 rounded-full flex items-center justify-center text-white transition-all duration-200"
                    title="Play/Pause">
              <svg class="play-icon w-4 h-4 ml-0.5" fill="currentColor" viewBox="0 0 24 24">
                <path d="M8 5v14l11-7z"/>
              </svg>
              <svg class="pause-icon w-4 h-4 hidden" fill="currentColor" viewBox="0 0 24 24">
                <path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z"/>
              </svg>
            </button>
            <button onclick="stopDirectVideo('${videoId}')" 
                    class="stop-btn w-8 h-8 bg-black bg-opacity-70 dark:bg-opacity-80 hover:bg-opacity-90 rounded-full flex items-center justify-center text-white transition-all duration-200"
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
        <div class="aspect-video w-full bg-gray-100 dark:bg-gray-800 relative group overflow-hidden cursor-pointer"
             onclick="playSocialVideo('${preview.url}', '${preview.videoType}')">
          <img src="${preview.previewImage}" alt="${preview.title}" 
               class="w-full h-full object-cover">
          <div class="absolute inset-0 flex items-center justify-center bg-black bg-opacity-30 dark:bg-opacity-50 group-hover:bg-opacity-20 dark:group-hover:bg-opacity-40 transition-colors duration-200">
            <div class="w-10 h-10 sm:w-12 sm:h-12 bg-white dark:bg-gray-300 bg-opacity-90 dark:bg-opacity-90 hover:bg-opacity-100 dark:hover:bg-opacity-100 rounded-full flex items-center justify-center flex-shrink-0 transform hover:scale-110 transition-all duration-200">
              <svg class="w-3 h-3 sm:w-4 sm:h-4 text-gray-800 dark:text-gray-900 ml-0.5" fill="currentColor" viewBox="0 0 24 24">
                <path d="M8 5v14l11-7z"/>
              </svg>
            </div>
          </div>
          <div class="absolute top-2 left-2 bg-black bg-opacity-80 dark:bg-opacity-90 text-white text-xs px-2 py-1 rounded capitalize">
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
    <div class="link-preview border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden my-2 bg-white dark:bg-gray-800 hover:shadow-md dark:hover:shadow-gray-900/30 transition-all duration-300 w-fit max-w-full">
      ${hasVideo ? `
        <div class="w-full max-w-[500px]">
          ${videoPreview}
          <div class="p-3 sm:p-4 min-w-0">
            <div class="flex items-start justify-between gap-2 min-w-0">
              <div class="flex-1 min-w-0">
                <h4 class="font-medium text-gray-900 dark:text-white text-sm sm:text-base line-clamp-2 mb-1 break-words">
                  <i class="pi pi-play-circle text-red-600 dark:text-red-500 mr-1 flex-shrink-0"></i>
                  <a href="${preview.url}" target="_blank" rel="noopener noreferrer" class="hover:text-blue-600 dark:hover:text-blue-400 break-words transition-colors duration-200">
                    ${preview.title}
                  </a>
                </h4>
                ${preview.description ? `
                  <p class="text-gray-600 dark:text-gray-400 text-xs sm:text-sm line-clamp-2 sm:line-clamp-3 mb-2 break-words leading-relaxed transition-colors duration-200">${preview.description}</p>
                ` : ''}
                <div class="flex items-center gap-1 text-xs sm:text-sm text-gray-500 dark:text-gray-400 min-w-0 transition-colors duration-200">
                  <i class="pi pi-video text-red-600 dark:text-red-500 flex-shrink-0"></i>
                  <span class="truncate min-w-0 flex-1">${preview.domain}</span>
                  ${preview.videoDuration ? `<span class="ml-2 flex-shrink-0 hidden xs:inline">• ${preview.videoDuration}</span>` : ''}
                </div>
                ${preview.videoDuration ? `
                  <div class="text-xs text-gray-500 dark:text-gray-500 mt-1 xs:hidden transition-colors duration-200">
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
            <div class="aspect-video overflow-hidden bg-gray-100 dark:bg-gray-700 transition-colors duration-200">
              <img src="${preview.previewImage}" alt="${preview.title}" 
                   class="w-full h-full object-cover"
                   onerror="this.parentElement.style.display='none'">
            </div>
          ` : ''}
          <div class="p-3 sm:p-4 min-w-0">
            <div class="flex items-start justify-between gap-2 min-w-0">
              <div class="flex-1 min-w-0">
                <h4 class="font-medium text-gray-900 dark:text-white text-sm sm:text-base line-clamp-2 mb-1 break-words transition-colors duration-200">
                  <span class="break-words">${preview.title}</span>
                </h4>
                ${preview.description ? `
                  <p class="text-gray-600 dark:text-gray-400 text-xs sm:text-sm line-clamp-2 sm:line-clamp-3 mb-2 break-words leading-relaxed transition-colors duration-200">${preview.description}</p>
                ` : ''}
                <div class="flex items-center gap-1 text-xs sm:text-sm text-gray-500 dark:text-gray-400 min-w-0 transition-colors duration-200">
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

// Debounced scroll handler to improve performance
let scrollTimeout: any = null;
function debouncedHandleScroll() {
  if (scrollTimeout) {
    clearTimeout(scrollTimeout);
  }

  scrollTimeout = setTimeout(() => {
    handleScroll();
    scrollTimeout = null;
  }, 100); // Increased for better performance
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

      // Store in pastePreviews map using current chat ID
      if (currentChatId.value) {
        pastePreviews.value.set(currentChatId.value, {
          content: processedContent,
          wordCount,
          charCount,
          show: true
        })
      }

      // Save draft immediately when large content is pasted
      if (currentChatId.value) {
        const textarea = document.getElementById('prompt') as HTMLTextAreaElement
        let currentDraft = textarea?.value || ''

        // Combine current textarea content with paste preview content
        const fullDraft = currentDraft + processedContent
        chatDrafts.value.set(currentChatId.value, fullDraft)
        saveChatDrafts()

        // Clear textarea but keep the draft with paste content
        if (textarea) {
          textarea.value = currentDraft // Keep only the typed content in textarea
          autoGrow({ target: textarea } as any)
        }
      }

      toast.info('Large content detected', {
        duration: 4000,
        description: `${wordCount} words, ${charCount} characters. Preview shown below.`
      })
    } else {
      // For small pastes, let the normal paste happen and then save draft
      setTimeout(() => {
        if (currentChatId.value) {
          const textarea = document.getElementById('prompt') as HTMLTextAreaElement
          if (textarea) {
            // For small pastes, save the normal content
            chatDrafts.value.set(currentChatId.value, textarea.value)
            saveChatDrafts()
          }
        }
      }, 100)
    }
  } catch (error) {
    console.error('Error handling paste:', error)
    // Don't prevent default on error - let normal paste proceed
  }
}

function removePastePreview() {
  // Remove paste preview for current chat
  if (currentChatId.value) {
    pastePreviews.value.delete(currentChatId.value)
    saveChatDrafts()

    // Also clear textarea if it contains paste content
    const textarea = document.getElementById('prompt') as HTMLTextAreaElement
    if (textarea && textarea.value.includes('#pastedText#')) {
      // Extract any non-pasted content
      const parts = textarea.value.split('#pastedText#')
      textarea.value = parts[0] || ''
      autoGrow({ target: textarea } as any)
    }
  }
}

// handleSubmit function
async function handleSubmit(e?: any, retryPrompt?: string) {
  e?.preventDefault?.()

  // Stop voice recording immediately when submitting
  if (isRecording.value || isTranscribing.value) {
    stopVoiceRecording(true)
  }

  // Use the global connection check
  if (!isUserOnline.value) {
    const isActuallyOnline = await checkInternetConnection()
    if (!isActuallyOnline) {
      toast.error('You are offline', {
        duration: 4000,
        description: 'Please check your internet connection and try again'
      })
      return
    }
  }

  let promptValue = retryPrompt || e?.target?.prompt?.value?.trim()

  // Check if we have paste preview content
  const currentPastePreview = pastePreviews.value.get(currentChatId.value)
  const hasPastePreview = currentPastePreview && currentPastePreview.show && !retryPrompt

  if (hasPastePreview) {
    promptValue += currentPastePreview.content
    pastePreviews.value.delete(currentChatId.value)
  }

  let fabricatedPrompt = promptValue
  if (!promptValue || isLoading.value) return

  if (!isAuthenticated.value) {
    toast.warning('Please create a session first', {
      duration: 4000,
      description: 'You need to be logged in.'
    })
    return
  }

  // Check request limits
  loadRequestCount()

  // Clear draft for current chat
  clearCurrentDraft()

  // Create new chat if none exists
  let submissionChatId = currentChatId.value
  const submissionChat = chats.value.find(chat => chat.id === submissionChatId)

  if (!submissionChatId || !submissionChat) {
    const newChatId = createNewChat(promptValue)
    if (!newChatId) return
    currentChatId.value = newChatId
    submissionChatId = newChatId
  }

  // Generate unique request ID
  const requestId = `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  const abortController = new AbortController()

  // Track the active request using global state
  activeRequests.value.set(requestId, abortController)
  requestChatMap.value.set(requestId, submissionChatId)

  // Handle link-only prompts
  if (isJustLinks(promptValue)) {
    return handleLinkOnlyRequest(promptValue, submissionChatId, requestId, abortController)
  }

  isLoading.value = true
  scrollToBottom();

  // Determine response mode - use light-response for pasted content, otherwise user preference
  let responseMode = parsedUserDetails?.value.responseMode || 'light-response'

  // Override to light-response if pasted content is detected
  if (hasPastePreview) {
    responseMode = 'light-response'
    console.log('Pasted content detected - using light-response mode')
  }

  const isSearchMode = responseMode === 'web-search' || responseMode === 'deep-search'

  // Merge with context for short prompts in non-search modes
  if (responseMode === 'light-response' && isPromptTooShort(promptValue) && currentMessages.value.length > 0) {
    const lastMessage = currentMessages.value[currentMessages.value.length - 1]
    fabricatedPrompt = `${lastMessage.prompt || ''} ${lastMessage.response || ''}\nUser: ${promptValue}`
  }

  // Clear input field
  if (!retryPrompt && e?.target?.prompt) {
    e.target.prompt.value = ""
    e.target.prompt.style.height = "auto"
  }

  // Clear voice transcription
  if (transcribedText.value) {
    transcribedText.value = ''
  }

  const tempResp: Res = {
    prompt: promptValue,
    response: responseMode ? `${responseMode}...` : "...",
  }

  // Add message to submission chat
  const targetChat = chats.value.find(chat => chat.id === submissionChatId)
  if (targetChat) {
    targetChat.messages.push(tempResp)
    targetChat.updatedAt = new Date().toISOString()

    // Update chat title if first message
    if (targetChat.messages.length === 1) {
      targetChat.title = generateChatTitle(promptValue)
    }
  }

  updateExpandedArray()
  processLinksInUserPrompt(promptValue)

  try {
    let response: Response
    let parseRes: any

    if (isSearchMode) {
      // Enhanced search request with proper parameters
      const searchParams = new URLSearchParams({
        query: encodeURIComponent(fabricatedPrompt),
        mode: responseMode === 'web-search' ? 'light-search' : 'deep-search',
        max_results: responseMode === 'deep-search' ? '5' : '5',
        content_depth: responseMode === 'deep-search' ? '2' : '1'
      })

      console.log(`Making ${responseMode} request`)

      response = await fetch(
        `${SPINDLE_URL}/search?${searchParams}`,
        {
          method: "GET",
          signal: abortController.signal,
          headers: {
            "Content-Type": "application/json",
          }
        }
      )
    } else {
      // Standard light-response mode
      console.log('Making light-response request')

      response = await fetch(WRAPPER_URL, {
        method: "POST",
        body: JSON.stringify(fabricatedPrompt),
        headers: {
          "Content-Type": "application/json"
        },
        signal: abortController.signal
      })
    }

    // Check if request was aborted
    if (abortController.signal.aborted) {
      console.log(`Request ${requestId} was aborted`)
      return
    }

    if (!response.ok) {
      const errorText = await response.text()
      throw new Error(`HTTP ${response.status}: ${errorText || response.statusText}`)
    }

    parseRes = await response.json()

    // Enhanced response processing for search modes
    let finalResponse = parseRes.error ? parseRes.error : parseRes.response

    if (isSearchMode) {
      // Check for results in both locations (results or json)
      const hasResults = parseRes.results || parseRes.json;
      if (hasResults) {
        finalResponse = formatSearchResults(parseRes, responseMode)
      } else {
        finalResponse = "No search results found for your query."
      }
    }

    // Update the response in chat
    const updatedTargetChat = chats.value.find(chat => chat.id === submissionChatId)
    if (updatedTargetChat) {
      const lastMessageIndex = updatedTargetChat.messages.length - 1
      if (lastMessageIndex >= 0) {
        const updatedMessage = {
          prompt: promptValue,
          response: finalResponse,
          status: response.status,
        }
        updatedTargetChat.messages[lastMessageIndex] = updatedMessage
        updatedTargetChat.updatedAt = new Date().toISOString()

        // Process links in the response
        await processLinksInResponse(lastMessageIndex)
      }
    }

    // Increment request count on success
    incrementRequestCount()

    // Show success notification if user switched away
    if (currentChatId.value !== submissionChatId) {
      toast.success('Response received', {
        duration: 3000,
        description: `Switch to "${targetChat?.title || 'chat'}" to view the response`
      })
    }

  } catch (err: any) {
    // Don't show error if request was intentionally aborted
    if (err.name === 'AbortError') {
      console.log(`Request ${requestId} was aborted`)
      return
    }

    console.error('AI Response Error:', err)

    // Update error in target chat
    const errorTargetChat = chats.value.find(chat => chat.id === submissionChatId)
    if (errorTargetChat && errorTargetChat.messages.length > 0) {
      const lastMessageIndex = errorTargetChat.messages.length - 1
      errorTargetChat.messages[lastMessageIndex] = {
        prompt: promptValue,
        response: `Error: ${err.message}`,
        status: 500,
      }
    }

    toast.error(`Failed to get AI response: ${err.message}`, {
      duration: 5000,
      description: 'Please try again or check your connection'
    })

    // Restore draft if request failed
    if (submissionChatId && promptValue.trim()) {
      chatDrafts.value.set(submissionChatId, promptValue)
      saveChatDrafts()
    }
  } finally {
    // Clean up request tracking
    activeRequests.value.delete(requestId)
    requestChatMap.value.delete(requestId)

    isLoading.value = false
    saveChats()

    // Trigger background sync if needed
    setTimeout(() => {
      performSmartSync().catch(error => {
        console.error('Background sync failed:', error);
      });
    }, 500);

    // Observe new video containers
    observeNewVideoContainers();
  }
}

// Helper function to format search results
function formatSearchResults(searchData: any, mode: string): string {
  // Check for results in different possible locations
  const results = searchData.results || searchData.json || [];
  if (results.length === 0) {
    return "No search results found for your query."
  }

  let formatted = "";

  if (mode === 'light-search' || mode === 'web-search') {
    const { total_pages } = searchData

    // Header with result count
    formatted += `Showing **${results.length}** of **${total_pages || results.length}** total results\n\n`
    formatted += `\n\n`

    results.forEach((result: any, index: number) => {
      const title = result.title || 'No Title'
      const url = result.url || '#'
      const description = result.description || 'No description available'

      // Result number and title
      formatted += `#### ${index + 1}. ${title} \n\n`

      // Description
      formatted += `${description} \n`

      // URL link
      formatted += `[${url}](${url}) \n\n`

      // Separator between results
      if (index < results.length - 1) {
        formatted += `\n\n\n`
      }
    })

  } else if (mode === 'deep-search') {
    const { json, total_pages, content_depth, search_time } = searchData

    // Header for deep search
    formatted += `**Advanced Analysis** • ${json.length} results analyzed at depth ${content_depth || 1}\n\n`
    formatted += `\n\n`

    // Process each result
    json.forEach((result: any, index: number) => {
      const title = result.title || 'No Title'
      const url = result.url || '#'
      const markdownContent = result.markdown_content || ''
      const depth = result.depth || 0
      const source: string = result.source || 'Unknown'

      // Result header
      formatted += `### ${index + 1}. ${title}\n\n`

      // Metadata in a quote block for styling
      formatted += `**URL:** [${url}](${url}) \n\n`
      formatted += `> **Source:** ${source.startsWith('https://') ? `[${source}](${source})` : (source.length > 60 ? source.slice(0, 60) + "..." : source)} • **Depth:** ${depth}  \n`;

      // Use the pre-formatted markdown content directly
      if (markdownContent) {
        formatted += `${markdownContent} \n\n`
      } else if (result.content) {
        // Fallback to plain content if no markdown
        formatted += `${result.content.substring(0, 500)}... \n\n`
      }

      // Separator between results
      if (index < json.length - 1) {
        formatted += `\n\n\n`
      }
    })

    // Summary section for deep search
    const totalContentResults = json.filter((r: any) => r.content || r.markdown_content).length

    formatted += `\n\n`
    formatted += `## Search Summary\n\n`

    if (totalContentResults > 0) {
      formatted += `- **Content Extracted:** ${totalContentResults} of ${json.length} results\n`
      formatted += `- **Search Depth:** ${content_depth || 1} level${(content_depth || 1) > 1 ? 's' : ''}\n`
      formatted += `- **Total Results:** ${json.length} results\n`
      formatted += `- **Pages Analyzed:** ${total_pages} pages\n`
      formatted += `- **Processing Time:** ${(search_time / 1e9).toFixed(2)}s\n\n`
    } else {
      formatted += `- **Total Results:** ${json.length} results\n`
      formatted += `- **Pages Analyzed:** ${total_pages} pages\n`
      formatted += `- **Processing Time:** ${(search_time / 1e9).toFixed(2)}s\n\n`
    }

    formatted += `> ✨ *All results have been deeply analyzed and formatted for your review.*\n`
  }

  return formatted
}

async function handleLinkOnlyRequest(promptValue: string, chatId: string, requestId: string, abortController: AbortController) {
  const urls = extractUrls(promptValue)

  isLoading.value = true

  const tempResp: Res = { prompt: promptValue, response: "..." }
  const targetChat = chats.value.find(chat => chat.id === chatId)

  if (targetChat) {
    targetChat.messages.push(tempResp)
    targetChat.updatedAt = new Date().toISOString()
  }

  try {
    let combinedResponse = `I've analyzed the link${urls.length > 1 ? "s" : ""} you shared:  \n\n`

    for (const url of urls) {
      if (abortController.signal.aborted) {
        console.log(`Link request ${requestId} was aborted`)
        return
      }

      try {
        const linkPreview = await fetchLinkPreview(url)

        // Use proper markdown with double spaces for line breaks
        combinedResponse += `### ${linkPreview.title || 'Untitled'}  \n\n`

        if (linkPreview.description) {
          combinedResponse += `${linkPreview.description}  \n\n`
        }

        combinedResponse += `**Source:** ${linkPreview.domain || new URL(url).hostname}  \n`
        combinedResponse += `**Url:** [${url}](${url})  \n\n`

        if (urls.length > 1) {
          combinedResponse += `  \n\n`
        }
      } catch (err: any) {
        combinedResponse += `### ⚠️ Error  \n\n`
        combinedResponse += `Failed to analyze: [${url}](${url})  \n\n`
        combinedResponse += `> ${err.message || 'Unknown error occurred'}  \n\n`

        if (urls.length > 1) {
          combinedResponse += `  \n\n`
        }
      }
    }

    // Add summary footer for multiple links
    if (urls.length > 1) {
      combinedResponse += `> ✨ *Analyzed ${urls.length} links* \n`
    }

    // Update the response in chat
    const updatedTargetChat = chats.value.find(chat => chat.id === chatId)
    if (updatedTargetChat) {
      const lastMessageIndex = updatedTargetChat.messages.length - 1
      updatedTargetChat.messages[lastMessageIndex] = {
        prompt: promptValue,
        response: combinedResponse.trim(),
        status: 200
      }
      updatedTargetChat.updatedAt = new Date().toISOString()
    }

    // ✅ ONLY INCREMENT ON SUCCESS for link-only prompts
    incrementRequestCount()

    // Show notification if user switched away
    if (currentChatId.value !== chatId) {
      toast.success('Links analyzed', {
        duration: 3000,
        description: `Switch to "${targetChat?.title || 'chat'}" to view the analysis`
      })
    }

  } finally {
    activeRequests.value.delete(requestId)
    requestChatMap.value.delete(requestId)
    isLoading.value = false
    saveChats()
  }
}

async function refreshResponse(oldPrompt?: string) {
  if (!isUserOnline.value) {
    const isActuallyOnline = await checkInternetConnection()
    if (!isActuallyOnline) {
      toast.error('You are offline', {
        duration: 4000,
        description: 'Please check your internet connection and try again'
      })
      return
    }
  }

  const chatIndex = chats.value.findIndex(chat => chat.id === currentChatId.value)
  if (chatIndex === -1) return

  const chat = chats.value[chatIndex]
  const msgIndex = chat.messages.findIndex(m => m.prompt === oldPrompt)
  if (msgIndex === -1) return

  const oldMessage = chat.messages[msgIndex]

  // Get the original response mode from message metadata or use current
  const originalMode = parsedUserDetails?.value?.responseMode || 'light-response'
  const isSearchMode = originalMode === 'web-search' || originalMode === 'deep-search'

  let fabricatedPrompt = oldPrompt
  if (originalMode === 'light-response' && oldPrompt && isPromptTooShort(oldPrompt) && currentMessages.value.length > 1) {
    const lastMessage = currentMessages.value[msgIndex - 1]
    fabricatedPrompt = `${lastMessage.prompt || ''} ${lastMessage.response || ''}\nUser: ${oldPrompt}`
  }

  // Check request limits for refresh too
  if (!checkRequestLimitBeforeSubmit()) {
    return
  }

  // Show placeholder while refreshing
  chat.messages[msgIndex] = {
    ...oldMessage,
    response: originalMode ? `${originalMode}...` : "Refreshing...",
  }

  // Clean up link previews if needed
  const responseUrls = extractUrls(oldMessage.response || '')
  const promptUrls = extractUrls(oldMessage.prompt || '')
  const urls = [...new Set([...responseUrls, ...promptUrls])]

  if (urls.length > 0) {
    urls.forEach(url => {
      linkPreviewCache.value.delete(url)
    })
    saveLinkPreviewCache()
  }

  // Handle link-only prompts
  if (oldPrompt && isJustLinks(oldPrompt)) {
    const urls = extractUrls(oldPrompt)

    try {
      let combinedResponse = `I've analyzed the link${urls.length > 1 ? "s" : ""} you shared:  \n\n`

      for (const url of urls) {
        try {
          const linkPreview = await fetchLinkPreview(url)

          // Use proper markdown with double spaces for line breaks
          combinedResponse += `### ${linkPreview.title || 'Untitled'}  \n\n`

          if (linkPreview.description) {
            combinedResponse += `${linkPreview.description}  \n\n`
          }

          combinedResponse += `**Source:** ${linkPreview.domain || new URL(url).hostname}  \n`
          combinedResponse += `**Url:** [${url}](${url})  \n\n`

          if (urls.length > 1) {
            combinedResponse += ` \n\n`
          }
        } catch (err: any) {
          combinedResponse += `### ⚠️ Error  \n\n`
          combinedResponse += `Failed to analyze: [${url}](${url})  \n\n`
          combinedResponse += `> ${err.message || 'Unknown error occurred'}  \n\n`

          if (urls.length > 1) {
            combinedResponse += ` \n\n`
          }
        }
      }

      // Replace the same message with the refreshed response
      chat.messages[msgIndex] = {
        ...oldMessage,
        response: combinedResponse.trim(),
        status: 200,
      }

      chat.updatedAt = new Date().toISOString()
      saveChats()

      // Re-run link previews if needed
      await processLinksInResponse(msgIndex)

      incrementRequestCount()

    } finally {
      observeNewVideoContainers()
    }

    return
  }

  try {
    let response: Response
    let parseRes: any

    if (isSearchMode) {
      // Refresh search request with same parameters
      const searchParams = new URLSearchParams({
        query: encodeURIComponent(fabricatedPrompt || ''),
        mode: originalMode === 'web-search' ? 'light-search' : 'deep-search',
        max_results: originalMode === 'deep-search' ? '5' : '5',
        content_depth: originalMode === 'deep-search' ? '2' : '1'
      })

      console.log(`Refreshing ${originalMode} request`)

      response = await fetch(
        `${SPINDLE_URL}/search?${searchParams}`,
        {
          method: "GET",
          headers: {
            "Content-Type": "application/json",
          }
        }
      )
    } else {
      // Standard light-response mode refresh
      response = await fetch(WRAPPER_URL, {
        method: "POST",
        body: JSON.stringify(fabricatedPrompt),
        headers: {
          "Content-Type": "application/json"
        }
      })
    }

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`)
    }

    parseRes = await response.json()

    let finalResponse = parseRes.error ? parseRes.error : parseRes.response

    if (isSearchMode) {
      // Check for results in both locations (results or json)
      const hasResults = parseRes.results || parseRes.json;
      if (hasResults) {
        finalResponse = formatSearchResults(parseRes, originalMode)
      } else {
        finalResponse = "No search results found for your query."
      }
    }

    // Replace the same message with the refreshed response
    chat.messages[msgIndex] = {
      ...oldMessage,
      response: finalResponse,
      status: response.status,
    }

    chat.updatedAt = new Date().toISOString()
    saveChats()

    // Re-run link previews if needed
    await processLinksInResponse(msgIndex)

    incrementRequestCount()

    toast.success('Response refreshed', {
      duration: 2000,
      description: 'The response has been updated with fresh data'
    })

  } catch (err: any) {
    console.error('Refresh error:', err)
    toast.error(`Failed to refresh response: ${err.message}`)

    // Restore original message on error
    chat.messages[msgIndex] = oldMessage
  } finally {
    saveChats()
    observeNewVideoContainers()
  }
}


// input area template logic
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

const scrollButtonPosition = computed(() => {
  // Base positions
  const basePosition = 'bottom-[130px] sm:bottom-[140px]'
  const withScrollButton = 'bottom-[130px] sm:bottom-[140px]'
  const withBanners = 'bottom-[210px] sm:bottom-[220px]'
  const withPastePreview = 'bottom-[300px] sm:bottom-[350px]'
  const withPasteAndBanners = 'bottom-[400px] sm:bottom-[420px]'

  // Priority order: paste + banners > banners > paste > scroll button > base
  if ((isRequestLimitExceeded.value || shouldShowUpgradePrompt.value) && pastePreview.value?.show) {
    return withPasteAndBanners
  } else if (isRequestLimitExceeded.value || shouldShowUpgradePrompt.value) {
    return withBanners
  } else if (pastePreview.value?.show) {
    return withPastePreview
  } else if (showScrollDownButton.value) {
    return withScrollButton
  } else {
    return basePosition
  }
})

// Update the scroll container padding computed property
const scrollContainerPadding = computed(() => {
  if ((isRequestLimitExceeded.value || shouldShowUpgradePrompt.value) && pastePreview.value?.show) {
    return 'pb-[200px] sm:pb-[190px]'
  } else if (isRequestLimitExceeded.value || shouldShowUpgradePrompt.value) {
    return 'pb-[190px] sm:pb-[200px]'
  } else if (pastePreview.value?.show) {
    return 'pb-[150px] sm:pb-[140px]'
  } else if (showScrollDownButton.value) {
    return 'pb-[140px] sm:pb-[120px]'
  } else {
    return 'pb-[110px] sm:pb-[120px]'
  }
})

// Add connection checking before authentication
async function handleStepSubmit(e: Event) {
  e.preventDefault()

  // Early validation with improved error handling
  if (!validateCurrentStep()) {
    handleValidationError()
    return
  }

  // Handle step progression vs final submission
  if (authStep.value < 4) {
    nextAuthStep()
    return
  }

  // Final step - create session
  await handleFinalAuthStep()
}

function handleValidationError() {
  const errorMessages = {
    1: {
      title: 'Invalid Username',
      message: 'Username must be 2-50 characters and contain only letters, numbers, and underscores'
    },
    2: {
      title: 'Invalid Email',
      message: 'Please enter a valid email address (e.g., name@example.com)'
    },
    3: {
      title: 'Weak Password',
      message: 'Password must be at least 7 characters with a mix of letters and numbers'
    },
    4: {
      title: 'Terms Required',
      message: 'Please accept the Terms of Service and Privacy Policy to continue'
    }
  }

  const error = errorMessages[authStep.value as keyof typeof errorMessages]
  if (error) {
    toast.warning(error.title, {
      duration: 4000,
      description: error.message,
      action: {
        label: 'Got it',
        onClick: () => { }
      }
    })
  }
}

async function handleFinalAuthStep() {
  try {
    isLoading.value = true

    // Add a small delay to show the loading state
    await new Promise(resolve => setTimeout(resolve, 500))

    const response = await handleAuth(authData.value)

    // Validate response structure
    if (!response) {
      throw new Error('No response received from authentication service')
    }

    if (response.error) {
      throw new Error(response.error)
    }

    if (!response.data || !response.success) {
      throw new Error('Authentication failed - invalid response structure')
    }

    // Success handling
    await handleAuthSuccess(response)

  } catch (err: any) {
    await handleAuthError(err)
  } finally {
    isLoading.value = false
  }
}

async function handleAuthSuccess(response: any) {
  // Reset form state
  setShowCreateSession(false)
  authStep.value = 1
  authData.value = {
    username: '',
    email: '',
    password: '',
    agreeToTerms: false
  }

  // Load user data
  await loadRequestCount()

  // Handle redirect logic
  await handlePostAuthRedirect()

  // Focus input field
  nextTick(() => {
    const textarea = document.getElementById("prompt") as HTMLTextAreaElement
    if (textarea) {
      textarea.focus()
      // Add a subtle animation to draw attention to the input
      textarea.classList.add('ring-2', 'ring-blue-500')
      setTimeout(() => {
        textarea.classList.remove('ring-2', 'ring-blue-500')
      }, 2000)
    }
  })
}

async function handlePostAuthRedirect() {
  // Check multiple sources for upgrade redirect
  const previousRoute = document.referrer
  const urlParams = new URLSearchParams(window.location.search)
  const isFromUpgrade =
    previousRoute.includes('/upgrade') ||
    urlParams.has('from') && urlParams.get('from') === 'upgrade' ||
    urlParams.has('redirect') && urlParams.get('redirect') === 'upgrade'

  if (isFromUpgrade) {
    console.log('Redirecting to upgrade page after authentication')
    // Small delay for better UX flow
    await new Promise(resolve => setTimeout(resolve, 1000))
    router.push('/upgrade')
  }
}

async function handleAuthError(err: any) {
  console.error('Authentication error:', err)

  // Map specific error types to user-friendly messages
  const errorMap = {
    timeout: {
      title: 'Connection Timeout',
      message: 'Server took too long to respond. Please check your connection and try again.'
    },
    network: {
      title: 'Network Error',
      message: 'Unable to reach our servers. Please check your internet connection.'
    },
    credentials: {
      title: 'Invalid Credentials',
      message: 'The username, email, or password you entered is incorrect.'
    },
    server: {
      title: 'Server Error',
      message: 'Our servers are experiencing issues. Please try again in a few minutes.'
    },
    rate_limit: {
      title: 'Too Many Attempts',
      message: 'Please wait a moment before trying again.'
    },
    default: {
      title: 'Authentication Failed',
      message: 'An unexpected error occurred. Please try again.'
    }
  }

  // Determine error type
  let errorType: keyof typeof errorMap = 'default'
  const errorMessage = err.message?.toLowerCase() || ''

  if (errorMessage.includes('timeout')) errorType = 'timeout'
  else if (errorMessage.includes('failed to fetch') || errorMessage.includes('network')) errorType = 'network'
  else if (errorMessage.includes('http 4') || errorMessage.includes('invalid') || errorMessage.includes('credential')) errorType = 'credentials'
  else if (errorMessage.includes('http 5')) errorType = 'server'
  else if (errorMessage.includes('rate') || errorMessage.includes('limit')) errorType = 'rate_limit'

  const error = errorMap[errorType]

  // Show error with retry option
  toast.error(error.title, {
    duration: 6000,
    description: error.message,
    action: {
      label: 'Try Again',
      onClick: () => {
        // Auto-focus the relevant input field based on step
        nextTick(() => {
          const focusMap = {
            1: 'username',
            2: 'email',
            3: 'password',
            4: 'agreeToTerms'
          }
          const fieldToFocus = focusMap[authStep.value as keyof typeof focusMap]
          if (fieldToFocus && fieldToFocus !== 'agreeToTerms') {
            const input = document.querySelector(`[name="${fieldToFocus}"]`) as HTMLInputElement
            if (input) {
              input.focus()
              input.select()
            }
          }
        })
      }
    }
  })
}

// Process links in a response and generate previews
async function processLinksInResponse(index: number) {
  const targetChat = chats.value.find(chat => chat.id === currentChatId.value)
  if (!targetChat || !targetChat.messages[index] || !targetChat.messages[index].response || targetChat.messages[index].response === "...") return

  const urls = extractUrls(targetChat.messages[index].response)
  if (urls.length > 0) {
    urls.slice(0, 3).forEach(url => {
      fetchLinkPreview(url).then(() => {
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
    })
  } else {
    copyCode(text)
    toast.info('Copied Instead', {
      duration: 3000,
    })
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

  const clickableClass = isClickable ? 'paste-preview-clickable cursor-pointer hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors duration-200' : ''

  return `
    <div class="paste-preview border border-gray-300 dark:border-gray-600 rounded-lg overflow-hidden my-2 bg-gray-100 dark:bg-gray-700 hover:shadow-md transition-all duration-300 w-full ${clickableClass}" 
         id="${componentId}" ${clickableAttributes}>
      <div class="w-full">
        <div class="bg-gray-600 dark:bg-gray-800 px-3 py-1 text-white dark:text-gray-200 text-xs font-medium flex items-center gap-2 transition-colors duration-200">
          <i class="pi pi-clipboard text-gray-300 dark:text-gray-400"></i>
          <span>PASTED CONTENT</span>
          <span class="ml-auto text-gray-200 dark:text-gray-400 hidden sm:inline">${wordCount} words • ${charCount} chars</span>
          <span class="ml-auto text-gray-200 dark:text-gray-400 sm:hidden">${charCount} chars</span>
          ${isClickable ? '<i class="pi pi-external-link ml-1 text-gray-300 dark:text-gray-500"></i>' : ''}
        </div>
        <div class="pb-3 px-3">
          <div class="relative">
            <div class="text-sm text-gray-800 dark:text-gray-200 leading-relaxed break-words whitespace-pre-wrap font-mono h-20 sm:h-24 overflow-hidden transition-colors duration-200">
              ${escapedPreview}
            </div>
            <div class="absolute bottom-0 left-0 right-0 h-8 bg-gradient-to-t from-gray-100 dark:from-gray-700 via-gray-100/80 dark:via-gray-700/80 to-transparent pointer-events-none transition-colors duration-200"></div>
          </div>
          <div class="flex items-center justify-between mt-2 text-xs text-gray-600 dark:text-gray-400 transition-colors duration-200">
            <span class="hidden sm:inline">${isClickable ? 'Click to view full content' : 'Large content detected'}</span>
            <span class="sm:hidden">${isClickable ? 'Tap to view' : 'Large content'}</span>
            ${!isClickable ? '<button class="remove-paste-preview text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-gray-100 underline font-medium transition-colors duration-200" type="button">Remove</button>' : ''}
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
    // Only process if we're still recording (FIX 6)
    if (!isRecording.value) return

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

    if (event.error !== 'aborted') { // Don't show toast for manual stops
      toast.error('Voice Input Error', {
        duration: 4000,
        description: event.error === 'not-allowed' ? 'Microphone access denied' : event.error
      })
    }
  }

  recognition.onend = () => {
    // Only restart if we're still supposed to be recording (FIX 5)
    if (isRecording.value && !isTranscribing.value) {
      setTimeout(() => {
        if (isRecording.value) { // Double check we're still recording
          recognition.start()
        }
      }, 500)
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
    if (textarea && (isRecording.value || transcribedText.value)) {
      // Only update if we're actively recording or have transcribed text
      textarea.value = transcribedText.value + interim
      autoGrow({ target: textarea } as any)
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
    stopVoiceRecording(false) // Don't clear text - let user decide
  } else {
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
function stopVoiceRecording(clearText: boolean = true) {
  isRecording.value = false
  isTranscribing.value = false
  stopTimer()
  voiceRecognition.value?.stop()

  // Clear transcribed text if requested (FIX 2 & 5)
  if (clearText) {
    clearVoiceTranscription()
  }
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
  transcriptionDuration.value = 0 // Reset duration
  const textarea = document.getElementById('prompt') as HTMLTextAreaElement
  if (textarea) {
    textarea.value = ''
    autoGrow({ target: textarea } as any)
    textarea.focus()
  }
}

// Select input mode and handle special actions
async function selectInputMode(mode: 'web-search' | 'deep-search' | 'light-response') {
  // Store original value for rollback
  const originalMode = parsedUserDetails.value.responseMode

  // Don't do anything if same mode
  if (originalMode === mode) {
    showInputModeDropdown.value = false
    return
  }

  try {
    // Update in-memory state - the watch will handle syncing
    parsedUserDetails.value.responseMode = mode
    showInputModeDropdown.value = false

    const modeNames = {
      'web-search': 'Web Search',
      'deep-search': 'Deep Search',
      'light-response': 'Light Response'
    }

    // Show success immediately - sync happens in background via watch
    toast.success(`Switched to ${modeNames[mode]}`, {
      duration: 2000,
      description: mode === 'web-search'
        ? 'Responses will include web search results'
        : mode === 'deep-search'
          ? 'Responses will be more detailed and thorough'
          : 'Responses will be quick and concise'
    })
  } catch (error) {
    console.error('Error selecting input mode:', error)
    parsedUserDetails.value.responseMode = originalMode

    toast.error('Failed to change mode', {
      duration: 3000,
      description: 'An error occurred'
    })
  }
}

// Close dropdown when clicking outside
const handleClickOutside = (e: MouseEvent) => {
  const dropdown = document.querySelector('.relative .absolute')
  const button = document.querySelector('.relative button')

  if (dropdown && !dropdown.contains(e.target as Node) &&
    button && !button.contains(e.target as Node)) {
    showInputModeDropdown.value = false
  }

  // Close suggestions dropup
  const suggestionsDropup = document.querySelector('.suggestions-dropup')
  const suggestionsButton = document.querySelector('.suggestions-button')

  if (suggestionsDropup && !suggestionsDropup.contains(e.target as Node) &&
    suggestionsButton && !suggestionsButton.contains(e.target as Node)) {
    showSuggestionsDropup.value = false
  }
}

const modeOptions: Record<string, ModeOption> = {
  'light-response': {
    mode: 'light-response',
    label: 'Quick Response',
    description: 'Fast & concise',
    icon: 'pi pi-bolt',
    title: 'Quick Response - Click to change mode'
  },
  'web-search': {
    mode: 'web-search',
    label: 'Web Search',
    description: 'Include web results',
    icon: 'pi pi-search',
    title: 'Web Search - Click to change mode'
  },
  'deep-search': {
    mode: 'deep-search',
    label: 'Deep Search',
    description: 'Detailed analysis',
    icon: 'pi pi-code',
    title: 'Deep Search - Click to change mode'
  }
}

onUpdated(() => {
  // Check for new video containers after DOM updates
  observeNewVideoContainers();
});

// Watch for chat switches to manage requests
watch(currentChatId, (newChatId, oldChatId) => {
  loadChatDrafts()

  if (oldChatId && newChatId !== oldChatId) {
    // Clear paste preview when switching chats
    // pastePreviews.value.delete(oldChatId)

    // Cancel ongoing requests for the old chat (optional - remove if you want them to continue)
    // cancelChatRequests(oldChatId)

    // User switched chats - stop any active recording
    if (isRecording.value || isTranscribing.value) {
      stopVoiceRecording(true)
      toast.info('Voice recording stopped', {
        duration: 2000,
        description: 'Switched to different chat'
      })
    }
  }
})

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
  planName: parsedUserDetails.value?.planName,
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

watch([() => currentMessages.value.length, () => chats.value], () => {
  nextTick(() => {
    setTimeout(() => {
      handleScroll(); // Recalculate scroll position after content changes
    }, 100);
  });
}, { deep: true });

watch(() => route.path, (newPath, oldPath) => {
  if (newPath === "/new") {
    createNewChat()
    setShowInput()
    router.replace(`${oldPath}`)
  }
}, { immediate: true })

onBeforeUnmount(() => {
  if (transcriptionTimer) clearInterval(transcriptionTimer);
  if (updateTimeout) clearTimeout(updateTimeout);

  // Clean up all active requests
  cancelAllRequests()

  // Clean up speech recognition
  if (isRecording.value || isTranscribing.value) {
    stopVoiceRecording(false) // Don't clear text during unmount
  }

  // Remove keyboard listener
  document.removeEventListener('keydown', handleModalKeydown)
  document.removeEventListener('click', handleClickOutside)

  // Clean up paste preview handlers (use the enhanced cleanup function)
  cleanupPastePreviewHandlers()

  // Restore body scroll if modal is open
  if (showPasteModal.value) {
    document.body.style.overflow = 'auto'
  }

  // Clear debounce timers
  if (chatsDebounceTimer) {
    clearTimeout(chatsDebounceTimer)
  }
  if (userDetailsDebounceTimer) {
    clearTimeout(userDetailsDebounceTimer)
  }
})

// Consolidated onMounted hook for better organization
onMounted(() => {
  if (route.path === "/new") {
    createNewChat()
    setShowInput()
    router.replace("/")
  }

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

  document.addEventListener('click', handleClickOutside)

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
    setShowInput()

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
  const resetCheckInterval = setInterval(loadRequestCount, 5 * 60 * 1000);

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

      // Set up scroll listener after initial render
      if (scrollableElem.value) {
        scrollableElem.value.addEventListener("scroll", debouncedHandleScroll, {
          passive: true
        });

        // Trigger initial scroll state calculation
        setTimeout(() => {
          handleScroll();
        }, 200);
      }
    }, 300); // Increased delay for more reliable initial render;
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

onUnmounted(() => {
  // Final cleanup for voice recording
  if (voiceRecognition.value) {
    voiceRecognition.value.abort()
  }

  if (transcriptionTimer) {
    clearInterval(transcriptionTimer)
  }

  if (updateTimeout) {
    clearTimeout(updateTimeout)
  }
})
</script>

<template>
  <div class="flex h-[100vh] bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100">
    <!-- Sidebar -->
    <SideNav v-if="isAuthenticated" :data="{
      chats,
      currentChatId,
      parsedUserDetails,
      screenWidth,
      isCollapsed,
    }" :functions="{
      setShowInput,
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
        }" :functions="{
          manualSync,
        }" />
        <!-- Empty State -->
        <CreateSessView v-if="!isAuthenticated" :data="{
          chats,
          currentChatId,
          isCollapsed,
          parsedUserDetails,
          screenWidth,
          syncStatus,
          isLoading,
          authStep,
          showCreateSession,
          authData,
          currentMessages,
        }" :functions="{
          validateCurrentStep,
          setShowInput,
          clearAllChats,
          toggleSidebar,
          logout,
          createNewChat,
          switchToChat,
          deleteChat,
          renameChat,
          manualSync,
          handleStepSubmit,
          prevAuthStep,
          updateAuthData,
          setShowCreateSession,
        }" />

        <div v-else-if="isAuthenticated && currentMessages.length === 0">
          <div
            class="flex md:max-w-3xl max-w-[100vw] max-md:px-4 flex-col md:flex-grow items-center gap-3 text-gray-600 dark:text-gray-400">
            <img :src="parsedUserDetails?.theme=== 'dark' || (parsedUserDetails?.theme=== 'system' && isDarkMode) ?
              '/logo-light.svg' : '/logo.svg'" alt="Gemmie Logo" class="w-[60px] h-[60px] rounded-md" />

            <p class="text-3xl text-black dark:text-white font-semibold">{{ parsedUserDetails?.username || 'Gemmie' }}
            </p>
            <div class="text-center max-w-md space-y-2">
              <p class="text-gray-600 dark:text-gray-400 leading-relaxed">
                Experience privacy-first conversations with advanced AI. Your data stays secure, local and synced to
                your
                all devices.
              </p>
              <div class="flex items-center justify-center gap-6 text-sm text-gray-500 dark:text-gray-400 mt-4">
                <div class="flex items-center gap-1">
                  <i class="pi pi-shield text-green-500 dark:text-green-400"></i>
                  <span>Private</span>
                </div>
                <div class="flex items-center gap-1">
                  <i class="pi pi-database text-blue-500 dark:text-blue-400"></i>
                  <span>Local Stored</span>
                </div>
                <div class="flex items-center gap-1">
                  <i class="pi pi-sync text-purple-500 dark:text-purple-400"></i>
                  <span>Synced</span>
                </div>
              </div>
            </div>
            
            <div class="flex flex-col gap-4 w-full mb-[100px] max-w-2xl relative">
              <!-- Suggestion Chips Grid -->
              <div class="w-full flex justify-center">
                <div class="flex flex-wrap justify-center gap-2">
                  <button v-for="(suggestion, index) in suggestionPrompts" :key="index" type="button"
                    @click="selectSuggestion(suggestion.prompt)"
                    class="group flex w-[100px] items-center gap-2 justify-center h-9 bg-white dark:bg-gray-800 border-[1px] border-gray-200 dark:border-gray-700 rounded-lg hover:border-blue-500 dark:hover:border-blue-400 hover:bg-gray-50 dark:hover:bg-gray-700 transition-all duration-200 transform hover:scale-105 shadow-sm hover:shadow-md">
                    <i
                      :class="[suggestion.icon, 'text-gray-500 dark:text-gray-300 text-sm group-hover:scale-110 transition-transform']"></i>
                    <span class="text-xs font-medium text-gray-700 dark:text-gray-300">
                      {{ suggestion.title }}
                    </span>
                  </button>
                </div>
              </div>

              <!-- Start Writing Button -->
              <button v-if="!showInput" @click="setShowInput"
                class="group px-6 py-3 bg-gradient-to-r from-blue-500 to-purple-600 dark:from-blue-600 dark:to-purple-700 text-white rounded-lg hover:from-blue-600 hover:to-purple-700 dark:hover:from-blue-700 dark:hover:to-purple-800 transition-all duration-300 transform hover:scale-105 shadow-lg hover:shadow-xl font-medium">
                <span class="flex items-center justify-center gap-2">
                  <i class="pi pi-pencil group-hover:rotate-12 transition-transform"></i>
                  Start Writing
                </span>
              </button>
            </div>
          </div>
        </div>

        <!-- Chat Messages Container -->
        <div ref="scrollableElem" v-else-if="currentMessages.length !== 0 && isAuthenticated"
          class="relative md:max-w-3xl max-w-[100vw] flex-grow no-scrollbar overflow-y-auto px-2 w-full space-y-3 sm:space-y-4 mt-[55px] pt-8  scroll-container"
          :class="scrollContainerPadding">
          <div v-if="currentMessages.length !== 0" v-for="(item, i) in currentMessages" :key="`chat-${i}`"
            class="flex flex-col gap-1">
            <!-- User Bubble -->
            <div class="flex w-full chat-message">
              <div class="flex flex-col w-full">
                <!-- User message bubble -->
                <div
                  class="flex items-start gap-2 font-medium bg-gray-100 dark:bg-gray-800 text-black dark:text-gray-100 px-4 rounded-2xl prose prose-sm dark:prose-invert chat-bubble w-fit max-w-full">
                  <!-- Avatar container -->
                  <div class="flex-shrink-0 py-3">
                    <div
                      class="flex items-center justify-center w-7 h-7 text-gray-100 dark:text-gray-800 bg-gray-700 dark:bg-gray-200 rounded-full text-sm font-semibold">
                      {{ parsedUserDetails.username.toUpperCase().slice(0, 2) }}
                    </div>
                  </div>

                  <!-- Message content container -->
                  <div class="flex-1 min-w-0">
                    <div class="break-words text-base leading-relaxed"
                      v-html="renderMarkdown((item && item?.prompt && item?.prompt?.length > 800) ? item?.prompt?.trim().split('#pastedText#')[0] : item.prompt || '')">
                    </div>
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
            <div class="flex w-full md:max-w-3xl max-w-full relative pb-[20px]">
              <div
                class="bg-none max-w-full w-full chat-message leading-relaxed text-black dark:text-gray-100 p-1 rounded-2xl prose prose-sm dark:prose-invert">
                <!-- Loading state -->
                <div v-if="isLoadingState(item.response)" 
                    class="flex w-full rounded-lg bg-gray-50 dark:bg-gray-800 p-2 items-center animate-pulse gap-2 text-gray-500 dark:text-gray-400">
                  <i class="pi pi-spin pi-spinner"></i>
                  <span class="text-sm">{{ getLoadingMessage(item.response) }}</span>
                </div>

                <!-- Regular response with enhanced link handling -->
                <div v-else>
                  <div class="break-words overflow-x-hidden" v-html="renderMarkdown(item.response || '')"></div>

                  <!-- Link Previews Section -->
                  <div v-if="extractUrls(item.response || '').length > 0" class="mt-2 sm:mt-3">
                    <div v-for="url in extractUrls(item.response || '').slice(0, 3)" :key="url">
                      <div v-if="linkPreviewCache.get(url)"
                        v-html="LinkPreviewComponent({ preview: linkPreviewCache.get(url)! })"></div>
                    </div>
                  </div>
                </div>

                <!-- Actions - Responsive with fewer labels on mobile -->
                <div
                  v-if="!isLoadingState(item.response)"
                  class="flex flex-wrap justify-end gap-2 sm:gap-3 mt-2 text-gray-500 dark:text-gray-400 text-sm">
                  <button @click="copyResponse(item.response, i)"
                    class="flex items-center gap-1 hover:text-blue-600 dark:hover:text-blue-400 transition-colors min-h-[32px]">
                    <i class="pi pi-copy"></i>
                    <span>{{ copiedIndex === i ? 'Copied!' : 'Copy' }}</span>
                  </button>

                  <button @click="shareResponse(item.response, item.prompt)"
                    class="flex items-center gap-1 hover:text-green-600 dark:hover:text-green-400 transition-colors min-h-[32px]">
                    <i class="pi pi-share-alt"></i>
                    <span>Share</span>
                  </button>

                  <button @click="refreshResponse(item.prompt)" :disabled="isLoading"
                    class="flex items-center gap-1 hover:text-orange-600 dark:hover:text-orange-400 transition-colors disabled:opacity-50 disabled:cursor-not-allowed min-h-[32px]">
                    <i class="pi pi-refresh"></i>
                    <span>Refresh</span>
                  </button>

                  <button @click="deleteMessage(i)" :disabled="isLoading"
                    class="flex items-center gap-1 hover:text-red-600 dark:hover:text-red-400 transition-colors disabled:opacity-50 disabled:cursor-not-allowed min-h-[32px]">
                    <i class="pi pi-trash"></i>
                    <span>Delete</span>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Responsive Scroll to Bottom Button -->
        <button v-if="showScrollDownButton && currentMessages.length !== 0 && isAuthenticated" @click="scrollToBottom()"
          :class="[
            'absolute bg-gray-50 dark:bg-gray-800 text-gray-500 dark:text-gray-400 border dark:border-gray-700 px-4 h-8 rounded-full shadow-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors z-20 flex items-center justify-center gap-2',
            scrollButtonPosition
          ]" :disabled="isRecording" :title="isRecording ? 'Recording in progress' : 'Scroll to bottom'">
          <i class="pi pi-arrow-down text-xs" :class="{ 'animate-bounce': !isRecording }"></i>
          <span class="text-sm font-medium">Scroll Down</span>
        </button>

        <!-- Input Area -->
        <div v-if="(currentMessages.length !== 0 || showInput === true) && isAuthenticated" :style="screenWidth > 720 && !isCollapsed ? 'left:270px;' :
          screenWidth > 720 && isCollapsed ? 'left:60px;' : 'left:0px;'"
          class="bg-white dark:bg-gray-900 z-20 bottom-0 right-0 fixed" :class="pastePreview?.show ? 'pt-2' : ''">

          <div class="flex items-center justify-center pb-3 sm:pb-5 px-2 sm:px-5">
            <form @submit="handleSubmit"
              class="w-full md:max-w-3xl relative flex bg-gray-50 dark:bg-gray-800 flex-col border-2 dark:border-gray-700 shadow rounded-2xl items-center">

              <!-- Paste Preview inside form - above other content -->
              <div v-if="pastePreview && pastePreview.show" class="w-full p-3 border-b dark:border-gray-700">
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
                        class="w-6 h-6 sm:w-8 sm:h-8 bg-red-100 dark:bg-red-900/30 rounded-full flex items-center justify-center flex-shrink-0">
                        <i class="pi pi-ban text-red-600 dark:text-red-400 text-xs sm:text-sm"></i>
                      </div>
                      <div class="min-w-0 flex-1">
                        <h3 class="text-xs sm:text-sm font-semibold text-red-800 dark:text-red-400 leading-tight">
                          {{ planStatus.isExpired ? 'Plan Expired' : 'Free Requests Exhausted' }}
                        </h3>
                        <p class="text-xs text-red-600 dark:text-red-400 leading-tight mt-0.5">
                          {{ planStatus.isExpired ? 'Renew your plan' : `Used all ${FREE_REQUEST_LIMIT} requests` }}
                        </p>
                      </div>
                    </div>
                    <button @click="$router.push('/upgrade')"
                      class="w-full bg-red-500 hover:bg-red-600 dark:bg-red-600 dark:hover:bg-red-700 text-white py-2 rounded-md text-xs font-medium transition-colors">
                      {{ planStatus.isExpired ? 'Renew Plan' : 'Upgrade Plan' }}
                    </button>
                  </div>

                  <!-- Desktop: Horizontal Layout -->
                  <div class="hidden sm:flex w-full items-center gap-3">
                    <div
                      class="w-8 h-8 bg-red-100 dark:bg-red-900/30 rounded-full flex items-center justify-center flex-shrink-0">
                      <i class="pi pi-ban text-red-600 dark:text-red-400 text-sm"></i>
                    </div>
                    <div class="min-w-0 flex-1">
                      <h3 class="text-sm font-semibold text-red-800 dark:text-red-400 mb-1">
                        {{ planStatus.isExpired ? 'Plan Expired' : 'Free Requests Exhausted' }}
                      </h3>
                      <p class="text-xs text-red-600 dark:text-red-400">
                        {{ planStatus.isExpired ?
                          'Please renew your plan to continue using the service.' :
                          `You've used all ${FREE_REQUEST_LIMIT} free requests. Upgrade to continue chatting.` }}
                      </p>
                    </div>
                    <button @click="$router.push('/upgrade')"
                      class="bg-red-500 px-3 hover:bg-red-600 dark:bg-red-600 dark:hover:bg-red-700 text-white py-2 rounded-md text-sm font-medium transition-colors flex-shrink-0 whitespace-nowrap">
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
                        class="w-6 h-6 sm:w-8 sm:h-8 bg-yellow-100 dark:bg-yellow-900/30 rounded-full flex items-center justify-center flex-shrink-0">
                        <i
                          class="pi pi-exclamation-triangle text-yellow-600 dark:text-yellow-400 text-xs sm:text-sm"></i>
                      </div>
                      <div class="min-w-0 flex-1">
                        <h3 class="text-xs sm:text-sm font-semibold text-yellow-800 dark:text-yellow-400 leading-tight">
                          {{ requestsRemaining }} requests left
                        </h3>
                        <p class="text-xs text-yellow-600 dark:text-yellow-400 leading-tight mt-0.5">
                          Upgrade for unlimited access
                        </p>
                      </div>
                    </div>
                    <button @click="$router.push('/upgrade')"
                      class="w-full bg-orange-500 hover:bg-orange-600 dark:bg-orange-600 dark:hover:bg-orange-700 text-white py-2 rounded-md text-xs font-medium transition-colors">
                      Upgrade Plan
                    </button>
                  </div>

                  <!-- Desktop: Horizontal Layout -->
                  <div class="hidden sm:flex w-full items-center gap-3">
                    <div
                      class="w-8 h-8 bg-yellow-100 dark:bg-yellow-900/30 rounded-full flex items-center justify-center flex-shrink-0">
                      <i class="pi pi-exclamation-triangle text-yellow-600 dark:text-yellow-400 text-sm"></i>
                    </div>
                    <div class="min-w-0 flex-1">
                      <h3 class="text-sm font-semibold text-yellow-800 dark:text-yellow-400 mb-1">
                        {{ requestsRemaining }} free requests remaining
                      </h3>
                      <p class="text-xs text-yellow-600 dark:text-yellow-400">
                        Upgrade to continue chatting without limits
                      </p>
                    </div>
                    <button @click="$router.push('/upgrade')"
                      class="bg-orange-500 px-3 hover:bg-orange-600 dark:bg-orange-600 dark:hover:bg-orange-700 text-white py-2 rounded-md text-sm font-medium transition-colors flex-shrink-0 whitespace-nowrap">
                      Upgrade Plan
                    </button>
                  </div>
                </div>
              </div>

              <!-- Input Area with Voice Recording - FLEX COL LAYOUT -->
              <div class="flex flex-col w-full bg-white dark:bg-gray-900 rounded-2xl px-2 sm:px-3 py-2 gap-1 sm:gap-2"
                :class="inputDisabled ? 'opacity-50 border border-t dark:border-gray-700 pointer-events-none' :
                  showUpgradeBanner ? 'border border-t dark:border-gray-700' : ''">

                <div class="w-full items-center justify-center flex">
                  <!-- Voice Recording Indicator (when active) - Now aligned horizontally -->
                  <div v-if="isRecording || isTranscribing"
                    class="flex items-center gap-1 sm:gap-2 px-2 py-1 bg-red-50 dark:bg-red-900/30 rounded-lg border border-red-200 dark:border-red-800 text-red-600 dark:text-red-400 text-xs sm:text-sm flex-shrink-0 h-fit">
                    <div class="flex items-center gap-1">
                      <div class="w-2 h-2 bg-red-500 dark:bg-red-400 rounded-full animate-pulse"></div>
                      <span class="hidden sm:inline">{{ isTranscribing ? 'Listening...' : 'Starting...' }}</span>
                      <span class="sm:hidden">{{ isTranscribing ? '🎤' : '⏳' }}</span>
                    </div>
                  </div>

                  <!-- Clear Voice Button (when transcribed text exists) -->
                  <button v-if="transcribedText && !isRecording" type="button" @click="clearVoiceTranscription"
                    class="rounded-lg w-8 h-8 sm:w-9 sm:h-9 flex items-center justify-center transition-colors text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 flex-shrink-0"
                    title="Clear voice transcription">
                    <svg class="w-4 h-4 sm:w-5 sm:h-5" fill="currentColor" viewBox="0 0 24 24">
                      <path
                        d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z" />
                    </svg>
                  </button>

                  <!-- Textarea - Now takes remaining space alongside the indicator -->
                  <textarea required id="prompt" name="prompt" @keydown="onEnter" @input="autoGrow" @paste="handlePaste"
                    :disabled="inputDisabled" rows="1" :class="[
                      'flex-grow py-3 px-3 placeholder:text-gray-500 dark:placeholder:text-gray-400 rounded-xl bg-white dark:bg-gray-900 dark:text-gray-100 text-sm outline-none resize-none max-h-[120px] sm:max-h-[150px] md:max-h-[200px] overflow-auto leading-relaxed min-w-0',
                      'disabled:opacity-50 disabled:cursor-not-allowed',
                      isRecording ? 'bg-red-50 border-red-200 dark:border-red-800' : 'focus:border-blue-500 dark:focus:border-blue-400'
                    ]" :placeholder="inputPlaceholderText">
                  </textarea>
                </div>

                <!-- Buttons Row - Below textarea -->
                <div class="flex items-center justify-between w-full gap-2">
                  <!-- Left side buttons -->
                  <div class="flex items-center gap-2">
                    <!-- Microphone Toggle Button -->
                    <button type="button" @click="toggleVoiceRecording" :disabled="inputDisabled" :class="[
                      'rounded-lg w-8 h-8 sm:w-9 sm:h-9 flex items-center justify-center transition-all duration-200 flex-shrink-0',
                      isRecording
                        ? 'bg-red-500 hover:bg-red-600 text-white shadow-lg transform scale-105 animate-pulse'
                        : 'bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-600 dark:text-gray-300 hover:text-gray-700 dark:hover:text-gray-200',
                      'disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none'
                    ]" :title="microphonePermission === 'denied'
                      ? 'Microphone access denied'
                      : isRecording
                        ? 'Stop voice input'
                        : 'Start voice input'">

                      <!-- Microphone Icon -->
                      <svg v-if="microphonePermission === 'prompt'" class="w-4 h-4 sm:w-5 sm:h-5" fill="currentColor"
                        viewBox="0 0 24 24">
                        <path d="M12 2a3 3 0 0 0-3 3v6a3 3 0 0 0 6 0V5a3 3 0 0 0-3-3Z" />
                        <path d="M19 10v1a7 7 0 0 1-14 0v-1a1 1 0 0 1 2 0v1a5 5 0 0 0 10 0v-1a1 1 0 0 1 2 0Z" />
                      </svg>

                      <svg v-else-if="!isRecording && microphonePermission === 'granted'" class="w-4 h-4 sm:w-5 sm:h-5"
                        fill="currentColor" viewBox="0 0 24 24">
                        <path d="M12 2a3 3 0 0 0-3 3v6a3 3 0 0 0 6 0V5a3 3 0 0 0-3-3Z" />
                        <path d="M19 10v1a7 7 0 0 1-14 0v-1a1 1 0 0 1 2 0v1a5 5 0 0 0 10 0v-1a1 1 0 0 1 2 0Z" />
                      </svg>

                      <!-- Stop Icon -->
                      <svg v-else-if="microphonePermission === 'granted' && isRecording" class="w-4 h-4 sm:w-5 sm:h-5"
                        fill="currentColor" viewBox="0 0 24 24">
                        <rect x="6" y="6" width="12" height="12" rx="2" />
                      </svg>

                      <!-- Microphone Denied Icon -->
                      <svg v-else-if="microphonePermission === 'denied' && !isRecording"
                        class="w-4 h-4 sm:w-5 sm:h-5 text-red-500 dark:text-red-400" fill="currentColor"
                        viewBox="0 0 24 24">
                        <path d="M12 2a3 3 0 0 0-3 3v6a3 3 0 0 0 6 0V5a3 3 0 0 0-3-3Z" />
                        <path d="M19 10v1a7 7 0 0 1-14 0v-1a1 1 0 0 1 2 0v1a5 5 0 0 0 10 0v-1a1 1 0 0 1 2 0Z" />
                        <line x1="4" y1="4" x2="20" y2="20" stroke="currentColor" stroke-width="2" />
                      </svg>
                    </button>

                    <!-- Mode Dropdown Container -->
                    <div class="relative flex-shrink-0">
                      <!-- Dropdown Button - Shows current mode -->
                      <button type="button" @click.stop="showInputModeDropdown = !showInputModeDropdown"
                        :disabled="inputDisabled" :class="[
                          'rounded-lg w-8 h-8 sm:w-9 sm:h-9 flex items-center justify-center transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed shadow-sm border',
                          parsedUserDetails?.responseMode === 'web-search'
                            ? 'border-green-300 bg-green-50 hover:bg-green-100 dark:border-green-600 dark:bg-green-900/30 dark:hover:bg-green-900/50 text-green-700 dark:text-green-300'
                            : parsedUserDetails?.responseMode === 'deep-search'
                              ? 'border-orange-300 bg-orange-50 hover:bg-orange-100 dark:border-orange-600 dark:bg-orange-900/30 dark:hover:bg-orange-900/50 text-orange-700 dark:text-orange-300'
                              : 'border-blue-300 bg-blue-50 hover:bg-blue-100 dark:border-blue-600 dark:bg-blue-900/30 dark:hover:bg-blue-900/50 text-blue-700 dark:text-blue-300'
                        ]" :title="modeOptions[parsedUserDetails.responseMode || ''].title">
                        <!-- Dynamic icon based on selected mode -->
                        <i :class="modeOptions[parsedUserDetails.responseMode || ''].icon" class="text-xs sm:text-sm"></i>
                      </button>

                      <!-- Dropdown Menu -->
                      <div v-show="showInputModeDropdown"
                        class="absolute bottom-12 left-0 bg-white dark:bg-gray-800 border-[1px] border-gray-300 dark:border-gray-600 rounded-lg shadow-2xl pt-2 z-[100] w-[220px] sm:w-[240px]"
                        @click.stop>
                        <div
                          class="px-2 py-1 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide border-b border-gray-200 dark:border-gray-700">
                          Response Mode
                        </div>

                        <!-- Mode Options -->
                        <button v-for="(option, key) in modeOptions" :key="option.mode" type="button"
                          @click="selectInputMode(option.mode)" :class="[
                            'w-full px-3 py-2.5 text-left text-sm flex items-center gap-3 transition-colors',
                            parsedUserDetails?.responseMode === option.mode
                              ? 'bg-green-100 dark:bg-green-900/40 text-green-700 dark:text-green-300 border-r-2 border-green-500'
                              : 'hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-800 dark:text-gray-200'
                          ]">
                          <i :class="[option.icon,
                          parsedUserDetails?.responseMode === option.mode
                            ? ' text-green-600 dark:text-green-400'
                            : ' text-gray-600 dark:text-gray-400'
                          ]"></i>
                          <div class="flex-1 min-w-0">
                            <div class="font-semibold">{{ option.label }}</div>
                            <div class="text-xs opacity-70">{{ option.description }}</div>
                          </div>
                          <i v-if="parsedUserDetails?.responseMode === option.mode"
                            class="pi pi-check text-green-600 dark:text-green-400 text-sm font-bold"></i>
                        </button>
                      </div>
                    </div>
                  </div>

                  <!-- Submit Button - Right side -->
                  <button type="submit" :disabled="inputDisabled"
                    class="rounded-lg w-8 h-8 sm:w-9 sm:h-9 flex items-center justify-center transition-colors text-white bg-blue-500 hover:bg-blue-600 dark:bg-blue-600 dark:hover:bg-blue-800 disabled:cursor-not-allowed disabled:opacity-50 disabled:bg-gray-400 flex-shrink-0 shadow-sm">
                    <i v-if="!isLoading" class="pi pi-arrow-up text-xs sm:text-sm"></i>
                    <i v-else class="pi pi-spin pi-spinner text-xs sm:text-sm"></i>
                  </button>
                </div>
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
