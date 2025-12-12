package handlers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/gorilla/mux"
	"github.com/imrany/gemmie/gemmie-server/internal/encrypt"
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
	// Check if streaming is supported, but don't fail if not
	flusher, canStream := w.(http.Flusher)

	if canStream {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Transfer-Encoding", "chunked")
		w.Header().Set("X-Content-Type-Options", "nosniff")
	} else {
		w.Header().Set("Content-Type", "application/json")
		slog.Warn("Response writer doesn't support streaming, using non-streaming mode")
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		w.Header().Set("Content-Type", "application/json")
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Chat ID is required",
		})
		return
	}

	var req store.Message
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}
	if req.Prompt == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Prompt is required",
		})
		return
	}

	chat, err := store.GetChatById(chatID)
	if err != nil || chat == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Chat not found",
		})
		return
	}
	if chat.UserId != userID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Access denied",
		})
		return
	}

	if req.Response == "" {
		// Create HTTP client with timeout
		client := &http.Client{
			Timeout: 120 * time.Second,
		}

		// Prepare Ollama request
		ollamaReq := map[string]any{
			"model":  "llama3.2:3b",
			"prompt": req.Prompt,
			"stream": canStream, // Only stream if we can flush
		}
		body, _ := json.Marshal(ollamaReq)

		slog.Info("Calling Ollama", "stream", canStream)

		ollamaResp, err := client.Post("https://wrapper.triple-ts-mediclinic.com/api/generate",
			"application/json", bytes.NewReader(body))
		if err != nil {
			slog.Error("Failed to contact Ollama", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: fmt.Sprintf("Failed to contact Ollama: %v", err),
			})
			return
		}
		defer ollamaResp.Body.Close()

		if ollamaResp.StatusCode != http.StatusOK {
			slog.Error("Ollama returned error", "status", ollamaResp.StatusCode)
			bodyBytes, _ := io.ReadAll(ollamaResp.Body)
			slog.Error("Ollama error response", "body", string(bodyBytes))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: fmt.Sprintf("Ollama returned status: %d", ollamaResp.StatusCode),
			})
			return
		}

		var finalResponse strings.Builder

		if canStream {
			// STREAMING MODE - send chunks as they arrive
			scanner := bufio.NewScanner(ollamaResp.Body)

			for scanner.Scan() {
				line := scanner.Bytes()

				var chunk struct {
					Response string `json:"response"`
					Done     bool   `json:"done"`
				}
				if err := json.Unmarshal(line, &chunk); err == nil {
					finalResponse.WriteString(chunk.Response)

					// Stream to client
					w.Write(line)
					w.Write([]byte("\n"))
					flusher.Flush()

					if chunk.Done {
						break
					}
				}
			}

			if err := scanner.Err(); err != nil {
				slog.Error("Streaming error", "error", err)
				errorChunk := map[string]interface{}{
					"error": err.Error(),
					"done":  true,
				}
				errorJSON, _ := json.Marshal(errorChunk)
				w.Write(errorJSON)
				w.Write([]byte("\n"))
				flusher.Flush()
				return
			}
		} else {
			// NON-STREAMING MODE - wait for complete response
			bodyBytes, err := io.ReadAll(ollamaResp.Body)
			if err != nil {
				slog.Error("Failed to read Ollama response", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(store.Response{
					Success: false,
					Message: "Failed to read AI response",
				})
				return
			}

			// Parse the complete response
			var ollamaData struct {
				Response string `json:"response"`
			}
			if err := json.Unmarshal(bodyBytes, &ollamaData); err != nil {
				slog.Error("Failed to parse Ollama response", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(store.Response{
					Success: false,
					Message: "Failed to parse AI response",
				})
				return
			}

			finalResponse.WriteString(ollamaData.Response)
		}

		req.Response = finalResponse.String()
	}

	// Save final message to DB
	message := store.Message{
		ID:         encrypt.GenerateID(nil),
		ChatId:     chatID,
		Prompt:     req.Prompt,
		Response:   req.Response,
		CreatedAt:  time.Now(),
		Model:      "llama3.2:3b",
		References: req.References,
	}

	if err := store.CreateMessage(message); err != nil {
		slog.Error("Failed to create message", "error", err)
		if !canStream {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: "Failed to save message",
			})
		}
		return
	}

	chat.MessageCount++
	chat.LastMessageAt = time.Now()
	chat.UpdatedAt = time.Now()
	if err := store.UpdateChat(*chat); err != nil {
		slog.Error("Failed to update chat", "chat_id", chatID, "error", err)
	}

	slog.Info("Message created successfully", "message_id", message.ID, "chat_id", chatID, "user_id", userID)

	// If not streaming, send final JSON response
	if !canStream {
		json.NewEncoder(w).Encode(store.Response{
			Success: true,
			Message: "Message created",
			Data:    message,
		})
	}

	// Send push notification asynchronously
	go func() {
		notifCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		subscriptions, err := store.GetSubscriptionsByUserID(notifCtx, userID)
		if err != nil {
			slog.Error("Failed to get subscriptions", "user_id", userID, "error", err)
			return
		}

		if len(subscriptions) == 0 {
			slog.Debug("No subscriptions found for user", "user_id", userID)
			return
		}

		promptPreview := message.Prompt
		if len(promptPreview) > 20 {
			promptPreview = promptPreview[:20] + "..."
		} else if len(promptPreview) == 0 {
			promptPreview = "your request"
		}

		payload := store.NotificationPayload{
			Title: "✅ Gemmie Finished Your Task!",
			Body:  fmt.Sprintf("The response for '%s' is ready — tap to review it now.", promptPreview),
			Data: map[string]any{
				"chat_id":    message.ChatId,
				"message_id": message.ID,
				"url":        "/chat/" + message.ChatId,
			},
			Tag:                "response-complete",
			RequireInteraction: true,
		}

		data, err := json.Marshal(payload)
		if err != nil {
			slog.Error("Failed to marshal notification payload", "error", err)
			return
		}

		successCount := 0
		failureCount := 0

		for _, sub := range subscriptions {
			if sub.Endpoint == "" || sub.P256dhKey == "" || sub.AuthKey == "" {
				slog.Warn("Invalid subscription data", "user_id", sub.UserID)
				store.DeleteSubscription(notifCtx, sub.Endpoint)
				failureCount++
				continue
			}

			resp, err := webpush.SendNotification(data, &webpush.Subscription{
				Endpoint: sub.Endpoint,
				Keys: webpush.Keys{
					Auth:   sub.AuthKey,
					P256dh: sub.P256dhKey,
				},
			}, &webpush.Options{
				Subscriber:      viper.GetString("VAPID_EMAIL"),
				VAPIDPublicKey:  viper.GetString("VAPID_PUBLIC_KEY"),
				VAPIDPrivateKey: viper.GetString("VAPID_PRIVATE_KEY"),
				TTL:             30,
			})
			if err != nil {
				slog.Error("Failed to send push notification",
					"user_id", sub.UserID,
					"error", err.Error(),
				)

				if resp != nil && (resp.StatusCode == 410 || resp.StatusCode == 404) {
					slog.Info("Deleting invalid subscription", "user_id", sub.UserID, "status", resp.StatusCode)
					store.DeleteSubscription(notifCtx, sub.Endpoint)
				} else if err.Error() == "P256 point not on curve" {
					slog.Warn("Deleting subscription with mismatched VAPID keys", "user_id", sub.UserID)
					store.DeleteSubscription(notifCtx, sub.Endpoint)
				}

				failureCount++
				continue
			}

			if resp != nil && resp.Body != nil {
				resp.Body.Close()
			}

			if resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
				successCount++
				slog.Debug("Push notification sent successfully",
					"user_id", sub.UserID,
					"status", resp.StatusCode,
				)
			} else {
				failureCount++
				if resp != nil {
					slog.Warn("Push notification failed",
						"user_id", sub.UserID,
						"status", resp.StatusCode,
					)
				}
			}
		}

		if successCount > 0 || failureCount > 0 {
			slog.Info("Push notification summary",
				"user_id", userID,
				"success", successCount,
				"failed", failureCount,
				"total", len(subscriptions),
			)
		}
	}()
}

// func CreateMessageHandler(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")
//
// 	userID := r.Header.Get("X-User-ID")
// 	if userID == "" {
// 		w.WriteHeader(http.StatusUnauthorized)
// 		json.NewEncoder(w).Encode(store.Response{
// 			Success: false,
// 			Message: "User ID header required",
// 		})
// 		return
// 	}
//
// 	vars := mux.Vars(r)
// 	chatID := vars["id"]
// 	if chatID == "" {
// 		w.WriteHeader(http.StatusBadRequest)
// 		json.NewEncoder(w).Encode(store.Response{
// 			Success: false,
// 			Message: "Chat ID is required",
// 		})
// 		return
// 	}
//
// 	// Parse request body
// 	var req store.Message
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		slog.Error("Invalid request body", "error", err)
// 		w.WriteHeader(http.StatusBadRequest)
// 		json.NewEncoder(w).Encode(store.Response{
// 			Success: false,
// 			Message: "Invalid request body",
// 		})
// 		return
// 	}
//
// 	// Validate required fields
// 	if req.Prompt == "" {
// 		w.WriteHeader(http.StatusBadRequest)
// 		json.NewEncoder(w).Encode(store.Response{
// 			Success: false,
// 			Message: "Prompt is required",
// 		})
// 		return
// 	}
//
// 	// Get chat to verify ownership first (before generating AI response)
// 	chat, err := store.GetChatById(chatID)
// 	if err != nil {
// 		slog.Error("Failed to get chat", "chat_id", chatID, "error", err)
// 		w.WriteHeader(http.StatusInternalServerError)
// 		json.NewEncoder(w).Encode(store.Response{
// 			Success: false,
// 			Message: "Failed to retrieve chat",
// 		})
// 		return
// 	}
//
// 	if chat == nil {
// 		w.WriteHeader(http.StatusNotFound)
// 		json.NewEncoder(w).Encode(store.Response{
// 			Success: false,
// 			Message: "Chat not found",
// 		})
// 		return
// 	}
//
// 	// Verify chat belongs to user
// 	if chat.UserId != userID {
// 		w.WriteHeader(http.StatusForbidden)
// 		json.NewEncoder(w).Encode(store.Response{
// 			Success: false,
// 			Message: "Access denied",
// 		})
// 		return
// 	}
//
// 	// Generate response using AI wrapper (only if response not provided)
// 	genaiResp := genai.GENAISERVICE{
// 		APIKey: viper.GetString("API_KEY"),
// 		Model:  viper.GetString("MODEL"),
// 	}
//
// 	var aiResponse string
// 	if req.Response == "" {
// 		genAIResponse, err := genaiResp.GenerateAIResponse(context.Background(), req.Prompt)
// 		if err != nil {
// 			slog.Error("Failed to generate AI response", "error", err)
// 			w.WriteHeader(http.StatusInternalServerError)
// 			json.NewEncoder(w).Encode(store.Response{
// 				Success: false,
// 				Message: fmt.Sprintf("Failed to generate AI response: %s", err.Error()),
// 			})
// 			return
// 		}
// 		aiResponse = genAIResponse.Response
// 	} else {
// 		aiResponse = req.Response
// 	}
//
// 	// Create new message
// 	message := store.Message{
// 		ID:         encrypt.GenerateID(nil),
// 		ChatId:     chatID,
// 		Prompt:     req.Prompt,
// 		Response:   aiResponse,
// 		CreatedAt:  time.Now(),
// 		Model:      genaiResp.Model,
// 		References: req.References,
// 	}
//
// 	if err := store.CreateMessage(message); err != nil {
// 		slog.Error("Failed to create message", "error", err)
// 		w.WriteHeader(http.StatusInternalServerError)
// 		json.NewEncoder(w).Encode(store.Response{
// 			Success: false,
// 			Message: "Failed to create message",
// 		})
// 		return
// 	}
//
// 	// Update chat's message count and last message time
// 	chat.MessageCount++
// 	chat.LastMessageAt = time.Now()
// 	chat.UpdatedAt = time.Now()
//
// 	if err := store.UpdateChat(*chat); err != nil {
// 		slog.Error("Failed to update chat", "chat_id", chatID, "error", err)
// 	}
//
// 	slog.Info("Message created successfully", "message_id", message.ID, "chat_id", chatID, "user_id", userID)
//
// 	w.WriteHeader(http.StatusCreated)
// 	json.NewEncoder(w).Encode(store.Response{
// 		Success: true,
// 		Message: "Message created successfully",
// 		Data:    message,
// 	})
//
// 	// Send push notification asynchronously to avoid blocking the response
// 	go func() {
// 		// Use a new context with timeout for the notification
// 		notifCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 		defer cancel()
//
// 		// Get subscriptions
// 		subscriptions, err := store.GetSubscriptionsByUserID(notifCtx, userID)
// 		if err != nil {
// 			slog.Error("Failed to get subscriptions", "user_id", userID, "error", err)
// 			return
// 		}
//
// 		if len(subscriptions) == 0 {
// 			slog.Debug("No subscriptions found for user", "user_id", userID)
// 			return
// 		}
//
// 		// Prepare notification payload
// 		promptPreview := message.Prompt
// 		if len(promptPreview) > 20 {
// 			promptPreview = promptPreview[:20] + "..."
// 		} else if len(promptPreview) == 0 {
// 			promptPreview = "your request"
// 		}
//
// 		payload := store.NotificationPayload{
// 			Title: "✅ Gemmie Finished Your Task!",
// 			Body:  fmt.Sprintf("The response for '%s' is ready — tap to review it now.", promptPreview),
// 			Data: map[string]any{
// 				"chat_id":    message.ChatId,
// 				"message_id": message.ID,
// 				"url":        "/chat/" + message.ChatId,
// 			},
// 			Tag:                "response-complete",
// 			RequireInteraction: true,
// 		}
//
// 		data, err := json.Marshal(payload)
// 		if err != nil {
// 			slog.Error("Failed to marshal notification payload", "error", err)
// 			return
// 		}
//
// 		// Send to all subscriptions
// 		successCount := 0
// 		failureCount := 0
//
// 		for _, sub := range subscriptions {
// 			// Validate subscription before sending
// 			if sub.Endpoint == "" || sub.P256dhKey == "" || sub.AuthKey == "" {
// 				slog.Warn("Invalid subscription data", "user_id", sub.UserID)
// 				// Delete invalid subscription
// 				store.DeleteSubscription(notifCtx, sub.Endpoint)
// 				failureCount++
// 				continue
// 			}
//
// 			resp, err := webpush.SendNotification(data, &webpush.Subscription{
// 				Endpoint: sub.Endpoint,
// 				Keys: webpush.Keys{
// 					Auth:   sub.AuthKey,
// 					P256dh: sub.P256dhKey,
// 				},
// 			}, &webpush.Options{
// 				Subscriber:      viper.GetString("VAPID_EMAIL"),
// 				VAPIDPublicKey:  viper.GetString("VAPID_PUBLIC_KEY"),
// 				VAPIDPrivateKey: viper.GetString("VAPID_PRIVATE_KEY"),
// 				TTL:             30,
// 			})
// 			if err != nil {
// 				slog.Error("Failed to send push notification",
// 					"user_id", sub.UserID,
// 					"error", err.Error(),
// 				)
//
// 				// Delete subscription if it's a key mismatch or invalid endpoint
// 				if resp != nil && (resp.StatusCode == 410 || resp.StatusCode == 404) {
// 					slog.Info("Deleting invalid subscription", "user_id", sub.UserID, "status", resp.StatusCode)
// 					store.DeleteSubscription(notifCtx, sub.Endpoint)
// 				} else if err.Error() == "P256 point not on curve" {
// 					// This means the subscription was created with different VAPID keys
// 					slog.Warn("Deleting subscription with mismatched VAPID keys", "user_id", sub.UserID)
// 					store.DeleteSubscription(notifCtx, sub.Endpoint)
// 				}
//
// 				failureCount++
// 				continue
// 			}
//
// 			// Close response body
// 			if resp != nil && resp.Body != nil {
// 				resp.Body.Close()
// 			}
//
// 			// Check response status
// 			if resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
// 				successCount++
// 				slog.Debug("Push notification sent successfully",
// 					"user_id", sub.UserID,
// 					"status", resp.StatusCode,
// 				)
// 			} else {
// 				failureCount++
// 				if resp != nil {
// 					slog.Warn("Push notification failed",
// 						"user_id", sub.UserID,
// 						"status", resp.StatusCode,
// 					)
// 				}
// 			}
// 		}
//
// 		if successCount > 0 || failureCount > 0 {
// 			slog.Info("Push notification summary",
// 				"user_id", userID,
// 				"success", successCount,
// 				"failed", failureCount,
// 				"total", len(subscriptions),
// 			)
// 		}
// 	}()
// }

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
