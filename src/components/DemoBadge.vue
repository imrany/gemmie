<script setup lang="ts">
import { inject } from "vue";

const { isDemoMode } = inject("globalState") as {
    isDemoMode: boolean;
};
</script>

<template>
    <!-- Add Demo Badge when in demo mode -->
    <Transition name="fade">
        <div
            v-if="isDemoMode"
            class="fixed top-4 left-1/2 z-50 -translate-x-1/2 px-4 py-2 bg-yellow-500 text-white rounded-full shadow-lg flex items-center gap-2 text-sm font-medium"
        >
            <i class="pi pi-bolt"></i>
            <span>Demo Mode</span>
        </div>
    </Transition>
</template>
