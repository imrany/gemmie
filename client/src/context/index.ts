import { createContext } from 'react'
import { User } from '../types/definitions'

export const GlobalContext=createContext<User>({
    uid:"",
    photoURL:"",
    email:"",
    displayName:"",
    phoneNumber:0,
    emailVerified:false
})