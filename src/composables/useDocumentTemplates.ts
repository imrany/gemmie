import { ref } from 'vue'
import type { UploadedFile } from '@/types/document'

export function useDocumentTemplates(
  addFile: (file: UploadedFile) => void,
  openEditor: (file: UploadedFile) => void
) {
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
    const newDoc: UploadedFile = {
      id: `doc-${Date.now()}-${Math.random().toString(36).slice(2)}`,
      name: `${template.name}.md`,
      url: 'custom',
      type: 'text/markdown',
      size: new Blob([template.content]).size,
      uploadedAt: new Date(),
      isCustom: true,
      content: template.content
    }

    addFile(newDoc)
    openEditor(newDoc)
  }

  return {
    documentTemplates,
    createDocumentFromTemplate
  }
}