<script setup lang="ts">
import { ref } from 'vue'
import { toast } from 'vue-sonner'

interface Props {
  data: {
    id: string
    language: string
    code: string
    highlighted: string
  }
}

const props = defineProps<Props>()
const copied = ref(false)

const copyCode = async () => {
  try {
    await navigator.clipboard.writeText(props.data.code)
    copied.value = true
    toast.success('Code copied to clipboard')
    setTimeout(() => {
      copied.value = false
    }, 2000)
  } catch (error) {
    toast.error('Failed to copy code')
  }
}
</script>

<template>
  <div class="code-container relative my-4">
    <pre class="bg-gray-900 rounded-lg overflow-x-auto">
      <code 
        :class="`hljs language-${data.language} text-sm`" 
        v-html="data.highlighted"
      ></code>
    </pre>
    <button
      @click="copyCode"
      class="absolute top-2 right-2 bg-gray-700 text-white px-3 py-1 rounded text-xs hover:bg-gray-600 transition-colors"
    >
      {{ copied ? 'Copied!' : 'Copy' }}
    </button>
  </div>
</template>
