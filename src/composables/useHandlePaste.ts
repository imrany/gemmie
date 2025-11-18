import { detectLargePaste } from "@/lib/globals";
import { detectContentType } from "@/lib/previewPasteContent";
import type { Ref } from "vue";
import { toast } from "vue-sonner/src/packages/state.js";

interface PastePreview {
  content: string;
  wordCount: number;
  charCount: number;
  show: boolean;
}

export function useHandlePaste({
  currentChatId,
  pastePreviews,
  chatDrafts,
  saveChatDrafts,
  autoGrow,
  currentPasteContent,
  showPasteModal,
}: {
  showPasteModal: Ref<boolean>;
  currentChatId: Ref<string | null>;
  pastePreviews: Ref<Map<string, PastePreview>>;
  chatDrafts: Ref<Map<string, string>>;
  saveChatDrafts: () => void;
  currentPasteContent: Ref<{
    content: string;
    wordCount: number;
    charCount: number;
    type: "text" | "code" | "json" | "markdown" | "xml" | "html";
  } | null>;
  autoGrow: (data: any) => void;
}) {
  function handlePaste(e: ClipboardEvent) {
    try {
      const pastedText = e.clipboardData?.getData("text") || "";

      if (!pastedText.trim()) return;

      if (detectLargePaste(pastedText)) {
        e.preventDefault();

        const wordCount = pastedText
          .trim()
          .split(/\s+/)
          .filter((word) => word.length > 0).length;
        const charCount = pastedText.length;

        // Enhanced paste preview with proper content handling
        const processedContent =
          wordCount > 100 ? `#pastedText#${pastedText}` : pastedText;

        // Store in pastePreviews map using current chat ID
        if (currentChatId.value) {
          pastePreviews.value.set(currentChatId.value, {
            content: processedContent,
            wordCount,
            charCount,
            show: true,
          });
        }

        // Save draft immediately when large content is pasted
        if (currentChatId.value) {
          const textarea = document.getElementById(
            "prompt",
          ) as HTMLTextAreaElement;
          const currentDraft = textarea?.value || "";

          // Combine current textarea content with paste preview content
          const fullDraft = currentDraft + processedContent;
          chatDrafts.value.set(currentChatId.value, fullDraft);
          saveChatDrafts();

          // Clear textarea but keep the draft with paste content
          if (textarea) {
            textarea.value = currentDraft; // Keep only the typed content in textarea
            autoGrow({ target: textarea } as any);
          }
        }

        toast.info("Large content detected", {
          duration: 4000,
          description: `${wordCount} words, ${charCount} characters. Preview shown below.`,
        });
      } else {
        // For small pastes, let the normal paste happen and then save draft
        setTimeout(() => {
          if (currentChatId.value) {
            const textarea = document.getElementById(
              "prompt",
            ) as HTMLTextAreaElement;
            if (textarea) {
              // For small pastes, save the normal content
              chatDrafts.value.set(currentChatId.value, textarea.value);
              saveChatDrafts();
            }
          }
        }, 100);
      }
    } catch (error) {
      console.error("Error handling paste:", error);
      // Don't prevent default on error - let normal paste proceed
    }
  }

  function removePastePreview() {
    // Remove paste preview for current chat
    if (currentChatId.value) {
      pastePreviews.value.delete(currentChatId.value);
      saveChatDrafts();

      // Also clear textarea if it contains paste content
      const textarea = document.getElementById("prompt") as HTMLTextAreaElement;
      if (textarea && textarea.value.includes("#pastedText#")) {
        // Extract any non-pasted content
        const parts = textarea.value.split("#pastedText#");
        textarea.value = parts[0] || "";
        autoGrow({ target: textarea } as any);
      }
    }
  }

  function handlePastePreviewClick(e: Event) {
    const target = e.target as HTMLElement;

    // Check if the clicked element itself or any parent has the clickable class
    const clickableElement = target.closest(".paste-preview-clickable");

    if (clickableElement) {
      // Prevent event bubbling to avoid conflicts
      e.preventDefault();
      e.stopPropagation();

      const content = clickableElement.getAttribute("data-paste-content");
      const wordCount = clickableElement.getAttribute("data-word-count");
      const charCount = clickableElement.getAttribute("data-char-count");

      if (content && wordCount && charCount) {
        try {
          const decodedContent = decodeURIComponent(content);
          const parsedWordCount = parseInt(wordCount, 10);
          const parsedCharCount = parseInt(charCount, 10);

          openPasteModal(decodedContent, parsedWordCount, parsedCharCount);
        } catch (error) {
          console.error("Error parsing paste preview data:", error);
          toast.error("Error opening paste preview", {
            duration: 3000,
            description: "Could not parse content data",
          });
        }
      }
    }
  }

  function handleRemovePastePreview(e: Event) {
    const target = e.target as HTMLElement;

    if (target.classList.contains("remove-paste-preview")) {
      e.preventDefault();
      e.stopPropagation();
      removePastePreview();
    }
  }

  function setupPastePreviewHandlers() {
    // Remove existing listeners to avoid duplicates
    document.removeEventListener("click", handlePastePreviewClick, true);
    document.removeEventListener("click", handleRemovePastePreview, true);

    // Add event delegation with capture phase for better reliability
    document.addEventListener("click", handlePastePreviewClick, true);
    document.addEventListener("click", handleRemovePastePreview, true);

    console.log("Paste preview handlers setup complete"); // Debug log
  }

  // Function to open paste modal
  function openPasteModal(
    content: string,
    wordCount: number,
    charCount: number,
  ) {
    try {
      // Handle the #pastedText# prefix if present
      const actualContent = content.startsWith("#pastedText#")
        ? content.substring(12)
        : content;

      // Detect content type - provide fallback if function not available
      let contentType: "text" | "code" | "json" | "markdown" | "xml" | "html" =
        "text";

      if (typeof detectContentType === "function") {
        contentType = detectContentType(actualContent);
      } else {
        // Simple content type detection as fallback
        if (
          actualContent.trim().startsWith("{") &&
          actualContent.trim().endsWith("}")
        ) {
          contentType = "json";
        } else if (
          actualContent.includes("```") ||
          actualContent.includes("function") ||
          actualContent.includes("class")
        ) {
          contentType = "code";
        } else if (
          actualContent.includes("#") ||
          actualContent.includes("**")
        ) {
          contentType = "markdown";
        } else if (actualContent.includes("<") && actualContent.includes(">")) {
          contentType = "html";
        }
      }

      currentPasteContent.value = {
        content: actualContent,
        wordCount,
        charCount,
        type: contentType,
      };

      showPasteModal.value = true;

      // Prevent body scroll
      document.body.style.overflow = "hidden";

      console.log("Paste modal opened successfully", {
        wordCount,
        charCount,
        type: contentType,
      }); // Debug log
    } catch (error) {
      console.error("Error opening paste modal:", error);
      toast.error("Error opening preview", {
        duration: 3000,
        description: "Could not display content preview",
      });
    }
  }

  function cleanupPastePreviewHandlers() {
    document.removeEventListener("click", handlePastePreviewClick, true);
    document.removeEventListener("click", handleRemovePastePreview, true);
    console.log("Paste preview handlers cleaned up"); // Debug log
  }

  function closePasteModal() {
    showPasteModal.value = false;
    currentPasteContent.value = null;

    // Restore body scroll
    document.body.style.overflow = "auto";
  }

  return {
    handlePaste,
    removePastePreview,
    cleanupPastePreviewHandlers,
    openPasteModal,
    setupPastePreviewHandlers,
    handlePastePreviewClick,
    closePasteModal,
    handleRemovePastePreview,
  };
}
