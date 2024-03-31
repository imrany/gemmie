import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { useState, useEffect } from "react";
import { auth, onAuthStateChanged } from "./firebaseConfig/config";
import { ToastContainer } from 'react-toastify';
import { Toaster } from "react-hot-toast";
import Layout from './pages/Layout';
import NotFound from './pages/NotFound';
import About from './pages/About';
import Login from "./pages/Login";
import ChatRoom from "./pages/ChatRoom";
import { GlobalContext } from "./context";
import { User } from "./types/definitions";
import CreateAccount from "./pages/CreateAccount";

function App() {
  const [user,setUser]=useState<User>({
    uid:"",
    photoURL:"",
    email:"",
    displayName:"",
    phoneNumber:0,
    emailVerified:false
  })
  const [isLoading,setIsLoading]=useState(true)
  const [isAuth,setIsAuth]=useState(false);
  useEffect(()=>{
    onAuthStateChanged(auth, (user) => {
      if (user) {
        // User is signed in
        console.log(user)
        let userData:User={
          uid:user.uid,
          photoURL:user.photoURL,
          email:user.email,
          displayName:user.displayName,
          phoneNumber:user.phoneNumber,
          emailVerified:user.emailVerified
        }
        setUser(userData)
        setIsAuth(true)
        setIsLoading(false)
      } else {
        // User is signed out
        console.log("user is signed out")
	      setIsAuth(false)
        setIsLoading(false)
      }
    });
  },[isAuth]);

  return (
    <BrowserRouter>
      <GlobalContext.Provider value={user}>
        {isLoading?(
          <div className="fixed top-0 bottom-0 left-0 z-20 right-0 bg-white">
            <div className="flex flex-col items-center h-[100vh] justify-center">
              <p className="text-xl font-semibold text-[var(--theme-blue)]">Loading...</p>
            </div>
          </div>
        ):(
          <>
            <ToastContainer 
              autoClose={5000}
              hideProgressBar={false}
              newestOnTop={false}
              closeOnClick
              rtl={false}
              pauseOnFocusLoss
              draggable
              pauseOnHover
              theme="light"
            />
            <Toaster/>
            <Routes>
              <Route path="/login" element={!isAuth?<Login/>:<Navigate to="/"/>}/>
              <Route path="/create" element={!isAuth?<CreateAccount />:<Navigate to="/"/>} />
              <Route path="/" element={isAuth?<Layout />:<Navigate to="/login"/>}>
                <Route index element={<About />} />
                <Route path="chat_room" element={<ChatRoom />} />
              </Route>
              <Route path="*" element={<NotFound />} />
            </Routes>
          </>
        )}
      </GlobalContext.Provider>
    </BrowserRouter>
  )
}

export default App
