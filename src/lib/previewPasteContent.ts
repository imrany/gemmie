import { toast } from "vue-sonner";

// Function to detect content type
export function detectContentType(content: string): 'text' | 'code' | 'json' | 'markdown' | 'xml' | 'html' {
  const trimmed = content.trim()
  
  // JSON detection
  if ((trimmed.startsWith('{') && trimmed.endsWith('}')) || 
      (trimmed.startsWith('[') && trimmed.endsWith(']'))) {
    try {
      JSON.parse(trimmed)
      return 'json'
    } catch (e) {
      // Continue checking other types
    }
  }
  
  // XML/HTML detection
  if (trimmed.startsWith('<') && trimmed.endsWith('>')) {
    if (trimmed.includes('<!DOCTYPE html') || trimmed.includes('<html')) {
      return 'html'
    }
    return 'xml'
  }
  
  // Markdown detection
  if (trimmed.includes('# ') || trimmed.includes('## ') || 
      trimmed.includes('```') || trimmed.includes('- ') || 
      trimmed.includes('* ') || trimmed.match(/\[.*\]\(.*\)/)) {
    return 'markdown'
  }
  
  // Code detection (common patterns)
  if (trimmed.includes('function ') || trimmed.includes('const ') || 
      trimmed.includes('import ') || trimmed.includes('class ') ||
      trimmed.includes('def ') || trimmed.includes('public class') ||
      trimmed.includes('<?php') || trimmed.includes('#!/')) {
    return 'code'
  }
  
  return 'text'
}

// Function to get language for syntax highlighting
export function getHighlightLanguage(type: string): string {
  switch (type) {
    case 'json': return 'json'
    case 'html': return 'html'
    case 'xml': return 'xml'
    case 'markdown': return 'markdown'
    case 'code': return 'javascript' // default, could be improved with better detection
    default: return 'plaintext'
  }
}

// Function to copy content to clipboard
export function copyPasteContent(content: string) {
  navigator.clipboard.writeText(content)
    .then(() => {
      toast.success('Content copied to clipboard!', {
        duration: 3000
      })
    })
    .catch(() => {
      toast.error('Failed to copy content', {
        duration: 3000
      })
    })
}

// Function to safely escape content for HTML attributes
export function escapeForAttribute(str: string): string {
  return str
    .replace(/\\/g, '\\\\')
    .replace(/'/g, "\\'")
    .replace(/"/g, '\\"')
    .replace(/\n/g, '\\n')
    .replace(/\r/g, '\\r')
    .replace(/\t/g, '\\t')
}