package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/imrany/gemmie/gemmie-server/internal/encrypt"
	"github.com/imrany/gemmie/gemmie-server/store"
)

// CreateChatRequest represents request payload for creating a new chat
type CreateChatRequest struct {
	Title string `json:"title"`
}

// CreateMessageRequest represents request payload for creating a new message
type CreateMessageRequest struct {
	Role    string `json:"role"` // "user" or "assistant"
	Content string `json:"content"`
	Model   string `json:"model,omitempty"`
}

// UpdateChatRequest represents request payload for updating a chat
type UpdateChatRequest struct {
	Title      string `json:"title,omitempty"`
	IsArchived bool   `json:"is_archived,omitempty"`
}

// ChatResponse represents a chat with its messages
type ChatResponse struct {
	store.Chat
	Messages []store.Message `json:"messages,omitempty"`
}

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

	var req CreateChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Validate title
	if req.Title == "" {
		req.Title = "New Chat"
	}

	// Create new chat
	chat := store.Chat{
		ID:            encrypt.GenerateID(nil),
		UserId:        userID,
		Title:         req.Title,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		IsArchived:    false,
		MessageCount:  0,
		LastMessageAt: time.Now(),
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
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Access denied",
		})
		return
	}

	// Get messages for this chat
	messages, err := store.GetMessagesByChatId(chatID)
	if err != nil {
		slog.Error("Failed to get messages", "chat_id", chatID, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to retrieve messages",
		})
		return
	}

	response := ChatResponse{
		Chat:     *chat,
		Messages: messages,
	}

	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Chat retrieved successfully",
		Data:    response,
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

	var req UpdateChatRequest
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
	chat.IsArchived = req.IsArchived
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

	// Delete all messages in the chat first
	if err := store.DeleteAllMessageByChatID(chatID); err != nil {
		slog.Error("Failed to delete chat messages", "chat_id", chatID, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to delete chat messages",
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

	var req CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Validate required fields
	if req.Role == "" || req.Content == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Role and content are required",
		})
		return
	}

	// Validate role
	if req.Role != "user" && req.Role != "assistant" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Role must be 'user' or 'assistant'",
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
		ID:        encrypt.GenerateID(nil),
		ChatId:    chatID,
		Role:      req.Role,
		Content:   req.Content,
		CreatedAt: time.Now(),
		Model:     req.Model,
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

// GetMessagesHandler handles GET /api/chats/{id}/messages
func GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
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

	// Get messages for this chat
	messages, err := store.GetMessagesByChatId(chatID)
	if err != nil {
		slog.Error("Failed to get messages", "chat_id", chatID, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to retrieve messages",
		})
		return
	}

	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Messages retrieved successfully",
		Data:    messages,
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
