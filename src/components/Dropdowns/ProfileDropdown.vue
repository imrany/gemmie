<script lang="ts" setup>
import type { UserDetails } from "@/types";
import { LogOut } from "lucide-vue-next";
import type { FunctionalComponent } from "vue";
import type { Ref } from "vue";
import { inject } from "vue";

const { parsedUserDetails, planStatus } = inject("globalState") as {
    parsedUserDetails: Ref<UserDetails>;
    planStatus: any;
};

const props = defineProps<{
    data: {
        showProfileMenu: any;
        planColor: string;
        profileOptions: {
            id: string;
            label: string;
            icon: FunctionalComponent<any>;
            action: () => void;
        }[];
    };
    functions: {
        logout: () => void;
        handleNavAction: (action: () => void) => void;
    };
}>();
</script>

<template>
    <transition name="fade">
        <div
            v-if="props.data.showProfileMenu"
            class="absolute max-w-[245px] max-md:text-base bottom-full left-3 right-3 mb-2 bg-white dark:bg-gray-900 border dark:border-gray-700 rounded-lg shadow-lg text-sm z-50"
        >
            <div class="px-4 py-2 border-b dark:border-gray-700">
                <p class="text-gray-500 dark:text-gray-400 font-medium">
                    {{
                        parsedUserDetails.email.length > 20
                            ? parsedUserDetails.email
                                  .trim()
                                  .split("@")[0]
                                  .slice(0, 10) +
                              "..." +
                              parsedUserDetails.email.trim().split("@")[1]
                            : parsedUserDetails.email || "No email"
                    }}
                </p>
                <div
                    v-if="parsedUserDetails.planName"
                    class="mt-1 text-xs font-normal"
                    :class="props.data.planColor"
                >
                    {{ parsedUserDetails.planName }}
                    <span v-if="planStatus.status === 'active'">
                        {{ planStatus.timeLeft }}</span
                    >
                    <span v-else-if="planStatus.isExpired"> Expired</span>
                </div>
            </div>
            <button
                v-for="option in props.data.profileOptions"
                :key="option.id"
                @click="props.functions.handleNavAction(option.action)"
                class="w-full flex gap-2 text-gray-600 dark:text-gray-400 items-center text-left px-4 py-2 hover:bg-gray-100 dark:hover:bg-gray-700/20 transition-colors"
            >
                <component :is="option.icon" class="w-4 h-4" />
                {{ option.label }}
            </button>
            <button
                @click="props.functions.handleNavAction(props.functions.logout)"
                class="w-full text-left flex gap-2 items-center px-4 py-2 text-red-600 dark:text-red-400 hover:bg-red-100 dark:hover:bg-red-900/20 rounded-b-lg transition-colors"
            >
                <LogOut class="w-4 h-4" />
                Log Out
            </button>
        </div>
    </transition>
</template>
