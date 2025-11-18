import * as pdfjsLib from 'pdfjs-dist'
import type { UploadedFile, EditableContent } from '@/types/document'

export async function generatePdfThumbnail(file: File): Promise<{ previewUrl: string; pages: number }> {
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

export async function extractPdfContent(url: string): Promise<EditableContent[]> {
  const arrayBuffer = await fetch(url).then(response => response.arrayBuffer())
  const pdf = await pdfjsLib.getDocument({ data: arrayBuffer }).promise
  const totalPages = pdf.numPages
  const editablePages: EditableContent[] = []

  for (let pageNum = 1; pageNum <= totalPages; pageNum++) {
    const page = await pdf.getPage(pageNum)
    const textContent = await page.getTextContent()

    let pageText = ''
    textContent.items.forEach((item: any) => {
      if ('str' in item) {
        pageText += item.str + ' '
      }
    })

    pageText = pageText.replace(/\s+/g, ' ').trim()

    editablePages.push({
      pageNum,
      content: pageText,
      originalContent: pageText,
      isModified: false,
      annotations: []
    })
  }

  return editablePages
}