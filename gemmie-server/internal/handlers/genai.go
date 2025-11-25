package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/imrany/gemmie/gemmie-server/internal/genai"
	"github.com/imrany/gemmie/gemmie-server/store"
	"github.com/spf13/viper"
)

// GenerateAIResponseHandler - POST /api/genai
func GenerateAIResponseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User ID header required",
		})
		return
	}

	var prompt string
	if err := json.NewDecoder(r.Body).Decode(&prompt); err != nil {
		slog.Error("Invalid request body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}
	if prompt == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Prompt is required",
		})
		return
	}

	// Generate response using AI wrapper (only if response not provided)
	genaiResp := genai.GENAISERVICE{
		APIKey: viper.GetString("API_KEY"),
		Model:  viper.GetString("MODEL"),
	}

	genAIResponse, err := genaiResp.GenerateAIResponse(context.Background(), prompt)
	if err != nil {
		slog.Error("Failed to generate AI response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: fmt.Sprintf("Failed to generate AI response: %s", err.Error()),
		})
		return
	}

	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "successfully generated AI response",
		Data:    genAIResponse,
	})
}
