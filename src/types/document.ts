export type UploadedFile = {
  id: string
  name: string
  url: string
  type: string
  size: number
  previewUrl?: string
  pages?: number
  uploadedAt: Date
  isCustom?: boolean
  content?: string
}

export type EditableContent = {
  pageNum: number
  content: string
  originalContent: string
  isModified: boolean
  annotations: Annotation[]
}

export type Annotation = {
  id: string
  type: 'highlight' | 'note' | 'bookmark'
  text: string
  startIndex: number
  endIndex: number
  color: string
  note?: string
  timestamp: Date
}

export type SearchResult = {
  pageNum: number
  text: string
  index: number
}

export type AIAction = 'summarize' | 'expand' | 'simplify' | 'translate' | 'paraphrase' | 'improve' | 'explain'

export type HistoryEntry = {
  content: string
  timestamp: number
  action?: string
}