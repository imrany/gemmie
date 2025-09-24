<script setup lang="ts">
import type { Chat } from '@/types';


const props = defineProps<{
  data: {
    activeChatMenu: string | null,
    chat:Chat | undefined,
    screenWidth : number,
  },
  functions: {
    deleteChat: (chatId: string) => void,
    startRename?: (chatId: string, chatTitle:string)=>void
    hideSidebar:() => void
  }
}>()

</script>
<template>
    <!-- Chat Dropdown Menu -->
    <transition name="fade">
        <div v-if="props.data.activeChatMenu === props.data.chat?.id"
            class="absolute top-8 right-0 bg-white border rounded-lg shadow-lg text-sm z-50 min-w-[120px]  max-md:min-w-[200px]" @click.stop>
            <button 
                @click="()=>{
                    if(props.functions.startRename&&props.data.chat){
                        props.functions.startRename(props.data.chat.id, props.data.chat.title || 'Untitled Chat')
                    }
                }"
                class="w-full flex items-center gap-2 text-left px-3 py-2 hover:bg-gray-100 rounded-t-lg">
                <i class="pi pi-pencil text-xs max-md:text-base"></i>
                <span class="max-md:text-base">Rename</span>
            </button>
            <button @click="() => { 
                props.functions.deleteChat(props.data.chat?.id||''); 
                props.data.activeChatMenu = null; 
                if (props.data.screenWidth < 720) props.functions.hideSidebar()
            }"
                class="w-full flex items-center gap-2 text-left px-3 py-2 text-red-600 hover:bg-red-100 rounded-b-lg">
                <i class="pi pi-trash text-xs max-md:text-base"></i>
                <span class="max-md:text-base">Delete</span>
            </button>
        </div>
    </transition>
</template>