import hljs from "highlight.js";
import "highlight.js/styles/night-owl.css";
import { marked } from 'marked';

/**
 * Enhanced Markdown Rendering with Full Support including Advanced Image Features
 * and Syntax Highlighting
 */

// Configure marked for synchronous use
marked.setOptions({
  async: false, // Force synchronous parsing
});

marked.use({
  renderer: {
    image({ href, title, text }) {
      const titleAttr = title ? ` title="${title}"` : "";
      const isSvg = href.toLowerCase().endsWith(".svg");
      const loadingAttr = isSvg ? "" : ' loading="lazy"';

      return `
        <div class="image-container my-4">
          <img 
            src="${href}" 
            alt="${text || 'Image'}" 
            ${titleAttr}
            ${loadingAttr}
            class="max-w-[300px] h-auto rounded-lg shadow-md hover:shadow-lg transition-all duration-200 cursor-zoom-in border border-gray-200 dark:border-gray-700"
            onclick="window.open('${href}', '_blank')"
            onerror="this.style.display='none'; this.nextElementSibling.style.display='block';"
          >
          <div style="display:none;" class="text-gray-600 dark:text-gray-400 p-4 border border-gray-300 dark:border-gray-600 rounded-lg bg-gray-50 dark:bg-gray-800">
            <i class="pi pi-image text-2xl mb-2"></i>
            <p>Failed to load: ${text || 'Image'}</p>
          </div>
          ${title ? `<p class="text-sm text-gray-500 dark:text-gray-400 mt-2 text-center italic">${title}</p>` : ''}
        </div>
      `;
    },
    link({ href, title, text }) {
      const titleAttr = title ? ` title="${title}"` : "";
      return `<a href="${href}" class="text-blue-600 underline hover:text-blue-800 link-with-preview" ${titleAttr} target="_blank" rel="noopener noreferrer">${text}</a>`;
    }
  }
});

// Process iframes and embeds (YouTube, Vimeo, etc.)
const processIframes = (content: string): string => {
  return (
    content
      // YouTube embed: [!youtube](video_id) or [!youtube](full_url)
      .replace(/\[!youtube\]\(([^\s)]+)\)/g, (match, id) => {
        const videoId =
          id.includes("youtube.com") || id.includes("youtu.be")
            ? id.split(/[\/=]/).pop()
            : id;
        return `<div class="video-container my-4 aspect-video">
  <iframe src="https://www.youtube.com/embed/${videoId}" class="w-full h-full rounded-lg shadow-md border border-gray-200 dark:border-gray-700" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>
</div>`;
      })

      // Vimeo embed: [!vimeo](video_id)
      .replace(/\[!vimeo\]\(([^\s)]+)\)/g, (match, id) => {
        const videoId = id.includes("vimeo.com") ? id.split("/").pop() : id;
        return `<div class="video-container my-4 aspect-video">
  <iframe src="https://player.vimeo.com/video/${videoId}" class="w-full h-full rounded-lg shadow-md border border-gray-200 dark:border-gray-700" frameborder="0" allow="autoplay; fullscreen; picture-in-picture" allowfullscreen></iframe>
</div>`;
      })

      // Generic iframe: [!iframe](url)
      .replace(/\[!iframe\]\(([^\s)]+)\)/g, (match, url) => {
        return `<div class="iframe-container my-4">
  <iframe src="${url}" class="w-full h-96 rounded-lg shadow-md border border-gray-200 dark:border-gray-700" frameborder="0" allowfullscreen></iframe>
</div>`;
      })

      // Twitter/X embed: [!twitter](tweet_url)
      .replace(/\[!twitter\]\(([^\s)]+)\)/g, (match, url) => {
        return `<div class="twitter-container my-4 p-4 border border-gray-300 dark:border-gray-600 rounded-lg bg-gray-50 dark:bg-gray-800">
  <a href="${url}" target="_blank" rel="noopener noreferrer" class="text-blue-600 dark:text-blue-400 hover:underline">View Tweet →</a>
</div>`;
      })

      // CodePen embed: [!codepen](pen_url)
      .replace(/\[!codepen\]\(([^\s)]+)\)/g, (match, url) => {
        const penId = url.split("/pen/")[1]?.split("?")[0];
        const user = url.split(".io/")[1]?.split("/")[0];
        return `<div class="codepen-container my-4 aspect-video">
  <iframe src="https://codepen.io/${user}/embed/${penId}?default-tab=result" class="w-full h-full rounded-lg shadow-md border border-gray-200 dark:border-gray-700" frameborder="0" allowfullscreen></iframe>
</div>`;
      })
  );
};

// Process blockquotes
const processBlockquotes = (content: string): string => {
  return content.replace(/^&gt; (.+)$/gm, (match, text) => {
    return `<blockquote class="border-l-4 border-blue-500 dark:border-blue-400 pl-4 my-4 text-gray-600 dark:text-gray-400 italic bg-blue-50 dark:bg-blue-900/20 py-3 rounded-r">${text}</blockquote>`;
  });
};

// Process task lists
const processTaskLists = (content: string): string => {
  return content
    .replace(
      /^- \[ \] (.+)$/gm,
      '<li class="task-list-item flex items-start gap-2 my-1"><input type="checkbox" disabled class="mt-1 accent-gray-400 dark:accent-gray-500 cursor-not-allowed"><span class="text-gray-700 dark:text-gray-300">$1</span></li>'
    )
    .replace(
      /^- \[x\] (.+)$/gm,
      '<li class="task-list-item flex items-start gap-2 my-1"><input type="checkbox" checked disabled class="mt-1 accent-blue-500 dark:accent-blue-400 cursor-not-allowed"><span class="line-through text-gray-500 dark:text-gray-500">$1</span></li>'
    );
};

// Process callouts/alerts (GitHub style)
const processCallouts = (content: string): string => {
  const calloutTypes = {
    NOTE: { icon: "ℹ️", color: "blue", label: "Note" },
    TIP: { icon: "💡", color: "green", label: "Tip" },
    IMPORTANT: { icon: "❗", color: "purple", label: "Important" },
    WARNING: { icon: "⚠️", color: "yellow", label: "Warning" },
    CAUTION: { icon: "🚨", color: "red", label: "Caution" },
  };

  return content.replace(
    /^\[!(NOTE|TIP|IMPORTANT|WARNING|CAUTION)\]\s*\n([\s\S]*?)(?=\n\n|\n\[!|$)/gm,
    (match, type, text) => {
      const callout = calloutTypes[type as keyof typeof calloutTypes];
      return `<div class="callout my-4 p-4 border-l-4 border-${
        callout.color
      }-500 dark:border-${callout.color}-400 bg-${callout.color}-50 dark:bg-${
        callout.color
      }-900/20 rounded-r-lg">
  <div class="flex items-start gap-2">
    <span class="text-lg flex-shrink-0">${callout.icon}</span>
    <div class="flex-1">
      <div class="font-semibold text-${callout.color}-900 dark:text-${
        callout.color
      }-100 mb-1">${callout.label}</div>
      <div class="text-${callout.color}-800 dark:text-${
        callout.color
      }-200 text-sm">${text.trim()}</div>
    </div>
  </div>
</div>`;
    }
  );
};

// Process footnotes
const processFootnotes = (content: string): string => {
  const footnotes: { [key: string]: string } = {};

  // Extract footnote definitions
  content = content.replace(/^\[\^(\w+)\]:\s*(.+)$/gm, (match, id, text) => {
    footnotes[id] = text;
    return "";
  });

  // Replace footnote references
  content = content.replace(/\[\^(\w+)\]/g, (match, id) => {
    if (footnotes[id]) {
      return `<sup class="text-blue-600 dark:text-blue-400"><a href="#fn-${id}" id="fnref-${id}" title="${footnotes[id]}">[${id}]</a></sup>`;
    }
    return match;
  });

  // Add footnotes section at the end if any exist
  if (Object.keys(footnotes).length > 0) {
    let footnotesHtml =
      '<hr class="my-6 border-gray-300 dark:border-gray-600"><div class="footnotes text-sm text-gray-600 dark:text-gray-400"><h4 class="font-semibold mb-2 text-gray-900 dark:text-gray-100">Footnotes</h4><ol class="list-decimal pl-6 space-y-1">';

    Object.entries(footnotes).forEach(([id, text]) => {
      footnotesHtml += `<li id="fn-${id}" class="text-gray-700 dark:text-gray-300">${text} <a href="#fnref-${id}" class="text-blue-600 dark:text-blue-400 hover:underline">↩</a></li>`;
    });

    footnotesHtml += "</ol></div>";
    content += footnotesHtml;
  }

  return content;
};

// Process abbreviations
const processAbbreviations = (content: string): string => {
  const abbreviations: { [key: string]: string } = {};

  // Extract abbreviation definitions
  content = content.replace(
    /^\*\[([^\]]+)\]:\s*(.+)$/gm,
    (match, abbr, title) => {
      abbreviations[abbr] = title;
      return "";
    }
  );

  // Replace abbreviations with abbr tags
  Object.entries(abbreviations).forEach(([abbr, title]) => {
    const regex = new RegExp(`\\b${abbr}\\b`, "g");
    content = content.replace(
      regex,
      `<abbr title="${title}" class="border-b border-dotted border-gray-400 dark:border-gray-500 cursor-help">${abbr}</abbr>`
    );
  });

  return content;
};

// Process definition lists
const processDefinitionLists = (content: string): string => {
  return content.replace(/^(\w.+)\n:\s+(.+)$/gm, (match, term, definition) => {
    return `<dl class="my-3">
  <dt class="font-semibold text-gray-900 dark:text-gray-100">${term}</dt>
  <dd class="ml-6 text-gray-700 dark:text-gray-300">${definition}</dd>
</dl>`;
  });
};

// Process strikethrough
const processStrikethrough = (content: string): string => {
  return content.replace(
    /~~(.+?)~~/g,
    '<del class="line-through text-gray-500 dark:text-gray-500">$1</del>'
  );
};

// Process highlights/marks
const processHighlights = (content: string): string => {
  return content.replace(
    /==(.+?)==/g,
    '<mark class="bg-yellow-200 dark:bg-yellow-500/30 px-1 rounded">$1</mark>'
  );
};

// Process subscript and superscript
const processSubSup = (content: string): string => {
  return content
    .replace(/\^(.+?)\^/g, '<sup class="text-xs">$1</sup>')
    .replace(/~(.+?)~/g, '<sub class="text-xs">$1</sub>');
};

// Process keyboard keys
const processKeyboard = (content: string): string => {
  return content.replace(
    /\[\[(.+?)\]\]/g,
    '<kbd class="px-2 py-1 text-xs font-mono bg-gray-100 dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded shadow-sm">$1</kbd>'
  );
};

// Process emojis (shortcodes)
const processEmojis = (content: string): string => {
  const emojiMap: { [key: string]: string } = {
    ":smile:": "😊",
    ":heart:": "❤️",
    ":thumbsup:": "👍",
    ":fire:": "🔥",
    ":rocket:": "🚀",
    ":star:": "⭐",
    ":check:": "✅",
    ":x:": "❌",
    ":warning:": "⚠️",
    ":info:": "ℹ️",
    // Add more as needed
  };

  Object.entries(emojiMap).forEach(([code, emoji]) => {
    content = content.replace(new RegExp(code, "g"), emoji);
  });

  return content;
};

// Process math (basic LaTeX-style)
const processMath = (content: string): string => {
  // Inline math: $...$
  content = content.replace(
    /\$([^\$]+)\$/g,
    '<span class="math inline-math font-mono text-sm bg-gray-100 dark:bg-gray-800 px-1 rounded">$1</span>'
  );

  // Block math: $$...$$
  content = content.replace(
    /\$\$([\s\S]+?)\$\$/g,
    '<div class="math block-math my-4 p-4 bg-gray-100 dark:bg-gray-800 rounded-lg overflow-x-auto"><code class="font-mono text-sm">$1</code></div>'
  );

  return content;
};

// ![Logo](logo.png)                     # Standard
// ![Avatar](user.jpg){.circle}          # Circular
// ![Banner](hero.jpg){.full}            # Full width
// ![Thumb](thumb.jpg){.small}           # Small
// ![Icon](icon.svg =50)                 # 50px width
// ![Photo](photo.jpg =800x600)          # Fixed dimensions
// ![Doc](doc.png){.float-left}          # Float left
// [![Click](img.jpg)](https://link.com) # Clickable
// ![Alt](img.png "Title")               # With title

// Enhanced Image Support with multiple formats and features
const processImages = (content: string): string => {
  return (
    content
      // 1. Clickable images: [![alt](src)](link)
      .replace(
        /\[!\[([^\]]*)\]\(([^)\s]+)\)\]\(([^)\s]+)\)/g,
        (match, alt, imageUrl, linkUrl) => {
          return `<div class="image-container my-4">
  <a href="${linkUrl}" target="_blank" rel="noopener noreferrer" class="block">
    <img src="${imageUrl}" alt="${
            alt || "Image"
          }" loading="lazy" class="max-w-full h-auto rounded-lg shadow-md hover:shadow-lg transition-all duration-200 cursor-pointer border border-gray-200 dark:border-gray-700" onerror="this.style.display='none'; this.nextElementSibling.style.display='block';">
    <div style="display:none;" class="text-gray-600 dark:text-gray-400 p-4 border border-gray-300 dark:border-gray-600 rounded-lg bg-gray-50 dark:bg-gray-800">
      <i class="pi pi-image text-2xl mb-2"></i>
      <p>Failed to load: ${alt || "Image"}</p>
    </div>
  </a>
</div>`;
        }
      )

      // 2. Images with width/height specification: ![alt](src =100x200)
      .replace(
        /!\[([^\]]*)\]\(([^)\s]+)\s*=(\d+)x(\d+)\)/g,
        (match, alt, url, width, height) => {
          return `<div class="image-container my-4">
  <img src="${url}" alt="${
            alt || "Image"
          }" width="${width}" height="${height}" loading="lazy" class="rounded-lg shadow-md hover:shadow-lg transition-all duration-200 cursor-zoom-in border border-gray-200 dark:border-gray-700" onclick="window.open('${url}', '_blank')" onerror="this.style.display='none'; this.nextElementSibling.style.display='block';">
  <div style="display:none;" class="text-gray-600 dark:text-gray-400 p-4 border border-gray-300 dark:border-gray-600 rounded-lg bg-gray-50 dark:bg-gray-800">
    <i class="pi pi-image text-2xl mb-2"></i>
    <p>Failed to load: ${alt || "Image"}</p>
  </div>
</div>`;
        }
      )

      // 3. Images with width only: ![alt](src =100)
      .replace(
        /!\[([^\]]*)\]\(([^)\s]+)\s*=(\d+)\)/g,
        (match, alt, url, width) => {
          return `<div class="image-container my-4">
  <img src="${url}" alt="${
            alt || "Image"
          }" width="${width}" loading="lazy" class="h-auto rounded-lg shadow-md hover:shadow-lg transition-all duration-200 cursor-zoom-in border border-gray-200 dark:border-gray-700" onclick="window.open('${url}', '_blank')" onerror="this.style.display='none'; this.nextElementSibling.style.display='block';">
  <div style="display:none;" class="text-gray-600 dark:text-gray-400 p-4 border border-gray-300 dark:border-gray-600 rounded-lg bg-gray-50 dark:bg-gray-800">
    <i class="pi pi-image text-2xl mb-2"></i>
    <p>Failed to load: ${alt || "Image"}</p>
  </div>
</div>`;
        }
      )

      // 4. Float left images: ![alt](src){.float-left}
      .replace(
        /!\[([^\]]*)\]\(([^)\s]+)\)\{\.float-left\}/g,
        (match, alt, url) => {
          return `<div class="image-container my-2 float-left mr-4 mb-4 max-w-xs">
  <img src="${url}" alt="${
            alt || "Image"
          }" loading="lazy" class="w-full h-auto rounded-lg shadow-md hover:shadow-lg transition-all duration-200 cursor-zoom-in border border-gray-200 dark:border-gray-700" onclick="window.open('${url}', '_blank')" onerror="this.style.display='none'; this.nextElementSibling.style.display='block';">
  <div style="display:none;" class="text-gray-600 dark:text-gray-400 p-4 border border-gray-300 dark:border-gray-600 rounded-lg bg-gray-50 dark:bg-gray-800">
    <i class="pi pi-image text-2xl mb-2"></i>
    <p>Failed to load: ${alt || "Image"}</p>
  </div>
</div>`;
        }
      )

      // 5. Float right images: ![alt](src){.float-right}
      .replace(
        /!\[([^\]]*)\]\(([^)\s]+)\)\{\.float-right\}/g,
        (match, alt, url) => {
          return `<div class="image-container my-2 float-right ml-4 mb-4 max-w-xs">
  <img src="${url}" alt="${
            alt || "Image"
          }" loading="lazy" class="w-full h-auto rounded-lg shadow-md hover:shadow-lg transition-all duration-200 cursor-zoom-in border border-gray-200 dark:border-gray-700" onclick="window.open('${url}', '_blank')" onerror="this.style.display='none'; this.nextElementSibling.style.display='block';">
  <div style="display:none;" class="text-gray-600 dark:text-gray-400 p-4 border border-gray-300 dark:border-gray-600 rounded-lg bg-gray-50 dark:bg-gray-800">
    <i class="pi pi-image text-2xl mb-2"></i>
    <p>Failed to load: ${alt || "Image"}</p>
  </div>
</div>`;
        }
      )

      // 6. Centered images: ![alt](src){.center}
      .replace(/!\[([^\]]*)\]\(([^)\s]+)\)\{\.center\}/g, (match, alt, url) => {
        return `<div class="image-container my-4 flex justify-center">
  <img src="${url}" alt="${
          alt || "Image"
        }" loading="lazy" class="max-w-full h-auto rounded-lg shadow-md hover:shadow-lg transition-all duration-200 cursor-zoom-in border border-gray-200 dark:border-gray-700" onclick="window.open('${url}', '_blank')" onerror="this.style.display='none'; this.nextElementSibling.style.display='block';">
  <div style="display:none;" class="text-gray-600 dark:text-gray-400 p-4 border border-gray-300 dark:border-gray-600 rounded-lg bg-gray-50 dark:bg-gray-800">
    <i class="pi pi-image text-2xl mb-2"></i>
    <p>Failed to load: ${alt || "Image"}</p>
  </div>
</div>`;
      })

      // 7. Small images: ![alt](src){.small}
      .replace(/!\[([^\]]*)\]\(([^)\s]+)\)\{\.small\}/g, (match, alt, url) => {
        return `<div class="image-container my-4">
  <img src="${url}" alt="${
          alt || "Image"
        }" loading="lazy" class="max-w-48 h-auto rounded-lg shadow-md hover:shadow-lg transition-all duration-200 cursor-zoom-in border border-gray-200 dark:border-gray-700" onclick="window.open('${url}', '_blank')" onerror="this.style.display='none'; this.nextElementSibling.style.display='block';">
  <div style="display:none;" class="text-gray-600 dark:text-gray-400 p-4 border border-gray-300 dark:border-gray-600 rounded-lg bg-gray-50 dark:bg-gray-800">
    <i class="pi pi-image text-2xl mb-2"></i>
    <p>Failed to load: ${alt || "Image"}</p>
  </div>
</div>`;
      })

      // 8. Medium images: ![alt](src){.medium}
      .replace(/!\[([^\]]*)\]\(([^)\s]+)\)\{\.medium\}/g, (match, alt, url) => {
        return `<div class="image-container my-4">
  <img src="${url}" alt="${
          alt || "Image"
        }" loading="lazy" class="max-w-md h-auto rounded-lg shadow-md hover:shadow-lg transition-all duration-200 cursor-zoom-in border border-gray-200 dark:border-gray-700" onclick="window.open('${url}', '_blank')" onerror="this.style.display='none'; this.nextElementSibling.style.display='block';">
  <div style="display:none;" class="text-gray-600 dark:text-gray-400 p-4 border border-gray-300 dark:border-gray-600 rounded-lg bg-gray-50 dark:bg-gray-800">
    <i class="pi pi-image text-2xl mb-2"></i>
    <p>Failed to load: ${alt || "Image"}</p>
  </div>
</div>`;
      })

      // 9. Large images: ![alt](src){.large}
      .replace(/!\[([^\]]*)\]\(([^)\s]+)\)\{\.large\}/g, (match, alt, url) => {
        return `<div class="image-container my-4">
  <img src="${url}" alt="${
          alt || "Image"
        }" loading="lazy" class="max-w-4xl h-auto rounded-lg shadow-md hover:shadow-lg transition-all duration-200 cursor-zoom-in border border-gray-200 dark:border-gray-700" onclick="window.open('${url}', '_blank')" onerror="this.style.display='none'; this.nextElementSibling.style.display='block';">
  <div style="display:none;" class="text-gray-600 dark:text-gray-400 p-4 border border-gray-300 dark:border-gray-600 rounded-lg bg-gray-50 dark:bg-gray-800">
    <i class="pi pi-image text-2xl mb-2"></i>
    <p>Failed to load: ${alt || "Image"}</p>
  </div>
</div>`;
      })

      // 10. Full width images: ![alt](src){.full}
      .replace(/!\[([^\]]*)\]\(([^)\s]+)\)\{\.full\}/g, (match, alt, url) => {
        return `<div class="image-container my-4 w-full">
  <img src="${url}" alt="${
          alt || "Image"
        }" loading="lazy" class="w-full h-auto rounded-lg shadow-md hover:shadow-lg transition-all duration-200 cursor-zoom-in border border-gray-200 dark:border-gray-700" onclick="window.open('${url}', '_blank')" onerror="this.style.display='none'; this.nextElementSibling.style.display='block';">
  <div style="display:none;" class="text-gray-600 dark:text-gray-400 p-4 border border-gray-300 dark:border-gray-600 rounded-lg bg-gray-50 dark:bg-gray-800">
    <i class="pi pi-image text-2xl mb-2"></i>
    <p>Failed to load: ${alt || "Image"}</p>
  </div>
</div>`;
      })

      // 11. Rounded circle images: ![alt](src){.circle}
      .replace(/!\[([^\]]*)\]\(([^)\s]+)\)\{\.circle\}/g, (match, alt, url) => {
        return `<div class="image-container my-4 flex justify-center">
  <img src="${url}" alt="${
          alt || "Image"
        }" loading="lazy" class="w-32 h-32 rounded-full object-cover shadow-md hover:shadow-lg transition-all duration-200 cursor-zoom-in border-2 border-gray-200 dark:border-gray-700" onclick="window.open('${url}', '_blank')" onerror="this.style.display='none'; this.nextElementSibling.style.display='block';">
  <div style="display:none;" class="text-gray-600 dark:text-gray-400 p-4 border border-gray-300 dark:border-gray-600 rounded-lg bg-gray-50 dark:bg-gray-800">
    <i class="pi pi-image text-2xl mb-2"></i>
    <p>Failed to load: ${alt || "Image"}</p>
  </div>
</div>`;
      })

      // 12. Border images: ![alt](src){.border}
      .replace(/!\[([^\]]*)\]\(([^)\s]+)\)\{\.border\}/g, (match, alt, url) => {
        return `<div class="image-container my-4">
  <img src="${url}" alt="${
          alt || "Image"
        }" loading="lazy" class="max-w-full h-auto rounded-lg border-4 border-gray-300 dark:border-gray-600 shadow-md hover:shadow-lg transition-all duration-200 cursor-zoom-in" onclick="window.open('${url}', '_blank')" onerror="this.style.display='none'; this.nextElementSibling.style.display='block';">
  <div style="display:none;" class="text-gray-600 dark:text-gray-400 p-4 border border-gray-300 dark:border-gray-600 rounded-lg bg-gray-50 dark:bg-gray-800">
    <i class="pi pi-image text-2xl mb-2"></i>
    <p>Failed to load: ${alt || "Image"}</p>
  </div>
</div>`;
      })

      // 13. Shadow images: ![alt](src){.shadow}
      .replace(/!\[([^\]]*)\]\(([^)\s]+)\)\{\.shadow\}/g, (match, alt, url) => {
        return `<div class="image-container my-4">
  <img src="${url}" alt="${
          alt || "Image"
        }" loading="lazy" class="max-w-full h-auto rounded-lg shadow-2xl hover:shadow-3xl transition-all duration-200 cursor-zoom-in" onclick="window.open('${url}', '_blank')" onerror="this.style.display='none'; this.nextElementSibling.style.display='block';">
  <div style="display:none;" class="text-gray-600 dark:text-gray-400 p-4 border border-gray-300 dark:border-gray-600 rounded-lg bg-gray-50 dark:bg-gray-800">
    <i class="pi pi-image text-2xl mb-2"></i>
    <p>Failed to load: ${alt || "Image"}</p>
  </div>
</div>`;
      })

      // 14. Images with custom CSS class (catch-all for other classes): ![alt](src){.custom-class}
      .replace(
        /!\[([^\]]*)\]\(([^)\s]+)\)\{\.([^}]+)\}/g,
        (match, alt, url, className) => {
          return `<div class="image-container my-4">
  <img src="${url}" alt="${
            alt || "Image"
          }" loading="lazy" class="${className} max-w-full h-auto rounded-lg shadow-md hover:shadow-lg transition-all duration-200 cursor-zoom-in" onclick="window.open('${url}', '_blank')" onerror="this.style.display='none'; this.nextElementSibling.style.display='block';">
  <div style="display:none;" class="text-gray-600 dark:text-gray-400 p-4 border border-gray-300 dark:border-gray-600 rounded-lg bg-gray-50 dark:bg-gray-800">
    <i class="pi pi-image text-2xl mb-2"></i>
    <p>Failed to load: ${alt || "Image"}</p>
  </div>
</div>`;
        }
      )

      // 15. Standard Markdown images with title: ![alt](src "title") - MUST BE LAST
      .replace(
        /!\[([^\]]*)\]\(([^)"]+)(?:\s+"([^"]+)")?\)/g,
        (match, alt, url, title) => {
          const titleAttr = title ? ` title="${title}"` : "";
          const isSvg = url.toLowerCase().endsWith(".svg");
          const loadingAttr = isSvg ? "" : ' loading="lazy"';

          return `<div class="image-container my-4">
  <img src="${url}" alt="${
            alt || "Image"
          }"${titleAttr}${loadingAttr} class="max-w-full max-h-auto max-md:h-[300px] max-md:object-cover rounded-lg shadow-md hover:shadow-lg transition-all duration-200 cursor-zoom-in border border-gray-200 dark:border-gray-700" onclick="window.open('${url}', '_blank')" onerror="this.style.display='none'; this.nextElementSibling.style.display='block';">
  <div style="display:none;" class="text-gray-600 dark:text-gray-400 p-4 border border-gray-300 dark:border-gray-600 rounded-lg bg-gray-50 dark:bg-gray-800">
    <i class="pi pi-image text-2xl mb-2"></i>
    <p>Failed to load: ${alt || "Image"}</p>
  </div>
  ${
    title
      ? `<p class="text-sm text-gray-500 dark:text-gray-400 mt-2 text-center italic">${title}</p>`
      : ""
  }
</div>`;
        }
      )
  );
};

// Process tables
const processTable = (content: string): string => {
  const tableRegex = /(\|.+\|[\r\n]+\|[-:\s|]+\|[\r\n]+(?:\|.+\|[\r\n]*)+)/g;

  return content.replace(tableRegex, (match) => {
    const lines = match
      .trim()
      .split("\n")
      .map((line) => line.trim());

    if (lines.length < 2) return match;

    // Parse header
    const headerCells = lines[0]
      .split("|")
      .filter((cell) => cell.trim())
      .map((cell) => cell.trim());

    // Parse alignment from separator row
    const alignments = lines[1]
      .split("|")
      .filter((cell) => cell.trim())
      .map((cell) => {
        const trimmed = cell.trim();
        if (trimmed.startsWith(":") && trimmed.endsWith(":")) return "center";
        if (trimmed.endsWith(":")) return "right";
        return "left";
      });

    // Parse body rows
    const bodyRows = lines.slice(2).map((line) =>
      line
        .split("|")
        .filter((cell) => cell.trim())
        .map((cell) => cell.trim())
    );

    // Build table HTML
    let tableHtml =
      '<div class="table-container my-4 overflow-x-auto"><table class="w-full border-collapse border border-gray-300 dark:border-gray-600">';

    // Header
    tableHtml += '<thead class="bg-gray-100 dark:bg-gray-800"><tr>';
    headerCells.forEach((cell, i) => {
      const align = alignments[i] || "left";
      tableHtml += `<th class="border border-gray-300 dark:border-gray-600 px-4 py-2 text-${align} font-semibold text-gray-900 dark:text-gray-100">${cell}</th>`;
    });
    tableHtml += "</tr></thead>";

    // Body
    tableHtml += "<tbody>";
    bodyRows.forEach((row) => {
      tableHtml += '<tr class="hover:bg-gray-50 dark:hover:bg-gray-700/50">';
      row.forEach((cell, i) => {
        const align = alignments[i] || "left";
        tableHtml += `<td class="border border-gray-300 dark:border-gray-600 px-4 py-2 text-${align} text-gray-800 dark:text-gray-200">${cell}</td>`;
      });
      tableHtml += "</tr>";
    });
    tableHtml += "</tbody></table></div>";

    return tableHtml;
  });
};

const processHeaders = (content: string): string =>{
  return content.replace(
      /^###### (.*$)/gm,
      '<h6 class="text-sm text-wrap break-words font-semibold text-gray-900 dark:text-gray-100 mb-1 mt-2">$1</h6>'
    )
    .replace(
      /^##### (.*$)/gm,
      '<h5 class="text-base text-wrap break-words font-semibold text-gray-900 dark:text-gray-100 mb-2 mt-3">$1</h5>'
    )
    .replace(
      /^#### (.*$)/gm,
      '<h4 class="text-base text-wrap break-words font-semibold text-gray-900 dark:text-gray-100 mb-2 mt-3">$1</h4>'
    )
    .replace(
      /^### (.*$)/gm,
      '<h3 class="text-lg text-wrap break-words font-semibold text-gray-900 dark:text-gray-100 mb-2 mt-4">$1</h3>'
    )
    .replace(
      /^## (.*$)/gm,
      '<h2 class="text-xl text-wrap break-words font-bold text-gray-900 dark:text-gray-100 mb-3 mt-6">$1</h2>'
    )
    .replace(
      /^# (.*$)/gm,
      '<h1 class="text-2xl text-wrap break-words font-bold text-gray-900 dark:text-gray-100 mb-4 mt-8">$1</h1>'
    );
}

const horizontalRules=(content: string): string =>{
  return content.replace(
    /^(---|\*\*\*|___)$/gm,
    '<hr class="border-t border-gray-300 dark:border-gray-600 my-3">'
  );
}

// Links with title support 
const processLinks = (content: string): string =>{
  return content.replace(
    /\[([^\]]+)\]\(([^)"]+)(?:\s+"([^"]+)")?\)/g,
    (match, text, url, title) => {
      const titleAttr = title ? ` title="${title}"` : "";
      return `<a href="${url}" class="text-blue-600 underline hover:text-blue-800 link-with-preview" ${titleAttr} target="_blank" rel="noopener noreferrer">${
        text.length>60? text.slice(0,60) + "...":text
      }</a>`;
    }
  );
}

const inlineCode =(escapeHtml: (unsafe: string)=> string,inlineCodePlaceholders: string[],content: string): string=>{
  return content.replace(/`([^`]+)`/g, (match, code) => {
    const placeholder = `___INLINE_CODE_${inlineCodePlaceholders.length}___`;
    inlineCodePlaceholders.push(
      `<code class="inline-code bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded text-sm font-mono text-red-600 dark:text-red-400">${escapeHtml(
        code
      )}</code>`
    );
    return placeholder;
  })
}

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

  // Protect HTML tags before escaping
  const htmlTagPlaceholders: string[] = [];
  html = html.replace(/<(\w+)[^>]*>[\s\S]*?<\/\1>|<[^>]+\/>/g, (match) => {
    const placeholder = `___HTML_TAG_${htmlTagPlaceholders.length}___`;
    htmlTagPlaceholders.push(match);
    return placeholder;
  });

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

    // <span class="text-xs font-mono text-gray-400 capitalize">${language}</span>
    codeBlockPlaceholders.push(`
      <div class="code-container relative my-4">
        <pre class="bg-gray-900 rounded-lg overflow-x-auto"><code class="hljs language-${
          language || "plaintext"
        } text-sm">${highlightedCode}</code></pre>
        <button class="copy-button absolute top-2 right-2 bg-gray-700 text-white px-3 py-1 rounded text-xs hover:bg-gray-600 transition-colors"
          data-code="${encodeURIComponent(code.trim())}"
        >Copy</button>
      </div> 
    `);

    return placeholder;
  });

  // Process inline code (`code`)
  html = inlineCode(escapeHtml, inlineCodePlaceholders, html);
  
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

  // Protect HTML tag placeholders
  html = html.replace(/(___HTML_TAG_\d+___)/g, (match) => {
    protectedParts.push(match);
    return `___PROTECTED_${protectedParts.length - 1}___`;
  });

  html = escapeHtml(html);

  // Restore protected placeholders
  html = html.replace(/___PROTECTED_(\d+)___/g, (match, index) => {
    return protectedParts[parseInt(index)] || match;
  });

  // Restore HTML tags
  html = html.replace(/___HTML_TAG_(\d+)___/g, (match, index) => {
    return htmlTagPlaceholders[parseInt(index)] || match;
  });

  html = processFootnotes(html); // 1. Footnotes (extract definitions first)
  html = processAbbreviations(html); // 2. Abbreviations (extract definitions)
  html = processImages(html); // 3. Images (before other markdown)
  html = processLinks(html);
  html = processIframes(html); // 4. Iframes and embeds
  html = processCallouts(html); // 5. Callouts/alerts
  html = processTable(html); // 6. Tables
  html = processBlockquotes(html); // 7. Blockquotes
  html = processTaskLists(html); // 8. Task lists
  html = processDefinitionLists(html); // 9. Definition lists

  // Process basic markdown with marked.js (including images)
  // html = marked.parse(html) as string;


  // Headers (H1-H6)  
  html = processHeaders(html);

  // Horizontal rules  
  html = horizontalRules(html)

  // Inline formatting
  html = processStrikethrough(html); // Strikethrough
  html = processHighlights(html); // Highlights
  html = processSubSup(html); // Sub/superscript
  html = processKeyboard(html); // Keyboard keys
  html = processEmojis(html); // Emoji shortcodes
  html = processMath(html); // Math expressions

  // Bold and italic  
  html = html
    .replace(
      /\*\*\*(.+?)\*\*\*/g,
      '<strong class="font-bold text-gray-900 dark:text-gray-100"><em class="italic">$1</em></strong>'
    )
    .replace(
      /\*\*(.+?)\*\*/g,
      '<strong class="font-bold text-gray-900 dark:text-gray-100">$1</strong>'
    )
    .replace(
      /\*(.+?)\*/g,
      '<em class="italic text-gray-800 dark:text-gray-200">$1</em>'
    );

  // Lists  
  html = html.replace(
    /^[\*\-\+] (.+$)/gm,
    '<li class="ml-4 leading-relaxed">• $1</li>'
  );
  html = html.replace(
    /^\d+\. (.+$)/gm,
    '<li class="ml-4 list-decimal leading-relaxed">$1</li>'
  );

  // Wrap lists
  html = html.replace(
    /(<li class="ml-4 leading-relaxed">• .*?<\/li>(?:\s*<li class="ml-4 leading-relaxed">• .*?<\/li>)*)/gs,
    '<ul class="my-3 space-y-1 text-gray-700 dark:text-gray-300">$1</ul>'
  );
  html = html.replace(
    /(<li class="ml-4 list-decimal leading-relaxed">.*?<\/li>(?:\s*<li class="ml-4 list-decimal leading-relaxed">.*?<\/li>)*)/gs,
    '<ol class="my-3 ml-8 space-y-1 text-gray-700 dark:text-gray-300">$1</ol>'
  );

  // Paragraphs
  const lines = html.split("\n");
  const processedLines: string[] = [];

  for (const line of lines) {
    const trimmedLine = line.trim();
    if (!trimmedLine) {
      processedLines.push("");
      continue;
    }

    // Skip if it's already HTML or a placeholder
    if (
      trimmedLine.startsWith("<") ||
      trimmedLine.includes("___CODE_BLOCK_") ||
      trimmedLine.includes("___INLINE_CODE_") ||
      line.trim() === '' ||
      line.match(/^[\*\-\+]\s/) || // List items
      line.match(/^\d+\.\s/) ||    // Ordered list items
      line.match(/^#+\s/) ||       // Headers
      line.match(/^>/) ||          // Blockquotes
      line.match(/^---|^\*\*\*|^___/) // Horizontal rules
    ) {
      processedLines.push(trimmedLine);
    } else {
      processedLines.push(
        `<p class="my-3 text-wrap break-words max-w-full text-gray-700 dark:text-gray-200 leading-relaxed">${trimmedLine}</p>`
      );
    }
  }

  html = processedLines.join("\n");

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
