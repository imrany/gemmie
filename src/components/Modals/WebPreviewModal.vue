<script lang="ts" setup>
import { ref, watch, onMounted, onBeforeUnmount, computed } from 'vue'

interface Props {
  data: {
    showWebPreview: boolean
    previewUrl: string | null
  }
  closeWebPreview: () => void
}

const props = defineProps<Props>()
const iframeRef = ref<HTMLIFrameElement | null>(null)
const isLoading = ref(true)
const loadError = ref(false)

// Computed property to safely extract hostname
const previewHostname = computed(() => {
  if (!props.data.previewUrl) return ''
  try {
    const url = new URL(props.data.previewUrl)
    return url.hostname
  } catch (error) {
    console.error('Invalid URL:', error)
    return props.data.previewUrl
  }
})

// Handle iframe load
const handleIframeLoad = () => {
  isLoading.value = false
  loadError.value = false
}

// Handle iframe error
const handleIframeError = () => {
  isLoading.value = false
  loadError.value = true
}

// Open in new tab
const openInNewTab = () => {
  if (props.data.previewUrl) {
    window.open(props.data.previewUrl, '_blank', 'noopener,noreferrer')
  }
}

// Copy URL to clipboard
const copyUrl = () => {
  if (props.data.previewUrl) {
    navigator.clipboard.writeText(props.data.previewUrl)
      .then(() => {
        // You can add a toast notification here if you have access to it
        console.log('URL copied to clipboard')
      })
      .catch(err => {
        console.error('Failed to copy URL:', err)
      })
  }
}

// Reset states when URL changes
watch(() => props.data.previewUrl, () => {
  isLoading.value = true
  loadError.value = false
})

// Prevent body scroll when modal is open
watch(() => props.data.showWebPreview, (newVal) => {
  if (newVal) {
    document.body.style.overflow = 'hidden'
  } else {
    document.body.style.overflow = 'auto'
  }
})

// Handle escape key
const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Escape' && props.data.showWebPreview) {
    props.closeWebPreview()
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
})

onBeforeUnmount(() => {
  document.removeEventListener('keydown', handleKeydown)
  document.body.style.overflow = 'auto'
})
</script>

<template>
  <Teleport to="body">
    <Transition name="modal-fade">
      <div v-if="data.showWebPreview" 
           class="fixed inset-0 z-[9999] flex items-center justify-center bg-black/70 backdrop-blur-sm"
           @click.self="closeWebPreview">
        
        <!-- Modal Container -->
        <div class="relative w-full h-full max-w-6xl max-h-[95vh] m-4 bg-white dark:bg-gray-900 rounded-xl shadow-2xl overflow-hidden flex flex-col"
             @click.stop>
          
          <!-- Header -->
          <div class="flex items-center justify-between px-4 py-3 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800">
            <!-- URL Display -->
            <div class="flex items-center gap-2 flex-1 min-w-0 mr-4">
              <i class="pi pi-globe text-gray-500 dark:text-gray-400 text-sm flex-shrink-0"></i>
              <div class="flex-1 min-w-0">
                <div class="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">
                  {{ previewHostname }}
                </div>
                <div class="text-xs text-gray-500 dark:text-gray-400 truncate">
                  {{ data.previewUrl }}
                </div>
              </div>
            </div>

            <!-- Action Buttons -->
            <div class="flex items-center gap-2 flex-shrink-0">
              <!-- Copy URL -->
              <button @click="copyUrl"
                      class="p-2 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-700 text-gray-600 dark:text-gray-300 transition-colors"
                      title="Copy URL">
                <i class="pi pi-copy text-sm"></i>
              </button>

              <!-- Open in New Tab -->
              <button @click="openInNewTab"
                      class="p-2 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-700 text-gray-600 dark:text-gray-300 transition-colors"
                      title="Open in new tab">
                <i class="pi pi-external-link text-sm"></i>
              </button>

              <!-- Close Button -->
              <button @click="closeWebPreview"
                      class="p-2 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-700 text-gray-600 dark:text-gray-300 transition-colors"
                      title="Close (Esc)">
                <i class="pi pi-times text-lg"></i>
              </button>
            </div>
          </div>

          <!-- Content Area -->
          <div class="flex-1 relative overflow-hidden bg-white dark:bg-gray-900">
            <!-- Loading Indicator -->
            <div v-if="isLoading" 
                 class="absolute inset-0 flex flex-col items-center justify-center bg-gray-50 dark:bg-gray-800 z-10">
              <i class="pi pi-spin pi-spinner text-4xl text-blue-500 dark:text-blue-400 mb-4"></i>
              <p class="text-sm text-gray-600 dark:text-gray-400">Loading preview...</p>
            </div>

            <!-- Error State -->
            <div v-if="loadError" 
                 class="absolute inset-0 flex flex-col items-center justify-center bg-gray-50 dark:bg-gray-800 z-10 p-8">
              <i class="pi pi-exclamation-triangle text-6xl text-orange-500 dark:text-orange-400 mb-4"></i>
              <h3 class="text-xl font-semibold text-gray-900 dark:text-gray-100 mb-2">
                Unable to Preview
              </h3>
              <p class="text-sm text-gray-600 dark:text-gray-400 text-center mb-6 max-w-md">
                This website cannot be displayed in a preview. This may be due to security restrictions set by the website.
              </p>
              <button @click="openInNewTab"
                      class="px-6 py-3 bg-blue-500 hover:bg-blue-600 dark:bg-blue-600 dark:hover:bg-blue-700 text-white rounded-lg transition-colors font-medium">
                <i class="pi pi-external-link mr-2"></i>
                Open in New Tab
              </button>
            </div>

            <!-- Iframe -->
            <iframe v-if="data.previewUrl"
                    ref="iframeRef"
                    :src="data.previewUrl"
                    @load="handleIframeLoad"
                    @error="handleIframeError"
                    class="w-full h-full border-0"
                    sandbox="allow-scripts allow-same-origin allow-forms allow-popups allow-popups-to-escape-sandbox"
                    loading="lazy">
            </iframe>
          </div>

          <!-- Footer (Optional - for additional info) -->
          <div class="px-4 py-2 border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800">
            <div class="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
              <div class="flex items-center gap-4">
                <span class="flex items-center gap-1">
                  <i class="pi pi-shield-check text-green-500"></i>
                  Sandboxed Preview
                </span>
                <span class="hidden sm:inline">Press Esc to close</span>
              </div>
              <span class="hidden sm:inline">Click outside to close</span>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity 0.3s ease;
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}

.modal-fade-enter-active .relative,
.modal-fade-leave-active .relative {
  transition: transform 0.3s ease;
}

.modal-fade-enter-from .relative,
.modal-fade-leave-to .relative {
  transform: scale(0.95);
}
</style>