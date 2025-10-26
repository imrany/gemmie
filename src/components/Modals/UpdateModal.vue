<script setup lang="ts">
import { updateState, hideUpdateModal, installUpdate } from "@/main";
import DialogBox from "../Dialog/DialogBox.vue";
import { Download, DownloadCloud, Info, X } from "lucide-vue-next";

const handleUpdate = () => {
    installUpdate();
};

const handleDismiss = () => {
    hideUpdateModal();
};
</script>

<template>
    <DialogBox
        name="updateModal"
        :show="updateState.showModal"
        :close-modal="handleDismiss"
    >
        <!-- Header -->
        <div class="flex flex-col items-center gap-4 mb-6 text-center">
            <div
                class="w-16 h-16 bg-gradient-to-br from-blue-500 to-blue-600 rounded-2xl flex items-center justify-center text-white shadow-lg shadow-blue-500/40 animate-pulse"
            >
                <DownloadCloud :size="30" />
            </div>
            <h2
                class="text-2xl font-bold text-gray-900 dark:text-gray-100 tracking-tight"
            >
                Update Available
            </h2>
        </div>

        <!-- Body -->
        <div class="mb-8 space-y-4">
            <p
                class="text-base font-medium text-gray-700 dark:text-gray-300 text-center leading-relaxed"
            >
                A new version of the application is available.
            </p>
            <p
                class="text-sm text-gray-600 dark:text-gray-400 text-center leading-relaxed"
            >
                Update now to get the latest features and improvements.
            </p>
            <div
                class="flex items-center justify-center gap-2 p-3 bg-blue-50 dark:bg-blue-950/30 border border-blue-200 dark:border-blue-800 rounded-lg"
            >
                <Info
                    class="w-4 h-4 text-blue-600 dark:text-blue-400 text-base flex-shrink-0"
                />
                <p
                    class="text-sm text-blue-900 dark:text-blue-300 leading-snug"
                >
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
                <X class="w-4 h-4" />
                <span>Later</span>
            </button>
            <button
                @click="handleUpdate"
                type="button"
                class="flex-1 px-6 py-3 bg-gradient-to-r from-blue-500 to-blue-600 hover:from-blue-600 hover:to-blue-700 active:scale-98 text-white font-semibold rounded-xl transition-all duration-200 flex items-center justify-center gap-2 shadow-lg shadow-blue-500/40 hover:shadow-xl hover:shadow-blue-500/50 hover:-translate-y-0.5 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2"
            >
                <Download class="w-4 h-4" />
                <span>Update Now</span>
            </button>
        </div>
    </DialogBox>
</template>

<style scoped>
.active\:scale-98:active {
    transform: scale(0.98);
}
</style>
