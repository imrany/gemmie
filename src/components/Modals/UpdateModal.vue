<script setup lang="ts">
import { updateState, hideUpdateModal, installUpdate } from "@/main"

const handleUpdate = () => {
  installUpdate()
}

const handleDismiss = () => {
  hideUpdateModal()
}
</script>

<template>
  <Transition
    enter-active-class="transition-opacity duration-300 ease-out"
    enter-from-class="opacity-0"
    enter-to-class="opacity-100"
    leave-active-class="transition-opacity duration-300 ease-in"
    leave-from-class="opacity-100"
    leave-to-class="opacity-0"
  >
    <div 
      v-if="updateState.showModal" 
      class="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-[9999] p-4"
      @click.self="handleDismiss"
    >
      <Transition
        enter-active-class="transition-all duration-300 ease-out"
        enter-from-class="scale-95 opacity-0 -translate-y-4"
        enter-to-class="scale-100 opacity-100 translate-y-0"
        leave-active-class="transition-all duration-300 ease-in"
        leave-from-class="scale-100 opacity-100 translate-y-0"
        leave-to-class="scale-95 opacity-0 -translate-y-4"
      >
        <div 
          v-if="updateState.showModal"
          class="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-md p-8 relative"
        >
          <!-- Header -->
          <div class="flex flex-col items-center gap-4 mb-6 text-center">
            <div class="w-16 h-16 bg-gradient-to-br from-blue-500 to-blue-600 rounded-2xl flex items-center justify-center text-white shadow-lg shadow-blue-500/40 animate-pulse">
              <i class="pi pi-sync text-3xl"></i>
            </div>
            <h2 class="text-2xl font-bold text-gray-900 dark:text-gray-100 tracking-tight">
              Update Available
            </h2>
          </div>
          
          <!-- Body -->
          <div class="mb-8 space-y-4">
            <p class="text-base font-medium text-gray-700 dark:text-gray-300 text-center leading-relaxed">
              A new version of the application is available.
            </p>
            <p class="text-sm text-gray-600 dark:text-gray-400 text-center leading-relaxed">
              Update now to get the latest features and improvements.
            </p>
            <div class="flex items-center justify-center gap-2 p-3 bg-blue-50 dark:bg-blue-950/30 border border-blue-200 dark:border-blue-800 rounded-lg">
              <i class="pi pi-info-circle text-blue-600 dark:text-blue-400 text-base flex-shrink-0"></i>
              <p class="text-sm text-blue-900 dark:text-blue-300 leading-snug">
                The page will reload automatically after updating.
              </p>
            </div>
          </div>
          
          <!-- Footer -->
          <div class="flex gap-3 text-sm">
            <button
              @click="handleDismiss"
              type="button"
              class="flex-1 px-6 py-3 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 active:scale-98 text-gray-700 dark:text-gray-300 font-semibold rounded-xl transition-all duration-200 flex items-center justify-center gap-2 border border-gray-200 dark:border-gray-600 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2"
            >
              <i class="pi pi-times"></i>
              <span>Later</span>
            </button>
            <button
              @click="handleUpdate"
              type="button"
              class="flex-1 px-6 py-3 bg-gradient-to-r from-blue-500 to-blue-600 hover:from-blue-600 hover:to-blue-700 active:scale-98 text-white font-semibold rounded-xl transition-all duration-200 flex items-center justify-center gap-2 shadow-lg shadow-blue-500/40 hover:shadow-xl hover:shadow-blue-500/50 hover:-translate-y-0.5 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2"
            >
              <i class="pi pi-download"></i>
              <span>Update Now</span>
            </button>
          </div>
        </div>
      </Transition>
    </div>
  </Transition>
</template>

<style scoped>
@keyframes pulse {
  0%, 100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.05);
  }
}

.animate-pulse {
  animation: pulse 2s ease-in-out infinite;
}

.active\:scale-98:active {
  transform: scale(0.98);
}

/* Responsive adjustments */
@media (max-width: 640px) {
  .flex.gap-3 {
    flex-direction: column-reverse;
  }
}

/* Reduced motion support */
@media (prefers-reduced-motion: reduce) {
  .animate-pulse {
    animation: none;
  }
  
  .transition-all,
  .transition-opacity {
    transition: none !important;
  }
}
</style>