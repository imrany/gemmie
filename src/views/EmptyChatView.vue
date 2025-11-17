<script setup lang="ts">
import { Skeleton } from "@/components/ui/skeleton";
import type { UserDetails } from "@/types";
import { Database, RefreshCw, Shield } from "lucide-vue-next";
import type { FunctionalComponent } from "vue";
import type { Ref } from "vue";
import { inject } from "vue";

const { parsedUserDetails, isDarkMode, isLoading } = inject("globalState") as {
    parsedUserDetails: UserDetails;
    isDarkMode: Ref<boolean>;
    isLoading: Ref<boolean>;
};

const { suggestionPrompts, selectSuggestion } = defineProps<{
    suggestionPrompts: {
        icon: FunctionalComponent;
        title: string;
        prompt: string;
    }[];
    selectSuggestion: (prompt: string) => void;
}>();
</script>

<template>
    <div>
        <div
            v-if="isLoading"
            class="md:max-w-3xl min-h-[calc(100vh-200px)] w-[100vw] flex-grow px-2 space-y-3 sm:space-y-4"
        >
            <div
                class="flex animate-pulse items-start gap-2 font-medium bg-gray-100 dark:bg-gray-800 text-black dark:text-gray-100 px-4 rounded-2xl prose prose-sm dark:prose-invert chat-bubble w-fit max-w-full"
            >
                <!-- Avatar container -->
                <div class="flex-shrink-0 py-3">
                    <div
                        class="flex items-center justify-center w-7 h-7 text-gray-100 dark:text-gray-800 bg-gray-700 dark:bg-gray-200 rounded-full text-sm font-semibold"
                    >
                        {{
                            parsedUserDetails.username.toUpperCase().slice(0, 2)
                        }}
                    </div>
                </div>
                <div class="flex-1 md:w-[700px] w-[95vw]"></div>
            </div>

            <div
                class="flex flex-col gap-2 w-full md:max-w-3xl max-w-full relative"
            >
                <Skeleton
                    v-for="index in [1, 2, 3, 4, 5, 6].reverse()"
                    :key="index"
                    :class="[
                        'chat-message bg-gray-100 dark:bg-gray-800',
                        `max-w-full w-full`,
                        index === 3 ? 'mt-4' : '',
                        `h-${index + 2}`,
                    ]"
                />
            </div>
        </div>

        <div
            class="flex h-screen justify-center md:max-w-3xl max-w-[100vw] max-md:px-4 flex-col md:flex-grow items-center gap-3 text-gray-600 dark:text-gray-400"
            v-else
        >
            <img
                :src="
                    parsedUserDetails?.theme === 'dark' ||
                    (parsedUserDetails?.theme === 'system' && isDarkMode)
                        ? '/logo-light.svg'
                        : '/logo.svg'
                "
                alt="Gemmie Logo"
                class="w-[60px] h-[60px] rounded-md"
            />

            <p class="text-gray-700 dark:text-gray-300 text-3xl font-semibold">
                Hey{{
                    parsedUserDetails?.username &&
                    parsedUserDetails?.username === "demo"
                        ? ", tester"
                        : ", " + parsedUserDetails?.username || ", there"
                }}
            </p>
            <div class="text-center max-w-md space-y-2">
                <p class="text-gray-600 dark:text-gray-400 leading-relaxed">
                    Experience privacy-first conversations with advanced AI.
                    Your data stays secure, local and synced to your all
                    devices.
                </p>
                <div
                    class="flex items-center justify-center gap-6 text-sm text-gray-500 dark:text-gray-400 mt-4"
                >
                    <div class="flex items-center gap-1">
                        <Shield
                            class="w-4 h-4 text-green-500 dark:text-green-400"
                        />
                        <span>Private</span>
                    </div>
                    <div class="flex items-center gap-1">
                        <Database
                            class="w-4 h-4 text-blue-500 dark:text-blue-400"
                        />
                        <span>Local Stored</span>
                    </div>
                    <div class="flex items-center gap-1">
                        <RefreshCw
                            class="w-4 h-4 text-purple-500 dark:text-purple-400"
                        />
                        <span>Synced</span>
                    </div>
                </div>
            </div>

            <div
                class="flex flex-col gap-4 w-full mb-[100px] max-w-2xl relative"
            >
                <!-- Suggestion Chips Grid -->
                <div class="w-full flex justify-center">
                    <div class="flex flex-wrap justify-center gap-2">
                        <button
                            v-for="(suggestion, index) in suggestionPrompts"
                            :key="index"
                            type="button"
                            @click="selectSuggestion(suggestion.prompt)"
                            class="group flex w-[100px] items-center gap-2 justify-center h-9 bg-white dark:bg-gray-800 border-[1px] border-gray-200 dark:border-gray-700 rounded-lg hover:border-blue-500 dark:hover:border-blue-400 hover:bg-gray-50 dark:hover:bg-gray-700 transition-all duration-200 transform hover:scale-105 shadow-sm hover:shadow-md"
                        >
                            <component
                                :is="suggestion.icon"
                                class="w-4 h-4 text-gray-500 dark:text-gray-300 text-sm group-hover:scale-110 transition-transform"
                            />
                            <span
                                class="text-xs font-medium text-gray-700 dark:text-gray-300"
                            >
                                {{ suggestion.title }}
                            </span>
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>
