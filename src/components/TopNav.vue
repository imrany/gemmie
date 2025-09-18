<script lang="ts" setup>
let props=defineProps<{
    data:{
        res:{
            response: string,
            prompt?: string,
            status?: number,
        }[],
        isCollapsed?:boolean,
        parsedUserDetails:{
            username:string
        },
        screenWidth:any,
        isSidebarHidden?:boolean,
    },
    functions:{
        hideSidebar:any,
        clearAllChats:any,
    }
}>()
</script>
<template>
    <div class="h-[44px] bg-white z-20 fixed top-0 right-0 border-b-[1px] transition-all duration-300 ease-in-out" :style="props.data.screenWidth>720&&!props.data.isCollapsed?'left:270px':props.data.screenWidth>720&&props.data.isCollapsed?'left:60px;':'left:0;'">
        <div class="flex h-full px-5 items-center justify-between w-full">
            <p v-if="props.data.parsedUserDetails.username!==undefined" class="my-3 text-black text-lg font-light">{{props.data.parsedUserDetails.username}}</p>
            <p v-else class="my-3 text-black text-lg font-light">Gemmie</p>
            <div v-if="props.data.screenWidth < 720" class="my-3 flex gap-2 items-center ml-auto">
                <!-- Sidebar Toggle Icon -->
                <button
                    @click="props.functions.hideSidebar"
                    title="Toggle Sidebar"
                    class="w-[30px] h-[30px] flex items-center justify-center hover:bg-gray-100 rounded-full cursor-pointer"
                >
                    <span v-if="props.data.isSidebarHidden" class="pi pi-bars text-base"></span>
                    <span v-else class="pi pi-times text-base"></span>
                </button>
            </div>
            <div v-else class="flex gap-2 items-center ml-auto">
                <button
                    @click="props.functions.clearAllChats"
                    title="Delete Chat"
                    v-if="props.data.res.length!==0"
                    class="w-full flex items-center bg-none text-gray-500 hover:text-red-500 gap-2 border hover:border-red-500 rounded-full py-1 px-3"
                >
                    <i class="pi pi-trash mb-[2px]"></i>
                    <p>Delete Chat</p>
                </button>
            </div>
        </div>
    </div>
</template>
