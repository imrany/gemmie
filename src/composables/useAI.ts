import { ref } from "vue";
import type { AIAction } from "@/types/document";
import { API_BASE_URL } from "@/lib/globals";
import type { ApiResponse } from "@/types";

export function useAI() {
  const showAIModal = ref(false);
  const aiContent = ref("");
  const originalTextForAI = ref("");
  const isGeneratingAI = ref(false);
  const currentAIAction = ref<AIAction>("summarize");
  const showAISuggestions = ref(false);
  const aiSuggestionsPosition = ref({ x: 0, y: 0 });
  const showAIPopover = ref(false);
  const aiPopoverPosition = ref({ x: 0, y: 0 });
  const aiPopoverContent = ref("");
  const isLoadingAIPopover = ref(false);

  const aiSuggestions = [
    {
      icon: "📝",
      label: "Summarize",
      action: "summarize",
      description: "Create a concise summary",
      shortcut: "Ctrl+Shift+S",
    },
    {
      icon: "🔍",
      label: "Expand",
      action: "expand",
      description: "Elaborate on the selected text",
      shortcut: "Ctrl+Shift+E",
    },
    {
      icon: "✨",
      label: "Improve",
      action: "improve",
      description: "Enhance writing quality",
      shortcut: "Ctrl+Shift+I",
    },
    {
      icon: "📖",
      label: "Simplify",
      action: "simplify",
      description: "Make text easier to understand",
      shortcut: "Ctrl+Shift+P",
    },
    {
      icon: "🔄",
      label: "Paraphrase",
      action: "paraphrase",
      description: "Rewrite in different words",
      shortcut: "Ctrl+Shift+R",
    },
    {
      icon: "🌐",
      label: "Translate",
      action: "translate",
      description: "Translate to different language",
      shortcut: "Ctrl+Shift+T",
    },
  ];

  const parsedUserDetails = JSON.parse(
    localStorage.getItem("userdetails") || "{}",
  );
  async function handlePrompt(
    prompt: string,
  ): Promise<{ Response: string; Prompt: string }> {
    try {
      const response = await fetch(`${API_BASE_URL}/genai`, {
        method: "POST",
        body: JSON.stringify(prompt),
        headers: {
          "Content-Type": "application/json",
          ...(parsedUserDetails.userId
            ? { "X-User-ID": parsedUserDetails.userId }
            : {}),
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const parseRes: ApiResponse<{ Response: string; Prompt: string }> =
        await response.json();
      if (parseRes.data) {
        return parseRes.data;
      } else {
        throw new Error(parseRes.message);
      }
    } catch (err: any) {
      console.error("AI request failed:", err);
      return err.message;
    }
  }

  function getAIPrompt(action: AIAction, text: string): string {
    const prompts = {
      summarize: `Please provide a concise summary of the following text. Focus on the main points and key ideas:\n\n"${text}"`,
      expand: `Please expand on the following text by adding more details, examples, and explanations:\n\n"${text}"`,
      simplify: `Please simplify the following text to make it easier to understand while keeping the core meaning:\n\n"${text}"`,
      translate: `Please translate the following text to English (if not already) and provide the translation:\n\n"${text}"`,
      paraphrase: `Please paraphrase the following text using different words and sentence structures while preserving the original meaning:\n\n"${text}"`,
      improve: `Please improve the following text by enhancing clarity, grammar, and overall writing quality:\n\n"${text}"`,
      explain: `Please explain the following text in simple terms, breaking down complex concepts:\n\n"${text}"`,
    };
    return prompts[action];
  }

  async function performAIAction(action: AIAction, text: string) {
    if (!text.trim()) return;

    const selection: any = window.getSelection();
    let x = window.innerWidth / 2 - 160;
    let y = window.innerHeight / 2 - 100;

    if (selection?.rangeCount > 0) {
      const range = selection.getRangeAt(0);
      const rect = range.getBoundingClientRect();

      x = rect.right + 20;
      y = rect.top;

      const viewport = {
        width: window.innerWidth,
        height: window.innerHeight,
      };

      if (x + 320 > viewport.width - 10) {
        x = rect.left - 330;
      }

      if (x < 10) {
        x = viewport.width / 2 - 160;
      }

      if (y + 300 > viewport.height - 10) {
        y = viewport.height - 310;
      }
      if (y < 10) {
        y = 10;
      }
    }

    aiPopoverPosition.value = { x, y };
    showAIPopover.value = true;
    isLoadingAIPopover.value = true;
    aiPopoverContent.value = "";
    originalTextForAI.value = text;
    currentAIAction.value = action;

    try {
      const prompt = getAIPrompt(action, text);
      const result = await handlePrompt(prompt);
      aiPopoverContent.value = result.Response || "No response generated";
    } catch (error) {
      console.error("Error performing AI action:", error);
      aiPopoverContent.value = "Error generating content. Please try again.";
    } finally {
      isLoadingAIPopover.value = false;
    }
  }

  function showAIToolbar() {
    const textarea = document.querySelector("textarea");
    if (!textarea) return;

    const rect = textarea.getBoundingClientRect();

    aiSuggestionsPosition.value = {
      x: rect.left + rect.width / 2 - 140,
      y: rect.top + rect.height / 2 - 100,
    };

    const viewport = {
      width: window.innerWidth,
      height: window.innerHeight,
    };

    if (aiSuggestionsPosition.value.x < 10) {
      aiSuggestionsPosition.value.x = 10;
    } else if (aiSuggestionsPosition.value.x + 280 > viewport.width - 10) {
      aiSuggestionsPosition.value.x = viewport.width - 290;
    }

    if (aiSuggestionsPosition.value.y < 10) {
      aiSuggestionsPosition.value.y = 10;
    } else if (aiSuggestionsPosition.value.y + 200 > viewport.height - 10) {
      aiSuggestionsPosition.value.y = viewport.height - 210;
    }

    showAISuggestions.value = true;
  }

  function hideAISuggestions() {
    showAISuggestions.value = false;
  }

  return {
    showAIModal,
    aiContent,
    originalTextForAI,
    isGeneratingAI,
    currentAIAction,
    showAISuggestions,
    aiSuggestionsPosition,
    showAIPopover,
    aiPopoverPosition,
    aiPopoverContent,
    isLoadingAIPopover,
    aiSuggestions,
    performAIAction,
    showAIToolbar,
    hideAISuggestions,
  };
}
