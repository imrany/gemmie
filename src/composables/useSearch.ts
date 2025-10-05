import { ref, computed } from 'vue'
import type { SearchResult, EditableContent } from '@/types/document'
import type { Ref } from 'vue'

export function useSearch(editablePages: Ref<EditableContent[]>) {
  const searchQuery = ref('')
  const searchResults = ref<SearchResult[]>([])
  const currentSearchIndex = ref(0)

  const hasSearchResults = computed(() => searchResults.value.length > 0)

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

  function goToSearchResult(index: number, goToPageFn: (pageNum: number) => void) {
    if (index >= 0 && index < searchResults.value.length) {
      currentSearchIndex.value = index
      const result = searchResults.value[index]
      goToPageFn(result.pageNum)
    }
  }

  function clearSearch() {
    searchQuery.value = ''
    searchResults.value = []
    currentSearchIndex.value = 0
  }

  return {
    searchQuery,
    searchResults,
    currentSearchIndex,
    hasSearchResults,
    performSearch,
    goToSearchResult,
    clearSearch
  }
}
