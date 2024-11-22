import express from "express";
import cors from "cors";
import axios from "axios";
import router from "./routes";
import { config } from "dotenv";
config();

const app=express();
const cors_option = {
    //origin:["http://localhost:3000","https://gemmie-hackathon-demo.web.app"],
    origin:"*",
    methods: ["GET", "POST", "DELETE", "UPDATE", "PATCH", "PUT"]
}

//middleware
app.use(express.json())
app.use(cors(cors_option))
app.use(express.urlencoded({extended:false}))
app.use('/api',router)

let port=process.env.PORT||8000
app.listen(port,()=>{
  console.log(`Server running on port ${port}`)
})

// sends an HTTP GET request to your server to prevent it from idling after 30 minutes of inactivity (optional)
setInterval(async()=> { 
    await axios.get(`${process.env.BASE_URL}`);
}, 3 * 60 * 1000);
