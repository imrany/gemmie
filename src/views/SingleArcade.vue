<script setup lang="ts">
import type { ApiResponse, RawArcade } from "@/types";
import { ref } from "vue";
import { inject } from "vue";
import { onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { toast } from "vue-sonner";
import { Loader2, AlertCircle, ArrowLeft } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import LoadingState from "@/components/LoadingState.vue";
import ErrorState from "@/components/ErrorState.vue";

const route = useRoute();
const router = useRouter();
const arcadeId = route.params.id as string;

const { unsecureApiCall } = inject("globalState") as {
    unsecureApiCall: <T>(
        endpoint: string,
        options: RequestInit,
    ) => Promise<ApiResponse<T>>;
};

const arcade = ref<RawArcade>();
const isLoading = ref(true);
const error = ref<string | null>(null);
const iframeLoaded = ref(false);

async function fetchArcade() {
    isLoading.value = true;
    error.value = null;
    iframeLoaded.value = false;

    try {
        const res = await unsecureApiCall<RawArcade>(
            `/arcades/${arcadeId.trim()}`,
            {
                method: "GET",
            },
        );

        if (!res.success) {
            throw new Error(res.message || "Failed to load arcade content");
        }

        if (res.data) {
            arcade.value = res.data;
            console.log(arcade);
        } else {
            throw new Error("No arcade data found");
        }
    } catch (err: any) {
        console.error("Fetch arcade error:", err);
        if (err.message?.includes("Failed to fetch")) {
            error.value =
                "Failed to connect to the server. Please check your internet connection.";
        } else if (err.message?.includes("404")) {
            error.value =
                "Arcade not found. It may have been removed or the ID is incorrect.";
        } else {
            error.value = err.message || "Failed to load arcade content";
        }
    } finally {
        isLoading.value = false;
    }
}

function handleIframeLoad() {
    iframeLoaded.value = true;
}

function handleIframeError() {
    error.value = "Failed to load content in preview";
    toast.error("Preview failed to load");
}

function goBack() {
    router.push("/arcade");
}

onMounted(async () => {
    await fetchArcade();
});
</script>

<template>
    <div class="w-screen h-screen overflow-hidden bg-gray-50 dark:bg-gray-900">
        <!-- Loading State -->
        <LoadingState
            description="Please wait while we fetch your content"
            label="Loading Arcade Content"
            v-if="isLoading"
        />

        <!-- Error State -->
        <ErrorState
            v-if="error"
            @retry="fetchArcade"
            back-button-text="Back to Arcade"
            :error="error"
        />

        <!-- Content State -->
        <div v-else-if="arcade?.code" class="relative w-full h-full">
            <!-- Loading overlay for iframe -->
            <Transition name="fade">
                <div
                    v-if="!iframeLoaded"
                    class="absolute inset-0 flex items-center justify-center bg-white dark:bg-gray-900 z-10"
                >
                    <div class="text-center">
                        <Loader2
                            class="w-10 h-10 text-blue-500 animate-spin mx-auto mb-3"
                        />
                        <p class="text-sm text-gray-500 dark:text-gray-400">
                            Rendering preview...
                        </p>
                    </div>
                </div>
            </Transition>

            <!-- Iframe -->
            <iframe
                :srcdoc="arcade.code"
                @load="handleIframeLoad"
                @error="handleIframeError"
                class="w-full h-full border-0 bg-white dark:bg-gray-900"
                sandbox="allow-scripts allow-popups allow-popups-to-escape-sandbox"
                title="Arcade Content Preview"
                referrerpolicy="no-referrer"
            />
        </div>

        <!-- No Content State -->
        <div
            v-else
            class="flex flex-col items-center justify-center h-full gap-4 px-4"
        >
            <AlertCircle class="w-16 h-16 text-gray-400 dark:text-gray-600" />
            <div class="text-center">
                <h2
                    class="text-xl font-semibold text-gray-900 dark:text-white mb-2"
                >
                    No Content Available
                </h2>
                <p class="text-sm text-gray-500 dark:text-gray-400 mb-4">
                    This arcade entry doesn't have any content to display.
                </p>
                <Button
                    @click="goBack"
                    variant="outline"
                    class="inline-flex items-center gap-2"
                >
                    <ArrowLeft class="w-4 h-4" />
                    Back to Arcade
                </Button>
            </div>
        </div>
    </div>
</template>

<style scoped>
/* Fade transition */
.fade-enter-active,
.fade-leave-active {
    transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
    opacity: 0;
}
</style>
