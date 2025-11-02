import type {
  Chat,
  Message,
  ConfirmDialogOptions,
  UserDetails,
  ApiResponse,
  CreateChatRequest,
  CreateMessageRequest,
} from "@/types";
<<<<<<< HEAD
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
import { toast } from "vue-sonner/src/packages/state.js";
import { generateErrorId } from "./usePlatformError";
=======
import { ref, type Ref, onUnmounted, onMounted } from "vue";
import { useRouter } from "vue-router";
import { toast } from "vue-sonner";
>>>>>>> new

// ============================================================================
// TYPES: Extended types for optimistic UI
// ============================================================================
interface ExtendedChat extends Chat {
  _status?: "creating" | "synced" | "failed" | "syncing";
}

interface ExtendedMessage extends Message {
  _status?: "sending" | "sent" | "failed" | "pending_chat_sync" | "syncing";
}

interface SyncQueueItem {
  type: "create" | "update" | "delete" | "message";
  chatId: string;
  data?: any;
  timestamp: number;
  retries: number;
  dependsOn?: string;
}

// ============================================================================
// MAIN COMPOSABLE
// ============================================================================
export function useChat({
  chats,
  currentChatId,
  isAuthenticated,
  parsedUserDetails,
  syncStatus,
  isLoading,
  confirmDialog,
}: {
  chats: Ref<ExtendedChat[]>;
  currentChatId: Ref<string | null>;
  isAuthenticated: Ref<boolean>;
  parsedUserDetails: Ref<UserDetails>;
  syncStatus: Ref<{
    lastSync: Date | null;
    syncing: boolean;
    hasUnsyncedChanges: boolean;
  }>;
  isLoading: Ref<boolean>;
  confirmDialog: Ref<ConfirmDialogOptions>;
}) {
<<<<<<< HEAD
  const activeChatMenu = ref<string | null>(null);
  const expanded = ref<boolean[]>([]);
=======
  const router = useRouter();
>>>>>>> new

  // ============================================================================
  // STATE MANAGEMENT
  // ============================================================================
  const syncQueue = ref<SyncQueueItem[]>([]);
  const tempIdMap = new Map<string, string>();
  const isOnline = ref(navigator.onLine);
  const isSyncing = ref(false);
  let syncInterval: ReturnType<typeof setInterval> | null = null;
  let saveTimer: ReturnType<typeof setTimeout> | null = null;

  // ============================================================================
  // VALIDATION: Validate server responses
  // ============================================================================
  function validateChat(data: any): data is Chat {
    return (
      data &&
      typeof data === "object" &&
      typeof data.id === "string" &&
      typeof data.title === "string" &&
      typeof data.created_at === "string" &&
      typeof data.updated_at === "string"
    );
  }

  function validateMessage(data: any): data is Message {
    return (
      data &&
      typeof data === "object" &&
      typeof data.id === "string" &&
      typeof data.chat_id === "string" &&
      (data.role === "user" || data.role === "assistant") &&
      typeof data.content === "string" &&
      typeof data.created_at === "string"
    );
  }

  function validateApiResponse<T>(response: any): response is ApiResponse<T> {
    return (
      response &&
      typeof response === "object" &&
      typeof response.success === "boolean"
    );
  }

  // ============================================================================
  // STORAGE: Debounced localStorage operations
  // ============================================================================
  function saveToStorage() {
    if (saveTimer) clearTimeout(saveTimer);

    saveTimer = setTimeout(() => {
      try {
        localStorage.setItem("chats", JSON.stringify(chats.value));
        console.log("💾 Saved to localStorage");
      } catch (error) {
        console.error("Failed to save to localStorage:", error);
        toast.error("Failed to save locally");
      }
    }, 300);
  }

  function loadFromStorage(): ExtendedChat[] {
    try {
      const cached = localStorage.getItem("chats");
      if (cached) {
        const parsed = JSON.parse(cached);
        if (Array.isArray(parsed)) {
          return parsed;
        }
      }
    } catch (error) {
      console.error("Failed to load from localStorage:", error);
    }
    return [];
  }

  // ============================================================================
  // NETWORK: Online/offline detection
  // ============================================================================
  function setupOnlineDetection() {
    const handleOnline = () => {
      isOnline.value = true;
      console.log("🌐 Back online");
      toast.success("Connection restored");
      processSyncQueue();
    };

    const handleOffline = () => {
      isOnline.value = false;
      console.log("📡 Offline");
      toast.warning("Working offline");
    };

    window.addEventListener("online", handleOnline);
    window.addEventListener("offline", handleOffline);

    return () => {
      window.removeEventListener("online", handleOnline);
      window.removeEventListener("offline", handleOffline);
    };
  }

  // ============================================================================
  // UTILITY: Generate chat title from message
  // ============================================================================
  function generateChatTitle(message: string): string {
    const words = message.trim().split(/\s+/);
    const title = words.slice(0, 6).join(" ");
    return title.length > 50 ? title.slice(0, 47) + "..." : title;
  }

  // ============================================================================
  // UTILITY: Switch to a chat
  // ============================================================================
  function switchToChat(chatId: string) {
    currentChatId.value = chatId;
    router.push(`/chat/${chatId}`);
  }

  // ============================================================================
  // UTILITY: Show confirmation dialog
  // ============================================================================
  function showConfirmDialog(options: ConfirmDialogOptions) {
    confirmDialog.value = options;
  }

  // ============================================================================
  // MERGE: Server wins on conflicts
  // ============================================================================
  function mergeChats(
    serverChats: Chat[],
    localChats: ExtendedChat[],
  ): ExtendedChat[] {
    const merged = new Map<string, ExtendedChat>();

    // Add local chats first (non-temp)
    localChats.forEach((chat) => {
      if (!chat.id.startsWith("temp_")) {
        merged.set(chat.id, chat);
      }
    });

    // Server chats override local
    serverChats.forEach((serverChat) => {
      const existingLocal = merged.get(serverChat.id);
      merged.set(serverChat.id, {
        ...serverChat,
        messages: serverChat.messages || existingLocal?.messages || [],
        _status: "synced",
      });
    });

    // Add back temp chats (not yet synced)
    localChats.forEach((chat) => {
      if (chat.id.startsWith("temp_")) {
        merged.set(chat.id, chat);
      }
    });

    return Array.from(merged.values()).sort(
      (a, b) =>
        new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime(),
    );
  }

  // ============================================================================
  // LOAD CHATS: Cache-first strategy
  // ============================================================================
  async function loadChats(forceFetch = false) {
    try {
      isLoading.value = true;

<<<<<<< HEAD
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

            // ✅ IMPROVED: Handle current chat deletion better
            const isDeletingCurrentChat = currentChatId.value === chatId;

            // Remove chat from array
            chats.value.splice(chatIndex, 1);

            // If we deleted the current chat, switch to another one or clear
            if (isDeletingCurrentChat) {
              if (chats.value.length > 0) {
                // Switch to the first chat (will trigger navigation via watcher)
                currentChatId.value = chats.value[0].id;
              } else {
                // No chats left - clear current chat ID
                currentChatId.value = "";
                // Navigation to /new will be handled by router or component logic
              }
              updateExpandedArray();
            }

            clearCurrentDraft();
            saveChats();

            // Trigger sync after deleting chat
            if (isAuthenticated.value && parsedUserDetails.value?.syncEnabled) {
              setTimeout(() => {
                performSmartSync();
              }, 1000);
            }

            toast.success("Chat deleted", {
              duration: 3000,
              description: `"${chatTitle}" has been removed`,
            });
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
=======
      // 1. Load from localStorage first (instant)
      if (!forceFetch) {
        const cached = loadFromStorage();
        if (cached.length > 0) {
          chats.value = cached;
          console.log("📦 Loaded", cached.length, "chats from cache");
>>>>>>> new
        }
      }

      // 2. Fetch fresh from server in background (if authenticated and online)
      if (isAuthenticated.value && isOnline.value) {
        const response = await fetch("/api/chats", {
          headers: {
            "X-User-ID": parsedUserDetails.value.userId,
          },
        });

        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }

        const result = await response.json();

        if (!validateApiResponse<Chat[]>(result)) {
          throw new Error("Invalid API response format");
        }

        const serverChats = Array.isArray(result.data) ? result.data : [];

        // Validate each chat
        const validChats = serverChats.filter((chat) => {
          if (!validateChat(chat)) {
            console.warn("Invalid chat data:", chat);
            return false;
          }
          return true;
        });

        // 3. Merge server data with local
        chats.value = mergeChats(validChats, chats.value);
        saveToStorage();

        syncStatus.value.lastSync = new Date();
        syncStatus.value.hasUnsyncedChanges = syncQueue.value.length > 0;
        console.log("✅ Synced", validChats.length, "chats from server");
      }

      // Set current chat if none selected
      if (chats.value.length > 0 && !currentChatId.value) {
        currentChatId.value = chats.value[0].id;
      }
    } catch (error) {
      console.error("Failed to load chats:", error);
      toast.error("Failed to sync chats from server");
      // Continue with cached data
    } finally {
      isLoading.value = false;
    }
  }

  // ============================================================================
  // CREATE CHAT: Optimistic UI
  // ============================================================================
  async function createNewChat(firstMessage?: string): Promise<string> {
    const tempId = `temp_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    const now = new Date().toISOString();

    // 1. Create optimistic chat
    const optimisticChat: ExtendedChat = {
      id: tempId,
      user_id: parsedUserDetails.value.userId,
      title: firstMessage ? generateChatTitle(firstMessage) : "New Chat",
      created_at: now,
      updated_at: now,
      is_archived: false,
      message_count: 0,
      last_message_at: now,
      messages: [],
      _status: "creating",
    };

    chats.value.unshift(optimisticChat);
    currentChatId.value = tempId;
    saveToStorage();

    // 2. Send to server in background
    if (isAuthenticated.value && isOnline.value) {
      try {
        const requestBody: CreateChatRequest = { title: optimisticChat.title };

<<<<<<< HEAD
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
=======
        const response = await fetch("/api/chats", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "X-User-ID": parsedUserDetails.value.userId,
          },
          body: JSON.stringify(requestBody),
>>>>>>> new
        });

        if (!response.ok) {
          throw new Error(`HTTP ${response.status}`);
        }

        const result = await response.json();

        if (!validateApiResponse<Chat>(result) || !validateChat(result.data)) {
          throw new Error("Invalid server response");
        }

        const serverChat = result.data;

        // 3. Replace temp with real data
        const index = chats.value.findIndex((c) => c.id === tempId);
        if (index !== -1) {
          tempIdMap.set(tempId, serverChat.id);

          chats.value[index] = {
            ...serverChat,
            messages: optimisticChat.messages,
            _status: "synced",
          };

          // Update current chat ID
          if (currentChatId.value === tempId) {
            currentChatId.value = serverChat.id;
            router.replace(`/chat/${serverChat.id}`);
          }

          saveToStorage();
          console.log("✅ Chat created on server:", serverChat.id);
        }

        return serverChat.id;
      } catch (error) {
        console.error("Failed to create chat on server:", error);

        // 4. Mark as failed
        const index = chats.value.findIndex((c) => c.id === tempId);
        if (index !== -1) {
          chats.value[index]._status = "failed";
          saveToStorage();
        }

        // Queue for retry
        syncQueue.value.push({
          type: "create",
          chatId: tempId,
          data: { title: optimisticChat.title },
          timestamp: Date.now(),
          retries: 0,
        });

        syncStatus.value.hasUnsyncedChanges = true;
        toast.error("Chat saved locally, will sync when online");

        return tempId;
      }
    } else {
      // Offline mode
      syncQueue.value.push({
        type: "create",
        chatId: tempId,
        data: { title: optimisticChat.title },
        timestamp: Date.now(),
        retries: 0,
      });
      syncStatus.value.hasUnsyncedChanges = true;
      return tempId;
    }
  }

  // ============================================================================
  // SEND MESSAGE: Optimistic UI
  // ============================================================================
  async function sendMessage(
    chatId: string,
    content: string,
    role: "user" | "assistant" = "user",
    model?: string,
  ): Promise<void> {
    const chat = chats.value.find((c) => c.id === chatId);
    if (!chat) {
      toast.error("Chat not found");
      return;
    }

    const tempMsgId = `temp_msg_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;

<<<<<<< HEAD
      // Clear paste preview for previous chat
      if (currentChatId.value) {
        pastePreviews.value.delete(currentChatId.value);

        // Save current draft before switching
        const textarea = document.getElementById(
          "prompt",
        ) as HTMLTextAreaElement;
        const currentDraft = textarea?.value || "";
        if (currentDraft.trim()) {
          chatDrafts.value.set(currentChatId.value, currentDraft);
          saveChatDrafts();
=======
    // 1. Optimistic message
    const optimisticMsg: ExtendedMessage = {
      id: tempMsgId,
      chat_id: chatId,
      role,
      content,
      created_at: new Date().toISOString(),
      model,
      _status: "sending",
    };

    if (!chat.messages) {
      chat.messages = [];
    }
    chat.messages.push(optimisticMsg);
    chat.updated_at = new Date().toISOString();
    chat.message_count = (chat.message_count || 0) + 1;
    chat.last_message_at = optimisticMsg.created_at;
    saveToStorage();

    // Check if chat is still temp (not synced yet)
    const realChatId = tempIdMap.get(chatId) || chatId;
    const isTempChat = realChatId.startsWith("temp_");

    // 2. Send to server in background
    if (isAuthenticated.value && isOnline.value && !isTempChat) {
      try {
        const requestBody: CreateMessageRequest = { role, content, model };

        const response = await fetch(`/api/chats/${realChatId}/messages`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "X-User-ID": parsedUserDetails.value.userId,
          },
          body: JSON.stringify(requestBody),
        });

        if (!response.ok) {
          throw new Error(`HTTP ${response.status}`);
>>>>>>> new
        }

<<<<<<< HEAD
      // Initialize new chat drafts
      chatDrafts.value.set(newChatId, "");
      pastePreviews.value.delete(newChatId);

      // Add new chat to beginning of list
      chats.value.unshift(newChat);

      // Update current chat ID (will trigger navigation via watcher)
      currentChatId.value = newChatId;

      updateExpandedArray();
      saveChats();

      // Only sync if we have meaningful changes
      // Don't sync immediately for empty new chats
      if (isAuthenticated.value && parsedUserDetails.value?.syncEnabled) {
        // Mark as having changes but don't force immediate sync
        syncStatus.value.hasUnsyncedChanges = true;
      }

      // Focus the textarea
      nextTick(() => {
        const textarea = document.getElementById(
          "prompt",
        ) as HTMLTextAreaElement;
        if (textarea) {
          textarea.value = "";
          autoGrow({ target: textarea } as any);
          textarea.focus();
=======
        const result = await response.json();

        if (
          !validateApiResponse<Message>(result) ||
          !validateMessage(result.data)
        ) {
          throw new Error("Invalid server response");
>>>>>>> new
        }

        const serverMsg = result.data;

        // 3. Replace temp with real message
        const msgIndex = chat.messages!.findIndex((m) => m.id === tempMsgId);
        if (msgIndex !== -1) {
          (chat.messages as ExtendedMessage[])![msgIndex] = {
            ...serverMsg,
            _status: "sent",
          };
          saveToStorage();
          console.log("✅ Message sent:", serverMsg.id);
        }
      } catch (error) {
        console.error("Failed to send message:", error);

        // 4. Mark as failed
        const msgIndex = chat.messages!.findIndex((m) => m.id === tempMsgId);
        if (msgIndex !== -1) {
          (chat.messages as ExtendedMessage[])![msgIndex]._status = "failed";
          saveToStorage();
        }

        // Queue for retry
        syncQueue.value.push({
          type: "message",
          chatId: realChatId,
          data: { role, content, model, tempId: tempMsgId },
          timestamp: Date.now(),
          retries: 0,
          dependsOn: isTempChat ? chatId : undefined,
        });

        syncStatus.value.hasUnsyncedChanges = true;
        toast.error("Message saved locally, will retry");
      }
    } else {
      // Offline or temp chat
      optimisticMsg._status = isTempChat ? "pending_chat_sync" : "failed";

      syncQueue.value.push({
        type: "message",
        chatId: realChatId,
        data: { role, content, model, tempId: tempMsgId },
        timestamp: Date.now(),
        retries: 0,
        dependsOn: isTempChat ? chatId : undefined,
      });

<<<<<<< HEAD
      console.log(`✅ Created new chat: ${newChatId}`);
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
=======
      syncStatus.value.hasUnsyncedChanges = true;
      saveToStorage();
>>>>>>> new
    }
  }

  // ============================================================================
  // DELETE CHAT: Optimistic removal
  // ============================================================================
  async function deleteChat(chatId: string) {
    const chatIndex = chats.value.findIndex((c) => c.id === chatId);
    if (chatIndex === -1) {
      toast.error("Chat not found");
      return;
    }

    const chat = chats.value[chatIndex];

    showConfirmDialog({
      visible: true,
      title: "Delete Chat",
      message: `Delete "${chat.title}"? This action cannot be undone.`,
      type: "danger",
      confirmText: "Delete",
      cancelText: "Cancel",
      onConfirm: async () => {
        // 1. Remove immediately (optimistic)
        const deletedChat = chats.value.splice(chatIndex, 1)[0];
        saveToStorage();

        // Switch to another chat if needed
        if (currentChatId.value === chatId) {
          if (chats.value.length > 0) {
            switchToChat(chats.value[0].id);
          } else {
            currentChatId.value = null;
            router.push("/");
          }
        }

        // 2. Send delete to server in background
        const realChatId = tempIdMap.get(chatId) || chatId;

        if (
          isAuthenticated.value &&
          isOnline.value &&
          !realChatId.startsWith("temp_")
        ) {
          try {
<<<<<<< HEAD
            // Clear all data
            chats.value = [];
            currentChatId.value = ""; // This will trigger navigation
            expanded.value = [];
            linkPreviewCache.value.clear();
            chatDrafts.value.clear();
            pastePreviews.value.clear();

            // Save changes
            saveChatDrafts();

            // Clean up localStorage
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

            // ✅ IMPROVED: Mark for sync after clearing
            if (isAuthenticated.value && parsedUserDetails.value?.syncEnabled) {
              syncStatus.value.hasUnsyncedChanges = true;
              setTimeout(() => {
                performSmartSync();
              }, 1000);
            }

            toast.success("All chats cleared", {
              duration: 5000,
              description: `Deleted ${totalChats} chats with ${totalMessages} messages`,
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
=======
            const response = await fetch(`/api/chats/${realChatId}`, {
              method: "DELETE",
              headers: {
                "X-User-ID": parsedUserDetails.value.userId,
              },
            });

            if (!response.ok) {
              throw new Error(`HTTP ${response.status}`);
            }

            console.log("✅ Chat deleted on server");
            toast.success("Chat deleted");
          } catch (error) {
            console.error("Failed to delete chat on server:", error);

            // Rollback - find correct position and restore
            const insertIndex = chats.value.findIndex(
              (c) =>
                new Date(c.updated_at).getTime() <
                new Date(deletedChat.updated_at).getTime(),
            );

            if (insertIndex === -1) {
              chats.value.push(deletedChat);
            } else {
              chats.value.splice(insertIndex, 0, deletedChat);
            }

            saveToStorage();
            toast.error("Failed to delete chat. Restored locally.");
>>>>>>> new
          }
        } else {
          // Just remove locally for temp chats
          toast.success("Chat deleted");
        }

        // Remove from temp map
        tempIdMap.delete(chatId);

        // Remove related items from sync queue
        syncQueue.value = syncQueue.value.filter(
          (item) => item.chatId !== chatId && item.dependsOn !== chatId,
        );
      },
      onCancel: () => {
        console.log("Delete cancelled");
      },
    });
  }

  // ============================================================================
  // RENAME CHAT: Optimistic update
  // ============================================================================
  async function renameChat(chatId: string, newTitle: string) {
    const chat = chats.value.find((c) => c.id === chatId);
    if (!chat) {
      toast.error("Chat not found");
      return;
    }

    const oldTitle = chat.title;
    const trimmedTitle = newTitle.trim();

    if (!trimmedTitle) {
      toast.error("Title cannot be empty");
      return;
    }

    // 1. Update immediately (optimistic)
    chat.title = trimmedTitle;
    chat.updated_at = new Date().toISOString();
    saveToStorage();

    // 2. Send to server in background
    const realChatId = tempIdMap.get(chatId) || chatId;

    if (
      isAuthenticated.value &&
      isOnline.value &&
      !realChatId.startsWith("temp_")
    ) {
      try {
        const response = await fetch(`/api/chats/${realChatId}`, {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
            "X-User-ID": parsedUserDetails.value.userId,
          },
          body: JSON.stringify({ title: trimmedTitle }),
        });

        if (!response.ok) {
          throw new Error(`HTTP ${response.status}`);
        }

        console.log("✅ Chat renamed on server");
        toast.success("Chat renamed");
      } catch (error) {
        console.error("Failed to rename chat on server:", error);

        // Rollback
        chat.title = oldTitle;
        saveToStorage();
        toast.error("Failed to rename chat");
      }
    } else {
      // Queue for later if temp chat
      if (realChatId.startsWith("temp_")) {
        syncQueue.value.push({
          type: "update",
          chatId: realChatId,
          data: { title: trimmedTitle },
          timestamp: Date.now(),
          retries: 0,
        });
        syncStatus.value.hasUnsyncedChanges = true;
      }
    }
  }

  // ============================================================================
  // SYNC QUEUE: Process failed operations
  // ============================================================================
  async function processSyncQueue() {
    if (
      !isAuthenticated.value ||
      !isOnline.value ||
      syncQueue.value.length === 0 ||
      isSyncing.value
    ) {
      return;
    }

    isSyncing.value = true;
    syncStatus.value.syncing = true;
    const maxRetries = 3;
    const itemsToRetry: SyncQueueItem[] = [];

    // Sort queue: creates first, then updates, then messages, then deletes
    const sortedQueue = [...syncQueue.value].sort((a, b) => {
      const order = { create: 0, update: 1, message: 2, delete: 3 };
      const orderDiff = order[a.type] - order[b.type];
      if (orderDiff !== 0) return orderDiff;
      return a.timestamp - b.timestamp;
    });

    for (const item of sortedQueue) {
      try {
        // Check dependencies
        if (item.dependsOn) {
          const dependency = chats.value.find((c) => c.id === item.dependsOn);
          if (!dependency || dependency._status !== "synced") {
            console.log(
              `Skipping ${item.type} - waiting for dependency ${item.dependsOn}`,
            );
            itemsToRetry.push(item);
            continue;
          }
        }

        if (item.type === "create") {
          // Retry chat creation
          const response = await fetch("/api/chats", {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
              "X-User-ID": parsedUserDetails.value.userId,
            },
            body: JSON.stringify({ title: item.data.title }),
          });

          if (!response.ok) throw new Error(`HTTP ${response.status}`);

          const result = await response.json();
          if (
            !validateApiResponse<Chat>(result) ||
            !validateChat(result.data)
          ) {
            throw new Error("Invalid response");
          }

          // Update local chat with server ID
          const chat = chats.value.find((c) => c.id === item.chatId);
          if (chat) {
            const oldId = chat.id;
            Object.assign(chat, result.data);
            chat._status = "synced";
            tempIdMap.set(oldId, result.data.id);

            // Update current chat ID if needed
            if (currentChatId.value === oldId) {
              currentChatId.value = result.data.id;
              router.replace(`/chat/${result.data.id}`);
            }
          }

          console.log("✅ Retried chat creation:", result.data.id);
        } else if (item.type === "message") {
          // Get real chat ID
          const realChatId = tempIdMap.get(item.chatId) || item.chatId;

          // Skip if chat is still temp
          if (realChatId.startsWith("temp_")) {
            itemsToRetry.push({
              ...item,
              dependsOn: item.chatId,
            });
            continue;
          }

          const response = await fetch(`/api/chats/${realChatId}/messages`, {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
              "X-User-ID": parsedUserDetails.value.userId,
            },
            body: JSON.stringify({
              role: item.data.role,
              content: item.data.content,
              model: item.data.model,
            }),
          });

          if (!response.ok) throw new Error(`HTTP ${response.status}`);

          const result = await response.json();
          if (
            !validateApiResponse<Message>(result) ||
            !validateMessage(result.data)
          ) {
            throw new Error("Invalid response");
          }

          // Update message status
          const chat = chats.value.find(
            (c) => c.id === item.chatId || c.id === realChatId,
          );
          if (chat?.messages) {
            const msg = chat.messages.find(
              (m: ExtendedMessage) => m.id === item.data.tempId,
            );
            if (msg) {
              Object.assign(msg, result.data);
              (msg as ExtendedMessage)._status = "sent";
            }
          }

          console.log("✅ Retried message send:", result.data.id);
        } else if (item.type === "update") {
          const realChatId = tempIdMap.get(item.chatId) || item.chatId;

          if (realChatId.startsWith("temp_")) {
            itemsToRetry.push(item);
            continue;
          }

          const response = await fetch(`/api/chats/${realChatId}`, {
            method: "PUT",
            headers: {
              "Content-Type": "application/json",
              "X-User-ID": parsedUserDetails.value.userId,
            },
            body: JSON.stringify(item.data),
          });

          if (!response.ok) throw new Error(`HTTP ${response.status}`);

          console.log("✅ Retried chat update:", realChatId);
        }
      } catch (error) {
        console.error(`Failed to retry ${item.type}:`, error);

        // Requeue if under retry limit
        if (item.retries < maxRetries) {
          itemsToRetry.push({
            ...item,
            retries: item.retries + 1,
          });
        } else {
          console.error(`Max retries exceeded for ${item.type}`);
          toast.error(
            `Failed to sync ${item.type} after ${maxRetries} attempts`,
          );
        }
      }
    }

    // Update queue with items to retry
    syncQueue.value = itemsToRetry;

    // Update sync status
    syncStatus.value.syncing = false;
    syncStatus.value.hasUnsyncedChanges = syncQueue.value.length > 0;
    isSyncing.value = false;

    if (syncQueue.value.length === 0) {
      saveToStorage();
      toast.success("All changes synced");
      syncStatus.value.lastSync = new Date();
    }
  }

  // ============================================================================
  // PERIODIC SYNC: Run every 30 seconds when online
  // ============================================================================
  function startPeriodicSync() {
    if (syncInterval) clearInterval(syncInterval);

    syncInterval = setInterval(() => {
      if (isAuthenticated.value && isOnline.value && !isSyncing.value) {
        processSyncQueue();
      }
    }, 30000);
  }

  function stopPeriodicSync() {
    if (syncInterval) {
      clearInterval(syncInterval);
      syncInterval = null;
    }
    if (saveTimer) {
      clearTimeout(saveTimer);
      saveTimer = null;
    }
  }

  // ============================================================================
  // MANUAL SYNC: User-triggered sync
  // ============================================================================
  async function manualSync() {
    if (!isOnline.value) {
      toast.error("You're offline. Please check your connection.");
      return;
    }

    toast.info("Syncing...");
    await loadChats(true);
    await processSyncQueue();
  }

  // ============================================================================
  // LIFECYCLE: Setup and cleanup
  // ============================================================================
  onMounted(() => {
    const cleanupOnlineDetection = setupOnlineDetection();
    startPeriodicSync();

    // Store cleanup function
    onUnmounted(() => {
      cleanupOnlineDetection();
      stopPeriodicSync();
    });
  });

  // ============================================================================
  // EXPORTS
  // ============================================================================
  return {
    loadChats,
    createNewChat,
    sendMessage,
    deleteChat,
    renameChat,
    manualSync,
    processSyncQueue,
    syncQueue,
    isOnline,
  };
}
