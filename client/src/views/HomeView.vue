<script lang="ts" setup>
    import { ref, onMounted } from "vue"
    import SideNav from "../components/SideNav.vue"
    import TopNav from "../components/TopNav.vue"

    type Res={
        text:string,
        prompt?:string
    }

    let showInput=ref(false)
    let readMore=ref(false)
    let note=ref("You are using Gemmie Demo, Gemini API Hackathon")

    let userDetails:any=localStorage.getItem("userdetails")
    let parsedUserDetails:any=JSON.parse(userDetails)===null?[]:JSON.parse(userDetails)

    let chats:any=localStorage.getItem("chats")
    let parsedChats:any=JSON.parse(chats)===null?[]:JSON.parse(chats)
    let res:Res[]=parsedChats.length===0?[]:parsedChats
    let isLoading=ref(false)
    async function handleSubmit(e:any){
        try {
            e.preventDefault()
            isLoading.value=true
            //let url=`http://127.0.0.1:8000/api/prompt`
            let url=`https://gemmie.onrender.com/api/prompt`
            let response=await fetch(url,{
                method:"POST",
                body:JSON.stringify({
                    prompt:e.target.prompt.value
                }),
                headers:{
                    "content-type":"application/json"
                }
            })
            let parseRes=await response.json()
            let resp:Res={
                prompt:parseRes.error?`An error has occurred`:parseRes.prompt,
                text:parseRes.error?parseRes.error:parseRes.text
            }
            res.push(resp)
            localStorage.setItem("chats",JSON.stringify(res))
            isLoading.value=false
            e.target.reset()
            scrollToBottom()
        } catch (error:any) {
            let resp:Res={
                prompt:`An error has occurred`,
                text:error.message
            }
            res.push(resp)
            localStorage.setItem("chats",JSON.stringify(res))
            isLoading.value=false
            console.log(error.message)
        }
    }

    function setShowInput(){
        showInput.value=true
    }
    function setReadMore(){
        readMore.value=true
    }
    function scrollToBottom(){
        document.getElementById("bottom")?.scrollIntoView()
    }

    onMounted(()=>{
        scrollToBottom()
    })
</script>

<template>
    <div class="flex h-[100vh]">
        <SideNav :data="{res,parsedUserDetails}" :functions="{setShowInput}"/>
        <div class="flex-grow flex-col ml-[300px]">
            <TopNav :data="{note,res,parsedUserDetails}"/>
            <div class="h-screen">
                <div class="mt-2 gap-2 flex flex-col items-center justify-center" :style="res.length!==0?'height:250px;':'height:600px;'">
                    <div v-if="res.length!==0" class="rounded-[50px] bg-gray-200 w-[50px] h-[50px] flex justify-center items-center">
                        <span class="pi pi-user"></span>
                    </div>
                    <div v-else class="rounded-[50px] bg-gray-200 w-[50px] h-[50px] flex justify-center items-center">
                        <span class="pi pi-comment"></span>
                    </div>

                    <div class="flex flex-col justify-center items-center">
                        <p v-if="parsedUserDetails.username!==undefined" class="text-xl font-semibold">{{parsedUserDetails.username}}</p>
                        <p v-else class="text-xl font-semibold">Gemmie</p>
                        <div v-if="res.length===0&&showInput===false" class="text-sm flex flex-col items-center gap-2 justify-center text-gray-500">
                            <p>Gemmie uses Gemini API to solve medical issues as a first aid assistant.</p>
                            <button v-if="parsedUserDetails.username!==undefined" @click="setShowInput" class="flex justify-center items-center mt-2 bg-gray-200 h-[35px] text-black px-3 w-fit rounded-md">Try it now</button>                    
                        </div>
                        <p v-else class="text-sm text-gray-500">Use Gemmie for medical inquiries only.</p>
                    </div>
                </div>
                <div class="pb-[50px]">
                    <div v-for="item in res" :key="item.text" class="flex flex-col">
                        <div class="p-3 flex items-center gap-2 hover:bg-slate-200">
                            <div class="w-[35px] h-[35px] flex justify-center items-center bg-gray-100 rounded-[50px]">
                                <span class="pi pi-user text-sm"></span>
                            </div>
                            <div class="flex text-sm justify-center flex-col">
                                <p class="font-semibold">{{ parsedUserDetails.username }}</p>
                                <p>{{ item.prompt }}</p>
                            </div>
                        </div>
                        <div class="p-3 flex gap-2 hover:bg-gray-200">
                            <div class="w-[35px] h-[35px] flex justify-center items-center bg-gray-100 rounded-[50px]">
                                <span class="pi pi-comment text-sm"></span>
                            </div>
                            <div class="text-sm justify-center flex flex-col">
                                <p class="font-semibold">Gemmie</p>
                                <p v-if="item.text.length>800&&readMore===false">
                                    <div v-html="item.text.slice(0,800)" class="flex flex-col gap-1"></div>
                                    <button @click="setReadMore" class="text-blue-500 mt-2 cursor-pointer">Read more</button>
                                </p>
                                <div v-html="item.text" class="flex flex-col gap-1" v-else></div>
                            </div>
                        </div>
                    </div>
                </div>

                <div v-if="res.length!==0||showInput===true" class="left-[300px] bg-white bottom-0 right-0 fixed z-5 h-[50px] p-2 border-t-[1px]">
                    <form v-on:submit="handleSubmit" class="flex h-full w-full gap-2">
                        <input required id="prompt" name="prompt" type="text" class="focus:outline-none active:outline-none outline-none border-none focus:border-none px-3 bg-gray-200 placeholder:text-gray-500 text-sm h-[35px] flex-grow py-1 rounded-[50px]" placeholder="How to stop a nose bleed?"/>
                        <button :disabled="isLoading" :class="isLoading===false?'ml-auto px-3 hover:bg-green-100 cursor-pointer rounded-md flex items-center justify-center':'ml-auto px-3 cursor-progress rounded-md flex items-center justify-center'">
                            <i v-if="isLoading===false" class="pi pi-send"></i>
                            <i v-else class="pi pi-spin pi-spinner"></i>
                        </button>
                    </form>
                </div>
                <div id="bottom"></div>
            </div>
        </div>
    </div>
</template>
