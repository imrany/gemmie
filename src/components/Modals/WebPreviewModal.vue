<script lang="ts" setup>
import { onClickOutside } from '@vueuse/core'
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
const errorMessage = ref('')
const hasAttemptedLoad = ref(false)
const showDropdown = ref(false)
const dropdownRef = ref(null)

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

// Check if error indicates connection refusal
const isConnectionRefused = (message: string): boolean => {
    const refusedIndicators = [
        'refused to connect',
        'connection refused',
        'failed to load',
        'blocked by cors',
        'x-frame-options',
        'content security policy',
        'can\'t display',
        'won\'t display',
        'not allowed',
        'access denied'
    ]
    return refusedIndicators.some(indicator =>
        message.toLowerCase().includes(indicator.toLowerCase())
    )
}

// Handle iframe load
const handleIframeLoad = () => {
    isLoading.value = false
    loadError.value = false
    errorMessage.value = ''
    hasAttemptedLoad.value = true

    // Check iframe content immediately after load
    checkIframeContent()
}

// Handle iframe error
const handleIframeError = () => {
    isLoading.value = false
    loadError.value = true
    errorMessage.value = 'Failed to load the webpage'
    hasAttemptedLoad.value = true

    // Auto-open in new tab immediately for connection refusal errors
    autoOpenInNewTabAndClose()
}

// Check iframe content for blocking or errors
const checkIframeContent = () => {
    if (iframeRef.value && !loadError.value) {
        try {
            const iframeDoc = iframeRef.value.contentDocument || iframeRef.value.contentWindow?.document

            // Check for common blocking indicators
            if (iframeDoc) {
                const bodyText = iframeDoc.body.innerText || ''
                const title = iframeDoc.title || ''

                // Detect common error messages in the iframe content itself
                const errorIndicators = [
                    'refused to connect',
                    'access denied',
                    'blocked',
                    'x-frame-options',
                    'csp',
                    'content security policy'
                ]

                const hasError = errorIndicators.some(indicator =>
                    bodyText.toLowerCase().includes(indicator.toLowerCase()) ||
                    title.toLowerCase().includes(indicator.toLowerCase())
                )

                // Check if iframe is empty or has very minimal content (likely blocked)
                const isEmpty = iframeDoc.body.children.length === 0 ||
                    bodyText.length < 50 && iframeDoc.body.children.length < 3

                if (hasError || isEmpty) {
                    errorMessage.value = bodyText.slice(0, 100) || 'Website refused to connect'
                    loadError.value = true
                    autoOpenInNewTabAndClose()
                }
            }
        } catch (error) {
            // Cross-origin restrictions - can't access iframe content
            // This is NORMAL for external sites that load successfully
            // Don't treat CORS errors as loading failures
            console.log('Cross-origin iframe, cannot check content - but it may be loading fine', error)
            
            // Only show error if we can see there's actually a problem
            // For now, assume it's loading fine if we get here without other errors
        }
    }
}

// Auto open in new tab and close modal
const autoOpenInNewTabAndClose = () => {
    if (props.data.previewUrl && hasAttemptedLoad.value) {
        console.log('Auto-opening URL in new tab due to connection refusal:', props.data.previewUrl)
        window.open(props.data.previewUrl, '_blank', 'noopener,noreferrer')
        props.closeWebPreview() // Close modal immediately
    }
}

// Open in new tab manually
const openInNewTab = () => {
    if (props.data.previewUrl) {
        window.open(props.data.previewUrl, '_blank', 'noopener,noreferrer')
    }
    showDropdown.value = false
}

// Copy URL to clipboard
const copyUrl = () => {
    if (props.data.previewUrl) {
        navigator.clipboard.writeText(props.data.previewUrl)
            .then(() => {
                console.log('URL copied to clipboard')
                // You could add a toast notification here
            })
            .catch(err => {
                console.error('Failed to copy URL:', err)
            })
    }
    showDropdown.value = false
}

// Reset states when URL changes
watch(() => props.data.previewUrl, () => {
    isLoading.value = true
    loadError.value = false
    errorMessage.value = ''
    hasAttemptedLoad.value = false
    showDropdown.value = false
})

// Watch for showWebPreview to set up content checking
watch(() => props.data.showWebPreview, (newVal) => {
    if (newVal) {
        document.body.style.overflow = 'hidden'
    } else {
        document.body.style.overflow = 'auto'
        showDropdown.value = false
    }
})

// Handle escape key
const handleKeydown = (e: KeyboardEvent) => {
    if (e.key === 'Escape' && props.data.showWebPreview) {
        if (showDropdown.value) {
            showDropdown.value = false
        } else {
            props.closeWebPreview()
        }
    }
}

onClickOutside(dropdownRef, () => {
    if (showDropdown.value) {
        showDropdown.value = false
    }
})

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
                <div class="relative w-full h-full max-w-6xl max-h-[95vh] m-4 bg-white dark:bg-gray-900 rounded-xl shadow-2xl overflow-hidden flex flex-col sm:m-4">

                    <!-- Header -->
                    <div
                        class="flex items-center justify-between px-4 py-3 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800">
                        <!-- URL Display -->
                        <div class="flex items-center gap-2 flex-1 min-w-0 mr-4">
                            <i class="pi pi-globe text-gray-500 dark:text-gray-400 text-sm flex-shrink-0"></i>
                            <div class="flex-1">  <!-- min-w-0 -->
                                <div class="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">
                                    {{ previewHostname.length>30? previewHostname.slice(0,30) + "...":previewHostname }}
                                </div>
                                <div class="text-xs text-gray-500 dark:text-gray-400 truncate">
                                    {{ data.previewUrl&&data.previewUrl?.length>30? data.previewUrl.slice(0, 30) + "...":data.previewUrl }}
                                </div>
                            </div>
                        </div>

                        <!-- Action Buttons - Desktop -->
                        <div class="hidden sm:flex items-center gap-2 flex-shrink-0">
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

                        <!-- Mobile Dropdown Menu -->
                        <div ref="dropdownRef" class="sm:hidden flex justify-end relative dropdown-container">
                            <button @click="showDropdown = !showDropdown"
                                class="p-2 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-700 text-gray-600 dark:text-gray-300 transition-colors"
                                title="More options">
                                <i class="pi pi-ellipsis-v text-lg"></i>
                            </button>

                            <!-- Dropdown Menu -->
                            <Transition name="dropdown-fade">
                                <div v-if="showDropdown"
                                    class="absolute right-0 top-full mt-1 w-48 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-50 py-1">
                                    <!-- Copy URL -->
                                    <button @click="copyUrl"
                                        class="w-full flex items-center gap-3 px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">
                                        <i class="pi pi-copy text-sm"></i>
                                        <span>Copy URL</span>
                                    </button>

                                    <!-- Open in New Tab -->
                                    <button @click="openInNewTab"
                                        class="w-full flex items-center gap-3 px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">
                                        <i class="pi pi-external-link text-sm"></i>
                                        <span>Open in New Tab</span>
                                    </button>

                                    <!-- Close Preview -->
                                    <button @click="closeWebPreview"
                                        class="w-full flex items-center gap-3 px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors border-t border-gray-200 dark:border-gray-700">
                                        <i class="pi pi-times text-sm"></i>
                                        <span>Close Preview</span>
                                    </button>
                                </div>
                            </Transition>
                        </div>
                    </div>

                    <!-- Content Area -->
                    <div class="flex-1 relative overflow-hidden bg-white dark:bg-gray-900">
                        <!-- Loading Indicator -->
                        <div v-if="isLoading && !loadError"
                            class="absolute inset-0 flex flex-col items-center justify-center bg-gray-50 dark:bg-gray-800 z-10">
                            <i class="pi pi-spin pi-spinner text-4xl text-blue-500 dark:text-blue-400 mb-4"></i>
                            <p class="text-sm text-gray-600 dark:text-gray-400">Loading preview...</p>
                        </div>

                        <!-- Error State -->
                        <div v-if="loadError"
                            class="absolute inset-0 flex flex-col items-center justify-center bg-gray-50 dark:bg-gray-800 z-10 p-8">
                            <i
                                class="pi pi-exclamation-triangle text-6xl text-orange-500 dark:text-orange-400 mb-4"></i>
                            <h3 class="text-xl font-semibold text-gray-900 dark:text-gray-100 mb-2">
                                Unable to Preview
                            </h3>
                            <p class="text-sm text-gray-600 dark:text-gray-400 text-center mb-4 max-w-md">
                                {{ errorMessage || 
                                'This website cannot be displayed in a preview due to security restrictions.' 
                                }}
                            </p>
                            <p class="text-xs text-gray-500 dark:text-gray-500 text-center mb-6 max-w-md">
                                The website has been opened in a new tab for you.
                            </p>
                            <div class="flex gap-3 flex-col sm:flex-row">
                                <button @click="openInNewTab"
                                    class="px-6 py-3 bg-blue-500 hover:bg-blue-600 dark:bg-blue-600 dark:hover:bg-blue-700 text-white rounded-lg transition-colors font-medium">
                                    <i class="pi pi-external-link mr-2"></i>
                                    Open Again
                                </button>
                                <button @click="closeWebPreview"
                                    class="px-6 py-3 bg-gray-500 hover:bg-gray-600 dark:bg-gray-600 dark:hover:bg-gray-700 text-white rounded-lg transition-colors font-medium">
                                    Close
                                </button>
                            </div>
                        </div>

                        <!-- Iframe -->
                        <iframe v-if="data.previewUrl && !loadError" ref="iframeRef" :src="data.previewUrl"
                            @load="handleIframeLoad" @error="handleIframeError" class="w-full h-full border-0 no-scrollbar overflow-hidden"
                            sandbox="allow-scripts allow-same-origin allow-forms allow-popups allow-popups-to-escape-sandbox"
                            loading="lazy">
                        </iframe>
                    </div>

                    <!-- Footer -->
                    <div class="px-4 py-2 border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800">
                        <div class="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
                            <div class="flex items-center gap-4">
                                <span class="flex items-center gap-1">
                                    <i class="pi pi-shield-check text-green-500"></i>
                                    <span class="hidden xs:inline">Sandboxed Preview</span>
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

/* Dropdown animations */
.dropdown-fade-enter-active,
.dropdown-fade-leave-active {
    transition: all 0.2s ease;
}

.dropdown-fade-enter-from {
    opacity: 0;
    transform: translateY(-8px);
}

.dropdown-fade-leave-to {
    opacity: 0;
    transform: translateY(-8px);
}

/* Make modal fullscreen on mobile */
@media (max-width: 640px) {
    .relative {
        margin: 0 !important;
        border-radius: 0 !important;
        max-width: 100% !important;
        max-height: 100% !important;
        width: 100% !important;
        height: 100% !important;
    }
}

/* Extra small screen adjustments */
@media (max-width: 475px) {
    .px-4 {
        padding-left: 1rem;
        padding-right: 1rem;
    }
    
    .text-sm {
        font-size: 0.875rem;
    }
    
    .text-xs {
        font-size: 0.75rem;
    }
}
</style>