<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  data: {
    id: string
    type: 'NOTE' | 'TIP' | 'IMPORTANT' | 'WARNING' | 'CAUTION'
    content: string
  }
}

const props = defineProps<Props>()

const calloutConfig = computed(() => {
  const configs = {
    NOTE: { icon: 'ℹ️', color: 'blue', label: 'Note' },
    TIP: { icon: '💡', color: 'green', label: 'Tip' },
    IMPORTANT: { icon: '❗', color: 'purple', label: 'Important' },
    WARNING: { icon: '⚠️', color: 'yellow', label: 'Warning' },
    CAUTION: { icon: '🚨', color: 'red', label: 'Caution' }
  }
  return configs[props.data.type]
})
</script>

<template>
  <div
    :class="`callout my-4 p-4 border-l-4 border-${calloutConfig.color}-500 dark:border-${calloutConfig.color}-400 bg-${calloutConfig.color}-50 dark:bg-${calloutConfig.color}-900/20 rounded-r-lg`"
  >
    <div class="flex items-start gap-2">
      <span class="text-lg flex-shrink-0">{{ calloutConfig.icon }}</span>
      <div class="flex-1">
        <div
          :class="`font-semibold text-${calloutConfig.color}-900 dark:text-${calloutConfig.color}-100 mb-1`"
        >
          {{ calloutConfig.label }}
        </div>
        <div
          :class="`text-${calloutConfig.color}-800 dark:text-${calloutConfig.color}-200 text-sm`"
        >
          {{ data.content }}
        </div>
      </div>
    </div>
  </div>
</template>
