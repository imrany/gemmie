import { Router } from "express";
import { GoogleGenerativeAI } from "@google/generative-ai";
import { config } from "dotenv";
import { Prompt, PromptResponse } from "../types";
config();

let router=Router()

// Access your API key as an environment variable (see "Set up your API key" above)
let apiKey:any=process.env.API_KEY;
const genAI = new GoogleGenerativeAI(apiKey);

router.post("/prompt",async(req:Prompt,res:any)=>{
    try {
        let prompt= req.body.prompt;
        // For text-only input, use the gemini-pro model
        const model = genAI.getGenerativeModel({ model: "gemini-pro"});
        const result = await model.generateContent(prompt);
        const response = result.response;
        let promptResponse:PromptResponse={
            text: response.text()
        }
        res.status(200).send(promptResponse)
    } catch (error:any) {
        res.status(501).send({error:error.message})
        console.log(error.message)
    }
})

export default router;