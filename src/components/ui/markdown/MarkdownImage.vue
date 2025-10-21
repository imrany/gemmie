<script setup lang="ts">
import { ref } from 'vue'

interface Props {
  data: {
    id: string
    src: string
    alt: string
    title?: string
    className?: string
    width?: string
    height?: string
    style?: string
  }
}

const props = defineProps<Props>()
const imageError = ref(false)

const handleImageError = () => {
  imageError.value = true
}

const openImage = () => {
  window.open(props.data.src, '_blank')
}

const getImageClasses = () => {
  const baseClasses = 'rounded-lg shadow-md hover:shadow-lg transition-all duration-200 cursor-zoom-in border border-gray-200 dark:border-gray-700'
  
  if (!props.data.className) {
    return `${baseClasses} max-w-full h-auto`
  }

  const classMap: Record<string, string> = {
    circle: 'w-32 h-32 rounded-full object-cover',
    small: 'max-w-48 h-auto',
    medium: 'max-w-md h-auto',
    large: 'max-w-4xl h-auto',
    full: 'w-full h-auto',
    'float-left': 'float-left mr-4 mb-4 max-w-xs',
    'float-right': 'float-right ml-4 mb-4 max-w-xs',
    center: 'mx-auto',
    clickable: 'cursor-pointer'
  }

  return `${baseClasses} ${classMap[props.data.className] || props.data.className}`
}

const containerClasses = () => {
  if (props.data.className === 'center') {
    return 'image-container my-4 flex justify-center'
  }
  return 'image-container my-4'
}
</script>

<template>
  <div :class="containerClasses()">
    <img
      v-if="!imageError"
      :src="data.src"
      :alt="data.alt"
      :title="data.title"
      :width="data.width"
      :height="data.height"
      :class="getImageClasses()"
      @click="openImage"
      @error="handleImageError"
      loading="lazy"
    />
    <div
      v-else
      class="text-gray-600 dark:text-gray-400 p-4 border border-gray-300 dark:border-gray-600 rounded-lg bg-gray-50 dark:bg-gray-800"
    >
      <i class="pi pi-image text-2xl mb-2"></i>
      <p>Failed to load: {{ data.alt }}</p>
    </div>
    <p
      v-if="data.title && !imageError"
      class="text-sm text-gray-500 dark:text-gray-400 mt-2 text-center italic"
    >
      {{ data.title }}
    </p>
  </div>
</template>
