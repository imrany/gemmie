import { Router } from "express";
import * as marked from "marked";
import removeMd from "remove-markdown"
import { GoogleGenerativeAI } from "@google/generative-ai";
import { config } from "dotenv";
import { Prompt, PromptResponse } from "../types";
config();

let router=Router()

// Access your API key as an environment variable (see "Set up your API key" above)
let apiKey:any=process.env.API_KEY;
const genAI = new GoogleGenerativeAI(apiKey);

router.post("/",async(req:Prompt,res:any)=>{
    try {
        let { prompt }= req.body;
        const { generateContent } = genAI.getGenerativeModel({ model: "gemini-pro"});
        const { response } = await generateContent(prompt);
        let text=removeMd(response.text())
        res.status(200).send(text)
    } catch (error:any) {
        res.status(500).send({error:error.message})
        console.log(error.message)
    }
})

router.post("/md",async(req:Prompt,res:any)=>{
    try {
        let { prompt }= req.body;
        const { generateContent } = genAI.getGenerativeModel({ model: "gemini-pro"});
        const { response } = await generateContent(prompt);
        let markdown=response.text()
        res.status(200).send(markdown)
    } catch (error:any) {
        res.status(500).send({error:error.message})
        console.log(error.message)
    }
})

router.post("/prompt",async(req:Prompt,res:any)=>{
    try {
        let { prompt }= req.body;
        // For text-only input, use the gemini-pro model
        const { generateContent } = genAI.getGenerativeModel({ model: "gemini-pro"});
        const { response } = await generateContent(prompt);
        let promptResponse:PromptResponse={
            prompt,
            text: marked.parse(response.text())
        }
        res.status(200).send(promptResponse)
    } catch (error:any) {
        res.status(500).send({error:error.message})
        console.log(error.message)
    }
})

export default router;
