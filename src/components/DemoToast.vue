<script setup lang="ts">
import type { Ref } from "vue";
import { watch, inject, h, ref } from "vue";
import { toast } from "vue-sonner";

const { isDemoMode } = inject("globalState") as {
    isDemoMode: Ref<boolean>;
};

// Track current toast ID
const currentToastId = ref<string | number | undefined>(undefined);

// Demo Mode Toast Component
const DemoModeToast = () =>
    h(
        "div",
        {
            class: "flex items-center gap-3 bg-gradient-to-r from-amber-50 to-orange-50 dark:from-amber-900/20 dark:to-orange-900/20 border-2 border-amber-300 dark:border-amber-700 rounded-lg shadow-lg px-4 py-3 min-w-[320px]",
        },
        [
            // Icon container
            h(
                "div",
                {
                    class: "flex-shrink-0 flex items-center justify-center w-8 h-8 sm:w-10 sm:h-10 bg-gradient-to-br from-amber-400 to-orange-500 rounded-full shadow-md",
                },
                [
                    h(
                        "svg",
                        {
                            class: "w-4 h-4 sm:w-5 sm:h-5 text-white",
                            fill: "currentColor",
                            viewBox: "0 0 20 20",
                        },
                        [
                            h("path", {
                                d: "M10.894 2.553a1 1 0 00-1.788 0l-7 14a1 1 0 001.169 1.409l5-1.429A1 1 0 009 15.571V11a1 1 0 112 0v4.571a1 1 0 00.725.962l5 1.428a1 1 0 001.17-1.408l-7-14z",
                            }),
                        ],
                    ),
                ],
            ),
            // Content
            h("div", { class: "flex-1 min-w-0" }, [
                h(
                    "p",
                    {
                        class: "text-xs sm:text-sm font-semibold text-amber-900 dark:text-amber-100 truncate",
                    },
                    "Demo Mode",
                ),
                h(
                    "p",
                    {
                        class: "text-[10px] sm:text-xs text-amber-700 dark:text-amber-300 mt-0.5 line-clamp-2",
                    },
                    "Data will not be saved permanently",
                ),
            ]),
            // Close button
            h(
                "button",
                {
                    class: "flex-shrink-0 flex items-center justify-center w-7 h-7 sm:w-8 sm:h-8 text-amber-600 dark:text-amber-400 hover:text-amber-800 dark:hover:text-amber-200 hover:bg-amber-200 dark:hover:bg-amber-800 rounded-full transition-colors",
                    onClick: (e: Event) => {
                        e.stopPropagation();
                        toast.dismiss(currentToastId.value);
                        hasShownToast.value = false;
                        currentToastId.value = undefined;
                    },
                },
                [
                    h(
                        "svg",
                        {
                            class: "w-3.5 h-3.5 sm:w-4 sm:h-4",
                            fill: "none",
                            stroke: "currentColor",
                            viewBox: "0 0 24 24",
                        },
                        [
                            h("path", {
                                "stroke-linecap": "round",
                                "stroke-linejoin": "round",
                                "stroke-width": "2",
                                d: "M6 18L18 6M6 6l12 12",
                            }),
                        ],
                    ),
                ],
            ),
        ],
    );

const showDemoModeToast = () => {
    // Dismiss any existing demo toast first
    if (currentToastId.value !== null) {
        toast.dismiss(currentToastId.value);
    }

    // Show new toast and store its ID
    const toastId = toast.custom(DemoModeToast, {
        duration: Infinity,
        position: "top-center",
    });

    currentToastId.value = toastId;
};

// Use ref for hasShownToast to make it reactive
const hasShownToast = ref(false);

// Watcher to show toast when demo mode becomes true
watch(
    isDemoMode,
    (isDemo) => {
        if (isDemo && !hasShownToast.value) {
            showDemoModeToast();
            hasShownToast.value = true;
        } else if (!isDemo) {
            // Dismiss toast and reset flag when user logs out of demo mode
            if (currentToastId.value !== undefined) {
                toast.dismiss(currentToastId.value);
                currentToastId.value = undefined;
            }
            hasShownToast.value = false;
        }
    },
    { immediate: true },
);
</script>
<template>
    <div></div>
</template>
