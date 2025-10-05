import { ref } from 'vue'
import type { EditableContent } from '@/types/document'
import type { Ref } from 'vue'

export function useExport(
  editablePages: Ref<EditableContent[]>,
  selectedPdfName: Ref<string>,
  totalPages: Ref<number>
) {
  const showPreview = ref(false)
  const previewContent = ref('')
  const showExportDropdown = ref(false)

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
  }

  function toggleExportDropdown() {
    showExportDropdown.value = !showExportDropdown.value
  }

  function closeExportDropdown() {
    showExportDropdown.value = false
  }

  return {
    showPreview,
    previewContent,
    showExportDropdown,
    generatePreview,
    closePreview,
    downloadAsText,
    downloadAsMarkdown,
    toggleExportDropdown,
    closeExportDropdown
  }
}
