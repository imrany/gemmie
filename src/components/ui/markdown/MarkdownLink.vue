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
const fetchPreview = async (url:string) => {
  // Reset preview when fetching new URL
  preview.value = {
    url: url,
    domain: getDomain(url),
    title: props.data.text,
    description: '',
    previewImage: '',
    loading: true,
    error: false
  }
  
  try {
    const data = await fetchLinkPreview(url, {
      cache: false
    })
    
    preview.value = {
      ...data,
      title: data.title || props.data.text,
      domain: data.domain || getDomain(url)
    }
  } catch (error) {
    console.error('Failed to fetch link preview:', error)
    preview.value = {
      url,
      domain: getDomain(url),
      title: props.data.text,
      description: '',
      previewImage: '',
      loading: false,
      error: true
    }
  }
}

// Watch for popover open and link changes
watch([isOpen, () => props.data?.href], ([newIsOpen, newHref], [oldIsOpen, oldHref]) => {
  const hrefChanged = newHref !== oldHref
  const shouldFetch = newIsOpen && newHref
  
  if (shouldFetch) {
    // If href changed or no preview exists, fetch fresh data
    if (hrefChanged || !preview.value || preview.value.url !== newHref) {
      fetchPreview(newHref)
    }
  }
}, { immediate: true })

// Watch specifically for href changes when popover is open
watch(() => props.data?.href, (newHref, oldHref) => {
  if (newHref && newHref !== oldHref && isOpen.value) {
    fetchPreview(newHref)
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