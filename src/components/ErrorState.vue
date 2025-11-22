<script setup lang="ts">
import { AlertCircle, ArrowLeft, RefreshCw } from "lucide-vue-next";
import { Button } from "./ui/button";
import type { Ref } from "vue";
import { inject } from "vue";
import { useRouter } from "vue-router";

const { isOnline } = inject("globalState") as {
    isOnline: Ref<boolean>;
};

defineProps<{
    error: string;
    backButtonText?: string;
}>();
const emit = defineEmits<{
    retry: [];
}>();

const router = useRouter();
function handleBack() {
    console.log(window.history, window.history.length > 1);
    if (window.history.length > 1) {
        router.push(
            window.history.state.forward
                ? window.history.state.forward
                : window.history.state.back,
        );
    } else {
        router.push("/");
    }
}
</script>
<template>
    <div class="flex flex-col items-center justify-center h-full gap-6 px-4">
        <div
            class="rounded-full bg-red-100 dark:bg-red-900/20 p-6 animate-pulse"
        >
            <AlertCircle class="w-16 h-16 text-red-500 dark:text-red-400" />
        </div>
        <div class="text-center max-w-md">
            <h2
                class="text-2xl font-semibold text-gray-900 dark:text-white mb-2"
            >
                Failed to Load Content
            </h2>
            <p class="text-sm text-gray-600 dark:text-gray-400 mb-6">
                {{ error }}
            </p>
            <div class="flex gap-3 items-center justify-center">
                <Button
                    @click="emit('retry')"
                    :disabled="!isOnline"
                    class="inline-flex items-center gap-2 h-[38px] bg-gray-100 text-gray-900 dark:bg-gray-100 dark:text-gray-900 hover:bg-gray-100"
                >
                    <RefreshCw class="w-4 h-4" />
                    Retry
                </Button>
                <Button
                    v-if="backButtonText"
                    @click="handleBack"
                    class="inline-flex items-center gap-2 h-[38px] bg-gray-800 dark:bg-gray-800 text-gray-200 dark:text-gray-200 hover:bg-gray-800 dark:hover:bg-gray-800"
                >
                    <ArrowLeft class="w-4 h-4" />
                    {{ backButtonText }}
                </Button>
            </div>
        </div>
    </div>
</template>
