// composables/useImageOCR.ts
import { ref } from "vue";
import { toast } from "vue-sonner";
import { API_BASE_URL } from "@/lib/globals";
import type { ApiResponse, OCRData, UserDetails } from "@/types";

export function useImageOCR() {
  const imageFile = ref<File | null>(null);
  const croppedBlob = ref<Blob | null>(null);
  const extractedData = ref<OCRData>({
    text: "",
    processing_ms: 0,
    text_length: 0,
  });
  const loading = ref(false);
  const error = ref<string | null>(null);
  const parsedUserDetails = ref<UserDetails>(
    JSON.parse(localStorage.getItem("userdetails") || "{}"),
  );

  // Capture image from <input type="file" accept="image/*" capture="environment">
  const captureImage = (file: File) => {
    imageFile.value = file;
    error.value = null;
    extractedData.value.text = "";
  };

  // Crop image using HTML canvas
  const cropImage = async (
    file: File,
    cropArea: { x: number; y: number; width: number; height: number },
  ): Promise<Blob> => {
    return new Promise((resolve, reject) => {
      const img = new Image();

      img.onload = () => {
        const canvas = document.createElement("canvas");
        canvas.width = cropArea.width;
        canvas.height = cropArea.height;
        const ctx = canvas.getContext("2d");

        if (!ctx) {
          reject(new Error("Canvas not supported"));
          return;
        }

        ctx.drawImage(
          img,
          cropArea.x,
          cropArea.y,
          cropArea.width,
          cropArea.height,
          0,
          0,
          cropArea.width,
          cropArea.height,
        );

        canvas.toBlob(
          (blob) => {
            if (blob) {
              croppedBlob.value = blob;
              resolve(blob);
            } else {
              reject(new Error("Failed to crop image"));
            }
          },
          "image/jpeg",
          0.95,
        ); // Added quality parameter
      };

      img.onerror = () => reject(new Error("Failed to load image"));
      img.src = URL.createObjectURL(file);
    });
  };

  // Send image (or cropped image) to Go backend for OCR
  const sendToBackend = async (useCropped: boolean = false) => {
    const blobToSend = useCropped ? croppedBlob.value : imageFile.value;

    if (!blobToSend) {
      error.value = "No image available to process";
      toast.error("Please select an image first");
      return;
    }

    try {
      loading.value = true;
      error.value = null;

      const formData = new FormData();
      formData.append("file", blobToSend, "capture.jpg");

      // Correct fetch syntax with template literal
      const response = await fetch(`${API_BASE_URL}/ocr/upload`, {
        method: "POST",
        headers: {
          ...(parsedUserDetails.value?.userId
            ? { "X-User-ID": parsedUserDetails.value.userId }
            : {}),
        },
        body: formData,
      });

      const data: ApiResponse<OCRData> = await response.json();

      if (!response.ok) {
        throw new Error(data.message || `HTTP error ${response.status}`);
      }

      if (!data.success) {
        throw new Error(data.message || "OCR failed");
      }

      extractedData.value = data.data as OCRData;
      console.log("Extracted Data:", extractedData.value);
      toast.success(data.message || "Text extracted successfully!");
    } catch (err: any) {
      error.value = err.message || "Failed to extract text";
      toast.error(`${err.message || "OCR failed"}`);
      console.error("OCR error:", err);
    } finally {
      loading.value = false;
    }
  };

  // Reset all state
  const reset = () => {
    imageFile.value = null;
    croppedBlob.value = null;
    extractedData.value.text = "";
    extractedData.value.processing_ms = 0;
    extractedData.value.text_length = 0;
    error.value = null;
    loading.value = false;
  };

  // Validate image file
  const validateImage = (file: File): boolean => {
    const maxSize = 4 * 1024 * 1024; // 4MB
    const allowedTypes = ["image/jpeg", "image/jpg", "image/png", "image/webp"];

    if (!allowedTypes.includes(file.type)) {
      error.value = "Invalid file type. Please upload PNG, JPG, or WebP";
      toast.error("Invalid file type");
      return false;
    }

    if (file.size > maxSize) {
      error.value = "File too large. Maximum size is 4MB";
      toast.error("File too large (max 4MB)");
      return false;
    }

    return true;
  };

  return {
    imageFile,
    croppedBlob,
    extractedData,
    loading,
    error,
    captureImage,
    cropImage,
    sendToBackend,
    reset,
    validateImage,
  };
}
