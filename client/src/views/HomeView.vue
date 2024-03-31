<script lang="ts" setup>
    import { ref } from "vue"
    type Res={
        text?:string,
        prompt?:string
    }

    let title=ref("Our Live Gemmie Demo, Give it a try ( Hackathon ).")
    let res:Res[]=[]
    let isLoading=ref(false)
    async function handleSubmit(e:any){
        try {
            e.preventDefault()
            isLoading.value=true
            // let url=`http://127.0.0.1:8000/api/prompt`
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
                prompt:parseRes.prompt,
                text:parseRes.error?parseRes.error:parseRes.text
            }
            res.push(resp)
            isLoading.value=false
            e.target.reset()
        } catch (error:any) {
            let resp:Res={
                prompt:`An error has occurred`,
                text:error.message
            }
            res.push(resp)
            isLoading.value=false
            console.log(error.message)
        }
    }
</script>

<template>
    <nav class="mt-1 bg-yellow-200 text-black font-semibold h-[50px] flex items-center justify-center">
        <p>{{ title }}</p>
    </nav>
    <div class="flex items-center flex-col py-6">
        <div class="w-[80vw] lg:w-[60vw]">
            <p></p>
            <div class="rounded-md p-4 bg-blue-100">
                <p v-if="res.length===0">Gemmie is a medical assistance. This program uses gemini to solve medical issues as a first aid assistant.</p>
                <div v-for="item in res" :key="item.text" v-else class="flex flex-col py-4  gap-2">
                    <div class="flex gap-2 items-center">
                        <p class="font-semibold">You:</p>
                        <p class="capitalize">{{ item.prompt }}</p>
                    </div>
                    <div class="flex gap-2 items-center p-2 rounded-md bg-gray-50">
                        <p>{{ item.text }}</p>
                    </div>
                </div>

                <form v-on:submit="handleSubmit" class="flex w-full gap-2 mt-6">
                    <textarea required id="prompt" name="prompt" type="text" class="focus:outline-none active:outline-none outline-none border-none focus:border-none px-2 h-[35px] flex-grow py-1 rounded-md" placeholder="How to stop a nose bleed?"></textarea>
                    <button :disabled="isLoading" :class="isLoading===false?'ml-auto px-3 hover:bg-green-100 cursor-pointer rounded-md flex items-center justify-center':'ml-auto px-3 cursor-progress rounded-md flex items-center justify-center'">
                        <i v-if="isLoading===false" class="pi pi-send"></i>
                        <i v-else class="pi pi-spin pi-spinner"></i>
                    </button>
                </form>
            </div>
        </div>
    </div>
</template>
