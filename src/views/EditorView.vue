<script setup lang="ts">
import { ref, onBeforeUnmount, nextTick, computed, onMounted, onUnmounted, watch } from "vue"
import * as pdfjsLib from "pdfjs-dist"
import pdfjsWorker from "pdfjs-dist/build/pdf.worker?url"
import type { Ref } from "vue"
import { WRAPPER_URL } from "@/utils/globals"

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
const activeSidebarTab = ref<'outline' | 'search' | 'annotations' | 'history' | 'documents'| any>('documents')

// Text selection state
const selectedText = ref('')
const selectionStart = ref(0)
const selectionEnd = ref(0)
const showContextMenu = ref(false)
const contextMenuPosition = ref({ x: 0, y: 0 })

// Search functionality
const searchQuery = ref('')
const searchResults = ref<SearchResult[]>([])
const currentSearchIndex = ref(0)

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
const showCreateModal = ref(false)
const newDocTitle = ref('')
const newDocContent = ref('')

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
 * Markdown Rendering
 */
function renderMarkdown(text: string): string {
  if (!text) return ''

  let html = text
    // Headers
    .replace(/^### (.*$)/gm, '<h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2 mt-4">$1</h3>')
    .replace(/^## (.*$)/gm, '<h2 class="text-xl font-bold text-gray-900 dark:text-gray-100 mb-3 mt-6">$1</h2>')
    .replace(/^# (.*$)/gm, '<h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4 mt-8">$1</h1>')

    // Bold and italic
    .replace(/\*\*\*(.+?)\*\*\*/g, '<strong><em class="font-bold italic text-gray-900 dark:text-gray-100">$1</em></strong>')
    .replace(/\*\*(.+?)\*\*/g, '<strong class="font-bold text-gray-900 dark:text-gray-100">$1</strong>')
    .replace(/\*(.+?)\*/g, '<em class="italic text-gray-800 dark:text-gray-200">$1</em>')

    // Code
    .replace(/`([^`]+)`/g, '<code class="bg-gray-100 dark:bg-gray-800 px-1 py-0.5 rounded text-sm font-mono text-red-600 dark:text-red-400">$1</code>')

    // Links
    .replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" class="text-blue-600 dark:text-blue-400 hover:underline" target="_blank">$1</a>')

    // Strikethrough
    .replace(/~~(.+?)~~/g, '<del class="line-through text-gray-500 dark:text-gray-400">$1</del>')

    // Lists
    .replace(/^\* (.+$)/gm, '<li class="ml-4 text-gray-800 dark:text-gray-200">• $1</li>')
    .replace(/^- (.+$)/gm, '<li class="ml-4 text-gray-800 dark:text-gray-200">• $1</li>')
    .replace(/^\d+\. (.+$)/gm, '<li class="ml-4 text-gray-800 dark:text-gray-200">$1</li>')

    // Blockquotes
    .replace(/^> (.+$)/gm, '<blockquote class="border-l-4 border-gray-300 dark:border-gray-600 pl-4 italic text-gray-700 dark:text-gray-300 my-2">$1</blockquote>')

    // Line breaks
    .replace(/\n/g, '<br>')

    // Wrap lists
    .replace(/(<li class="ml-4[^>]*>.*?<\/li>)(?:\s*<br>\s*)?(?=<li class="ml-4|$)/gs, '$1')
    .replace(/(<li class="ml-4[^>]*>.*?<\/li>(?:\s*<li class="ml-4[^>]*>.*?<\/li>)*)/gs, '<ul class="my-2">$1</ul>')

  return html
}

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

async function performAIAction(action: AIAction, text: string) {
  if (!text.trim()) return

  isGeneratingAI.value = true
  originalTextForAI.value = text
  currentAIAction.value = action

  try {
    const prompt = getAIPrompt(action, text)
    const result = await handlePrompt(prompt)
    aiContent.value = result.response || result.answer || 'No response generated'
    showAIModal.value = true
  } catch (error) {
    console.error('Error performing AI action:', error)
    aiContent.value = 'Error generating content. Please try again.'
    showAIModal.value = true
  } finally {
    isGeneratingAI.value = false
  }
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
  const textarea = document.querySelector('textarea');
  
  if (!textarea) return;

  const rect = textarea.getBoundingClientRect();
  const scrollX = window.pageXOffset || document.documentElement.scrollLeft;
  const scrollY = window.pageYOffset || document.documentElement.scrollTop;
  console.log(rect)

  // Center position (assuming fixed toolbar dimensions)
  aiSuggestionsPosition.value = {
    x: rect.left + (rect.width / 2) - 100 + scrollX, // minus half of toolbar width
    y: (rect.height / 2) - scrollY   // minus half of toolbar height
  };

  showAISuggestions.value = true;
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

/**
 * Create New Document
 */
function createNewDocument() {
  if (!newDocTitle.value.trim()) return

  const newDoc: UploadedFile = {
    id: `doc-${Date.now()}-${Math.random().toString(36).slice(2)}`,
    name: newDocTitle.value.trim() + (newDocTitle.value.includes('.') ? '' : '.md'),
    url: 'custom',
    type: 'text/markdown',
    size: new Blob([newDocContent.value]).size,
    uploadedAt: new Date(),
    isCustom: true,
    content: newDocContent.value
  }

  uploadedFiles.value.push(newDoc)
  saveToLocalStorage()
  openTextEditor(newDoc)

  showCreateModal.value = false
  newDocTitle.value = ''
  newDocContent.value = ''
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

/**
 * Context Menu Functionality
 */
function handleRightClick(event: MouseEvent) {
  event.preventDefault()

  const selection = window.getSelection()
  const hasSelection = selection && selection.toString().trim().length > 0

  if (hasSelection) {
    selectedText.value = selection.toString().trim()
    const range = selection.getRangeAt(0)
    selectionStart.value = range.startOffset
    selectionEnd.value = range.endOffset
  }

  showContextMenu.value = true
  contextMenuPosition.value = {
    x: event.clientX,
    y: event.clientY
  }
}

function hideContextMenu() {
  showContextMenu.value = false
}

function contextMenuAction(action: string) {
  const selection = window.getSelection()
  const selectedTextValue = selection?.toString().trim()

  switch (action) {
    case 'copy':
      if (selectedTextValue) {
        navigator.clipboard.writeText(selectedTextValue)
      }
      break
    case 'selectAll':
      const textarea = document.querySelector('textarea')
      if (textarea) textarea.select()
      break
    case 'paste':
      navigator.clipboard.readText().then(text => {
        const textarea = document.querySelector('textarea')
        if (textarea) {
          saveToHistory()
          const start = textarea.selectionStart
          const end = textarea.selectionEnd
          const currentContent = textarea.value
          const newContent = currentContent.substring(0, start) + text + currentContent.substring(end)
          updatePageContent(newContent)
        }
      }).catch(() => {
        console.warn('Failed to read clipboard')
      })
      break
    case 'highlight':
      if (selectedTextValue) {
        highlightText('#ffff00')
      }
      break
    case 'note':
      if (selectedTextValue) {
        const note = prompt('Add a note for this text:')
        if (note) {
          addAnnotation('note', selectedTextValue, note)
        }
      }
      break
    case 'summarize':
    case 'expand':
    case 'simplify':
    case 'translate':
    case 'paraphrase':
    case 'improve':
      if (selectedTextValue) {
        selectedText.value = selectedTextValue
        performAIAction(action as AIAction, selectedTextValue)
      }
      break
    case 'bold':
      if (selectedTextValue) {
        makeBold()
      }
      break
    case 'erase':
      if (selectedTextValue) {
        eraseText()
      }
      break
  }
  hideContextMenu()
}

/**
 * Text actions
 */
function eraseText() {
  if (!selectedText.value) return

  const currentPageData = getCurrentPageContent()
  if (currentPageData) {
    saveToHistory()
    const newContent = currentPageData.content.replace(selectedText.value, '')
    updatePageContent(newContent)
    addToHistory('Erased text', `Removed "${selectedText.value.substring(0, 30)}..."`)
  }
}

function makeBold() {
  if (!selectedText.value) return

  const currentPageData = getCurrentPageContent()
  if (currentPageData) {
    saveToHistory()
    const newContent = currentPageData.content.replace(selectedText.value, `**${selectedText.value}**`)
    updatePageContent(newContent)
    addToHistory('Made text bold', `Bolded "${selectedText.value.substring(0, 30)}..."`)
  }
}

function highlightText(color: string = '#ffff00') {
  if (!selectedText.value) return

  addAnnotation('highlight', selectedText.value, '', color)
  addToHistory('Highlighted text', `Highlighted "${selectedText.value.substring(0, 30)}..."`)
}

function addAnnotation(type: Annotation['type'], text: string, note: string = '', color: string = '#ffff00') {
  const currentPageData = getCurrentPageContent()
  if (currentPageData) {
    const annotation: Annotation = {
      id: `ann-${Date.now()}-${Math.random().toString(36).slice(2)}`,
      type,
      text,
      startIndex: selectionStart.value,
      endIndex: selectionEnd.value,
      color,
      note,
      timestamp: new Date()
    }

    currentPageData.annotations.push(annotation)
    currentPageData.isModified = true
  }
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
  editHistory.value = []
  undoHistory.value = []
  redoHistory.value = []

  try {
    const arrayBuffer = await fetch(url).then(response => response.arrayBuffer())
    const pdf = await pdfjsLib.getDocument({ data: arrayBuffer }).promise

    totalPages.value = pdf.numPages
    currentPage.value = 1

    for (let pageNum = 1; pageNum <= totalPages.value; pageNum++) {
      try {
        const page = await pdf.getPage(pageNum)
        const textContent = await page.getTextContent()

        let pageText = ''
        let currentY = -1

        textContent.items.forEach((item: any) => {
          if ('str' in item && 'transform' in item) {
            const y = item.transform[5]

            if (currentY !== -1 && Math.abs(currentY - y) > 5) {
              pageText += '\n'
            }

            pageText += item.str + ' '
            currentY = y
          }
        })

        pageText = pageText
          .replace(/\s+/g, ' ')
          .replace(/\n\s+/g, '\n')
          .trim()

        if (!pageText) {
          pageText = `[Page ${pageNum} - No extractable text content or image-based content]`
        }

        editablePages.value.push({
          pageNum,
          content: pageText,
          originalContent: pageText,
          isModified: false,
          annotations: []
        })

      } catch (error: any) {
        console.error(`Error extracting text from page ${pageNum}:`, error)
        editablePages.value.push({
          pageNum,
          content: `[Error loading page ${pageNum}: ${error.message}]`,
          originalContent: '',
          isModified: false,
          annotations: []
        })
      }
    }
  } catch (error: any) {
    console.error('Error extracting PDF content:', error)
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

function handleGlobalClick(event: MouseEvent) {
  if (showContextMenu.value) {
    hideContextMenu()
  }

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
  loadFromLocalStorage()
  document.addEventListener('click', handleGlobalClick)
})

onUnmounted(() => {
  document.removeEventListener('click', handleGlobalClick)
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
    <div class="bg-white h-screen w-full flex dark:bg-gray-800">
      <!-- Sidebar -->
      <div :class="[
        'bg-gray-100 border-r border-gray-300 flex flex-col transition-all duration-300 dark:bg-gray-700 dark:border-gray-600',
        sidebarOpen ? 'w-80' : 'w-12'
      ]">

        <!-- Sidebar Header -->
        <div class="p-3 border-b border-gray-300 flex items-center justify-between dark:border-gray-600">
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
        <div v-if="sidebarOpen" class="flex border-b border-gray-300 dark:border-gray-600">
          <button v-for="tab in [
            { key: 'documents', label: 'Docs' },
            { key: 'outline', label: 'Pages' },
            { key: 'search', label: 'Search' },
            { key: 'annotations', label: 'Notes' },
            { key: 'history', label: 'History' }
          ]" :key="tab.key" @click="activeSidebarTab = tab.key" :class="[
            'flex-1 p-2 text-xs font-medium transition-colors border-b-2 text-center',
            activeSidebarTab === tab.key
              ? 'border-blue-500 text-blue-600 bg-blue-50 dark:border-blue-400 dark:text-blue-400 dark:bg-blue-900/20'
              : 'border-transparent text-gray-800 hover:text-blue-600 dark:text-gray-400 dark:hover:text-blue-400'
          ]">
            {{ tab.label }}
          </button>
        </div>

        <!-- Sidebar Content -->
        <div v-if="sidebarOpen" class="flex-1 overflow-y-auto">

          <!-- Documents Tab -->
          <div v-if="activeSidebarTab === 'documents'" class="p-3">
            <!-- Create New Document Button -->
            <button @click="showCreateModal = true"
              class="w-full mb-4 p-3 border-2 border-dashed border-blue-500 rounded-lg text-center transition-colors hover:border-blue-600 bg-blue-50 hover:bg-blue-100 dark:border-blue-400 dark:hover:border-blue-300 dark:bg-blue-900/20 dark:hover:bg-blue-900/30">
              <div class="flex flex-col items-center gap-2 text-blue-600 dark:text-blue-400">
                <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                </svg>
                <span class="font-medium text-sm">Create New Document</span>
              </div>
            </button>

            <!-- Optional PDF Upload Area -->
            <div :class="[
              'border-2 border-dashed rounded-lg p-3 text-center transition-all duration-200 mb-4',
              isDragOver ? 'border-blue-500 bg-blue-50 dark:border-blue-400 dark:bg-blue-900/20' : 'border-gray-300 hover:border-blue-500 bg-gray-50 dark:border-gray-600 dark:hover:border-blue-400 dark:bg-gray-600/30'
            ]" @dragover="handleDragOver" @dragleave="handleDragLeave" @drop="handleDrop">
              <input id="pdfUpload" type="file" accept="application/pdf" class="hidden" multiple
                @change="handleFileUpload" />

              <div class="flex flex-col items-center gap-2">
                <svg class="w-5 h-5 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor"
                  viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                </svg>
                <label for="pdfUpload"
                  class="cursor-pointer text-blue-600 hover:text-blue-700 font-medium text-sm dark:text-blue-400 dark:hover:text-blue-300">
                  Upload PDF (Optional)
                </label>
              </div>
            </div>

            <div class="flex items-center justify-between mb-3">
              <h4 class="text-sm font-medium text-gray-800 dark:text-gray-300">
                My Documents ({{ uploadedFiles.length }})
              </h4>
              <div v-if="hasUnsavedChanges"
                class="flex items-center gap-1 text-xs text-orange-600 dark:text-orange-400">
                <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                Auto-saving...
              </div>
            </div>

            <!-- Documents List -->
            <div class="space-y-2">
              <div v-for="file in uploadedFiles" :key="file.id" @click="openPdfEditor(file)" :class="[
                'p-2 rounded cursor-pointer transition-colors text-sm border',
                selectedFileId === file.id
                  ? 'bg-blue-50 text-blue-600 border-blue-500 dark:bg-blue-900/20 dark:text-blue-400 dark:border-blue-400'
                  : 'hover:bg-gray-50 border-gray-300 dark:hover:bg-gray-600 dark:border-gray-600'
              ]">
                <div class="flex items-center gap-2">
                  <svg v-if="file.isCustom" class="w-4 h-4 text-blue-500 flex-shrink-0 dark:text-blue-400" fill="none"
                    stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                      d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                  <svg v-else class="w-4 h-4 text-red-500 flex-shrink-0" fill="none" stroke="currentColor"
                    viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                      d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707L13.293 3.293A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                  </svg>
                  <div class="flex-1 min-w-0">
                    <div class="font-medium truncate text-gray-800 dark:text-gray-300">{{ file.name }}</div>
                    <div class="text-xs text-gray-500 dark:text-gray-400">
                      {{ file.isCustom ? 'Markdown Document' : formatFileSize(file.size) }}
                      <span v-if="file.pages"> • {{ file.pages }} pages</span>
                    </div>
                  </div>
                  <button @click.stop="removeFile(file.id)"
                    class="w-6 h-6 rounded-full bg-red-100 hover:bg-red-200 flex items-center justify-center text-red-600 transition-colors dark:bg-red-900/20 dark:hover:bg-red-900/30 dark:text-red-400"
                    title="Delete document">
                    <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                        d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                </div>
              </div>
              <div v-if="uploadedFiles.length === 0" class="text-gray-500 text-center py-4 text-sm dark:text-gray-400">
                No documents created yet
              </div>
            </div>
          </div>

          <!-- Other sidebar tabs content -->
          <div v-if="activeSidebarTab === 'outline'" class="p-3">
            <h4 class="text-sm font-medium text-gray-800 dark:text-gray-300 mb-3">Document Pages</h4>
            <div class="space-y-1">
              <div v-for="page in editablePages" :key="page.pageNum" @click="goToPage(page.pageNum)" :class="[
                'p-2 rounded cursor-pointer transition-colors text-sm',
                currentPage === page.pageNum
                  ? 'bg-blue-50 text-blue-600 border border-blue-500 dark:bg-blue-900/20 dark:text-blue-400 dark:border-blue-400'
                  : 'hover:bg-gray-50 text-gray-800 dark:hover:bg-gray-600 dark:text-gray-300'
              ]">
                <div class="flex items-center justify-between">
                  <span class="font-medium">{{ totalPages > 1 ? `Page ${page.pageNum}` : 'Document' }}</span>
                  <div class="flex items-center gap-1">
                    <div v-if="page.isModified" class="w-2 h-2 bg-orange-500 rounded-full" title="Modified"></div>
                    <span v-if="page.annotations.length > 0" class="text-xs text-blue-600 dark:text-blue-400"
                      :title="`${page.annotations.length} annotations`">
                      {{ page.annotations.length }}
                    </span>
                  </div>
                </div>
                <div class="text-xs text-gray-500 dark:text-gray-400 mt-1 truncate">
                  {{ page.content.substring(0, 60) }}...
                </div>
              </div>
            </div>
          </div>

          <div v-if="activeSidebarTab === 'search'" class="p-3">
            <div class="mb-3">
              <input v-model="searchQuery" @keyup.enter="performSearch" type="text" placeholder="Search in document..."
                class="w-full px-3 py-2 border border-gray-300 rounded text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100" />
              <button @click="performSearch"
                class="w-full mt-2 px-3 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 transition-colors">
                <svg class="inline w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
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
                'p-2 rounded cursor-pointer text-xs transition-colors',
                index === currentSearchIndex
                  ? 'bg-yellow-100 border border-yellow-300 dark:bg-yellow-900/20 dark:border-yellow-600'
                  : 'hover:bg-gray-50 border border-transparent dark:hover:bg-gray-600'
              ]">
                <div class="font-medium text-gray-800 dark:text-gray-300 mb-1">Page {{ result.pageNum }}</div>
                <div class="text-gray-800 dark:text-gray-400">{{ result.text }}</div>
              </div>
            </div>
          </div>

          <div v-if="activeSidebarTab === 'annotations'" class="p-3">
            <h4 class="text-sm font-medium text-gray-800 dark:text-gray-300 mb-3">Annotations ({{ allAnnotations.length
              }})</h4>
            <div class="space-y-2">
              <div v-for="ann in allAnnotations" :key="ann.id" @click="goToPage(ann.pageNum)"
                class="p-2 rounded border cursor-pointer hover:bg-gray-50 text-xs dark:hover:bg-gray-600"
                :style="{ borderLeftColor: ann.color, borderLeftWidth: '3px' }">
                <div class="flex items-center justify-between mb-1">
                  <span class="font-medium capitalize text-gray-800 dark:text-gray-300">{{ ann.type }}</span>
                  <span class="text-gray-500 dark:text-gray-400">Page {{ ann.pageNum }}</span>
                </div>
                <div class="text-gray-800 dark:text-gray-300 mb-1">"{{ ann.text.substring(0, 50) }}..."</div>
                <div v-if="ann.note" class="text-gray-500 dark:text-gray-400">{{ ann.note }}</div>
                <div class="text-gray-500 dark:text-gray-400 text-xs mt-1">
                  {{ ann.timestamp.toLocaleDateString() }}
                </div>
              </div>
              <div v-if="allAnnotations.length === 0" class="text-gray-500 text-center py-4 dark:text-gray-400">
                No annotations yet. Select text to add notes.
              </div>
            </div>
          </div>

          <div v-if="activeSidebarTab === 'history'" class="p-3">
            <h4 class="text-sm font-medium text-gray-800 dark:text-gray-300 mb-3">Edit History</h4>
            <div class="space-y-2">
              <div v-for="(entry, index) in editHistory" :key="index"
                class="p-2 rounded border border-gray-300 hover:bg-gray-50 text-xs dark:border-gray-600 dark:hover:bg-gray-600">
                <div class="flex items-center justify-between mb-1">
                  <span class="font-medium text-gray-800 dark:text-gray-300">{{ entry.action }}</span>
                  <span class="text-gray-500 dark:text-gray-400">P{{ entry.pageNum }}</span>
                </div>
                <div class="text-gray-500 dark:text-gray-400 mb-1">{{ entry.preview }}</div>
                <div class="text-gray-500 dark:text-gray-400 text-xs">
                  {{ entry.timestamp.toLocaleTimeString() }}
                </div>
              </div>
              <div v-if="editHistory.length === 0" class="text-gray-500 text-center py-4 dark:text-gray-400">
                No edits yet
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Main Editor Area -->
      <div class="flex-1 flex flex-col">
        <!-- Header -->
        <div class="flex flex-col border-b border-gray-300 bg-gray-50 dark:border-gray-600 dark:bg-gray-700/50">
          <!-- Title bar -->
          <div class="flex items-center justify-between p-3 sm:p-4 border-b border-gray-300 dark:border-gray-600">
            <div class="flex items-center gap-3 min-w-0 flex-1">
              <svg class="w-6 h-6 text-blue-600 flex-shrink-0 dark:text-blue-400" fill="none" stroke="currentColor"
                viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
              </svg>
              <div class="min-w-0 flex-1">
                <h3 class="font-semibold text-gray-900 truncate dark:text-gray-100">{{ selectedPdfName || 'Document Editor' }}</h3>
                <p class="text-sm text-gray-600 dark:text-gray-400">
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
              <div class="flex items-center gap-1 border-r border-gray-300 pr-2 dark:border-gray-500">
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
              <div class="flex bg-gray-200 rounded-lg p-1 dark:bg-gray-600">
                <button @click="editorMode = 'edit'" :class="[
                  'px-3 py-1 text-sm font-medium rounded transition-colors',
                  editorMode === 'edit'
                    ? 'bg-white text-gray-900 shadow-sm dark:bg-gray-700 dark:text-gray-100'
                    : 'text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-200'
                ]" title="Edit Mode">
                  Edit
                </button>
                <button @click="editorMode = 'preview'" :class="[
                  'px-3 py-1 text-sm font-medium rounded transition-colors',
                  editorMode === 'preview'
                    ? 'bg-white text-gray-900 shadow-sm dark:bg-gray-700 dark:text-gray-100'
                    : 'text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-200'
                ]" title="Preview Mode">
                  Preview
                </button>
                <button @click="editorMode = 'split'" :class="[
                  'px-3 py-1 text-sm font-medium rounded transition-colors',
                  editorMode === 'split'
                    ? 'bg-white text-gray-900 shadow-sm dark:bg-gray-700 dark:text-gray-100'
                    : 'text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-200'
                ]" title="Split Mode">
                  Split
                </button>
              </div>

              <!-- AI Assistant Button -->
              <button @click="showAIToolbar"
                class="w-10 h-10 rounded-md bg-purple-600 hover:bg-purple-700 transition-colors text-white flex items-center justify-center"
                title="AI Assistant (Ctrl+/)">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M13 10V3L4 14h7v7l9-11h-7z" />
                </svg>
              </button>

              <!-- Export Dropdown Menu -->
              <div class="relative export-dropdown">
                <button @click="toggleExportDropdown"
                  class="w-10 h-10 rounded-md bg-white hover:bg-gray-50 transition-colors text-gray-700 flex items-center justify-center border border-gray-200 dark:bg-gray-700 dark:hover:bg-gray-600 dark:text-gray-300 dark:border-gray-600"
                  title="Export Options">
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
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

          <!-- Markdown Toolbar -->
          <div v-if="selectedPdfName && showMarkdownToolbar && (editorMode === 'edit' || editorMode === 'split')"
            class="flex items-center gap-1 p-2 bg-gray-100/50 border-b border-gray-300 overflow-x-auto dark:bg-gray-600/50 dark:border-gray-600">

            <!-- Text Formatting -->
            <div class="flex items-center gap-1 border-r border-gray-300 pr-2 mr-2 dark:border-gray-500">
              <button @click="insertBold"
                class="w-8 h-8 rounded bg-white hover:bg-gray-50 flex items-center justify-center border border-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                title="Bold (Ctrl+B)">
                <svg class="w-4 h-4 font-bold" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3"
                    d="M6 4h8a4 4 0 010 8H6zM6 12h9a4 4 0 010 8H6z" />
                </svg>
              </button>
              <button @click="insertItalic"
                class="w-8 h-8 rounded bg-white hover:bg-gray-50 flex items-center justify-center border border-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                title="Italic (Ctrl+I)">
                <svg class="w-4 h-4 italic" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 4l4 16M6 8h12M4 16h12" />
                </svg>
              </button>
              <button @click="insertCode"
                class="w-8 h-8 rounded bg-white hover:bg-gray-50 flex items-center justify-center border border-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                title="Code">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
                </svg>
              </button>
            </div>

            <!-- Headers -->
            <div class="flex items-center gap-1 border-r border-gray-300 pr-2 mr-2 dark:border-gray-500">
              <button @click="insertHeader(1)"
                class="w-8 h-8 rounded bg-white hover:bg-gray-50 flex items-center justify-center border border-gray-300 text-xs font-bold dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                title="Header 1">
                H1
              </button>
              <button @click="insertHeader(2)"
                class="w-8 h-8 rounded bg-white hover:bg-gray-50 flex items-center justify-center border border-gray-300 text-xs font-bold dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                title="Header 2">
                H2
              </button>
              <button @click="insertHeader(3)"
                class="w-8 h-8 rounded bg-white hover:bg-gray-50 flex items-center justify-center border border-gray-300 text-xs font-bold dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                title="Header 3">
                H3
              </button>
            </div>

            <!-- Lists and Links -->
            <div class="flex items-center gap-1 border-r border-gray-300 pr-2 mr-2 dark:border-gray-500">
              <button @click="insertList"
                class="w-8 h-8 rounded bg-white hover:bg-gray-50 flex items-center justify-center border border-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                title="Bullet List">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M8 6h13M8 12h13M8 18h13M3 6h.01M3 12h.01M3 18h.01" />
                </svg>
              </button>
              <button @click="insertNumberedList"
                class="w-8 h-8 rounded bg-white hover:bg-gray-50 flex items-center justify-center border border-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                title="Numbered List">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M9 5H7a2 2 0 00-2 2v6a2 2 0 002 2h2m0-8h10m-10 8h10" />
                </svg>
              </button>
              <button @click="insertLink"
                class="w-8 h-8 rounded bg-white hover:bg-gray-50 flex items-center justify-center border border-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                title="Link (Ctrl+K)">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
                </svg>
              </button>
            </div>

            <!-- Special Elements -->
            <div class="flex items-center gap-1">
              <button @click="insertQuote"
                class="w-8 h-8 rounded bg-white hover:bg-gray-50 flex items-center justify-center border border-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                title="Quote">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                </svg>
              </button>
              <button @click="insertTable"
                class="w-8 h-8 rounded bg-white hover:bg-gray-50 flex items-center justify-center border border-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                title="Table">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M3 10h18M3 14h18M10 3v18M14 3v18" />
                </svg>
              </button>
            </div>

            <!-- Toolbar Toggle -->
            <button @click="showMarkdownToolbar = !showMarkdownToolbar"
              class="ml-auto w-8 h-8 rounded bg-white hover:bg-gray-50 flex items-center justify-center border border-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
              title="Toggle Toolbar">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <!-- Page/Document Navigation -->
          <div v-if="selectedPdfName" class="flex flex-wrap items-center gap-3 p-3 bg-gray-100/50 dark:bg-gray-600/50">
            <!-- Page navigation -->
            <div v-if="totalPages > 1"
              class="flex items-center gap-2 border-r border-gray-300 pr-3 dark:border-gray-500">
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
            <div class="flex items-center gap-2 border-r border-gray-300 pr-3 dark:border-gray-500">
              <span class="text-sm text-gray-700 dark:text-gray-400">Font:</span>
              <select v-model="fontSize"
                class="px-2 py-1 text-sm border border-gray-300 rounded bg-white dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100">
                <option :value="12">12px</option>
                <option :value="14">14px</option>
                <option :value="16">16px</option>
                <option :value="18">18px</option>
                <option :value="20">20px</option>
              </select>
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

            <!-- Page actions -->
            <div class="flex items-center gap-2">
              <button @click="resetPageContent" :disabled="!getCurrentPageContent()?.isModified"
                class="px-3 py-1.5 bg-orange-100 text-orange-700 rounded-md hover:bg-orange-200 transition-colors text-sm font-medium disabled:opacity-50 dark:bg-orange-900/20 dark:text-orange-300 dark:hover:bg-orange-900/30">
                <svg class="inline w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
                </svg>
                Reset Page
              </button>
              <button @click="resetAllContent" :disabled="getModifiedPagesCount() === 0"
                class="px-3 py-1.5 bg-red-100 text-red-700 rounded-md hover:bg-red-200 transition-colors text-sm font-medium disabled:opacity-50 dark:bg-red-900/20 dark:text-red-300 dark:hover:bg-red-900/30">
                <svg class="inline w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
                Reset All
              </button>
            </div>
          </div>
        </div>

        <!-- Content Area -->
        <div class="flex-1 overflow-hidden bg-gray-50/30 relative dark:bg-gray-800/30">
          <!-- Loading state -->
          <div v-if="isLoading" class="flex items-center justify-center h-96">
            <div class="text-center">
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
              <p class="text-red-600 mb-4 dark:text-red-400">{{ loadError }}</p>
              <button @click="extractPdfContent(selectedPdfUrl)"
                class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 mr-2">
                Try Again
              </button>
              <button @click="closeEditor"
                class="px-4 py-2 bg-gray-200 text-gray-800 rounded hover:bg-gray-300 dark:bg-gray-600 dark:text-gray-300 dark:hover:bg-gray-500">
                Close
              </button>
            </div>
          </div>

          <!-- No document selected state -->
          <div v-else-if="!selectedPdfUrl" class="flex items-center justify-center h-96">
            <div class="text-center max-w-md mx-auto p-6">
              <svg class="w-12 h-12 text-gray-400 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100 mb-2">No Document Selected</h3>
              <p class="text-gray-600 dark:text-gray-400 mb-4">Create a new document to get started with markdown
                editing and live preview.</p>
              <button @click="showCreateModal = true"
                class="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">
                <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                </svg>
                Create New Document
              </button>
            </div>
          </div>

          <!-- Editor content -->
          <div v-else-if="editablePages.length > 0" class="h-full overflow-auto px-5">
            <!-- Edit Mode -->
            <div v-if="editorMode === 'edit'"
              class="bg-white mx-auto dark:bg-gray-800 h-full flex flex-col">
              <!-- Page header -->
              <div class="py-2 w-full flex items-center justify-between dark:border-gray-600">
                <div class="text-sm justify-end ml-auto flex gap-2 text-gray-600 dark:text-gray-400">
                  {{ totalPages > 1 ? `Page ${currentPage}` : 'Document' }}
                  <p v-if="getCurrentPageContent()?.annotations?.length > 0">{{ getCurrentPageContent()?.annotations.length }} annotations</p>
                  {{ getCurrentPageContent()?.content?.length || 0 }} characters
                </div>
              </div>

              <!-- Text editor -->
              <div class="flex-1 py-2 relative">
                <textarea v-if="getCurrentPageContent()" :value="getCurrentPageContent()?.content"
                  @input="updatePageContent(($event.target as HTMLTextAreaElement).value)"
                  @contextmenu="handleRightClick" @keydown="handleTextareaKeydown"
                  class="w-full outline-none h-full resize-none bg-white text-gray-900 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100"
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
            <div v-else-if="editorMode === 'preview'"
              class="bg-white mx-auto dark:bg-gray-800 h-full flex flex-col">
              <!-- Page header -->
              <div class="py-2 w-full flex items-center justify-between dark:border-gray-600">
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
                <div v-if="getCurrentPageContent()" class="prose prose-gray dark:prose-invert max-w-none"
                  v-html="renderedMarkdown"></div>
                <div v-else class="text-center py-8 text-gray-500 dark:text-gray-400">
                  <p>No content to preview</p>
                </div>
              </div>
            </div>

            <!-- Split Mode -->
            <div v-else-if="editorMode === 'split'"
              class="bg-white dark:bg-gray-800 h-full flex flex-col">
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
              <div class="flex-1 flex">
                <!-- Editor side -->
                <div class="flex-1 p-4 border-r border-gray-300 dark:border-gray-600">
                  <div class="mb-2">
                    <span class="text-sm font-medium text-gray-600 dark:text-gray-400">Editor</span>
                  </div>
                  <textarea v-if="getCurrentPageContent()" :value="getCurrentPageContent()?.content"
                    @input="updatePageContent(($event.target as HTMLTextAreaElement).value)"
                    @contextmenu="handleRightClick" @keydown="handleTextareaKeydown"
                    class="w-full h-full p-4 outline-none border border-gray-300 rounded-lg resize-none focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white text-gray-900 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100"
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
                    class="h-full border border-gray-300 rounded-lg p-4 overflow-auto bg-gray-50 dark:border-gray-600 dark:bg-gray-700/50">
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
              <p class="text-sm text-gray-600 dark:text-gray-400">This PDF may contain only images or have no
                extractable text.</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- AI Suggestions Popup -->
    <div v-if="showAISuggestions"
      class="ai-suggestions fixed z-50 bg-white border border-gray-300 rounded-lg shadow-xl dark:bg-gray-800 dark:border-gray-600"
      :style="{
        left: aiSuggestionsPosition.x + 'px',
        top: aiSuggestionsPosition.y + 'px'
      }">
      <div class="p-2">
        <div class="text-xs font-medium text-gray-500 px-2 py-1 dark:text-gray-400">AI Assistant</div>
        <div class="space-y-1">
          <button v-for="suggestion in aiSuggestions" :key="suggestion.action" @click="selectAISuggestion(suggestion)"
            class="w-full px-3 py-2 text-left hover:bg-gray-100 rounded flex items-center gap-3 text-sm text-gray-800 dark:hover:bg-gray-700 dark:text-gray-300 transition-colors">
            <span class="text-lg">{{ suggestion.icon }}</span>
            <div class="flex-1">
              <div class="font-medium">{{ suggestion.label }}</div>
              <div class="text-xs text-gray-500 dark:text-gray-400">{{ suggestion.description }}</div>
            </div>
            <div class="text-xs text-gray-400 dark:text-gray-500">{{ suggestion.shortcut }}</div>
          </button>
        </div>
      </div>
    </div>

    <!-- Modals -->

    <!-- Create New Document Modal -->
    <div v-if="showCreateModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4"
      @click.self="showCreateModal = false">
      <div class="bg-white rounded-lg max-w-md w-full border border-gray-300 dark:bg-gray-800 dark:border-gray-600">
        <div class="p-4 border-b border-gray-300 flex items-center justify-between dark:border-gray-600">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Create New Document</h3>
          <button @click="showCreateModal = false"
            class="w-8 h-8 rounded-full bg-gray-100 hover:bg-gray-200 flex items-center justify-center transition-colors dark:bg-gray-700 dark:hover:bg-gray-600">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div class="p-4">
          <div class="mb-4">
            <label class="block text-sm font-medium text-gray-800 dark:text-gray-300 mb-2">
              Document Title
            </label>
            <input type="text" v-model="newDocTitle" placeholder="My New Document"
              class="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white text-gray-900 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100" />
          </div>
          <div class="mb-4">
            <label class="block text-sm font-medium text-gray-800 dark:text-gray-300 mb-2">
              Initial Content (Optional)
            </label>
            <textarea v-model="newDocContent" placeholder="Start with some markdown content..." rows="4"
              class="w-full px-3 py-2 border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white text-gray-900 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 resize-none"></textarea>
          </div>
        </div>
        <div class="p-4 border-t border-gray-300 flex gap-2 justify-end dark:border-gray-600">
          <button @click="showCreateModal = false"
            class="px-4 py-2 text-gray-600 hover:text-gray-800 transition-colors dark:text-gray-400 dark:hover:text-gray-300">
            Cancel
          </button>
          <button @click="createNewDocument" :disabled="!newDocTitle.trim()"
            class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors disabled:opacity-50">
            Create Document
          </button>
        </div>
      </div>
    </div>

    <!-- AI Modal -->
    <div v-if="showAIModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4"
      @click.self="closeAIModal">
      <div class="bg-white rounded-lg max-w-2xl w-full border border-gray-300 dark:bg-gray-800 dark:border-gray-600">
        <div class="p-4 border-b border-gray-300 flex items-center justify-between dark:border-gray-600">
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
        <div class="p-6">
          <div class="mb-4">
            <h4 class="text-sm font-medium text-gray-800 dark:text-gray-300 mb-2">Original Text:</h4>
            <div
              class="p-3 bg-gray-50 rounded text-sm text-gray-700 dark:bg-gray-700 dark:text-gray-300 max-h-32 overflow-y-auto">
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
        <div class="p-4 border-t border-gray-300 flex gap-2 justify-end dark:border-gray-600">
          <button @click="discardAIContent"
            class="px-4 py-2 text-gray-600 hover:text-gray-800 transition-colors dark:text-gray-400 dark:hover:text-gray-300">
            Discard
          </button>
          <button @click="saveAIContent" :disabled="isGeneratingAI || !aiContent.trim()"
            class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors disabled:opacity-50">
            Apply Changes
          </button>
        </div>
      </div>
    </div>

    <!-- Preview Modal -->
    <div v-if="showPreview" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4"
      @click.self="closePreview">
      <div
        class="bg-white rounded-lg max-w-4xl w-full h-5/6 border border-gray-300 dark:bg-gray-800 dark:border-gray-600 flex flex-col">
        <div class="p-4 border-b border-gray-300 flex items-center justify-between dark:border-gray-600">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Document Preview</h3>
          <button @click="closePreview"
            class="w-8 h-8 rounded-full bg-gray-100 hover:bg-gray-200 flex items-center justify-center transition-colors dark:bg-gray-700 dark:hover:bg-gray-600">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div class="flex-1 p-6 overflow-auto">
          <pre
            class="whitespace-pre-wrap text-sm text-gray-800 dark:text-gray-200 font-mono">{{ previewContent }}</pre>
        </div>
      </div>
    </div>

    <!-- Context Menu -->
    <div v-if="showContextMenu" :style="{
      position: 'fixed',
      left: contextMenuPosition.x + 'px',
      top: contextMenuPosition.y + 'px',
      zIndex: 1001
    }" class="bg-white border border-gray-300 rounded-lg shadow-lg py-2 min-w-48 dark:bg-gray-800 dark:border-gray-600"
      @click.stop>
      <div
        class="px-3 py-2 text-xs font-medium text-gray-500 border-b border-gray-300 dark:text-gray-400 dark:border-gray-600">
        AI Actions
      </div>
      <button v-for="suggestion in aiSuggestions.slice(0, 3)" :key="suggestion.action"
        @click="contextMenuAction(suggestion.action)"
        class="w-full px-3 py-2 text-left hover:bg-gray-100 flex items-center gap-2 text-sm text-gray-800 dark:hover:bg-gray-700 dark:text-gray-300">
        <span class="text-base">{{ suggestion.icon }}</span>
        <span>{{ suggestion.label }}</span>
      </button>

      <div class="border-t border-gray-300 my-1 dark:border-gray-600"></div>

      <button @click="contextMenuAction('copy')"
        class="w-full px-3 py-2 text-left hover:bg-gray-100 flex items-center gap-2 text-sm text-gray-800 dark:hover:bg-gray-700 dark:text-gray-300">
        <svg class="w-4 h-4 text-blue-600 dark:text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
        </svg>
        <span>Copy</span>
      </button>

      <button @click="contextMenuAction('highlight')"
        class="w-full px-3 py-2 text-left hover:bg-gray-100 flex items-center gap-2 text-sm text-gray-800 dark:hover:bg-gray-700 dark:text-gray-300">
        <svg class="w-4 h-4 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zM7 3V1m0 20v-2m8-10a4 4 0 00-4-4V3a2 2 0 012-2h4a2 2 0 012 2v2a4 4 0 00-4 4z" />
        </svg>
        <span>Highlight</span>
      </button>

      <button @click="contextMenuAction('note')"
        class="w-full px-3 py-2 text-left hover:bg-gray-100 flex items-center gap-2 text-sm text-gray-800 dark:hover:bg-gray-700 dark:text-gray-300">
        <svg class="w-4 h-4 text-orange-600 dark:text-orange-400" fill="none" stroke="currentColor"
          viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
        </svg>
        <span>Add Note</span>
      </button>
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
  width: 8px;
  height: 8px;
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
</style>