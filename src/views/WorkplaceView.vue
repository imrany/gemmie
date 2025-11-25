<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, type Ref } from 'vue'
import { useRouter } from 'vue-router'
import { inject } from 'vue'
import * as pdfjsLib from 'pdfjs-dist'
import pdfjsWorker from 'pdfjs-dist/build/pdf.worker?url'

// Configure PDF.js worker
pdfjsLib.GlobalWorkerOptions.workerSrc = pdfjsWorker

// Utils and Types
import { renderMarkdown } from '@/lib/markdownSupport'
import { formatFileSize } from '@/lib/formatters'
import type { EditableContent, UploadedFile } from '@/types/document'

// Composables
import { useDocumentStore } from '@/composables/useDocumentStore'
import { useMarkdownEditor } from '@/composables/useMarkdownEditor'
import { useHistory } from '@/composables/useHistory'
import { useKeyboardShortcuts } from '@/composables/useKeyboardShortcuts'
import { useAI } from '@/composables/useAI'
import { useSearch } from '@/composables/useSearch'
import { useFileUpload } from '@/composables/useFileUpload'
import { useDocumentTemplates } from '@/composables/useDocumentTemplates'
import { useExport } from '@/composables/useExport'
import { extractPdfContent } from '@/lib/pdfHelpers'
import type { UserDetails } from '@/types'

// Global state
const globalState = inject('globalState') as {
  screenWidth: Ref<number>,
  toggleTheme: () => void
  isDarkMode: Ref<boolean>,
  parsedUserDetails: Ref<UserDetails>
}
const {
  screenWidth,
  isDarkMode,
  toggleTheme,
  parsedUserDetails
} = globalState
const router = useRouter()

// Window dimensions
const innerWidth = window.innerWidth
const innerHeight = window.innerHeight

// ============================================
// CORE STATE
// ============================================
const selectedPdfUrl = ref('')
const selectedPdfName = ref('')
const currentPage = ref(1)
const totalPages = ref(0)
const editablePages = ref<EditableContent[]>([])
const isLoading = ref(false)
const loadError = ref<string>('')
const hasUnsavedChanges = ref(false)
const autoSaveTimer = ref<any | null>(null)

// ============================================
// UI STATE
// ============================================
const activeDocumentMenu = ref<string | null>(null)
const imageMenuOpen = ref(false)
const sidebarOpen = ref(true)
const activeSidebarTab = ref<'outline' | 'search' | 'annotations' | 'history' | 'documents' | string>('documents')
const sidebarWidth = ref(290)
const isResizingSidebar = ref(false)
const maxSidebarWidth = 350
const editorMode = ref<'edit' | 'preview' | 'split'>('edit')
const showMarkdownToolbar = ref(true)
const fontSize = ref(14)
const lineHeight = ref(1.5)
const selectedText = ref('')

// ============================================
// INITIALIZE COMPOSABLES
// ============================================

// Document store
const documentStore = useDocumentStore()
const { uploadedFiles, selectedFileId } = documentStore

// Helper function for current page content
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

// History
const history = useHistory(
  editablePages,
  currentPage,
  getCurrentPageContent,
  (content: string, saveHistory: boolean) => {
    const pageIndex = currentPage.value - 1
    if (pageIndex >= 0 && pageIndex < editablePages.value.length) {
      if (saveHistory && editablePages.value[pageIndex].content !== content) {
        history.saveToHistory()
      }
      editablePages.value[pageIndex].content = content
      editablePages.value[pageIndex].isModified =
        content !== editablePages.value[pageIndex].originalContent
    }
  }
)

const {
  editHistory
} = history

// Markdown editor
const markdownEditor = useMarkdownEditor(
  editablePages,
  currentPage,
  history.saveToHistory
)

// AI features
const ai = useAI()

// Search
const search = useSearch(editablePages)

// File upload
const fileUpload = useFileUpload(
  uploadedFiles,
  documentStore.addFile,
  openPdfEditor
)

// Templates
const templates = useDocumentTemplates(
  documentStore.addFile,
  openTextEditor
)

// Export
const exportUtils = useExport(
  editablePages,
  selectedPdfName,
  totalPages
)

// ============================================
// KEYBOARD SHORTCUTS
// ============================================
const shortcuts = useKeyboardShortcuts({
  undo: history.undo,
  redo: history.redo,
  insertBold: markdownEditor.insertBold,
  insertItalic: markdownEditor.insertItalic,
  insertStrikethrough: markdownEditor.insertStrikethrough,
  insertCode: markdownEditor.insertCode,
  insertLink: markdownEditor.insertLink,
  insertHeader: markdownEditor.insertHeader,
  insertList: markdownEditor.insertList,
  insertNumberedList: markdownEditor.insertNumberedList,
  insertTaskList: markdownEditor.insertTaskList,
  insertQuote: markdownEditor.insertQuote,
  insertCodeBlock: markdownEditor.insertCodeBlock,
  insertImage: markdownEditor.insertImage,
  insertHighlight: markdownEditor.insertHighlight,
  insertHorizontalRule: markdownEditor.insertHorizontalRule,
  insertTable: markdownEditor.insertTable,
  saveDocument: saveDocumentChanges,
  showAIToolbar: ai.showAIToolbar,
  handleAIShortcut: (action: string) => {
    const selection = window.getSelection()
    const selectedTextValue = selection?.toString().trim()
    if (selectedTextValue) {
      ai.performAIAction(action as any, selectedTextValue)
    } else {
      const currentPageData = getCurrentPageContent()
      if (currentPageData) {
        ai.performAIAction(action as any, currentPageData.content)
      }
    }
  }
})

// ============================================
// COMPUTED PROPERTIES
// ============================================
const { canUndo, canRedo } = history
const { hasSearchResults } = search

const allAnnotations = computed(() =>
  editablePages.value.flatMap(page =>
    page.annotations.map(ann => ({ ...ann, pageNum: page.pageNum }))
  )
)

const renderedMarkdown = computed(() => {
  const currentPageData = getCurrentPageContent()
  if (!currentPageData) return ''

  const rawHtml = renderMarkdown(currentPageData.content)
  return rawHtml
})

// ============================================
// EXPOSED FUNCTIONS (from composables)
// ============================================

// Markdown functions
const {
  insertBold,
  insertItalic,
  insertStrikethrough,
  insertHighlight,
  insertCode,
  insertCodeBlock,
  insertLink,
  insertImage,
  insertHeader,
  insertList,
  insertNumberedList,
  insertTaskList,
  insertQuote,
  insertTable,
  insertHorizontalRule,
  updatePageContent,
  insertImageWithTitle,
  insertImageWithDimensions,
  insertImageCentered,
  insertImageSmall,
  insertImageWithBorder,
  insertImageLink,
} = markdownEditor

// History functions
const { undo, redo, addToHistory } = history

// Search functions
const { searchQuery, searchResults, currentSearchIndex, performSearch } = search

// File upload
const { isDragOver, handleFileUpload, handleDragOver, handleDragLeave, handleDrop } = fileUpload

// Templates
const { documentTemplates, createDocumentFromTemplate } = templates

// Export
const {
  showPreview,
  previewContent,
  showExportDropdown,
  generatePreview,
  closePreview,
  downloadAsText,
  downloadAsMarkdown,
  toggleExportDropdown,
  closeExportDropdown
} = exportUtils

// AI
const {
  showAIModal,
  aiContent,
  originalTextForAI,
  isGeneratingAI,
  currentAIAction,
  showAISuggestions,
  aiSuggestionsPosition,
  showAIPopover,
  aiPopoverPosition,
  aiPopoverContent,
  isLoadingAIPopover,
  aiSuggestions,
  hideAISuggestions
} = ai

// ============================================
// AI MODAL FUNCTIONS
// ============================================
function saveAIContent() {
  const currentPageData = getCurrentPageContent()
  if (currentPageData && aiContent.value.trim()) {
    history.saveToHistory()
    const newContent = currentPageData.content.replace(
      originalTextForAI.value,
      aiContent.value.trim()
    )
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

// ============================================
// DOCUMENT MANAGEMENT
// ============================================
async function openPdfEditor(file: UploadedFile) {
  if (file.isCustom) {
    openTextEditor(file)
    return
  }

  selectedPdfUrl.value = file.url
  selectedPdfName.value = file.name
  selectedFileId.value = file.id
  documentStore.saveCurrentDocument()

  isLoading.value = true
  loadError.value = ''

  try {
    const pages = await extractPdfContent(file.url)
    editablePages.value = pages
    totalPages.value = pages.length
    currentPage.value = 1
    history.clearHistory()
    search.clearSearch()
  } catch (error: any) {
    loadError.value = `Failed to extract PDF content: ${error.message}`
  } finally {
    isLoading.value = false
  }
}

function openTextEditor(doc: UploadedFile) {
  selectedPdfUrl.value = doc.url
  selectedPdfName.value = doc.name
  selectedFileId.value = doc.id
  documentStore.saveCurrentDocument()

  const pageContent: EditableContent = {
    pageNum: 1,
    content: doc.content || `# Welcome to ${doc.name}\n\nStart writing...`,
    originalContent: doc.content || '',
    isModified: false,
    annotations: []
  }

  editablePages.value = [pageContent]
  currentPage.value = 1
  totalPages.value = 1
  history.clearHistory()
  search.clearSearch()
  hasUnsavedChanges.value = false
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
  history.clearHistory()
  search.clearSearch()
  hideAISuggestions()

  if (autoSaveTimer.value) {
    clearTimeout(autoSaveTimer.value)
    autoSaveTimer.value = null
  }

  documentStore.clearCurrentDocument()
}

function removeFile(id: string) {
  if (selectedFileId.value === id) {
    closeEditor()
  }
  documentStore.removeFile(id)
}

// ============================================
// PAGE MANAGEMENT
// ============================================
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

function goToSearchResult(index: number) {
  search.goToSearchResult(index, goToPage)
}

function resetPageContent() {
  const pageIndex = currentPage.value - 1
  if (pageIndex >= 0 && pageIndex < editablePages.value.length) {
    history.saveToHistory()
    editablePages.value[pageIndex].content = editablePages.value[pageIndex].originalContent
    editablePages.value[pageIndex].isModified = false
    editablePages.value[pageIndex].annotations = []
    addToHistory('Reset page', `Reset page ${currentPage.value} to original`)
  }
}

function resetAllContent() {
  history.saveToHistory()
  editablePages.value.forEach(page => {
    page.content = page.originalContent
    page.isModified = false
    page.annotations = []
  })
  hasUnsavedChanges.value = false
  history.clearHistory()
  addToHistory('Reset all', 'Reset all pages to original')
}

function getModifiedPagesCount(): number {
  return editablePages.value.filter(page => page.isModified).length
}

// ============================================
// AUTO-SAVE
// ============================================
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

  documentStore.saveToLocalStorage()
  hasUnsavedChanges.value = false

  editablePages.value.forEach(page => {
    page.originalContent = page.content
    page.isModified = false
  })
}

// ============================================
// EVENT HANDLERS
// ============================================
function handleTextareaKeydown(event: KeyboardEvent) {
  shortcuts.handleEditorKeydown(event)
}

function handleTextareaContextMenu(event: MouseEvent) {
  event.preventDefault()

  // const textarea = event.target as HTMLTextAreaElement
  const selection = window.getSelection()
  const selectedTextValue = selection?.toString().trim()

  // Use selected text if available, otherwise use all content
  const textToUse = selectedTextValue || getCurrentPageContent()?.content || ''

  if (!textToUse || textToUse.length < 3) {
    return // Don't show menu for empty or very short text
  }

  selectedText.value = textToUse

  // Position the popup near the mouse cursor
  const viewport = {
    width: window.innerWidth,
    height: window.innerHeight
  }

  const popupWidth = 280
  const popupHeight = 350

  let x = event.clientX
  let y = event.clientY + 10

  // Keep popup within viewport
  if (x + popupWidth > viewport.width - 10) {
    x = viewport.width - popupWidth - 10
  }
  if (x < 10) {
    x = 10
  }

  if (y + popupHeight > viewport.height - 10) {
    y = event.clientY - popupHeight - 10
  }
  if (y < 10) {
    y = 10
  }

  x = Math.max(10, Math.min(x, viewport.width - popupWidth - 10))
  y = Math.max(10, Math.min(y, viewport.height - popupHeight - 10))

  aiSuggestionsPosition.value = { x, y }
  showAISuggestions.value = true
}

function selectAISuggestion(suggestion: typeof aiSuggestions[0]) {
  hideAISuggestions()
  if (selectedText.value) {
    ai.performAIAction(suggestion.action as any, selectedText.value)
  }
}

function applyAIResult() {
  const currentPageData = getCurrentPageContent()
  if (currentPageData && aiPopoverContent.value.trim()) {
    history.saveToHistory()
    const newContent = currentPageData.content.replace(
      originalTextForAI.value,
      aiPopoverContent.value.trim()
    )
    updatePageContent(newContent)
    addToHistory(`AI ${currentAIAction.value}`, `Applied ${currentAIAction.value} to selected text`)
  }
  showAIPopover.value = false
  aiPopoverContent.value = ''
  originalTextForAI.value = ''
}

// ============================================
// SIDEBAR RESIZE
// ============================================
function startSidebarResize() {
  isResizingSidebar.value = true
  document.addEventListener('mousemove', handleSidebarResize)
  document.addEventListener('mouseup', stopSidebarResize)
}

function handleSidebarResize(event: MouseEvent) {
  if (!isResizingSidebar.value) return
  const newWidth = event.clientX
  sidebarWidth.value = Math.max(290, Math.min(newWidth, maxSidebarWidth))
}

function stopSidebarResize() {
  isResizingSidebar.value = false
}

// ============================================
// DRAG HANDLERS
// ============================================
const isDraggingAISuggestions = ref(false)
const isDraggingAIPopover = ref(false)
const dragOffset = ref({ x: 0, y: 0 })

function startSuggestionsDrag(event: MouseEvent) {
  event.preventDefault()
  event.stopPropagation()

  isDraggingAISuggestions.value = true

  const rect = (event.currentTarget as HTMLElement).closest('.ai-suggestions')?.getBoundingClientRect()
  if (rect) {
    dragOffset.value = {
      x: event.clientX - rect.left,
      y: event.clientY - rect.top
    }
  }

  document.addEventListener('mousemove', handleSuggestionsDrag)
  document.addEventListener('mouseup', stopSuggestionsDrag)
}

function handleSuggestionsDrag(event: MouseEvent) {
  if (!isDraggingAISuggestions.value) return

  event.preventDefault()

  let newX = event.clientX - dragOffset.value.x
  let newY = event.clientY - dragOffset.value.y

  const viewport = {
    width: window.innerWidth,
    height: window.innerHeight
  }

  newX = Math.max(10, Math.min(newX, viewport.width - 290))
  newY = Math.max(10, Math.min(newY, viewport.height - 250))

  aiSuggestionsPosition.value = { x: newX, y: newY }
}

function stopSuggestionsDrag() {
  isDraggingAISuggestions.value = false
  document.removeEventListener('mousemove', handleSuggestionsDrag)
  document.removeEventListener('mouseup', stopSuggestionsDrag)
}

function startDrag(event: MouseEvent, component: string) {
  if (component !== 'popover') return

  const target = event.target as HTMLElement
  const isDragHandle = target.classList.contains('drag-handle') || target.closest('.drag-handle')

  if (!isDragHandle) return

  event.preventDefault()
  event.stopPropagation()

  isDraggingAIPopover.value = true

  dragOffset.value = {
    x: event.clientX - aiPopoverPosition.value.x,
    y: event.clientY - aiPopoverPosition.value.y
  }

  const handleMouseMove = (e: MouseEvent) => {
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

  // Close document menu
  if (activeDocumentMenu.value) {
    const target = event.target as HTMLElement
    if (!target.closest('.document-menu')) {
      closeDocumentMenu()
    }
  }
}

function handlePreview() {
  closeExportDropdown()
  generatePreview()
}

function getImageDropdownPosition(): { left: number, top: number } {
  const imageButton = document.querySelector('.image-dropdown-button') as HTMLElement;

  if (!imageButton) {
    // Fallback positioning
    return {
      left: Math.min(window.innerWidth - 300, 100),
      top: 150
    };
  }

  const rect = imageButton.getBoundingClientRect();
  return {
    left: rect.left,
    top: rect.bottom + window.scrollY
  };
}

function toggleDocumentMenu(fileId: string, event: Event) {
  event.stopPropagation()
  activeDocumentMenu.value = activeDocumentMenu.value === fileId ? null : fileId
}

function closeDocumentMenu() {
  activeDocumentMenu.value = null
}

function renameDocument(file: UploadedFile) {
  const newName = prompt('Enter new name:', file.name)
  if (newName && newName.trim()) {
    file.name = newName.trim()
    documentStore.saveToLocalStorage()
  }
  closeDocumentMenu()
}

function duplicateDocument(file: UploadedFile) {
  const duplicatedFile: UploadedFile = {
    ...file,
    id: Date.now().toString() + Math.random(),
    name: `${file.name} (Copy)`,
    uploadedAt: new Date()
  }
  documentStore.addFile(duplicatedFile)
  closeDocumentMenu()
}

function getThemeTitle(): string {
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches

  if (parsedUserDetails.value.theme === 'system') {
    return `System Theme (${prefersDark ? 'Dark' : 'Light'})`
  } else if (parsedUserDetails.value.theme === 'light') {
    return 'Light Theme'
  } else {
    return 'Dark Theme'
  }
}

// ============================================
// WATCHERS
// ============================================
watch([editablePages], () => {
  hasUnsavedChanges.value = editablePages.value.some(page => page.isModified)
  if (hasUnsavedChanges.value) {
    scheduleAutoSave()
  }
}, { deep: true })

watch(()=> screenWidth.value, (newValue)=>{
  console.log(newValue)
  if (newValue < 768) {
    router.push("/")
  }
})

// ============================================
// LIFECYCLE
// ============================================
onMounted(() => {
  if (screenWidth.value < 768) {
    router.push("/")
  }

  documentStore.loadFromLocalStorage()

  const currentDocId = documentStore.getCurrentDocumentId()
  if (currentDocId) {
    const lastOpenedDoc = uploadedFiles.value.find(file => file.id === currentDocId)
    if (lastOpenedDoc) {
      setTimeout(() => openPdfEditor(lastOpenedDoc), 100)
    }
  }

  document.addEventListener('click', handleGlobalClick)
  document.addEventListener('mouseup', stopSidebarResize)
})

onUnmounted(() => {
  if (hasUnsavedChanges.value) {
    saveDocumentChanges()
  }

  document.removeEventListener('click', handleGlobalClick)
  document.removeEventListener('mouseup', stopSidebarResize)
  document.removeEventListener('mousemove', handleSuggestionsDrag)
  document.removeEventListener('mouseup', stopSuggestionsDrag)

  if (autoSaveTimer.value) {
    clearTimeout(autoSaveTimer.value)
  }
})
</script>

<template>
  <div @dblclick="imageMenuOpen = false"
    class="w-full bg-gray-50 text-gray-900 min-h-screen dark:bg-gray-900 dark:text-gray-100">
    <!-- Main Editor Interface -->
    <div class="bg-white h-screen w-full flex flex-col lg:flex-row dark:bg-gray-800">
      <!-- Mobile Header (visible on mobile only) -->
      <div class="lg:hidden bg-gray-100 border-b border-gray-300 p-3 dark:bg-gray-700 dark:border-gray-600">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2">
            <button @click="sidebarOpen = !sidebarOpen" class="p-2 rounded hover:bg-gray-200 dark:hover:bg-gray-600">
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
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
              </svg>
              <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
              </svg>
            </button>

            <div class="relative">
              <button @click="toggleExportDropdown" class="p-2 rounded hover:bg-gray-200 dark:hover:bg-gray-600">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
                </svg>
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Sidebar Overlay (mobile) -->
      <div v-if="sidebarOpen && innerWidth < 1024" class="lg:hidden fixed inset-0 z-40 bg-black bg-opacity-50"
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
                'p-3 rounded cursor-pointer transition-colors text-sm border relative',
                selectedFileId === file.id
                  ? 'bg-blue-50 text-blue-600 border-blue-500 dark:bg-blue-900/20 dark:text-blue-400 dark:border-blue-400'
                  : 'hover:bg-gray-50 border-gray-300 dark:hover:bg-gray-600 dark:border-gray-600'
              ]">
                <div class="flex items-start gap-3">
                  <svg v-if="file.isCustom" class="w-5 h-5 text-blue-500 flex-shrink-0 mt-0.5 dark:text-blue-400"
                    fill="none" stroke="currentColor" viewBox="0 0 24 24">
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

                  <!-- More Options Button -->
                  <div class="document-menu relative flex-shrink-0">
                    <button @click.stop="toggleDocumentMenu(file.id, $event)"
                      class="w-8 h-8 rounded-full hover:bg-gray-200 dark:hover:bg-gray-600 flex items-center justify-center text-gray-600 dark:text-gray-400 transition-colors"
                      :class="{ 'bg-gray-200 dark:bg-gray-600': activeDocumentMenu === file.id }" title="More options">
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                          d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
                      </svg>
                    </button>

                    <!-- Dropdown Menu -->
                    <div v-if="activeDocumentMenu === file.id"
                      class="absolute right-0 top-full mt-1 bg-white dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-lg shadow-lg py-1 min-w-48 z-50"
                      @click.stop>

                      <button @click.stop="renameDocument(file)"
                        class="w-full px-4 py-2 text-left hover:bg-gray-100 dark:hover:bg-gray-600 flex items-center gap-3 text-sm text-gray-800 dark:text-gray-200 transition-colors">
                        <svg class="w-4 h-4 text-gray-600 dark:text-gray-400" fill="none" stroke="currentColor"
                          viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                            d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                        </svg>
                        <span>Rename</span>
                      </button>

                      <button @click.stop="duplicateDocument(file)"
                        class="w-full px-4 py-2 text-left hover:bg-gray-100 dark:hover:bg-gray-600 flex items-center gap-3 text-sm text-gray-800 dark:text-gray-200 transition-colors">
                        <svg class="w-4 h-4 text-gray-600 dark:text-gray-400" fill="none" stroke="currentColor"
                          viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                            d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                        </svg>
                        <span>Duplicate</span>
                      </button>

                      <div class="border-t border-gray-200 dark:border-gray-600"></div>

                      <button @click.stop="removeFile(file.id); closeDocumentMenu()"
                        class="w-full px-4 py-2 text-left hover:bg-red-50 dark:hover:bg-red-900/20 flex items-center gap-3 text-sm text-red-600 dark:text-red-400 transition-colors">
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                            d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                        </svg>
                        <span>Delete</span>
                      </button>
                    </div>
                  </div>
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
                    <span v-if="page.annotations.length > 0"
                      class="text-xs text-blue-600 dark:text-blue-400 px-1 py-0.5 bg-blue-100 dark:bg-blue-900/20 rounded"
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
            <h4 class="text-sm font-medium text-gray-800 dark:text-gray-300 mb-3">Annotations ({{ allAnnotations.length
            }})</h4>
            <div class="space-y-2">
              <div v-for="ann in allAnnotations" :key="ann.id" @click="goToPage(ann.pageNum)"
                class="p-3 rounded border cursor-pointer hover:bg-gray-50 text-xs dark:hover:bg-gray-600"
                :style="{ borderLeftColor: ann.color, borderLeftWidth: '3px' }">
                <div class="flex items-center justify-between mb-1">
                  <span class="font-medium capitalize text-gray-800 dark:text-gray-300">{{ ann.type }}</span>
                  <span class="text-gray-500 dark:text-gray-400">Page {{ ann.pageNum }}</span>
                </div>
                <div class="text-gray-800 dark:text-gray-300 mb-1 break-words">"{{ ann.text.substring(0, 60) }}..."
                </div>
                <div v-if="ann.note" class="text-gray-500 dark:text-gray-400 break-words">{{ ann.note }}</div>
                <div class="text-gray-500 dark:text-gray-400 text-xs mt-1">
                  {{ ann.timestamp.toLocaleDateString() }}
                </div>
              </div>
              <div v-if="allAnnotations.length === 0" class="text-gray-500 text-center py-8 dark:text-gray-400">
                <svg class="w-8 h-8 text-gray-300 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z" />
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
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
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
        <div
          class="hidden lg:flex flex-col border-b border-gray-300 bg-gray-50 dark:border-gray-600 dark:bg-gray-700/50">
          <!-- Title bar -->
          <div class="flex items-center justify-between p-3 xl:p-4 border-b border-gray-300 dark:border-gray-600">
            <div class="flex items-center gap-3 min-w-0 flex-1">
              <img
                :src="parsedUserDetails?.theme === 'dark' || (parsedUserDetails?.theme === 'system' && isDarkMode) ? '/favicon-light.svg' : '/favicon.svg'"
                alt="Gemmie Logo" class="w-8 h-8 rounded-md bg-gray-50 dark:bg-gray-700/50" />

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

              <!-- Theme Toggle Button -->
              <button @click="toggleTheme"
                class="w-8 xl:w-10 h-8 xl:h-10 rounded-md bg-white hover:bg-gray-50 transition-colors text-gray-700 flex items-center justify-center border border-gray-200 dark:bg-gray-700 dark:hover:bg-gray-600 dark:text-gray-300 dark:border-gray-600"
                :title="getThemeTitle()">
                <!-- System icon (shown when theme is system) -->
                <svg v-if="parsedUserDetails?.theme === 'system'" class="w-4 xl:w-5 h-4 xl:h-5" fill="none" stroke="currentColor"
                  viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
                <!-- Sun icon (shown in light mode) -->
                <svg v-else-if="parsedUserDetails?.theme=== 'light'" class="w-4 xl:w-5 h-4 xl:h-5" fill="none"
                  stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
                </svg>
                <!-- Moon icon (shown in dark mode) -->
                <svg v-else class="w-4 xl:w-5 h-4 xl:h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
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
              <button @click="insertStrikethrough"
                class="w-7 h-7 xl:w-8 xl:h-8 rounded bg-white hover:bg-gray-50 flex items-center justify-center border border-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                title="Strike through">
                <svg class="w-3 h-3 xl:w-4 xl:h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
                </svg>
              </button>
              <button @click="insertHighlight"
                class="w-7 h-7 xl:w-8 xl:h-8 rounded bg-white hover:bg-gray-50 flex items-center justify-center border border-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                title="Highlight">
                <svg class="w-3 h-3 xl:w-4 xl:h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                </svg>
              </button>
            </div>

            <!-- Code & Blocks -->
            <div class="flex items-center gap-1 border-r border-gray-300 pr-2 mr-2 dark:border-gray-500">
              <button @click="insertCode"
                class="w-7 h-7 xl:w-8 xl:h-8 rounded bg-white hover:bg-gray-50 flex items-center justify-center border border-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                title="Inline Code">
                <svg class="w-3 h-3 xl:w-4 xl:h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
                </svg>
              </button>
              <button @click="insertCodeBlock"
                class="w-7 h-7 xl:w-8 xl:h-8 rounded bg-white hover:bg-gray-50 flex items-center justify-center border border-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                title="Code Block">
                <svg class="w-3 h-3 xl:w-4 xl:h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
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

            <!-- Lists -->
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
              <button @click="insertTaskList"
                class="w-7 h-7 xl:w-8 xl:h-8 rounded bg-white hover:bg-gray-50 flex items-center justify-center border border-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                title="Task List">
                <svg class="w-3 h-3 xl:w-4 xl:h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                </svg>
              </button>
            </div>

            <!-- Links & Media -->
            <div class="flex items-center gap-1 border-r border-gray-300 pr-2 mr-2 dark:border-gray-500">
              <button @click="insertLink"
                class="w-7 h-7 xl:w-8 xl:h-8 rounded bg-white hover:bg-gray-50 flex items-center justify-center border border-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                title="Link (Ctrl+K)">
                <svg class="w-3 h-3 xl:w-4 xl:h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
                </svg>
              </button>

              <!-- Enhanced Image Dropdown Menu -->
              <div class="relative z-50">
                <button @click="imageMenuOpen = !imageMenuOpen"
                  class="image-dropdown-button w-7 h-7 xl:w-8 xl:h-8 rounded bg-white hover:bg-gray-50 flex items-center justify-center border border-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600 transition-colors"
                  :class="{ 'bg-blue-50 border-blue-300 dark:bg-blue-900/30 dark:border-blue-500': imageMenuOpen }"
                  title="Insert Image with Options">
                  <svg class="w-3 h-3 xl:w-4 xl:h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                      d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                  </svg>
                </button>

                <!-- Dropdown Menu -->
                <div v-if="imageMenuOpen" @click.stop="imageMenuOpen = false"
                  class="fixed w-72 bg-white dark:bg-gray-800 rounded-lg shadow-xl border border-gray-200 dark:border-gray-700 z-[1000] overflow-hidden backdrop-blur-sm"
                  :style="{
                    left: getImageDropdownPosition().left + 'px',
                    top: getImageDropdownPosition().top + 'px'
                  }">

                  <!-- Header -->
                  <div class="px-4 py-3 border-b border-gray-100 dark:border-gray-600 bg-gray-50 dark:bg-gray-700/50">
                    <div class="flex items-center justify-between">
                      <h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100">Insert Image</h3>
                      <button @click="imageMenuOpen = false"
                        class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300">
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                            d="M6 18L18 6M6 6l12 12" />
                        </svg>
                      </button>
                    </div>
                    <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">Choose image format</p>
                  </div>

                  <!-- Image Options -->
                  <div class="max-h-96 overflow-y-auto">
                    <!-- Basic Image -->
                    <button @click="insertImage(); imageMenuOpen = false"
                      class="w-full px-3 py-2 text-left hover:bg-gray-50 dark:hover:bg-gray-900/20 border-b border-gray-100 dark:border-gray-600 flex items-center gap-3 group transition-colors">
                      <div
                        class="w-8 h-8 bg-blue-100 dark:bg-blue-900 rounded-lg flex items-center justify-center group-hover:scale-110 transition-transform">
                        <svg class="w-5 h-5 text-blue-600 dark:text-blue-400" fill="none" stroke="currentColor"
                          viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                            d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                        </svg>
                      </div>
                      <div class="flex-1 min-w-0">
                        <div
                          class="font-medium text-gray-900 dark:text-gray-100 text-sm group-hover:text-blue-600 dark:group-hover:text-blue-400">
                          Basic Image</div>
                        <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">![alt](url)</div>
                      </div>
                      <div
                        class="text-xs text-gray-500 px-2 py-1 bg-gray-100 group-hover:bg-gray-200 dark:bg-gray-700 rounded">
                        Default</div>
                    </button>

                    <!-- Image with Title -->
                    <button @click="insertImageWithTitle(); imageMenuOpen = false"
                      class="w-full px-3 py-2 text-left hover:bg-gray-50 dark:hover:bg-gray-900/20 border-b border-gray-100 dark:border-gray-600 flex items-center gap-3 group transition-colors">
                      <div
                        class="w-8 h-8 bg-green-100 dark:bg-green-900 rounded-lg flex items-center justify-center group-hover:scale-110 transition-transform">
                        <svg class="w-5 h-5 text-green-600 dark:text-green-400" fill="none" stroke="currentColor"
                          viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                            d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                      </div>
                      <div class="flex-1 min-w-0">
                        <div
                          class="font-medium text-gray-900 dark:text-gray-100 text-sm group-hover:text-green-600 dark:group-hover:text-green-400">
                          Image with Title</div>
                        <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">![alt](url "title")</div>
                      </div>
                    </button>

                    <!-- Image with Dimensions -->
                    <button @click="insertImageWithDimensions(); imageMenuOpen = false"
                      class="w-full px-3 py-2 text-left hover:bg-gray-50 dark:hover:bg-gray-900/20 border-b border-gray-100 dark:border-gray-600 flex items-center gap-3 group transition-colors">
                      <div
                        class="w-8 h-8 bg-purple-100 dark:bg-purple-900 rounded-lg flex items-center justify-center group-hover:scale-110 transition-transform">
                        <svg class="w-5 h-5 text-purple-600 dark:text-purple-400" fill="none" stroke="currentColor"
                          viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                            d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5v-4m0 4h-4m4 0l-5-5" />
                        </svg>
                      </div>
                      <div class="flex-1 min-w-0">
                        <div
                          class="font-medium text-gray-900 dark:text-gray-100 text-sm group-hover:text-purple-600 dark:group-hover:text-purple-400">
                          Image with Size</div>
                        <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">![alt](url =WxH)</div>
                      </div>
                    </button>

                    <!-- Centered Image -->
                    <button @click="insertImageCentered(); imageMenuOpen = false"
                      class="w-full px-3 py-2 text-left hover:bg-gray-50 dark:hover:bg-gray-900/20 border-b border-gray-100 dark:border-gray-600 flex items-center gap-3 group transition-colors">
                      <div
                        class="w-8 h-8 bg-yellow-100 dark:bg-yellow-900 rounded-lg flex items-center justify-center group-hover:scale-110 transition-transform">
                        <svg class="w-5 h-5 text-yellow-600 dark:text-yellow-400" fill="none" stroke="currentColor"
                          viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14" />
                        </svg>
                      </div>
                      <div class="flex-1 min-w-0">
                        <div
                          class="font-medium text-gray-900 dark:text-gray-100 text-sm group-hover:text-yellow-600 dark:group-hover:text-yellow-400">
                          Centered Image</div>
                        <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">![alt](url){.center}</div>
                      </div>
                    </button>

                    <!-- Small Image -->
                    <button @click="insertImageSmall(); imageMenuOpen = false"
                      class="w-full px-3 py-2 text-left hover:bg-gray-50 dark:hover:bg-gray-900/20 border-b border-gray-100 dark:border-gray-600 flex items-center gap-3 group transition-colors">
                      <div
                        class="w-8 h-8 bg-red-100 dark:bg-red-900 rounded-lg flex items-center justify-center group-hover:scale-110 transition-transform">
                        <svg class="w-5 h-5 text-red-600 dark:text-red-400" fill="none" stroke="currentColor"
                          viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                        </svg>
                      </div>
                      <div class="flex-1 min-w-0">
                        <div
                          class="font-medium text-gray-900 dark:text-gray-100 text-sm group-hover:text-red-600 dark:group-hover:text-red-400">
                          Small Image</div>
                        <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">![alt](url){.small}</div>
                      </div>
                    </button>

                    <!-- Image with Border -->
                    <button @click="insertImageWithBorder(); imageMenuOpen = false"
                      class="w-full px-3 py-2 text-left hover:bg-gray-50 dark:hover:bg-gray-900/20 border-b border-gray-100 dark:border-gray-600 flex items-center gap-3 group transition-colors">
                      <div
                        class="w-8 h-8 bg-indigo-100 dark:bg-indigo-900 rounded-lg flex items-center justify-center group-hover:scale-110 transition-transform">
                        <svg class="w-5 h-5 text-indigo-600 dark:text-indigo-400" fill="none" stroke="currentColor"
                          viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                            d="M4 6a2 2 0 012-2h12a2 2 0 012 2v12a2 2 0 01-2 2H6a2 2 0 01-2-2V6z" />
                        </svg>
                      </div>
                      <div class="flex-1 min-w-0">
                        <div
                          class="font-medium text-gray-900 dark:text-gray-100 text-sm group-hover:text-indigo-600 dark:group-hover:text-indigo-400">
                          Image with Border</div>
                        <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">![alt](url){.border}</div>
                      </div>
                    </button>

                    <!-- Image Link -->
                    <button @click="insertImageLink(); imageMenuOpen = false"
                      class="w-full px-3 py-2 text-left hover:bg-gray-50 dark:hover:bg-gray-900/20 flex items-center gap-3 group transition-colors">
                      <div
                        class="w-8 h-8 bg-blue-100 dark:bg-blue-900 rounded-lg flex items-center justify-center group-hover:scale-110 transition-transform">
                        <svg class="w-5 h-5 text-blue-600 dark:text-blue-400" fill="none" stroke="currentColor"
                          viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                            d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
                        </svg>
                      </div>
                      <div class="flex-1 min-w-0">
                        <div
                          class="font-medium text-gray-900 dark:text-gray-100 text-sm group-hover:text-blue-600 dark:group-hover:text-blue-400">
                          Clickable Image</div>
                        <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">[![alt](url)](link)</div>
                      </div>
                    </button>
                  </div>

                  <!-- Footer -->
                  <div class="px-3 py-2 border-t border-gray-100 dark:border-gray-600 bg-gray-50 dark:bg-gray-700/50">
                    <p class="text-xs text-gray-500 dark:text-gray-400 text-center">
                      Select text before inserting to use as alt text
                    </p>
                  </div>
                </div>
              </div>
            </div>

            <!-- Special Elements -->
            <div class="flex items-center gap-1">
              <button @click="insertQuote"
                class="w-7 h-7 xl:w-8 xl:h-8 rounded bg-white hover:bg-gray-50 flex items-center justify-center border border-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                title="Blockquote">
                <svg class="w-3 h-3 xl:w-4 xl:h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                </svg>
              </button>
              <button @click="insertTable"
                class="hidden sm:flex w-7 h-7 xl:w-8 xl:h-8 rounded bg-white hover:bg-gray-50 items-center justify-center border border-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                title="Insert Table">
                <svg class="w-3 h-3 xl:w-4 xl:h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M3 10h18M3 14h18M10 3v18M14 3v18" />
                </svg>
              </button>
              <button @click="insertHorizontalRule"
                class="w-7 h-7 xl:w-8 xl:h-8 rounded bg-white hover:bg-gray-50 flex items-center justify-center border border-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
                title="Horizontal Rule">
                <svg class="w-3 h-3 xl:w-4 xl:h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14" />
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
          <div v-if="selectedPdfName"
            class="flex flex-col sm:flex-row sm:items-center gap-3 p-3 bg-gray-100/50 dark:bg-gray-600/50">
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
            <div
              class="flex flex-wrap items-center gap-2 border-r border-gray-300 pr-3 sm:border-r-0 sm:pr-0 dark:border-gray-500">
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
                <h2 class="text-xl sm:text-2xl font-bold text-gray-900 dark:text-gray-100 mb-2">Welcome to Gemmie Editor
                </h2>
                <p class="text-base text-gray-600 dark:text-gray-400 mb-6 sm:mb-8 px-4">Choose a template to get started
                  with your new document</p>
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
              <div
                class="lg:hidden py-2 w-full flex items-center justify-between border-b border-gray-200 dark:border-gray-600">
                <div class="flex items-center gap-2">
                  <span class="text-sm font-medium text-gray-700 dark:text-gray-400">
                    {{ totalPages > 1 ? `Page ${currentPage}` : 'Document' }}
                  </span>
                  <span v-if="getCurrentPageContent()?.isModified" class="w-2 h-2 bg-orange-500 rounded-full"
                    title="Modified"></span>
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
                  @keydown="handleTextareaKeydown" @contextmenu="handleTextareaContextMenu"
                  class="w-full outline-none h-full px-3 sm:px-6 overflow-y-auto resize-none bg-white dark:bg-inherit text-gray-900 dark:text-gray-100"
                  :style="{
                    fontSize: fontSize + 'px',
                    lineHeight: lineHeight.toString(),
                    fontFamily: 'system-ui, -apple-system, sans-serif'
                  }" placeholder="Start writing your markdown content here...

Right-click for AI assistance
Use Ctrl+B for bold, Ctrl+I for italic
Press Ctrl+Z to undo, Ctrl+Y to redo">
                </textarea>
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
              <div class="flex-1 relative">
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
                    @keydown="handleTextareaKeydown" @contextmenu="handleTextareaContextMenu"
                    class="w-full h-64 lg:h-full p-4 outline-none border border-gray-300 rounded-lg resize-none focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white text-gray-900 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100"
                    :style="{
                      fontSize: fontSize + 'px',
                      lineHeight: lineHeight.toString(),
                      fontFamily: 'system-ui, -apple-system, sans-serif'
                    }" placeholder="Start writing markdown...">
</textarea>
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
              <p class="text-sm text-gray-600 dark:text-gray-400 px-4">This PDF may contain only images or have no
                extractable text.</p>
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
            <div class="hidden sm:block text-xs text-gray-400 dark:text-gray-500 flex-shrink-0">{{
              suggestion.shortcut.replace('Ctrl+Shift+', 'C+S+') }}</div>
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
      <div
        class="bg-white rounded-lg max-w-2xl w-full max-h-[90vh] border border-gray-300 dark:bg-gray-800 dark:border-gray-600 flex flex-col">
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
        <div
          class="p-4 border-t border-gray-300 flex flex-col sm:flex-row gap-2 justify-end dark:border-gray-600 flex-shrink-0">
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
          <pre
            class="whitespace-pre-wrap text-sm text-gray-800 dark:text-gray-200 font-mono break-words">{{ previewContent }}</pre>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Clear floats after floated images */
.image-container::after {
  content: "";
  display: table;
  clear: both;
}

/* Ensure text wraps properly around floated images */
.image-container.float-left+*,
.image-container.float-right+* {
  overflow: hidden;
}

/* Enhanced blockquote styling */
blockquote {
  margin: 1rem 0;
  padding-left: 1rem;
  border-left: 4px solid #d1d5db;
  font-style: italic;
}

.dark blockquote {
  border-left-color: #4b5563;
}

/* Nested quotes */
blockquote blockquote {
  border-left-width: 3px;
  margin-left: 0.5rem;
  opacity: 0.9;
}

blockquote blockquote blockquote {
  border-left-width: 2px;
  opacity: 0.8;
}

/* Ensure proper text wrapping */
.whitespace-pre-wrap {
  white-space: pre-wrap;
}

.break-words {
  word-wrap: break-word;
  overflow-wrap: break-word;
}

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
