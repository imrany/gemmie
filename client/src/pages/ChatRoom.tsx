import { useState, useContext, useEffect } from "react";
import { MdClose, MdChevronRight, MdDelete } from "react-icons/md";
import { GlobalContext } from "../context";
import { addDoc, collection, db, getDocs, doc, deleteDoc, onSnapshot } from "../firebaseConfig/config";
import { Chat } from "../types/definitions";
import { err_toast } from "../components/Feedback";

function ChatRoom() {
    const { email } =useContext(GlobalContext)
    const [showAddReplyForm,setShowAddReplyForm]=useState(false)
    const [showChats,setShowChats]=useState(true)
    const [disable,setDisable]=useState(false)
    let md_screen:boolean=screen.width>930||screen.width===930?true:false;
    const [chats,setChats]=useState<Chat[]>([
        {
            chat_number:0,
            from:"",
            message:"",
            time:"",
            today:""
        }
    ])

    const date=new Date()
    let time=date.getHours()<10?`0${date.getHours()}:${date.getMinutes()}`:`${date.getHours()}:${date.getMinutes()}`;
    let today:string;
    switch (date.getDay()) {
        case 1:
            today="Mon"
            break;
        case 2:
            today="Tue"
            break;
        case 3:
            today="Wed"
            break;
        case 4:
            today="Thu"
            break;
        case 5:
            today="Fri"
            break;
        case 6:
            today="Sat"
            break;
        case 7:
            today="Sun"
            break;
    } 

    const updateChats = onSnapshot(collection(db, "chats"), () => {
        return "changed"
    });

    async function handleReply(e:any) {
        try{
            e.preventDefault()
            setDisable(true)
            let message:Chat={
                chat_number:chats.length+1,
                from:email,
                message:e.target.message.value,
                time,
                today
            }
            await addDoc(collection(db,"chats"),message);
            if(md_screen===false){
                setShowChats(!md_screen)
                setShowAddReplyForm(false)
            }
            e.target.reset()
            setDisable(false)
        }catch(error:any){
            setDisable(false)
            err_toast(error.message)
            console.log(error)
        }
    }

    async function fetchChatsFromFirebase(){
        try {
            const querySnapshot=await getDocs(collection(db,"chats"))
            let list:Chat[]=[]
            querySnapshot.forEach((doc) => {
                let data={
                    id:doc.id,
                    from:doc.data().from,
                    message:doc.data().message,
                    time:doc.data().time,
                    today:doc.data().today,
                    chat_number:doc.data().chat_number
                }
                list.push(data)
            });
            list.sort((a,b)=>{
                return a.chat_number - b.chat_number;
            })
            console.log(list)
            setChats([...list])
        } catch (error:any) {
            console.log(error)
            err_toast(error.message)
        }
    }

    async function handleDeleteText(id:any){
        try{
            await deleteDoc(doc(db,"chats",id))
        }catch(error:any){
            console.log(error.message)
        }
    }

    window.onresize=function(){
        if(screen.width<768){
            setShowChats(false)
        }else{
            setShowChats(true)
        }
    }
    useEffect(()=>{
        fetchChatsFromFirebase()
        window.scrollTo(0,0)
    },[updateChats])
    return (
        <div className="p-10 flex gap-4 max-sm:flex-wrap min-h-[50vh]">
            {showChats?(<div className="flex-grow border-r-[2px] pr-3 border-dotted">
                <div className="flex sidebar flex-col overflow-y-auto h-[65vh] gap-6 text-sm mb-8">
                    {chats.length!==0?chats.map((chat)=>{
                        return(
                            <div className="flex flex-col py-8 border-b-[1px]" key={chat.id}>
                                <div className="flex items-center justify-between">
                                    <p>{email!==chat.from?(<span>From: <span className="text-[var(--theme-blue)]">{chat.from}</span></span>):(<span>You: </span>)}</p>
                                    <p className="text-gray-500 text-xs mr-4">{chat.today}, {chat.time}</p>
                                </div>
                                <p className="text-base">{chat.message}</p>
                                {email!==chat.from?"":(
                                    <div className="ml-auto flex gap-3">
                                        <MdDelete title="Delete this text" onClick={()=>handleDeleteText(chat.id)} className="w-6 h-6 cursor-pointer hover:text-red-600"/>
                                    </div>
                                )}
                            </div>
                        )
                    }):(
                        <tr className="flex h-full text-sm gap-2 items-center justify-center">
                            <p className="">No chats yet</p>
                            <button 
                                className="underline text-[var(--theme-blue)]" 
                                onClick={()=>{
                                    setShowAddReplyForm(true)
                                    if(screen.width<768){
                                        setShowChats(false)
                                    }
                                }}
                            >
                                Start chatting
                            </button>
                        </tr>
                    )}
                </div>
                {!showAddReplyForm&&chats.length!==0?(
                    <button 
                        className="underline text-[var(--theme-blue)] flex items-center" 
                        onClick={()=>{
                            setShowAddReplyForm(true)
                            if(screen.width<768){
                                setShowChats(false)
                            }
                        }}
                    >
                        <span>Reply</span>
                        <MdChevronRight className="w-4 h-4"/>
                    </button>
                ):""}
            </div>):""}
            {showAddReplyForm?(
                <div className="h-[65vh] flex flex-col justify-center max-md:w-[90vw] w-[50vw] lg:w-[25vw]">
                    <div className="md:border-[1px] text-sm md:p-4 rounded-lg">
                        <div className="flex justify-between items-center ">
                            <p className="text-[20px] font-semibold">Chat Room</p>
                            <button title="close" 
                                onClick={()=>{
                                    setShowAddReplyForm(false)
                                    if(screen.width<768){
                                        setShowChats(true)
                                    }
                                }}
                            >
                                <MdClose className="w-5 h-5"/>
                            </button>
                        </div>
                        <form onSubmit={handleReply} className="flex mt-5 flex-col text-sm">
                            <label className="mb-[8px] text-[#0f172a]" htmlFor="message">Message <span className="text-red-500">*</span></label>
                            <div className="pb-4">
                                <input id="message" name="message" type="text" className={`px-[10px] w-full py-2 focus:outline-[var(--theme-blue)] focus:outline-[1px] bg-white border-[1px] rounded-lg`} placeholder="Hey there..." required/>
                            </div>
                            <button disabled={disable} className={disable===true?"cursor-wait mt-5 capitalize py-3 px-6 text-white rounded-md bg-[var(--theme-dark)]":"mt-5 capitalize py-3 px-6 text-white rounded-md bg-[var(--theme-blue)]"}>
                            {disable===false?(<span>
                                Send Text
                            </span>):(
                                <i className="italic">Sending Text...</i>
                            )}
                            </button>
                        </form>
                    </div>
                </div>
            ):""}
        </div>
    );
};

export default ChatRoom;
