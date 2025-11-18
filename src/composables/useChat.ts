import type {
  Chat,
  ConfirmDialogOptions,
  LinkPreview,
  PlatformError,
  UserDetails,
  ApiResponse,
  CreateChatRequest,
} from "@/types";
import {
  API_BASE_URL,
  copyCode,
  createErrorContext,
  detectLargePaste,
  extractUrls,
  generateChatId,
  generateChatTitle,
  getErrorStatus,
} from "@/utils/globals";
import { nextTick, ref, type Ref } from "vue";
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
  isAuthenticated,
  saveLinkPreviewCache,
  isLoading,
  confirmDialog,
}: {
  confirmDialog: Ref<ConfirmDialogOptions>;
  isAuthenticated: Ref<boolean>;
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
  const activeChatMenu = ref<string | null>(null);
  const expanded = ref<boolean[]>([]);

  function reportError(error: PlatformError) {
    console.error("Platform Error:", error);
  }

  async function apiCall<T>(
    endpoint: string,
    options: RequestInit = {},
  ): Promise<ApiResponse<T>> {
    const userId = parsedUserDetails.value?.userId;
    if (!userId) {
      throw new Error("User not authenticated");
    }

    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        "X-User-ID": userId,
        ...options.headers,
      },
    });

    const data: ApiResponse<T> = await response.json();

    if (!response.ok || !data.success) {
      throw new Error(data.message || "API request failed");
    }

    return data;
  }

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
            action: "showConfirmDialog",
            message: "Error in confirm callback: " + error.message,
            description: "An error occurred while processing your request",
            status: getErrorStatus(error),
            context: createErrorContext({
              errorName: error.name,
              errorStack: error.stack,
            }),
            userId: parsedUserDetails.value?.userId || "unknown",
            severity: "high",
            id: generateErrorId(),
            createdAt: new Date().toISOString(),
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
            action: "showConfirmDialog",
            message: "Error in cancel callback: " + error.message,
            status: getErrorStatus(error),
            context: createErrorContext({
              errorName: error.name,
              errorStack: error.stack,
            }),
            userId: parsedUserDetails.value?.userId || "unknown",
            severity: "low",
            id: generateErrorId(),
            createdAt: new Date().toISOString(),
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

  function shareResponse(title: string, text?: string) {
    if (navigator.share) {
      navigator
        .share({
          title:
            title && title.length > 200
              ? `${title.slice(0, 200)}...\n\n`
              : `${title}\n\n`,
          text: text || "",
        })
        .then(() => {
          console.log("Share successful");
        })
        .catch((err) => {
          console.log("Share canceled", err);
        });
    } else {
      copyCode(text || "");
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
      chat.updated_at &&
      typeof chat.updated_at === "string"
    );
  }

  function isNewerChat(serverChat: Chat, localChat: Chat): boolean {
    try {
      return (
        new Date(serverChat.updated_at).getTime() >
        new Date(localChat.updated_at).getTime()
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
            new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime()
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

  async function deleteChat(chatId: string) {
    if (isLoading.value || !chatId) return;

    try {
      const chatIndex = chats.value.findIndex((chat) => chat.id === chatId);
      if (chatIndex === -1) {
        toast.error("Chat not found");
        return;
      }

      const chatToDelete = chats.value[chatIndex];
      const chatTitle = chatToDelete.title || "Untitled Chat";
      const messageCount = chatToDelete.message_count || 0;

      showConfirmDialog({
        visible: true,
        title: "Delete Chat",
        message: `Are you sure you want to delete "${chatTitle}"?\n\nThis will permanently remove ${messageCount} message(s). This action cannot be undone.`,
        type: "danger",
        confirmText: "Delete",
        onConfirm: async () => {
          try {
            let serverDeleteSuccess = true;

            if (isAuthenticated.value) {
              try {
                await apiCall(`/chats/${chatId}`, { method: "DELETE" });
              } catch (apiError: any) {
                console.error("Failed to delete chat from server:", apiError);
                toast.warning("Chat not fully deleted", {
                  description:
                    "Server sync failed. Chat will remain until server deletion succeeds.",
                });
                serverDeleteSuccess = false;
              }
            }

            if (serverDeleteSuccess) {
              if (chatToDelete.messages?.length) {
                chatToDelete.messages.forEach((message) => {
                  try {
                    const promptUrls = extractUrls(message.prompt || "");
                    const responseUrls = extractUrls(message.response || "");
                    const urls = [...new Set([...promptUrls, ...responseUrls])];

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

              const isDeletingCurrentChat = currentChatId.value === chatId;

              chats.value.splice(chatIndex, 1);

              if (isDeletingCurrentChat) {
                if (chats.value.length > 0) {
                  currentChatId.value = chats.value[0].id;
                } else {
                  currentChatId.value = "";
                }
                updateExpandedArray();
              }

              clearCurrentDraft();
              saveChats();

              toast.success("Chat deleted", {
                duration: 3000,
                description: `"${chatTitle}" has been removed`,
              });
            }
          } catch (error: any) {
            reportError({
              action: "onconfirm delete chat",
              message: "Failed to delete chat: " + error.message,
              description: "Failed to delete this chat. Please try again.",
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
              id: generateErrorId(),
              createdAt: new Date().toISOString(),
            } as PlatformError);
          }
        },
      });
    } catch (error: any) {
      reportError({
        action: "delete chat",
        message: "Failed to delete chat: " + error.message,
        description: "Failed to delete this chat. Please try again.",
        status: getErrorStatus(error),
        context: createErrorContext({
          chatId,
          errorName: error.name,
        }),
        userId: parsedUserDetails.value?.userId || "unknown",
        id: generateErrorId(),
        createdAt: new Date().toISOString(),
      } as PlatformError);
    }
  }

  async function loadChats() {
    try {
      if (isAuthenticated.value) {
        try {
          const response = await apiCall<Chat[]>("/chats");
          if (response.data) {
            chats.value = response.data;
            saveChats();
          }
        } catch (apiError) {
          console.error("Failed to load chats from server:", apiError);
        }
      } else {
        const stored = localStorage.getItem("chats");
        if (stored) {
          const parsedChats = JSON.parse(stored);
          if (Array.isArray(parsedChats)) {
            chats.value = parsedChats;
          }
        }
      }

      updateExpandedArray();
    } catch (error) {
      console.error("Failed to load chats:", error);
      chats.value = [];
    }
  }

  async function loadChat() {
    try {
      if (isAuthenticated.value) {
        try {
          isLoading.value = true;
          const response = await apiCall<Chat>(`/chats/${currentChatId.value}`);
          if (response.data) {
            const existingChatIndex = chats.value.findIndex(
              (chat) => chat.id === response.data!.id,
            );
            if (existingChatIndex !== -1) {
              chats.value[existingChatIndex] = response.data;
            } else {
              chats.value.push(response.data);
            }
            saveChats();
            isLoading.value = false;
          }
        } catch (apiError) {
          isLoading.value = false;
          console.error("Failed to load chats from server:", apiError);
        }
      } else {
        const stored = localStorage.getItem("chats");
        if (stored) {
          const parsedChats = JSON.parse(stored);
          if (Array.isArray(parsedChats)) {
            chats.value = parsedChats;
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
          }
        } else if (currentDraft.trim()) {
          const textarea = document.getElementById(
            "prompt",
          ) as HTMLTextAreaElement;
          if (textarea) {
            textarea.value = currentDraft;
            autoGrow({ target: textarea } as any);
            shouldFocus = true;
          }
          pastePreviews.value.delete(currentChatId.value);
        } else {
          const textarea = document.getElementById(
            "prompt",
          ) as HTMLTextAreaElement;
          if (textarea) {
            textarea.value = "";
            autoGrow({ target: textarea } as any);
          }
          pastePreviews.value.delete(currentChatId.value);
        }

        if (shouldFocus) {
          nextTick(() => {
            const textarea = document.getElementById(
              "prompt",
            ) as HTMLTextAreaElement;
            if (textarea) {
              setTimeout(() => {
                textarea.focus();
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
    } catch (error: any) {
      reportError({
        action: "saveChats",
        message: "Failed to save chats: " + error.message,
        description: "Your recent changes may not be saved.",
        status: getErrorStatus(error),
        userId: parsedUserDetails.value?.userId || "unknown",
        severity: "critical",
        id: generateErrorId(),
        createdAt: new Date().toISOString(),
      } as PlatformError);
    }
  }

  async function createNewChat(firstMessage?: string): Promise<string> {
    try {
      const newChatId = generateChatId();
      const now = new Date().toISOString();
      const title = firstMessage ? generateChatTitle(firstMessage) : "New Chat";

      const newChat: Chat = {
        id: newChatId,
        user_id: parsedUserDetails.value?.userId || "",
        title,
        messages: [],
        created_at: now,
        updated_at: now,
        is_archived: false,
        message_count: 0,
        last_message_at: now,
        is_private: true,
      };

      if (isAuthenticated.value) {
        try {
          const response = await apiCall<Chat>("/chats", {
            method: "POST",
            body: JSON.stringify({ title } as CreateChatRequest),
          });

          if (response.data) {
            newChat.id = response.data.id;
            newChat.created_at = response.data.created_at;
            newChat.updated_at = response.data.updated_at;
            newChat.is_private = response.data.is_private;

            chats.value.unshift(newChat);
            currentChatId.value = newChat.id;
            updateExpandedArray();
            saveChats();
          } else {
            console.error("Failed to create chat on server: Invalid response");
            toast.warning("Chat created locally only, server sync failed");
          }
        } catch (apiError) {
          console.error("Failed to create chat on server:", apiError);
          toast.warning("Chat created locally only, server sync failed");
        }
      } else {
        chats.value.unshift(newChat);
        currentChatId.value = newChat.id;
        updateExpandedArray();
        saveChats();
      }

      if (currentChatId.value) {
        pastePreviews.value.delete(currentChatId.value);

        const textarea = document.getElementById(
          "prompt",
        ) as HTMLTextAreaElement;
        const currentDraft = textarea?.value || "";
        if (currentDraft.trim()) {
          chatDrafts.value.set(currentChatId.value, currentDraft);
          saveChatDrafts();
        }
      }

      chatDrafts.value.set(newChat.id, "");
      pastePreviews.value.delete(newChat.id);

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

      console.log(`✅ Created new chat: ${newChat.id}`);
      return newChat.id;
    } catch (error: any) {
      reportError({
        action: "createNewChat",
        message: "Error creating new chat: " + error.message,
        description: "Failed to create a new chat. Try again",
        status: getErrorStatus(error),
        userId: parsedUserDetails.value?.userId || "unknown",
        severity: "critical",
        id: generateErrorId(),
        createdAt: new Date().toISOString(),
      } as PlatformError);
      return "";
    }
  }

  async function processLinksInResponse(index: number) {
    const targetChat = chats.value.find(
      (chat) => chat.id === currentChatId.value,
    );
    if (
      !targetChat ||
      !targetChat.messages ||
      !targetChat.messages[index] ||
      !targetChat.messages[index].response
    )
      return;

    const message = targetChat.messages[index];
    processLinks(message.response);
  }

  async function processLinksInUserPrompt(index: number) {
    const targetChat = chats.value.find(
      (chat) => chat.id === currentChatId.value,
    );
    if (!targetChat || !targetChat.messages || !targetChat.messages[index])
      return;

    const message = targetChat.messages[index];
    processLinks(message.prompt || "");
  }

  const processLinks = (message: string) => {
    const urls = extractUrls(message);
    if (urls.length > 0) {
      urls.slice(0, 3).forEach((url) => {
        fetchLinkPreview(url).then(() => {
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

  async function updateChat(
    chatId: string,
    update: Record<string, any>,
  ): Promise<Chat | undefined> {
    try {
      const chat = chats.value.find((c) => c.id === chatId);
      if (!chat) {
        toast.error("Chat not found");
        return;
      }

      if (isAuthenticated.value) {
        const response = await apiCall<Chat>(`/chats/${chatId}`, {
          method: "PUT",
          body: JSON.stringify(update),
        });

        if (response.success && response.data) {
          chat.title = response.data.title.trim();
          chat.updated_at = response.data.updated_at;
          chat.last_message_at = response.data.last_message_at;
          chat.is_archived = response.data.is_archived;
          chat.is_private = response.data.is_private;
          saveChats();
          return response.data;
        } else {
          console.error("Failed to update chat on server:", response);
        }
      }
    } catch (error: any) {
      reportError({
        action: "updateChat",
        message: "Error updating chat: " + error.message,
        description: "Failed to update this chat. Please try again.",
        status: getErrorStatus(error),
        userId: parsedUserDetails.value?.userId || "unknown",
        severity: "critical",
        id: generateErrorId(),
        createdAt: new Date().toISOString(),
        context: createErrorContext({ chatId }),
      } as PlatformError);
    }
  }

  async function renameChat(chatId: string, newTitle: string): Promise<string> {
    if (!chatId || !newTitle || typeof newTitle !== "string") {
      toast.error("Invalid chat title");
      return "";
    }

    const trimmedTitle = newTitle.trim();
    if (!trimmedTitle) {
      toast.error("Chat title cannot be empty");
      return "";
    }
    const res = await updateChat(chatId, { title: trimmedTitle });
    return res?.title || "";
  }

  function clearAllChats() {
    if (isLoading.value) return;

    try {
      const totalChats = chats.value.length;
      const totalMessages = chats.value.reduce(
        (sum, chat) => sum + (chat.message_count || 0),
        0,
      );

      if (totalChats === 0) {
        toast.info("There are no chats to clear");
        return;
      }

      showConfirmDialog({
        visible: true,
        title: "Clear All Chats",
        message: `⚠️ DELETE ALL CHATS?\n\nThis will permanently delete:\n• ${totalChats} chat(s)\n• ${totalMessages} total message(s)\n\nThis action cannot be undone!`,
        type: "danger",
        confirmText: "Delete All",
        onConfirm: async () => {
          try {
            if (isAuthenticated.value) {
              const response = await apiCall(`/chats`, {
                method: "DELETE",
              });

              if (!response.success) {
                console.error("Failed to delete chats:", response.message);
                return;
              }
            }

            chats.value = [];
            currentChatId.value = "";
            expanded.value = [];
            linkPreviewCache.value.clear();
            chatDrafts.value.clear();
            pastePreviews.value.clear();

            saveChatDrafts();

            const keysToRemove = [
              "chats",
              "linkPreviews",
              "chatDrafts",
              "pastePreviews",
            ];
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

            toast.success("All chats cleared", {
              duration: 5000,
              description: `Deleted ${totalChats} chats with ${totalMessages} messages`,
            });
          } catch (error: any) {
            reportError({
              action: "clearAllChats",
              message: "Error clearing all chats: " + error.message,
              description: "Failed to clear all chats. Try again",
              status: getErrorStatus(error),
              userId: parsedUserDetails.value?.userId || "unknown",
              severity: "critical",
              id: generateErrorId(),
              createdAt: new Date().toISOString(),
            } as PlatformError);
          }
        },
      });
    } catch (error: any) {
      reportError({
        action: "clearAllChats",
        message: "Error clearing all chats: " + error.message,
        description: "Failed to clear all chats. Try again",
        status: getErrorStatus(error),
        userId: parsedUserDetails.value?.userId || "unknown",
        severity: "critical",
        id: generateErrorId(),
        createdAt: new Date().toISOString(),
      } as PlatformError);
    }
  }

  return {
    updateChat,
    expanded,
    activeChatMenu,
    draftSaveTimeout,
    isValidChat,
    isNewerChat,
    mergeChats,
    deleteChat,
    saveChats,
    autoGrow,
    saveChatDrafts,
    loadChatDrafts,
    clearCurrentDraft,
    autoSaveDraft,
    createNewChat,
    copyResponse,
    shareResponse,
    loadChats,
    loadChat,
    processLinksInUserPrompt,
    processLinksInResponse,
    toggleChatMenu,
    renameChat,
    clearAllChats,
    apiCall,
  };
}
