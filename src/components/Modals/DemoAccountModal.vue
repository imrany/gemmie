<script setup lang="ts">
import type { Ref } from "vue";
import { inject } from "vue";
import DialogBox from "../Dialog/DialogBox.vue";

const { isDarkMode } = inject("globalState") as {
    isDarkMode: Ref<boolean>;
};

const { showDemoModal, closeDemoModal, createRealAccount, startDemoMode } =
    defineProps<{
        showDemoModal: boolean;
        createRealAccount: () => void;
        startDemoMode: () => void;
        closeDemoModal: () => void;
    }>();
</script>
<template>
    <div>
        <DialogBox
            :close-modal="closeDemoModal"
            name="demoAccountModal"
            :show="showDemoModal"
        >
            <!-- Icon -->
            <img
                :src="isDarkMode ? '/logo-light.svg' : '/logo.svg'"
                alt="Gemmie Logo"
                class="w-[54px] h-[54px] mx-auto mb-5 rounded-md"
            />

            <!-- Title -->
            <h2
                class="text-2xl font-bold text-center text-gray-900 dark:text-white mb-2"
            >
                Welcome to Gemmie!
            </h2>
            <p class="text-center text-gray-600 dark:text-gray-300 mb-6">
                Try our demo account first or create your own session
            </p>

            <!-- Options -->
            <div class="space-y-3">
                <!-- Demo Account Option -->
                <button
                    @click="startDemoMode"
                    class="group w-full p-4 rounded-xl bg-gradient-to-r from-blue-500 to-purple-600 hover:from-blue-600 hover:to-purple-700 transition-all duration-300 transform hover:scale-[1.02] shadow-lg"
                >
                    <div class="flex items-center justify-between text-white">
                        <div class="flex items-center gap-3">
                            <div
                                class="w-10 h-10 bg-white/20 rounded-lg flex items-center justify-center"
                            >
                                <i class="pi pi-bolt text-xl"></i>
                            </div>
                            <div class="text-left">
                                <p class="font-semibold">Try Demo Account</p>
                                <p class="text-xs opacity-90">
                                    Explore features instantly
                                </p>
                            </div>
                        </div>
                        <i
                            class="pi pi-arrow-right group-hover:translate-x-1 transition-transform"
                        ></i>
                    </div>
                </button>

                <!-- Create Real Account Option -->
                <button
                    @click="createRealAccount"
                    class="group w-full p-4 rounded-xl border-2 border-gray-200 dark:border-gray-600 hover:border-blue-500 dark:hover:border-blue-500 hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-all duration-300"
                >
                    <div class="flex items-center justify-between">
                        <div class="flex items-center gap-3">
                            <div
                                class="w-10 h-10 bg-gray-100 dark:bg-gray-700 rounded-lg flex items-center justify-center"
                            >
                                <i
                                    class="pi pi-user-plus text-blue-600 dark:text-blue-400 text-xl"
                                ></i>
                            </div>
                            <div class="text-left">
                                <p
                                    class="font-semibold text-gray-900 dark:text-white"
                                >
                                    Create Account
                                </p>
                                <p
                                    class="text-xs text-gray-600 dark:text-gray-400"
                                >
                                    Full features & sync
                                </p>
                            </div>
                        </div>
                        <i
                            class="pi pi-arrow-right text-gray-400 group-hover:translate-x-1 group-hover:text-blue-600 dark:group-hover:text-blue-400 transition-all"
                        ></i>
                    </div>
                </button>
            </div>

            <!-- Demo Info -->
            <div
                class="mt-4 p-3 bg-blue-50 dark:bg-blue-900/20 rounded-lg border border-blue-200 dark:border-blue-800"
            >
                <div class="flex items-start gap-2">
                    <i
                        class="pi pi-info-circle text-blue-600 dark:text-blue-400 text-sm mt-0.5"
                    ></i>
                    <div class="text-xs text-blue-800 dark:text-blue-200">
                        <p class="font-semibold mb-1">Demo Account Features:</p>
                        <ul class="list-disc list-inside space-y-0.5 ml-1">
                            <li>Full AI chat functionality</li>
                            <li>No login required</li>
                            <li>Data not saved permanently</li>
                            <li>Upgrade to real account anytime</li>
                        </ul>
                    </div>
                </div>
            </div>

            <!-- Skip button -->
            <button
                @click="closeDemoModal"
                class="w-full mt-4 text-sm text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 transition-colors"
            >
                I'll decide later
            </button>
        </DialogBox>
    </div>
</template>
