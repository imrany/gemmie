<script lang="ts" setup>
    let props=defineProps<{
        data:{
            res:{
                text:string,
                prompt?:string
            }[],
            parsedUserDetails:{
                username:string
            },
            screenWidth:any,
        },
        functions:{
            setShowInput:any,
            toggleSideNav:any,
        }
    }>()
    function reload(){
        window.location.reload()
    }
    function clear(){
        localStorage.removeItem("chats");
        reload()
    }
    function handleSubmit(e:any){
        e.preventDefault()
        let userDetails={
            username:e.target.username.value
        }
        localStorage.setItem("userdetails",JSON.stringify(userDetails))
        e.target.reset()
        reload()
    }
</script>
<template>
    <div id="side_nav" :class="props.data.screenWidth>720?'':'none'" class="bg-white z-5 fixed top-0 left-0 bottom-0 border-r-[1px]" :style="props.data.screenWidth>720?'width:300px;':'right:0; z-index:10;'">
        <div class="flex items-center justify-between p-3">
            <p class="font-semibold text-xl text-black">Gemini</p>
            <div v-if="props.data.screenWidth>720" class="flex gap-2 items-center">
                <button @click="reload" title="Refresh Page" class="w-[30px] h-[30px] flex items-center justify-center hover:bg-gray-100 rounded-[50px] cursor-pointer">
                    <span class="pi pi-refresh text-sm"></span>
                </button>
                <button v-if="props.data.res.length!==0" @click="clear" title="Clear Chats" class="w-[30px] h-[30px] flex items-center justify-center hover:bg-gray-100 rounded-[50px] cursor-pointer">
                    <span class="pi pi-trash text-sm"></span>
                </button>
            </div>
            <div v-else>
                <button @click="props.functions.toggleSideNav" title="Close" class="w-[30px] h-[30px] flex items-center justify-center hover:bg-gray-100 rounded-[50px] cursor-pointer">
                    <span class="pi pi-times text-xl"></span>
                </button>

            </div>
        </div>
        <div v-if="props.data.res.length!==0&&props.data.parsedUserDetails.username.length!==0" class="flex items-center px-4 cursor-pointer bg-slate-200 h-[70px]">
            <div class="flex gap-2 items-center flex-grow">
                <div class="w-[40px] h-[40px] flex justify-center items-center bg-gray-100 rounded-[50px]">
                    <span class="pi pi-user text-sm"></span>
                </div>
                <div class="flex flex-col text-xs gap-1">
                    <p>{{props.data.parsedUserDetails.username}}</p>
                    <p>You: {{ props.data.res.slice(props.data.res.length-1,props.data.res.length)[0].prompt }}</p>
                </div>
            </div>
            <div class="ml-auto text-xs">
                <p class="text-xs capitalize">Today</p>
            </div>
        </div>
        <div v-else-if="props.data.parsedUserDetails.username===undefined||props.data.parsedUserDetails.username.length===0" class="flex flex-col items-center text-sm mt-7 w-full px-4 cursor-pointer justify-center">
            <form @submit="handleSubmit" class="flex gap-2 justify-center flex-col">
                <input required id="username" name="username" type="text" class="w-[250px] focus:outline-none active:outline-none outline-none border-none focus:border-none px-3 bg-gray-100 placeholder:text-gray-500 text-sm h-[35px] flex-grow py-1 rounded-md" placeholder="Enter your username"/>
                <button class="rounded-md flex justify-center items-center bg-gray-800 text-white h-[37px] px-5">Submit</button>
            </form>
        </div>
        <div v-else-if="props.data.parsedUserDetails.username.length!==0" class="flex flex-col mt-7 px-3 text-sm justify-center">
            <div class="flex justify-center flex-col">
                <p class="text-lg font-semibold">Hello {{ props.data.parsedUserDetails.username }},</p>
                <p class="text-sm">Welcome to Gemmie, you are using Gemmie Demo, Google AI Hackathon.</p>
                <button @click="props.functions.setShowInput()" class="rounded-md flex justify-center items-center bg-gray-200 h-[40px]  mt-2 px-3">Write a prompt</button>
            </div>
        </div>
    </div>
</template>
