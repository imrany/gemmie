import { Outlet, Link, useLocation } from "react-router-dom";
import { auth, signOut } from "../firebaseConfig/config"
import Footer from "../components/Footer";
import { useContext, useEffect, useState } from "react";
import { GlobalContext } from "../context";
import { MdClose, MdMenu } from "react-icons/md";
import { err_toast, success_toast } from "../components/Feedback";

export default function Layout(){
  const { email } =useContext(GlobalContext);
  let [showMobileSidebar, setShowMobileSidebar]=useState(false)
  const location=useLocation()
  async function logout(){
    try{
      await signOut(auth);
      success_toast(`Successfull sign out`)
    }catch(error:any){
      console.log(error)
      err_toast(error.message)
    }
  }

  let links=[
    {
      name:"About",
      to:"/"
    },
    {
      name:"Chat Room",
      to:"/chat_room"
    }
  ]

  // Close the dropdown if the user clicks outside of it
  window.onclick = function(event:any) {
    if (!event.target.matches('.dropbtn')) {
      var dropdowns = document.getElementsByClassName("dropdown-content");
      var i;
      for (i = 0; i < dropdowns.length; i++) {
        var openDropdown = dropdowns[i];
        if (openDropdown.classList.contains('show')) {
          openDropdown.classList.remove('show');
        }
      }
    }
  }

  let show_mobile_sidebar_btn=document.querySelector("#show_mobile_sidebar_btn")
  show_mobile_sidebar_btn?.addEventListener("click",()=>{
    setShowMobileSidebar(true)
  })

  useEffect(()=>{
    if(screen.width>768){
      let header:any=document.getElementById("header")
      header.innerHTML=`
        <div class="px-2 pt-1 flex items-center justify-between">
            <p class="text-sm ml-auto">Logged in as <span class="underline text-[var(--theme-blue)]">${email}</span></p>
        </div>
        <div class="px-2 py-2 flex items-center justify-between ">
            <div class="flex gap-2 items-center">
                <img src="/uni_logo.png" alt="ruiru logo" width="30" height="30"/>
                <p class="text-lg font-semibold">Realtime messaging platform</p>
            </div>
            <div class="flex gap-8 items-center">
            <div class="flex flex-col justify-center">
                <p class="text-[var(--theme-yellow)] font-semibold">Call Us:</p>
                <a href="tel:+254734720752" target="_blank" rel="noopener noreferrer">+254734720752</a>
            </div>
  
            <div class="flex flex-col justify-center">
                <p class="text-[var(--theme-yellow)] font-semibold">Email:</p>
                <a href="mailto:imranmat254@gmail.com" target="_blank" rel="noopener noreferrer">imranmat254@gmail.com</a>
            </div>
  
            <div class="flex flex-col justify-center">
                <p class="text-[var(--theme-yellow)] font-semibold">Virtual Tour:</p>
                <a href="#" target="_blank" rel="noopener noreferrer">Click to Visit</a>
            </div>
            </div>
        </div>
      `
    }
  },[location.pathname])
  return (
    <>
      <nav className="md-nav border-b-[1px] shadow-sm">
        <div className="" id={screen.width>768?"header":""}>
          <div className="px-2 pt-1 flex items-center justify-between">
            <p className="text-sm ml-auto">Logged in as <span className="underline text-[var(--theme-blue)]">{email}</span></p>
          </div>
          <div className="px-2 py-2 flex items-center justify-between ">
            <div className="flex gap-2 items-center">
              <img src="/uni_logo.png" alt="ruiru logo" width={30} height={30}/>
              <p className="text-lg font-semibold">Realtime messaging platform</p>
            </div>
            <div className="flex gap-8 items-center">
              <div className="flex flex-col justify-center">
                <p className="text-[var(--theme-yellow)] font-semibold">Call Us:</p>
                <a href="tel:+254734720752" target="_blank" rel="noopener noreferrer">+254734720752</a>
              </div>

              <div className="flex flex-col justify-center">
                <p className="text-[var(--theme-yellow)] font-semibold">Email:</p>
                <a href="mailto:imranmat2542gmail.com" target="_blank" rel="noopener noreferrer">imranmat254@gmail.com</a>
              </div>

              <div className="flex flex-col justify-center">
                <p className="text-[var(--theme-yellow)] font-semibold">Virtual Tour:</p>
                <a href="#" target="_blank" rel="noopener noreferrer">Click to Visit</a>
              </div>
            </div>
          </div>
        </div>
        <div className="flex text-white bg-[var(--theme-blue)] pr-2">
	        {links.map((link,index)=>(<Link to={link.to} className={location.pathname===link.to?"px-2 py-3 bg-white text-[#213547]":"px-2 py-3 hover:bg-slate-200 hover:text-[#213547]"} key={index}>{link.name}</Link>))}
          <button onClick={logout} className="px-2 hover:bg-slate-200 hover:text-[#213547]">Log out</button>
        </div>
      </nav>

      <nav className="max-md-nav bg-[var(--theme-blue)] text-white shadow-sm">
        <div className="" id={screen.width<768?"header":""}>
          <div className="px-2 py-2 flex items-center justify-between ">
            <div className="flex gap-2 items-center">
              <img src="/uni_logo.png" alt="ruiru logo" width={25} height={25}/>
              <p className="text-base font-semibold">Realtime messaging platform</p>
            </div>
            <button
              id="show_mobile_sidebar_btn"
              onClick={()=>setShowMobileSidebar(true)}
              className="rounded-md p-1 border-[1px]"
            >
              <MdMenu className="w-5 h-5"/>
            </button>
          </div>
        </div>
      </nav>
      {showMobileSidebar===true?(
        <div id="mobile-sidebar" className="max-md-nav py-6 bg-white z-10 fixed top-0 bottom-0 left-0 right-0 h-[100vh]">
          <div className="flex flex-col">
            <div className="flex items-center px-5">
              <button onClick={()=>setShowMobileSidebar(false)} className="ml-auto">
                <MdClose className="w-7 h-7"/>
              </button>
            </div>

            <div className="flex flex-col overflow-y-auto">
              {links.map((link,index)=>(<Link to={link.to} className={location.pathname===link.to?"px-9 py-4 bg-white text-[#213547]":"px-9 py-4 hover:bg-slate-200 hover:text-[#213547]"} onClick={()=>setShowMobileSidebar(false)} key={index}>{link.name}</Link>))}
              <button onClick={logout} className="px-9 py-4 text-left text-[#213547]">Log out</button>
            </div>
          </div>
        </div>
      ):""}
      <Outlet />
      <Footer/>
    </>
  )
};
