# Gemini
A health program using Google Gemini model (hackathon)
For text-only input, use the gemini-pro model

## Routes
- http://127.0.0.1:8000/api/prompt : A post api route that receives post request

### Test 
Testing `/api/prompt` route on with javascript 
```javascript
async function runPrompt(){
    try{
        let request={
            prompt:"How can i stop nose bleeding?"
        }
        let url="http://127.0.0.1:8000/api/prompt"
        let response =await fetch(url,{
            method:"POST",
            body:JSON.stringify(request),
            headers:{
                "content-type":"application/json"
            }
        })
        let data=await response.json()
        console.log(data)
    }catch(error){
        console.error(error.message)
    }
}
runPrompt();
```
