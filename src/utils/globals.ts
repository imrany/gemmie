import { ref } from "vue"
import { toast } from "vue-sonner"

function getBaseURL() {
  if (import.meta.env.DEV) {
    return 'http://localhost:8081'
  } else {
    // return window.location.origin
    return 'https://gemmie.villebiz.com'
  }
}

export function isPromptTooShort(prompt: string): boolean {
  return prompt.trim().split(/\s+/).length < 30
}

// Generate unique chat ID
export function generateChatId(): string {
  return 'chat_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9)
}

// Generate chat title from first message
export function generateChatTitle(firstMessage: string): string {
  const title = firstMessage.slice(0, 50).trim()
  return title.length < firstMessage.length ? title + '...' : title
}

// Extract URLs from text using regex (removed extra pipe character)
export function extractUrls(text: string): string[] {
  const urlRegex = /https?:\/\/[^\s<>"{}|\\^`[\]]+/gi
  return text.match(urlRegex) || []
}

// Helper function to show confirmation dialog
export function copyCode(text: string, button?: HTMLElement) {
  navigator.clipboard.writeText(text)
    .then(() => {
      if (button) {
        button.innerText = 'Copied!'
        setTimeout(() => (button.innerText = 'Copy code'), 2000)
      }
    })
    .catch(err => {
      console.error('Failed to copy text: ', err)
      toast.error('Failed to copy code to clipboard', {
        duration: 3000,
        description: ''
      })
    })
}

export function validateCredentials(username: string, email: string, password: string): string | null {
  // Username: 3–12 chars, no spaces, only letters, numbers, underscores, hyphens
  const usernameRegex = /^[a-zA-Z0-9_-]{3,12}$/
  if (!usernameRegex.test(username)) {
    return "Username must be 3–12 characters, no spaces, only letters, numbers, _ or -"
  }

  // Email: basic check
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  if (!emailRegex.test(email)) {
    return "Invalid email format"
  }

  // Password: at least 8 chars
  if (password.length < 8) {
    return "Password must be at least 8 characters"
  }

  return null
}

export async function getTransaction(external_reference: string){
  try{
    const parseRes=await fetch(`${API_BASE_URL}/transactions/${external_reference}`,{
      method:"GET"
    })
    const res=await parseRes.json()
    if(res.success){
      return res.data
    }else{
      toast.error(res.message,{ 
        duration:5000,
        action:{
          label:"Try again",
          onClick:()=>toast.dismiss()
        }
      })
    }
  }catch(err:any){
    toast.error(`Transaction failed, Try again`,{
      duration:5000,
        action:{
          label:"Try again",
          onClick:()=>toast.dismiss()
        }
    })
  }
}

export const plans = ref([
  {
    name: "Student",
    id: "student",
    // price: 50,
    price: 1,
    duration: "per 5 hours",
    description: "Perfect for quick sessions and light academic use.",
    features: [
      "Strong privacy",
      "Good for light work",
      "Analyze and summarize text",
      "Write, edit and create content",
      "Get web results and insights",
    ],
    popular: false,
  },
  {
    name: "Pro",
    id: "pro",
    price: 100,
    duration: "per 24 hours",
    description: "Great for professionals needing reliable AI help all day.",
    features: [
      "Everything in Student Plan",
      "Handles heavy workloads",
      "Downloadable content e.g PDFs",
      "Full privacy",
      "Data sync across all devices",
    ],
    popular: true,
  },
  {
    name: "Hobbyist",
    id: "hobbyist",
    price: 500,
    duration: "per week",
    description: "Best for hobbyists and regular users exploring AI deeply.",
    features: [
      "Everything in Pro Plan",
      "More usage time",
      "Persistent sync",
      "Extended access for projects",
    ],
    popular: false,
  },
])

export let API_BASE_URL = getBaseURL() + '/api'
export let SOCKET_URL = getBaseURL().replace(/^http/, 'ws') + '/ws'
export const WRAPPER_URL = 'https://wrapper.villebiz.com/v1/genai'