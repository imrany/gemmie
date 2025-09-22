<script lang="ts" setup>
import type { ConfirmDialogOptions } from '@/types';
let props = defineProps<{
  confirmDialog: ConfirmDialogOptions
}>()
</script>

<template>
    <div v-if="props.confirmDialog.visible"
      class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg p-6 max-w-md w-full mx-4 shadow-2xl">
        <div class="flex items-center gap-3 mb-4">
          <i :class="props.confirmDialog.type === 'danger' ? 'pi pi-exclamation-triangle text-red-500' :
            props.confirmDialog.type === 'warning' ? 'pi pi-exclamation-circle text-orange-500' :
              'pi pi-info-circle text-blue-500'" class="text-2xl"></i>
          <h3 class="text-lg font-semibold text-gray-900">{{ props.confirmDialog.title }}</h3>
        </div>

        <p class="text-gray-700 mb-6 leading-relaxed whitespace-pre-line">{{ props.confirmDialog.message }}</p>

        <div class="flex gap-3 justify-end">
          <button @click="props.confirmDialog.onCancel"
            class="px-4 py-2 text-gray-600 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors">
            {{ props.confirmDialog.cancelText }}
          </button>
          <button @click="() => { props.confirmDialog.onConfirm(); props.confirmDialog.visible = false }" :class="props.confirmDialog.type === 'danger' ? 'bg-red-600 hover:bg-red-700' :
            props.confirmDialog.type === 'warning' ? 'bg-orange-600 hover:bg-orange-700' :
              'bg-blue-600 hover:bg-blue-700'" class="px-4 py-2 text-white rounded-lg transition-colors">
            {{ props.confirmDialog.confirmText }}
          </button>
        </div>
      </div>
    </div>
</template>