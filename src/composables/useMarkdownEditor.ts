import { nextTick } from "vue";
import type { EditableContent } from "@/types/document";
import type { Ref } from "vue";

export function useMarkdownEditor(
  editablePages: Ref<EditableContent[]>,
  currentPage: Ref<number>,
  saveToHistory: () => void
) {
  function updatePageContent(
    content: string,
    shouldSaveHistory: boolean = true
  ) {
    const pageIndex = currentPage.value - 1;
    if (pageIndex >= 0 && pageIndex < editablePages.value.length) {
      if (
        shouldSaveHistory &&
        editablePages.value[pageIndex].content !== content
      ) {
        saveToHistory();
      }

      editablePages.value[pageIndex].content = content;
      editablePages.value[pageIndex].isModified =
        content !== editablePages.value[pageIndex].originalContent;
    }
  }

  function insertMarkdown(
    before: string,
    after: string = "",
    placeholder: string = ""
  ) {
    const textarea = document.querySelector("textarea") as HTMLTextAreaElement;
    if (!textarea) return;

    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const selectedText = textarea.value.substring(start, end);
    const replacement = selectedText || placeholder;

    const newText = before + replacement + after;
    const newContent =
      textarea.value.substring(0, start) +
      newText +
      textarea.value.substring(end);

    saveToHistory();
    updatePageContent(newContent);

    nextTick(() => {
    //   textarea.focus();
      if (selectedText) {
        textarea.setSelectionRange(
          start + before.length,
          start + before.length + replacement.length
        );
      } else {
        textarea.setSelectionRange(
          start + before.length,
          start + before.length + placeholder.length
        );
      }
    });
  }

  function insertBold() {
    insertMarkdown("**", "**", "bold text");
  }

  function insertItalic() {
    insertMarkdown("*", "*", "italic text");
  }

  function insertStrikethrough() {
    insertMarkdown("~~", "~~", "strikethrough text");
  }

  function insertHighlight() {
    insertMarkdown("==", "==", "highlighted text");
  }

  function insertCode() {
    insertMarkdown("`", "`", "code");
  }

  function insertCodeBlock() {
    const textarea = document.querySelector("textarea") as HTMLTextAreaElement;
    if (!textarea) return;

    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const selectedText = textarea.value.substring(start, end);
    const beforeCursor = textarea.value.substring(0, start);
    const needsNewline =
      beforeCursor.length > 0 && !beforeCursor.endsWith("\n");

    const prefix = (needsNewline ? "\n" : "") + "```\n";
    const suffix = "\n```";
    const placeholder = "code block";

    const replacement = selectedText || placeholder;
    const newText = prefix + replacement + suffix;
    const newContent =
      textarea.value.substring(0, start) +
      newText +
      textarea.value.substring(end);

    saveToHistory();
    updatePageContent(newContent);

    nextTick(() => {
    //   textarea.focus();
      const offset = prefix.length;
      textarea.setSelectionRange(
        start + offset,
        start + offset + replacement.length
      );
    });
  }

  function insertLink() {
    const textarea = document.querySelector("textarea") as HTMLTextAreaElement;
    if (!textarea) return;

    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const selectedText = textarea.value.substring(start, end);

    const linkText = selectedText || "link text";
    const newText = `[${linkText}](url)`;
    const newContent =
      textarea.value.substring(0, start) +
      newText +
      textarea.value.substring(end);

    saveToHistory();
    updatePageContent(newContent);

    nextTick(() => {
    //   textarea.focus();
      const urlStart = start + linkText.length + 3;
      const urlEnd = urlStart + 3;
      textarea.setSelectionRange(urlStart, urlEnd);
    });
  }

  function insertImage() {
    const textarea = document.querySelector("textarea") as HTMLTextAreaElement;
    if (!textarea) return;

    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const selectedText = textarea.value.substring(start, end);

    const altText = selectedText || "image description";
    const newText = `![${altText}](image-url)`;
    const newContent =
      textarea.value.substring(0, start) +
      newText +
      textarea.value.substring(end);

    saveToHistory();
    updatePageContent(newContent);

    nextTick(() => {
    //   textarea.focus();
      const urlStart = start + altText.length + 4;
      const urlEnd = urlStart + 9;
      textarea.setSelectionRange(urlStart, urlEnd);
    });
  }

  function insertImageWithTitle() {
    const textarea = document.querySelector("textarea") as HTMLTextAreaElement;
    if (!textarea) return;

    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const selectedText = textarea.value.substring(start, end);

    const altText = selectedText || "image description";
    const newText = `![${altText}](image-url "optional title")`;
    const newContent =
      textarea.value.substring(0, start) +
      newText +
      textarea.value.substring(end);

    saveToHistory();
    updatePageContent(newContent);

    nextTick(() => {
    //   textarea.focus();
      const urlStart = start + altText.length + 4;
      const urlEnd = urlStart + 9;
      textarea.setSelectionRange(urlStart, urlEnd);
    });
  }

  function insertImageWithDimensions() {
    const textarea = document.querySelector("textarea") as HTMLTextAreaElement;
    if (!textarea) return;

    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const selectedText = textarea.value.substring(start, end);

    const altText = selectedText || "image description";
    const newText = `![${altText}](image-url =800x600)`;
    const newContent =
      textarea.value.substring(0, start) +
      newText +
      textarea.value.substring(end);

    saveToHistory();
    updatePageContent(newContent);

    nextTick(() => {
    //   textarea.focus();
      const urlStart = start + altText.length + 4;
      const urlEnd = urlStart + 9;
      textarea.setSelectionRange(urlStart, urlEnd);
    });
  }

  function insertImageCentered() {
    const textarea = document.querySelector("textarea") as HTMLTextAreaElement;
    if (!textarea) return;

    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const selectedText = textarea.value.substring(start, end);

    const altText = selectedText || "image description";
    const newText = `![${altText}](image-url){.center}`;
    const newContent =
      textarea.value.substring(0, start) +
      newText +
      textarea.value.substring(end);

    saveToHistory();
    updatePageContent(newContent);

    nextTick(() => {
    //   textarea.focus();
      const urlStart = start + altText.length + 4;
      const urlEnd = urlStart + 9;
      textarea.setSelectionRange(urlStart, urlEnd);
    });
  }

  function insertImageSmall() {
    const textarea = document.querySelector("textarea") as HTMLTextAreaElement;
    if (!textarea) return;

    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const selectedText = textarea.value.substring(start, end);

    const altText = selectedText || "image description";
    const newText = `![${altText}](image-url){.small}`;
    const newContent =
      textarea.value.substring(0, start) +
      newText +
      textarea.value.substring(end);

    saveToHistory();
    updatePageContent(newContent);

    nextTick(() => {
    //   textarea.focus();
      const urlStart = start + altText.length + 4;
      const urlEnd = urlStart + 9;
      textarea.setSelectionRange(urlStart, urlEnd);
    });
  }

  function insertImageWithBorder() {
    const textarea = document.querySelector("textarea") as HTMLTextAreaElement;
    if (!textarea) return;

    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const selectedText = textarea.value.substring(start, end);

    const altText = selectedText || "image description";
    const newText = `![${altText}](image-url){.border}`;
    const newContent =
      textarea.value.substring(0, start) +
      newText +
      textarea.value.substring(end);

    saveToHistory();
    updatePageContent(newContent);

    nextTick(() => {
    //   textarea.focus();
      const urlStart = start + altText.length + 4;
      const urlEnd = urlStart + 9;
      textarea.setSelectionRange(urlStart, urlEnd);
    });
  }

  function insertImageLink() {
    const textarea = document.querySelector("textarea") as HTMLTextAreaElement;
    if (!textarea) return;

    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const selectedText = textarea.value.substring(start, end);

    const altText = selectedText || "image description";
    const newText = `[![${altText}](image-url)](https://example.com)`;
    const newContent =
      textarea.value.substring(0, start) +
      newText +
      textarea.value.substring(end);

    saveToHistory();
    updatePageContent(newContent);

    nextTick(() => {
    //   textarea.focus();
      const urlStart = start + altText.length + 6;
      const urlEnd = urlStart + 9;
      textarea.setSelectionRange(urlStart, urlEnd);
    });
  }

  function insertHeader(level: number) {
    const textarea = document.querySelector("textarea") as HTMLTextAreaElement;
    if (!textarea) return;

    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const selectedText = textarea.value.substring(start, end);
    const beforeCursor = textarea.value.substring(0, start);
    const needsNewline =
      beforeCursor.length > 0 && !beforeCursor.endsWith("\n");

    const prefix = (needsNewline ? "\n" : "") + "#".repeat(level) + " ";
    const placeholder = `Header ${level}`;
    const replacement = selectedText || placeholder;

    const newText = prefix + replacement;
    const newContent =
      textarea.value.substring(0, start) +
      newText +
      textarea.value.substring(end);

    saveToHistory();
    updatePageContent(newContent);

    nextTick(() => {
      textarea.focus();
      textarea.setSelectionRange(
        start + prefix.length,
        start + prefix.length + replacement.length
      );
    });
  }

  function insertLinePrefix(prefix: string, placeholder: string) {
    const textarea = document.querySelector("textarea") as HTMLTextAreaElement;
    if (!textarea) return;

    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const selectedText = textarea.value.substring(start, end);
    const beforeCursor = textarea.value.substring(0, start);
    const needsNewline =
      beforeCursor.length > 0 && !beforeCursor.endsWith("\n");

    let newText: string;
    if (selectedText && selectedText.includes("\n")) {
      const lines = selectedText.split("\n");
      newText =
        (needsNewline ? "\n" : "") +
        lines.map((line) => prefix + line).join("\n");
    } else {
      const replacement = selectedText || placeholder;
      newText = (needsNewline ? "\n" : "") + prefix + replacement;
    }

    const newContent =
      textarea.value.substring(0, start) +
      newText +
      textarea.value.substring(end);

    saveToHistory();
    updatePageContent(newContent);

    nextTick(() => {
    //   textarea.focus();
      const offset = (needsNewline ? 1 : 0) + prefix.length;
      textarea.setSelectionRange(
        start + offset,
        start +
          offset +
          (selectedText ? selectedText.length : placeholder.length)
      );
    });
  }

  function insertList() {
    insertLinePrefix("- ", "List item");
  }

  function insertNumberedList() {
    insertLinePrefix("1. ", "List item");
  }

  function insertTaskList() {
    insertLinePrefix("- [ ] ", "Task item");
  }

  function insertQuote() {
    insertLinePrefix("> ", "Quote text");
  }

  function insertTable() {
    const tableTemplate = `| Column 1 | Column 2 | Column 3 |
|----------|----------|----------|
| Row 1    | Data     | Data     |
| Row 2    | Data     | Data     |`;

    const textarea = document.querySelector("textarea") as HTMLTextAreaElement;
    if (!textarea) return;

    const start = textarea.selectionStart;
    const beforeCursor = textarea.value.substring(0, start);
    const needsNewline =
      beforeCursor.length > 0 && !beforeCursor.endsWith("\n");

    const newText = (needsNewline ? "\n" : "") + tableTemplate;
    const newContent =
      textarea.value.substring(0, start) +
      newText +
      textarea.value.substring(start);

    saveToHistory();
    updatePageContent(newContent);

    nextTick(() => {
    //   textarea.focus();
      textarea.setSelectionRange(
        start + newText.length,
        start + newText.length
      );
    });
  }

  function insertHorizontalRule() {
    const textarea = document.querySelector("textarea") as HTMLTextAreaElement;
    if (!textarea) return;

    const start = textarea.selectionStart;
    const beforeCursor = textarea.value.substring(0, start);
    const needsNewline =
      beforeCursor.length > 0 && !beforeCursor.endsWith("\n");

    const newText = (needsNewline ? "\n\n" : "\n") + "---" + "\n\n";
    const newContent =
      textarea.value.substring(0, start) +
      newText +
      textarea.value.substring(start);

    saveToHistory();
    updatePageContent(newContent);

    nextTick(() => {
    //   textarea.focus();
      textarea.setSelectionRange(
        start + newText.length,
        start + newText.length
      );
    });
  }

  return {
  updatePageContent,
  insertBold,
  insertItalic,
  insertStrikethrough,
  insertHighlight,
  insertCode,
  insertCodeBlock,
  insertLink,
  insertImage,
  insertImageWithTitle,        
  insertImageWithDimensions,   
  insertImageCentered,         
  insertImageSmall,            
  insertImageWithBorder,       
  insertImageLink,             
  insertHeader,
  insertList,
  insertNumberedList,
  insertTaskList,
  insertQuote,
  insertTable,
  insertHorizontalRule
}
}
