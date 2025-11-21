<script setup lang="ts">
import { Loader2 } from "lucide-vue-next";

defineProps<{
    label: string;
    description: string;
}>();
</script>
<template>
    <!-- Loading State -->
    <div class="flex flex-col items-center justify-center h-full gap-4">
        <Loader2 class="w-12 h-12 text-blue-500 animate-spin" />
        <div class="text-center">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ label ? label : "Loading" }}
            </h2>
            <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">
                {{ description ? description + "..." : "Please wait..." }}
            </p>
        </div>
    </div>
</template>
