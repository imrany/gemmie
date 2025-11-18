import { ref } from "vue";
import type { UploadedFile } from "@/types/document";
import { generatePdfThumbnail } from "@/lib/pdfHelpers";
import type { Ref } from "vue";

export function useFileUpload(
  uploadedFiles: Ref<UploadedFile[]>,
  addFile: (file: UploadedFile) => void,
  openEditor: (file: UploadedFile) => void,
) {
  const isDragOver = ref(false);

  async function handleFileUpload(event: Event) {
    const target = event.target as HTMLInputElement;
    const files = target.files ? Array.from(target.files) : [];

    for (const file of files) {
      if (file.type === "application/pdf") {
        const url = URL.createObjectURL(file);
        const { previewUrl, pages } = await generatePdfThumbnail(file);

        const newFile: UploadedFile = {
          id: `${Date.now()}-${Math.random().toString(36).slice(2)}`,
          name: file.name,
          url,
          type: file.type,
          size: file.size,
          previewUrl,
          pages,
          uploadedAt: new Date(),
        };

        addFile(newFile);
        openEditor(newFile);
      }
    }

    target.value = "";
  }

  function handleDragOver(e: DragEvent) {
    e.preventDefault();
    isDragOver.value = true;
  }

  function handleDragLeave(e: DragEvent) {
    e.preventDefault();
    isDragOver.value = false;
  }

  function handleDrop(e: DragEvent) {
    e.preventDefault();
    isDragOver.value = false;
    const files = e.dataTransfer?.files;
    if (files) {
      const fakeEvent = { target: { files } } as any;
      handleFileUpload(fakeEvent);
    }
  }

  return {
    isDragOver,
    handleFileUpload,
    handleDragOver,
    handleDragLeave,
    handleDrop,
  };
}
