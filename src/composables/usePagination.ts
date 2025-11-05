import type { Message } from "@/types";
import { nextTick, type Ref } from "vue";

export function usePagination({
  currentChatId,
  currentMessages,
  isDeepSearchResult,
  deepSearchPagination,
  scrollToLastMessage,
}: {
  currentChatId: Ref<string | undefined>;
  currentMessages: Ref<Message[]>;
  isDeepSearchResult: (response: string) => boolean;
  deepSearchPagination: Ref<
    Map<string, Map<number, { currentPage: number; totalPages: number }>>
  >;
  scrollToLastMessage: () => void;
}) {
  function getPagination(messageIndex: number) {
    if (!currentChatId.value) return { currentPage: 0, totalPages: 0 };

    const message = currentMessages.value[messageIndex];
    if (!message || !isDeepSearchResult(message.response)) {
      return { currentPage: 0, totalPages: 0 };
    }

    // Get or create chat pagination map
    let chatPagination = deepSearchPagination.value.get(currentChatId.value);
    if (!chatPagination) {
      chatPagination = new Map();
      deepSearchPagination.value.set(currentChatId.value, chatPagination);
    }

    // Get or initialize pagination for this message
    let pagination = chatPagination.get(messageIndex);
    if (!pagination) {
      // Extract total pages from the deep search result
      try {
        const parsed = JSON.parse(message.response);
        const totalPages = parsed.results?.length || 0;
        pagination = { currentPage: 0, totalPages };
        chatPagination.set(messageIndex, pagination);
      } catch (e) {
        pagination = { currentPage: 0, totalPages: 0 };
      }
    }

    return pagination;
  }

  // Navigation functions
  function nextResult(messageIndex: number) {
    if (!currentChatId.value) return;

    const pagination = getPagination(messageIndex);
    if (pagination.currentPage < pagination.totalPages - 1) {
      // Get or create chat pagination map
      let chatPagination = deepSearchPagination.value.get(currentChatId.value);
      if (!chatPagination) {
        chatPagination = new Map();
        deepSearchPagination.value.set(currentChatId.value, chatPagination);
      }

      // Update the specific message pagination
      chatPagination.set(messageIndex, {
        ...pagination,
        currentPage: pagination.currentPage + 1,
      });

      // Force reactivity
      deepSearchPagination.value = new Map(deepSearchPagination.value);

      nextTick(() => scrollToLastMessage());
    }
  }

  function prevResult(messageIndex: number) {
    if (!currentChatId.value) return;

    const pagination = getPagination(messageIndex);
    if (pagination.currentPage > 0) {
      // Get or create chat pagination map
      let chatPagination = deepSearchPagination.value.get(currentChatId.value);
      if (!chatPagination) {
        chatPagination = new Map();
        deepSearchPagination.value.set(currentChatId.value, chatPagination);
      }

      // Update the specific message pagination
      chatPagination.set(messageIndex, {
        ...pagination,
        currentPage: pagination.currentPage - 1,
      });

      // Force reactivity
      deepSearchPagination.value = new Map(deepSearchPagination.value);

      nextTick(() => scrollToLastMessage());
    }
  }

  function goToPage(messageIndex: number, pageIndex: number) {
    if (!currentChatId.value) return;

    const pagination = getPagination(messageIndex);
    if (pageIndex >= 0 && pageIndex < pagination.totalPages) {
      // FIXED: Proper validation
      // Get or create chat pagination map
      let chatPagination = deepSearchPagination.value.get(currentChatId.value);
      if (!chatPagination) {
        chatPagination = new Map();
        deepSearchPagination.value.set(currentChatId.value, chatPagination);
      }

      // Update the specific message pagination
      chatPagination.set(messageIndex, {
        ...pagination,
        currentPage: pageIndex,
      });

      // Force reactivity
      deepSearchPagination.value = new Map(deepSearchPagination.value);

      nextTick(() => scrollToLastMessage());
    }
  }

  return {
    prevResult,
    goToPage,
    nextResult,
    getPagination,
  };
}
