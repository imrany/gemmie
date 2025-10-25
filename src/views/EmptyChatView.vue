<script setup lang="ts">
import type { UserDetails } from "@/types";
import type { Ref } from "vue";
import { inject } from "vue";

const { parsedUserDetails, showInput, setShowInput, isDarkMode } = inject(
    "globalState",
) as {
    parsedUserDetails: UserDetails;
    showInput: Ref<boolean>;
    setShowInput: () => void;
    isDarkMode: Ref<boolean>;
};

const { suggestionPrompts, selectSuggestion } = defineProps<{
    suggestionPrompts: {
        icon: string;
        title: string;
        prompt: string;
    }[];
    selectSuggestion: (prompt: string) => void;
}>();
</script>

<template>
    <div
        class="flex h-screen justify-center md:max-w-3xl max-w-[100vw] max-md:px-4 flex-col md:flex-grow items-center gap-3 text-gray-600 dark:text-gray-400"
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
            {{ parsedUserDetails?.username || "Gemmie" }}
        </p>
        <div class="text-center max-w-md space-y-2">
            <p class="text-gray-600 dark:text-gray-400 leading-relaxed">
                Experience privacy-first conversations with advanced AI. Your
                data stays secure, local and synced to your all devices.
            </p>
            <div
                class="flex items-center justify-center gap-6 text-sm text-gray-500 dark:text-gray-400 mt-4"
            >
                <div class="flex items-center gap-1">
                    <i
                        class="pi pi-shield text-green-500 dark:text-green-400"
                    ></i>
                    <span>Private</span>
                </div>
                <div class="flex items-center gap-1">
                    <i
                        class="pi pi-database text-blue-500 dark:text-blue-400"
                    ></i>
                    <span>Local Stored</span>
                </div>
                <div class="flex items-center gap-1">
                    <i
                        class="pi pi-sync text-purple-500 dark:text-purple-400"
                    ></i>
                    <span>Synced</span>
                </div>
            </div>
        </div>

        <div class="flex flex-col gap-4 w-full mb-[100px] max-w-2xl relative">
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
                        <i
                            :class="[
                                suggestion.icon,
                                'text-gray-500 dark:text-gray-300 text-sm group-hover:scale-110 transition-transform',
                            ]"
                        ></i>
                        <span
                            class="text-xs font-medium text-gray-700 dark:text-gray-300"
                        >
                            {{ suggestion.title }}
                        </span>
                    </button>
                </div>
            </div>

            <!-- Start Writing Button -->
            <button
                v-if="!showInput"
                @click="setShowInput"
                class="group px-6 py-3 bg-gradient-to-r from-blue-500 to-purple-600 dark:from-blue-600 dark:to-purple-700 text-white rounded-lg hover:from-blue-600 hover:to-purple-700 dark:hover:from-blue-700 dark:hover:to-purple-800 transition-all duration-300 transform hover:scale-105 shadow-lg hover:shadow-xl font-medium"
            >
                <span class="flex items-center justify-center gap-2">
                    <i
                        class="pi pi-pencil group-hover:rotate-12 transition-transform"
                    ></i>
                    Start Writing
                </span>
            </button>
        </div>
    </div>
</template>
