export function useKeyboardShortcuts(callbacks: {
  undo: () => void
  redo: () => void
  insertBold: () => void
  insertItalic: () => void
  insertStrikethrough: () => void
  insertCode: () => void
  insertLink: () => void
  insertHeader: (level: number) => void
  insertList: () => void
  insertNumberedList: () => void
  insertTaskList: () => void
  insertQuote: () => void
  insertCodeBlock: () => void
  insertImage: () => void
  insertHighlight: () => void
  insertHorizontalRule: () => void
  insertTable: () => void
  saveDocument: () => void
  showAIToolbar: () => void
  handleAIShortcut: (action: string) => void
}) {
  function handleEditorKeydown(event: KeyboardEvent) {
    const isMac = navigator.platform.toUpperCase().indexOf('MAC') >= 0
    const modKey = isMac ? event.metaKey : event.ctrlKey
    
    // Handle basic shortcuts (Ctrl/Cmd + key)
    if (modKey && !event.shiftKey && !event.altKey) {
      handleBasicShortcuts(event)
    }
    
    // Handle AI shortcuts (Ctrl/Cmd + Shift + key)
    if (modKey && event.shiftKey && !event.altKey) {
      handleAIShortcuts(event)
    }
    
    // Handle advanced shortcuts (Ctrl/Cmd + Alt + key)
    if (modKey && event.altKey && !event.shiftKey) {
      handleAdvancedShortcuts(event)
    }
    
    // Handle special cases
    handleSpecialCases(event, modKey)
  }

  function handleBasicShortcuts(event: KeyboardEvent) {
    const shortcuts: Record<string, () => void> = {
      'z': () => callbacks.undo(),
      'y': () => callbacks.redo(),
      'b': () => callbacks.insertBold(),
      'i': () => callbacks.insertItalic(),
      'u': () => callbacks.insertStrikethrough(),
      'e': () => callbacks.insertCode(),
      'k': () => callbacks.insertLink(),
      's': () => {
        event.preventDefault()
        callbacks.saveDocument()
      },
      '/': () => {
        event.preventDefault()
        callbacks.showAIToolbar()
      },
      'q': () => { // Added quote to basic shortcuts for quick access
        event.preventDefault()
        callbacks.insertQuote()
      },
    }
    
    const action = shortcuts[event.key.toLowerCase()]
    if (action) {
      event.preventDefault()
      action()
    }
  }

  function handleAIShortcuts(event: KeyboardEvent) {
    const shortcuts: Record<string, string> = {
      'S': 'summarize',
      'E': 'expand',
      'I': 'improve',
      'P': 'simplify',
      'R': 'paraphrase',
      'T': 'translate',
      'C': 'continue',
      'F': 'fix',
      'G': 'grammar',
      'Q': 'quote', // AI-powered quote formatting
    }
    
    const action = shortcuts[event.key]
    if (action) {
      event.preventDefault()
      callbacks.handleAIShortcut(action)
    }
  }

  function handleAdvancedShortcuts(event: KeyboardEvent) {
    const shortcuts: Record<string, () => void> = {
      '1': () => callbacks.insertHeader(1),
      '2': () => callbacks.insertHeader(2),
      '3': () => callbacks.insertHeader(3),
      '4': () => callbacks.insertHeader(4),
      '5': () => callbacks.insertHeader(5),
      '6': () => callbacks.insertHeader(6),
      'l': () => callbacks.insertList(),
      'n': () => callbacks.insertNumberedList(), // Changed from 'o' to 'n' for numbered
      't': () => callbacks.insertTaskList(),
      'q': () => callbacks.insertQuote(), // Alternative shortcut
      'c': () => callbacks.insertCodeBlock(),
      'i': () => callbacks.insertImage(),
      'h': () => callbacks.insertHighlight(),
      'r': () => callbacks.insertHorizontalRule(),
      'd': () => callbacks.insertTable(),
      's': () => callbacks.saveDocument(), // Alternative save shortcut
    }
    
    const action = shortcuts[event.key.toLowerCase()]
    if (action) {
      event.preventDefault()
      action()
    }
  }

  function handleSpecialCases(event: KeyboardEvent, modKey: boolean) {
    // Handle Enter for hard line breaks within quotes
    if (event.key === 'Enter') {
      const textarea = event.target as HTMLTextAreaElement
      if (textarea) {
        const start = textarea.selectionStart
        const value = textarea.value
        
        // Check if we're inside a quote
        const linesBefore = value.substring(0, start).split('\n')
        const currentLine = linesBefore[linesBefore.length - 1]
        
        if (currentLine.startsWith('>')) {
          // In quote - insert new line with quote continuation
          event.preventDefault()
          const quoteLevel = (currentLine.match(/^>+/)?.[0] || '').length
          const quotePrefix = '>'.repeat(quoteLevel) + ' '
          
          const newValue = value.substring(0, start) + '\n' + quotePrefix + value.substring(start)
          textarea.value = newValue
          textarea.setSelectionRange(start + 1 + quotePrefix.length, start + 1 + quotePrefix.length)
          
          // Trigger input event for Vue reactivity
          textarea.dispatchEvent(new Event('input', { bubbles: true }))
        }
      }
    }
    
    // Handle Tab key for nested quotes
    if (event.key === 'Tab') {
      const textarea = event.target as HTMLTextAreaElement
      if (textarea) {
        const start = textarea.selectionStart
        const value = textarea.value
        
        // Check if we're at the beginning of a quote line
        const lines = value.substring(0, start).split('\n')
        const currentLineIndex = lines.length - 1
        const currentLine = lines[currentLineIndex]
        
        if (currentLine.startsWith('>') && textarea.selectionStart === textarea.selectionEnd) {
          event.preventDefault()
          
          if (event.shiftKey) {
            // Shift+Tab: Decrease quote nesting
            if (currentLine.startsWith('>>')) {
              const newLine = currentLine.replace(/^>/, '')
              const newValue = value.split('\n')
              newValue[currentLineIndex] = newLine
              textarea.value = newValue.join('\n')
              textarea.setSelectionRange(start - 1, start - 1)
            }
          } else {
            // Tab: Increase quote nesting
            const newLine = '>' + currentLine
            const newValue = value.split('\n')
            newValue[currentLineIndex] = newLine
            textarea.value = newValue.join('\n')
            textarea.setSelectionRange(start + 1, start + 1)
          }
          
          textarea.dispatchEvent(new Event('input', { bubbles: true }))
        }
      }
    }
  }

  function getShortcutHint(action: string): string {
    const isMac = navigator.platform.toUpperCase().indexOf('MAC') >= 0
    const mod = isMac ? '⌘' : 'Ctrl'
    
    const shortcuts: Record<string, string> = {
      'bold': `${mod}+B`,
      'italic': `${mod}+I`,
      'strikethrough': `${mod}+U`,
      'code': `${mod}+E`,
      'link': `${mod}+K`,
      'quote': `${mod}+Q`, // Added basic quote shortcut
      'h1': `${mod}+Alt+1`,
      'h2': `${mod}+Alt+2`,
      'h3': `${mod}+Alt+3`,
      'h4': `${mod}+Alt+4`,
      'h5': `${mod}+Alt+5`,
      'h6': `${mod}+Alt+6`,
      'list': `${mod}+Alt+L`,
      'numbered': `${mod}+Alt+N`,
      'task': `${mod}+Alt+T`,
      'quote_alt': `${mod}+Alt+Q`, // Alternative quote shortcut
      'codeblock': `${mod}+Alt+C`,
      'image': `${mod}+Alt+I`,
      'table': `${mod}+Alt+D`,
      'save': `${mod}+S`,
      'save_alt': `${mod}+Alt+S`,
      'undo': `${mod}+Z`,
      'redo': isMac ? '⌘+Shift+Z' : 'Ctrl+Y',
      'ai_toolbar': `${mod}+/`,
    }
    
    return shortcuts[action] || ''
  }

  function getShortcutDescription(action: string): string {
    const descriptions: Record<string, string> = {
      'bold': 'Make text bold',
      'italic': 'Make text italic',
      'strikethrough': 'Strikethrough text',
      'code': 'Insert inline code',
      'link': 'Insert link',
      'quote': 'Insert blockquote',
      'h1': 'Insert heading 1',
      'h2': 'Insert heading 2',
      'h3': 'Insert heading 3',
      'list': 'Insert bullet list',
      'numbered': 'Insert numbered list',
      'task': 'Insert task list',
      'codeblock': 'Insert code block',
      'image': 'Insert image',
      'table': 'Insert table',
      'save': 'Save document',
      'undo': 'Undo last action',
      'redo': 'Redo last action',
      'ai_toolbar': 'Show AI assistant',
    }
    
    return descriptions[action] || ''
  }

  return {
    handleEditorKeydown,
    getShortcutHint,
    getShortcutDescription
  }
}