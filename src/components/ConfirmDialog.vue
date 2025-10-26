<script lang="ts" setup>
import type { ConfirmDialogOptions } from "@/types";
import DialogBox from "./Dialog/DialogBox.vue";
let props = defineProps<{
    confirmDialog: ConfirmDialogOptions;
}>();

const close = () => {
    // eslint-disable-next-line vue/no-mutating-props
    props.confirmDialog.visible = false;
};
</script>

<template>
    <DialogBox
        :close-modal="close"
        :show="props.confirmDialog.visible || false"
        name="confirmDialog"
    >
        <div class="flex items-center gap-3 mb-4">
            <i
                :class="
                    props.confirmDialog.type === 'danger'
                        ? 'pi pi-exclamation-triangle text-red-500 dark:text-red-400'
                        : props.confirmDialog.type === 'warning'
                          ? 'pi pi-exclamation-circle text-orange-500 dark:text-orange-400'
                          : 'pi pi-info-circle text-blue-500 dark:text-blue-400'
                "
                class="text-2xl"
            ></i>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ props.confirmDialog.title }}
            </h3>
        </div>

        <p
            class="text-gray-700 dark:text-gray-300 mb-6 leading-relaxed whitespace-pre-line"
        >
            {{ props.confirmDialog.message }}
        </p>

        <div class="flex gap-3 justify-end">
            <button
                @click="props.confirmDialog.onCancel"
                class="px-4 py-2 text-gray-600 dark:text-gray-400 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors duration-200 focus:outline-none focus:ring-2 focus:ring-gray-400 dark:focus:ring-gray-500 focus:ring-opacity-50"
            >
                {{ props.confirmDialog.cancelText }}
            </button>
            <button
                @click="
                    () => {
                        props.confirmDialog.onConfirm();
                    }
                "
                :class="
                    props.confirmDialog.type === 'danger'
                        ? 'bg-red-600 hover:bg-red-700 dark:bg-red-500 dark:hover:bg-red-600 focus:ring-red-500'
                        : props.confirmDialog.type === 'warning'
                          ? 'bg-orange-600 hover:bg-orange-700 dark:bg-orange-500 dark:hover:bg-orange-600 focus:ring-orange-500'
                          : 'bg-blue-600 hover:bg-blue-700 dark:bg-blue-500 dark:hover:bg-blue-600 focus:ring-blue-500'
                "
                class="px-4 py-2 text-white rounded-lg transition-colors duration-200 focus:outline-none focus:ring-2 focus:ring-opacity-50"
            >
                {{ props.confirmDialog.confirmText }}
            </button>
        </div>
    </DialogBox>
</template>
