import hljs from "highlight.js";

/**
 * Enhanced Markdown Rendering with Full Support including Advanced Image Features
 * and Syntax Highlighting
 */
function renderMarkdown(text: string): string {
  if (!text) return "";

  let html = text;

  // Escape HTML to prevent XSS
  const escapeHtml = (unsafe: string): string => {
    return unsafe
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;")
      .replace(/'/g, "&#039;");
  };

  // Process ALL code blocks and inline code FIRST, before any HTML escaping
  const codeBlockPlaceholders: string[] = [];
  const inlineCodePlaceholders: string[] = [];
  
  // Process code blocks (```code```)
  html = html.replace(/```(\w+)?\n([\s\S]*?)```/g, (match, lang, code) => {
    const language = lang || "plaintext";
    let highlightedCode = code;

    try {
      if (hljs.getLanguage(language)) {
        highlightedCode = hljs.highlight(code.trim(), { language }).value;
      } else {
        highlightedCode = hljs.highlightAuto(code.trim()).value;
      }
    } catch (error) {
      console.warn("Syntax highlighting failed:", error);
      highlightedCode = escapeHtml(code.trim());
    }

    const placeholder = `___CODE_BLOCK_${codeBlockPlaceholders.length}___`;
    
    codeBlockPlaceholders.push(`
<div class="code-container relative my-4 bg-gray-900 dark:bg-gray-950 rounded-lg overflow-hidden">
  <div class="flex items-center justify-between px-4 py-2 bg-gray-800 dark:bg-gray-900 border-b border-gray-700">
    <span class="text-xs font-mono text-gray-400 capitalize">${language}</span>
    <button 
      class="copy-button flex items-center gap-1 px-2 py-1 text-xs text-gray-300 hover:text-white bg-gray-700 hover:bg-gray-600 rounded transition-colors"
      data-code="${encodeURIComponent(code.trim())}"
    >
      <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"/>
      </svg>
      Copy
    </button>
  </div>
  <pre class="p-4 overflow-x-auto"><code class="hljs language-${language} text-sm leading-relaxed">${highlightedCode}</code></pre>
</div>`);

    return placeholder;
  });

  // Process inline code (`code`)
  html = html.replace(/`([^`]+)`/g, (match, code) => {
    const placeholder = `___INLINE_CODE_${inlineCodePlaceholders.length}___`;
    inlineCodePlaceholders.push(
      `<code class="inline-code bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded text-sm font-mono text-red-600 dark:text-red-400 border border-gray-200 dark:border-gray-700">${escapeHtml(code)}</code>`
    );
    return placeholder;
  });

  // NOW escape the remaining HTML (placeholders remain untouched)
  const protectedParts: string[] = [];
  
  // Protect code block placeholders
  html = html.replace(/(___CODE_BLOCK_\d+___)/g, (match) => {
    protectedParts.push(match);
    return `___PROTECTED_${protectedParts.length - 1}___`;
  });
  
  // Protect inline code placeholders  
  html = html.replace(/(___INLINE_CODE_\d+___)/g, (match) => {
    protectedParts.push(match);
    return `___PROTECTED_${protectedParts.length - 1}___`;
  });

  // Now escape the HTML
  html = escapeHtml(html);

  // Restore protected placeholders
  html = html.replace(/___PROTECTED_(\d+)___/g, (match, index) => {
    return protectedParts[parseInt(index)] || match;
  });

  // Enhanced Image Support with multiple formats and features
  const processImages = (content: string): string => {
    return (
      content
        // Clickable images: [![alt](src)](link)
        .replace(
          /\[!\[([^\]]*)\]\(([^)\s]+)\)\]\(([^)\s]+)\)/g,
          (match, alt, imageUrl, linkUrl) => {
            return `<div class="image-container my-4">
<a href="${linkUrl}" target="_blank" rel="noopener noreferrer">
<img src="${imageUrl}" alt="${alt || "Image"}" loading="lazy" class="max-w-full h-auto rounded-lg shadow-sm transition-all duration-200 hover:shadow-md cursor-pointer" onerror="this.style.display='none'; this.nextElementSibling.style.display='block';">
<div style="display:none;" class="text-gray-600 dark:text-gray-400 p-4 border border-gray-300 dark:border-gray-600 rounded-lg">${alt || "Image"}</div>
</a>
</div>`;
          }
        )
        // Images with width/height specification: ![alt](src =100x200)
        .replace(
          /!\[([^\]]*)\]\(([^)\s]+)\s*=(\d+)x(\d+)\)/g,
          (match, alt, url, width, height) => {
            return `<div class="image-container my-4">
<img src="${url}" alt="${alt || "Image"}" width="${width}" height="${height}" loading="lazy" class="max-w-full h-auto rounded-lg shadow-sm transition-all duration-200 hover:shadow-md cursor-zoom-in" onclick="window.open('${url}', '_blank')" onerror="this.style.display='none'; this.nextElementSibling.style.display='block';">
<div style="display:none;" class="text-gray-600 dark:text-gray-400 p-4 border border-gray-300 dark:border-gray-600 rounded-lg">${alt || "Image"}</div>
</div>`;
          }
        )
        // Float left images: ![alt](src){.float-left}
        .replace(
          /!\[([^\]]*)\]\(([^)\s]+)\)\{\.float-left\}/g,
          (match, alt, url) => {
            return `<div class="image-container my-4 float-left mr-4 mb-4">
<img src="${url}" alt="${alt || "Image"}" loading="lazy" class="max-w-48 h-auto rounded-lg shadow-sm transition-all duration-200 hover:shadow-md cursor-zoom-in" onclick="window.open('${url}', '_blank')" onerror="this.style.display='none'; this.nextElementSibling.style.display='block';">
<div style="display:none;" class="text-gray-600 dark:text-gray-400 p-4 border border-gray-300 dark:border-gray-600 rounded-lg">${alt || "Image"}</div>
</div>`;
          }
        )
        // Float right images: ![alt](src){.float-right}
        .replace(
          /!\[([^\]]*)\]\(([^)\s]+)\)\{\.float-right\}/g,
          (match, alt, url) => {
            return `<div class="image-container my-4 float-right ml-4 mb-4">
<img src="${url}" alt="${alt || "Image"}" loading="lazy" class="max-w-48 h-auto rounded-lg shadow-sm transition-all duration-200 hover:shadow-md cursor-zoom-in" onclick="window.open('${url}', '_blank')" onerror="this.style.display='none'; this.nextElementSibling.style.display='block';">
<div style="display:none;" class="text-gray-600 dark:text-gray-400 p-4 border border-gray-300 dark:border-gray-600 rounded-lg">${alt || "Image"}</div>
</div>`;
          }
        )
        // Centered images: ![alt](src){.center}
        .replace(
          /!\[([^\]]*)\]\(([^)\s]+)\)\{\.center\}/g,
          (match, alt, url) => {
            return `<div class="image-container my-4 flex justify-center">
<img src="${url}" alt="${alt || "Image"}" loading="lazy" class="max-w-full h-auto rounded-lg shadow-sm transition-all duration-200 hover:shadow-md cursor-zoom-in" onclick="window.open('${url}', '_blank')" onerror="this.style.display='none'; this.nextElementSibling.style.display='block';">
<div style="display:none;" class="text-gray-600 dark:text-gray-400 p-4 border border-gray-300 dark:border-gray-600 rounded-lg">${alt || "Image"}</div>
</div>`;
          }
        )
        // Small images: ![alt](src){.small}
        .replace(
          /!\[([^\]]*)\]\(([^)\s]+)\)\{\.small\}/g,
          (match, alt, url) => {
            return `<div class="image-container my-4">
<img src="${url}" alt="${alt || "Image"}" loading="lazy" class="max-w-48 h-auto rounded-lg shadow-sm transition-all duration-200 hover:shadow-md cursor-zoom-in" onclick="window.open('${url}', '_blank')" onerror="this.style.display='none'; this.nextElementSibling.style.display='block';">
<div style="display:none;" class="text-gray-600 dark:text-gray-400 p-4 border border-gray-300 dark:border-gray-600 rounded-lg">${alt || "Image"}</div>
</div>`;
          }
        )
        // Border images: ![alt](src){.border}
        .replace(
          /!\[([^\]]*)\]\(([^)\s]+)\)\{\.border\}/g,
          (match, alt, url) => {
            return `<div class="image-container my-4">
<img src="${url}" alt="${alt || "Image"}" loading="lazy" class="max-w-full h-auto rounded-lg border-2 border-gray-300 dark:border-gray-600 shadow-sm transition-all duration-200 hover:shadow-md cursor-zoom-in" onclick="window.open('${url}', '_blank')" onerror="this.style.display='none'; this.nextElementSibling.style.display='block';">
<div style="display:none;" class="text-gray-600 dark:text-gray-400 p-4 border border-gray-300 dark:border-gray-600 rounded-lg">${alt || "Image"}</div>
</div>`;
          }
        )
        // Images with CSS class: ![alt](src){.class-name}
        .replace(
          /!\[([^\]]*)\]\(([^)\s]+)\)\{\.([^}]+)\}/g,
          (match, alt, url, className) => {
            return `<div class="image-container my-4">
<img src="${url}" alt="${alt || "Image"}" loading="lazy" class="${className} max-w-full h-auto rounded-lg shadow-sm transition-all duration-200 hover:shadow-md cursor-zoom-in" onclick="window.open('${url}', '_blank')" onerror="this.style.display='none'; this.nextElementSibling.style.display='block';">
<div style="display:none;" class="text-gray-600 dark:text-gray-400 p-4 border border-gray-300 dark:border-gray-600 rounded-lg">${alt || "Image"}</div>
</div>`;
          }
        )
        // Standard Markdown images: ![alt](src "title") - must be LAST
        .replace(
          /!\[([^\]]*)\]\(([^)"]+)(?:\s+"([^"]+)")?\)/g,
          (match, alt, url, title) => {
            const titleAttr = title ? ` title="${title}"` : "";
            const isSvg = url.toLowerCase().endsWith(".svg");
            const loadingAttr = isSvg ? "" : ' loading="lazy"';

            return `<div class="image-container my-4">
<img src="${url}" alt="${alt || "Image"}"${titleAttr}${loadingAttr} class="max-w-full h-auto rounded-lg shadow-sm transition-all duration-200 hover:shadow-md cursor-zoom-in" onclick="window.open('${url}', '_blank')" onerror="this.style.display='none'; this.nextElementSibling.style.display='block';">
<div style="display:none;" class="text-gray-600 dark:text-gray-400 p-4 border border-gray-300 dark:border-gray-600 rounded-lg">${alt || "Image"}</div>
</div>`;
          }
        )
    );
  };

  // Process images FIRST (before other markdown processing)
  html = processImages(html);

  // Process tables
  const processTable = (content: string): string => {
    const tableRegex = /(\|.+\|[\r\n]+\|[-:\s|]+\|[\r\n]+(?:\|.+\|[\r\n]*)+)/g;
    
    return content.replace(tableRegex, (match) => {
      const lines = match.trim().split('\n').map(line => line.trim());
      
      if (lines.length < 2) return match;
      
      // Parse header
      const headerCells = lines[0]
        .split('|')
        .filter(cell => cell.trim())
        .map(cell => cell.trim());
      
      // Parse alignment from separator row
      const alignments = lines[1]
        .split('|')
        .filter(cell => cell.trim())
        .map(cell => {
          const trimmed = cell.trim();
          if (trimmed.startsWith(':') && trimmed.endsWith(':')) return 'center';
          if (trimmed.endsWith(':')) return 'right';
          return 'left';
        });
      
      // Parse body rows
      const bodyRows = lines.slice(2).map(line => 
        line.split('|')
          .filter(cell => cell.trim())
          .map(cell => cell.trim())
      );
      
      // Build table HTML
      let tableHtml = '<div class="table-container my-4 overflow-x-auto"><table class="w-full border-collapse border border-gray-300 dark:border-gray-600">';
      
      // Header
      tableHtml += '<thead class="bg-gray-100 dark:bg-gray-800"><tr>';
      headerCells.forEach((cell, i) => {
        const align = alignments[i] || 'left';
        tableHtml += `<th class="border border-gray-300 dark:border-gray-600 px-4 py-2 text-${align} font-semibold text-gray-900 dark:text-gray-100">${cell}</th>`;
      });
      tableHtml += '</tr></thead>';
      
      // Body
      tableHtml += '<tbody>';
      bodyRows.forEach(row => {
        tableHtml += '<tr class="hover:bg-gray-50 dark:hover:bg-gray-700/50">';
        row.forEach((cell, i) => {
          const align = alignments[i] || 'left';
          tableHtml += `<td class="border border-gray-300 dark:border-gray-600 px-4 py-2 text-${align} text-gray-800 dark:text-gray-200">${cell}</td>`;
        });
        tableHtml += '</tr>';
      });
      tableHtml += '</tbody></table></div>';
      
      return tableHtml;
    });
  };

  html = processTable(html);

  // Headers (H1-H6)
  html = html
    .replace(
      /^###### (.*$)/gm,
      '<h6 class="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-1 mt-2">$1</h6>'
    )
    .replace(
      /^##### (.*$)/gm,
      '<h5 class="text-base font-semibold text-gray-900 dark:text-gray-100 mb-2 mt-3">$1</h5>'
    )
    .replace(
      /^#### (.*$)/gm,
      '<h4 class="text-base font-semibold text-gray-900 dark:text-gray-100 mb-2 mt-3">$1</h4>'
    )
    .replace(
      /^### (.*$)/gm,
      '<h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2 mt-4">$1</h3>'
    )
    .replace(
      /^## (.*$)/gm,
      '<h2 class="text-xl font-bold text-gray-900 dark:text-gray-100 mb-3 mt-6">$1</h2>'
    )
    .replace(
      /^# (.*$)/gm,
      '<h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4 mt-8">$1</h1>'
    );

  // Horizontal rules
  html = html.replace(
    /^(---|\*\*\*|___)$/gm,
    '<hr class="border-t border-gray-300 dark:border-gray-600 my-6">'
  );

  // Links with title support (after images)
  html = html.replace(
    /\[([^\]]+)\]\(([^)"]+)(?:\s+"([^"]+)")?\)/g,
    (match, text, url, title) => {
      const titleAttr = title ? ` title="${title}"` : "";
      return `<a href="${url}" class="text-blue-600 dark:text-blue-400 hover:underline"${titleAttr} target="_blank" rel="noopener noreferrer">${text}</a>`;
    }
  );

  // Bold and italic
  html = html
    .replace(/\*\*\*(.+?)\*\*\*/g, '<strong><em>$1</em></strong>')
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.+?)\*/g, '<em>$1</em>');

  // Lists
  html = html.replace(/^[\*\-\+] (.+$)/gm, '<li class="ml-4">• $1</li>');
  html = html.replace(/^\d+\. (.+$)/gm, '<li class="ml-4 list-decimal">$1</li>');

  // Wrap lists
  html = html.replace(/(<li class="ml-4">• .*?<\/li>(?:\s*<li class="ml-4">• .*?<\/li>)*)/gs, '<ul class="my-2">$1</ul>');
  html = html.replace(/(<li class="ml-4 list-decimal">.*?<\/li>(?:\s*<li class="ml-4 list-decimal">.*?<\/li>)*)/gs, '<ol class="my-2 ml-8">$1</ol>');

  // Paragraphs
  const lines = html.split('\n');
  const processedLines: string[] = [];
  
  for (const line of lines) {
    const trimmedLine = line.trim();
    if (!trimmedLine) {
      processedLines.push('');
      continue;
    }
    
    // Skip if it's already HTML or a placeholder
    if (
      trimmedLine.startsWith('<') ||
      trimmedLine.includes('___CODE_BLOCK_') ||
      trimmedLine.includes('___INLINE_CODE_')
    ) {
      processedLines.push(trimmedLine);
    } else {
      processedLines.push(`<p class="text-gray-800 dark:text-gray-200 mb-4">${trimmedLine}</p>`);
    }
  }
  
  html = processedLines.join('\n');

  // FINALLY: Restore code blocks and inline code
  html = html.replace(/___INLINE_CODE_(\d+)___/g, (match, index) => {
    return inlineCodePlaceholders[parseInt(index)] || match;
  });

  html = html.replace(/___CODE_BLOCK_(\d+)___/g, (match, index) => {
    return codeBlockPlaceholders[parseInt(index)] || match;
  });

  return html;
}

export { renderMarkdown };