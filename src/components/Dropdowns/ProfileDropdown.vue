<script lang="ts" setup>
import type { UserDetails } from "@/types";
import { LogOut } from "lucide-vue-next";
import type { FunctionalComponent } from "vue";
import type { Ref } from "vue";
import { inject } from "vue";
import { Button } from "../ui/button";

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
            class="absolute px-2 pb-2 max-w-[245px] max-md:text-base bottom-full left-3 right-3 mb-2 bg-white dark:bg-gray-900 border dark:border-gray-700 rounded-lg shadow-lg text-sm z-50"
        >
            <div class="py-3 px-1">
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
            <Button
                variant="ghost"
                v-for="option in props.data.profileOptions"
                :key="option.id"
                @click="props.functions.handleNavAction(option.action)"
                class="font-medium group text-gray-500 dark:text-gray-500 my-1 hover:text-gray-800 hover:dark:text-gray-200 hover:bg-gray-200 h-[32px] px-2 py-0 dark:hover:bg-gray-700/50 text-sm justify-start w-full flex items-center rounded-md relative transition-all duration-150"
            >
                <component :is="option.icon" class="w-4 h-4" />
                {{ option.label }}
            </Button>
            <Button
                variant="ghost"
                @click="props.functions.handleNavAction(props.functions.logout)"
                class="font-medium group text-red-600 dark:text-red-400 hover:text-red-600 hover:dark:text-red-400 hover:bg-red-100 dark:hover:bg-red-900/20 h-[30px] px-2 py-0 text-sm justify-start w-full flex items-center rounded-md relative transition-all duration-150"
            >
                <LogOut class="w-4 h-4" />
                Log Out
            </Button>
        </div>
    </transition>
</template>
