<script setup lang="ts">
import { ref, onBeforeUnmount, nextTick, computed, onMounted, onUnmounted, watch } from "vue"
import * as pdfjsLib from "pdfjs-dist"
import pdfjsWorker from "pdfjs-dist/build/pdf.worker?url"
import type { Ref } from "vue"
import { WRAPPER_URL } from "@/utils/globals"
import { inject } from "vue"
import { useRouter } from "vue-router"
import { renderMarkdown } from "@/utils/markdownSupport"

// configure worker
pdfjsLib.GlobalWorkerOptions.workerSrc = pdfjsWorker

type UploadedFile = {
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

type EditableContent = {
  pageNum: number
  content: string
  originalContent: string
  isModified: boolean
  annotations: Annotation[]
}

type Annotation = {
  id: string
  type: 'highlight' | 'note' | 'bookmark'
  text: string
  startIndex: number
  endIndex: number
  color: string
  note?: string
  timestamp: Date
}

type SearchResult = {
  pageNum: number
  text: string
  index: number
}

type AIAction = 'summarize' | 'expand' | 'simplify' | 'translate' | 'paraphrase' | 'improve' | 'explain'

type HistoryEntry = {
  content: string
  timestamp: number
  action?: string
}

const globalState = inject('globalState') as {
  screenWidth: number
}
const {
  screenWidth
}=globalState;

const router=useRouter()
const uploadedFiles = ref<UploadedFile[]>([])
const selectedPdfUrl = ref("")
const selectedPdfName = ref("")
const selectedFileId = ref("")
const isDragOver = ref(false)
const showExportDropdown = ref(false)

// PDF Editor state
const currentPage = ref(1)
const totalPages = ref(0)
const editablePages = ref<EditableContent[]>([])
const isLoading = ref(false)
const loadError = ref<string>('')
const fontSize = ref(14)
const lineHeight = ref(1.5)

// Sidebar state
const sidebarOpen = ref(true)
const activeSidebarTab = ref<'outline' | 'search' | 'annotations' | 'history' | 'documents' | any>('documents')

// Text selection state
const selectedText = ref('')
const showContextMenu = ref(false)

const innerWidth=window.innerWidth
const innerHeight = window.innerHeight

const sidebarWidth = ref(290) // default width
const isResizingSidebar = ref(false)
const maxSidebarWidth = 350 // max width constraint

// Search functionality
const searchQuery = ref('')
const searchResults = ref<SearchResult[]>([])
const currentSearchIndex = ref(0)

const showAIPopover = ref(false)
const aiPopoverPosition = ref({ x: 0, y: 0 })
const aiPopoverContent = ref('')
const isLoadingAIPopover = ref(false)

const isDraggingAISuggestions = ref(false)
const isDraggingAIPopover = ref(false)
const dragOffset = ref({ x: 0, y: 0 })

// History
const editHistory = ref<Array<{ action: string, pageNum: number, timestamp: Date, preview: string }>>([])

// Undo/Redo functionality
const undoHistory = ref<HistoryEntry[]>([])
const redoHistory = ref<HistoryEntry[]>([])
const maxHistorySize = 100

// Preview state
const showPreview = ref(false)
const previewContent = ref('')

// Create New Document Modal
const documentTemplates = ref([
  {
    id: 'blank',
    name: 'Blank Document',
    icon: '📄',
    content: '# New Document\n\nStart writing here...'
  },
  {
    id: 'meeting-notes',
    name: 'Meeting Notes',
    icon: '📝',
    content: '# Meeting Notes\n\n**Date:** \n**Attendees:** \n\n## Agenda\n- \n\n## Discussion\n\n## Action Items\n- [ ] \n\n## Next Steps\n'
  },
  {
    id: 'project-plan',
    name: 'Project Plan',
    icon: '📋',
    content: '# Project Plan\n\n## Overview\n\n## Objectives\n- \n\n## Timeline\n\n## Resources\n\n## Milestones\n- [ ] \n\n## Risks\n'
  },
  {
    id: 'blog-post',
    name: 'Blog Post',
    icon: '✍️',
    content: '# Blog Post Title\n\n## Introduction\n\n## Main Content\n\n### Section 1\n\n### Section 2\n\n## Conclusion\n\n---\n*Published on [Date]*'
  },
  {
    id: 'research-notes',
    name: 'Research Notes',
    icon: '🔬',
    content: '# Research Notes\n\n**Topic:** \n**Date:** \n**Sources:** \n\n## Key Findings\n\n## Quotes\n> \n\n## Questions\n- \n\n## Next Steps\n'
  }
])

function createDocumentFromTemplate(template: any) {
  const newDoc = {
    id: `doc-${Date.now()}-${Math.random().toString(36).slice(2)}`,
    name: `${template.name}.md`,
    url: 'custom',
    type: 'text/markdown',
    size: new Blob([template.content]).size,
    uploadedAt: new Date(),
    isCustom: true,
    content: template.content
  }

  uploadedFiles.value.push(newDoc)
  saveToLocalStorage()
  openTextEditor(newDoc)
}

// AI Features state
const showAIModal = ref(false)
const aiContent = ref('')
const originalTextForAI = ref('')
const isGeneratingAI = ref(false)
const currentAIAction = ref<AIAction>('summarize')
const showAISuggestions = ref(false)
const aiSuggestionsPosition = ref({ x: 0, y: 0 })

// Auto-save state
const hasUnsavedChanges = ref(false)
const autoSaveTimer = ref<any | null>(null)

// Editor state
const editorMode = ref<'edit' | 'preview' | 'split'>('edit')
const showMarkdownToolbar = ref(true)

// AI Suggestions
const aiSuggestions = ref([
  {
    icon: '📝',
    label: 'Summarize',
    action: 'summarize',
    description: 'Create a concise summary',
    shortcut: 'Ctrl+Shift+S'
  },
  {
    icon: '🔍',
    label: 'Expand',
    action: 'expand',
    description: 'Elaborate on the selected text',
    shortcut: 'Ctrl+Shift+E'
  },
  {
    icon: '✨',
    label: 'Improve',
    action: 'improve',
    description: 'Enhance writing quality',
    shortcut: 'Ctrl+Shift+I'
  },
  {
    icon: '📖',
    label: 'Simplify',
    action: 'simplify',
    description: 'Make text easier to understand',
    shortcut: 'Ctrl+Shift+P'
  },
  {
    icon: '🔄',
    label: 'Paraphrase',
    action: 'paraphrase',
    description: 'Rewrite in different words',
    shortcut: 'Ctrl+Shift+R'
  },
  {
    icon: '🌐',
    label: 'Translate',
    action: 'translate',
    description: 'Translate to different language',
    shortcut: 'Ctrl+Shift+T'
  }
])

// Computed properties
const hasSearchResults = computed(() => searchResults.value.length > 0)
const allAnnotations = computed(() =>
  editablePages.value.flatMap(page =>
    page.annotations.map(ann => ({ ...ann, pageNum: page.pageNum }))
  )
)

const canUndo = computed(() => undoHistory.value.length > 0)
const canRedo = computed(() => redoHistory.value.length > 0)

const renderedMarkdown = computed(() => {
  const currentPageData = getCurrentPageContent()
  if (!currentPageData) return ''
  return renderMarkdown(currentPageData.content)
})

// Local storage keys
const STORAGE_KEY = 'pdf_editor_documents'
const CURRENT_DOCUMENT_KEY = 'pdf_editor_current_document'

// Watch for changes to trigger auto-save
watch([editablePages], () => {
  hasUnsavedChanges.value = editablePages.value.some(page => page.isModified)
  if (hasUnsavedChanges.value) {
    scheduleAutoSave()
  }
}, { deep: true })

/**
 * Undo/Redo functionality
 */
function saveToHistory() {
  const currentPageData = getCurrentPageContent()
  if (!currentPageData) return

  const historyEntry: HistoryEntry = {
    content: currentPageData.content,
    timestamp: Date.now()
  }

  undoHistory.value.push(historyEntry)
  redoHistory.value = [] // Clear redo history when new change is made

  // Limit history size
  if (undoHistory.value.length > maxHistorySize) {
    undoHistory.value.shift()
  }
}

function undo() {
  if (!canUndo.value) return

  const currentPageData = getCurrentPageContent()
  if (!currentPageData) return

  // Save current state to redo history
  redoHistory.value.push({
    content: currentPageData.content,
    timestamp: Date.now()
  })

  // Restore previous state
  const previousState = undoHistory.value.pop()!
  updatePageContent(previousState.content, false) // Don't save to history

  addToHistory('Undo', 'Undid last change')
}

function redo() {
  if (!canRedo.value) return

  const currentPageData = getCurrentPageContent()
  if (!currentPageData) return

  // Save current state to undo history
  undoHistory.value.push({
    content: currentPageData.content,
    timestamp: Date.now()
  })

  // Restore next state
  const nextState = redoHistory.value.pop()!
  updatePageContent(nextState.content, false) // Don't save to history

  addToHistory('Redo', 'Redid last change')
}

/**
 * Markdown Toolbar Actions
 */
function insertMarkdown(before: string, after: string = '', placeholder: string = '') {
  const textarea = document.querySelector('textarea') as HTMLTextAreaElement
  if (!textarea) return

  const start = textarea.selectionStart
  const end = textarea.selectionEnd
  const selectedText = textarea.value.substring(start, end)
  const replacement = selectedText || placeholder

  const newText = before + replacement + after
  const newContent = textarea.value.substring(0, start) + newText + textarea.value.substring(end)

  saveToHistory()
  updatePageContent(newContent)

  // Restore cursor position
  nextTick(() => {
    textarea.focus()
    if (selectedText) {
      textarea.setSelectionRange(start + before.length, start + before.length + replacement.length)
    } else {
      textarea.setSelectionRange(start + before.length, start + before.length + placeholder.length)
    }
  })
}

function insertBold() {
  insertMarkdown('**', '**', 'bold text')
}

function insertItalic() {
  insertMarkdown('*', '*', 'italic text')
}

function insertCode() {
  insertMarkdown('`', '`', 'code')
}

function insertLink() {
  insertMarkdown('[', '](url)', 'link text')
}

function insertHeader(level: number) {
  const prefix = '#'.repeat(level) + ' '
  insertMarkdown(prefix, '', `Header ${level}`)
}

function insertList() {
  insertMarkdown('- ', '', 'List item')
}

function insertNumberedList() {
  insertMarkdown('1. ', '', 'List item')
}

function insertQuote() {
  insertMarkdown('> ', '', 'Quote text')
}

function insertTable() {
  const tableTemplate = `| Column 1 | Column 2 | Column 3 |
|----------|----------|----------|
| Row 1    | Data     | Data     |
| Row 2    | Data     | Data     |`

  const textarea = document.querySelector('textarea') as HTMLTextAreaElement
  if (!textarea) return

  const start = textarea.selectionStart
  saveToHistory()
  const newContent = textarea.value.substring(0, start) + tableTemplate + textarea.value.substring(start)
  updatePageContent(newContent)
}

/**
 * Keyboard shortcuts
 */
function handleEditorKeydown(event: KeyboardEvent) {
  // Handle undo/redo
  if (event.ctrlKey || event.metaKey) {
    switch (event.key) {
      case 'z':
        if (event.shiftKey) {
          event.preventDefault()
          redo()
        } else {
          event.preventDefault()
          undo()
        }
        break
      case 'y':
        event.preventDefault()
        redo()
        break
      case 'b':
        event.preventDefault()
        insertBold()
        break
      case 'i':
        event.preventDefault()
        insertItalic()
        break
      case 'k':
        event.preventDefault()
        insertLink()
        break
      case 's':
        event.preventDefault()
        saveDocumentChanges()
        break
    }
  }

  // AI Suggestions trigger (Ctrl + /)
  if (event.key === '/' && event.ctrlKey) {
    event.preventDefault()
    showAIToolbar()
  }

  // AI Shortcuts
  if (event.ctrlKey && event.shiftKey) {
    switch (event.key) {
      case 'S':
        event.preventDefault()
        handleAIShortcut('summarize')
        break
      case 'E':
        event.preventDefault()
        handleAIShortcut('expand')
        break
      case 'I':
        event.preventDefault()
        handleAIShortcut('improve')
        break
      case 'P':
        event.preventDefault()
        handleAIShortcut('simplify')
        break
      case 'R':
        event.preventDefault()
        handleAIShortcut('paraphrase')
        break
      case 'T':
        event.preventDefault()
        handleAIShortcut('translate')
        break
    }
  }
}

/**
 * AI Integration
 */
async function handlePrompt(prompt: string) {
  try {
    const response = await fetch(WRAPPER_URL, {
      method: "POST",
      body: JSON.stringify(prompt),
      headers: { "content-type": "application/json" }
    })

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`)
    }

    const parseRes = await response.json()
    return parseRes
  } catch (err: any) {
    console.error('AI request failed:', err)
    throw err
  }
}

function getAIPrompt(action: AIAction, text: string): string {
  const prompts = {
    summarize: `Please provide a concise summary of the following text. Focus on the main points and key ideas:\n\n"${text}"`,
    expand: `Please expand on the following text by adding more details, examples, and explanations:\n\n"${text}"`,
    simplify: `Please simplify the following text to make it easier to understand while keeping the core meaning:\n\n"${text}"`,
    translate: `Please translate the following text to English (if not already) and provide the translation:\n\n"${text}"`,
    paraphrase: `Please paraphrase the following text using different words and sentence structures while preserving the original meaning:\n\n"${text}"`,
    improve: `Please improve the following text by enhancing clarity, grammar, and overall writing quality:\n\n"${text}"`,
    explain: `Please explain the following text in simple terms, breaking down complex concepts:\n\n"${text}"`
  }
  return prompts[action]
}

function saveAIContent() {
  const currentPageData = getCurrentPageContent()
  if (currentPageData && aiContent.value.trim()) {
    saveToHistory()
    const newContent = currentPageData.content.replace(originalTextForAI.value, aiContent.value.trim())
    updatePageContent(newContent)
    addToHistory(`AI ${currentAIAction.value}`, `Applied ${currentAIAction.value} to selected text`)
  }
  closeAIModal()
}

function discardAIContent() {
  closeAIModal()
}

function closeAIModal() {
  showAIModal.value = false
  aiContent.value = ''
  originalTextForAI.value = ''
  isGeneratingAI.value = false
}

// Simplified AI Suggestions drag handling
function startSuggestionsDrag(event: MouseEvent | any) {
  event.preventDefault()
  event.stopPropagation()

  isDraggingAISuggestions.value = true
  
  const rect = event?.currentTarget?.closest('.ai-suggestions').getBoundingClientRect()
  dragOffset.value = {
    x: event.clientX - rect.left,
    y: event.clientY - rect.top
  }

  document.addEventListener('mousemove', handleSuggestionsDrag)
  document.addEventListener('mouseup', stopSuggestionsDrag)
}

function handleSuggestionsDrag(event: MouseEvent) {
  if (!isDraggingAISuggestions.value) return
  
  event.preventDefault()

  let newX = event.clientX - dragOffset.value.x
  let newY = event.clientY - dragOffset.value.y

  // Viewport constraints
  const viewport = {
    width: window.innerWidth,
    height: window.innerHeight
  }
  
  // Constrain to viewport
  newX = Math.max(10, Math.min(newX, viewport.width - 290))
  newY = Math.max(10, Math.min(newY, viewport.height - 250))

  aiSuggestionsPosition.value = { x: newX, y: newY }
}

function stopSuggestionsDrag() {
  isDraggingAISuggestions.value = false
  document.removeEventListener('mousemove', handleSuggestionsDrag)
  document.removeEventListener('mouseup', stopSuggestionsDrag)
}

// Simplified version - remove the complex startDrag function and replace with:
function startDrag(event:MouseEvent, component:string) {
  // Only handle popover now, since suggestions has its own handler
  if (component !== 'popover') return
  
  const target:any = event.target
  const isDragHandle = target.classList.contains('drag-handle') || target.closest('.drag-handle')

  if (!isDragHandle) return

  event.preventDefault()
  event.stopPropagation()

  isDraggingAIPopover.value = true
  
  dragOffset.value = {
    x: event.clientX - aiPopoverPosition.value.x,
    y: event.clientY - aiPopoverPosition.value.y
  }

  const handleMouseMove = (e:MouseEvent) => {
    if (!isDraggingAIPopover.value) return
    
    e.preventDefault()
    let newX = e.clientX - dragOffset.value.x
    let newY = e.clientY - dragOffset.value.y

    const viewport = {
      width: window.innerWidth,
      height: window.innerHeight
    }
    
    newX = Math.max(10, Math.min(newX, viewport.width - 320 - 10))
    newY = Math.max(10, Math.min(newY, viewport.height - 200 - 10))

    aiPopoverPosition.value = { x: newX, y: newY }
  }
  
  const handleMouseUp = () => {
    isDraggingAIPopover.value = false
    document.removeEventListener('mousemove', handleMouseMove)
    document.removeEventListener('mouseup', handleMouseUp)
  }
  
  document.addEventListener('mousemove', handleMouseMove)
  document.addEventListener('mouseup', handleMouseUp)
}

// Sidebar resize functionality
function startSidebarResize() {
  isResizingSidebar.value = true
  document.addEventListener('mousemove', handleSidebarResize)
  document.addEventListener('mouseup', stopSidebarResize)
}

function handleSidebarResize(event: MouseEvent) {
  if (!isResizingSidebar.value) return

  const newWidth = event.clientX
  // Constrain between min and max widths
  sidebarWidth.value = Math.max(290, Math.min(newWidth, maxSidebarWidth))
}

function stopSidebarResize() {
  isResizingSidebar.value = false
}

/**
 * Textarea AI Support
 */
function handleTextareaKeydown(event: KeyboardEvent) {
  handleEditorKeydown(event)
}

function handleAIShortcut(action: AIAction) {
  const selection = window.getSelection()
  const selectedTextValue = selection?.toString().trim()

  if (selectedTextValue) {
    selectedText.value = selectedTextValue
    performAIAction(action, selectedTextValue)
  } else {
    // If no text selected, use the entire content
    const currentPageData = getCurrentPageContent()
    if (currentPageData) {
      performAIAction(action, currentPageData.content)
    }
  }
}

function showAIToolbar() {
  const textarea = document.querySelector('textarea')
  if (!textarea) return

  const rect = textarea.getBoundingClientRect()

  // Position in center of visible textarea area
  aiSuggestionsPosition.value = {
    x: rect.left + (rect.width / 2) - 140, // Half of component width (280px)
    y: rect.top + (rect.height / 2) - 100  // Rough center
  }

  // Ensure it stays within viewport
  const viewport = {
    width: window.innerWidth,
    height: window.innerHeight
  }

  // Adjust horizontal position
  if (aiSuggestionsPosition.value.x < 10) {
    aiSuggestionsPosition.value.x = 10
  } else if (aiSuggestionsPosition.value.x + 280 > viewport.width - 10) {
    aiSuggestionsPosition.value.x = viewport.width - 290
  }

  // Adjust vertical position
  if (aiSuggestionsPosition.value.y < 10) {
    aiSuggestionsPosition.value.y = 10
  } else if (aiSuggestionsPosition.value.y + 200 > viewport.height - 10) {
    aiSuggestionsPosition.value.y = viewport.height - 210
  }

  showAISuggestions.value = true
}

function hideAISuggestions() {
  showAISuggestions.value = false
}

function selectAISuggestion(suggestion: typeof aiSuggestions.value[0]) {
  hideAISuggestions()
  handleAIShortcut(suggestion.action as AIAction)
}

/**
 * Auto-save functionality
 */
function scheduleAutoSave() {
  if (autoSaveTimer.value) {
    clearTimeout(autoSaveTimer.value)
  }

  autoSaveTimer.value = setTimeout(() => {
    saveDocumentChanges()
  }, 2000)
}

function saveDocumentChanges() {
  if (!selectedFileId.value || !hasUnsavedChanges.value) return

  const fileIndex = uploadedFiles.value.findIndex(f => f.id === selectedFileId.value)
  if (fileIndex !== -1 && uploadedFiles.value[fileIndex].isCustom) {
    const allContent = editablePages.value.map(page => page.content).join('\n\n')
    uploadedFiles.value[fileIndex].content = allContent
    uploadedFiles.value[fileIndex].size = new Blob([allContent]).size
  }

  saveToLocalStorage()
  hasUnsavedChanges.value = false

  editablePages.value.forEach(page => {
    page.originalContent = page.content
    page.isModified = false
  })
}

/**
 * Local Storage Management
 */
function saveToLocalStorage() {
  try {
    const documentsData = uploadedFiles.value.map(file => ({
      ...file,
      url: file.url.startsWith('blob:') && !file.isCustom ? '' : file.url,
      content: file.isCustom ? file.content : undefined
    }))
    localStorage.setItem(STORAGE_KEY, JSON.stringify(documentsData))
  } catch (error) {
    console.warn('Failed to save to localStorage:', error)
  }
}

function loadFromLocalStorage() {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored) {
      const documentsData = JSON.parse(stored)
      uploadedFiles.value = documentsData.map((doc: any) => ({
        ...doc,
        uploadedAt: new Date(doc.uploadedAt)
      }))
    }

    const currentDocId = localStorage.getItem(CURRENT_DOCUMENT_KEY)
    if (currentDocId) {
      const lastOpenedDoc = uploadedFiles.value.find(file => file.id === currentDocId)
      if (lastOpenedDoc) {
        setTimeout(() => openPdfEditor(lastOpenedDoc), 100)
      }
    }
  } catch (error) {
    console.warn('Failed to load from localStorage:', error)
  }
}

function saveCurrentDocument() {
  if (selectedFileId.value) {
    localStorage.setItem(CURRENT_DOCUMENT_KEY, selectedFileId.value)
  }
}

function openTextEditor(doc: UploadedFile) {
  selectedPdfUrl.value = doc.url
  selectedPdfName.value = doc.name
  selectedFileId.value = doc.id
  isLoading.value = false
  loadError.value = ''

  saveCurrentDocument()

  const pageContent: EditableContent = {
    pageNum: 1,
    content: doc.content || `# Welcome to ${doc.name}\n\nStart writing your content here...\n\n## Features\n\n- **Markdown support** with live preview\n- *Italic* and **bold** text\n- \`Code snippets\`\n- > Blockquotes\n- Lists and more!\n\nPress **Ctrl+/** for AI assistance.`,
    originalContent: doc.content || '',
    isModified: false,
    annotations: []
  }

  editablePages.value = [pageContent]
  currentPage.value = 1
  totalPages.value = 1
  editHistory.value = []
  searchResults.value = []
  searchQuery.value = ''
  hasUnsavedChanges.value = false
  undoHistory.value = []
  redoHistory.value = []
}


function handleTextSelection(event: Event) {
  if (event.type === 'keyup') return

  setTimeout(() => {
    const selection: any = window.getSelection()
    const selectedTextValue = selection?.toString().trim()

    if (selectedTextValue && selectedTextValue.length > 3) {
      selectedText.value = selectedTextValue
      const range = selection.getRangeAt(0)
      const rect = range.getBoundingClientRect()
      const scrollX = window.pageXOffset || document.documentElement.scrollLeft
      const scrollY = window.pageYOffset || document.documentElement.scrollTop

      // Viewport constraints
      const viewport = {
        width: window.innerWidth,
        height: window.innerHeight
      }

      // Component dimensions
      const popupWidth = 280
      const popupHeight = 200

      // Start with position below and centered on selection
      let x = rect.left + (rect.width / 2) - (popupWidth / 2)
      let y = rect.bottom + 10

      // Adjust horizontal positioning
      if (x + popupWidth > viewport.width - 10) {
        // Too far right - align with right edge of viewport
        x = viewport.width - popupWidth - 10
      }
      if (x < 10) {
        // Too far left - align with left edge of viewport
        x = 10
      }

      // If horizontal positioning is extreme, center in viewport
      if (x < viewport.width * 0.1 || x > viewport.width * 0.7) {
        x = (viewport.width / 2) - (popupWidth / 2)
      }

      // Adjust vertical positioning
      if (y + popupHeight > viewport.height - 10) {
        // Too far down - show above selection instead
        y = rect.top - popupHeight - 10
      }

      // If still outside viewport vertically, position at vertical center
      if (y < 10 || y + popupHeight > viewport.height - 10) {
        y = (viewport.height / 2) - (popupHeight / 2)
      }

      // Final bounds check
      x = Math.max(10, Math.min(x, viewport.width - popupWidth - 10))
      y = Math.max(10, Math.min(y, viewport.height - popupHeight - 10))

      aiSuggestionsPosition.value = { 
        x: x + scrollX, 
        y: y + scrollY 
      }
      showAISuggestions.value = true
    } else {
      showAISuggestions.value = false
    }
  }, 100)
}


async function performAIAction(action: AIAction, text: string) {
  if (!text.trim()) return

  // Position popover near the current cursor/selection
  const selection: any = window.getSelection()
  let x = window.innerWidth / 2 - 160 // Center horizontally by default
  let y = window.innerHeight / 2 - 100 // Center vertically by default

  if (selection?.rangeCount > 0) {
    const range = selection.getRangeAt(0)
    const rect = range.getBoundingClientRect()

    // Try to position to the right of selection
    x = rect.right + 20
    y = rect.top

    // Viewport constraints
    const viewport = {
      width: window.innerWidth,
      height: window.innerHeight
    }

    // If too far right, position to the left
    if (x + 320 > viewport.width - 10) {
      x = rect.left - 330
    }

    // If still off-screen, center horizontally
    if (x < 10) {
      x = viewport.width / 2 - 160
    }

    // Vertical adjustments
    if (y + 300 > viewport.height - 10) {
      y = viewport.height - 310
    }
    if (y < 10) {
      y = 10
    }
  }

  aiPopoverPosition.value = { x, y }
  showAIPopover.value = true
  isLoadingAIPopover.value = true
  aiPopoverContent.value = ''
  originalTextForAI.value = text
  currentAIAction.value = action
  hideAISuggestions()

  try {
    const prompt = getAIPrompt(action, text)
    const result = await handlePrompt(prompt)
    aiPopoverContent.value = result.response || result.answer || 'No response generated'
  } catch (error) {
    console.error('Error performing AI action:', error)
    aiPopoverContent.value = 'Error generating content. Please try again.'
  } finally {
    isLoadingAIPopover.value = false
  }
}

function hideContextMenu() {
  showContextMenu.value = false
}

/**
 * Search functionality
 */
function performSearch() {
  if (!searchQuery.value.trim()) {
    searchResults.value = []
    return
  }

  const results: SearchResult[] = []
  const query = searchQuery.value.toLowerCase()

  editablePages.value.forEach(page => {
    const content = page.content.toLowerCase()
    let index = content.indexOf(query)

    while (index !== -1) {
      const start = Math.max(0, index - 20)
      const end = Math.min(content.length, index + query.length + 20)
      const context = page.content.substring(start, end)

      results.push({
        pageNum: page.pageNum,
        text: context,
        index
      })

      index = content.indexOf(query, index + 1)
    }
  })

  searchResults.value = results
  currentSearchIndex.value = 0
}

function goToSearchResult(index: number) {
  if (index >= 0 && index < searchResults.value.length) {
    currentSearchIndex.value = index
    const result = searchResults.value[index]
    goToPage(result.pageNum)
  }
}

/**
 * History management
 */
function addToHistory(action: string, preview: string) {
  editHistory.value.unshift({
    action,
    pageNum: currentPage.value,
    timestamp: new Date(),
    preview
  })

  if (editHistory.value.length > 50) {
    editHistory.value = editHistory.value.slice(0, 50)
  }
}

/**
 * Navigation functions
 */
function goToPage(pageNum: number) {
  if (pageNum >= 1 && pageNum <= totalPages.value) {
    currentPage.value = pageNum
  }
}

function previousPage() {
  goToPage(currentPage.value - 1)
}

function nextPage() {
  goToPage(currentPage.value + 1)
}

/**
 * Update page content
 */
function updatePageContent(content: string, saveHistory: boolean = true) {
  const pageIndex = currentPage.value - 1
  if (pageIndex >= 0 && pageIndex < editablePages.value.length) {
    if (saveHistory && editablePages.value[pageIndex].content !== content) {
      // Only save to history if content actually changed
      saveToHistory()
    }

    editablePages.value[pageIndex].content = content
    editablePages.value[pageIndex].isModified =
      content !== editablePages.value[pageIndex].originalContent
  }
}

function resetPageContent() {
  const pageIndex = currentPage.value - 1
  if (pageIndex >= 0 && pageIndex < editablePages.value.length) {
    saveToHistory()
    editablePages.value[pageIndex].content = editablePages.value[pageIndex].originalContent
    editablePages.value[pageIndex].isModified = false
    editablePages.value[pageIndex].annotations = []
    addToHistory('Reset page', `Reset page ${currentPage.value} to original`)
  }
}

function resetAllContent() {
  saveToHistory()
  editablePages.value.forEach(page => {
    page.content = page.originalContent
    page.isModified = false
    page.annotations = []
  })
  editHistory.value = []
  hasUnsavedChanges.value = false
  undoHistory.value = []
  redoHistory.value = []
  addToHistory('Reset all', 'Reset all pages to original')
}

function getCurrentPageContent(): EditableContent {
  const pageIndex = currentPage.value - 1
  return pageIndex >= 0 && pageIndex < editablePages.value.length
    ? editablePages.value[pageIndex]
    : {
      annotations: [],
      content: "",
      isModified: false,
      originalContent: "",
      pageNum: 0
    }
}

/**
 * File handling
 */
async function handleFileUpload(event: Event) {
  const target = event.target as HTMLInputElement
  const files = target.files ? Array.from(target.files) : []
  for (const file of files) {
    if (file.type === "application/pdf") {
      const url = URL.createObjectURL(file)
      const { previewUrl, pages } = await generatePdfThumbnail(file)
      const newFile = {
        id: `${Date.now()}-${Math.random().toString(36).slice(2)}`,
        name: file.name,
        url,
        type: file.type,
        size: file.size,
        previewUrl,
        pages,
        uploadedAt: new Date(),
      }
      uploadedFiles.value.push(newFile)
      saveToLocalStorage()
      openPdfEditor(newFile)
    }
  }
  target.value = ""
}

function handleDragOver(e: DragEvent) {
  e.preventDefault()
  isDragOver.value = true
}

function handleDragLeave(e: DragEvent) {
  e.preventDefault()
  isDragOver.value = false
}

function handleDrop(e: DragEvent) {
  e.preventDefault()
  isDragOver.value = false
  const files = e.dataTransfer?.files
  if (files) {
    const fakeEvent = { target: { files } } as any
    handleFileUpload(fakeEvent)
  }
}

async function generatePdfThumbnail(file: File): Promise<{ previewUrl: string; pages: number }> {
  try {
    const arrayBuffer = await file.arrayBuffer()
    const pdf = await pdfjsLib.getDocument({ data: arrayBuffer }).promise
    const page = await pdf.getPage(1)
    const viewport = page.getViewport({ scale: 0.3 })

    const canvas = document.createElement("canvas")
    const context = canvas.getContext("2d")
    if (!context) {
      return { previewUrl: "", pages: 0 }
    }

    canvas.width = viewport.width
    canvas.height = viewport.height

    await page.render({ canvasContext: context, viewport, canvas }).promise
    return {
      previewUrl: canvas.toDataURL('image/jpeg', 0.7),
      pages: pdf.numPages
    }
  } catch (err) {
    console.error("Error generating PDF thumbnail:", err)
    return { previewUrl: "", pages: 0 }
  }
}

function removeFile(id: string) {
  const index = uploadedFiles.value.findIndex(file => file.id === id)
  if (index > -1) {
    if (selectedFileId.value === id) {
      closeEditor()
    }

    const file = uploadedFiles.value[index]
    if (file.url.startsWith("blob:")) {
      URL.revokeObjectURL(file.url)
    }

    uploadedFiles.value.splice(index, 1)
    saveToLocalStorage()

    if (selectedFileId.value === id) {
      localStorage.removeItem(CURRENT_DOCUMENT_KEY)
    }
  }
}

/**
 * PDF content extraction
 */
async function openPdfEditor(file: UploadedFile) {
  if (file.isCustom) {
    openTextEditor(file)
    return
  }

  selectedPdfUrl.value = file.url
  selectedPdfName.value = file.name
  selectedFileId.value = file.id
  saveCurrentDocument()
  await extractPdfContent(file.url)
}

async function extractPdfContent(url: string) {
  isLoading.value = true
  loadError.value = ''
  editablePages.value = []

  try {
    const arrayBuffer = await fetch(url).then(response => response.arrayBuffer())
    const pdf = await pdfjsLib.getDocument({ data: arrayBuffer }).promise
    totalPages.value = pdf.numPages
    currentPage.value = 1

    let fullContent = ''

    for (let pageNum = 1; pageNum <= totalPages.value; pageNum++) {
      const page = await pdf.getPage(pageNum)
      const textContent = await page.getTextContent()

      let pageText = ''
      textContent.items.forEach((item) => {
        if ('str' in item) {
          pageText += item.str + ' '
        }
      })

      pageText = pageText.replace(/\s+/g, ' ').trim()
      if (pageText) {
        fullContent += `## Page ${pageNum}\n\n${pageText}\n\n`
      }

      editablePages.value.push({
        pageNum,
        content: pageText,
        originalContent: pageText,
        isModified: false,
        annotations: []
      })
    }

    // Save PDF content as a document
    const fileIndex = uploadedFiles.value.findIndex(f => f.id === selectedFileId.value)
    if (fileIndex !== -1) {
      uploadedFiles.value[fileIndex].content = fullContent
      uploadedFiles.value[fileIndex].isCustom = true
      saveToLocalStorage()
    }

  } catch (error: any) {
    loadError.value = `Failed to extract PDF content: ${error.message}`
  } finally {
    isLoading.value = false
  }
}

/**
 * Export functions
 */
function generatePreview() {
  let content = `${selectedPdfName.value} - Final Document\n`
  content += '='.repeat(60) + '\n\n'

  editablePages.value.forEach(page => {
    if (totalPages.value > 1) {
      content += `PAGE ${page.pageNum}\n`
      content += '-'.repeat(20) + '\n'
    }
    content += page.content + '\n\n'

    if (page.annotations.length > 0) {
      content += `📝 Annotations:\n`
      page.annotations.forEach(ann => {
        const icon = ann.type === 'highlight' ? '🔆' : ann.type === 'note' ? '📌' : '🔖'
        content += `${icon} ${ann.type.toUpperCase()}: "${ann.text}"`
        if (ann.note) content += ` - ${ann.note}`
        content += '\n'
      })
      content += '\n'
    }
    content += '\n'
  })

  previewContent.value = content
  showPreview.value = true
}

function closePreview() {
  showPreview.value = false
  previewContent.value = ''
}

function downloadAsText() {
  let content = `${selectedPdfName.value} - Content\n`
  content += '='.repeat(50) + '\n\n'

  editablePages.value.forEach(page => {
    if (totalPages.value > 1) {
      content += `--- Page ${page.pageNum} ---\n`
    }
    content += page.content + '\n\n'

    if (page.annotations.length > 0) {
      content += `Annotations:\n`
      page.annotations.forEach(ann => {
        content += `- ${ann.type}: "${ann.text}"${ann.note ? ` (${ann.note})` : ''}\n`
      })
      content += '\n'
    }
  })

  const blob = new Blob([content], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = selectedPdfName.value.replace(/\.[^/.]+$/, '') + '_content.txt'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
  closeExportDropdown()
}

function downloadAsMarkdown() {
  let content = `# ${selectedPdfName.value}\n\n`

  editablePages.value.forEach(page => {
    if (totalPages.value > 1) {
      content += `## Page ${page.pageNum}\n\n`
    }
    content += page.content + '\n\n'

    if (page.annotations.length > 0) {
      content += `### Annotations\n\n`
      page.annotations.forEach(ann => {
        content += `- **${ann.type}**: "${ann.text}"`
        if (ann.note) content += `\n  - Note: ${ann.note}`
        content += '\n'
      })
      content += '\n'
    }
  })

  const blob = new Blob([content], { type: 'text/markdown;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = selectedPdfName.value.replace(/\.[^/.]+$/, '') + '.md'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
  closeExportDropdown()
}

function closeEditor() {
  if (hasUnsavedChanges.value) {
    saveDocumentChanges()
  }

  selectedPdfUrl.value = ""
  selectedPdfName.value = ""
  selectedFileId.value = ""
  editablePages.value = []
  currentPage.value = 1
  totalPages.value = 0
  isLoading.value = false
  loadError.value = ''
  searchResults.value = []
  searchQuery.value = ''
  editHistory.value = []
  hasUnsavedChanges.value = false
  undoHistory.value = []
  redoHistory.value = []
  hideContextMenu()
  hideAISuggestions()

  if (autoSaveTimer.value) {
    clearTimeout(autoSaveTimer.value)
    autoSaveTimer.value = null
  }

  localStorage.removeItem(CURRENT_DOCUMENT_KEY)
}

function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

function getModifiedPagesCount(): number {
  return editablePages.value.filter(page => page.isModified).length
}

function cleanupFiles() {
  uploadedFiles.value.forEach((file) => {
    if (file.url.startsWith("blob:")) {
      URL.revokeObjectURL(file.url)
    }
  })
}

function applyAIResult() {
  const currentPageData = getCurrentPageContent()
  if (currentPageData && aiPopoverContent.value.trim()) {
    saveToHistory()
    const newContent = currentPageData.content.replace(originalTextForAI.value, aiPopoverContent.value.trim())
    updatePageContent(newContent)
    addToHistory(`AI ${currentAIAction.value}`, `Applied ${currentAIAction.value} to selected text`)
  }
  showAIPopover.value = false
  aiPopoverContent.value = ''
  originalTextForAI.value = ''
}

function handleGlobalClick(event: MouseEvent) {
  if (showExportDropdown.value) {
    const target = event.target as HTMLElement
    if (!target.closest('.export-dropdown')) {
      closeExportDropdown()
    }
  }

  if (showAISuggestions.value) {
    const target = event.target as HTMLElement
    if (!target.closest('.ai-suggestions')) {
      hideAISuggestions()
    }
  }

  // if (showAIPopover.value) {
  //   const target = event.target as HTMLElement
  //   if (!target.closest('.ai-popover')) {
  //     showAIPopover.value = false
  //   }
  // }
}

const toggleExportDropdown = () => {
  showExportDropdown.value = !showExportDropdown.value
}

const closeExportDropdown = () => {
  showExportDropdown.value = false
}

const handlePreview = () => {
  closeExportDropdown()
  generatePreview()
}

onMounted(() => {
  if(screenWidth < 720){
    router.push("/")
  }
  loadFromLocalStorage()
  document.addEventListener('mouseup', handleTextSelection)
  document.addEventListener('click', handleGlobalClick)
  document.addEventListener('mouseup', stopSidebarResize)
})

onUnmounted(() => {
  document.removeEventListener('mouseup', handleTextSelection)
  document.removeEventListener('click', handleGlobalClick)
  document.removeEventListener('mouseup', stopSidebarResize)
  document.removeEventListener('mousemove', handleSuggestionsDrag)
  document.removeEventListener('mouseup', stopSuggestionsDrag)

  if (autoSaveTimer.value) {
    clearTimeout(autoSaveTimer.value)
  }
})

onBeforeUnmount(() => {
  if (hasUnsavedChanges.value) {
    saveDocumentChanges()
  }
  cleanupFiles()
  document.removeEventListener('click', handleGlobalClick)
})
</script>

<template>
  <div class="w-full bg-gray-50 text-gray-900 min-h-screen dark:bg-gray-900 dark:text-gray-100">
    <!-- Main Editor Interface -->
    <div class="bg-white h-screen w-full flex flex-col lg:flex-row dark:bg-gray-800">
      <!-- Mobile Header (visible on mobile only) -->
      <div class="lg:hidden bg-gray-100 border-b border-gray-300 p-3 dark:bg-gray-700 dark:border-gray-600">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2">
            <button @click="sidebarOpen = !sidebarOpen" 
              class="p-2 rounded hover:bg-gray-200 dark:hover:bg-gray-600">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
              </svg>
            </button>
            <h3 class="font-medium text-gray-800 dark:text-gray-300 truncate">
              {{ selectedPdfName || "Gemmie Editor" }}
            </h3>
          </div>
          
          <!-- Mobile action buttons -->
          <div class="flex items-center gap-1">
            <button v-if="selectedPdfName" @click="editorMode = editorMode === 'edit' ? 'preview' : 'edit'" 
              class="p-2 rounded bg-blue-100 text-blue-700 hover:bg-blue-200 dark:bg-blue-900/20 dark:text-blue-400">
              <svg v-if="editorMode === 'edit'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
              </svg>
              <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
              </svg>
            </button>
            
            <div class="relative">
              <button @click="toggleExportDropdown" class="p-2 rounded hover:bg-gray-200 dark:hover:bg-gray-600">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
                </svg>
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Sidebar Overlay (mobile) -->
      <div v-if="sidebarOpen && innerWidth < 1024" 
        class="lg:hidden fixed inset-0 z-40 bg-black bg-opacity-50" 
        @click="sidebarOpen = false"></div>

      <!-- Sidebar -->
      <div :class="[
        'bg-gray-100 border-r border-gray-300 flex flex-col transition-all duration-300 dark:bg-gray-700 dark:border-gray-600 relative z-50',
        // Mobile: sidebar slides in from left, full height
        'lg:relative fixed inset-y-0 left-0',
        sidebarOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0',
        // Desktop: normal behavior with resizing
        sidebarOpen ? 'lg:flex' : 'lg:w-12',
        // Width management
        'w-80 sm:w-96 lg:w-auto'
      ]" :style="sidebarOpen && innerWidth >= 1024 ? { width: sidebarWidth + 'px' } : {}">

        <!-- Resize handle (desktop only) -->
        <div v-if="sidebarOpen && innerWidth >= 1024"
          class="absolute right-0 top-0 bottom-0 w-1 cursor-col-resize hover:bg-blue-500 transition-colors z-10 hidden lg:block"
          @mousedown="startSidebarResize">
        </div>

        <!-- Sidebar Header -->
        <div class="p-3 border-b border-gray-300 flex items-center justify-between dark:border-gray-600 flex-shrink-0">
          <h3 v-if="sidebarOpen" class="font-medium text-gray-800 dark:text-gray-300">Tools</h3>
          <button @click="sidebarOpen = !sidebarOpen"
            class="w-8 h-8 rounded hover:bg-gray-200 flex items-center justify-center transition-colors text-gray-800 dark:hover:bg-gray-600 dark:text-gray-400">
            <svg v-if="sidebarOpen" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
            </svg>
            <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
            </svg>
          </button>
        </div>

        <!-- Sidebar Tabs -->
        <div v-if="sidebarOpen" class="flex flex-wrap border-b border-gray-300 dark:border-gray-600 flex-shrink-0">
          <button v-for="tab in [
            { key: 'documents', label: 'Docs', icon: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z' },
            { key: 'outline', label: 'Pages', icon: 'M4 6h16M4 10h16M4 14h16M4 18h16' },
            { key: 'search', label: 'Search', icon: 'M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z' },
            { key: 'annotations', label: 'Notes', icon: 'M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z' },
            { key: 'history', label: 'History', icon: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z' }
          ]" :key="tab.key" @click="activeSidebarTab = tab.key" :class="[
            'flex-1 min-w-0 p-2 text-xs font-medium transition-colors border-b-2 text-center flex flex-col items-center gap-1 sm:flex-row sm:justify-center sm:gap-2',
            activeSidebarTab === tab.key
              ? 'border-blue-500 text-blue-600 bg-blue-50 dark:border-blue-400 dark:text-blue-400 dark:bg-blue-900/20'
              : 'border-transparent text-gray-800 hover:text-blue-600 dark:text-gray-400 dark:hover:text-blue-400'
          ]">
            <svg class="w-4 h-4 flex-shrink-0 sm:hidden" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" :d="tab.icon" />
            </svg>
            <span class="truncate">{{ tab.label }}</span>
          </button>
        </div>

        <!-- Sidebar Content -->
        <div v-if="sidebarOpen" class="flex-1 overflow-y-auto">
          <!-- Documents Tab -->
          <div v-if="activeSidebarTab === 'documents'" class="p-3">
            <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between mb-3 gap-2">
              <h4 class="text-sm font-medium text-gray-800 dark:text-gray-300">
                Documents ({{ uploadedFiles.length }})
              </h4>
              <div v-if="hasUnsavedChanges"
                class="flex items-center gap-1 text-xs text-orange-600 dark:text-orange-400 self-start sm:self-center">
                <svg class="w-3 h-3 animate-pulse" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                Auto-saving...
              </div>
            </div>

            <!-- Documents List -->
            <div class="space-y-2">
              <div v-for="file in uploadedFiles" :key="file.id" @click="openPdfEditor(file)" :class="[
                'p-3 rounded cursor-pointer transition-colors text-sm border',
                selectedFileId === file.id
                  ? 'bg-blue-50 text-blue-600 border-blue-500 dark:bg-blue-900/20 dark:text-blue-400 dark:border-blue-400'
                  : 'hover:bg-gray-50 border-gray-300 dark:hover:bg-gray-600 dark:border-gray-600'
              ]">
                <div class="flex items-start gap-3">
                  <svg v-if="file.isCustom" class="w-5 h-5 text-blue-500 flex-shrink-0 mt-0.5 dark:text-blue-400" fill="none"
                    stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                      d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                  <svg v-else class="w-5 h-5 text-red-500 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor"
                    viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                      d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707L13.293 3.293A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                  </svg>
                  <div class="flex-1 min-w-0">
                    <div class="font-medium truncate text-gray-800 dark:text-gray-300">{{ file.name }}</div>
                    <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">
                      {{ file.isCustom ? 'Markdown Document' : formatFileSize(file.size) }}
                      <span v-if="file.pages"> • {{ file.pages }} pages</span>
                    </div>
                  </div>
                  <button @click.stop="removeFile(file.id)"
                    class="w-8 h-8 rounded-full bg-red-100 hover:bg-red-200 flex items-center justify-center text-red-600 transition-colors dark:bg-red-900/20 dark:hover:bg-red-900/30 dark:text-red-400 flex-shrink-0"
                    title="Delete document">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                        d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                </div>
              </div>

              <div v-if="uploadedFiles.length === 0" class="text-gray-500 text-center py-8 text-sm dark:text-gray-400">
                <svg class="w-12 h-12 text-gray-300 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                <p class="mb-2">No documents yet</p>
                <p class="text-xs px-4">Use the toolbar buttons above to create or upload documents</p>
              </div>
            </div>
          </div>

          <!-- Other sidebar content remains the same but with responsive improvements -->
          <div v-if="activeSidebarTab === 'outline'" class="p-3">
            <h4 class="text-sm font-medium text-gray-800 dark:text-gray-300 mb-3">Document Pages</h4>
            <div class="space-y-1">
              <div v-for="page in editablePages" :key="page.pageNum" @click="goToPage(page.pageNum)" :class="[
                'p-3 rounded cursor-pointer transition-colors text-sm',
                currentPage === page.pageNum
                  ? 'bg-blue-50 text-blue-600 border border-blue-500 dark:bg-blue-900/20 dark:text-blue-400 dark:border-blue-400'
                  : 'hover:bg-gray-50 text-gray-800 dark:hover:bg-gray-600 dark:text-gray-300'
              ]">
                <div class="flex items-center justify-between mb-1">
                  <span class="font-medium truncate">{{ totalPages > 1 ? `Page ${page.pageNum}` : 'Document' }}</span>
                  <div class="flex items-center gap-1 flex-shrink-0 ml-2">
                    <div v-if="page.isModified" class="w-2 h-2 bg-orange-500 rounded-full" title="Modified"></div>
                    <span v-if="page.annotations.length > 0" class="text-xs text-blue-600 dark:text-blue-400 px-1 py-0.5 bg-blue-100 dark:bg-blue-900/20 rounded"
                      :title="`${page.annotations.length} annotations`">
                      {{ page.annotations.length }}
                    </span>
                  </div>
                </div>
                <div class="text-xs text-gray-500 dark:text-gray-400 truncate">
                  {{ page.content.substring(0, 80) }}...
                </div>
              </div>
            </div>
          </div>

          <!-- Search Tab -->
          <div v-if="activeSidebarTab === 'search'" class="p-3">
            <div class="mb-3 space-y-2">
              <input v-model="searchQuery" @keyup.enter="performSearch" type="text" placeholder="Search in document..."
                class="w-full px-3 py-2 border border-gray-300 rounded text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100" />
              <button @click="performSearch"
                class="w-full px-3 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 transition-colors flex items-center justify-center gap-2">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
                Search
              </button>
            </div>

            <div v-if="hasSearchResults" class="space-y-2">
              <div class="text-xs text-gray-500 dark:text-gray-400 mb-2">
                {{ searchResults.length }} results found
              </div>
              <div v-for="(result, index) in searchResults" :key="index" @click="goToSearchResult(index)" :class="[
                'p-3 rounded cursor-pointer text-xs transition-colors',
                index === currentSearchIndex
                  ? 'bg-yellow-100 border border-yellow-300 dark:bg-yellow-900/20 dark:border-yellow-600'
                  : 'hover:bg-gray-50 border border-transparent dark:hover:bg-gray-600'
              ]">
                <div class="font-medium text-gray-800 dark:text-gray-300 mb-1">Page {{ result.pageNum }}</div>
                <div class="text-gray-600 dark:text-gray-400 break-words">{{ result.text }}</div>
              </div>
            </div>
          </div>

          <!-- Annotations Tab -->
          <div v-if="activeSidebarTab === 'annotations'" class="p-3">
            <h4 class="text-sm font-medium text-gray-800 dark:text-gray-300 mb-3">Annotations ({{ allAnnotations.length }})</h4>
            <div class="space-y-2">
              <div v-for="ann in allAnnotations" :key="ann.id" @click="goToPage(ann.pageNum)"
                class="p-3 rounded border cursor-pointer hover:bg-gray-50 text-xs dark:hover:bg-gray-600"
                :style="{ borderLeftColor: ann.color, borderLeftWidth: '3px' }">
                <div class="flex items-center justify-between mb-1">
                  <span class="font-medium capitalize text-gray-800 dark:text-gray-300">{{ ann.type }}</span>
                  <span class="text-gray-500 dark:text-gray-400">Page {{ ann.pageNum }}</span>
                </div>
                <div class="text-gray-800 dark:text-gray-300 mb-1 break-words">"{{ ann.text.substring(0, 60) }}..."</div>
                <div v-if="ann.note" class="text-gray-500 dark:text-gray-400 break-words">{{ ann.note }}</div>
                <div class="text-gray-500 dark:text-gray-400 text-xs mt-1">
                  {{ ann.timestamp.toLocaleDateString() }}
                </div>
              </div>
              <div v-if="allAnnotations.length === 0" class="text-gray-500 text-center py-8 dark:text-gray-400">
                <svg class="w-8 h-8 text-gray-300 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z" />
                </svg>
                <p class="mb-1">No annotations yet</p>
                <p class="text-xs px-4">Select text to add notes</p>
              </div>
            </div>
          </div>

          <!-- History Tab -->
          <div v-if="activeSidebarTab === 'history'" class="p-3">
            <h4 class="text-sm font-medium text-gray-800 dark:text-gray-300 mb-3">Edit History</h4>
            <div class="space-y-2">
              <div v-for="(entry, index) in editHistory" :key="index"
                class="p-3 rounded border border-gray-300 hover:bg-gray-50 text-xs dark:border-gray-600 dark:hover:bg-gray-600">
                <div class="flex items-center justify-between mb-1">
                  <span class="font-medium text-gray-800 dark:text-gray-300 truncate">{{ entry.action }}</span>
                  <span class="text-gray-500 dark:text-gray-400 flex-shrink-0 ml-2">P{{ entry.pageNum }}</span>
                </div>
                <div class="text-gray-500 dark:text-gray-400 mb-1 break-words">{{ entry.preview }}</div>
                <div class="text-gray-500 dark:text-gray-400 text-xs">
                  {{ entry.timestamp.toLocaleTimeString() }}
                </div>
              </div>
              <div v-if="editHistory.length === 0" class="text-gray-500 text-center py-8 dark:text-gray-400">
                <svg class="w-8 h-8 text-gray-300 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <p>No edits yet</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Main Editor Area -->
      <div class="flex-1 flex flex-col min-w-0">
        <!-- Header (desktop only, hidden on mobile) -->
        <div class="hidden lg:flex flex-col border-b border-gray-300 bg-gray-50 dark:border-gray-600 dark:bg-gray-700/50">
          <!-- Title bar -->
          <div class="flex items-center justify-between p-3 xl:p-4 border-b border-gray-300 dark:border-gray-600">
            <div class="flex items-center gap-3 min-w-0 flex-1">
              <svg class="w-6 h-6 text-blue-600 flex-shrink-0 dark:text-blue-400" fill="none" stroke="currentColor"
                viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
              </svg>
              <div class="min-w-0 flex-1">
                <h3 class="font-semibold text-gray-900 truncate dark:text-gray-100">
                  {{ selectedPdfName || "Gemmie Editor" }}
                </h3>
                <p class="text-sm text-gray-600 dark:text-gray-400 truncate">
                  Markdown Editor with Live Preview
                  <span v-if="getModifiedPagesCount() > 0" class="text-orange-600 ml-2 dark:text-orange-400">
                    ({{ getModifiedPagesCount() }} page{{ getModifiedPagesCount() > 1 ? 's' : '' }} modified)
                  </span>
                  <span v-if="hasUnsavedChanges" class="text-blue-600 ml-1 dark:text-blue-400">• Auto-saving</span>
                </p>
              </div>
            </div>

            <div class="flex items-center gap-2 flex-shrink-0">
              <!-- Undo/Redo buttons -->
              <div class="hidden xl:flex items-center gap-1 border-r border-gray-300 pr-2 dark:border-gray-500">
                <button @click="undo" :disabled="!canUndo"
                  class="w-8 h-8 rounded bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center border border-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                  title="Undo (Ctrl+Z)">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                      d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
                  </svg>
                </button>
                <button @click="redo" :disabled="!canRedo"
                  class="w-8 h-8 rounded bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center border border-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                  title="Redo (Ctrl+Y)">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                      d="M21 10h-10a8 8 0 00-8 8v2m18-10l-6 6m6-6l-6-6" />
                  </svg>
                </button>
              </div>

              <!-- Editor Mode Toggle -->
              <div class="hidden sm:flex bg-gray-200 rounded-lg p-1 dark:bg-gray-600">
                <button @click="editorMode = 'edit'" :class="[
                  'px-2 xl:px-3 py-1 text-sm font-medium rounded transition-colors',
                  editorMode === 'edit'
                    ? 'bg-white text-gray-900 shadow-sm dark:bg-gray-700 dark:text-gray-100'
                    : 'text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-200'
                ]" title="Edit Mode">
                  Edit
                </button>
                <button @click="editorMode = 'preview'" :class="[
                  'px-2 xl:px-3 py-1 text-sm font-medium rounded transition-colors',
                  editorMode === 'preview'
                    ? 'bg-white text-gray-900 shadow-sm dark:bg-gray-700 dark:text-gray-100'
                    : 'text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-200'
                ]" title="Preview Mode">
                  Preview
                </button>
                <button @click="editorMode = 'split'" :class="[
                  'hidden lg:block px-2 xl:px-3 py-1 text-sm font-medium rounded transition-colors',
                  editorMode === 'split'
                    ? 'bg-white text-gray-900 shadow-sm dark:bg-gray-700 dark:text-gray-100'
                    : 'text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-200'
                ]" title="Split Mode">
                  Split
                </button>
              </div>

              <!-- Create Document Button -->
              <button @click="closeEditor"
                class="w-8 xl:w-10 h-8 xl:h-10 rounded-md bg-purple-600 hover:bg-purple-700 transition-colors text-white flex items-center justify-center"
                title="Create New Document">
                <svg class="w-4 xl:w-5 h-4 xl:h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                </svg>
              </button>

              <!-- Export Dropdown Menu -->
              <div class="relative export-dropdown">
                <button @click="toggleExportDropdown"
                  class="w-8 xl:w-10 h-8 xl:h-10 rounded-md bg-white hover:bg-gray-50 transition-colors text-gray-700 flex items-center justify-center border border-gray-200 dark:bg-gray-700 dark:hover:bg-gray-600 dark:text-gray-300 dark:border-gray-600"
                  title="Export Options">
                  <svg class="w-4 xl:w-5 h-4 xl:h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                      d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
                  </svg>
                </button>

                <!-- Dropdown Menu -->
                <div v-if="showExportDropdown"
                  class="absolute right-0 top-full mt-1 bg-white border border-gray-300 rounded-lg shadow-lg py-2 min-w-48 z-50 dark:bg-gray-800 dark:border-gray-600"
                  @click.stop>
                  <button @click="handlePreview" :disabled="editablePages.length === 0"
                    class="w-full px-4 py-2 text-left hover:bg-gray-100 flex items-center gap-3 text-sm text-gray-800 disabled:opacity-50 dark:hover:bg-gray-700 dark:text-gray-300"
                    title="Preview Final Document">
                    <svg class="w-4 h-4 text-purple-600 dark:text-purple-400" fill="none" stroke="currentColor"
                      viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                        d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                        d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                    </svg>
                    <span>Preview</span>
                  </button>

                  <button @click="downloadAsText" :disabled="editablePages.length === 0"
                    class="w-full px-4 py-2 text-left hover:bg-gray-100 flex items-center gap-3 text-sm text-gray-800 disabled:opacity-50 dark:hover:bg-gray-700 dark:text-gray-300"
                    title="Download as Text File">
                    <svg class="w-4 h-4 text-blue-600 dark:text-blue-400" fill="none" stroke="currentColor"
                      viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                        d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                    <span>Download TXT</span>
                  </button>

                  <button @click="downloadAsMarkdown" :disabled="editablePages.length === 0"
                    class="w-full px-4 py-2 text-left hover:bg-gray-100 flex items-center gap-3 text-sm text-gray-800 disabled:opacity-50 dark:hover:bg-gray-700 dark:text-gray-300"
                    title="Download as Markdown">
                    <svg class="w-4 h-4 text-green-600 dark:text-green-400" fill="none" stroke="currentColor"
                      viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                        d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                    <span>Download MD</span>
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- Markdown Toolbar (responsive) -->
          <div v-if="selectedPdfName && showMarkdownToolbar && (editorMode === 'edit' || editorMode === 'split')"
            class="flex items-center gap-1 p-2 bg-gray-100/50 border-b border-gray-300 overflow-x-auto dark:bg-gray-600/50 dark:border-gray-600">

            <!-- Text Formatting -->
            <div class="flex items-center gap-1 border-r border-gray-300 pr-2 mr-2 dark:border-gray-500">
              <button @click="insertBold"
                class="w-7 h-7 xl:w-8 xl:h-8 rounded bg-white hover:bg-gray-50 flex items-center justify-center border border-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                title="Bold (Ctrl+B)">
                <svg class="w-3 h-3 xl:w-4 xl:h-4 font-bold" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3"
                    d="M6 4h8a4 4 0 010 8H6zM6 12h9a4 4 0 010 8H6z" />
                </svg>
              </button>
              <button @click="insertItalic"
                class="w-7 h-7 xl:w-8 xl:h-8 rounded bg-white hover:bg-gray-50 flex items-center justify-center border border-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                title="Italic (Ctrl+I)">
                <svg class="w-3 h-3 xl:w-4 xl:h-4 italic" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 4l4 16M6 8h12M4 16h12" />
                </svg>
              </button>
              <button @click="insertCode"
                class="w-7 h-7 xl:w-8 xl:h-8 rounded bg-white hover:bg-gray-50 flex items-center justify-center border border-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                title="Code">
                <svg class="w-3 h-3 xl:w-4 xl:h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
                </svg>
              </button>
            </div>

            <!-- Headers -->
            <div class="flex items-center gap-1 border-r border-gray-300 pr-2 mr-2 dark:border-gray-500">
              <button @click="insertHeader(1)"
                class="w-7 h-7 xl:w-8 xl:h-8 rounded bg-white hover:bg-gray-50 flex items-center justify-center border border-gray-300 text-xs font-bold dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                title="Header 1">
                H1
              </button>
              <button @click="insertHeader(2)"
                class="w-7 h-7 xl:w-8 xl:h-8 rounded bg-white hover:bg-gray-50 flex items-center justify-center border border-gray-300 text-xs font-bold dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                title="Header 2">
                H2
              </button>
              <button @click="insertHeader(3)"
                class="hidden sm:flex w-7 h-7 xl:w-8 xl:h-8 rounded bg-white hover:bg-gray-50 items-center justify-center border border-gray-300 text-xs font-bold dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                title="Header 3">
                H3
              </button>
            </div>

            <!-- Lists and Links -->
            <div class="flex items-center gap-1 border-r border-gray-300 pr-2 mr-2 dark:border-gray-500">
              <button @click="insertList"
                class="w-7 h-7 xl:w-8 xl:h-8 rounded bg-white hover:bg-gray-50 flex items-center justify-center border border-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                title="Bullet List">
                <svg class="w-3 h-3 xl:w-4 xl:h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M8 6h13M8 12h13M8 18h13M3 6h.01M3 12h.01M3 18h.01" />
                </svg>
              </button>
              <button @click="insertNumberedList"
                class="w-7 h-7 xl:w-8 xl:h-8 rounded bg-white hover:bg-gray-50 flex items-center justify-center border border-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                title="Numbered List">
                <svg class="w-3 h-3 xl:w-4 xl:h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M9 5H7a2 2 0 00-2 2v6a2 2 0 002 2h2m0-8h10m-10 8h10" />
                </svg>
              </button>
              <button @click="insertLink"
                class="w-7 h-7 xl:w-8 xl:h-8 rounded bg-white hover:bg-gray-50 flex items-center justify-center border border-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                title="Link (Ctrl+K)">
                <svg class="w-3 h-3 xl:w-4 xl:h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
                </svg>
              </button>
            </div>

            <!-- Special Elements -->
            <div class="flex items-center gap-1">
              <button @click="insertQuote"
                class="w-7 h-7 xl:w-8 xl:h-8 rounded bg-white hover:bg-gray-50 flex items-center justify-center border border-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                title="Quote">
                <svg class="w-3 h-3 xl:w-4 xl:h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                </svg>
              </button>
              <button @click="insertTable"
                class="hidden sm:flex w-7 h-7 xl:w-8 xl:h-8 rounded bg-white hover:bg-gray-50 items-center justify-center border border-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                title="Table">
                <svg class="w-3 h-3 xl:w-4 xl:h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M3 10h18M3 14h18M10 3v18M14 3v18" />
                </svg>
              </button>
            </div>

            <!-- Toolbar Toggle -->
            <button @click="showMarkdownToolbar = !showMarkdownToolbar"
              class="ml-auto w-7 h-7 xl:w-8 xl:h-8 rounded bg-white hover:bg-gray-50 flex items-center justify-center border border-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
              title="Toggle Toolbar">
              <svg class="w-3 h-3 xl:w-4 xl:h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <!-- Page/Document Navigation (responsive) -->
          <div v-if="selectedPdfName" class="flex flex-col sm:flex-row sm:items-center gap-3 p-3 bg-gray-100/50 dark:bg-gray-600/50">
            <!-- Page navigation -->
            <div v-if="totalPages > 1"
              class="flex items-center gap-2 border-r border-gray-300 pr-3 sm:border-r-0 sm:pr-0 sm:border-b sm:border-gray-300 sm:pb-3 sm:mb-0 dark:border-gray-500">
              <button @click="previousPage" :disabled="currentPage <= 1"
                class="w-8 h-8 rounded bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center border border-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
                </svg>
              </button>
              <span class="text-sm text-gray-700 whitespace-nowrap dark:text-gray-400">
                {{ currentPage }} / {{ totalPages }}
              </span>
              <button @click="nextPage" :disabled="currentPage >= totalPages"
                class="w-8 h-8 rounded bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center border border-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                </svg>
              </button>
            </div>

            <!-- Text formatting -->
            <div class="flex flex-wrap items-center gap-2 border-r border-gray-300 pr-3 sm:border-r-0 sm:pr-0 dark:border-gray-500">
              <div class="flex items-center gap-2">
                <span class="text-sm text-gray-700 dark:text-gray-400">Font:</span>
                <select v-model="fontSize"
                  class="px-2 py-1 text-sm border border-gray-300 rounded bg-white dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100">
                  <option :value="12">12px</option>
                  <option :value="14">14px</option>
                  <option :value="16">16px</option>
                  <option :value="18">18px</option>
                  <option :value="20">20px</option>
                </select>
              </div>
              <div class="flex items-center gap-2">
                <span class="text-sm text-gray-700 dark:text-gray-400">Line:</span>
                <select v-model="lineHeight"
                  class="px-2 py-1 text-sm border border-gray-300 rounded bg-white dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100">
                  <option :value="1.2">1.2</option>
                  <option :value="1.4">1.4</option>
                  <option :value="1.5">1.5</option>
                  <option :value="1.6">1.6</option>
                  <option :value="1.8">1.8</option>
                  <option :value="2.0">2.0</option>
                </select>
              </div>
            </div>

            <!-- Page actions -->
            <div class="flex flex-wrap items-center gap-2">
              <button @click="resetPageContent" :disabled="!getCurrentPageContent()?.isModified"
                class="px-2 xl:px-3 py-1.5 bg-orange-100 text-orange-700 rounded-md hover:bg-orange-200 transition-colors text-sm font-medium disabled:opacity-50 dark:bg-orange-900/20 dark:text-orange-300 dark:hover:bg-orange-900/30">
                <svg class="inline w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
                </svg>
                Reset Page
              </button>
              <button @click="resetAllContent" :disabled="getModifiedPagesCount() === 0"
                class="px-2 xl:px-3 py-1.5 bg-red-100 text-red-700 rounded-md hover:bg-red-200 transition-colors text-sm font-medium disabled:opacity-50 dark:bg-red-900/20 dark:text-red-300 dark:hover:bg-red-900/30">
                <svg class="inline w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
                <span class="hidden sm:inline">Reset All</span>
                <span class="sm:hidden">Reset</span>
              </button>
            </div>
          </div>
        </div>

        <!-- Content Area -->
        <div class="flex-1 overflow-hidden bg-gray-50/30 relative dark:bg-gray-800/30">
          <!-- Loading state -->
          <div v-if="isLoading" class="flex items-center justify-center h-96">
            <div class="text-center px-4">
              <svg class="animate-spin w-8 h-8 text-blue-600 mx-auto mb-2 dark:text-blue-400" fill="none"
                viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z">
                </path>
              </svg>
              <p class="text-gray-900 dark:text-gray-100">Extracting text from PDF...</p>
              <p class="text-sm text-gray-600 dark:text-gray-400 mt-1">This may take a moment for large documents</p>
            </div>
          </div>

          <!-- Error state -->
          <div v-else-if="loadError" class="flex items-center justify-center h-96">
            <div class="text-center max-w-md mx-auto p-6">
              <svg class="w-12 h-12 text-red-600 mx-auto mb-2 dark:text-red-400" fill="none" stroke="currentColor"
                viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.34 16.5c-.77.833.192 2.5 1.732 2.5z" />
              </svg>
              <p class="text-red-600 mb-4 dark:text-red-400 px-4">{{ loadError }}</p>
              <div class="flex flex-col sm:flex-row gap-2 justify-center">
                <button @click="extractPdfContent(selectedPdfUrl)"
                  class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">
                  Try Again
                </button>
                <button @click="closeEditor"
                  class="px-4 py-2 bg-gray-200 text-gray-800 rounded hover:bg-gray-300 dark:bg-gray-600 dark:text-gray-300 dark:hover:bg-gray-500">
                  Close
                </button>
              </div>
            </div>
          </div>

          <!-- Template Selection Welcome Screen (responsive) -->
          <div v-else-if="!selectedPdfUrl"
            class="flex font-light text-sm pt-8 sm:pt-16 lg:pt-32 overflow-y-auto items-start justify-center h-full p-4 sm:p-8">
            <div class="max-w-6xl w-full">
              <div class="text-center mb-6 sm:mb-8">
                <h2 class="text-xl sm:text-2xl font-bold text-gray-900 dark:text-gray-100 mb-2">Welcome to Gemmie Editor</h2>
                <p class="text-base text-gray-600 dark:text-gray-400 mb-6 sm:mb-8 px-4">Choose a template to get started with your new document</p>
              </div>

              <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-6 mb-6 sm:mb-8">
                <button v-for="template in documentTemplates" :key="template.id"
                  @click="createDocumentFromTemplate(template)"
                  class="group p-4 sm:p-6 bg-white border-2 border-gray-200 rounded-xl hover:border-blue-500 hover:shadow-lg transition-all duration-200 text-left dark:bg-gray-800 dark:border-gray-700 dark:hover:border-blue-400">
                  <div class="flex items-start gap-3 sm:gap-4">
                    <div class="flex-shrink-0">
                      <span class="text-xl sm:text-2xl group-hover:scale-110 transition-transform duration-200">
                        {{ template.icon }}
                      </span>
                    </div>
                    <div class="flex-1 min-w-0">
                      <h3
                        class="text-sm sm:text-base font-semibold text-gray-900 dark:text-gray-100 mb-1 group-hover:text-blue-600 dark:group-hover:text-blue-400 leading-tight">
                        {{ template.name }}
                      </h3>
                      <p class="text-xs sm:text-sm text-gray-600 dark:text-gray-400 line-clamp-2 mb-2 sm:mb-0">
                        {{ template.content.substring(2, 100).replace(/\n/g, ' ').trim() }}...
                      </p>
                      <div
                        class="mt-2 sm:mt-3 text-xs text-blue-600 dark:text-blue-400 opacity-0 group-hover:opacity-100 transition-opacity">
                        Click to create →
                      </div>
                    </div>
                  </div>
                </button>
              </div>

              <!-- Optional PDF Upload Section (responsive) -->
              <div class="pt-4 sm:pt-6 border-t border-gray-200 dark:border-gray-700">
                <div class="text-center mb-4">
                  <p class="text-gray-600 dark:text-gray-400 px-4">Or upload a PDF to extract and edit its content</p>
                </div>
                <div :class="[
                  'border-2 border-dashed rounded-xl px-4 sm:px-8 py-6 text-center transition-all duration-200',
                  isDragOver ? 'border-blue-500 bg-blue-50 dark:border-blue-400 dark:bg-blue-900/20' : 'border-gray-300 hover:border-blue-400 bg-gray-50 dark:border-gray-600 dark:hover:border-blue-400 dark:bg-gray-700/30'
                ]" @dragover="handleDragOver" @dragleave="handleDragLeave" @drop="handleDrop">
                  <input id="pdfUploadWelcome" type="file" accept="application/pdf" class="hidden" multiple
                    @change="handleFileUpload" />
                  <svg class="w-8 h-8 text-gray-400 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                      d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                  </svg>
                  <label for="pdfUploadWelcome" class="cursor-pointer">
                    <span
                      class="text-base font-medium text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300">
                      Upload PDF Files
                    </span>
                    <p class="text-gray-500 dark:text-gray-400 mt-1">or drag and drop them here</p>
                  </label>
                </div>
              </div>
            </div>
          </div>

          <!-- Editor content (responsive) -->
          <div v-else-if="editablePages.length > 0" class="h-full overflow-auto px-3 sm:px-5">
            <!-- Edit Mode -->
            <div v-if="editorMode === 'edit'" class="bg-white mx-auto dark:bg-gray-800 h-full flex flex-col">
              <!-- Mobile page header -->
              <div class="lg:hidden py-2 w-full flex items-center justify-between border-b border-gray-200 dark:border-gray-600">
                <div class="flex items-center gap-2">
                  <span class="text-sm font-medium text-gray-700 dark:text-gray-400">
                    {{ totalPages > 1 ? `Page ${currentPage}` : 'Document' }}
                  </span>
                  <span v-if="getCurrentPageContent()?.isModified" 
                    class="w-2 h-2 bg-orange-500 rounded-full" title="Modified"></span>
                </div>
                
                <!-- Page navigation for mobile -->
                <div v-if="totalPages > 1" class="flex items-center gap-2">
                  <button @click="previousPage" :disabled="currentPage <= 1"
                    class="w-8 h-8 rounded bg-gray-100 hover:bg-gray-200 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center dark:bg-gray-700 dark:hover:bg-gray-600">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
                    </svg>
                  </button>
                  <span class="text-sm text-gray-600 dark:text-gray-400">{{ currentPage }}/{{ totalPages }}</span>
                  <button @click="nextPage" :disabled="currentPage >= totalPages"
                    class="w-8 h-8 rounded bg-gray-100 hover:bg-gray-200 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center dark:bg-gray-700 dark:hover:bg-gray-600">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                    </svg>
                  </button>
                </div>
              </div>

              <!-- Desktop page header -->
              <div class="hidden lg:block py-2 w-full">
                <div class="text-sm justify-end ml-auto flex gap-2 text-gray-600 dark:text-gray-400">
                  {{ totalPages > 1 ? `Page ${currentPage}` : '' }}
                  <p v-if="getCurrentPageContent()?.annotations?.length > 0">
                    {{ getCurrentPageContent()?.annotations.length }} annotations
                  </p>
                  {{ getCurrentPageContent()?.content?.length || 0 }} characters
                </div>
              </div>

              <!-- Text editor -->
              <div class="flex-1 py-2 relative">
                <textarea v-if="getCurrentPageContent()" :value="getCurrentPageContent()?.content"
                  @input="updatePageContent(($event.target as HTMLTextAreaElement).value)"
                  @keydown="handleTextareaKeydown"
                  class="w-full outline-none h-full px-3 sm:px-6 overflow-y-auto resize-none bg-white dark:bg-inherit text-gray-900 dark:text-gray-100"
                  :style="{
                    fontSize: fontSize + 'px',
                    lineHeight: lineHeight.toString(),
                    fontFamily: 'system-ui, -apple-system, sans-serif'
                  }" placeholder="Start writing your markdown content here... 

Press Ctrl+/ for AI assistance
Use Ctrl+B for bold, Ctrl+I for italic
Press Ctrl+Z to undo, Ctrl+Y to redo"></textarea>
              </div>
            </div>

            <!-- Preview Mode -->
            <div v-else-if="editorMode === 'preview'" class="bg-white mx-auto dark:bg-gray-800 h-full flex flex-col">
              <!-- Page header -->
              <div class="py-2 w-full flex items-center justify-between">
                <div class="text-sm justify-end ml-auto items-center flex gap-4 text-gray-600 dark:text-gray-400">
                  {{ totalPages > 1 ? `Page ${currentPage} - Preview` : 'Document Preview' }}
                  <span
                    class="px-2 py-1 bg-green-100 text-green-700 text-xs rounded dark:bg-green-900/20 dark:text-green-300">
                    Live Preview
                  </span>
                </div>
              </div>

              <!-- Preview content -->
              <div class="flex-1 py-2 relative">
                <div v-if="getCurrentPageContent()" 
                  class="prose prose-gray dark:prose-invert max-w-none px-3 sm:px-6 prose-sm sm:prose-base"
                  v-html="renderedMarkdown"></div>
                <div v-else class="text-center py-8 text-gray-500 dark:text-gray-400">
                  <p>No content to preview</p>
                </div>
              </div>
            </div>

            <!-- Split Mode (desktop only) -->
            <div v-else-if="editorMode === 'split'" class="bg-white dark:bg-gray-800 h-full flex flex-col">
              <!-- Page header -->
              <div class="border-b border-gray-300 p-4 flex items-center justify-between dark:border-gray-600">
                <div class="flex items-center gap-2">
                  <h3 class="font-medium text-gray-900 dark:text-gray-100">
                    {{ totalPages > 1 ? `Page ${currentPage} - Split View` : 'Document - Split View' }}
                  </h3>
                  <span v-if="getCurrentPageContent()?.isModified"
                    class="px-2 py-1 bg-orange-100 text-orange-700 text-xs rounded dark:bg-orange-900/20 dark:text-orange-300">
                    Modified
                  </span>
                </div>
                <div class="text-sm text-gray-600 dark:text-gray-400">
                  Editor & Preview
                </div>
              </div>

              <!-- Split content -->
              <div class="flex-1 flex flex-col lg:flex-row">
                <!-- Editor side -->
                <div class="flex-1 p-4 border-b lg:border-b-0 lg:border-r border-gray-300 dark:border-gray-600">
                  <div class="mb-2">
                    <span class="text-sm font-medium text-gray-600 dark:text-gray-400">Editor</span>
                  </div>
                  <textarea v-if="getCurrentPageContent()" :value="getCurrentPageContent()?.content"
                    @input="updatePageContent(($event.target as HTMLTextAreaElement).value)"
                    @keydown="handleTextareaKeydown"
                    class="w-full h-64 lg:h-full p-4 outline-none border border-gray-300 rounded-lg resize-none focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white text-gray-900 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100"
                    :style="{
                      fontSize: fontSize + 'px',
                      lineHeight: lineHeight.toString(),
                      fontFamily: 'system-ui, -apple-system, sans-serif'
                    }" placeholder="Start writing markdown..."></textarea>
                </div>

                <!-- Preview side -->
                <div class="flex-1 p-4">
                  <div class="mb-2">
                    <span class="text-sm font-medium text-gray-600 dark:text-gray-400">Preview</span>
                  </div>
                  <div
                    class="h-64 lg:h-full border border-gray-300 rounded-lg p-4 overflow-auto bg-gray-50 dark:border-gray-600 dark:bg-gray-700/50">
                    <div v-if="getCurrentPageContent()" class="prose prose-gray dark:prose-invert max-w-none prose-sm"
                      v-html="renderedMarkdown"></div>
                    <div v-else class="text-center py-8 text-gray-500 dark:text-gray-400">
                      <p>Preview will appear here</p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- No content state -->
          <div v-else class="flex items-center justify-center h-96">
            <div class="text-center max-w-md mx-auto p-6">
              <svg class="w-12 h-12 text-gray-400 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              <p class="text-gray-900 dark:text-gray-100 mb-2">No text content found</p>
              <p class="text-sm text-gray-600 dark:text-gray-400 px-4">This PDF may contain only images or have no extractable text.</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- AI Suggestions Popup (responsive positioning) -->
    <div v-if="showAISuggestions"
      class="ai-suggestions fixed z-50 bg-white border border-gray-300 rounded-lg shadow-xl dark:bg-gray-800 dark:border-gray-600 max-w-xs sm:max-w-sm"
      :style="{
        left: Math.min(aiSuggestionsPosition.x, innerWidth - 320) + 'px',
        top: Math.min(aiSuggestionsPosition.y, innerHeight - 300) + 'px'
      }">
      <!-- Drag Handle -->
      <div
        class="drag-handle cursor-move px-3 py-2 border-b border-gray-200 dark:border-gray-600 flex justify-between items-center"
        @mousedown="startSuggestionsDrag">
        <div class="text-xs font-medium text-gray-500 dark:text-gray-400">AI Assistant</div>
      </div>

      <div class="p-2">
        <div class="space-y-1">
          <button v-for="suggestion in aiSuggestions" :key="suggestion.action"
            @click.stop="selectAISuggestion(suggestion)"
            class="w-full px-3 py-2 text-left hover:bg-gray-100 rounded flex items-center gap-3 text-sm text-gray-800 dark:hover:bg-gray-700 dark:text-gray-300 transition-colors">
            <span class="text-lg flex-shrink-0">{{ suggestion.icon }}</span>
            <div class="flex-1 min-w-0">
              <div class="font-medium">{{ suggestion.label }}</div>
              <div class="text-xs text-gray-500 dark:text-gray-400 truncate">{{ suggestion.description }}</div>
            </div>
            <div class="hidden sm:block text-xs text-gray-400 dark:text-gray-500 flex-shrink-0">{{ suggestion.shortcut.replace('Ctrl+Shift+', 'C+S+') }}</div>
          </button>
        </div>
      </div>
    </div>

    <!-- AI Popover (responsive positioning) -->
    <div v-if="showAIPopover"
      class="ai-popover fixed z-50 bg-white border border-gray-300 rounded-lg shadow-xl w-80 max-w-[90vw] dark:bg-gray-800 dark:border-gray-600"
      :style="{
        left: Math.min(aiPopoverPosition.x, innerWidth - 340) + 'px',
        top: Math.min(aiPopoverPosition.y, innerHeight - 250) + 'px'
      }">
      <!-- Drag Handle -->
      <div
        class="drag-handle cursor-move px-4 py-3 border-b border-gray-200 dark:border-gray-600 flex justify-between items-center"
        @mousedown="startDrag($event, 'popover')">
        <div class="text-sm font-medium text-gray-700 dark:text-gray-300 capitalize">
          {{ currentAIAction }} Result
        </div>
        <button @click="showAIPopover = false"
          class="w-6 h-6 hover:bg-gray-200 rounded flex items-center justify-center transition-colors dark:hover:bg-gray-600 flex-shrink-0"
          @mousedown.stop>
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <div class="p-4">
        <div v-if="isLoadingAIPopover" class="flex items-center gap-2 text-blue-600 dark:text-blue-400">
          <svg class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z">
            </path>
          </svg>
          <span class="text-sm">Generating {{ currentAIAction }}...</span>
        </div>
        <div v-else>
          <div class="text-sm text-gray-800 max-h-[400px] overflow-y-auto dark:text-gray-100 mb-3 break-words">
            {{ aiPopoverContent }}
          </div>
          <div class="flex gap-2 flex-wrap">
            <button @click="applyAIResult" @mousedown.stop
              class="px-3 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 transition-colors">
              Apply
            </button>
            <button @click="showAIPopover = false" @mousedown.stop
              class="px-3 py-2 bg-gray-200 text-gray-800 rounded text-sm hover:bg-gray-300 transition-colors dark:bg-gray-600 dark:text-gray-300 dark:hover:bg-gray-500">
              Dismiss
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- AI Modal (responsive) -->
    <div v-if="showAIModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4"
      @click.self="closeAIModal">
      <div class="bg-white rounded-lg max-w-2xl w-full max-h-[90vh] border border-gray-300 dark:bg-gray-800 dark:border-gray-600 flex flex-col">
        <div class="p-4 border-b border-gray-300 flex items-center justify-between dark:border-gray-600 flex-shrink-0">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 capitalize">
            {{ currentAIAction }} Text
          </h3>
          <button @click="closeAIModal"
            class="w-8 h-8 rounded-full bg-gray-100 hover:bg-gray-200 flex items-center justify-center transition-colors dark:bg-gray-700 dark:hover:bg-gray-600">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div class="p-6 flex-1 overflow-y-auto">
          <div class="mb-4">
            <h4 class="text-sm font-medium text-gray-800 dark:text-gray-300 mb-2">Original Text:</h4>
            <div
              class="p-3 bg-gray-50 rounded text-sm text-gray-700 dark:bg-gray-700 dark:text-gray-300 max-h-32 overflow-y-auto break-words">
              {{ originalTextForAI }}
            </div>
          </div>
          <div class="mb-4">
            <h4 class="text-sm font-medium text-gray-800 dark:text-gray-300 mb-2">Result:</h4>
            <div v-if="isGeneratingAI"
              class="p-3 bg-blue-50 rounded text-sm text-blue-700 dark:bg-blue-900/20 dark:text-blue-300">
              <div class="flex items-center gap-2">
                <svg class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor"
                    d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z">
                  </path>
                </svg>
                Generating {{ currentAIAction }}...
              </div>
            </div>
            <textarea v-else v-model="aiContent" rows="4"
              class="w-full p-3 border border-gray-300 rounded text-sm resize-none focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white text-gray-900 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100"
              :placeholder="`${currentAIAction} result will appear here...`"></textarea>
          </div>
        </div>
        <div class="p-4 border-t border-gray-300 flex flex-col sm:flex-row gap-2 justify-end dark:border-gray-600 flex-shrink-0">
          <button @click="discardAIContent"
            class="px-4 py-2 text-gray-600 hover:text-gray-800 transition-colors dark:text-gray-400 dark:hover:text-gray-300 order-2 sm:order-1">
            Discard
          </button>
          <button @click="saveAIContent" :disabled="isGeneratingAI || !aiContent.trim()"
            class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors disabled:opacity-50 order-1 sm:order-2">
            Apply Changes
          </button>
        </div>
      </div>
    </div>

    <!-- Preview Modal (responsive) -->
    <div v-if="showPreview" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4"
      @click.self="closePreview">
      <div
        class="bg-white rounded-lg max-w-4xl w-full h-5/6 max-h-[90vh] border border-gray-300 dark:bg-gray-800 dark:border-gray-600 flex flex-col">
        <div class="p-4 border-b border-gray-300 flex items-center justify-between dark:border-gray-600 flex-shrink-0">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Document Preview</h3>
          <button @click="closePreview"
            class="w-8 h-8 rounded-full bg-gray-100 hover:bg-gray-200 flex items-center justify-center transition-colors dark:bg-gray-700 dark:hover:bg-gray-600">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div class="flex-1 p-6 overflow-auto">
          <pre class="whitespace-pre-wrap text-sm text-gray-800 dark:text-gray-200 font-mono break-words">{{ previewContent }}</pre>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.animate-spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(360deg);
  }
}

/* Custom scrollbar for webkit browsers */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

@media (min-width: 1024px) {
  ::-webkit-scrollbar {
    width: 8px;
    height: 8px;
  }
}

::-webkit-scrollbar-track {
  background: #f1f5f9;
}

::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: #94a3b8;
}

.dark ::-webkit-scrollbar-track {
  background: #374151;
}

.dark ::-webkit-scrollbar-thumb {
  background: #6b7280;
}

.dark ::-webkit-scrollbar-thumb:hover {
  background: #9ca3af;
}

/* Prose styles for better markdown rendering */
.prose {
  max-width: none;
}

.prose h1 {
  font-size: 2em;
  line-height: 1.2;
}

.prose h2 {
  font-size: 1.5em;
  line-height: 1.3;
}

.prose h3 {
  font-size: 1.25em;
  line-height: 1.4;
}

.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.cursor-move {
  cursor: move;
}

.cursor-col-resize {
  cursor: col-resize;
}

.ai-suggestions,
.ai-popover {
  user-select: none;
}

/* Responsive adjustments for mobile */
@media (max-width: 640px) {
  .prose {
    font-size: 14px;
  }
  
  .prose h1 {
    font-size: 1.5em;
  }
  
  .prose h2 {
    font-size: 1.3em;
  }
  
  .prose h3 {
    font-size: 1.1em;
  }
}

/* Touch-friendly interactions */
@media (max-width: 1024px) {
  .ai-suggestions button,
  .ai-popover button {
    min-height: 44px;
    padding: 12px;
  }
}

.drag-handle {
  touch-action: none;
}
</style>