<script setup lang="ts">
import type { Chat } from "@/types";
import { Pencil, Trash } from "lucide-vue-next";
import { Button } from "../ui/button";

const props = defineProps<{
    data: {
        activeChatMenu: string | null;
        chat: Chat | undefined;
        screenWidth: number;
    };
    functions: {
        deleteChat: (chatId: string) => void;
        startRename?: (chatId: string, chatTitle: string) => void;
        hideSidebar: () => void;
    };
}>();
</script>
<template>
    <!-- Chat Dropdown Menu -->
    <transition name="fade">
        <div
            v-if="props.data.activeChatMenu === props.data.chat?.id"
            class="absolute top-8 right-0 px-1 py-2 bg-white border rounded-lg shadow-lg text-sm z-50 min-w-[150px] max-md:min-w-[200px] dark:bg-gray-800 dark:border-gray-700 dark:shadow-md"
        >
            <Button
                @click="
                    () => {
                        if (props.functions.startRename && props.data.chat) {
                            props.functions.startRename(
                                props.data.chat.id,
                                props.data.chat.title || 'Untitled Chat',
                            );
                        }
                    }
                "
                variant="ghost"
                class="w-full flex font-medium group hover:font-semibold items-center gap-2 text-left px-2 py-1 h-[30px] justify-start hover:bg-gray-100 rounded-t-lg dark:hover:bg-gray-700 dark:text-gray-300"
            >
                <Pencil class="w-4 h-4 max-md:w-5 max-md:h-5" />
                <span class="max-md:text-base">Rename</span>
            </Button>
            <Button
                @click="
                    () => {
                        props.functions.deleteChat(props.data.chat?.id || '');
                        // eslint-disable-next-line vue/no-mutating-props
                        props.data.activeChatMenu = null;
                        if (props.data.screenWidth < 720)
                            props.functions.hideSidebar();
                    }
                "
                variant="ghost"
                class="w-full flex font-medium group hover:font-semibold items-center gap-2 text-left px-2 py-1 h-[30px] justify-start hover:text-red-600 text-red-600 hover:bg-red-100 rounded-b-lg dark:text-red-200 dark:hover:bg-red-900"
            >
                <Trash class="w-5 h-5 sm:h-4 sm:w-4" />
                <span class="max-md:text-base">Delete</span>
            </Button>
        </div>
    </transition>
</template>
