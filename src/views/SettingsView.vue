<script lang="ts" setup>
import router from '@/router';
import type { Chat, ConfirmDialogOptions } from '@/types';
import { inject, ref, reactive, type Ref, computed, watch, onMounted } from 'vue';
import SideNav from '@/components/SideNav.vue';
import { toast } from 'vue-sonner';
import { useRoute } from 'vue-router';
import type { Theme } from 'vue-sonner/src/packages/types.js';

const globalState = inject('globalState') as {
  screenWidth: Ref<number>,
  confirmDialog: Ref<ConfirmDialogOptions>,
  isCollapsed: Ref<boolean>,
  isSidebarHidden: Ref<boolean>,
  authData: Ref<{
    username: string;
    email: string;
    password: string;
    workFunction?: string,
    preferences?: string
  }>,
  syncStatus: Ref<{ lastSync: Date | null; syncing: boolean; hasUnsyncedChanges: boolean; }>,
  isAuthenticated: Ref<boolean>,
  parsedUserDetails: Ref<any>,
  currentChatId: Ref<string>,
  chats: Ref<Chat[]>
  logout: () => void,
  isLoading: Ref<boolean>,
  showInput: Ref<boolean>,
  planStatus: Ref<{ status: string; timeLeft: string; expiryDate: string; isExpired: boolean; }>,
  isDarkMode: Ref<boolean>,
  currentTheme: Ref<Theme>,
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
  authData,
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
  apiCall,
  isDarkMode,
  currentTheme,
  toggleTheme,
  isLocalDataEmpty,
  toggleSync,
  syncEnabled, // Add this
} = globalState

const route = useRoute()
const parsedUserDetails = globalState.parsedUserDetails
const tabParam = route.params.tab as 'profile' | 'account' | 'billing' | undefined

// Local form state
const profileData = reactive({
  username: parsedUserDetails.value?.username || '',
  email: parsedUserDetails.value?.email || '',
  workFunction: parsedUserDetails.value?.workFunction || '',
  preferences: parsedUserDetails.value?.preferences || ''
})

const activeTab = ref<'profile' | 'account' | 'billing'>(tabParam ?? 'profile')
const isSaving = ref(false)

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

    // Always save to localStorage
    localStorage.setItem("userdetails", JSON.stringify(parsedUserDetails.value))

    // Server sync only if enabled
    if (parsedUserDetails?.value.sync_enabled) {
      try {
        await apiCall('/sync', {
          method: 'POST',
          body: JSON.stringify({
            workFunction: profileData.workFunction,
            preferences: profileData.preferences,
            chats: JSON.stringify(chats.value),
            link_previews: "{}",
            current_chat_id: currentChatId.value,
            sync_enabled: parsedUserDetails?.value.sync_enabled
          })
        })
        toast.success('Profile updated and synced successfully')
      } catch (serverError) {
        console.warn('Failed to sync profile to server:', serverError)
        toast.success('Profile updated locally (sync failed)')
      }
    } else {
      toast.success('Profile updated locally')
    }

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

// Fix the watch function - watch the actual sync_enabled property
watch(() => parsedUserDetails.value?.sync_enabled, (newVal) => {
  console.log('Sync enabled changed:', newVal)
  // The toggleSync function will handle the UI updates
}, { deep: true })

// Lifecycle hooks
onMounted(() => {
  if (parsedUserDetails.value) {
    resetProfileData()
  } else if (isAuthenticated.value) {
    logout()
    router.push('/')
  }
  if (!isAuthenticated.value) {
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
  <div class="flex h-[100vh] bg-gray-50 dark:bg-gray-900 transition-colors duration-300">
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
      manualSync
    }" />

    <!-- Main Content -->
    <div :class="screenWidth > 720 && isAuthenticated
      ? (!isCollapsed
        ? 'flex-grow flex flex-col ml-[270px] font-light text-sm transition-all duration-300 ease-in-out'
        : 'flex-grow flex flex-col ml-[60px] font-light text-sm transition-all duration-300 ease-in-out')
      : 'text-sm font-light flex-grow flex flex-col transition-all duration-300 ease-in-out'">

      <div class="flex flex-col p-4 md:p-6 bg-gray-50 dark:bg-gray-900 min-h-0 flex-1">
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
            <button @click="activeTab = 'profile'"
              :class="activeTab === 'profile'
                ? 'border-blue-500 text-blue-600 dark:text-blue-400 whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm'
                : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300 dark:hover:border-gray-600 whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm transition-colors'">
              Profile
            </button>
            <button @click="activeTab = 'account'"
              :class="activeTab === 'account'
                ? 'border-blue-500 text-blue-600 dark:text-blue-400 whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm'
                : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300 dark:hover:border-gray-600 whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm transition-colors'">
              Account
            </button>
            <button @click="activeTab = 'billing'"
              :class="activeTab === 'billing'
                ? 'border-blue-500 text-blue-600 dark:text-blue-400 whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm'
                : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300 dark:hover:border-gray-600 whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm transition-colors'">
              Billing
            </button>
          </nav>
        </div>

        <div class="flex flex-col md:flex-row gap-4 md:gap-8 flex-1 min-h-0">
          <!-- Tabs Sidebar (hidden on mobile) -->
          <div class="w-48 flex-col gap-2 hidden md:flex flex-shrink-0">
            <button @click="activeTab = 'profile'" :class="activeTab === 'profile'
              ? 'bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400 border-l-4 border-blue-600 dark:border-blue-400 font-medium'
              : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800'"
              class="text-left px-4 py-2 rounded-md transition-all duration-200">
              Profile
            </button>
            <button @click="activeTab = 'account'" :class="activeTab === 'account'
              ? 'bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400 border-l-4 border-blue-600 dark:border-blue-400 font-medium'
              : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800'"
              class="text-left px-4 py-2 rounded-md transition-all duration-200">
              Account
            </button>
            <button @click="activeTab = 'billing'" :class="activeTab === 'billing'
              ? 'bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400 border-l-4 border-blue-600 dark:border-blue-400 font-medium'
              : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800'"
              class="text-left px-4 py-2 rounded-md transition-all duration-200">
              Billing
            </button>
          </div>

          <!-- Content Area -->
          <div class="flex-1 min-h-0 overflow-auto">
            <!-- Profile -->
            <div v-if="activeTab === 'profile'"
              class="bg-white dark:bg-gray-800 p-4 md:p-6 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 w-full max-w-2xl mx-auto transition-colors duration-300">
              <h2 class="text-xl font-medium mb-4 md:mb-6 text-gray-900 dark:text-white">Profile</h2>

              <form @submit.prevent="saveProfile" class="flex flex-col gap-4 md:gap-6">
                <!-- Username -->
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Your username
                  </label>
                  <input v-model="profileData.username" type="text" disabled
                    class="border border-gray-300 dark:border-gray-600 rounded-lg px-3 md:px-4 py-2 w-full text-sm bg-gray-50 dark:bg-gray-700 text-gray-600 dark:text-gray-400 cursor-not-allowed transition-colors" />
                  <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">Username cannot be changed</p>
                </div>

                <!-- Email -->
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Your email</label>
                  <input v-model="profileData.email" type="email" disabled
                    class="border border-gray-300 dark:border-gray-600 rounded-lg px-3 md:px-4 py-2 w-full text-sm bg-gray-50 dark:bg-gray-700 text-gray-600 dark:text-gray-400 cursor-not-allowed transition-colors" />
                  <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">Email cannot be changed</p>
                </div>

                <!-- Theme Selection - THIS IS THE DESKTOP VERSION -->
                <div class="max-md:hidden">
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">
                    Theme Preference
                  </label>
                  <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
                    <!-- Light Theme -->
                    <button type="button" @click="toggleTheme('light')"
                      :class="currentTheme === 'light'
                        ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20 ring-2 ring-blue-500 ring-opacity-50'
                        : 'border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600'"
                      class="flex flex-col items-center p-4 border rounded-lg transition-all duration-200 group">
                      <div
                        class="w-12 h-12 mb-2 rounded-lg bg-gradient-to-br from-gray-100 to-gray-200 border border-gray-300 flex items-center justify-center">
                        <i class="pi pi-sun text-yellow-500 text-lg"></i>
                      </div>
                      <span class="text-sm font-medium text-gray-900 dark:text-white">Light</span>
                      <span class="text-xs text-gray-500 dark:text-gray-400 mt-1">Always light</span>
                    </button>

                    <!-- Dark Theme -->
                    <button type="button" @click="toggleTheme('dark')"
                      :class="currentTheme === 'dark'
                        ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20 ring-2 ring-blue-500 ring-opacity-50'
                        : 'border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600'"
                      class="flex flex-col items-center p-4 border rounded-lg transition-all duration-200 group">
                      <div
                        class="w-12 h-12 mb-2 rounded-lg bg-gradient-to-br from-gray-800 to-gray-900 border border-gray-700 flex items-center justify-center">
                        <i class="pi pi-moon text-blue-300 text-lg"></i>
                      </div>
                      <span class="text-sm font-medium text-gray-900 dark:text-white">Dark</span>
                      <span class="text-xs text-gray-500 dark:text-gray-400 mt-1">Always dark</span>
                    </button>

                    <!-- System Theme -->
                    <button type="button" @click="toggleTheme('system')"
                      :class="currentTheme === 'system'
                        ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20 ring-2 ring-blue-500 ring-opacity-50'
                        : 'border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600'"
                      class="flex flex-col items-center p-4 border rounded-lg transition-all duration-200 group">
                      <div
                        class="w-12 h-12 mb-2 rounded-lg bg-gradient-to-br from-gray-100 to-gray-800 border border-gray-300 dark:border-gray-600 flex items-center justify-center relative overflow-hidden">
                        <div class="absolute top-0 left-0 w-1/2 h-full bg-gray-100"></div>
                        <div class="absolute top-0 right-0 w-1/2 h-full bg-gray-800"></div>
                        <i class="pi pi-desktop text-gray-600 dark:text-gray-300 text-lg relative z-10"></i>
                      </div>
                      <span class="text-sm font-medium text-gray-900 dark:text-white">System</span>
                      <span class="text-xs text-gray-500 dark:text-gray-400 mt-1">Follow device</span>
                    </button>
                  </div>
                  <p class="text-xs text-gray-500 dark:text-gray-400 mt-2">
                    Choose how Gemmie appears. System theme follows your device's dark/light mode.
                  </p>
                </div>

                <!-- Compact Theme Selection -->
                <div class="md:hidden">
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">
                    Theme Preference
                  </label>
                  <div class="flex gap-2 p-1 bg-gray-100 dark:bg-gray-700 rounded-lg w-fit">
                    <button type="button" @click="toggleTheme('light')" :class="currentTheme === 'light'
                      ? 'bg-white dark:bg-gray-600 text-blue-600 shadow-sm'
                      : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white'"
                      class="flex items-center gap-2 px-4 py-2 rounded-md transition-all duration-200 text-sm font-medium">
                      <i class="pi pi-sun"></i>
                      <span>Light</span>
                    </button>

                    <button type="button" @click="toggleTheme('dark')" :class="currentTheme === 'dark'
                      ? 'bg-white dark:bg-gray-600 text-blue-600 shadow-sm'
                      : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white'"
                      class="flex items-center gap-2 px-4 py-2 rounded-md transition-all duration-200 text-sm font-medium">
                      <i class="pi pi-moon"></i>
                      <span>Dark</span>
                    </button>

                    <button type="button" @click="toggleTheme('system')" :class="currentTheme === 'system'
                      ? 'bg-white dark:bg-gray-600 text-blue-600 shadow-sm'
                      : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white'"
                      class="flex items-center gap-2 px-4 py-2 rounded-md transition-all duration-200 text-sm font-medium">
                      <i class="pi pi-desktop"></i>
                      <span>System</span>
                    </button>
                  </div>
                  <p class="text-xs text-gray-500 dark:text-gray-400 mt-2">
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
                  <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
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
              class="bg-white dark:bg-gray-800 p-4 md:p-6 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 w-full max-w-2xl mx-auto transition-colors duration-300">
              <h2 class="text-xl font-medium mb-6 text-gray-900 dark:text-white">Account</h2>

              <div class="space-y-6">
                <!-- Sync Toggle -->
                <div class="flex items-center justify-between">
                  <div>
                    <h3 class="text-sm font-medium text-gray-800 dark:text-gray-200">Auto Sync</h3>
                    <p class="text-xs max-w-[150px] text-gray-500 dark:text-gray-400">
                      {{ parsedUserDetails?.sync_enabled ? 'Data is synced across all your devices automatically' :
                        'Data is only stored locally on this device' }}
                    </p>
                  </div>
                  <button @click="() => toggleSync()"
                    class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors duration-200"
                    :class="parsedUserDetails?.sync_enabled ? 'bg-blue-600' : 'bg-gray-300 dark:bg-gray-600'">
                    <span class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform duration-200"
                      :class="parsedUserDetails?.sync_enabled ? 'translate-x-6' : 'translate-x-1'" />
                  </button>
                </div>

                <!-- Manual Sync Button (only show if sync is enabled) -->
                <div v-if="parsedUserDetails?.sync_enabled" class="flex flex-wrap gap-3 items-center justify-between">
                  <div>
                    <h3 class="text-sm font-medium text-gray-800 dark:text-gray-200">Manual Sync</h3>
                    <p class="text-xs text-gray-500 dark:text-gray-400">Force sync your data now</p>
                  </div>
                  <button @click="manualSync" :disabled="syncStatus.syncing"
                    class="px-4 py-2 border font-medium border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700 disabled:bg-gray-100 dark:disabled:bg-gray-800 disabled:cursor-not-allowed rounded-lg transition-all duration-200 flex items-center gap-2 text-gray-700 dark:text-gray-300">
                    <i v-if="syncStatus.syncing" class="pi pi-spin pi-spinner text-sm"></i>
                    <span>{{ syncStatus.syncing ? 'Syncing...' : 'Sync Now' }}</span>
                  </button>
                </div>

                <!-- Logout -->
                <div class="flex flex-wrap gap-3 items-center justify-between">
                  <div>
                    <h3 class="text-sm font-medium text-gray-800 dark:text-gray-200">Log out of all devices</h3>
                    <p class="text-xs text-gray-500 dark:text-gray-400">This will sign you out everywhere</p>
                  </div>
                  <button @click="logout"
                    class="px-4 py-2 font-medium text-white bg-yellow-500 hover:bg-yellow-600 rounded-lg transition-all duration-200">
                    Log out
                  </button>
                </div>

                <!-- Delete account -->
                <div class="flex flex-wrap gap-3 items-center justify-between">
                  <div>
                    <h3 class="text-sm font-medium text-gray-800 dark:text-gray-200">Delete your account</h3>
                    <p class="text-xs text-gray-500 dark:text-gray-400">Permanently delete your account and all data
                    </p>
                  </div>
                  <button @click="router.push('/auth/delete_account')"
                    class="px-4 py-2 bg-red-600 font-medium hover:bg-red-700 text-white rounded-lg transition-all duration-200">
                    Delete account
                  </button>
                </div>

                <!-- Session ID -->
                <div class="space-y-2">
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">Session ID</label>
                  <input type="text" :value="parsedUserDetails?.sessionId" readonly
                    class="w-full px-4 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded-lg bg-gray-50 dark:bg-gray-700 text-gray-600 dark:text-gray-400 font-mono transition-colors" />
                </div>

                <!-- Sync Status -->
                <div v-if="parsedUserDetails?.sync_enabled" class="space-y-2">
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">Sync Status</label>
                  <div class="text-sm text-gray-600 dark:text-gray-400">
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
              class="bg-white dark:bg-gray-800 p-4 md:p-6 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 w-full max-w-2xl mx-auto transition-colors duration-300">
              <h2 class="text-xl font-medium mb-6 text-gray-900 dark:text-white">Billing</h2>

              <!-- Show M-Pesa number if available -->
              <div v-if="parsedUserDetails?.phone_number" class="mb-6">
                <div
                  class="flex items-center justify-between p-4 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg">
                  <div class="flex items-center gap-3">
                    <div class="w-10 h-10 bg-green-100 dark:bg-green-800 rounded-full flex items-center justify-center">
                      <i class="pi pi-mobile text-green-600 dark:text-green-400"></i>
                    </div>
                    <div>
                      <h3 class="text-sm font-medium text-gray-900 dark:text-white">M-Pesa Number</h3>
                      <p class="text-sm text-gray-600 dark:text-gray-300">{{ parsedUserDetails.phone_number }}</p>
                    </div>
                  </div>
                  <div
                    class="flex items-center gap-1 text-xs text-green-700 dark:text-green-400 bg-green-100 dark:bg-green-800 px-2 py-1 rounded">
                    <i class="pi pi-check-circle"></i>
                    <span>Verified</span>
                  </div>
                </div>
              </div>
              <div v-if="!parsedUserDetails?.phone_number" class="text-center py-12">
                <i class="pi pi-credit-card text-4xl text-gray-300 dark:text-gray-600 mb-4"></i>
                <p class="text-gray-600 dark:text-gray-400 mb-2">No billing information available</p>
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
                      <p class="text-sm text-gray-600 dark:text-gray-400 capitalize">{{ parsedUserDetails?.plan ||
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
                    <div v-else class="text-sm text-gray-600 dark:text-gray-400">
                      Expires on {{ new Date(planStatus?.expiryDate || '').toLocaleDateString() }}
                    </div>
                  </div>
                </div>

                <!-- Payment History (placeholder) -->
                <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-4">
                  <h3 class="font-medium text-gray-900 dark:text-white mb-3">Payment History</h3>
                  <div class="text-center py-8">
                    <i class="pi pi-history text-2xl text-gray-300 dark:text-gray-600 mb-2"></i>
                    <p class="text-sm text-gray-500 dark:text-gray-400">No payment history available</p>
                  </div>
                </div>

                <!-- Billing Information -->
                <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-4">
                  <h3 class="font-medium text-gray-900 dark:text-white mb-3">Billing Information</h3>
                  <div class="space-y-3">
                    <div class="flex justify-between text-sm">
                      <span class="text-gray-600 dark:text-gray-400">Payment Method</span>
                      <span class="font-medium text-gray-900 dark:text-white">M-Pesa</span>
                    </div>
                    <div class="flex justify-between text-sm">
                      <span class="text-gray-600 dark:text-gray-400">Phone Number</span>
                      <span class="font-medium text-gray-900 dark:text-white">{{ parsedUserDetails.phone_number
                        }}</span>
                    </div>
                    <div class="flex justify-between text-sm">
                      <span class="text-gray-600 dark:text-gray-400">Currency</span>
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