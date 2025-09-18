<script lang="ts" setup>
import { ref, onMounted, nextTick } from "vue"
import { marked } from "marked"
import hljs from "highlight.js"
import "highlight.js/styles/github-dark.css"
import SideNav from "../components/SideNav.vue"
import TopNav from "../components/TopNav.vue"

type Res = {
  response: string,
  prompt?: string,
  status?: number,
}

type LinkPreview = {
  url: string,
  title?: string,
  description?: string,
  image?: string,
  domain?: string,
  loading?: boolean,
  error?: boolean
}

// ---------- State ----------
let showInput = ref(false)
let screenWidth = ref(screen.width)
// local state for collapse toggle
const isCollapsed = ref(false)
const isSidebarHidden = ref(true)

let userDetails: any = localStorage.getItem("userdetails")
let parsedUserDetails: any = userDetails ? JSON.parse(userDetails) : null

let chats: any = localStorage.getItem("chats")
let parsedChats: any = JSON.parse(chats) === null ? [] : JSON.parse(chats)
let res = ref<Res[]>(parsedChats.length === 0 ? [] : parsedChats.map((item: any) => ({
  prompt: item.prompt ?? "",
  response: item.response,
  status: item.status
})))

let isLoading = ref(false)
let expanded = ref<boolean[]>(res.value.map(() => false))

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
function handleAuth(e: Event) {
  e.preventDefault()
  const form = e.target as HTMLFormElement
  const formData = new FormData(form)
  
  const email = formData.get('email') as string
  const password = formData.get('password') as string
  const createdAt = new Date().toISOString()
  
  if (!email || !password) {
    alert('Please fill in all fields')
    return
  }
  
  const username = formData.get('username') as string
  
  const userData = {
    email,
    username,
    createdAt,
    sessionId: btoa(email + ':' + password + ':' + createdAt)
  }
  
  try {
    localStorage.setItem('userdetails', JSON.stringify(userData))
    parsedUserDetails = userData
    
    if (chats && JSON.parse(chats).length > 0) {
      res.value = JSON.parse(chats)
    } else {
      localStorage.setItem('chats', JSON.stringify([]))
      res.value = []
    }
    
    alert(`Welcome ${username}! Your session has been created.`)
    
    nextTick(() => {
      const textarea = document.getElementById("prompt") as HTMLTextAreaElement
      if (textarea) textarea.focus()
    })
  } catch (err) {
    console.error('Failed to save user data:', err)
    alert('Failed to create session. Please try again.')
  }
}

function logout() {
  if (confirm('Are you sure you want to logout? This will clear your session on this device.')) {
    try {
      localStorage.removeItem('userdetails')
      localStorage.removeItem('isCollapsed')
      // Keep link previews cached even after logout
      // localStorage.removeItem('linkPreviews') // Commented out to persist previews
      
      parsedUserDetails = null
      res.value = []
      expanded.value = []
      showInput.value = false
      isCollapsed.value = false
      
      alert('Logged out successfully!')
    } catch (err) {
      console.error('Error during logout:', err)
      alert('Error during logout. Please try again.')
    }
  }
}

function isAuthenticated(): boolean {
  return parsedUserDetails && parsedUserDetails.email && parsedUserDetails.username && parsedUserDetails.sessionId
}

// ---------- Helpers ----------
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
      alert('Failed to copy text. Please try again.')
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

// Enhanced markdown renderer with link previews
async function renderMarkdownWithPreviews(text?: string) {
  if (!text || typeof text !== "string") return ""
  
  try {
    // First, render the markdown
    let html = marked.parse(text)
    
    // Extract URLs from the original text for preview generation
    const urls = extractUrls(text)
    
    // Generate previews for found URLs
    if (urls.length > 0) {
      const previews = await Promise.all(
        urls.slice(0, 3).map(url => fetchLinkPreview(url)) // Limit to 3 previews max
      )
      
      // Append previews to the HTML
      const previewsHtml = previews
        .filter(preview => !preview.loading)
        .map(preview => LinkPreviewComponent({ preview }))
        .join('')
      
      if (previewsHtml) {
        html += `<div class="link-previews mt-3">${previewsHtml}</div>`
      }
    }
    
    return html
  } catch (err) {
    console.error("Markdown parse error:", err)
    return text
  }
}

function renderMarkdown(text?: string) {
  if (!text || typeof text !== "string") return ""
  try {
    return marked.parse(text)
  } catch (err) {
    console.error("Markdown parse error:", err)
    return text
  }
}

// --- Handle submit with optimistic update ---
async function handleSubmit(e?: any, retryPrompt?: string) {
  e?.preventDefault?.()

  let promptValue = retryPrompt || e?.target?.prompt?.value?.trim()
  if (!promptValue || isLoading.value) return

  if (!isAuthenticated()) {
    alert('Please create a session first')
    return
  }

  isLoading.value = true

  if (!retryPrompt && e?.target?.prompt) {
    e.target.prompt.value = ""
    e.target.prompt.style.height = "auto"
  }

  const tempResp: Res = { prompt: promptValue, response: "..." }
  res.value.push(tempResp)
  expanded.value.push(false)
  
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
    
    res.value[res.value.length - 1] = {
      prompt: promptValue,
      response: parseRes.error ? parseRes.error : parseRes.response,
      status: response.status
    }

    // Trigger link preview generation for the new response
    await processLinksInResponse(res.value.length - 1)
    
  } catch (err: any) {
    res.value[res.value.length - 1] = {
      prompt: promptValue,
      response: `⚠️ Error: ${err.message || 'Failed to get response. Please try again.'}`
    }
  } finally {
    isLoading.value = false
    debounceSave()
    await nextTick()
    scrollToBottom()
  }
}

// Process links in a response and generate previews
async function processLinksInResponse(index: number) {
  const response = res.value[index]
  if (!response.response || response.response === "...") return

  const urls = extractUrls(response.response)
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

// --- Debounced localStorage save ---
let saveTimeout: any
function debounceSave() {
  clearTimeout(saveTimeout)
  saveTimeout = setTimeout(() => {
    try {
      localStorage.setItem("chats", JSON.stringify(res.value))
    } catch (err) {
      console.error("Failed to save to localStorage:", err)
    }
  }, 300)
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
function copyResponse(text: string) {
  copyCode(text)
}

function toggleSidebar() {
  isCollapsed.value = !isCollapsed.value
  localStorage.setItem("isCollapsed", String(isCollapsed.value))
}

function shareResponse(text: string, prompt?: string) {
  if (navigator.share) {
    navigator.share({
      title: prompt && prompt.length > 200 ? `${prompt.slice(0,200)}...\n\n` : `${prompt || "Gemmie Chat"}\n\n`,
      text
    }).catch(err => console.log("Share canceled", err))
  } else {
    copyCode(text)
    alert("Sharing not supported, copied to clipboard instead!")
  }
}

function refreshResponse(prompt?: string) {
  if (prompt && !isLoading.value) {
    handleSubmit(undefined, prompt)
  }
}

function deleteChat(index: number) {
  if (isLoading.value) return
  
  res.value.splice(index, 1)
  expanded.value.splice(index, 1)
  
  try {
    localStorage.setItem("chats", JSON.stringify(res.value))
  } catch (err) {
    console.error("Failed to save after deletion:", err)
  }
}

// ---------- UI Helpers ----------
function setShowInput() {
  if (res.value.length !== 0) {
    return
  }
  if (!isAuthenticated()) {
    alert('Please create a session first')
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
    if(sideNav.classList.contains("none")){
      sideNav.classList.add("w-full", "bg-white", "z-20", "fixed", "top-0", "left-0", "bottom-0", "border-r-[1px]", "flex", "flex-col", "transition-all", "duration-300", "ease-in-out")
    }else{
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
      preventDefault: () => {}, 
      target: { prompt: textarea }
    }
    handleSubmit(formEvent)
  }
}

function clearAllChats() {
  if (isLoading.value) return
  
  if (confirm("Are you sure you want to clear all chats? This action cannot be undone.")) {
    res.value = []
    expanded.value = []
    try {
      localStorage.setItem("chats", JSON.stringify([]))
      // Option to clear link previews as well (uncomment if desired)
      // localStorage.removeItem('linkPreviews')
      // linkPreviewCache.value.clear()
    } catch (err) {
      console.error("Failed to clear chats:", err)
    }
  }
}

// Add function to manually clear link preview cache
function clearLinkPreviewCache() {
  if (confirm("Clear all link preview cache? This will require refetching previews for existing links.")) {
    try {
      localStorage.removeItem('linkPreviews')
      linkPreviewCache.value.clear()
      alert('Link preview cache cleared successfully!')
    } catch (err) {
      console.error('Failed to clear link preview cache:', err)
      alert('Failed to clear link preview cache.')
    }
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

  // Load cached link previews
  loadLinkPreviewCache()

  // Pre-process existing chat links on page load
  if (res.value.length > 0) {
    res.value.forEach((item, index) => {
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

  scrollToBottom()
  
  document.addEventListener("click", (e: any) => {
    if (e.target && e.target.classList.contains("copy-button")) {
      const code = decodeURIComponent(e.target.getAttribute("data-code"))
      copyCode(code, e.target)
    }
  })

  if (showInput.value || res.value.length > 0) {
    nextTick(() => {
      const textarea = document.getElementById("prompt") as HTMLTextAreaElement
      if (textarea) textarea.focus()
    })
  }
})
</script>

<template>
  <div class="flex h-[100vh]">
    <!-- Sidebar -->
    <SideNav v-if="isAuthenticated()" :data="{ res, parsedUserDetails, screenWidth, isCollapsed }" 
             :functions="{ setShowInput, hideSidebar, clearAllChats, toggleSidebar, logout }" />

    <!-- Main Chat Window -->
    <div :class="screenWidth>720&&isAuthenticated() ? (!isCollapsed?
      'flex-grow flex flex-col items-center justify-center ml-[270px] font-light text-sm transition-all duration-300 ease-in-out' 
      :
      'flex-grow flex flex-col items-center justify-center ml-[60px] font-light text-sm transition-all duration-300 ease-in-out' 
    )
    : 'text-sm font-light flex-grow items-center justify-center flex flex-col transition-all duration-300 ease-in-out'">
      <TopNav v-if="isAuthenticated()" :data="{ res, parsedUserDetails, screenWidth, isCollapsed, isSidebarHidden }" 
        :functions="{ hideSidebar, clearAllChats }"
      />

      <div :class="(screenWidth>720&&isAuthenticated()) ? 'h-screen flex flex-col items-center justify-center w-[85%]':'h-screen flex flex-col items-center justify-center'">
        <!-- Empty State -->
        <div v-if="res.length===0||!isAuthenticated()" class="flex flex-col items-center justify-center h-[90vh]">
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
              <button
                v-if="isAuthenticated()"
                @click="setShowInput"
                class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-500 transition-colors">
                Write a prompt
              </button>
            </div>

            <div v-if="!isAuthenticated()" class="flex-grow text-sm max-w-md">
              <p class="text-lg text-gray-500 mb-4">Create a session on this device to get started</p>
              <form @submit="handleAuth" class="mt-3 flex-col flex gap-2">
                <input
                  required
                  type="text"
                  name="username"
                  placeholder="johndoe"
                  class="border border-gray-300 rounded-lg px-3 py-2 w-full focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
                <input
                  required
                  type="email"
                  name="email"
                  placeholder="johndoe@example.com"
                  class="border border-gray-300 rounded-lg px-3 py-2 w-full focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
                <input
                  required
                  type="password"
                  name="password"
                  placeholder="Password"
                  minlength="6"
                  class="border border-gray-300 rounded-lg px-3 py-2 w-full focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
                <button
                  type="submit"
                  class="mt-3 w-full bg-blue-600 text-white rounded-lg px-4 py-2 hover:bg-blue-500 transition-colors"
                >
                  Create Session
                </button>
              </form>
              <p class="text-sm text-gray-400 mt-2">
                Your credentials are only stored locally on your device for session management.
              </p>
            </div>
          </div>
          <div>
            <p class="text-sm mt-2 text-gray-400">Gemmie can make mistakes. Check important info.</p>
          </div>
        </div>

        <!-- Chat Messages -->
        <div v-else-if="res.length!==0&&isAuthenticated()" class="flex-grow no-scrollbar overflow-y-auto px-4 space-y-4 pt-[90px] pb-[120px]">
          <div v-for="(item, i) in res" :key="`chat-${i}`" class="flex flex-col gap-2">
            
            <!-- User Bubble -->
            <div class="flex justify-end">
              <div :class="screenWidth>720 ? 'max-w-[70%]' : 'max-w-[95%]'"
                   class="bg-gray-50 text-black p-3 rounded-2xl prose prose-sm max-w-none chat-bubble">
                <p class="text-xs opacity-80 text-right mb-1">{{ parsedUserDetails?.username || "You" }}</p>
                <div v-html="renderMarkdown(item.prompt || '')"></div>
                
                <!-- Link Previews Section for User Messages -->
                <div v-if="extractUrls(item.prompt || '').length > 0" class="mt-3">
                  <div v-for="url in extractUrls(item.prompt || '').slice(0, 3)" :key="`user-${i}-${url}`">
                    <div v-if="linkPreviewCache.get(url)" v-html="LinkPreviewComponent({ preview: linkPreviewCache.get(url)! })"></div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Bot Bubble -->
            <div class="flex justify-start relative">
              <div :class="screenWidth>720 ? 'max-w-[70%]' : 'max-w-[95%]'"
                   class="bg-none leading-relaxed text-black p-3 rounded-2xl prose prose-sm max-w-none">
                
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
                      <div v-if="linkPreviewCache.get(url)" v-html="LinkPreviewComponent({ preview: linkPreviewCache.get(url)! })"></div>
                    </div>
                  </div>
                </div>

                <!-- Actions (hidden during loading) -->
                <div v-if="item.response !== '...'" class="flex gap-3 mt-2 text-gray-500 text-sm">
                  <button @click="copyResponse(item.response)" 
                          class="flex items-center gap-1 hover:text-blue-600 transition-colors">
                    <i class="pi pi-copy"></i> Copy
                  </button>
                  <button @click="shareResponse(item.response, item.prompt)" 
                          class="flex items-center gap-1 hover:text-green-600 transition-colors">
                    <i class="pi pi-share-alt"></i> Share
                  </button>
                  <button @click="refreshResponse(item.prompt)" 
                          :disabled="isLoading"
                          class="flex items-center gap-1 hover:text-orange-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
                    <i class="pi pi-refresh"></i> Refresh
                  </button>
                  <button @click="deleteChat(i)" 
                          :disabled="isLoading"
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
        <div
          v-if="(res.length !== 0 || showInput === true)&&isAuthenticated()"
          :style="screenWidth>720&&!isCollapsed ? 'left:270px;' :
            screenWidth>720&&isCollapsed? 'left:60px;': 'left:0px;'"
          class="bg-white bottom-0 right-0 fixed pb-5 px-5"
        >
          <div class="flex items-center justify-center w-full">
            <form @submit="handleSubmit" :class="screenWidth > 720 ?'relative flex px-3 py-2 border-2 shadow rounded-2xl items-center gap-2 w-[85%]':'relative flex px-3 py-2 border-2 shadow rounded-2xl w-full items-center gap-2'">
              <textarea
                required
                id="prompt"
                name="prompt"
                @keydown="onEnter"
                @input="autoGrow"
                :disabled="isLoading"
                rows="1"
                class="flex-grow py-2 bg-white text-sm 
                      outline-none resize-none border-none
                      max-h-[200px] overflow-auto leading-relaxed
                      disabled:opacity-50 disabled:cursor-not-allowed"
                :placeholder="isLoading ? 'Please wait...' : 'Ask me a question...'"
              ></textarea>
              <button
                type="submit"
                :disabled="isLoading"
                class="rounded-lg w-[26px] h-[26px] flex items-center justify-center transition-colors
                      text-white bg-blue-600 hover:bg-blue-500 disabled:cursor-not-allowed disabled:opacity-50
                      disabled:bg-gray-400 flex-shrink-0"
              >
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