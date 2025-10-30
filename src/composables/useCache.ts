import type { LinkPreview } from "@/types";
import { ref } from "vue";

export function useCache() {
  const linkPreviewCache = ref<Map<string, LinkPreview>>(new Map());

  function loadLinkPreviewCache() {
    try {
      const cached = localStorage.getItem("linkPreviews");
      if (cached) {
        const parsedCache = JSON.parse(cached);
        if (typeof parsedCache === "object" && parsedCache !== null) {
          linkPreviewCache.value = new Map(Object.entries(parsedCache));
        }
      }
    } catch (error) {
      console.error("Failed to load link preview cache:", error);
      localStorage.removeItem("linkPreviews");
      linkPreviewCache.value.clear();
    }
  }

  function saveLinkPreviewCache() {
    try {
      const cacheObject = Object.fromEntries(linkPreviewCache.value);
      localStorage.setItem("linkPreviews", JSON.stringify(cacheObject));
    } catch (error) {
      console.error("Failed to save link preview cache:", error);
      if (linkPreviewCache.value.size > 100) {
        const entries = Array.from(linkPreviewCache.value.entries());
        const recent = entries.slice(-50);
        linkPreviewCache.value = new Map(recent);
        try {
          const reducedCache = Object.fromEntries(linkPreviewCache.value);
          localStorage.setItem("linkPreviews", JSON.stringify(reducedCache));
        } catch (retryError) {
          console.error("Failed to save reduced cache:", retryError);
        }
      }
    }
  }

  return {
    linkPreviewCache,
    loadLinkPreviewCache,
    saveLinkPreviewCache,
  };
}
