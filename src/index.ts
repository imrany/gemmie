import express from "express";
import cors from "cors";
import router from "./routes"
import { config } from "dotenv";
config();

const app=express();
const cors_option = {
  origin:["http://localhost:3000",],
  methods: ["GET", "POST", "DELETE", "UPDATE", "PATCH", "PUT"]
}

//middleware
app.use(express.json())
app.use(cors(cors_option))
app.use(express.urlencoded({extended:false}))
// app.set('view engine','ejs');
app.use('/api',router)

//routes
app.get("/",(req:any,res:any)=>{
  res.send(`Hello gemmie`)
});

let port=8000||process.env.PORT
app.listen(port,()=>{
  console.log(`Server running on port ${port} \nView... http://127.0.0.1:8000`)
})
