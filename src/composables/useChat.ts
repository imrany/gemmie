import type { Chat, LinkPreview } from "@/types";
import { copyCode, extractUrls } from "@/utils/globals";
import type { Ref } from "vue";
import { toast } from "vue-sonner/src/packages/state.js";

export function useChat({
  copiedIndex,
  chats,
  currentChatId,
  updateExpandedArray,
  linkPreviewCache,
  fetchLinkPreview,
}: {
  copiedIndex: Ref<number | null>;
  chats: Ref<Chat[]>;
  currentChatId: Ref<string | null>;
  updateExpandedArray: () => void;
  linkPreviewCache: Ref<Map<string, LinkPreview | null>>;
  fetchLinkPreview: (url: string) => Promise<LinkPreview | null>;
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

  // Process links in a response and generate previews
  async function processLinksInResponse(index: number) {
    const targetChat = chats.value.find(
      (chat) => chat.id === currentChatId.value,
    );
    if (
      !targetChat ||
      !targetChat.messages[index] ||
      !targetChat.messages[index].response ||
      targetChat.messages[index].response === "..."
    )
      return;

    processLinks(targetChat.messages[index].response);
  }

  // Process links in user prompts
  async function processLinksInUserPrompt(index: number) {
    const targetChat = chats.value.find(
      (chat) => chat.id === currentChatId.value,
    );
    if (
      !targetChat ||
      !targetChat.messages[index] ||
      !targetChat.messages[index].prompt ||
      targetChat.messages[index].prompt === ""
    )
      return;
    processLinks(targetChat.messages[index].prompt || "");
  }

  const processLinks = (message: string) => {
    const urls = extractUrls(message);
    if (urls.length > 0) {
      // Start loading previews for user prompt links
      urls.slice(0, 3).forEach((url) => {
        fetchLinkPreview(url).then(() => {
          // Trigger reactivity update
          linkPreviewCache.value = new Map(linkPreviewCache.value);
        });
      });
    }
  };

  return {
    copyResponse,
    shareResponse,
    loadChats,
    processLinksInUserPrompt,
    processLinksInResponse,
  };
}
