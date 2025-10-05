import { ref, computed } from 'vue'
import type { HistoryEntry, EditableContent } from '@/types/document'
import type { Ref } from 'vue'

export function useHistory(
  editablePages: Ref<EditableContent[]>,
  currentPage: Ref<number>,
  getCurrentPageContent: () => EditableContent,
  updatePageContent: (content: string, saveHistory: boolean) => void
) {
  const undoHistory = ref<HistoryEntry[]>([])
  const redoHistory = ref<HistoryEntry[]>([])
  const editHistory = ref<Array<{ action: string, pageNum: number, timestamp: Date, preview: string }>>([])
  const maxHistorySize = 100

  const canUndo = computed(() => undoHistory.value.length > 0)
  const canRedo = computed(() => redoHistory.value.length > 0)

  function saveToHistory() {
    const currentPageData = getCurrentPageContent()
    if (!currentPageData) return

    const historyEntry: HistoryEntry = {
      content: currentPageData.content,
      timestamp: Date.now()
    }

    undoHistory.value.push(historyEntry)
    redoHistory.value = []

    if (undoHistory.value.length > maxHistorySize) {
      undoHistory.value.shift()
    }
  }

  function undo() {
    if (!canUndo.value) return

    const currentPageData = getCurrentPageContent()
    if (!currentPageData) return

    redoHistory.value.push({
      content: currentPageData.content,
      timestamp: Date.now()
    })

    const previousState = undoHistory.value.pop()!
    updatePageContent(previousState.content, false)

    addToHistory('Undo', 'Undid last change')
  }

  function redo() {
    if (!canRedo.value) return

    const currentPageData = getCurrentPageContent()
    if (!currentPageData) return

    undoHistory.value.push({
      content: currentPageData.content,
      timestamp: Date.now()
    })

    const nextState = redoHistory.value.pop()!
    updatePageContent(nextState.content, false)

    addToHistory('Redo', 'Redid last change')
  }

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

  function clearHistory() {
    undoHistory.value = []
    redoHistory.value = []
    editHistory.value = []
  }

  return {
    undoHistory,
    redoHistory,
    editHistory,
    canUndo,
    canRedo,
    saveToHistory,
    undo,
    redo,
    addToHistory,
    clearHistory
  }
}
