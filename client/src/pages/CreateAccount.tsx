import { useState } from "react"
import { FaInfoCircle } from "react-icons/fa";
import { auth, createUserWithEmailAndPassword } from "../firebaseConfig/config";
import { err_toast, success_toast } from "../components/Feedback";
import { Link } from "react-router-dom";

export default function CreateAccount() {
    let [disable,setDisable]=useState(false); 
    let [passwordErr,setPasswordErr]=useState(<></>); 

    async function handleSignUp(e:any){
        try {
            e.preventDefault()
            let userInput={
                email:e.target.email.value,
                password:e.target.confirmpassword.value
            }
            if(e.target.confirmpassword.value===e.target.password.value){
                setDisable(true)
                let userCredential=await createUserWithEmailAndPassword(auth, userInput.email, userInput.password);
                const user = userCredential.user;
                console.log(user)
                success_toast(`Sign in successfull`)
                setDisable(false)
            }else{
                setPasswordErr(
                    <p className="text-red-500 text-[11px]">Passwords doesn't match</p>
                )
            }
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
                <p className="text-[30px] text-[#1e293b] mb-[8px] font-semibold">Create Account</p>
                <p className="text-[#64748b] text-[14px]">Enter your credentials to access our realtime messaging platform.</p>
                </div>
                <form onSubmit={(e)=>handleSignUp(e)} className="flex flex-col text-sm">
                <div className="flex flex-col mb-3">
                    {passwordErr}
                    <label className="mb-[8px] font-semibold text-[#0f172a]" htmlFor="email">Email</label>
                    <div className="pb-4">
                        <input id="email" name="email" type="email" className={`px-[10px] w-full py-2 focus:outline-[var(--theme-blue)] focus:outline-[1px] bg-white border-[1px] rounded-lg`} placeholder="Enter email" required/>
                    </div>

                    <label htmlFor="password" className=" font-semibold mb-[8px] text-[#0f172a]">Password</label>
                    <div className="flex flex-col">
                        <div className="flex">
                            <input id="password" name="password" placeholder="Enter password" type="password" className={`flex-grow px-[10px] py-2 focus:outline-[var(--theme-blue)] focus:outline-[1px] bg-white border-[1px] rounded-l-lg`} minLength={8} maxLength={24} required/>
                        </div>
                    </div>
                    <label htmlFor="confirmpassword" className=" font-semibold mb-[8px] text-[#0f172a]">Confirm Password</label>
                    <div className="flex flex-col">
                        <div className="flex">
                            <input id="confirmpassword" placeholder="Confirm password" name="confirmpassword" type="password" className={`flex-grow px-[10px] py-2 focus:outline-[var(--theme-blue)] focus:outline-[1px] bg-white border-[1px] rounded-l-lg`} minLength={8} maxLength={24} required/>
                        </div>
                    </div>
                </div>

                <button disabled={disable} className={disable===true?"cursor-wait mt-5 capitalize py-3 px-6 text-white rounded-md bg-[var(--theme-dark)]":"mt-5 capitalize py-3 px-6 text-white rounded-md bg-[var(--theme-blue)]"}>
                    {disable===false?(<span>
                        Create
                    </span>):(
                        <i className="italic">Creating account...</i>
                    )}
                </button>
                <div className="flex mt-5">
                    <p className="mr-3">{"Do you have an account?"}</p>
                    <Link to="/login" className="underline text-[var(--theme-blue)]">Login instead</Link>
                </div>
                <div className="mt-5 text-xs flex items-center gap-x-1 text-[var(--gray-text)]">
                    <FaInfoCircle className="w-5 h-5"/>
                    <p>By creating an account you've agreed with our privacy policies.</p>
                    </div>
                </form>
            </div>
        </main>
    );
};
