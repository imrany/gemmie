<script setup lang="ts">
import { Skeleton } from "@/components/ui/skeleton";
import type { UserDetails } from "@/types";
import type { FunctionalComponent } from "vue";
import type { Ref } from "vue";
import { inject } from "vue";

const { parsedUserDetails, isDarkMode, screenWidth, isChatLoading } = inject(
    "globalState",
) as {
    parsedUserDetails: UserDetails;
    isDarkMode: Ref<boolean>;
    screenWidth: Ref<number>;
    isChatLoading: Ref<boolean>;
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
    <div
        :class="
            screenWidth > 720
                ? 'h-screen bg-inherit flex flex-col items-center justify-center'
                : 'bg-inherit h-screen flex flex-col items-center justify-center'
        "
    >
        <div
            v-if="isChatLoading"
            class="relative md:max-w-3xl min-h-[calc(100vh-200px)] max-w-[100vw] flex-grow no-scrollbar overflow-y-auto px-2 w-full space-y-3 sm:space-y-4 mt-[65px] scroll-container"
        >
            <div class="flex flex-col gap-4">
                <!-- User message skeleton -->
                <div
                    class="flex animate-pulse items-start gap-3 font-medium bg-gray-100 dark:bg-gray-800 text-black dark:text-gray-100 px-4 py-3 rounded-2xl w-full ml-auto"
                >
                    <!-- Avatar container -->
                    <div class="flex-shrink-0">
                        <div
                            class="flex items-center justify-center w-7 h-7 text-gray-100 dark:text-gray-800 bg-gray-700 dark:bg-gray-200 rounded-full text-sm font-semibold"
                        >
                            {{
                                parsedUserDetails.username
                                    .toUpperCase()
                                    .slice(0, 2)
                            }}
                        </div>
                    </div>
                    <!-- Message content skeleton -->
                    <div class="flex-1 min-w-[200px] max-w-[500px]">
                        <div
                            class="h-4 bg-gray-300 dark:bg-gray-600 rounded w-3/4"
                        ></div>
                    </div>
                </div>

                <!-- Assistant response skeleton -->
                <div class="flex flex-col gap-2 w-full max-w-full">
                    <Skeleton
                        v-for="(height, index) in [4, 3, 5, 4, 3, 2]"
                        :key="index"
                        :class="[
                            'chat-message bg-gray-100 dark:bg-gray-800 rounded-lg',
                            `h-${height}`,
                            'w-full',
                            index === 2 ? 'mt-2' : '',
                        ]"
                    />
                </div>
            </div>
        </div>

        <div
            class="relative h-full flex mt-16 md:max-w-3xl max-w-[100vw] max-md:px-4 flex-col md:flex-grow items-center gap-3 text-gray-600 dark:text-gray-400"
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
            <div class="text-center max-w-md space-y-4 gap-4">
                <p class="text-gray-700 dark:text-gray-300 leading-relaxed">
                    I'm here to assist you with any questions or tasks you might
                    have. What can I do for you today?
                </p>
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
