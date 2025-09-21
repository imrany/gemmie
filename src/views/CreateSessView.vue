<script lang="ts" setup>
import { computed } from 'vue'
import type { Chat } from '@/types'

// --- All Props Directly from HomeView ---
const props = defineProps<{
  chats: Chat[]
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
  isAuthenticated: () => boolean
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

// Handle input updates
const handleUsernameInput = (event: Event) => {
  const value = (event.target as HTMLInputElement).value
  props.updateAuthData({ username: value })
}

const handleEmailInput = (event: Event) => {
  const value = (event.target as HTMLInputElement).value
  props.updateAuthData({ email: value })
}

const handlePasswordInput = (event: Event) => {
  const value = (event.target as HTMLInputElement).value
  props.updateAuthData({ password: value })
}

// Show/hide logic
const shouldShowWelcome = computed(() => {
  return (props.showCreateSession === false && props.screenWidth < 720) || 
         props.screenWidth > 720
})

const shouldShowAuth = computed(() => {
  return (!props.isAuthenticated() && props.showCreateSession === true && props.screenWidth < 720) || 
         (!props.isAuthenticated() && props.screenWidth > 720)
})
</script>

<template>
  <div class="flex flex-col items-center justify-center h-[90vh]">
    <div class="max-md:flex-col flex gap-10 items-center justify-center h-full w-full max-md:px-5">
      
      <!-- Welcome Section -->
      <div v-if="shouldShowWelcome"
           class="flex flex-col md:flex-grow items-center gap-3 text-gray-600">
        <div class="rounded-full bg-gray-200 w-[60px] h-[60px] flex justify-center items-center">
          <span class="pi pi-comment text-lg"></span>
        </div>
        <p class="text-3xl font-semibold">{{ parsedUserDetails?.username || 'Gemmie' }}</p>
        <div class="text-center text-base md:max-w-[400px]">
          <p>Your private AI assistant.</p>
          <p class="text-sm text-gray-500">
            We take your privacy seriously. All your data stays on your device.
            Your chats are stored locally in your browser, so they are never sent to any server.
            Make sure to back up your chats if you clear your browser data or switch to a new device.
          </p>
        </div>
        <button v-if="isAuthenticated()" 
                @click="setShowInput"
                class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-500 transition-colors">
          Write a prompt
        </button>
        <button v-else-if="screenWidth < 720" 
                @click="() => setShowCreateSession(true)"
                class="px-4 py-2 w-full bg-blue-600 text-white rounded-lg hover:bg-blue-500 transition-colors">
          Get Started
        </button>
      </div>

      <!-- Multi-step Auth Form -->
      <div v-if="shouldShowAuth"
           class="flex-grow text-sm md:px-4 px-1 relative overflow-hidden"
           :class="screenWidth > 720 ? 'max-w-md w-full' : (!isAuthenticated() && showCreateSession === true) ?
               'flex flex-col justify-center w-full h-full translate-x-0 opacity-100' : 'translate-x-full opacity-0'">

        <!-- Progress indicator -->
        <div class="flex justify-center mb-6">
          <div class="flex items-center space-x-2">
            <div v-for="step in 3" 
                 :key="step" 
                 :class="step <= authStep ? 'bg-blue-600' : 'bg-gray-300'"
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

            <form @submit.prevent="handleStepSubmit" class="space-y-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">
                  Choose a username
                </label>
                <input 
                  v-model="authData.username" 
                  required 
                  type="text" 
                  placeholder="johndoe"
                  class="border border-gray-300 rounded-lg px-4 py-2 w-full text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
                  :class="authData.username && !validateCurrentStep ? 'border-red-300' : ''"
                  @input="handleUsernameInput" />
                <p class="text-xs text-gray-500 mt-1">This will be your display name</p>
              </div>

              <button type="submit" 
                      :disabled="!validateCurrentStep"
                      class="w-full bg-blue-600 text-white rounded-lg px-4 py-2 font-medium hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed transition-colors duration-200">
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

            <form @submit.prevent="handleStepSubmit" class="space-y-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">
                  Email address
                </label>
                <input 
                  v-model="authData.email" 
                  required 
                  type="email" 
                  placeholder="johndoe@example.com"
                  class="border border-gray-300 rounded-lg px-4 py-2 w-full text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
                  :class="authData.email && !validateCurrentStep ? 'border-red-300' : ''"
                  @input="handleEmailInput" />
                <p class="text-xs text-gray-500 mt-1">Used for session identification only</p>
              </div>

              <div class="flex gap-3">
                <button type="button" 
                        @click="prevAuthStep"
                        class="flex-1 bg-gray-100 text-gray-700 rounded-lg px-4 py-2 font-medium hover:bg-gray-200 transition-colors duration-200">
                  Back
                </button>
                <button type="submit" 
                        :disabled="!validateCurrentStep"
                        class="flex-1 bg-blue-600 text-white rounded-lg px-4 py-2 font-medium hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed transition-colors duration-200">
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

            <form @submit.prevent="handleStepSubmit" class="space-y-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">
                  Password
                </label>
                <input 
                  v-model="authData.password" 
                  required 
                  type="password"
                  placeholder="Enter a secure password" 
                  minlength="8"
                  class="border border-gray-300 rounded-lg px-4 py-2 w-full text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
                  :class="authData.password && !validateCurrentStep ? 'border-red-300' : ''"
                  @input="handlePasswordInput" />
                <div class="mt-2">
                  <div class="flex items-center gap-2 text-xs">
                    <div :class="authData.password.length >= 8 ? 'text-green-600' : 'text-gray-400'"
                         class="flex items-center gap-1">
                      <i :class="authData.password.length >= 8 ? 'pi pi-check' : 'pi pi-circle'"
                         class="text-xs"></i>
                      <span>At least 8 characters</span>
                    </div>
                  </div>
                </div>
              </div>

              <div class="flex gap-3">
                <button type="button" 
                        @click="prevAuthStep"
                        class="flex-1 bg-gray-100 text-gray-700 rounded-lg px-4 py-2 font-medium hover:bg-gray-200 transition-colors duration-200">
                  Back
                </button>
                <button type="submit" 
                        :disabled="!validateCurrentStep || isLoading"
                        class="flex-1 bg-blue-600 text-white rounded-lg px-4 py-2 font-medium hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed transition-colors duration-200 flex items-center justify-center gap-2">
                  <i v-if="isLoading" class="pi pi-spin pi-spinner"
                     :class="isLoading ? '' : 'pi pi-check'"></i>
                  <span>{{ isLoading ? 'Creating...' : 'Create Session' }}</span>
                </button>
              </div>
            </form>
          </div>
        </div>

        <!-- Footer note -->
        <div class="text-center">
          <p class="text-xs text-gray-500 leading-relaxed">
            Your data is securely encrypted and synced across all your devices.
            <br>Create an account to access your chats anywhere.
          </p>
        </div>
      </div>
    </div>

    <p v-if="isAuthenticated()" class="text-sm mt-2 text-gray-400">
      Gemmie can make mistakes. Check important info.
    </p>
  </div>
</template>