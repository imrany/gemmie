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
  
  // Extract username from email (part before @)
  // const username = email.split('@')[0]
  const username = formData.get('username') as string
  
  // Create user session data
  const userData = {
    email,
    username,
    createdAt,
    sessionId: btoa(email + ':' + password + ':' + createdAt)
  }
  
  try {
    // Save user details to localStorage
    localStorage.setItem('userdetails', JSON.stringify(userData))
    // Update reactive state
    parsedUserDetails = userData
    
    // Initialize chats for new user or use existing chats
    if (chats && JSON.parse(chats).length > 0) {
      res.value = JSON.parse(chats)
    } else {
      localStorage.setItem('chats', JSON.stringify([]))
      res.value = []
    }
    
    // Show success message
    alert(`Welcome ${username}! Your session has been created.`)
    
    // Auto-focus on input after successful auth
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
      // localStorage.removeItem('chats')
      localStorage.removeItem('isCollapsed')
      
      // Reset all reactive state
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

marked.use({
  renderer: {
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

// --- Handle submit with optimistic update ---
async function handleSubmit(e?: any, retryPrompt?: string) {
  e?.preventDefault?.()

  let promptValue = retryPrompt || e?.target?.prompt?.value?.trim()
  if (!promptValue || isLoading.value) return

  // Check authentication before submitting
  if (!isAuthenticated()) {
    alert('Please create a session first')
    return
  }

  // Set loading state
  isLoading.value = true

  // Clear textarea instantly for snappy UX
  if (!retryPrompt && e?.target?.prompt) {
    e.target.prompt.value = ""
    // Reset textarea height
    e.target.prompt.style.height = "auto"
  }

  // Optimistically push user prompt + loading bubble
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
    
    // Replace last "loading" bubble with real response
    res.value[res.value.length - 1] = {
      prompt: promptValue,
      response: parseRes.error ? parseRes.error : parseRes.response,
      status: response.status
    }
  } catch (err: any) {
    // Replace loading bubble with error message
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

// --- Smarter textarea auto-grow with max height ---
function autoGrow(e: Event) {
  const el = e.target as HTMLTextAreaElement
  const maxHeight = 200 // px - adjust as needed
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

function shareResponse(text: string) {
  if (navigator.share) {
    navigator.share({
      title: `Gemmie Chat`,
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
  if (isLoading.value) return // Prevent deletion while loading
  
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
  // Focus the textarea after it's shown
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

// Debounced resize handler
let resizeTimeout: any
window.onresize = () => {
  clearTimeout(resizeTimeout)
  resizeTimeout = setTimeout(() => {
    screenWidth.value = screen.width
  }, 100)
}

function onEnter(e: KeyboardEvent) {
  if (e.key !== 'Enter' || e.shiftKey || isLoading.value) {
    return // let Shift+Enter create a newline or ignore if not Enter key
  }
  
  e.preventDefault()
  
  // Create a fake form event for handleSubmit
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
    } catch (err) {
      console.error("Failed to clear chats:", err)
    }
  }
}

onMounted(() => {
  const saved = localStorage.getItem("isCollapsed")
  if (saved && saved !== 'null') {
    try {
      isCollapsed.value = JSON.parse(saved)
    } catch (err) {
      console.error('Error parsing isCollapsed:', err)
    }
  }

  scrollToBottom()
  
  // Handle copy button clicks
  document.addEventListener("click", (e: any) => {
    if (e.target && e.target.classList.contains("copy-button")) {
      const code = decodeURIComponent(e.target.getAttribute("data-code"))
      copyCode(code, e.target)
    }
  })

  // Focus input if it's visible
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
                <p>
                  Your private AI assistant. 
                </p>
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
                   class="bg-gray-50 text-black p-3 rounded-2xl prose prose-sm max-w-none">
                <p class="text-xs opacity-80 text-right mb-1">{{ parsedUserDetails?.username || "You" }}</p>
                <p class="text-wrap whitespace-pre-wrap">{{ item.prompt }}</p>
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
                
                <!-- Regular response -->
                <div v-else v-html="renderMarkdown(item.response || '')"></div>

                <!-- Actions (hidden during loading) -->
                <div v-if="item.response !== '...'" class="flex gap-3 mt-2 text-gray-500 text-sm">
                  <button @click="copyResponse(item.response)" 
                          class="flex items-center gap-1 hover:text-blue-600 transition-colors">
                    <i class="pi pi-copy"></i> Copy
                  </button>
                  <button @click="shareResponse(item.response)" 
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