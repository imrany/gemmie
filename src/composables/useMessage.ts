import type { Chat } from "@/types";
import type { Ref } from "vue";

export function useMessage({
  chats,
  currentChatId,
  deepSearchPagination,
  updateExpandedArray,
  saveChats,
}: {
  chats: Ref<Chat[]>;
  currentChatId: Ref<string>;
  deepSearchPagination: Ref<
    Map<string, Map<number, { currentPage: number; totalPages: number }>>
  >;
  updateExpandedArray: () => void;
  saveChats: () => void;
}) {
  const getLoadingMessage = (response: string): string => {
    if (response === "web-search..." || response === "light-search...")
      return "Web searching...";
    if (response === "deep-search...") return "Deep searching...";
    if (response === "light-response...") return "Thinking...";
    if (response === "refreshing...") return "Refreshing...";
    return "Thinking...";
  };

  function formatSearchResults(
    searchData: any,
    mode: string,
    messageIndex?: number,
  ): string {
    const results = searchData.results || searchData.json || [];
    if (results.length === 0) {
      return "No search results found for your query.";
    }

    // For deep-search, set up pagination
    if (mode === "deep-search" && messageIndex !== undefined) {
      // Get or create chat pagination map
      let chatPagination = deepSearchPagination.value.get(currentChatId.value);
      if (!chatPagination) {
        chatPagination = new Map();
        deepSearchPagination.value.set(currentChatId.value, chatPagination);
      }

      chatPagination.set(messageIndex, {
        currentPage: 0,
        totalPages: results.length,
      });

      // Force reactivity
      deepSearchPagination.value = new Map(deepSearchPagination.value);
    }

    // Store results as JSON for client-side pagination
    if (mode === "deep-search") {
      return JSON.stringify({
        mode: "deep-search",
        results: results,
        metadata: {
          total_pages: searchData.total_pages,
          content_depth: searchData.content_depth,
          search_time: searchData.search_time,
        },
      });
    }

    // For web-search, keep existing format (all results shown)
    let formatted = "";
    const { total_pages } = searchData;

    formatted += `Showing **${results.length}** of **${total_pages || results.length}** total results\n\n`;
    formatted += `\n\n`;

    results.forEach((result: any, index: number) => {
      const title = result.title || "No Title";
      const url = result.url || "#";
      const description = result.description || "No description available";

      formatted += `#### ${index + 1}. ${title} \n\n`;
      formatted += `${description} \n`;
      formatted += `[${url}](${url}) \n\n`;

      if (index < results.length - 1) {
        formatted += `\n\n\n`;
      }
    });

    return formatted;
  }

  function removeTemporaryMessage(chatId: string, messageIndex: number) {
    if (messageIndex < 0) return;

    const targetChat = chats.value.find((chat) => chat.id === chatId);
    if (targetChat && targetChat.messages.length > messageIndex) {
      // Remove the temporary message
      targetChat.messages.splice(messageIndex, 1);

      // If this was the only message, we might want to handle the empty chat
      if (targetChat.messages.length === 0) {
        // Optionally delete the empty chat or keep it with a default title
        targetChat.title = "New Chat";
      }

      targetChat.last_message_at = new Date().toISOString();
      updateExpandedArray();
      saveChats();
    }
  }

  return {
    getLoadingMessage,
    formatSearchResults,
    removeTemporaryMessage,
  };
}
