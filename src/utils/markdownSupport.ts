/**
 * Enhanced Markdown Rendering with Full Support
 */
function renderMarkdown(text: string): string {
  if (!text) return ''

  let html = text

  // Escape HTML to prevent XSS
  html = html
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')

  // Code blocks (must be before inline code)
  html = html.replace(/```(\w+)?\n([\s\S]+?)```/g, (match, lang, code) => {
    const language = lang ? ` data-lang="${lang}"` : ''
    return `<pre class="bg-gray-900 dark:bg-gray-950 text-gray-100 p-4 rounded-lg overflow-x-auto my-4"><code${language} class="font-mono text-sm">${code.trim()}</code></pre>`
  })

  // Inline code (after code blocks)
  html = html.replace(/`([^`]+)`/g, '<code class="bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded text-sm font-mono text-red-600 dark:text-red-400">$1</code>')

  // Headers (H1-H6)
  html = html
    .replace(/^###### (.*$)/gm, '<h6 class="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-1 mt-2">$1</h6>')
    .replace(/^##### (.*$)/gm, '<h5 class="text-base font-semibold text-gray-900 dark:text-gray-100 mb-2 mt-3">$1</h5>')
    .replace(/^#### (.*$)/gm, '<h4 class="text-base font-semibold text-gray-900 dark:text-gray-100 mb-2 mt-3">$1</h4>')
    .replace(/^### (.*$)/gm, '<h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2 mt-4">$1</h3>')
    .replace(/^## (.*$)/gm, '<h2 class="text-xl font-bold text-gray-900 dark:text-gray-100 mb-3 mt-6">$1</h2>')
    .replace(/^# (.*$)/gm, '<h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4 mt-8">$1</h1>')

  // Horizontal rules
  html = html.replace(/^(---|\*\*\*|___)$/gm, '<hr class="border-t border-gray-300 dark:border-gray-600 my-6">')

  // Bold, italic, and combined (order matters!)
  html = html
    .replace(/\*\*\*(.+?)\*\*\*/g, '<strong><em class="font-bold italic text-gray-900 dark:text-gray-100">$1</em></strong>')
    .replace(/___(.+?)___/g, '<strong><em class="font-bold italic text-gray-900 dark:text-gray-100">$1</em></strong>')
    .replace(/\*\*(.+?)\*\*/g, '<strong class="font-bold text-gray-900 dark:text-gray-100">$1</strong>')
    .replace(/__(.+?)__/g, '<strong class="font-bold text-gray-900 dark:text-gray-100">$1</strong>')
    .replace(/\*(.+?)\*/g, '<em class="italic text-gray-800 dark:text-gray-200">$1</em>')
    .replace(/_(.+?)_/g, '<em class="italic text-gray-800 dark:text-gray-200">$1</em>')

  // Strikethrough
  html = html.replace(/~~(.+?)~~/g, '<del class="line-through text-gray-500 dark:text-gray-400">$1</del>')

  // Highlight
  html = html.replace(/==(.+?)==/g, '<mark class="bg-yellow-200 dark:bg-yellow-700 px-1">$1</mark>')

  // Links with title support
  html = html.replace(/\[([^\]]+)\]\(([^)"]+)(?:\s+"([^"]+)")?\)/g, (match, text, url, title) => {
    const titleAttr = title ? ` title="${title}"` : ''
    return `<a href="${url}" class="text-blue-600 dark:text-blue-400 hover:underline"${titleAttr} target="_blank" rel="noopener noreferrer">${text}</a>`
  })

  // Images with alt text
  html = html.replace(/!\[([^\]]*)\]\(([^)"]+)(?:\s+"([^"]+)")?\)/g, (match, alt, url, title) => {
    const titleAttr = title ? ` title="${title}"` : ''
    return `<img src="${url}" alt="${alt}"${titleAttr} class="max-w-full h-auto rounded-lg my-4">`
  })

  // Task lists (must be before regular lists)
  html = html.replace(/^- \[([ x])\] (.+$)/gm, (match, checked, text) => {
    const isChecked = checked.toLowerCase() === 'x'
    return `<li class="flex items-start ml-4 text-gray-800 dark:text-gray-200"><input type="checkbox" ${isChecked ? 'checked' : ''} disabled class="mt-1 mr-2"><span>${text}</span></li>`
  })

  // Unordered lists
  html = html.replace(/^[\*\-\+] (.+$)/gm, '<li class="ml-4 text-gray-800 dark:text-gray-200">• $1</li>')

  // Ordered lists
  html = html.replace(/^\d+\. (.+$)/gm, '<li class="ml-4 text-gray-800 dark:text-gray-200 list-decimal">$1</li>')

  // Blockquotes (nested support)
  html = html.replace(/^> (.+$)/gm, '<blockquote class="border-l-4 border-gray-300 dark:border-gray-600 pl-4 italic text-gray-700 dark:text-gray-300 my-2">$1</blockquote>')
  html = html.replace(/(<blockquote[^>]*>.*?<\/blockquote>)(?:\s*<br>\s*)?(?=<blockquote|$)/gs, '$1')

  // Tables
  html = html.replace(/^\|(.+)\|$/gm, (match) => {
    const cells = match.slice(1, -1).split('|').map(cell => cell.trim())
    return `<tr>${cells.map(cell => `<td class="border border-gray-300 dark:border-gray-600 px-4 py-2 text-gray-800 dark:text-gray-200">${cell}</td>`).join('')}</tr>`
  })
  html = html.replace(/(<tr>.*?<\/tr>(?:\s*<br>\s*<tr>.*?<\/tr>)*)/gs, '<table class="border-collapse border border-gray-300 dark:border-gray-600 my-4 w-full">$1</table>')

  // Wrap consecutive list items in ul/ol
  html = html.replace(/(<li class="ml-4[^>]*>• .*?<\/li>)(?:\s*<br>\s*)?(?=<li class="ml-4[^>]*>• |$)/gs, '$1')
  html = html.replace(/(<li class="ml-4[^>]*>• .*?<\/li>(?:\s*<li class="ml-4[^>]*>• .*?<\/li>)*)/gs, '<ul class="my-2 space-y-1">$1</ul>')

  html = html.replace(/(<li class="ml-4[^>]*list-decimal[^>]*>.*?<\/li>)(?:\s*<br>\s*)?(?=<li class="ml-4[^>]*list-decimal|$)/gs, '$1')
  html = html.replace(/(<li class="ml-4[^>]*list-decimal[^>]*>.*?<\/li>(?:\s*<li class="ml-4[^>]*list-decimal[^>]*>.*?<\/li>)*)/gs, '<ol class="my-2 space-y-1 list-decimal ml-8">$1</ol>')

  // Paragraphs (wrap standalone text)
  html = html.split('<br>').map(line => {
    line = line.trim()
    if (!line) return ''
    if (line.startsWith('<h') || line.startsWith('<ul') || line.startsWith('<ol') || 
        line.startsWith('<pre') || line.startsWith('<blockquote') || 
        line.startsWith('<hr') || line.startsWith('<table') || line.startsWith('<img')) {
      return line
    }
    return `<p class="text-gray-800 dark:text-gray-200 mb-4">${line}</p>`
  }).join('')

  return html
}

export { 
    renderMarkdown 
}