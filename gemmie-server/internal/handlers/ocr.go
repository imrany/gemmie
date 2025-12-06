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
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/imrany/gemmie/gemmie-server/store"
	"github.com/spf13/viper"
)

var (
	MaxFileSize       = int64(10 << 20) // 10 MB
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

	// Create temp directory if not exists
	if err := os.MkdirAll(viper.GetString("OCR_UPLOAD_DIR"), 0755); err != nil {
		slog.Error("Failed to create ocr upload directory", "error", err, "name", viper.GetString("OCR_UPLOAD_DIR"))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Server error: failed to create temporary directory",
		})
		return
	}

	// Generate unique filename
	fileExt := filepath.Ext(header.Filename)
	uniqueFilename := fmt.Sprintf("%s_%s%s", userID, uuid.New().String(), fileExt)
	tmpPath := filepath.Join(viper.GetString("OCR_UPLOAD_DIR"), uniqueFilename)

	// Save file temporarily
	out, err := os.Create(tmpPath)
	if err != nil {
		slog.Error("Failed to create temp file", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to save uploaded file",
		})
		return
	}
	defer out.Close()

	// Copy file contents
	written, err := io.Copy(out, file)
	if err != nil {
		slog.Error("Failed to write file", "error", err)
		os.Remove(tmpPath) // Cleanup
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to save file",
		})
		return
	}

	slog.Info("File uploaded successfully",
		"user_id", userID,
		"filename", header.Filename,
		"size_bytes", written,
		"temp_path", tmpPath,
	)

	// Perform OCR
	startTime := time.Now()
	text, err := performOCR(tmpPath)
	duration := time.Since(startTime)

	// Cleanup temp file
	if err := os.Remove(tmpPath); err != nil {
		slog.Warn("Failed to delete temp file", "path", tmpPath, "error", err)
	}

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
			Data: map[string]interface{}{
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

// performOCR runs Tesseract OCR on the image using the system binary
func performOCR(imagePath string) (string, error) {
	// Run tesseract and output to stdout
	cmd := exec.Command("tesseract", imagePath, "stdout", "-l", "eng")
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
	allowed := strings.Split(AllowedImageTypes, ",")
	for _, t := range allowed {
		if strings.EqualFold(contentType, t) {
			return true
		}
	}
	return false
}

// CleanupOldTempFiles removes temp files older than 1 hour (call this periodically)
func CleanupOldTempFiles() error {
	if _, err := os.Stat(viper.GetString("OCR_UPLOAD_DIR")); os.IsNotExist(err) {
		return nil // Directory doesn't exist, nothing to clean
	}

	entries, err := os.ReadDir(viper.GetString("OCR_UPLOAD_DIR"))
	if err != nil {
		return err
	}

	now := time.Now()
	cleanedCount := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Delete files older than 1 hour
		if now.Sub(info.ModTime()) > time.Hour {
			path := filepath.Join(viper.GetString("OCR_UPLOAD_DIR"), entry.Name())
			if err := os.Remove(path); err != nil {
				slog.Warn("Failed to cleanup old temp file", "path", path, "error", err)
			} else {
				cleanedCount++
			}
		}
	}

	if cleanedCount > 0 {
		slog.Info("Cleaned up old temp files", "count", cleanedCount)
	}

	return nil
}

// StartCleanupScheduler starts a background goroutine to cleanup old temp files
func StartCleanupScheduler() {
	ticker := time.NewTicker(30 * time.Minute)
	go func() {
		for range ticker.C {
			if err := CleanupOldTempFiles(); err != nil {
				slog.Error("Failed to cleanup temp files", "error", err)
			}
		}
	}()
	slog.Info("OCR temp file cleanup scheduler started")
}
