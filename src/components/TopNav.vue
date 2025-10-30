<script lang="ts" setup>
import type { CurrentChat } from "@/types";
import {
    AlignJustify,
    CheckCircle,
    CloudUpload,
    LoaderCircle,
} from "lucide-vue-next";
import type { Ref } from "vue";
import { inject } from "vue";

const {
    hideSidebar,
    isAuthenticated,
    screenWidth,
    syncStatus,
    manualSync,
    isCollapsed,
    currentChat,
} = inject("globalState") as {
    isCollapsed: Ref<boolean>;
    currentChat: Ref<CurrentChat | undefined>;

    isAuthenticated: Ref<boolean>;
    syncStatus: Ref<{
        lastSync: Date | null;
        syncing: boolean;
        hasUnsyncedChanges: boolean;
        showSyncIndicator: boolean;
        syncMessage: string;
        syncProgress: number;
        lastError: string | null;
        retryCount: number;
        maxRetries: number;
    }>;
    screenWidth: Ref<number>;
    hideSidebar: () => void;
    manualSync: () => void;
};
</script>

<template>
    <div
        class="bg-white dark:bg-gray-900 h-[52px] z-30 fixed top-0 right-0 transition-all duration-300 ease-in-out"
        :style="
            screenWidth > 720 && !isCollapsed
                ? 'left:270px'
                : screenWidth > 720 && isCollapsed
                  ? 'left:60px'
                  : 'left:0'
        "
    >
        <div class="flex h-full px-4 items-center justify-between w-full">
            <!-- Brand -->
            <p
                v-if="currentChat && screenWidth > 720"
                class="text-gray-600 dark:text-gray-400 font-medium truncate text-sm select-none"
            >
                <span v-if="currentChat.title.length > 30"
                    >{{ currentChat.title.slice(0, 30) }}...</span
                >
                <span v-else>{{ currentChat.title }}</span>
            </p>
            <p
                v-else-if="screenWidth <= 720"
                class="text-gray-700 dark:text-gray-300 text-xl max-md:text-2xl font-semibold tracking-wide select-none"
            >
                Gemmie
            </p>

            <div class="flex gap-3 items-center ml-auto">
                <!-- Sync Status -->
                <div v-if="isAuthenticated" class="relative">
                    <div
                        v-if="syncStatus.syncing"
                        class="flex items-center gap-2 bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-400 px-3 py-1.5 rounded-full text-xs border border-blue-200 dark:border-blue-800 shadow-sm animate-pulse"
                    >
                        <LoaderCircle class="w-4 h-4 animate-spin" />
                        <span>Syncing...</span>
                    </div>

                    <div
                        v-else-if="syncStatus.hasUnsyncedChanges"
                        class="flex items-center gap-2 bg-orange-50 dark:bg-orange-900/20 text-orange-700 dark:text-orange-400 px-3 py-1.5 rounded-full text-xs border border-orange-200 dark:border-orange-800 shadow-sm cursor-pointer hover:bg-orange-100 dark:hover:bg-orange-900/30 transition"
                        @click="manualSync"
                    >
                        <CloudUpload class="w-4 h-4" />
                        <span>Sync pending</span>
                    </div>

                    <div
                        v-else-if="syncStatus.lastSync"
                        class="flex items-center gap-2 bg-green-50 dark:bg-green-900/20 text-green-700 dark:text-green-400 px-3 py-1.5 rounded-full text-xs border border-green-200 dark:border-green-800 shadow-sm"
                    >
                        <CheckCircle class="w-4 h-4" />
                        <span>Synced</span>
                    </div>
                </div>

                <!-- Mobile Sidebar Toggle -->
                <div
                    v-if="screenWidth < 720"
                    class="flex gap-2 items-center ml-auto"
                >
                    <button
                        @click="hideSidebar"
                        title="Toggle Sidebar"
                        class="w-9 h-9 flex items-center justify-center text-lg hover:bg-gray-100 dark:hover:bg-gray-800 rounded-full cursor-pointer transition-colors"
                    >
                        <AlignJustify
                            class="w-5 h-5 text-gray-700 dark:text-gray-300"
                        />
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>
