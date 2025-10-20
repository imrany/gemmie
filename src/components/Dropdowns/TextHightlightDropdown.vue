<script setup lang="ts">
import { ref, nextTick, onMounted, onUnmounted, inject, type Ref } from 'vue' // Add lifecycle hooks
import { toast } from 'vue-sonner'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'

const {
  isOpenTextHighlightPopover,
} = inject("globalState") as {
  isOpenTextHighlightPopover: Ref<boolean>,
}
const selectedText = ref('')
const triggerElement = ref<HTMLElement | null>(null)
const triggerStyle = ref({ top: '0px', left: '0px' })
const actualSide = ref('top')

async function handleMouseUp() {
  const selection = window.getSelection()
  const text = selection?.toString().trim()

  if (text && text.length > 0) {
    selectedText.value = text

    const range = selection?.getRangeAt(0)
    const rect = range?.getBoundingClientRect()

    if (rect && triggerElement.value) {
      // Position trigger element at the center of selection, above the text
      triggerStyle.value = {
        top: (rect.top + window.scrollY - 10) + 'px',
        left: (rect.left + window.scrollX + rect.width / 2) + 'px'
      }

      isOpenTextHighlightPopover.value = true
    }
  }
}

function copyText() {
  navigator.clipboard.writeText(selectedText.value)
  toast.success('Copied to clipboard!')
  isOpenTextHighlightPopover.value = false
}

function speakText() {
  const utterance = new SpeechSynthesisUtterance(selectedText.value)
  utterance.rate = 1
  window.speechSynthesis.speak(utterance)
  toast.success('Playing...')
  isOpenTextHighlightPopover.value = false
}

function translateText() {
  window.open(`https://translate.google.com/?text=${encodeURIComponent(selectedText.value)}`, '_blank')
  isOpenTextHighlightPopover.value = false
}

onMounted(() => {
  document.addEventListener('mouseup', handleMouseUp)
})

onUnmounted(() => {
  document.removeEventListener('mouseup', handleMouseUp)
})
</script>

<template>
  <div>
    <Popover v-model:open="isOpenTextHighlightPopover">
      <PopoverTrigger as-child>
        <div ref="triggerElement" class="fixed w-0 h-0 pointer-events-none" :style="{
          top: triggerStyle.top,
          left: triggerStyle.left,
          transform: 'translateX(-50%)',
          zIndex: 50
        }" />
      </PopoverTrigger>
      <PopoverContent side="top" :avoid-collisions="true" :collision-padding="10" align="center"
        class="w-fit h-fit rounded-md shadow-md transition-all duration-200 bg-white dark:bg-slate-900 border border-gray-300 dark:border-slate-600 p-0 relative"
        :side-offset="6" @escape-key-down="isOpenTextHighlightPopover = false">

        <!-- Pointer -->
        <div
          class="absolute w-2 h-2 bg-white dark:bg-slate-900 border-b border-r border-gray-300 dark:border-slate-600 rotate-45"
          :class="{
            '-bottom-1 left-1/2 transform -translate-x-1/2': actualSide === 'top',
            '-top-1 left-1/2 transform -translate-x-1/2': actualSide === 'bottom'
          }"></div>

        <div class="flex items-center relative z-10 rounded-md">
          <button @click="copyText"
            class="px-3 py-2 text-sm text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-slate-800 transition-colors duration-200 rounded-l-md">
            Copy
          </button>

          <button @click="speakText"
            class="px-3 py-2 text-sm text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-slate-800 transition-colors duration-200 border-l border-gray-200 dark:border-slate-700">
            Listen
          </button>

          <button @click="translateText"
            class="px-3 py-2 text-sm text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-slate-800 transition-colors duration-200 rounded-r-md border-l border-gray-200 dark:border-slate-700">
            Translate
          </button>
        </div>
      </PopoverContent>
    </Popover>
  </div>
</template>