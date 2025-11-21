<script lang="ts" setup>
import ProtectedPage from "@/layout/ProtectedPage.vue";
import type { UserDetails } from "@/types";
import type { Ref } from "vue";
import { inject } from "vue";

const { screenWidth, parsedUserDetails, isCollapsed } = inject(
    "globalState",
) as {
    screenWidth: Ref<number>;
    parsedUserDetails: Ref<UserDetails>;
    isCollapsed: Ref<boolean>;
};
</script>
<template>
    <ProtectedPage>
        <!-- Main Content - Centered -->
        <div
            :class="[
                'flex-grow flex flex-col transition-all duration-300 ease-in-out h-screen overflow-hidden min-w-0',
                screenWidth > 720 && parsedUserDetails?.username
                    ? isCollapsed
                        ? 'ml-[60px]'
                        : 'ml-[270px]'
                    : '',
            ]"
        ></div>
    </ProtectedPage>
</template>
