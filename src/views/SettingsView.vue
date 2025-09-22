<script lang="ts" setup>
import router from '@/router';
import type { Chat, ConfirmDialogOptions } from '@/types';
import { inject, ref, reactive, type Ref, computed, watch, onMounted } from 'vue';
import SideNav from '@/components/SideNav.vue';
import { toast } from 'vue-sonner';
import { useRoute } from 'vue-router';

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
  hideSidebar: () => void,
  setShowInput: () => void,
  clearAllChats: () => void,
  switchToChat: (chatId: string) => void,
  createNewChat: (initialMessage?: string) => void,
  deleteChat: (chatId: string) => void,
  renameChat: (chatId: string, newTitle: string) => void,
  toggleSidebar: () => void,
  manualSync: () => Promise<any>,
  apiCall: (endpoint: string, options?: RequestInit) => Promise<any>
}

const {
  screenWidth,
  isCollapsed,
  authData,
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
  apiCall
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

// Get sync setting from user details or localStorage, default to true
const syncEnabled = ref(
  parsedUserDetails.value?.sync_enabled ?? 
  (localStorage.getItem('syncEnabled') !== 'false')
)

const showDropdown = ref(false)
const activeTab = ref<'profile' | 'account' | 'billing'>(tabParam ?? 'profile')
const isSaving = ref(false)

/**
 * Save profile changes - only sync to server if sync is enabled
 */
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
    if (syncEnabled.value) {
      try {
        await apiCall('/sync', {
          method: 'POST',
          body: JSON.stringify({
            username: profileData.username.trim(),
            workFunction: profileData.workFunction,
            preferences: profileData.preferences,
            chats: JSON.stringify(chats.value),
            link_previews: "{}",
            current_chat_id: currentChatId.value,
            sync_enabled: syncEnabled.value
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

/**
 * Toggle auto-sync - this affects all future sync operations
 */
async function toggleSync() {
  const newSyncValue = !syncEnabled.value
  
  try {
    // Update local state first
    syncEnabled.value = newSyncValue
    parsedUserDetails.value.sync_enabled = newSyncValue
    
    // Save to localStorage
    localStorage.setItem('syncEnabled', String(newSyncValue))
    localStorage.setItem("userdetails", JSON.stringify(parsedUserDetails.value))

    // If enabling sync, sync current data to server
    if (newSyncValue) {
      try {
        await apiCall('/sync', {
          method: 'POST',
          body: JSON.stringify({
            username: parsedUserDetails.value.username,
            workFunction: parsedUserDetails.value.workFunction || '',
            preferences: parsedUserDetails.value.preferences || '',
            chats: JSON.stringify(chats.value),
            link_previews: "{}",
            current_chat_id: currentChatId.value,
            sync_enabled: newSyncValue
          })
        })
        toast.success('Auto-sync enabled and data synced to server')
      } catch (error) {
        console.warn('Failed to sync to server:', error)
        toast.success('Auto-sync enabled (initial sync failed)')
      }
    } else {
      // When disabling sync, still notify the server about the preference change
      try {
        await apiCall('/sync', {
          method: 'POST',
          body: JSON.stringify({
            username: parsedUserDetails.value.username,
            workFunction: parsedUserDetails.value.workFunction || '',
            preferences: parsedUserDetails.value.preferences || '',
            chats: JSON.stringify(chats.value),
            link_previews: "{}",
            current_chat_id: currentChatId.value,
            sync_enabled: newSyncValue
          })
        })
      } catch (error) {
        console.warn('Failed to update sync preference on server:', error)
      }
      toast.info('Auto-sync disabled - data will only be saved locally')
    }

  } catch (error) {
    console.error('Failed to toggle sync:', error)
    // Revert the change on error
    syncEnabled.value = !newSyncValue
    parsedUserDetails.value.sync_enabled = !newSyncValue
    toast.error('Failed to update sync setting')
  }
}

/**
 * Reset form when switching tabs
 */
function resetProfileData() {
  profileData.username = parsedUserDetails.value?.username || ''
  profileData.email = parsedUserDetails.value?.email || ''
  profileData.workFunction = parsedUserDetails.value?.workFunction || ''
  profileData.preferences = parsedUserDetails.value?.preferences || ''
}

/**
 * Detect unsaved changes
 */
const hasUnsavedChanges = computed(() => {
  return profileData.username !== (parsedUserDetails.value?.username || '') ||
    profileData.workFunction !== (parsedUserDetails.value?.workFunction || '') ||
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

/**
 * Watch for changes in parsedUserDetails and update syncEnabled accordingly
 */
watch(parsedUserDetails, (newVal) => {
  if (newVal) {
    // Update sync setting from user details if it exists
    if (typeof newVal.sync_enabled === 'boolean') {
      syncEnabled.value = newVal.sync_enabled
    }
  }
}, { deep: true })

/**
 * Lifecycle hooks
 */
onMounted(() => {
  if (parsedUserDetails.value) {
    resetProfileData()
    // Ensure sync setting is in sync with user details
    if (typeof parsedUserDetails.value.sync_enabled === 'boolean') {
      syncEnabled.value = parsedUserDetails.value.sync_enabled
    }
  } else if (isAuthenticated.value) {
    logout()
    router.push('/')
  }
  if (!isAuthenticated.value) {
    router.push('/login')
  }
})

watch(isAuthenticated, (val) => {
  if (val === false) {
    router.push('/')
  }
})
</script>

<template>
    <div class="flex h-[100vh]">
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

            <div class="h-screen flex flex-col p-6 overflow-y-auto">
                <div class="flex items-center justify-between mb-5">
                    <h1 class="text-2xl font-semibold">Settings</h1>
                    <button @click="router.back()" title="Go Back"
                        class="md:hidden flex items-center justify-center w-8 h-8 hover:bg-gray-100 rounded-full cursor-pointer transition-colors">
                        <span class="pi pi-times text-lg text-gray-700"></span>
                    </button>
                </div>

                <!-- Mobile Dropdown Toggle -->
                <div class="relative mb-4 md:hidden">
                    <button @click="showDropdown = !showDropdown"
                        class="w-full flex justify-between font-medium items-center px-4 py-2 border border-gray-300 rounded-md bg-gray-50 shadow-sm text-sm focus:outline-none">
                        <span>
                            {{ activeTab === 'profile' ? 'Profile' : activeTab === 'account' ? 'Account' : 'Billing' }}
                        </span>
                        <i :class="showDropdown ? 'pi pi-chevron-up' : 'pi pi-chevron-down'" class="text-gray-500"></i>
                    </button>

                    <!-- Dropdown Menu -->
                    <div v-if="showDropdown" class="absolute mt-1 w-full bg-white border rounded-md shadow-lg z-10">
                        <button v-for="tab in ['profile', 'account', 'billing']" :key="tab"
                            @click="activeTab = tab as 'profile' | 'account' | 'billing'; showDropdown = false" :class="activeTab === tab
                                ? 'w-full text-left px-4 py-2 bg-blue-50 text-blue-600 font-medium'
                                : 'w-full text-left px-4 py-2 hover:bg-gray-100'">
                            {{ tab.charAt(0).toUpperCase() + tab.slice(1) }}
                        </button>
                    </div>
                </div>

                <div class="flex flex-grow gap-8">
                    <!-- Tabs Sidebar (hidden on mobile) -->
                    <div class="w-48 flex-col gap-2 hidden md:flex flex-shrink-0">
                        <button @click="activeTab = 'profile'" :class="activeTab === 'profile'
                            ? 'bg-blue-50 text-blue-600 border-l-4 border-blue-600 font-medium'
                            : 'text-gray-700 hover:bg-gray-100'" class="text-left px-4 py-2 rounded-md transition-all">
                            Profile
                        </button>
                        <button @click="activeTab = 'account'" :class="activeTab === 'account'
                            ? 'bg-blue-50 text-blue-600 border-l-4 border-blue-600 font-medium'
                            : 'text-gray-700 hover:bg-gray-100'" class="text-left px-4 py-2 rounded-md transition-all">
                            Account
                        </button>
                        <button @click="activeTab = 'billing'" :class="activeTab === 'billing'
                            ? 'bg-blue-50 text-blue-600 border-l-4 border-blue-600 font-medium'
                            : 'text-gray-700 hover:bg-gray-100'" class="text-left px-4 py-2 rounded-md transition-all">
                            Billing
                        </button>
                    </div>

                    <!-- Content Area -->
                    <div class="flex-grow">
                        <!-- Profile -->
                        <div v-if="activeTab === 'profile'"
                            class="bg-white p-6 rounded-lg shadow-sm border w-full max-w-2xl">
                            <h2 class="text-xl font-medium mb-6">Profile</h2>

                            <form @submit.prevent="saveProfile" class="flex flex-col gap-6">
                                <!-- Username -->
                                <div>
                                    <label class="block text-sm font-medium text-gray-700 mb-1">
                                        What should Gemmie call you?
                                    </label>
                                    <input v-model="profileData.username" type="text" required
                                        class="border border-gray-300 rounded-lg px-4 py-2 w-full text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
                                        placeholder="Enter your preferred name" />
                                </div>

                                <!-- Email -->
                                <div>
                                    <label class="block text-sm font-medium text-gray-700 mb-1">Your email</label>
                                    <input v-model="profileData.email" type="email" disabled
                                        class="border border-gray-300 rounded-lg px-4 py-2 w-full text-sm bg-gray-50 text-gray-600 cursor-not-allowed" />
                                    <p class="text-xs text-gray-500 mt-1">Email cannot be changed</p>
                                </div>

                                <!-- Work Function -->
                                <div>
                                    <label class="block text-sm font-medium text-gray-700 mb-1">
                                        What best describes your work?
                                    </label>
                                    <select v-model="profileData.workFunction"
                                        class="border border-gray-300 rounded-lg px-4 py-2 w-full text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200">
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
                                    <label class="block text-sm font-medium text-gray-700 mb-1">
                                        What personal preferences should Gemmie consider in responses?
                                        <span
                                            class="ml-1 text-xs text-orange-600 bg-orange-100 px-2 py-0.5 rounded">Beta</span>
                                    </label>
                                    <textarea v-model="profileData.preferences" rows="3"
                                        class="border border-gray-300 rounded-lg px-4 py-2 w-full text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200"
                                        placeholder="e.g., Be concise, use technical explanations, avoid jargon" />
                                    <p class="text-xs text-gray-500 mt-1">
                                        Your preferences will apply to all conversations, within guidelines.
                                    </p>
                                </div>

                                <!-- Save Button -->
                                <div class="flex justify-end">
                                    <button type="submit" :disabled="isSaving || !hasUnsavedChanges"
                                        class="px-6 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed text-white rounded-lg shadow-sm transition-all flex items-center gap-2">
                                        <i v-if="isSaving" class="pi pi-spin pi-spinner"></i>
                                        <span>{{ isSaving ? 'Saving...' : 'Save changes' }}</span>
                                    </button>
                                </div>
                            </form>
                        </div>

                        <!-- Account -->
                        <div v-if="activeTab === 'account'"
                            class="bg-white p-6 rounded-lg shadow-sm border w-full max-w-2xl">
                            <h2 class="text-xl font-medium mb-6">Account</h2>

                            <div class="space-y-6">
                                <!-- Sync Toggle -->
                                <div class="flex items-center justify-between">
                                    <div>
                                        <h3 class="text-sm font-medium text-gray-800">Auto Sync</h3>
                                        <p class="text-xs text-gray-500">
                                            {{ syncEnabled ? 'Data is synced across all your devices automatically' :
                                            'Data is only stored locally on this device' }}
                                        </p>
                                    </div>
                                    <button @click="toggleSync"
                                        class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors"
                                        :class="syncEnabled ? 'bg-blue-600' : 'bg-gray-300'">
                                        <span
                                            class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform"
                                            :class="syncEnabled ? 'translate-x-6' : 'translate-x-1'" />
                                    </button>
                                </div>

                                <!-- Manual Sync Button (only show if sync is enabled) -->
                                <div v-if="syncEnabled" class="flex items-center justify-between">
                                    <div>
                                        <h3 class="text-sm font-medium text-gray-800">Manual Sync</h3>
                                        <p class="text-xs text-gray-500">Force sync your data now</p>
                                    </div>
                                    <button @click="manualSync"
                                        :disabled="syncStatus.syncing"
                                        class="px-4 py-2 border font-medium border-gray-300 hover:bg-gray-50 disabled:bg-gray-100 disabled:cursor-not-allowed rounded-lg transition-all flex items-center gap-2">
                                        <i v-if="syncStatus.syncing" class="pi pi-spin pi-spinner text-sm"></i>
                                        <span>{{ syncStatus.syncing ? 'Syncing...' : 'Sync Now' }}</span>
                                    </button>
                                </div>

                                <!-- Logout -->
                                <div class="flex items-center justify-between">
                                    <div>
                                        <h3 class="text-sm font-medium text-gray-800">Log out of all devices</h3>
                                        <p class="text-xs text-gray-500">This will sign you out everywhere</p>
                                    </div>
                                    <button @click="logout"
                                        class="px-4 py-2 border font-medium border-gray-300 hover:bg-gray-50 rounded-lg transition-all">
                                        Log out
                                    </button>
                                </div>

                                <!-- Delete account -->
                                <div class="flex items-center justify-between">
                                    <div>
                                        <h3 class="text-sm font-medium text-gray-800">Delete your account</h3>
                                        <p class="text-xs text-gray-500">Permanently delete your account and all data
                                        </p>
                                    </div>
                                    <button @click="router.push('/auth/delete_account')"
                                        class="px-4 py-2 bg-red-600 font-medium hover:bg-red-700 text-white rounded-lg transition-all">
                                        Delete account
                                    </button>
                                </div>

                                <!-- Session ID -->
                                <div class="space-y-2">
                                    <label class="block text-sm font-medium text-gray-700">Session ID</label>
                                    <input type="text" :value="parsedUserDetails.sessionId" readonly
                                        class="w-full px-4 py-2 text-sm border border-gray-300 rounded-lg bg-gray-50 text-gray-600 font-mono" />
                                </div>

                                <!-- Sync Status -->
                                <div v-if="syncEnabled" class="space-y-2">
                                    <label class="block text-sm font-medium text-gray-700">Sync Status</label>
                                    <div class="text-sm text-gray-600">
                                        <div class="flex items-center gap-2">
                                            <div :class="syncStatus.syncing ? 'bg-yellow-500' : 
                                                         syncStatus.hasUnsyncedChanges ? 'bg-orange-500' : 'bg-green-500'" 
                                                 class="w-2 h-2 rounded-full"></div>
                                            <span>
                                                {{ syncStatus.syncing ? 'Syncing...' : 
                                                   syncStatus.hasUnsyncedChanges ? 'Unsynced changes' : 'Synced' }}
                                            </span>
                                        </div>
                                        <div v-if="syncStatus.lastSync" class="text-xs text-gray-500 mt-1">
                                            Last sync: {{ new Date(syncStatus.lastSync).toLocaleString() }}
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <!-- Billing -->
                        <div v-if="activeTab === 'billing'"
                            class="bg-white p-6 rounded-lg shadow-sm border w-full max-w-2xl">
                            <h2 class="text-xl font-medium mb-6">Billing</h2>
                            <div class="text-center py-12">
                                <i class="pi pi-credit-card text-4xl text-gray-300 mb-4"></i>
                                <p class="text-gray-600 mb-2">No billing information available</p>
                                <p class="text-sm text-gray-500">Your billing details will appear here when you upgrade
                                    to a paid plan</p>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
/* Ensure proper scrolling on mobile */
@media (max-width: 768px) {
    .h-screen {
        height: 100vh;
        height: 100dvh;
        /* Better mobile support */
    }
}
</style>