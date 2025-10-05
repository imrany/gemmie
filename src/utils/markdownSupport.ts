/**
 * Enhanced Markdown Rendering with Full Support including Advanced Image Features
 */
function renderMarkdown(text: string): string {
  if (!text) return "";

  let html = text;

  // Escape HTML to prevent XSS
  html = html
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;");

  // Enhanced Blockquotes with nested support
  const processBlockquotes = (content: string): string => {
    const lines = content.split("\n");
    let inBlockquote = false;
    let blockquoteLevel = 0;
    let blockquoteContent: string[] = [];
    const result: string[] = [];

    const closeBlockquote = (level: number) => {
      if (blockquoteContent.length > 0) {
        const nestedContent = processBlockquotes(blockquoteContent.join("\n"));
        const nestedClass = level > 1 ? ` nested-quote-level-${level}` : "";
        result.push(
          `<blockquote class="border-l-4 border-gray-300 dark:border-gray-600 pl-4 my-3 text-gray-700 dark:text-gray-300${nestedClass}">${nestedContent}</blockquote>`
        );
        blockquoteContent = [];
      }
    };

    for (const line of lines) {
      const trimmedLine = line.trim();

      // Check for blockquote (">" at start of line)
      if (trimmedLine.startsWith(">")) {
        const quoteMatch = trimmedLine.match(/^(>+)\s*(.*)/);
        if (quoteMatch) {
          const [, quotes, content] = quoteMatch;
          const level = quotes.length;

          if (!inBlockquote || level !== blockquoteLevel) {
            closeBlockquote(blockquoteLevel);
            blockquoteLevel = level;
            inBlockquote = true;
          }

          if (content) {
            blockquoteContent.push(content);
          } else {
            blockquoteContent.push(""); // Empty line in blockquote
          }
          continue;
        }
      }

      // Not a blockquote line
      if (inBlockquote) {
        closeBlockquote(blockquoteLevel);
        inBlockquote = false;
        blockquoteLevel = 0;
      }

      // Process regular line
      if (trimmedLine) {
        result.push(processInlineMarkdown(line));
      } else {
        result.push("<br>");
      }
    }

    // Close any remaining blockquote
    if (inBlockquote) {
      closeBlockquote(blockquoteLevel);
    }

    return result.join("\n");
  };

  // Process inline markdown (for non-blockquote content)
  const processInlineMarkdown = (line: string): string => {
    let processed = line;

    // Code blocks (must be before inline code)
    processed = processed.replace(
      /```(\w+)?\n([\s\S]+?)```/g,
      (match, lang, code) => {
        const language = lang ? ` data-lang="${lang}"` : "";
        return `<pre class="bg-gray-900 dark:bg-gray-950 text-gray-100 p-4 rounded-lg overflow-x-auto my-4"><code${language} class="font-mono text-sm">${code.trim()}</code></pre>`;
      }
    );

    // Inline code
    processed = processed.replace(
      /`([^`]+)`/g,
      '<code class="bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded text-sm font-mono text-red-600 dark:text-red-400">$1</code>'
    );

    // Bold, italic, etc. (your existing inline processing)
    processed = processed
      .replace(
        /\*\*\*(.+?)\*\*\*/g,
        '<strong><em class="font-bold italic text-gray-900 dark:text-gray-100">$1</em></strong>'
      )
      .replace(
        /___(.+?)___/g,
        '<strong><em class="font-bold italic text-gray-900 dark:text-gray-100">$1</em></strong>'
      )
      .replace(
        /\*\*(.+?)\*\*/g,
        '<strong class="font-bold text-gray-900 dark:text-gray-100">$1</strong>'
      )
      .replace(
        /__(.+?)__/g,
        '<strong class="font-bold text-gray-900 dark:text-gray-100">$1</strong>'
      )
      .replace(
        /\*(.+?)\*/g,
        '<em class="italic text-gray-800 dark:text-gray-200">$1</em>'
      )
      .replace(
        /_(.+?)_/g,
        '<em class="italic text-gray-800 dark:text-gray-200">$1</em>'
      )
      .replace(
        /~~(.+?)~~/g,
        '<del class="line-through text-gray-500 dark:text-gray-400">$1</del>'
      )
      .replace(
        /==(.+?)==/g,
        '<mark class="bg-yellow-200 dark:bg-yellow-700 px-1">$1</mark>'
      );

    return processed;
  };

  // Start processing with blockquotes
  html = processBlockquotes(html);

  // Enhanced Image Support with multiple formats and features
  const processImages = (content: string): string => {
    return content
      // Standard Markdown images: ![alt](src "title")
      .replace(
        /!\[([^\]]*)\]\(([^)"]+)(?:\s+"([^"]+)")?\)/g,
        (match, alt, url, title) => {
          const titleAttr = title ? ` title="${title}"` : "";
          const isSvg = url.toLowerCase().endsWith('.svg');
          const loadingAttr = isSvg ? '' : ' loading="lazy"';
          
          return `
            <div class="image-container my-4">
              <img 
                src="${url}" 
                alt="${alt || 'Image'}" 
                ${titleAttr}
                ${loadingAttr}
                class="max-w-full h-auto rounded-lg shadow-sm transition-all duration-200 hover:shadow-md cursor-zoom-in"
                onclick="window.open('${url}', '_blank')"
              >
              ${alt ? `<div class="image-caption text-sm text-gray-600 dark:text-gray-400 text-center mt-2">${alt}</div>` : ''}
            </div>
          `;
        }
      )
      // Images with width/height specification: ![alt](src =100x200)
      .replace(
        /!\[([^\]]*)\]\(([^)\s]+)\s*=(\d+)x(\d+)\)/g,
        (match, alt, url, width, height) => {
          return `
            <div class="image-container my-4">
              <img 
                src="${url}" 
                alt="${alt || 'Image'}" 
                width="${width}"
                height="${height}"
                loading="lazy"
                class="max-w-full h-auto rounded-lg shadow-sm transition-all duration-200 hover:shadow-md cursor-zoom-in"
                onclick="window.open('${url}', '_blank')"
              >
              ${alt ? `<div class="image-caption text-sm text-gray-600 dark:text-gray-400 text-center mt-2">${alt}</div>` : ''}
            </div>
          `;
        }
      )
      // Images with CSS class: ![alt](src){.class-name}
      .replace(
        /!\[([^\]]*)\]\(([^)\s]+)\)\{\.([^}]+)\}/g,
        (match, alt, url, className) => {
          return `
            <div class="image-container my-4">
              <img 
                src="${url}" 
                alt="${alt || 'Image'}" 
                loading="lazy"
                class="${className} max-w-full h-auto rounded-lg shadow-sm transition-all duration-200 hover:shadow-md cursor-zoom-in"
                onclick="window.open('${url}', '_blank')"
              >
              ${alt ? `<div class="image-caption text-sm text-gray-600 dark:text-gray-400 text-center mt-2">${alt}</div>` : ''}
            </div>
          `;
        }
      )
      // Centered images: ![alt](src){.center}
      .replace(
        /!\[([^\]]*)\]\(([^)\s]+)\)\{\.center\}/g,
        (match, alt, url) => {
          return `
            <div class="image-container my-4 flex justify-center">
              <img 
                src="${url}" 
                alt="${alt || 'Image'}" 
                loading="lazy"
                class="max-w-full h-auto rounded-lg shadow-sm transition-all duration-200 hover:shadow-md cursor-zoom-in"
                onclick="window.open('${url}', '_blank')"
              >
              ${alt ? `<div class="image-caption text-sm text-gray-600 dark:text-gray-400 text-center mt-2">${alt}</div>` : ''}
            </div>
          `;
        }
      )
      // Small images: ![alt](src){.small}
      .replace(
        /!\[([^\]]*)\]\(([^)\s]+)\)\{\.small\}/g,
        (match, alt, url) => {
          return `
            <div class="image-container my-4">
              <img 
                src="${url}" 
                alt="${alt || 'Image'}" 
                loading="lazy"
                class="max-w-48 h-auto rounded-lg shadow-sm transition-all duration-200 hover:shadow-md cursor-zoom-in"
                onclick="window.open('${url}', '_blank')"
              >
              ${alt ? `<div class="image-caption text-sm text-gray-600 dark:text-gray-400 text-center mt-2">${alt}</div>` : ''}
            </div>
          `;
        }
      );
  };

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

  // Links with title support
  html = html.replace(
    /\[([^\]]+)\]\(([^)"]+)(?:\s+"([^"]+)")?\)/g,
    (match, text, url, title) => {
      const titleAttr = title ? ` title="${title}"` : "";
      return `<a href="${url}" class="text-blue-600 dark:text-blue-400 hover:underline"${titleAttr} target="_blank" rel="noopener noreferrer">${text}</a>`;
    }
  );

  // Process images with enhanced features
  html = processImages(html);

  // Task lists
  html = html.replace(/^- \[([ x])\] (.+$)/gm, (match, checked, text) => {
    const isChecked = checked.toLowerCase() === "x";
    return `<li class="flex items-start ml-4 text-gray-800 dark:text-gray-200"><input type="checkbox" ${
      isChecked ? "checked" : ""
    } disabled class="mt-1 mr-2"><span>${text}</span></li>`;
  });

  // Unordered lists
  html = html.replace(
    /^[\*\-\+] (.+$)/gm,
    '<li class="ml-4 text-gray-800 dark:text-gray-200">• $1</li>'
  );

  // Ordered lists
  html = html.replace(
    /^\d+\. (.+$)/gm,
    '<li class="ml-4 text-gray-800 dark:text-gray-200 list-decimal">$1</li>'
  );

  // Enhanced Table Support with proper header detection
  const tableRegex = /(\|.*\|(?:\s*\n\|.*\|)*)/g;
  html = html.replace(tableRegex, (tableBlock) => {
    const rows = tableBlock
      .trim()
      .split("\n")
      .filter((row) => row.trim());

    if (rows.length < 2) return tableBlock; // Not a valid table

    const processedRows: string[] = [];
    let hasHeader = false;

    rows.forEach((row, index) => {
      const cells = row
        .slice(1, -1)
        .split("|")
        .map((cell) => cell.trim());

      // Check if this is a separator row
      if (cells.every((cell) => /^:?-+:?$/.test(cell))) {
        hasHeader = true;
        return; // Skip separator row
      }

      const isHeaderRow = hasHeader && index === 0;
      const cellTag = isHeaderRow ? "th" : "td";
      const cellClass = isHeaderRow
        ? "border border-gray-300 dark:border-gray-600 px-4 py-3 bg-gray-50 dark:bg-gray-700 font-semibold text-gray-900 dark:text-gray-100 text-left"
        : "border border-gray-300 dark:border-gray-600 px-4 py-2 text-gray-800 dark:text-gray-200";

      const rowHtml = `<tr>${cells
        .map((cell) => `<${cellTag} class="${cellClass}">${cell}</${cellTag}>`)
        .join("")}</tr>`;
      processedRows.push(rowHtml);
    });

    if (processedRows.length > 0) {
      return `<table class="border-collapse border border-gray-300 dark:border-gray-600 my-4 w-full text-sm">${processedRows.join(
        ""
      )}</table>`;
    }

    return tableBlock;
  });

  // Wrap consecutive list items in ul/ol
  html = html.replace(
    /(<li class="ml-4[^>]*>• .*?<\/li>)(?:\s*<br>\s*)?(?=<li class="ml-4[^>]*>• |$)/gs,
    "$1"
  );
  html = html.replace(
    /(<li class="ml-4[^>]*>• .*?<\/li>(?:\s*<li class="ml-4[^>]*>• .*?<\/li>)*)/gs,
    '<ul class="my-2 space-y-1">$1</ul>'
  );

  html = html.replace(
    /(<li class="ml-4[^>]*list-decimal[^>]*>.*?<\/li>)(?:\s*<br>\s*)?(?=<li class="ml-4[^>]*list-decimal|$)/gs,
    "$1"
  );
  html = html.replace(
    /(<li class="ml-4[^>]*list-decimal[^>]*>.*?<\/li>(?:\s*<li class="ml-4[^>]*list-decimal[^>]*>.*?<\/li>)*)/gs,
    '<ol class="my-2 space-y-1 list-decimal ml-8">$1</ol>'
  );

  // Paragraphs (wrap standalone text)
  html = html
    .split("<br>")
    .map((line) => {
      line = line.trim();
      if (!line) return "";
      if (
        line.startsWith("<h") ||
        line.startsWith("<ul") ||
        line.startsWith("<ol") ||
        line.startsWith("<pre") ||
        line.startsWith("<blockquote") ||
        line.startsWith("<hr") ||
        line.startsWith("<table") ||
        line.startsWith("<img") ||
        line.startsWith("<div class=\"image-container\"")
      ) {
        return line;
      }
      return `<p class="text-gray-800 dark:text-gray-200 mb-4">${line}</p>`;
    })
    .join("");

  return html;
}

export { renderMarkdown };