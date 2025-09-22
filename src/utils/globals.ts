import { toast } from "vue-sonner"

export const API_BASE_URL = 'http://localhost:8081/api'
// export const API_BASE_URL = 'https://gemmie.villebiz.com/api'
export const WRAPPER_URL = 'https://wrapper.villebiz.com/v1/genai'

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