import { useState } from "react"
import { FaEye, FaEyeSlash, FaInfoCircle } from "react-icons/fa";
import { auth, signInWithEmailAndPassword } from "../firebaseConfig/config";
import { err_toast, success_toast } from "../components/Feedback";
import { Link } from "react-router-dom";

function Login() {
    let [eye_icon,setEye_icon]=useState(<FaEye className="h-5 w-5"/>);
    let [disable,setDisable]=useState(false); 

    function toggle_password(){
        let password=document.getElementById("password");
        if(password?.getAttribute("type")=="password"){
          password?.setAttribute("type","text");
          setEye_icon(<FaEyeSlash className="h-5 w-5"/>);
          return;
        }
        password?.setAttribute("type","password");
        setEye_icon(<FaEye className="h-5 w-5"/>);
    }

    

    async function handleLogin(e:any){
        try {
            e.preventDefault();
            setDisable(true)
            let userInput={
                email:e.target.email.value,
                password:e.target.password.value
            }
            let userCredential=await signInWithEmailAndPassword(auth, userInput.email, userInput.password);
            const user = userCredential.user;
            console.log(user)
            success_toast(`Sign in successfull`)
            setDisable(false)
        } catch (error:any) {
            setDisable(false)
            const errorCode = error.code;
            const errorMessage = error.message;
            console.log(error,errorCode,errorMessage)
            errorCode==="auth/network-request-failed"?err_toast(`No internet`):err_toast(error.message)
        }
    }
    return (
        <main className="flex md:h-screen max-md:h-[85vh] max-md:justify-center flex-col items-center md:p-4">
            <div className="flex flex-col sm:w-[440px] max-sm:w-[85vw]">
                <div className="sm:my-[40px] max-sm:my-[20px]">
                <p className="text-[30px] text-[#1e293b] mb-[8px] font-semibold">Sign in</p>
                <p className="text-[#64748b] text-[14px]">Enter your credentials to access our realtime messaging platform</p>
                </div>
                <form onSubmit={(e)=>handleLogin(e)} className="flex flex-col text-sm">
                <div className="flex flex-col mb-3">
                    <label className="mb-[8px] font-semibold text-[#0f172a]" htmlFor="email">Email</label>
                    <div className="pb-4">
                        <input id="email" name="email" type="email" className={`px-[10px] w-full py-2 focus:outline-[var(--theme-blue)] focus:outline-[1px] bg-white border-[1px] rounded-lg`} placeholder="johndoe@gmail.com" required/>
                    </div>

                    <label htmlFor="password" className=" font-semibold mb-[8px] text-[#0f172a]">Password</label>
                    <div className="flex flex-col">
                        <div className="flex">
                            <input id="password" name="password" placeholder="Enter password" type="password" className={`flex-grow px-[10px] py-2 focus:outline-[var(--theme-blue)] focus:outline-[1px] bg-white border-[1px] rounded-l-lg`} minLength={8} maxLength={24} required/>
                            <button type="button" onClick={toggle_password} className="rounded-r-lg px-3 py-2 border-[1px] w-[53px] bg-white text-[#64748b]">
                            {eye_icon}
                            </button>
                        </div>
                    </div>
                </div>

                <a href="#" target="_blank" rel="noopener noreferrer" className="text-[#475569] underline text-[14px] ml-auto">Forget Password?</a>
                <button disabled={disable} className={disable===true?"cursor-wait mt-5 capitalize py-3 px-6 text-white rounded-md bg-[var(--theme-dark)]":"mt-5 capitalize py-3 px-6 text-white rounded-md bg-[var(--theme-blue)]"}>
                    {disable===false?(<span>
                        Sign in
                    </span>):(
                        <i className="italic">Signing in...</i>
                    )}
                </button>
                <div className="flex mt-5">
                    <p className="mr-3">{"Don't have an account"}</p>
                    <Link to="/create" className="underline text-[var(--theme-blue)]">Create an account</Link>
                </div>
                <div className="mt-5 text-xs flex items-center gap-x-1 text-[var(--gray-text)]">
                    <FaInfoCircle className="w-5 h-5"/>
                    <p>Only members should sign in.</p>
                    </div>
                </form>
            </div>
        </main>
    );
};

export default Login;
