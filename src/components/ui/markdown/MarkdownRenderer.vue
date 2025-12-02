<!-- MarkdownRenderer.vue -->
<script setup lang="ts">
import { computed, ref } from "vue";
import hljs from "highlight.js";
import "highlight.js/styles/night-owl.css";
import CodeBlock from "./CodeBlock.vue";
import MarkdownImage from "./MarkdownImage.vue";
import MarkdownTable from "./MarkdownTable.vue";
import MarkdownCallout from "./MarkdownCallout.vue";
import MarkdownLink from "./MarkdownLink.vue";
import { inject } from "vue";
import type { Ref } from "vue";

interface Props {
    content: string;
}

const props = defineProps<Props>();
const { openPreview, showPreviewSidebar, closePreview, screenWidth } = inject(
    "globalState",
) as {
    screenWidth: Ref<number>;
    openPreview: (
        code: string,
        language: string,
        data?: {
            fileSize: string;
            wordCount: number;
            charCount: number;
        },
    ) => void;
    showPreviewSidebar: Ref<boolean>;
    closePreview: () => void;
};

interface CodeBlockData {
    id: string;
    language: string;
    code: string;
    highlighted: string;
}

interface ImageData {
    id: string;
    src: string;
    alt: string;
    title?: string;
    className?: string;
    width?: string;
    height?: string;
    style?: string;
}

interface TableData {
    id: string;
    headers: string[];
    alignments: ("left" | "right" | "center")[];
    rows: string[][];
}

interface CalloutData {
    id: string;
    type: "NOTE" | "TIP" | "IMPORTANT" | "WARNING" | "CAUTION";
    content: string;
}

interface LinkData {
    id: string;
    href: string;
    text: string;
    title?: string;
}

const codeBlocks = ref<CodeBlockData[]>([]);
const images = ref<ImageData[]>([]);
const tables = ref<TableData[]>([]);
const callouts = ref<CalloutData[]>([]);
const links = ref<LinkData[]>([]);

const escapeHtml = (unsafe: string): string => {
    return unsafe
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#039;");
};

const generateId = () => Math.random().toString(36).substring(2, 9);

// Check if code is HTML/previewable
const isPreviewableCode = (language: string, code: string): boolean => {
    const htmlLanguages = ["html", "htm", "xml"];
    return (
        htmlLanguages.includes(language.toLowerCase()) ||
        code.trim().startsWith("<!DOCTYPE") ||
        code.includes("<html") ||
        code.includes("<body")
    );
};

const processCodeBlocks = (text: string): string => {
    codeBlocks.value = [];

    return text.replace(/```(\w+)?\n([\s\S]*?)```/g, (match, lang, code) => {
        const language = lang || "plaintext";
        let highlightedCode = code;

        try {
            if (hljs.getLanguage(language)) {
                highlightedCode = hljs.highlight(code.trim(), {
                    language,
                }).value;
            } else {
                highlightedCode = hljs.highlightAuto(code.trim()).value;
            }
        } catch (error) {
            console.warn("Syntax highlighting failed:", error);
            highlightedCode = escapeHtml(code.trim());
        }

        const id = generateId();
        codeBlocks.value.push({
            id,
            language,
            code: code.trim(),
            highlighted: highlightedCode,
        });

        return `___CODE_BLOCK_${id}___`;
    });
};

const processImages = (text: string): string => {
    images.value = [];
    let result = text;

    // Clickable images: [![alt](src)](link)
    result = result.replace(
        /\[!\[([^\]]*)\]\(([^)\s]+)\)\]\(([^)\s]+)\)/g,
        (match, alt, src, link) => {
            const id = generateId();
            images.value.push({
                id,
                src,
                alt: alt || "Image",
                className: "clickable",
                style: `cursor: pointer; onclick="window.open('${link}', '_blank')"`,
            });
            return `___IMAGE_${id}___`;
        },
    );

    // Images with dimensions: ![alt](src =100x200)
    result = result.replace(
        /!\[([^\]]*)\]\(([^)\s]+)\s*=(\d+)x(\d+)\)/g,
        (match, alt, src, width, height) => {
            const id = generateId();
            images.value.push({ id, src, alt: alt || "Image", width, height });
            return `___IMAGE_${id}___`;
        },
    );

    // Images with width only: ![alt](src =100)
    result = result.replace(
        /!\[([^\]]*)\]\(([^)\s]+)\s*=(\d+)\)/g,
        (match, alt, src, width) => {
            const id = generateId();
            images.value.push({ id, src, alt: alt || "Image", width });
            return `___IMAGE_${id}___`;
        },
    );

    // Images with classes: ![alt](src){.class}
    result = result.replace(
        /!\[([^\]]*)\]\(([^)\s]+)\)\{\.([^}]+)\}/g,
        (match, alt, src, className) => {
            const id = generateId();
            images.value.push({ id, src, alt: alt || "Image", className });
            return `___IMAGE_${id}___`;
        },
    );

    // Standard markdown images: ![alt](src "title")
    result = result.replace(
        /!\[([^\]]*)\]\(([^)"]+)(?:\s+"([^"]+)")?\)/g,
        (match, alt, src, title) => {
            const id = generateId();
            images.value.push({ id, src, alt: alt || "Image", title });
            return `___IMAGE_${id}___`;
        },
    );

    return result;
};

const processTables = (text: string): string => {
    tables.value = [];
    const tableRegex = /(\|.+\|[\r\n]+\|[-:\s|]+\|[\r\n]+(?:\|.+\|[\r\n]*)+)/g;

    return text.replace(tableRegex, (match) => {
        const lines = match
            .trim()
            .split("\n")
            .map((line) => line.trim());
        if (lines.length < 2) return match;

        const headers = lines[0]
            .split("|")
            .filter((cell) => cell.trim())
            .map((cell) => cell.trim());

        const alignments = lines[1]
            .split("|")
            .filter((cell) => cell.trim())
            .map((cell) => {
                const trimmed = cell.trim();
                if (trimmed.startsWith(":") && trimmed.endsWith(":"))
                    return "center";
                if (trimmed.endsWith(":")) return "right";
                return "left";
            }) as ("left" | "right" | "center")[];

        const rows = lines.slice(2).map((line) =>
            line
                .split("|")
                .filter((cell) => cell.trim())
                .map((cell) => cell.trim()),
        );

        const id = generateId();
        tables.value.push({ id, headers, alignments, rows });

        return `___TABLE_${id}___`;
    });
};

const processCallouts = (text: string): string => {
    callouts.value = [];

    return text.replace(
        /^\[!(NOTE|TIP|IMPORTANT|WARNING|CAUTION)\]\s*\n([\s\S]*?)(?=\n\n|\n\[!|$)/gm,
        (match, type, content) => {
            const id = generateId();
            callouts.value.push({
                id,
                type: type as
                    | "NOTE"
                    | "TIP"
                    | "IMPORTANT"
                    | "WARNING"
                    | "CAUTION",
                content: content.trim(),
            });
            return `___CALLOUT_${id}___`;
        },
    );
};

const processLinks = (text: string): string => {
    links.value = [];

    return text.replace(
        /\[([^\]]+)\]\(([^)"]+)(?:\s+"([^"]+)")?\)/g,
        (match, text, href, title) => {
            const id = generateId();
            links.value.push({ id, href, text, title });
            return `___LINK_${id}___`;
        },
    );
};

const processBlockquotes = (text: string): string => {
    return text.replace(
        /^&gt; (.+)$/gm,
        '<blockquote class="border-l-4 border-blue-500 dark:border-blue-400 pl-4 my-4 text-gray-600 dark:text-gray-400 italic bg-blue-50 dark:bg-blue-900/20 py-3 rounded-r">$1</blockquote>',
    );
};

const processBasicMarkdown = (text: string): string => {
    let html = text;

    // Headers
    html = html
        .replace(
            /^###### (.*$)/gm,
            '<h6 class="text-sm text-wrap break-words font-semibold text-gray-900 dark:text-gray-100 mb-1 mt-2">$1</h6>',
        )
        .replace(
            /^##### (.*$)/gm,
            '<h5 class="text-base text-wrap break-words font-semibold text-gray-900 dark:text-gray-100 mb-2 mt-3">$1</h5>',
        )
        .replace(
            /^#### (.*$)/gm,
            '<h4 class="text-base text-wrap break-words font-semibold text-gray-900 dark:text-gray-100 mb-2 mt-3">$1</h4>',
        )
        .replace(
            /^### (.*$)/gm,
            '<h3 class="text-lg text-wrap break-words font-semibold text-gray-900 dark:text-gray-100 mb-2 mt-4">$1</h3>',
        )
        .replace(
            /^## (.*$)/gm,
            '<h2 class="text-xl text-wrap break-words font-bold text-gray-900 dark:text-gray-100 mb-3 mt-6">$1</h2>',
        )
        .replace(
            /^# (.*$)/gm,
            '<h1 class="text-2xl text-wrap break-words font-bold text-gray-900 dark:text-gray-100 mb-4 mt-8">$1</h1>',
        );

    // Horizontal rules
    html = html.replace(
        /^(---|\*\*\*|___)$/gm,
        '<hr class="border-t border-gray-300 dark:border-gray-600 my-3">',
    );

    // Bold and italic
    html = html
        .replace(
            /\*\*\*(.+?)\*\*\*/g,
            '<strong class="font-bold text-gray-900 dark:text-gray-100"><em class="italic">$1</em></strong>',
        )
        .replace(
            /\*\*(.+?)\*\*/g,
            '<strong class="font-bold text-gray-900 dark:text-gray-100">$1</strong>',
        )
        .replace(
            /\*(.+?)\*/g,
            '<em class="italic text-gray-800 dark:text-gray-200">$1</em>',
        );

    // Inline code
    html = html.replace(
        /`([^`]+)`/g,
        '<code class="inline-code bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded text-sm font-mono text-red-600 dark:text-red-400">$1</code>',
    );

    // Lists
    html = html.replace(
        /^[*\-+] (.+$)/gm,
        '<li class="ml-3 leading-relaxed">• $1</li>',
    );
    html = html.replace(
        /^\d+\. (.+$)/gm,
        '<li class="ml-3 list-decimal leading-relaxed">$1</li>',
    );

    // Wrap lists
    html = html.replace(
        /(<li class="ml-3 leading-relaxed">• .*?<\/li>(?:\s*<li class="ml-3 leading-relaxed">• .*?<\/li>)*)/gs,
        '<ul class="my-3 space-y-1 text-gray-700 dark:text-gray-300">$1</ul>',
    );
    html = html.replace(
        /(<li class="ml-3 list-decimal leading-relaxed">.*?<\/li>(?:\s*<li class="ml-3 list-decimal leading-relaxed">.*?<\/li>)*)/gs,
        '<ol class="my-3 ml-4 space-y-1 text-gray-700 dark:text-gray-300">$1</ol>',
    );

    // Blockquotes
    html = processBlockquotes(html);

    // Paragraphs
    const lines = html.split("\n");
    const processedLines: string[] = [];

    for (const line of lines) {
        const trimmedLine = line.trim();
        if (!trimmedLine) {
            processedLines.push("");
            continue;
        }

        if (
            trimmedLine.startsWith("<") ||
            trimmedLine.includes("___") ||
            line.match(/^[*\-+]\s/) ||
            line.match(/^\d+\.\s/) ||
            line.match(/^#+\s/) ||
            line.match(/^>/) ||
            line.match(/^---|^\*\*\*|^___/)
        ) {
            processedLines.push(trimmedLine);
        } else {
            processedLines.push(
                `<p class="my-3 text-wrap break-words max-w-full text-gray-700 dark:text-gray-200 leading-relaxed">${trimmedLine}</p>`,
            );
        }
    }

    return processedLines.join("\n");
};

const renderedContent = computed(() => {
    let html = props.content;

    if (!html) return "";

    // Process in order
    html = processCodeBlocks(html);
    html = processImages(html);
    html = processTables(html);
    html = processCallouts(html);
    html = processLinks(html);
    html = escapeHtml(html);
    html = processBasicMarkdown(html);

    return html;
});

const getComponentData = (placeholder: string) => {
    if (placeholder.includes("CODE_BLOCK_")) {
        const id = placeholder.replace(/___CODE_BLOCK_|___/g, "");
        return codeBlocks.value.find((block) => block.id === id);
    } else if (placeholder.includes("IMAGE_")) {
        const id = placeholder.replace(/___IMAGE_|___/g, "");
        return images.value.find((img) => img.id === id);
    } else if (placeholder.includes("TABLE_")) {
        const id = placeholder.replace(/___TABLE_|___/g, "");
        return tables.value.find((table) => table.id === id);
    } else if (placeholder.includes("CALLOUT_")) {
        const id = placeholder.replace(/___CALLOUT_|___/g, "");
        return callouts.value.find((callout) => callout.id === id);
    } else if (placeholder.includes("LINK_")) {
        const id = placeholder.replace(/___LINK_|___/g, "");
        return links.value.find((link) => link.id === id);
    }
    return null;
};

const renderPart = (part: string) => {
    if (part.includes("___CODE_BLOCK_")) return "code-block";
    else if (part.includes("___IMAGE_")) return "image";
    else if (part.includes("___TABLE_")) return "table";
    else if (part.includes("___CALLOUT_")) return "callout";
    else if (part.includes("___LINK_")) return "link";
    return "html";
};

const contentParts = computed(() => {
    const parts = renderedContent.value.split(/(___\w+_\w+___)/g);
    return parts.filter((part) => part.trim());
});

// Handle escape key to close modal
const handleKeydown = (e: KeyboardEvent) => {
    if (e.key === "Escape" && showPreviewSidebar.value) {
        closePreview();
    }
};

// Add event listener on mount
if (typeof window !== "undefined") {
    window.addEventListener("keydown", handleKeydown);
}
</script>

<template>
    <div class="markdown-renderer">
        <template v-for="(part, index) in contentParts" :key="index">
            <CodeBlock
                v-if="renderPart(part) === 'code-block'"
                :data="getComponentData(part) as CodeBlockData"
                :maxLines="screenWidth > 720 ? 7 : 5"
                :is-previewable="
                    isPreviewableCode(
                        (getComponentData(part) as CodeBlockData)?.language,
                        (getComponentData(part) as CodeBlockData)?.code,
                    )
                "
                @preview="
                    openPreview(
                        (getComponentData(part) as CodeBlockData)?.code,
                        (getComponentData(part) as CodeBlockData)?.language,
                        undefined,
                    )
                "
            />
            <MarkdownImage
                v-else-if="renderPart(part) === 'image'"
                :data="getComponentData(part) as ImageData"
            />
            <MarkdownTable
                v-else-if="renderPart(part) === 'table'"
                :data="getComponentData(part) as TableData"
            />
            <MarkdownCallout
                v-else-if="renderPart(part) === 'callout'"
                :data="getComponentData(part) as CalloutData"
            />
            <MarkdownLink
                v-else-if="renderPart(part) === 'link'"
                :data="getComponentData(part) as LinkData"
            />
            <div v-else v-html="part"></div>
        </template>
    </div>
</template>
