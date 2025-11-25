package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/imrany/gemmie/gemmie-server/internal/encrypt"
	"github.com/imrany/gemmie/gemmie-server/internal/genai"
	"github.com/imrany/gemmie/gemmie-server/store"
	"github.com/spf13/viper"
)

// CreateChatHandler handles POST /api/chats
func CreateChatHandler(w http.ResponseWriter, r *http.Request) {
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

	// Verify user exists
	user, err := store.GetUserByID(userID)
	if err != nil {
		slog.Error("Database error", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Database error",
		})
		return
	}

	if user == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User not found",
		})
		return
	}

	var req store.Chat
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Create new chat
	chat := store.Chat{
		ID:            req.ID,
		UserId:        userID,
		Title:         req.Title,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		IsArchived:    req.IsArchived,
		MessageCount:  0,
		Messages:      []store.Message{},
		LastMessageAt: time.Now(),
		IsPrivate:     req.IsPrivate,
	}

	if err := store.CreateChat(chat); err != nil {
		slog.Error("Failed to create chat", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to create chat",
		})
		return
	}

	slog.Info("Chat created successfully", "chat_id", chat.ID, "user_id", userID)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Chat created successfully",
		Data:    chat,
	})
}

// GetChatsHandler handles GET /api/chats
func GetChatsHandler(w http.ResponseWriter, r *http.Request) {
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

	// Verify user exists
	user, err := store.GetUserByID(userID)
	if err != nil {
		slog.Error("Database error", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Database error",
		})
		return
	}

	if user == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User not found",
		})
		return
	}

	// Get user chats
	chats, err := store.GetChatsByUserId(userID)
	if err != nil {
		slog.Error("Failed to get chats", "user_id", userID, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to retrieve chats",
		})
		return
	}

	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Chats retrieved successfully",
		Data:    chats,
	})
}

// GetChatHandler handles GET /api/chats/{id}
func GetChatHandler(w http.ResponseWriter, r *http.Request) {
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

	vars := mux.Vars(r)
	chatID := vars["id"]

	if chatID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Chat ID is required",
		})
		return
	}

	// Get chat
	chat, err := store.GetChatById(chatID)
	if err != nil {
		slog.Error("Failed to get chat", "chat_id", chatID, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to retrieve chat",
		})
		return
	}

	if chat == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Chat not found",
		})
		return
	}

	// Verify chat belongs to user
	if chat.UserId != userID {
		if chat.IsPrivate == true {
			// If chat is private, access denied
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: "Access denied",
			})
			return
		}
		chat.IsReadOnly = true
	}

	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Chat retrieved successfully",
		Data:    *chat,
	})
}

// UpdateChatHandler handles PUT /api/chats/{id}
func UpdateChatHandler(w http.ResponseWriter, r *http.Request) {
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

	vars := mux.Vars(r)
	chatID := vars["id"]

	if chatID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Chat ID is required",
		})
		return
	}

	var req store.Chat
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Get existing chat
	chat, err := store.GetChatById(chatID)
	if err != nil {
		slog.Error("Failed to get chat", "chat_id", chatID, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to retrieve chat",
		})
		return
	}

	if chat == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Chat not found",
		})
		return
	}

	// Verify chat belongs to user
	if chat.UserId != userID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Access denied",
		})
		return
	}

	// Update chat fields
	if req.Title != "" {
		chat.Title = req.Title
	}
	if !req.LastMessageAt.IsZero() {
		chat.LastMessageAt = req.LastMessageAt
	}
	chat.IsArchived = req.IsArchived
	if req.IsPrivate != chat.IsPrivate {
		chat.IsPrivate = req.IsPrivate
	}
	chat.UpdatedAt = time.Now()

	if err := store.UpdateChat(*chat); err != nil {
		slog.Error("Failed to update chat", "chat_id", chatID, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to update chat",
		})
		return
	}

	slog.Info("Chat updated successfully", "chat_id", chatID, "user_id", userID)

	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Chat updated successfully",
		Data:    chat,
	})
}

// DeleteChatHandler handles DELETE /api/chats/{id}
func DeleteChatHandler(w http.ResponseWriter, r *http.Request) {
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

	vars := mux.Vars(r)
	chatID := vars["id"]

	if chatID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Chat ID is required",
		})
		return
	}

	// Get chat to verify ownership
	chat, err := store.GetChatById(chatID)
	if err != nil {
		slog.Error("Failed to get chat", "chat_id", chatID, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to retrieve chat",
		})
		return
	}

	if chat == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Chat not found",
		})
		return
	}

	// Verify chat belongs to user
	if chat.UserId != userID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Access denied",
		})
		return
	}

	// Delete the chat
	if err := store.DeleteChatByID(chatID); err != nil {
		slog.Error("Failed to delete chat", "chat_id", chatID, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to delete chat",
		})
		return
	}

	slog.Info("Chat deleted successfully", "chat_id", chatID, "user_id", userID)

	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Chat deleted successfully",
	})
}

// DeleteAllChats handles DELETE /api/chats
func DeleteAllChatsHandler(w http.ResponseWriter, r *http.Request) {
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

	if err := store.DeleteAllChatsByUserID(userID); err != nil {
		slog.Error("Failed to delete all chats", "user_id", userID, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to delete all chats",
		})
		return
	}

	slog.Info("All chats deleted successfully", "user_id", userID)

	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "All chats deleted successfully",
	})
}

// CreateMessageHandler handles POST /api/chats/{id}/messages
func CreateMessageHandler(w http.ResponseWriter, r *http.Request) {
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

	vars := mux.Vars(r)
	chatID := vars["id"]

	if chatID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Chat ID is required",
		})
		return
	}

	// send request to ai wrapper
	var req store.Message
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Invalid request body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Validate required fields
	if req.Prompt == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Prompt is required",
		})
		return
	}

	// generate response using ai wrapper
	genaiResp := genai.GENAISERVICE{
		APIKey: viper.GetString("API_KEY"),
		Model:  viper.GetString("MODEL"),
	}
	genAIResponse, err := genaiResp.GenerateAIResponse(context.Background(), req.Prompt)
	if err != nil {
		slog.Error("Failed to generate AI response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: fmt.Sprintf("Failed to generate AI response: %s", err.Error()),
		})
		return
	}

	// Get chat to verify ownership
	chat, err := store.GetChatById(chatID)
	if err != nil {
		slog.Error("Failed to get chat", "chat_id", chatID, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to retrieve chat",
		})
		return
	}

	if chat == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Chat not found",
		})
		return
	}

	// Verify chat belongs to user
	if chat.UserId != userID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Access denied",
		})
		return
	}

	// Create new message
	message := store.Message{
		ID:         encrypt.GenerateID(nil),
		ChatId:     chatID,
		Prompt:     genAIResponse.Prompt,
		Response:   genAIResponse.Response,
		CreatedAt:  time.Now(),
		Model:      genaiResp.Model,
		References: req.References,
	}

	if err := store.CreateMessage(message); err != nil {
		slog.Error("Failed to create message", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to create message",
		})
		return
	}

	// Update chat's message count and last message time
	chat.MessageCount++
	chat.LastMessageAt = time.Now()
	chat.UpdatedAt = time.Now()

	if err := store.UpdateChat(*chat); err != nil {
		slog.Error("Failed to update chat", "chat_id", chatID, "error", err)
		// Don't fail the request, message was created successfully
	}

	slog.Info("Message created successfully", "message_id", message.ID, "chat_id", chatID, "user_id", userID)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Message created successfully",
		Data:    message,
	})
}

// DeleteMessageHandler handles DELETE /api/messages/{id}
func DeleteMessageHandler(w http.ResponseWriter, r *http.Request) {
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

	vars := mux.Vars(r)
	messageID := vars["id"]

	if messageID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Message ID is required",
		})
		return
	}

	// Get message to verify ownership
	message, err := store.GetMessageById(messageID)
	if err != nil {
		slog.Error("Failed to get message", "message_id", messageID, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to retrieve message",
		})
		return
	}

	if message == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Message not found",
		})
		return
	}

	// Get chat to verify ownership
	chat, err := store.GetChatById(message.ChatId)
	if err != nil {
		slog.Error("Failed to get chat", "chat_id", message.ChatId, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to retrieve chat",
		})
		return
	}

	if chat == nil || chat.UserId != userID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Access denied",
		})
		return
	}

	// Delete the message
	if err := store.DeleteMessageByID(messageID); err != nil {
		slog.Error("Failed to delete message", "message_id", messageID, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to delete message",
		})
		return
	}

	slog.Info("Message deleted successfully", "message_id", messageID, "user_id", userID)

	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Message deleted successfully",
	})
}

// UpdateMessageHandler handles PUT /api/messages/{id}
func UpdateMessageHandler(w http.ResponseWriter, r *http.Request) {
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

	vars := mux.Vars(r)
	messageID := vars["id"]

	if messageID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Message ID is required",
		})
		return
	}

	// Get message to verify ownership
	message, err := store.GetMessageById(messageID)
	if err != nil {
		slog.Error("Failed to get message", "message_id", messageID, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to retrieve message",
		})
		return
	}

	if message == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Message not found",
		})
		return
	}

	// Get chat to verify ownership
	chat, err := store.GetChatById(message.ChatId)
	if err != nil {
		slog.Error("Failed to get chat", "chat_id", message.ChatId, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to retrieve chat",
		})
		return
	}

	if chat == nil || chat.UserId != userID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Access denied",
		})
		return
	}

	// Parse request body
	var req store.Message
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode request body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	message.Prompt = req.Prompt
	message.Response = req.Response
	message.Model = req.Model
	message.References = req.References

	// Update message
	if err := store.UpdateMessage(*message); err != nil {
		slog.Error("Failed to update message", "message_id", messageID, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to update message",
		})
		return
	}

	slog.Info("Message updated successfully", "message_id", messageID, "user_id", userID)

	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Message updated successfully",
	})
}
