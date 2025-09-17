<script lang="ts" setup>
import { ref, onMounted, nextTick } from "vue"
import { marked } from "marked"
import hljs from "highlight.js"
import "highlight.js/styles/github-dark.css" // pick any highlight.js theme
import SideNav from "../components/SideNav.vue"
import TopNav from "../components/TopNav.vue"

type Res = {
  response: string,
  prompt?: string,
  status?: number,
}

let showInput = ref(false)
let note = ref("You are using Gemmie, your personal AI assistant")
let screenWidth = ref(screen.width)

let userDetails: any = localStorage.getItem("userdetails")
let parsedUserDetails: any = JSON.parse(userDetails) === null ? [] : JSON.parse(userDetails)

let chats: any = localStorage.getItem("chats")
let parsedChats: any = JSON.parse(chats) === null ? [] : JSON.parse(chats)
let res: any = parsedChats.length === 0 ? [] : parsedChats
let isLoading = ref(false)

// track expanded state per index
let expanded = ref<boolean[]>(res.map(() => false))

// ✅ Configure marked with highlight.js
marked.use({
  renderer: {
    code({ text, lang }) {
      if (lang && hljs.getLanguage(lang)) {
        return `<pre><code class="hljs language-${lang}">${hljs.highlight(text, { language: lang }).value}</code></pre>`
      }
      return `<pre><code class="hljs">${hljs.highlightAuto(text).value}</code></pre>`
    }
  }
})


async function handleSubmit(e: any) {
  try {
    e.preventDefault()
    isLoading.value = true
    let url = `https://wrapper.villebiz.com/v1/genai`

    let response = await fetch(url, {
      method: "POST",
      body: JSON.stringify(e.target.prompt.value),
      headers: {
        "content-type": "application/json"
      }
    })

    let parseRes = await response.json()
    let resp: Res = {
      prompt: parseRes.prompt,
      response: parseRes.error ? parseRes.error : parseRes.response,
      status: response.status
    }

    res.push(resp)
    expanded.value.push(false)

    localStorage.setItem("chats", JSON.stringify(res))
    isLoading.value = false
    e.target.reset()
    await nextTick()
    scrollToBottom()
  } catch (error: any) {
    let resp: Res = {
      prompt: e.target.prompt.value,
      response: error.message
    }
    res.push(resp)
    expanded.value.push(false)
    localStorage.setItem("chats", JSON.stringify(res))
    isLoading.value = false
    console.log(error.message)
  }
}

function setShowInput() {
  showInput.value = true
}

function toggleExpand(index: number) {
  expanded.value[index] = !expanded.value[index]
  nextTick(() => {
    scrollToBottom()
  })
}

function scrollToBottom() {
  let elem = document.getElementById("scrollableElem")
  if (elem) elem.scrollIntoView({ behavior: "smooth", block: "end" })
}

function toggleSideNav() {
  let sideNav: any = document.getElementById("side_nav")
  sideNav.classList.contains("none") ? sideNav.classList.remove("none") : sideNav.classList.add("none")
}

window.onresize = () => {
  screenWidth.value = screen.width
}

onMounted(() => {
  scrollToBottom()
})

function renderMarkdown(text: string) {
  return marked.parse(text)
}
</script>

<template>
  <div class="flex h-[100vh]">
    <!-- Sidebar -->
    <SideNav :data="{ res, parsedUserDetails, screenWidth }" :functions="{ setShowInput, toggleSideNav }" />

    <!-- Main Chat Window -->
    <div :class="screenWidth>720 ? 'flex-grow flex flex-col ml-[300px]' : 'flex-grow flex flex-col'">
      <TopNav :data="{ note, res, parsedUserDetails, screenWidth }" />

      <div class="h-screen flex flex-col">
        <!-- Empty State -->
        <div v-if="res.length===0" class="flex flex-col items-center justify-center h-full">
          <div class="flex flex-col items-center gap-3 text-gray-600">
            <div class="rounded-full bg-gray-200 w-[60px] h-[60px] flex justify-center items-center">
              <span class="pi pi-comment text-lg"></span>
            </div>
            <p class="text-xl font-semibold">{{ parsedUserDetails.username || 'Gemmie' }}</p>
            <p class="text-center text-sm max-w-[300px]">
              Here's Gemmie your AI assistant. Ask me a question
            </p>
            <button v-if="parsedUserDetails.username!==undefined"
                    @click="setShowInput"
                    class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-500">
              Write a prompt
            </button>
            <button v-if="parsedUserDetails.username===undefined && screenWidth<720"
                    @click="toggleSideNav"
                    class="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-500">
              Get Started
            </button>
          </div>
        </div>

        <!-- Chat Messages -->
        <div v-else class="flex-grow overflow-y-auto px-4 space-y-3 pt-[90px] py-[70px]">
          <div v-for="(item, i) in res" :key="i" class="flex flex-col gap-2">
            <!-- User Bubble -->
            <div class="flex justify-end">
              <div :class="screenWidth>720 ? 'max-w-[70%]' : 'max-w-[95%]'"
                   class="bg-blue-500 text-white p-3 rounded-2xl shadow">
                <p class="text-xs opacity-80">{{ parsedUserDetails.username || "You" }}</p>
                <p>{{ item.prompt }}</p>
              </div>
            </div>

            <!-- Bot Bubble -->
            <div class="flex justify-start">
              <div :class="screenWidth>720 ? 'max-w-[70%]' : 'max-w-[95%]'"
                   class="bg-gray-100 text-black p-3 rounded-2xl shadow prose prose-sm max-w-none">
                <p class="text-xs font-semibold text-gray-500 mb-1">Gemmie</p>

                <div v-if="item.response.length>200 && !expanded[i]">
                  <div class="less" v-html="renderMarkdown(item.response.slice(0,200))"></div>
                  <button @click="toggleExpand(i)" class="text-blue-500 mt-2 text-sm">Show more</button>
                </div>

                <div v-else>
                  <div v-html="renderMarkdown(item.response)"></div>
                  <button v-if="item.response.length>200"
                          @click="toggleExpand(i)"
                          class="text-blue-500 mt-2 text-sm">Show less</button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Input -->
        <div v-if="res.length!==0 || showInput===true"
             :style="screenWidth>720?'left:300px;':'left:0;'"
             class="bg-white bottom-0 right-0 fixed border-t p-2">
          <form @submit="handleSubmit" class="flex gap-2 w-full">
            <input required id="prompt" name="prompt" type="text"
                   class="flex-grow px-3 py-2 rounded-full bg-gray-200 text-sm focus:ring-2 focus:ring-blue-400 outline-none"
                   placeholder="Ask me a question..." />
            <button :disabled="isLoading"
                    :class="isLoading ? 'cursor-progress' : 'hover:bg-blue-100'"
                    class="rounded-full p-2 flex items-center justify-center">
              <i v-if="!isLoading" class="pi pi-send text-blue-600"></i>
              <i v-else class="pi pi-spin pi-spinner text-gray-500"></i>
            </button>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>
