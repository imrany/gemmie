import type {
  Chat,
  ConfirmDialogOptions,
  LinkPreview,
  PlatformError,
  UserDetails,
} from "@/types";
import {
  copyCode,
  createErrorContext,
  detectLargePaste,
  extractUrls,
  generateChatId,
  generateChatTitle,
  getErrorStatus,
} from "@/utils/globals";
import { nextTick, ref, type Ref } from "vue";
import { useRouter } from "vue-router";
import { toast } from "vue-sonner/src/packages/state.js";
import { generateErrorId } from "./usePlatformError";

export function useChat({
  copiedIndex,
  chats,
  currentChatId,
  updateExpandedArray,
  linkPreviewCache,
  fetchLinkPreview,
  chatDrafts,
  pastePreviews,
  parsedUserDetails,
  performSmartSync,
  isAuthenticated,
  syncStatus,
  saveLinkPreviewCache,
  isLoading,
  confirmDialog,
}: {
  syncStatus: Ref<{
    lastSync: Date | null;
    syncing: boolean;
    hasUnsyncedChanges: boolean;
    lastError: string | null;
    retryCount: number;
    maxRetries: number;
    showSyncIndicator: boolean;
    syncMessage: string;
    syncProgress: number;
  }>;
  confirmDialog: Ref<ConfirmDialogOptions>;
  isAuthenticated: Ref<boolean>;
  performSmartSync: () => void;
  parsedUserDetails: Ref<UserDetails>;
  copiedIndex: Ref<number | null>;
  chats: Ref<Chat[]>;
  currentChatId: Ref<string | null>;
  updateExpandedArray: () => void;
  linkPreviewCache: Ref<Map<string, LinkPreview | null>>;
  fetchLinkPreview: (url: string) => Promise<LinkPreview | null>;
  chatDrafts: Ref<Map<string, string>>;
  saveLinkPreviewCache: () => void;
  isLoading: Ref<boolean>;
  pastePreviews: Ref<
    Map<
      string,
      {
        content: string;
        wordCount: number;
        charCount: number;
        show: boolean;
      }
    >
  >;
}) {
  const router = useRouter();
  const activeChatMenu = ref<string | null>(null);
  const expanded = ref<boolean[]>([]);

  function showConfirmDialog(options: ConfirmDialogOptions) {
    confirmDialog.value = {
      visible: true,
      title: options.title,
      message: options.message,
      type: options.type || "info",
      confirmText: options.confirmText || "Confirm",
      cancelText: options.cancelText || "Cancel",
      onConfirm: () => {
        try {
          options.onConfirm();
        } catch (error: any) {
          reportError({
            action: `showConfirmDialog`,
            message: "Error in confirm callback :" + error.message,
            description: `An error occurred while processing your request`,
            status: getErrorStatus(error),
            context: createErrorContext({
              errorName: error.name,
              errorStack: error.stack,
            }),
            userId: parsedUserDetails.value?.userId || "unknown",
            severity: "high",
          } as PlatformError);
        } finally {
          nextTick(() => {
            confirmDialog.value.visible = false;
          });
        }
      },
      onCancel: () => {
        try {
          options.onCancel?.();
        } catch (error: any) {
          reportError({
            action: `showConfirmDialog`,
            message: "Error in cancel callback :" + error.message,
            status: getErrorStatus(error),
            context: createErrorContext({
              errorName: error.name,
              errorStack: error.stack,
            }),
            userId: parsedUserDetails.value?.userId || "unknown",
            severity: "low",
          } as PlatformError);
        } finally {
          nextTick(() => {
            confirmDialog.value.visible = false;
          });
        }
      },
    };
  }

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

  function isValidChat(chat: any): chat is Chat {
    return (
      chat &&
      typeof chat === "object" &&
      chat.id &&
      typeof chat.id === "string" &&
      Array.isArray(chat.messages) &&
      chat.updatedAt &&
      typeof chat.updatedAt === "string"
    );
  }

  function isNewerChat(serverChat: Chat, localChat: Chat): boolean {
    try {
      return (
        new Date(serverChat.updatedAt).getTime() >
        new Date(localChat.updatedAt).getTime()
      );
    } catch (error) {
      return false;
    }
  }

  function mergeChats(serverChats: Chat[], localChats: Chat[]): Chat[] {
    try {
      if (!Array.isArray(serverChats)) serverChats = [];
      if (!Array.isArray(localChats)) localChats = [];

      const merged = new Map<string, Chat>();

      localChats.forEach((chat) => {
        if (isValidChat(chat)) {
          merged.set(chat.id, chat);
        }
      });

      serverChats.forEach((serverChat) => {
        if (!isValidChat(serverChat)) return;

        const localChat = merged.get(serverChat.id);
        if (!localChat || isNewerChat(serverChat, localChat)) {
          merged.set(serverChat.id, serverChat);
        }
      });

      return Array.from(merged.values()).sort((a, b) => {
        try {
          return (
            new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()
          );
        } catch (error) {
          return 0;
        }
      });
    } catch (error) {
      console.error("Error merging chats:", error);
      return localChats || [];
    }
  }

  function deleteChat(chatId: string) {
    if (isLoading.value || !chatId) return;

    try {
      const chatIndex = chats.value.findIndex((chat) => chat.id === chatId);
      if (chatIndex === -1) {
        toast.error("Chat not found");
        return;
      }

      const chatToDelete = chats.value[chatIndex];
      const chatTitle = chatToDelete.title || "Untitled Chat";
      const messageCount = chatToDelete.messages?.length || 0;

      showConfirmDialog({
        visible: true,
        title: "Delete Chat",
        message: `Are you sure you want to delete "${chatTitle}"?\n\nThis will permanently remove ${messageCount} message(s). This action cannot be undone.`,
        type: "danger",
        confirmText: "Delete",
        onConfirm: () => {
          try {
            // Remove link previews from cache before deleting
            if (chatToDelete.messages?.length > 0) {
              chatToDelete.messages.forEach((message) => {
                try {
                  const responseUrls = extractUrls(message.response || "");
                  const promptUrls = extractUrls(message.prompt || "");
                  const urls = [...new Set([...responseUrls, ...promptUrls])];

                  urls.forEach((url) => {
                    linkPreviewCache.value.delete(url);
                  });
                } catch (error: any) {
                  reportError({
                    createdAt: new Date().toISOString(),
                    id: generateErrorId(),
                    action: "deleteChat",
                    message:
                      "Error extracting URLs for cache cleanup: " +
                      error.message,
                    status: getErrorStatus(error),
                    userId: parsedUserDetails.value?.userId || "unknown",
                    context: createErrorContext({
                      chatId,
                      errorName: error.name,
                    }),
                    severity: "low",
                  } as PlatformError);
                }
              });
              saveLinkPreviewCache();
            }

            // Remove chat from array
            chats.value.splice(chatIndex, 1);

            // If we deleted the current chat, switch to another one
            if (currentChatId.value === chatId) {
              if (chats.value.length > 0) {
                switchToChat(chats.value[0].id);
              } else {
                currentChatId.value = "";
                updateExpandedArray();
              }
            }

            clearCurrentDraft();
            saveChats();

            // Trigger sync after deleting chat
            if (isAuthenticated.value && parsedUserDetails.value?.syncEnabled) {
              setTimeout(() => {
                performSmartSync();
              }, 1000);
            }
          } catch (error: any) {
            reportError({
              action: `onconfirm delete chat`,
              message: "Failed to delete chat :" + error.message,
              description: `Failed to delete this chat. Please try again.`,
              status: getErrorStatus(error),
              context: createErrorContext({
                chatId,
                chatTitle,
                messageCount,
                errorName: error.name,
                errorStack: error.stack,
              }),
              userId: parsedUserDetails.value?.userId || "unknown",
              severity: "high",
            } as PlatformError);
          }
        },
      });
    } catch (error: any) {
      reportError({
        action: `delete chat`,
        message: "Failed to delete chat :" + error.message,
        description: `Failed to delete this chat. Please try again.`,
        status: getErrorStatus(error),
        context: createErrorContext({
          chatId,
          errorName: error.name,
        }),
        userId: parsedUserDetails.value?.userId || "unknown",
      } as PlatformError);
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

  function autoGrow(e: Event) {
    const el = e.target as HTMLTextAreaElement;
    const maxHeight = 200;
    el.style.height = "auto";
    if (el.scrollHeight <= maxHeight) {
      el.style.height = el.scrollHeight + "px";
      el.style.overflowY = "hidden";
    } else {
      el.style.height = maxHeight + "px";
      el.style.overflowY = "auto";
    }

    autoSaveDraft();
  }

  function saveChatDrafts() {
    try {
      const draftsObject = Object.fromEntries(chatDrafts.value);
      localStorage.setItem("chatDrafts", JSON.stringify(draftsObject));

      const previewsObject = Object.fromEntries(pastePreviews.value);
      localStorage.setItem("pastePreviews", JSON.stringify(previewsObject));
    } catch (error) {
      console.error("Failed to save chat drafts:", error);
    }
  }

  function loadChatDrafts() {
    try {
      const saved = localStorage.getItem("chatDrafts");
      if (saved) {
        const parsedDrafts = JSON.parse(saved);
        chatDrafts.value = new Map(Object.entries(parsedDrafts));
      }

      const savedPastePreviews = localStorage.getItem("pastePreviews");
      if (savedPastePreviews) {
        const parsedPreviews = JSON.parse(savedPastePreviews);
        pastePreviews.value = new Map(Object.entries(parsedPreviews));
      }

      if (currentChatId.value) {
        const currentDraft = chatDrafts.value.get(currentChatId.value) || "";
        const currentPastePreview = pastePreviews.value.get(
          currentChatId.value,
        );

        let shouldFocus = false;

        if (currentPastePreview && currentPastePreview.show) {
          const textarea = document.getElementById(
            "prompt",
          ) as HTMLTextAreaElement;
          if (textarea) {
            const draftWithoutPaste = currentDraft.replace(
              currentPastePreview.content,
              "",
            );
            textarea.value = draftWithoutPaste;
            autoGrow({ target: textarea } as any);
            shouldFocus = true;
            console.log("📝 Focus: Paste preview detected");
          }
        } else if (currentDraft && detectLargePaste(currentDraft)) {
          const wordCount = currentDraft
            .trim()
            .split(/\s+/)
            .filter((word) => word.length > 0).length;
          const charCount = currentDraft.length;

          pastePreviews.value.set(currentChatId.value, {
            content: currentDraft,
            wordCount,
            charCount,
            show: true,
          });

          const textarea = document.getElementById(
            "prompt",
          ) as HTMLTextAreaElement;
          if (textarea) {
            textarea.value = "";
            autoGrow({ target: textarea } as any);
            shouldFocus = true;
            console.log(
              "📝 Focus: Large paste detected and converted to preview",
            );
          }
        } else if (currentDraft.trim()) {
          const textarea = document.getElementById(
            "prompt",
          ) as HTMLTextAreaElement;
          if (textarea) {
            textarea.value = currentDraft;
            autoGrow({ target: textarea } as any);
            shouldFocus = true;
            console.log("📝 Focus: Draft content detected");
          }
          pastePreviews.value.delete(currentChatId.value);
        } else {
          const textarea = document.getElementById(
            "prompt",
          ) as HTMLTextAreaElement;
          if (textarea) {
            textarea.value = "";
            autoGrow({ target: textarea } as any);
            // Don't focus on empty drafts
            console.log("📝 No focus: Empty draft");
          }
          pastePreviews.value.delete(currentChatId.value);
        }

        // Focus the textarea if we have content to work with
        if (shouldFocus) {
          nextTick(() => {
            const textarea = document.getElementById(
              "prompt",
            ) as HTMLTextAreaElement;
            if (textarea) {
              // Small delay to ensure DOM is updated
              setTimeout(() => {
                textarea.focus();
                console.log("🎯 Textarea focused due to draft/preview content");
              }, 100);
            }
          });
        }
      }
    } catch (error) {
      console.error("Failed to load chat drafts:", error);
    }
  }

  function clearCurrentDraft() {
    if (currentChatId.value) {
      chatDrafts.value.delete(currentChatId.value);
      pastePreviews.value.delete(currentChatId.value);
      saveChatDrafts();

      const textarea = document.getElementById("prompt") as HTMLTextAreaElement;
      if (textarea) {
        textarea.value = "";
        autoGrow({ target: textarea } as any);
      }
    }
  }

  let draftSaveTimeout: any = null;

  function autoSaveDraft() {
    if (draftSaveTimeout) {
      clearTimeout(draftSaveTimeout);
    }

    draftSaveTimeout = setTimeout(() => {
      if (currentChatId.value) {
        const textarea = document.getElementById(
          "prompt",
        ) as HTMLTextAreaElement;
        let currentDraft = textarea?.value || "";

        const currentPastePreview = pastePreviews.value.get(
          currentChatId.value,
        );
        if (currentPastePreview?.show) {
          currentDraft += currentPastePreview.content;
        }

        if (currentDraft.trim().length > 0) {
          chatDrafts.value.set(currentChatId.value, currentDraft);
          saveChatDrafts();
        } else {
          chatDrafts.value.delete(currentChatId.value);
          saveChatDrafts();
        }
      }
    }, 1000);
  }

  function switchToChat(chatId: string): boolean {
    try {
      if (!chatId || typeof chatId !== "string") {
        console.error("Invalid chat ID provided");
        return false;
      }

      // Validate chat exists
      const targetChat = chats.value.find((chat) => chat.id === chatId);
      if (!targetChat) {
        console.warn(`Chat with ID ${chatId} not found`);
        toast.error("Chat not found");
        return false;
      }

      // Skip if already on the target chat
      if (currentChatId.value === chatId) {
        console.log("Already on target chat, skipping switch");
        return true;
      }

      // Save current draft if switching from another chat
      if (currentChatId.value) {
        const textarea = document.getElementById(
          "prompt",
        ) as HTMLTextAreaElement;
        let currentDraft = textarea?.value || "";

        const currentPastePreview = pastePreviews.value.get(
          currentChatId.value,
        );
        if (currentPastePreview?.show && currentPastePreview.content) {
          currentDraft += currentPastePreview.content;
        }

        if (currentDraft.trim().length === 0) {
          chatDrafts.value.delete(currentChatId.value);
          pastePreviews.value.delete(currentChatId.value);
        } else {
          chatDrafts.value.set(currentChatId.value, currentDraft);
        }
        saveChatDrafts();
      }

      // Update current chat ID
      const previousChatId = currentChatId.value;
      currentChatId.value = chatId;
      router.push(`/chat/${chatId}`);

      // Update expanded array for new chat
      updateExpandedArray();

      // Load draft for new chat and update UI
      nextTick(() => {
        loadChatDrafts();
      });

      console.log(`✅ Switched from chat ${previousChatId} to ${chatId}`);
      return true;
    } catch (error: any) {
      console.error("Error in switchToChat:", error);
      reportError({
        action: `switchToChat`,
        message: "Error switching to chat: " + error.message,
        description: `Failed to switch to chat : ${chatId.slice(0, 10)}...`,
        status: getErrorStatus(error),
        userId: parsedUserDetails.value?.userId || "unknown",
        severity: "critical",
        context: createErrorContext({
          targetChatId: chatId,
          currentChatId: currentChatId.value,
          errorName: error.name,
        }),
      } as PlatformError);
      return false;
    }
  }

  function saveChats() {
    try {
      if (!Array.isArray(chats.value)) {
        console.error("Chats is not an array, resetting to empty array");
        chats.value = [];
      }

      const chatsJson = JSON.stringify(chats.value);

      if (chatsJson.length > 5000000) {
        toast.warning("Chat data is getting large", {
          duration: 5000,
          description: "Consider clearing old chats to improve performance",
        });
      }

      localStorage.setItem("chats", chatsJson);

      if (isAuthenticated.value && parsedUserDetails.value?.syncEnabled) {
        syncStatus.value.hasUnsyncedChanges = true;
      }
    } catch (error: any) {
      reportError({
        action: `saveChats`,
        message: "Failed to save chats: " + error.message,
        description: `Your recent changes may not be saved.`,
        status: getErrorStatus(error),
        userId: parsedUserDetails.value?.userId || "unknown",
        severity: "critical",
      } as PlatformError);
    }
  }

  function createNewChat(firstMessage?: string): string {
    try {
      const newChatId = generateChatId();
      const now = new Date().toISOString();

      const newChat: Chat = {
        id: newChatId,
        title: firstMessage ? generateChatTitle(firstMessage) : "New Chat",
        messages: [],
        createdAt: now,
        updatedAt: now,
      };

      if (currentChatId.value) {
        pastePreviews.value.delete(currentChatId.value);
      }

      if (currentChatId.value) {
        const textarea = document.getElementById(
          "prompt",
        ) as HTMLTextAreaElement;
        const currentDraft = textarea?.value || "";
        if (currentDraft.trim()) {
          chatDrafts.value.set(currentChatId.value, currentDraft);
          saveChatDrafts();
        }
      }

      chatDrafts.value.set(newChatId, "");
      pastePreviews.value.delete(newChatId);

      chats.value.unshift(newChat);
      currentChatId.value = newChatId;
      updateExpandedArray();
      saveChats();

      // Trigger sync after creating new chat
      if (isAuthenticated.value && parsedUserDetails.value?.syncEnabled) {
        setTimeout(() => {
          performSmartSync();
        }, 1000);
      }

      nextTick(() => {
        const textarea = document.getElementById(
          "prompt",
        ) as HTMLTextAreaElement;
        if (textarea) {
          textarea.value = "";
          autoGrow({ target: textarea } as any);
          textarea.focus();
        }
      });

      router.push(`/chat/${newChatId}`);
      return newChatId;
    } catch (error: any) {
      reportError({
        action: `createNewChat`,
        message: "Error creating new chat: " + error.message,
        description: `Failed to create a new chat. Try again`,
        status: getErrorStatus(error),
        userId: parsedUserDetails.value?.userId || "unknown",
        severity: "critical",
      } as PlatformError);
      return "";
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

  function toggleChatMenu(chatId: string, event: Event) {
    try {
      event.stopPropagation();
      activeChatMenu.value = activeChatMenu.value === chatId ? null : chatId;
    } catch (error) {
      console.error("Error toggling chat menu:", error);
    }
  }

  //  renameChat function
  function renameChat(chatId: string, newTitle: string) {
    try {
      if (!chatId || !newTitle || typeof newTitle !== "string") {
        toast.error("Invalid chat title");
        return;
      }

      const chat = chats.value.find((c) => c.id === chatId);
      if (!chat) {
        toast.error("Chat not found");
        return;
      }

      const trimmedTitle = newTitle.trim();
      if (!trimmedTitle) {
        toast.error("Chat title cannot be empty");
        return;
      }

      chat.title = trimmedTitle;
      chat.updatedAt = new Date().toISOString();
      saveChats();

      // Trigger sync after renaming chat
      if (isAuthenticated.value && parsedUserDetails.value?.syncEnabled) {
        setTimeout(() => {
          performSmartSync();
        }, 1000);
      }
    } catch (error: any) {
      reportError({
        action: `renameChat`,
        message: "Error renaming chat: " + error.message,
        description: `Failed to rename this chat. Please try again.`,
        status: getErrorStatus(error),
        userId: parsedUserDetails.value?.userId || "unknown",
        severity: "critical",
      } as PlatformError);
    }
  }

  function clearAllChats() {
    if (isLoading.value) return;

    try {
      const totalChats = chats.value.length;
      const totalMessages = chats.value.reduce(
        (sum, chat) => sum + (chat.messages?.length || 0),
        0,
      );

      if (totalChats === 0) {
        toast.info("There are no chats to clear", {
          duration: 3000,
          description: "Your chat list is already empty.",
        });
        return;
      }

      showConfirmDialog({
        visible: true,
        title: "Clear All Chats",
        message: `⚠️ DELETE ALL CHATS?\n\nThis will permanently delete:\n• ${totalChats} chat(s)\n• ${totalMessages} total message(s)\n\nThis action cannot be undone!`,
        type: "danger",
        confirmText: "Delete All",
        onConfirm: () => {
          try {
            chats.value = [];
            currentChatId.value = "";
            expanded.value = [];
            linkPreviewCache.value.clear();

            chatDrafts.value.clear();
            saveChatDrafts();

            const keysToRemove = ["chats", "linkPreviews", "chatDrafts"];
            keysToRemove.forEach((key) => {
              try {
                localStorage.removeItem(key);
              } catch (error) {
                console.error(
                  `Failed to remove ${key} from localStorage:`,
                  error,
                );
              }
            });

            saveChats();
            toast.error(
              `${totalChats} chats with ${totalMessages} messages deleted`,
              {
                duration: 5000,
                description: "All chat data has been cleared",
              },
            );
          } catch (error: any) {
            reportError({
              action: `clearAllChats`,
              message: "Error clearing all chats: " + error.message,
              description: `Failed to clear all chats. Try again`,
              status: getErrorStatus(error),
              userId: parsedUserDetails.value?.userId || "unknown",
              severity: "critical",
            } as PlatformError);
          }
        },
      });
    } catch (error: any) {
      reportError({
        action: `clearAllChats`,
        message: "Error clearing all chats: " + error.message,
        description: `Failed to clear all chats. Try again`,
        status: getErrorStatus(error),
        userId: parsedUserDetails.value?.userId || "unknown",
        severity: "critical",
      } as PlatformError);
    }
  }

  return {
    expanded,
    activeChatMenu,
    renameChat,
    clearAllChats,
    toggleChatMenu,
    draftSaveTimeout,
    isNewerChat,
    isValidChat,
    mergeChats,
    deleteChat,
    saveChats,
    autoGrow,
    saveChatDrafts,
    loadChatDrafts,
    clearCurrentDraft,
    autoSaveDraft,
    switchToChat,
    createNewChat,

    copyResponse,
    shareResponse,
    loadChats,
    processLinksInUserPrompt,
    processLinksInResponse,
  };
}
