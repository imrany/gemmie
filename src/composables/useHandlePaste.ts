import { detectLargePaste } from "@/lib/globals";
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
}: {
  currentChatId: Ref<string | null>;
  pastePreviews: Ref<Map<string, PastePreview>>;
  chatDrafts: Ref<Map<string, string>>;
  saveChatDrafts: () => void;
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
      chatDrafts.value.delete(currentChatId.value);
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

  return {
    handlePaste,
    removePastePreview,
  };
}
