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

// Enhanced validateCredentials function
export function validateCredentials(username: string, email: string, password: string, agreeToTerms: boolean): string | null {
  if (!username || username.trim().length < 2) {
    return 'Username must be at least 2 characters long'
  }
  
  if (username.trim().length > 50) {
    return 'Username must be less than 50 characters'
  }
  
  if (!email || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
    return 'Please enter a valid email address'
  }
  
  if (!password || password.length < 8) {
    return 'Password must be at least 8 characters long'
  }
  
  if (password.length > 25) {
    return 'Password must be less than 24 characters'
  }

  if(!agreeToTerms){
    return 'Must accept our terms of service and privacy policy'
  }
  
  return null
}

export async function getTransaction(external_reference: string){
  try{
    const parseRes=await fetch(`${API_BASE_URL}/transactions/${external_reference}`,{
      method:"GET"
    })
    const res=await parseRes.json()
    return res
  }catch(err:any){
    console.error("Error fetching transaction:", err)
    return {success:false,message:"Error fetching transaction"}
  }
}

export const plans = ref([
  {
    name: "Student",
    id: "student",
    price: 50,
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

export const API_BASE_URL = getBaseURL() + '/api'
export const SOCKET_URL = getBaseURL().replace(/^http/, 'ws') + '/ws'
export const WRAPPER_URL = 'https://wrapper.villebiz.com/v1/genai'
export const SPINDLE_URL = 'https://spindle.villebiz.com'

// connection status checking
export function checkConnectionStatus(): Promise<boolean> {
  return new Promise((resolve) => {
    // Try to fetch a simple endpoint or ping
    fetch(`${API_BASE_URL}/health`, { 
      method: 'GET', 
      signal: AbortSignal.timeout(5000) 
    })
    .then(response => resolve(response.ok))
    .catch(() => resolve(false))
  })
}

// paste detection function
export function detectLargePaste(text: string): boolean {
  const wordCount = text.trim().split(/\s+/).length
  const charCount = text.length
  return wordCount > 100 || charCount > 800
}