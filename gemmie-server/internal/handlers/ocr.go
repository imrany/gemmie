package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/imrany/gemmie/gemmie-server/store"
)

var (
	MaxFileSize       = int64(4 << 20) // 4 MB
	AllowedImageTypes = "image/jpeg,image/jpg,image/png,image/webp"
)

// OCRUploadHandler handles image upload and OCR text extraction
func OCRUploadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Validate user authentication
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User ID header required",
		})
		return
	}

	// Verify user exists
	user, err := store.GetUserByID(userID)
	if err != nil || user == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User not found",
		})
		return
	}

	// Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, MaxFileSize)

	// Parse multipart form
	if err := r.ParseMultipartForm(MaxFileSize); err != nil {
		slog.Error("Failed to parse multipart form", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "File too large or invalid request (max 10MB)",
		})
		return
	}

	// Get uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		slog.Error("Failed to read file", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to read uploaded file",
		})
		return
	}
	defer file.Close()

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if !isAllowedImageType(contentType) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: fmt.Sprintf("Invalid file type: %s. Allowed types: PNG, JPG, WebP", contentType),
		})
		return
	}

	// Validate file size
	if header.Size > MaxFileSize {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "File size exceeds 10MB limit",
		})
		return
	}

	// Read file into memory
	buf, err := io.ReadAll(file)
	if err != nil {
		slog.Error("Failed to read file into memory", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to read uploaded file",
		})
		return
	}

	slog.Info("File received successfully",
		"user_id", userID,
		"filename", header.Filename,
		"size_bytes", len(buf),
	)

	// Perform OCR directly from bytes
	startTime := time.Now()
	text, err := performOCRBytes(buf)
	duration := time.Since(startTime)

	if err != nil {
		slog.Error("OCR processing failed",
			"user_id", userID,
			"error", err,
			"duration", duration,
		)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "OCR processing failed: " + err.Error(),
		})
		return
	}

	// Check if text was extracted
	if strings.TrimSpace(text) == "" {
		slog.Info("No text found in image", "user_id", userID)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(store.Response{
			Success: true,
			Message: "No text found in the image",
			Data: map[string]any{
				"text":          "",
				"processing_ms": duration.Milliseconds(),
			},
		})
		return
	}

	slog.Info("OCR completed successfully",
		"user_id", userID,
		"text_length", len(text),
		"duration", duration,
	)

	// Return extracted text
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Text extracted successfully",
		Data: map[string]any{
			"text":          strings.TrimSpace(text),
			"text_length":   len(strings.TrimSpace(text)),
			"processing_ms": duration.Milliseconds(),
		},
	})
}

// performOCRBytes runs Tesseract OCR on image data provided as bytes
func performOCRBytes(imageData []byte) (string, error) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "ocr-*.png")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name()) // ensure cleanup

	// Write the bytes to the temp file
	if _, err := tmpFile.Write(imageData); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	// Run tesseract on the temp file
	cmd := exec.Command("tesseract", tmpFile.Name(), "stdout", "-l", "eng")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("tesseract failed: %s (%v)", stderr.String(), err)
	}

	return out.String(), nil
}

// isAllowedImageType checks if the content type is allowed
func isAllowedImageType(contentType string) bool {
	allowed := strings.SplitSeq(AllowedImageTypes, ",")
	for t := range allowed {
		if strings.EqualFold(contentType, t) {
			return true
		}
	}
	return false
}
