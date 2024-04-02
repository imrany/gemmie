<script lang="ts" setup>
let props=defineProps<{
    data:{
        res:any,
        note: string,
        parsedUserDetails:{
            username:string
        }
    }
}>()
</script>
<template>
    <div class="h-[44px] bg-white z-5 left-[300px] fixed top-0 right-0 border-b-[1px]">
        <div class="flex px-5 py-3 items-center justify-between w-full">
            <p v-if="props.data.parsedUserDetails.username!==undefined" class="text-black text-sm font-semibold">{{props.data.parsedUserDetails.username}}</p>
            <p v-else class="text-black text-sm font-semibold">Gemmie</p>
        </div>
        <div class="bg-blue-200 flex items-center justify-center text-xs h-[20px]">
            <p>{{ props.data.note }}</p>
        </div>
    </div>
</template>