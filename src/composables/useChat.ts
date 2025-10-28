import type { Chat } from "@/types";
import { copyCode } from "@/utils/globals";
import type { Ref } from "vue";
import { toast } from "vue-sonner/src/packages/state.js";

export function useChat({
  copiedIndex,
  chats,
  currentChatId,
  updateExpandedArray,
}: {
  copiedIndex: Ref<number | null>;
  chats: Ref<Chat[]>;
  currentChatId: Ref<string | null>;
  updateExpandedArray: () => void;
}) {
  function copyResponse(text: string, index?: number) {
    navigator.clipboard
      .writeText(text)
      .then(() => {
        copiedIndex.value = index ?? null;

        setTimeout(() => {
          copiedIndex.value = null;
        }, 2000);
      })
      .catch((err) => {
        console.error("Failed to copy text: ", err);
        toast.error("Copy Failed", {
          duration: 3000,
          description: "",
        });
      });
  }

  function shareResponse(text: string, prompt?: string) {
    if (navigator.share) {
      navigator
        .share({
          title:
            prompt && prompt.length > 200
              ? `${prompt.slice(0, 200)}...\n\n`
              : `${prompt || "Gemmie Chat"}\n\n`,
          text,
        })
        .then(() => {
          console.log("Share successful");
        })
        .catch((err) => {
          console.log("Share canceled", err);
        });
    } else {
      copyCode(text);
      toast.info("Copied Instead", {
        duration: 3000,
      });
    }
  }

  // Load chats from localStorage
  function loadChats() {
    try {
      const stored = localStorage.getItem("chats");
      if (stored) {
        const parsedChats = JSON.parse(stored);
        if (Array.isArray(parsedChats)) {
          chats.value = parsedChats;
          if (chats.value.length > 0 && !currentChatId.value) {
            currentChatId.value = chats.value[0].id;
          }
        }
      }
      updateExpandedArray();
    } catch (error) {
      console.error("Failed to load chats:", error);
      chats.value = [];
    }
  }

  return {
    copyResponse,
    shareResponse,
    loadChats,
  };
}
