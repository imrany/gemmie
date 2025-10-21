<script setup lang="ts">
import { computed, ref, watch, inject } from 'vue'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import LinkPreviewComponent from '@/components/LinkPreview.vue'
import type { LinkPreview } from '@/types'

interface Props {
  data: {
    id: string
    href: string
    text: string
    title?: string
  }
}

const {
    fetchLinkPreview,
}=inject("globalState") as{
    fetchLinkPreview: (url: string, options:{
        cache?: boolean
    }) => Promise<LinkPreview>,
}

const props = defineProps<Props>()

const displayText = computed(() => {
  return props.data.text.length > 60 
    ? props.data.text.slice(0, 60) + '...' 
    : props.data.text
})

const isOpen = ref(false)
const preview = ref<LinkPreview | null>(null)

// Extract domain from URL
const getDomain = (url: string) => {
  try {
    const urlObj = new URL(url)
    return urlObj.hostname.replace('www.', '')
  } catch {
    return url
  }
}

// Fetch link preview
const fetchPreview = async () => {
  if (preview.value) return
  
  // Set loading state
  preview.value = {
    url: props.data.href,
    domain: getDomain(props.data.href),
    title: props.data.text,
    description: '',
    previewImage: '',
    loading: true,
    error: false
  }
  
  try {
    // Replace this with your actual API endpoint
    const data= await fetchLinkPreview(props.data.href, {
        cache: false
    })
    
    preview.value = {
      url: props.data.href,
      domain: getDomain(props.data.href),
      title: data.title || props.data.text,
      description: data.description || '',
      previewImage: data.previewImage || data.previewImage || '',
      video: data.video || '',
      videoThumbnail: data.videoThumbnail || '',
      videoType: data.videoType || undefined,
      videoDuration: data.videoDuration || '',
      embedHtml: data.embedHtml || '',
      loading: false,
      error: false
    }
  } catch (error) {
    preview.value = {
      url: props.data.href,
      domain: getDomain(props.data.href),
      title: props.data.text,
      description: '',
      previewImage: '',
      loading: false,
      error: true
    }
  }
}

// Fetch preview when popover opens
watch(isOpen, (newValue) => {
  if (newValue && !preview.value) {
    fetchPreview()
  }
})
</script>

<template>
  <Popover v-model:open="isOpen">
    <PopoverTrigger as-child>
      <a
        :href="data.href"
        :title="data.title"
        class="text-blue-600 underline hover:text-blue-800 link-with-preview"
        target="_blank"
        rel="noopener noreferrer"
        @mouseenter="isOpen = true"
        @mouseleave="isOpen = false"
        @click.stop
      >
        {{ displayText }}
      </a>
    </PopoverTrigger>
    <PopoverContent 
      class="w-auto p-0 border-0"
      @mouseenter="isOpen = true"
      @mouseleave="isOpen = false"
    >
      <LinkPreviewComponent v-if="preview" :preview="preview" />
    </PopoverContent>
  </Popover>
</template>