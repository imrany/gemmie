<script lang="ts" setup>
import { computed, inject, ref, type Ref, onMounted, onUnmounted } from 'vue'
import type { Chat } from '@/types'

const globalState = inject('globalState') as {
  isAuthenticated: Ref<boolean>
}
const { isAuthenticated } = globalState

// Carousel state
const currentSlide = ref(0)
const autoSlideInterval = ref<number | null>(null)

// Sample chat suggestions - you can customize these
const chatSuggestions = ref([
  {
    title: "Creative Writing",
    description: "Get help with stories, poems, and creative content",
    icon: "pi-pencil",
    color: "from-purple-500 to-pink-500",
    prompt: "Help me write a creative story about..."
  },
  {
    title: "Code Assistant",
    description: "Debug code, learn programming concepts, and get help",
    icon: "pi-code",
    color: "from-blue-500 to-cyan-500",
    prompt: "Can you help me debug this code?"
  },
  {
    title: "Learning & Research",
    description: "Explore topics, get explanations, and expand knowledge",
    icon: "pi-book",
    color: "from-green-500 to-teal-500",
    prompt: "Explain this topic in simple terms..."
  },
  {
    title: "Problem Solving",
    description: "Work through challenges and find solutions",
    icon: "pi-sun",
    color: "from-yellow-500 to-orange-500",
    prompt: "I need help solving this problem..."
  },
  {
    title: "Data Analysis",
    description: "Analyze data, create charts, and find insights",
    icon: "pi-chart-line",
    color: "from-indigo-500 to-purple-500",
    prompt: "Help me analyze this data..."
  },
  {
    title: "Business Strategy",
    description: "Strategic planning, market analysis, and business advice",
    icon: "pi-briefcase",
    color: "from-orange-500 to-red-500",
    prompt: "I need business advice on..."
  }
])

const features = ref([
  {
    title: "Privacy First",
    description: "Your conversations stay private and secure with end-to-end encryption",
    icon: "pi-shield",
    color: "text-green-500"
  },
  {
    title: "Local Storage",
    description: "All data is stored locally on your device for maximum privacy",
    icon: "pi-database",
    color: "text-blue-500"
  },
  {
    title: "Cross-Device Sync",
    description: "Seamlessly access your chats from any device, anywhere",
    icon: "pi-sync",
    color: "text-purple-500"
  },
  {
    title: "Always Available",
    description: "24/7 AI assistance at your fingertips, even offline",
    icon: "pi-clock",
    color: "text-indigo-500"
  },
  {
    title: "Fast Response",
    description: "Lightning-fast AI responses powered by advanced algorithms",
    icon: "pi-bolt",
    color: "text-yellow-500"
  },
  {
    title: "Smart Memory",
    description: "Remembers context across conversations for better assistance",
    icon: "pi-book",
    color: "text-pink-500"
  }
])

const tips = ref([
  {
    title: "Be Specific",
    description: "The more details you provide, the better I can help you",
    icon: "pi-pencil",
    example: "Instead of 'help with code', try 'debug my React component that won't render'"
  },
  {
    title: "Ask Follow-ups",
    description: "Don't hesitate to ask for clarification or more details",
    icon: "pi-comments",
    example: "Can you explain that in simpler terms? or What about edge cases?"
  },
  {
    title: "Use Examples",
    description: "Provide examples of what you're working with",
    icon: "pi-list",
    example: "Here's my current code... or This is the error I'm getting..."
  },
  {
    title: "Set Context",
    description: "Let me know your skill level and what you're trying to achieve",
    icon: "pi-map",
    example: "I'm a beginner in Python and want to build a web scraper"
  }
])

const carouselSlides = computed(() => [
  { id: 'welcome', title: 'Welcome', icon: 'pi-home' },
  { id: 'suggestions', title: 'Chat Ideas', icon: 'pi-comment' },
  { id: 'features', title: 'Features', icon: 'pi-star' },
  { id: 'tips', title: 'Tips', icon: 'pi-info-circle' }
])

// --- All Props Directly from HomeView ---
const props = defineProps<{
  chats: Chat[],
  currentChatId: string
  isCollapsed?: boolean
  parsedUserDetails: {
    username: string
  } | null
  screenWidth: number
  syncStatus: {
    lastSync: Date | null
    syncing: boolean
    hasUnsyncedChanges: boolean
  }
  isLoading: boolean
  authStep: number
  showCreateSession: boolean
  authData: {
    username: string
    email: string
    password: string
  }
  currentMessages: any[]
  validateCurrentStep: boolean
  setShowInput: () => void
  hideSidebar: () => void
  clearAllChats: () => void
  toggleSidebar: () => void
  logout: () => void
  createNewChat: () => void
  switchToChat: (chatId: string) => void
  deleteChat: (chatId: string) => void
  renameChat: (chatId: string, newTitle: string) => void
  manualSync: () => void
  handleStepSubmit: (e: Event) => void
  prevAuthStep: () => void
  updateAuthData: (data: Partial<{ username: string; email: string; password: string }>) => void
  setShowCreateSession: (value: boolean) => void
}>()

// Carousel functions
const nextSlide = () => {
  currentSlide.value = (currentSlide.value + 1) % carouselSlides.value.length
}

const prevSlide = () => {
  currentSlide.value = currentSlide.value === 0 ? carouselSlides.value.length - 1 : currentSlide.value - 1
}

const goToSlide = (index: number) => {
  currentSlide.value = index
}

const startAutoSlide = () => {
  if (autoSlideInterval.value) return
  autoSlideInterval.value = window.setInterval(nextSlide, 6000) // 6 seconds
}

const stopAutoSlide = () => {
  if (autoSlideInterval.value) {
    clearInterval(autoSlideInterval.value)
    autoSlideInterval.value = null
  }
}

const handleSuggestionClick = (suggestion: typeof chatSuggestions.value[0]) => {
  props.createNewChat()
  // You might want to emit an event here to set the initial message
  // emit('setInitialMessage', suggestion.prompt)
}

// Handle input updates
const handleUsernameInput = (event: Event) => {
  const value = (event.target as HTMLInputElement).value.trim()
  props.updateAuthData({ username: value })
}

const handleEmailInput = (event: Event) => {
  const value = (event.target as HTMLInputElement).value.trim()
  props.updateAuthData({ email: value })
}

const handlePasswordInput = (event: Event) => {
  const value = (event.target as HTMLInputElement).value.trim()
  props.updateAuthData({ password: value })
}

// Layout logic
const isDesktop = computed(() => props.screenWidth >= 720)
const isMobile = computed(() => props.screenWidth < 720)

// Lifecycle
onMounted(() => {
  startAutoSlide()
})

onUnmounted(() => {
  stopAutoSlide()
})
</script>

<template>
  <div class="flex flex-col items-center justify-center min-h-screen w-screen">
    <!-- Desktop Layout: Side by side -->
    <div v-if="isDesktop" class="flex gap-10 items-center justify-center h-full w-full max-w-7xl mx-auto px-8">
      <!-- LEFT SECTION: Carousel -->
      <div class="flex-1 max-w-2xl" @mouseenter="stopAutoSlide" @mouseleave="startAutoSlide">
        <!-- Carousel Container -->
        <div class="relative h-[600px] overflow-hidden backdrop-blur-sm rounded-3xl bg-white/80 border border-white/50">
          <!-- Slide 1: Welcome -->
          <div
            :class="currentSlide === 0 ? 'translate-x-0 opacity-100' : currentSlide > 0 ? '-translate-x-full opacity-0' : 'translate-x-full opacity-0'"
            class="absolute inset-0 transition-all duration-700 ease-in-out transform p-12 flex flex-col items-center justify-center">
            <div class="text-center">
              <img src="/logo.svg" alt="Gemmie Logo" class="rounded-full w-24 h-24 mb-8 mx-auto" />
              <h2 class="text-4xl font-bold text-gray-900 mb-4">
                Welcome to Gemmie
              </h2>
              <p class="text-gray-600 leading-relaxed mb-3 max-w-lg text-base">
                Experience privacy-first conversations with advanced AI. Your data stays secure, local, and synced
                across all your devices.
              </p>
              <div class="flex items-center justify-center gap-12">
                <div class="flex flex-col items-center gap-3">
                  <div class="w-10 h-10 bg-green-100 rounded-full flex items-center justify-center">
                    <i class="pi pi-shield text-green-600"></i>
                  </div>
                  <span class="text-gray-700 font-medium">Private</span>
                </div>
                <div class="flex flex-col items-center gap-3">
                  <div class="w-10 h-10 bg-blue-100 rounded-full flex items-center justify-center">
                    <i class="pi pi-database text-blue-600 "></i>
                  </div>
                  <span class="text-gray-700 font-medium">Local</span>
                </div>
                <div class="flex flex-col items-center gap-3">
                  <div class="w-10 h-10 bg-purple-100 rounded-full flex items-center justify-center">
                    <i class="pi pi-sync text-purple-600"></i>
                  </div>
                  <span class="text-gray-700 font-medium">Synced</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Slide 2: Chat Suggestions -->
          <div
            :class="currentSlide === 1 ? 'translate-x-0 opacity-100' : currentSlide > 1 ? '-translate-x-full opacity-0' : 'translate-x-full opacity-0'"
            class="absolute inset-0 transition-all duration-700 ease-in-out transform p-8">
            <div class="text-center mb-2">
              <h2 class="text-3xl font-bold text-gray-900 mb-2">What can I help with?</h2>
              <p class="text-gray-600 text-lg">Try one of these popular conversation starters</p>
            </div>

            <div class="grid grid-cols-2 gap-6 h-96 overflow-x-hidden overflow-y-auto custom-scrollbar">
              <button v-for="suggestion in chatSuggestions" :key="suggestion.title"
                @click="handleSuggestionClick(suggestion)"
                class="group px-6 py-3 rounded-xl bg-white/80 backdrop-blur-sm border border-white/50 hover:bg-white/95 hover:shadow-xl transition-all duration-300 transform hover:scale-[1.02] text-left">
                <div
                  :class="`w-10 h-10 rounded-lg bg-gradient-to-r ${suggestion.color} flex items-center justify-center mb-4 group-hover:scale-110 transition-transform`">
                  <i :class="`pi ${suggestion.icon} text-white`"></i>
                </div>
                <h3 class="font-semibold text-gray-900 mb-2">{{ suggestion.title }}</h3>
                <p class="text-sm text-gray-600 leading-relaxed">{{ suggestion.description }}</p>
              </button>
            </div>
          </div>

          <!-- Slide 3: Features -->
          <div
            :class="currentSlide === 2 ? 'translate-x-0 opacity-100' : currentSlide > 2 ? '-translate-x-full opacity-0' : 'translate-x-full opacity-0'"
            class="absolute inset-0 transition-all duration-700 ease-in-out transform p-8">
            <div class="text-center mb-8">
              <h2 class="text-3xl font-bold text-gray-900 mb-3">Why Choose Gemmie?</h2>
              <p class="text-gray-600 text-lg">Built with your privacy and convenience in mind</p>
            </div>

            <div class="space-y-2 h-96 overflow-y-auto custom-scrollbar">
              <div v-for="feature in features" :key="feature.title"
                class="flex items-start gap-4 py-3 px-6 rounded-xl bg-white/80 backdrop-blur-sm border border-white/50 hover:bg-white/95 transition-all duration-300">
                <div class="flex-shrink-0 w-10 h-10 rounded-lg bg-gray-100 flex items-center justify-center">
                  <i :class="`pi ${feature.icon} ${feature.color}`"></i>
                </div>
                <div>
                  <h3 class="font-semibold text-gray-900 mb-2">{{ feature.title }}</h3>
                  <p class="text-sm text-gray-600 leading-relaxed">{{ feature.description }}</p>
                </div>
              </div>
            </div>
          </div>

          <!-- Slide 4: Tips -->
          <div :class="currentSlide === 3 ? 'translate-x-0 opacity-100' : '-translate-x-full opacity-0'"
            class="absolute inset-0 transition-all duration-700 ease-in-out transform p-8">
            <div class="text-center mb-8">
              <h2 class="text-3xl font-bold text-gray-900 mb-3">Pro Tips</h2>
              <p class="text-gray-600 text-lg">Get the most out of your AI conversations</p>
            </div>

            <div class="space-y-4 h-96 overflow-y-auto custom-scrollbar">
              <div v-for="tip in tips" :key="tip.title"
                class="py-3 px-6 rounded-xl bg-white/80 backdrop-blur-sm border border-white/50 hover:bg-white/95 transition-all duration-300">
                <div class="flex items-start gap-4 mb-3">
                  <div class="flex-shrink-0 w-10 h-10 rounded-lg bg-blue-100 flex items-center justify-center">
                    <i :class="`pi ${tip.icon} text-blue-600`"></i>
                  </div>
                  <div>
                    <h3 class="font-semibold text-gray-900 mb-2">{{ tip.title }}</h3>
                    <p class="text-sm text-gray-600 leading-relaxed">{{ tip.description }}</p>
                  </div>
                </div>
                <div class="ml-14 p-3 bg-blue-50 rounded-lg">
                  <p class="text-xs text-blue-700 italic">{{ tip.example }}</p>
                </div>
              </div>
            </div>
          </div>

          <!-- Slide Indicators -->
          <div class="absolute bottom-6 left-1/2 -translate-x-1/2 flex gap-3">
            <button v-for="(slide, index) in carouselSlides" :key="slide.id" @click="goToSlide(index)"
              :class="currentSlide === index ? 'bg-black shadow-lg scale-110' : 'bg-gray-100 hover:bg-white'"
              class="group w-10 h-10 rounded-full transition-all duration-300 flex items-center justify-center"
              :title="slide.title">
              <i
                :class="currentSlide === index ? `text-white pi ${slide.icon}` : `text-gray-500 pi ${slide.icon}`"></i>
            </button>
          </div>

          <!-- Slide Title -->
          <div class="max-md:hidden absolute top-6 left-6">
            <div class="px-4 shadow py-2 bg-gray-100 backdrop-blur-sm rounded-full">
              <span class="text-sm font-medium text-gray-700">
                {{ carouselSlides[currentSlide]?.title }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- RIGHT SECTION: Auth Form -->
      <div class="flex-1 max-w-md">
        <!-- Multi-step Auth Form -->
        <div class="text-sm relative overflow-hidden backdrop-blur-sm  p-8 border border-white/50">

          <!-- Progress indicator -->
          <div class="flex justify-center mb-8">
            <div class="flex items-center space-x-2">
              <div v-for="step in 3" :key="step" :class="step <= authStep ? 'bg-blue-600' : 'bg-gray-300'"
                class="w-3 h-3 rounded-full transition-colors duration-300">
              </div>
            </div>
          </div>

          <!-- Multi-step form container -->
          <div class="relative h-96">
            <!-- Step 1: Username -->
            <div :class="authStep === 1 ? 'translate-x-0 opacity-100' :
              authStep > 1 ? '-translate-x-full opacity-0' : 'translate-x-full opacity-0'"
              class="absolute inset-0 transition-all duration-500 ease-in-out transform">
              <div class="text-center mb-8">
                <h2 class="text-2xl font-semibold text-gray-900 mb-3">Welcome!</h2>
                <p class="text-gray-600">Let's start by creating your username</p>
              </div>

              <form @submit.prevent="handleStepSubmit" class="space-y-6">
                <div>
                  <label class="block text-sm font-medium text-gray-700 mb-3">
                    Choose a username
                  </label>
                  <input v-model="authData.username" required type="text" placeholder="johndoe"
                    class="border border-gray-300 rounded-xl px-4 py-3 w-full text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
                    :class="authData.username && !validateCurrentStep ? 'border-red-300' : ''"
                    @input="handleUsernameInput" />
                  <p class="text-xs text-gray-500 mt-2">This will be your display name</p>
                </div>

                <button type="submit" :disabled="!validateCurrentStep"
                  class="w-full bg-gradient-to-r from-blue-500 to-purple-600 text-white rounded-xl px-6 py-3 font-semibold hover:from-blue-600 hover:to-purple-700 disabled:from-gray-300 disabled:to-gray-400 disabled:cursor-not-allowed transition-all duration-300 transform hover:scale-[1.02] shadow-lg">
                  Continue
                </button>
              </form>
            </div>

            <!-- Step 2: Email -->
            <div :class="authStep === 2 ? 'translate-x-0 opacity-100' :
              authStep > 2 ? '-translate-x-full opacity-0' : 'translate-x-full opacity-0'"
              class="absolute inset-0 transition-all duration-500 ease-in-out transform">
              <div class="text-center mb-8">
                <h2 class="text-2xl font-semibold text-gray-900 mb-3">Hi {{ authData.username }}!</h2>
                <p class="text-gray-600">What's your email address?</p>
              </div>

              <form @submit.prevent="handleStepSubmit" class="space-y-6">
                <div>
                  <label class="block text-sm font-medium text-gray-700 mb-3">
                    Email address
                  </label>
                  <input v-model="authData.email" required type="email" placeholder="johndoe@example.com"
                    class="border border-gray-300 rounded-xl px-4 py-3 w-full text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
                    :class="authData.email && !validateCurrentStep ? 'border-red-300' : ''" @input="handleEmailInput" />
                  <p class="text-xs text-gray-500 mt-2">Used for session identification only</p>
                </div>

                <div class="flex gap-4">
                  <button type="button" @click="prevAuthStep"
                    class="flex-1 flex gap-2 items-center justify-center bg-gray-100 backdrop-blur-sm text-gray-700 rounded-xl px-4 py-3 font-medium hover:bg-gray-200 transition-all duration-200">
                    <i class="pi pi-arrow-left"></i> Back
                  </button>
                  <button type="submit" :disabled="!validateCurrentStep"
                    class="bg-gradient-to-r from-blue-500 to-purple-600 flex-1 flex gap-2 items-center justify-center hover:from-blue-600 hover:to-purple-700 disabled:from-gray-300 disabled:to-gray-400 disabled:cursor-not-allowed transform hover:scale-[1.02] shadow-lg rounded-xl px-4 py-3 font-medium text-white transition-all duration-200">
                    Continue
                  </button>
                </div>
              </form>
            </div>

            <!-- Step 3: Password -->
            <div :class="authStep === 3 ? 'translate-x-0 opacity-100' : 'translate-x-full opacity-0'"
              class="absolute inset-0 transition-all duration-500 ease-in-out transform">
              <div class="text-center mb-8">
                <h2 class="text-2xl font-semibold text-gray-900 mb-3">Almost there!</h2>
                <p class="text-gray-600">Create a secure password</p>
              </div>

              <form @submit.prevent="handleStepSubmit" class="space-y-6">
                <div>
                  <label class="block text-sm font-medium text-gray-700 mb-3">
                    Password
                  </label>
                  <input v-model="authData.password" required type="password" placeholder="Enter a secure password"
                    minlength="8"
                    class="border border-gray-300 rounded-xl px-4 py-3 w-full text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
                    :class="authData.password && !validateCurrentStep ? 'border-red-300' : ''"
                    @input="handlePasswordInput" />
                  <div class="mt-3">
                    <div class="flex items-center gap-2 text-xs">
                      <div :class="authData.password.length >= 8 ? 'text-green-600' : 'text-gray-400'"
                        class="flex items-center gap-1">
                        <i :class="authData.password.length >= 8 ? 'pi pi-check' : 'pi pi-circle'" class="text-xs"></i>
                        <span>At least 8 characters</span>
                      </div>
                    </div>
                  </div>
                </div>

                <div class="flex gap-4">
                  <button type="button" @click="prevAuthStep"
                    class="flex-1 flex gap-2 items-center justify-center bg-gray-100 backdrop-blur-sm text-gray-700 rounded-xl px-4 py-3 font-medium hover:bg-gray-200 transition-all duration-200">
                    <i class="pi pi-arrow-left"></i> Back
                  </button>
                  <button type="submit" :disabled="!validateCurrentStep || isLoading"
                    class="bg-gradient-to-r from-blue-500 to-purple-600 flex-1 flex gap-2 items-center justify-center hover:from-blue-600 hover:to-purple-700 disabled:from-gray-300 disabled:to-gray-400 disabled:cursor-not-allowed transform hover:scale-[1.02] shadow-lg rounded-xl px-4 py-3 font-medium text-white transition-all duration-200">
                    <i v-if="isLoading" class="pi pi-spin pi-spinner" :class="isLoading ? '' : 'pi pi-check'"></i>
                    <span>{{ isLoading ? 'Creating...' : 'Create Session' }}</span>
                  </button>
                </div>
              </form>
            </div>
          </div>

          <!-- Footer note -->
          <div class="text-center mt-6">
            <p class="text-xs text-gray-500 leading-relaxed">
              Your data is securely encrypted and synced across all your devices.
              <br>Create an account to access your chats anywhere.
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- Mobile Layout: Vertical stack with carousel -->
    <div v-if="isMobile" class="flex flex-col gap-8 items-center  justify-center h-full w-full px-2">
      <!-- Mobile Carousel (always shown) -->
      <div v-if="!showCreateSession" class="w-full max-w-sm" @touchstart="stopAutoSlide" @touchend="startAutoSlide">
        <div class="relative h-[440px] overflow-hidden backdrop-blur-sm rounded-2xl bg-white/80 border border-white/50">
          <!-- Mobile Slide 1: Welcome -->
          <div
            :class="currentSlide === 0 ? 'translate-x-0 opacity-100' : currentSlide > 0 ? '-translate-x-full opacity-0' : 'translate-x-full opacity-0'"
            class="absolute inset-0 transition-all duration-700 ease-in-out transform p-6 flex flex-col items-center justify-center">
            <div class="text-center">
              <img src="/logo.svg" alt="Gemmie Logo" class="rounded-full w-16 h-16 mb-4 mx-auto" />
              <h2 class="text-2xl font-bold text-gray-900 mb-2">
                Welcome to Gemmie
              </h2>
              <p class="text-gray-600 leading-relaxed mb-6 text-sm">
                Experience privacy-first conversations with advanced AI.
              </p>
              <div class="flex items-center justify-center gap-6 text-xs">
                <div class="flex flex-col items-center gap-1">
                  <div class="w-8 h-8 bg-green-100 rounded-full flex items-center justify-center">
                    <i class="pi pi-shield text-green-600 text-sm"></i>
                  </div>
                  <span class="text-gray-700 font-medium">Private</span>
                </div>
                <div class="flex flex-col items-center gap-1">
                  <div class="w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center">
                    <i class="pi pi-database text-blue-600 text-sm"></i>
                  </div>
                  <span class="text-gray-700 font-medium">Local</span>
                </div>
                <div class="flex flex-col items-center gap-1">
                  <div class="w-8 h-8 bg-purple-100 rounded-full flex items-center justify-center">
                    <i class="pi pi-sync text-purple-600 text-sm"></i>
                  </div>
                  <span class="text-gray-700 font-medium">Synced</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Mobile Slide 2: Chat Suggestions -->
          <div
            :class="currentSlide === 1 ? 'translate-x-0 opacity-100' : currentSlide > 1 ? '-translate-x-full opacity-0' : 'translate-x-full opacity-0'"
            class="absolute inset-0 transition-all duration-700 ease-in-out transform p-4">
            <div class="text-center mb-4">
              <h2 class="text-xl font-bold text-gray-900 mb-2">What can I help with?</h2>
              <p class="text-gray-600 text-sm">Try these conversation starters</p>
            </div>

            <div class="grid grid-cols-2 gap-3 h-72 overflow-y-auto">
              <button v-for="suggestion in chatSuggestions" :key="suggestion.title"
                @click="handleSuggestionClick(suggestion)"
                class="group p-3 rounded-lg bg-white/70 backdrop-blur-sm border border-white/50 hover:bg-white/90 hover:shadow-lg transition-all duration-300 transform hover:scale-[1.02] text-left">
                <div
                  :class="`w-8 h-8 rounded-lg bg-gradient-to-r ${suggestion.color} flex items-center justify-center mb-2 group-hover:scale-110 transition-transform`">
                  <i :class="`pi ${suggestion.icon} text-white text-sm`"></i>
                </div>
                <h3 class="font-semibold text-gray-900 text-xs mb-1">{{ suggestion.title }}</h3>
                <p class="text-xs text-gray-600 leading-tight">{{ suggestion.description }}</p>
              </button>
            </div>
          </div>

          <!-- Mobile Slide 3: Features -->
          <div
            :class="currentSlide === 2 ? 'translate-x-0 opacity-100' : currentSlide > 2 ? '-translate-x-full opacity-0' : 'translate-x-full opacity-0'"
            class="absolute inset-0 transition-all duration-700 ease-in-out transform p-4">
            <div class="text-center mb-4">
              <h2 class="text-xl font-bold text-gray-900 mb-2">Why Choose Gemmie?</h2>
              <p class="text-gray-600 text-sm">Built with your privacy in mind</p>
            </div>

            <div class="space-y-3 h-72 overflow-y-auto">
              <div v-for="feature in features" :key="feature.title"
                class="flex items-start gap-3 p-3 rounded-lg bg-white/70 backdrop-blur-sm border border-white/50 hover:bg-white/90 transition-all duration-300 hover:shadow-md">
                <div class="flex-shrink-0 w-8 h-8 rounded-lg bg-gray-100 flex items-center justify-center">
                  <i :class="`pi ${feature.icon} ${feature.color} text-sm`"></i>
                </div>
                <div>
                  <h3 class="font-semibold text-gray-900 text-xs mb-1">{{ feature.title }}</h3>
                  <p class="text-xs text-gray-600 leading-tight">{{ feature.description }}</p>
                </div>
              </div>
            </div>
          </div>

          <!-- Mobile Slide 4: Tips -->
          <div :class="currentSlide === 3 ? 'translate-x-0 opacity-100' : '-translate-x-full opacity-0'"
            class="absolute inset-0 transition-all duration-700 ease-in-out transform p-4">
            <div class="text-center mb-4">
              <h2 class="text-xl font-bold text-gray-900 mb-2">Pro Tips</h2>
              <p class="text-gray-600 text-sm">Get the most out of conversations</p>
            </div>

            <div class="space-y-3 h-72 overflow-y-auto">
              <div v-for="tip in tips" :key="tip.title"
                class="p-3 rounded-lg bg-white/70 backdrop-blur-sm border border-white/50 hover:bg-white/90 transition-all duration-300 hover:shadow-md">
                <div class="flex items-start gap-2 mb-2">
                  <div class="flex-shrink-0 w-6 h-6 rounded-lg bg-blue-100 flex items-center justify-center">
                    <i :class="`pi ${tip.icon} text-blue-600 text-xs`"></i>
                  </div>
                  <div>
                    <h3 class="font-semibold text-gray-900 text-xs mb-1">{{ tip.title }}</h3>
                    <p class="text-xs text-gray-600 leading-tight">{{ tip.description }}</p>
                  </div>
                </div>
                <div class="ml-8 p-2 bg-blue-50 rounded-lg">
                  <p class="text-xs text-blue-700 italic">{{ tip.example }}</p>
                </div>
              </div>
            </div>
          </div>

          <!-- Mobile Slide Indicators -->
          <div class="absolute bottom-2 left-1/2 -translate-x-1/2 flex gap-2">
            <button v-for="(slide, index) in carouselSlides" :key="slide.id" @click="goToSlide(index)"
              :class="currentSlide === index ? 'bg-black shadow-lg scale-110' : 'bg-gray-100 hover:bg-gray-100'"
              class="group w-8 h-8 rounded-full transition-all duration-300 flex items-center justify-center"
              :title="slide.title">
              <i
                :class="currentSlide === index ? `text-white pi ${slide.icon} text-xs ` : `text-gray-500 pi ${slide.icon} text-xs`"></i>
            </button>
          </div>
        </div>
      </div>

      <!-- Mobile Auth Section -->
      <div v-if="showCreateSession" class="w-full max-w-sm">
        <div class="bg-white/80 backdrop-blur-sm rounded-2xl p-2 border border-white/50">
          <!-- Progress indicator -->
          <div class="flex justify-center mb-6">
            <div class="flex items-center space-x-2">
              <div v-for="step in 3" :key="step" :class="step <= authStep ? 'bg-blue-600' : 'bg-gray-300'"
                class="w-2.5 h-2.5 rounded-full transition-colors duration-300">
              </div>
            </div>
          </div>

          <!-- Mobile Multi-step form container -->
          <div class="relative h-72">
            <!-- Mobile Step 1: Username -->
            <div :class="authStep === 1 ? 'translate-x-0 opacity-100' :
              authStep > 1 ? '-translate-x-full opacity-0' : 'translate-x-full opacity-0'"
              class="absolute inset-0 transition-all duration-500 ease-in-out transform">
              <div class="text-center mb-6">
                <h2 class="text-xl font-semibold text-gray-900 mb-2">Welcome!</h2>
                <p class="text-gray-600 text-sm">Let's start by creating your username</p>
              </div>

              <form @submit.prevent="handleStepSubmit" class="space-y-4">
                <div>
                  <label class="block text-sm font-medium text-gray-700 mb-2">
                    Choose a username
                  </label>
                  <input v-model="authData.username" required type="text" placeholder="johndoe"
                    class="border border-gray-300 rounded-lg px-4 py-2.5 w-full text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
                    :class="authData.username && !validateCurrentStep ? 'border-red-300' : ''"
                    @input="handleUsernameInput" />
                  <p class="text-xs text-gray-500 mt-1">This will be your display name</p>
                </div>

                <button type="submit" :disabled="!validateCurrentStep"
                  class="w-full bg-gradient-to-r from-blue-500 to-purple-600 text-white rounded-lg px-6 py-2.5 font-semibold hover:from-blue-600 hover:to-purple-700 disabled:from-gray-300 disabled:to-gray-400 disabled:cursor-not-allowed transition-all duration-300 transform hover:scale-[1.02] shadow-lg">
                  Continue
                </button>
              </form>
            </div>

            <!-- Mobile Step 2: Email -->
            <div :class="authStep === 2 ? 'translate-x-0 opacity-100' :
              authStep > 2 ? '-translate-x-full opacity-0' : 'translate-x-full opacity-0'"
              class="absolute inset-0 transition-all duration-500 ease-in-out transform">
              <div class="text-center mb-6">
                <h2 class="text-xl font-semibold text-gray-900 mb-2">Hi {{ authData.username }}!</h2>
                <p class="text-gray-600 text-sm">What's your email address?</p>
              </div>

              <form @submit.prevent="handleStepSubmit" class="space-y-4">
                <div>
                  <label class="block text-sm font-medium text-gray-700 mb-2">
                    Email address
                  </label>
                  <input v-model="authData.email" required type="email" placeholder="johndoe@example.com"
                    class="border border-gray-300 rounded-lg px-4 py-2.5 w-full text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
                    :class="authData.email && !validateCurrentStep ? 'border-red-300' : ''" @input="handleEmailInput" />
                  <p class="text-xs text-gray-500 mt-1">Used for session identification only</p>
                </div>

                <div class="flex gap-3">
                  <button type="button" @click="prevAuthStep"
                    class="flex-1 flex gap-2 items-center justify-center bg-gray-100 backdrop-blur-sm text-gray-700 rounded-lg px-4 py-2.5 font-medium hover:bg-gray-200 transition-all duration-200">
                    <i class="pi pi-arrow-left"></i> Back
                  </button>
                  <button type="submit" :disabled="!validateCurrentStep"
                    class="bg-gradient-to-r from-blue-500 to-purple-600 flex-1 flex gap-2 items-center justify-center hover:from-blue-600 hover:to-purple-700 disabled:from-gray-300 disabled:to-gray-400 disabled:cursor-not-allowed transform hover:scale-[1.02] shadow-lg rounded-lg px-4 py-2.5 font-medium text-white transition-all duration-200">
                    Continue
                  </button>
                </div>
              </form>
            </div>

            <!-- Mobile Step 3: Password -->
            <div :class="authStep === 3 ? 'translate-x-0 opacity-100' : 'translate-x-full opacity-0'"
              class="absolute inset-0 transition-all duration-500 ease-in-out transform">
              <div class="text-center mb-6">
                <h2 class="text-xl font-semibold text-gray-900 mb-2">Almost there!</h2>
                <p class="text-gray-600 text-sm">Create a secure password</p>
              </div>

              <form @submit.prevent="handleStepSubmit" class="space-y-4">
                <div>
                  <label class="block text-sm font-medium text-gray-700 mb-2">
                    Password
                  </label>
                  <input v-model="authData.password" required type="password" placeholder="Enter a secure password"
                    minlength="8"
                    class="border border-gray-300 rounded-lg px-4 py-2.5 w-full text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
                    :class="authData.password && !validateCurrentStep ? 'border-red-300' : ''"
                    @input="handlePasswordInput" />
                  <div class="mt-2">
                    <div class="flex items-center gap-2 text-xs">
                      <div :class="authData.password.length >= 8 ? 'text-green-600' : 'text-gray-400'"
                        class="flex items-center gap-1">
                        <i :class="authData.password.length >= 8 ? 'pi pi-check' : 'pi pi-circle'" class="text-xs"></i>
                        <span>At least 8 characters</span>
                      </div>
                    </div>
                  </div>
                </div>

                <div class="flex gap-3">
                  <button type="button" @click="prevAuthStep"
                    class="flex-1 flex gap-2 items-center justify-center bg-gray-100 backdrop-blur-sm text-gray-700 rounded-lg px-4 py-2.5 font-medium hover:bg-gray-200 transition-all duration-200">
                    <i class="pi pi-arrow-left"></i> Back
                  </button>
                  <button type="submit" :disabled="!validateCurrentStep || isLoading"
                    class="bg-gradient-to-r from-blue-500 to-purple-600 flex-1 flex gap-2 items-center justify-center hover:from-blue-600 hover:to-purple-700 disabled:from-gray-300 disabled:to-gray-400 disabled:cursor-not-allowed transform hover:scale-[1.02] shadow-lg rounded-lg px-4 py-2.5 font-medium text-white transition-all duration-200">
                    <i v-if="isLoading" class="pi pi-spin pi-spinner" :class="isLoading ? '' : 'pi pi-check'"></i>
                    <span>{{ isLoading ? 'Creating...' : 'Create Session' }}</span>
                  </button>
                </div>
              </form>
            </div>
          </div>

          <!-- Mobile Footer note -->
          <div class="text-center mt-4">
            <p class="text-xs text-gray-500 leading-relaxed">
              Your data is securely encrypted and synced across all your devices.
            </p>
          </div>
        </div>
      </div>

      <!-- Mobile Get Started Button (when not in auth mode) -->
      <div v-else class="w-full max-w-sm">
        <button @click="() => setShowCreateSession(true)"
          class="group w-full px-6 py-3 bg-gradient-to-r from-indigo-500 to-blue-600 text-white rounded-xl hover:from-indigo-600 hover:to-blue-700 transition-all duration-300 transform hover:scale-105 shadow-lg hover:shadow-xl font-medium">
          <span class="flex items-center justify-center gap-2">
            <i class="pi pi-arrow-right group-hover:translate-x-1 transition-transform"></i>
            Get Started
          </span>
        </button>
      </div>
    </div>

    <!-- Footer disclaimer -->
    <div v-if="isAuthenticated" class="absolute bottom-4 text-center">
      <p class="text-xs text-gray-500">
        Gemmie can make mistakes. Check important info.
      </p>
    </div>
  </div>
</template>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}

.custom-scrollbar::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 3px;
}

.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.3);
  border-radius: 3px;
}

.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.5);
}
</style>