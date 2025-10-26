<script setup lang="ts">
import { X } from "lucide-vue-next";

const { name, show, closeModal } = defineProps<{
    name: string;
    show: boolean;
    closeModal: () => void;
}>();
</script>
<template>
    <Transition :name="name">
        <div
            v-if="show"
            class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm"
            @click.self="closeModal"
        >
            <div
                class="relative scale-90 bg-white dark:bg-gray-800 rounded-xl shadow-2xl max-w-md w-full p-6 transform transition-all"
                @click.stop
            >
                <!-- Close button -->
                <button
                    @click="closeModal"
                    class="absolute top-4 right-4 w-8 h-8 flex items-center justify-center rounded-full hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
                >
                    <X class="w-4 h-4 text-gray-500 dark:text-gray-400" />
                </button>

                <div>
                    <slot />
                </div>
            </div>
        </div>
    </Transition>
</template>

<style scoped>
.touch-pan-y {
    touch-action: pan-y;
}

/* Modal transitions */
.modal-enter-active,
.modal-leave-active {
    transition: opacity 0.3s ease;
}

.modal-enter-active > div,
.modal-leave-active > div {
    transition: transform 0.3s ease;
}

.modal-enter-from,
.modal-leave-to {
    opacity: 0;
}

.modal-enter-from > div,
.modal-leave-to > div {
    transform: scale(0.9);
}

/* Fade transition for demo badge */
.fade-enter-active,
.fade-leave-active {
    transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
    opacity: 0;
}
</style>
