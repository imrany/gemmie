<script lang="ts" setup>
import router from '@/router';
import type { Chat, ConfirmDialogOptions, UserDetails } from '@/types';
import { inject, ref, reactive, type Ref, computed, watch, onMounted } from 'vue';
import SideNav from '@/components/SideNav.vue';
import { toast } from 'vue-sonner';
import { useRoute } from 'vue-router';

const globalState = inject('globalState') as {
  screenWidth: Ref<number>,
  confirmDialog: Ref<ConfirmDialogOptions>,
  isCollapsed: Ref<boolean>,
  authData: Ref<{
    username: string;
    email: string;
    password: string;
    workFunction?: string,
    preferences?: string
  }>,
  syncStatus: Ref<{ lastSync: Date | null; syncing: boolean; hasUnsyncedChanges: boolean; }>,
  isAuthenticated: Ref<boolean>,
  parsedUserDetails: Ref<UserDetails>,
  currentChatId: Ref<string>,
  chats: Ref<Chat[]>
  logout: () => void,
  isLoading: Ref<boolean>,
  showInput: Ref<boolean>,
  planStatus: Ref<{ status: string; timeLeft: string; expiryDate: string; isExpired: boolean; }>,
  isDarkMode: Ref<boolean>,
  syncEnabled: Ref<boolean>, // Add this

  toggleTheme: (theme: string) => void,
  hideSidebar: () => void,
  setShowInput: () => void,
  clearAllChats: () => void,
  switchToChat: (chatId: string) => void,
  createNewChat: (initialMessage?: string) => void,
  deleteChat: (chatId: string) => void,
  renameChat: (chatId: string, newTitle: string) => void,
  toggleSidebar: () => void,
  manualSync: () => Promise<any>,
  apiCall: (endpoint: string, options?: RequestInit) => Promise<any>,
  isLocalDataEmpty: () => boolean,
  toggleSync: (syncEnabled?: boolean) => Promise<void> // Fix the type
}

const {
  screenWidth,
  isCollapsed,
  planStatus,
  isAuthenticated,
  currentChatId,
  chats,
  logout,
  syncStatus,
  toggleSidebar,
  manualSync,
  hideSidebar,
  setShowInput,
  clearAllChats,
  switchToChat,
  createNewChat,
  deleteChat,
  renameChat,
  toggleTheme,
  toggleSync,
  parsedUserDetails,
} = globalState

const route = useRoute()
const tabParam = route.params.tab as 'general' | 'account' | 'privacy' | 'billing' | undefined

// Local form state
const profileData = reactive({
  username: parsedUserDetails.value?.username || '',
  email: parsedUserDetails.value?.email || '',
  workFunction: parsedUserDetails.value?.workFunction || '',
  preferences: parsedUserDetails.value?.preferences || ''
})

const activeTab = ref<'general' | 'account' | 'privacy' | 'billing'>(tabParam ?? 'general')
const isSaving = ref(false)

const isTogglingSync = ref(false)

async function handleToggleSync() {
  try {
    isTogglingSync.value = true
    await toggleSync()
  } catch (error: any) {
    console.error('Failed to toggle sync:', error)
  } finally {
    isTogglingSync.value = false
  }
}

// Save profile changes - only sync to server if sync is enabled
async function saveProfile() {
  if (!profileData.username.trim()) {
    toast.error('Username is required')
    return
  }

  try {
    isSaving.value = true

    // Update global state
    parsedUserDetails.value.username = profileData.username.trim()
    parsedUserDetails.value.workFunction = profileData.workFunction
    parsedUserDetails.value.preferences = profileData.preferences
  } catch (error) {
    console.error('Failed to save profile:', error)
    toast.error('Failed to save profile changes')
  } finally {
    isSaving.value = false
  }
}

// Reset form when switching tabs
function resetProfileData() {
  profileData.username = parsedUserDetails.value?.username || ''
  profileData.email = parsedUserDetails.value?.email || ''
  profileData.workFunction = parsedUserDetails.value?.workFunction || ''
  profileData.preferences = parsedUserDetails.value?.preferences || ''
}

// Detect unsaved changes
const hasUnsavedChanges = computed(() => {
  return profileData.workFunction !== (parsedUserDetails.value?.workFunction || '') ||
    profileData.preferences !== (parsedUserDetails.value?.preferences || '')
})

/**
 * Watch tab changes
 */
watch(activeTab, (newVal, oldVal) => {
  if (hasUnsavedChanges.value) {
    const confirmLeave = confirm("You have unsaved changes. Leave without saving?")
    if (!confirmLeave) {
      activeTab.value = oldVal
      return
    }
  }
  router.push({ name: 'settings', params: { tab: newVal } })
})

// Lifecycle hooks
onMounted(() => {
  if (isAuthenticated.value && parsedUserDetails?.value) {
    resetProfileData()
    router.push({ 
      name: 'settings', 
      params: { tab: activeTab.value || 'general' } 
    })
  } else {
    logout()
    router.push('/')
  }
})

watch(isAuthenticated, (val) => {
  if (val === false) {
    router.push('/')
  }
})
</script>

<template>
  <div class="flex h-[100vh] bg-white dark:bg-gray-900 transition-colors duration-300">
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
      manualSync
    }" />

    <!-- Main Content -->
    <div :class="screenWidth > 720 && isAuthenticated
      ? (!isCollapsed
        ? 'flex-grow flex flex-col ml-[270px] font-light text-sm transition-all duration-300 ease-in-out'
        : 'flex-grow flex flex-col ml-[60px] font-light text-sm transition-all duration-300 ease-in-out')
      : 'text-sm font-light flex-grow flex flex-col transition-all duration-300 ease-in-out'">

      <div class="flex flex-col p-4 md:p-6 min-h-0 flex-1">
        <div class="flex items-center justify-between mb-4">
          <button @click="router.push('/')" title="Go Back"
            class="md:hidden flex items-center justify-center w-8 h-8 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-full cursor-pointer transition-colors">
            <span class="pi pi-chevron-left text-lg text-gray-700 dark:text-gray-300"></span>
          </button>
          <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">Settings</h1>
        </div>

        <!-- Horizontal Tabs -->
        <div class="border-b border-gray-200 dark:border-gray-700 mb-2 md:hidden">
          <nav class="flex space-x-8" aria-label="Tabs">
            <button @click="activeTab = 'general'"
              :class="activeTab === 'general'
                ? 'border-gray-500 text-gray-600 dark:text-gray-200 whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm'
                : 'border-transparent text-gray-500 dark:text-gray-200 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300 dark:hover:border-gray-600 whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm transition-colors'">
              General
            </button>
            <button @click="activeTab = 'account'"
              :class="activeTab === 'account'
                ? 'border-gray-500 text-gray-600 dark:text-gray-200 whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm'
                : 'border-transparent text-gray-500 dark:text-gray-200 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300 dark:hover:border-gray-600 whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm transition-colors'">
              Account
            </button>
            <button @click="activeTab = 'privacy'"
              :class="activeTab === 'privacy'
                ? 'border-gray-500 text-gray-600 dark:text-gray-200 whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm'
                : 'border-transparent text-gray-500 dark:text-gray-200 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300 dark:hover:border-gray-600 whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm transition-colors'">
              Privacy
            </button>
            <button @click="activeTab = 'billing'"
              :class="activeTab === 'billing'
                ? 'border-gray-500 text-gray-600 dark:text-gray-200 whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm'
                : 'border-transparent text-gray-500 dark:text-gray-200 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300 dark:hover:border-gray-600 whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm transition-colors'">
              Billing
            </button>
          </nav>
        </div>

        <div class="flex flex-col md:flex-row gap-4 md:gap-8 flex-1 min-h-0">
          <!-- Tabs Sidebar (hidden on mobile) -->
          <div class="w-48 flex-col font-normal text-base gap-2 hidden md:flex flex-shrink-0">
            <button @click="activeTab = 'general'" :class="activeTab === 'general'
              ? 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-200 font-medium'
              : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800'"
              class="text-left px-4 py-2 rounded-md transition-all duration-200">
              General
            </button>
            <button @click="activeTab = 'account'" :class="activeTab === 'account'
              ? 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-200 font-medium'
              : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800'"
              class="text-left px-4 py-2 rounded-md transition-all duration-200">
              Account
            </button>
            <button @click="activeTab = 'privacy'" :class="activeTab === 'privacy'
              ? 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-200 font-medium'
              : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800'"
              class="text-left px-4 py-2 rounded-md transition-all duration-200">
              Privacy
            </button>
            <button @click="activeTab = 'billing'" :class="activeTab === 'billing'
              ? 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-200 font-medium'
              : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800'"
              class="text-left px-4 py-2 rounded-md transition-all duration-200">
              Billing
            </button>
          </div>

          <!-- Content Area -->
          <div class="flex-1 min-h-0 overflow-auto custom-scrollbar">
            <!-- General -->
            <div v-if="activeTab === 'general'"
              class="bg-white dark:bg-gray-800 p-4 md:p-6 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 w-full transition-colors duration-300"
              :class="[
                screenWidth >= 1536 ? 'max-w-4xl' : 
                screenWidth >= 1280 ? 'max-w-3xl' : 
                screenWidth >= 1024 ? 'max-w-2xl' : 
                'max-w-none'
              ]">

              <form @submit.prevent="saveProfile" class="flex flex-col gap-4 md:gap-6">
                <div :class="['grid md:grid-cols-2  gap-4'
                ]">
                  <!-- Username -->
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                      Username
                    </label>
                    <input v-model="profileData.username" type="text" disabled
                      class="border border-gray-300 dark:border-gray-600 rounded-lg px-3 md:px-4 py-2 w-full text-sm bg-gray-50 dark:bg-gray-700 text-gray-600 dark:text-gray-200 cursor-not-allowed transition-colors" />
                    <p class="text-xs text-gray-500 dark:text-gray-200 mt-1">Username cannot be changed</p>
                  </div>
  
                  <!-- Email -->
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Email</label>
                    <input v-model="profileData.email" type="email" disabled
                      class="border border-gray-300 dark:border-gray-600 rounded-lg px-3 md:px-4 py-2 w-full text-sm bg-gray-50 dark:bg-gray-700 text-gray-600 dark:text-gray-200 cursor-not-allowed transition-colors" />
                    <p class="text-xs text-gray-500 dark:text-gray-200 mt-1">Email cannot be changed</p>
                  </div>
                </div>

                <!-- Theme Selection -->
                <div class="max-md:hidden">
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">
                    Theme Preference
                  </label>
                  <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
                    <!-- Light Theme -->
                    <button type="button" @click="toggleTheme('light')"
                      :class="parsedUserDetails?.theme === 'light'
                        ? 'border-blue-500 bg-gray-100 dark:bg-gray-800 ring-2 ring-blue-500 ring-opacity-50'
                        : 'border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600'"
                      class="flex flex-col items-center p-4 border rounded-lg transition-all duration-200 group">
                      <div
                        class="w-12 h-12 mb-2 rounded-lg bg-gradient-to-br from-gray-100 to-gray-200 border border-gray-300 flex items-center justify-center">
                        <i class="pi pi-sun text-yellow-500 text-lg"></i>
                      </div>
                      <span class="text-sm font-medium text-gray-900 dark:text-white">Light</span>
                      <span class="text-xs text-gray-500 dark:text-gray-200 mt-1">Always light</span>
                    </button>

                    <!-- Dark Theme -->
                    <button type="button" @click="toggleTheme('dark')"
                      :class="parsedUserDetails?.theme === 'dark'
                        ? 'border-blue-500 bg-gray-100 dark:bg-gray-800 ring-2 ring-blue-500 ring-opacity-50'
                        : 'border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600'"
                      class="flex flex-col items-center p-4 border rounded-lg transition-all duration-200 group">
                      <div
                        class="w-12 h-12 mb-2 rounded-lg bg-gradient-to-br from-gray-800 to-gray-900 border border-gray-700 flex items-center justify-center">
                        <i class="pi pi-moon text-blue-300 text-lg"></i>
                      </div>
                      <span class="text-sm font-medium text-gray-900 dark:text-white">Dark</span>
                      <span class="text-xs text-gray-500 dark:text-gray-200 mt-1">Always dark</span>
                    </button>

                    <!-- System Theme -->
                    <button type="button" @click="toggleTheme('system')"
                      :class="parsedUserDetails?.theme === 'system'
                        ? 'border-blue-500 bg-gray-100 dark:bg-gray-800 ring-2 ring-blue-500 ring-opacity-50'
                        : 'border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600'"
                      class="flex flex-col items-center p-4 border rounded-lg transition-all duration-200 group">
                      <div
                        class="w-12 h-12 mb-2 rounded-lg bg-gradient-to-br from-gray-100 to-gray-800 border border-gray-300 dark:border-gray-600 flex items-center justify-center relative overflow-hidden">
                        <div class="absolute top-0 left-0 w-1/2 h-full bg-gray-100"></div>
                        <div class="absolute top-0 right-0 w-1/2 h-full bg-gray-800"></div>
                        <i class="pi pi-desktop text-gray-600 dark:text-gray-300 text-lg relative z-10"></i>
                      </div>
                      <span class="text-sm font-medium text-gray-900 dark:text-white">System</span>
                      <span class="text-xs text-gray-500 dark:text-gray-200 mt-1">Follow device</span>
                    </button>
                  </div>
                  <p class="text-xs text-gray-500 dark:text-gray-200 mt-2">
                    Choose how Gemmie appears. System theme follows your device's dark/light mode.
                  </p>
                </div>

                <!-- Compact Theme Selection -->
                <div class="md:hidden">
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">
                    Theme Preference
                  </label>
                  <div class="flex gap-2 p-1 bg-gray-100 dark:bg-gray-700 rounded-lg w-fit">
                    <button type="button" @click="toggleTheme('light')" :class="parsedUserDetails?.theme === 'light'
                      ? 'bg-white dark:bg-gray-600 text-blue-600 shadow-sm'
                      : 'text-gray-600 dark:text-gray-200 hover:text-gray-900 dark:hover:text-white'"
                      class="flex items-center gap-2 px-4 py-2 rounded-md transition-all duration-200 text-sm font-medium">
                      <i class="pi pi-sun"></i>
                      <span>Light</span>
                    </button>

                    <button type="button" @click="toggleTheme('dark')" :class="parsedUserDetails?.theme === 'dark'
                      ? 'bg-white dark:bg-gray-600 text-blue-600 shadow-sm'
                      : 'text-gray-600 dark:text-gray-200 hover:text-gray-900 dark:hover:text-white'"
                      class="flex items-center gap-2 px-4 py-2 rounded-md transition-all duration-200 text-sm font-medium">
                      <i class="pi pi-moon"></i>
                      <span>Dark</span>
                    </button>

                    <button type="button" @click="toggleTheme('system')" :class="parsedUserDetails?.theme === 'system'
                      ? 'bg-white dark:bg-gray-600 text-blue-600 shadow-sm'
                      : 'text-gray-600 dark:text-gray-200 hover:text-gray-900 dark:hover:text-white'"
                      class="flex items-center gap-2 px-4 py-2 rounded-md transition-all duration-200 text-sm font-medium">
                      <i class="pi pi-desktop"></i>
                      <span>System</span>
                    </button>
                  </div>
                  <p class="text-xs text-gray-500 dark:text-gray-200 mt-2">
                    Choose how Gemmie appears. System theme follows your device's settings.
                  </p>
                </div>

                <!-- Work Function -->
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    What best describes your work?
                  </label>
                  <select v-model="profileData.workFunction"
                    class="border border-gray-300 dark:border-gray-600 rounded-lg px-3 md:px-4 py-2 w-full text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200 bg-white dark:bg-gray-700 text-gray-900 dark:text-white">
                    <option value="">Select your work function</option>
                    <option value="software-developer">Software Developer</option>
                    <option value="designer">Designer</option>
                    <option value="researcher">Researcher</option>
                    <option value="student">Student</option>
                    <option value="writer">Writer</option>
                    <option value="teacher">Teacher/Educator</option>
                    <option value="business">Business Professional</option>
                    <option value="healthcare">Healthcare</option>
                    <option value="other">Other</option>
                  </select>
                </div>

                <!-- Preferences -->
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    What personal preferences should Gemmie consider in responses?
                    <span
                      class="ml-1 text-xs text-orange-600 dark:text-orange-400 bg-orange-100 dark:bg-orange-900/30 px-2 py-0.5 rounded">Beta</span>
                  </label>
                  <textarea v-model="profileData.preferences" rows="3"
                    class="border border-gray-300 dark:border-gray-600 rounded-lg px-3 md:px-4 py-2 w-full text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200 bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400"
                    placeholder="e.g., Be concise, use technical explanations, avoid jargon" />
                  <p class="text-xs text-gray-500 dark:text-gray-200 mt-1">
                    Your preferences will apply to all conversations, within guidelines.
                  </p>
                </div>

                <!-- Save Button -->
                <div class="flex justify-end">
                  <button type="submit" :disabled="isSaving || !hasUnsavedChanges"
                    class="px-6 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 dark:disabled:bg-gray-600 disabled:cursor-not-allowed text-white rounded-lg shadow-sm transition-all duration-200 flex items-center gap-2">
                    <i v-if="isSaving" class="pi pi-spin pi-spinner"></i>
                    <span>{{ isSaving ? 'Saving...' : 'Save changes' }}</span>
                  </button>
                </div>
              </form>
            </div>

            <!-- Account -->
            <div v-if="activeTab === 'account'"
              class="bg-white dark:bg-gray-800 p-4 md:p-6 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 w-full transition-colors duration-300"
              :class="[
                screenWidth >= 1536 ? 'max-w-4xl' : 
                screenWidth >= 1280 ? 'max-w-3xl' : 
                screenWidth >= 1024 ? 'max-w-2xl' : 
                'max-w-none'
              ]">
              <h2 class="text-xl font-medium mb-6 text-gray-900 dark:text-white">Account</h2>

              <div class="space-y-6">
                <!-- Logout -->
                <div class="flex flex-wrap gap-3 items-center justify-between">
                  <div>
                    <h3 class="text-sm font-medium text-gray-800 dark:text-gray-200">Log out of all devices</h3>
                    <p class="text-xs text-gray-500 dark:text-gray-200">This will sign you out everywhere</p>
                  </div>
                  <button @click="logout"
                    class="px-4 py-2 font-medium text-gray-700 dark:text-gray-300 bg-none border dark:border-gray-700 rounded-lg transition-all duration-200">
                    Log out
                  </button>
                </div>

                <!-- Delete account -->
                <div class="flex flex-wrap gap-3 items-center justify-between">
                  <div>
                    <h3 class="text-sm font-medium text-gray-800 dark:text-gray-200">Delete your account</h3>
                    <p class="text-xs text-gray-500 dark:text-gray-200">Permanently delete your account and all data
                    </p>
                  </div>
                  <button @click="router.push('/auth/delete_account')"
                    class="px-4 py-2 bg-black dark:bg-white font-medium text-white dark:text-black rounded-lg transition-all duration-200">
                    Delete account
                  </button>
                </div>

                <!-- Session ID -->
                <div class="space-y-2">
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">Session ID</label>
                  <input type="text" :value="parsedUserDetails?.sessionId" readonly
                    class="w-full px-4 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded-lg bg-gray-50 dark:bg-gray-700 text-gray-600 dark:text-gray-200 font-mono transition-colors" />
                </div>
              </div>
            </div>

            <!-- Privacy -->
            <div v-if="activeTab === 'privacy'"
              class="bg-white dark:bg-gray-800 p-4 md:p-6 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 w-full transition-colors duration-300"
              :class="[
                screenWidth >= 1536 ? 'max-w-4xl' : 
                screenWidth >= 1280 ? 'max-w-3xl' : 
                screenWidth >= 1024 ? 'max-w-2xl' : 
                'max-w-none'
              ]">
              <h2 class="text-xl font-medium mb-6 text-gray-900 dark:text-white">Privacy</h2>

              <div class="space-y-6">
                <!-- Sync Toggle -->
                <div class="flex items-center justify-between">
                  <div>
                    <h3 class="text-sm font-medium text-gray-800 dark:text-gray-200">Auto Sync</h3>
                    <p class="text-xs max-w-[150px] text-gray-500 dark:text-gray-200">
                      {{ parsedUserDetails?.syncEnabled ? 'Data is synced across all your devices automatically' :
                        'Data is only stored locally on this device' }}
                    </p>
                  </div>
                  <button @click="handleToggleSync" :disabled="isTogglingSync"
                    class="disabled:opacity-50 disabled:cursor-not-allowed relative inline-flex h-6 w-11 items-center rounded-full transition-colors duration-200"
                    :class="parsedUserDetails?.syncEnabled ? 'bg-blue-600' : 'bg-gray-300 dark:bg-gray-600'">
                    <span class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform duration-200"
                      :class="parsedUserDetails?.syncEnabled ? 'translate-x-6' : 'translate-x-1'" />
                  </button>
                </div>

                <!-- Manual Sync Button (only show if sync is enabled) -->
                <div v-if="parsedUserDetails?.syncEnabled" class="flex flex-wrap gap-3 items-center justify-between">
                  <div>
                    <h3 class="text-sm font-medium text-gray-800 dark:text-gray-200">Manual Sync</h3>
                    <p class="text-xs text-gray-500 dark:text-gray-200">Force sync your data now</p>
                  </div>
                  <button @click="manualSync" :disabled="syncStatus.syncing"
                    class="px-4 py-2 border font-medium border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700 disabled:bg-gray-100 dark:disabled:bg-gray-800 disabled:cursor-not-allowed rounded-lg transition-all duration-200 flex items-center gap-2 text-gray-700 dark:text-gray-300">
                    <i v-if="syncStatus.syncing" class="pi pi-spin pi-spinner text-sm"></i>
                    <span>{{ syncStatus.syncing ? 'Syncing...' : 'Sync Now' }}</span>
                  </button>
                </div>

                <!-- Sync Status -->
                <div v-if="parsedUserDetails?.syncEnabled" class="space-y-2">
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">Sync Status</label>
                  <div class="text-sm text-gray-600 dark:text-gray-200">
                    <div class="flex items-center gap-2">
                      <div :class="syncStatus.syncing ? 'bg-yellow-500' :
                        syncStatus.hasUnsyncedChanges ? 'bg-orange-500' : 'bg-green-500'" class="w-2 h-2 rounded-full">
                      </div>
                      <span>
                        {{ syncStatus.syncing ? 'Syncing...' :
                          syncStatus.hasUnsyncedChanges ? 'Unsynced changes' : 'Synced' }}
                      </span>
                    </div>
                    <div v-if="syncStatus.lastSync" class="text-xs text-gray-500 dark:text-gray-500 mt-1">
                      Last sync: {{ new Date(syncStatus.lastSync).toLocaleString() }}
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Billing -->
            <div v-if="activeTab === 'billing'"
              class="bg-white dark:bg-gray-800 p-4 md:p-6 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 w-full transition-colors duration-300"
              :class="[
                screenWidth >= 1536 ? 'max-w-4xl' : 
                screenWidth >= 1280 ? 'max-w-3xl' : 
                screenWidth >= 1024 ? 'max-w-2xl' : 
                'max-w-none'
              ]">
              <h2 class="text-xl font-medium mb-6 text-gray-900 dark:text-white">Billing</h2>

              <!-- Show M-Pesa number if available -->
              <div v-if="parsedUserDetails?.phoneNumber" class="mb-6">
                <div
                  class="flex items-center justify-between p-4 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg">
                  <div class="flex items-center gap-3">
                    <div class="w-10 h-10 bg-green-100 dark:bg-green-800 rounded-full flex items-center justify-center">
                      <i class="pi pi-mobile text-green-600 dark:text-green-400"></i>
                    </div>
                    <div>
                      <h3 class="text-sm font-medium text-gray-900 dark:text-white">M-Pesa Number</h3>
                      <p class="text-sm text-gray-600 dark:text-gray-300">{{ parsedUserDetails.phoneNumber }}</p>
                    </div>
                  </div>
                  <div
                    class="flex items-center gap-1 text-xs text-green-700 dark:text-green-400 bg-green-100 dark:bg-green-800 px-2 py-1 rounded">
                    <i class="pi pi-check-circle"></i>
                    <span>Verified</span>
                  </div>
                </div>
              </div>
              <div v-if="!parsedUserDetails?.phoneNumber" class="text-center py-12">
                <i class="pi pi-credit-card text-4xl text-gray-300 dark:text-gray-600 mb-4"></i>
                <p class="text-gray-600 dark:text-gray-200 mb-2">No billing information available</p>
                <p class="text-sm text-gray-500 dark:text-gray-500">Your billing details will appear here when you
                  upgrade to a paid plan
                </p>
              </div>

              <!-- If phone number exists but no other billing info -->
              <div v-else class="space-y-6">
                <!-- Current Plan -->
                <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-4">
                  <div class="flex items-center justify-between">
                    <div>
                      <h3 class="font-medium text-gray-900 dark:text-white">
                        Current Plan
                      </h3>
                      <p class="text-sm text-gray-600 dark:text-gray-200 capitalize">{{ parsedUserDetails?.plan ||
                        'Free' }} Plan</p>
                      <h3
                        :class="planStatus?.isExpired ? 'font-medium text-red-900 dark:text-red-400 capitalize' : 'font-medium text-green-500 dark:text-green-400 capitalize'">
                        {{ planStatus?.status }}</h3>
                    </div>
                    <button v-if="planStatus?.isExpired || parsedUserDetails?.plan === ''"
                      @click="router.push('/upgrade')"
                      class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm font-medium transition-colors duration-200">
                      Manage Plan
                    </button>
                    <div v-else class="text-sm text-gray-600 dark:text-gray-200">
                      Expires on {{ new Date(planStatus?.expiryDate || '').toLocaleDateString() }}
                    </div>
                  </div>
                </div>

                <!-- Payment History -->
                <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-4">
                  <h3 class="font-medium text-gray-900 dark:text-white mb-3">Payment History</h3>

                  <!-- No Transactions -->
                  <div v-if="!parsedUserDetails?.userTransactions || parsedUserDetails.userTransactions.length === 0"
                    class="text-center py-8">
                    <i class="pi pi-history text-2xl text-gray-300 dark:text-gray-600 mb-2"></i>
                    <p class="text-sm text-gray-500 dark:text-gray-200">No payment history available</p>
                  </div>

                  <!-- Transactions List -->
                  <div v-else class="space-y-3">
                    <details v-for="transaction in parsedUserDetails.userTransactions" :key="transaction.id"
                      class="border border-gray-200 dark:border-gray-700 rounded-lg p-2 hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors cursor-pointer group">

                      <!-- Summary (Always Visible) -->
                      <summary class="flex items-start justify-between select-none font-medium">
                        <div class="flex items-center gap-3 flex-1">
                          <!-- Status Icon -->
                          <div :class="transaction.Status === 'Success'
                            ? 'bg-green-100 dark:bg-green-900/30'
                            : transaction.Status === 'Pending'
                              ? 'bg-yellow-100 dark:bg-yellow-900/30'
                              : 'bg-red-100 dark:bg-red-900/30'"
                            class="w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0">
                            <i :class="transaction.Status === 'Success'
                              ? 'pi pi-check-circle text-green-600 dark:text-green-400'
                              : transaction.Status === 'Pending'
                                ? 'pi pi-clock text-yellow-600 dark:text-yellow-400'
                                : 'pi pi-times-circle text-red-600 dark:text-red-400'"></i>
                          </div>

                          <!-- Transaction Info -->
                          <div class="flex-1 min-w-0">
                            <p class="text-sm text-gray-900 dark:text-white">
                              Payment
                              <span :class="transaction.Status === 'Success'
                                ? 'text-green-600 dark:text-green-400'
                                : transaction.Status === 'Pending'
                                  ? 'text-yellow-600 dark:text-yellow-400'
                                  : 'text-red-600 dark:text-red-400'" class="font-semibold">
                                {{ transaction.Status }}
                              </span>
                            </p>
                            <p class="text-xs text-gray-500 dark:text-gray-200">
                              {{ new Date(transaction.CreatedAt).toLocaleString() }}
                            </p>
                          </div>
                        </div>

                        <!-- Amount -->
                        <div class="text-right flex-shrink-0 ml-2">
                          <p class="font-semibold text-gray-900 dark:text-white text-sm">
                            KES {{ transaction.Amount }}
                          </p>
                        </div>
                      </summary>

                      <!-- Details (Collapsible) -->
                      <div class="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700 space-y-2">
                        <!-- M-Pesa Receipt -->
                        <div class="flex justify-between text-xs">
                          <span class="text-gray-600 dark:text-gray-200">M-Pesa Receipt</span>
                          <span class="font-mono text-gray-900 dark:text-white font-medium">
                            {{ transaction.MpesaReceiptNumber }}
                          </span>
                        </div>

                        <!-- Phone Number -->
                        <div class="flex justify-between text-xs">
                          <span class="text-gray-600 dark:text-gray-200">Phone Number</span>
                          <span class="font-mono text-gray-900 dark:text-white font-medium">
                            {{ transaction.Phone }}
                          </span>
                        </div>

                        <!-- Checkout Request ID -->
                        <div class="flex justify-between text-xs">
                          <span class="text-gray-600 dark:text-gray-200">Checkout ID</span>
                          <span class="font-mono text-gray-900 dark:text-white font-medium truncate max-w-[200px]">
                            {{ transaction.CheckoutRequestID }}
                          </span>
                        </div>

                        <!-- Merchant Request ID -->
                        <div class="flex justify-between text-xs">
                          <span class="text-gray-600 dark:text-gray-200">Merchant ID</span>
                          <span class="font-mono text-gray-900 dark:text-white font-medium truncate max-w-[200px]">
                            {{ transaction.MerchantRequestID }}
                          </span>
                        </div>

                        <!-- External Reference -->
                        <div class="flex justify-between text-xs">
                          <span class="text-gray-600 dark:text-gray-200">Reference</span>
                          <span class="font-mono text-gray-900 dark:text-white font-medium">
                            {{ transaction.ExternalReference }}
                          </span>
                        </div>

                        <!-- Result Description -->
                        <div class="flex justify-between text-xs">
                          <span class="text-gray-600 dark:text-gray-200">Result</span>
                          <span class="text-gray-900 dark:text-white">
                            {{ transaction.ResultDesc }}
                          </span>
                        </div>

                        <!-- Updated Timestamp -->
                        <div class="flex justify-between text-xs pt-2 border-t border-gray-200 dark:border-gray-700">
                          <span class="text-gray-600 dark:text-gray-200">Updated</span>
                          <span class="text-gray-900 dark:text-white">
                            {{ new Date(transaction.UpdatedAt).toLocaleString() }}
                          </span>
                        </div>
                      </div>
                    </details>
                  </div>
                </div>

                <!-- Billing Information -->
                <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-4">
                  <h3 class="font-medium text-gray-900 dark:text-white mb-3">Billing Information</h3>
                  <div class="space-y-3">
                    <div class="flex justify-between text-sm">
                      <span class="text-gray-600 dark:text-gray-200">Payment Method</span>
                      <span class="font-medium text-gray-900 dark:text-white">M-Pesa</span>
                    </div>
                    <div class="flex justify-between text-sm">
                      <span class="text-gray-600 dark:text-gray-200">Phone Number</span>
                      <span class="font-medium text-gray-900 dark:text-white">{{ parsedUserDetails.phoneNumber
                      }}</span>
                    </div>
                    <div class="flex justify-between text-sm">
                      <span class="text-gray-600 dark:text-gray-200">Currency</span>
                      <span class="font-medium text-gray-900 dark:text-white">KES</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>