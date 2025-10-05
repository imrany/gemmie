import { ref, watch } from "vue";
import type { UploadedFile } from "@/types/document";

const STORAGE_KEY = "pdf_editor_documents";
const CURRENT_DOCUMENT_KEY = "pdf_editor_current_document";

export function useDocumentStore() {
  const uploadedFiles = ref<UploadedFile[]>([]);
  const selectedFileId = ref("");

  function saveToLocalStorage() {
    try {
      const documentsData = uploadedFiles.value.map((file) => ({
        ...file,
        url: file.url.startsWith("blob:") && !file.isCustom ? "" : file.url,
        content: file.isCustom ? file.content : undefined,
      }));
      localStorage.setItem(STORAGE_KEY, JSON.stringify(documentsData));
    } catch (error) {
      console.warn("Failed to save to localStorage:", error);
    }
  }

  function loadFromLocalStorage() {
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (stored) {
        const documentsData = JSON.parse(stored);
        uploadedFiles.value = documentsData.map((doc: any) => ({
          ...doc,
          uploadedAt: new Date(doc.uploadedAt),
        }));
      }
    } catch (error) {
      console.warn("Failed to load from localStorage:", error);
    }
  }

  function saveCurrentDocument() {
    if (selectedFileId.value) {
      localStorage.setItem(CURRENT_DOCUMENT_KEY, selectedFileId.value);
    }
  }

  function getCurrentDocumentId() {
    return localStorage.getItem(CURRENT_DOCUMENT_KEY);
  }

  function clearCurrentDocument() {
    localStorage.removeItem(CURRENT_DOCUMENT_KEY);
  }

  function removeFile(id: string) {
    const index = uploadedFiles.value.findIndex((file) => file.id === id);
    if (index > -1) {
      const file = uploadedFiles.value[index];
      if (file.url.startsWith("blob:")) {
        URL.revokeObjectURL(file.url);
      }
      uploadedFiles.value.splice(index, 1);
      saveToLocalStorage();

      if (selectedFileId.value === id) {
        clearCurrentDocument();
      }
    }
  }

  function addFile(file: UploadedFile) {
    uploadedFiles.value.push(file);
    saveToLocalStorage();
  }

  return {
    uploadedFiles,
    selectedFileId,
    saveToLocalStorage,
    loadFromLocalStorage,
    saveCurrentDocument,
    getCurrentDocumentId,
    clearCurrentDocument,
    removeFile,
    addFile,
  };
}
