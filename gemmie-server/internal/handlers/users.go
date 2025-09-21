package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/imrany/gemmie/gemmie-server/internal/encrypt"
	"github.com/imrany/gemmie/gemmie-server/store"
)

// DeleteAccountRequest represents request payload for account deletion
type DeleteAccountRequest struct {
    Email    string `json:"email"`
    Username string `json:"username"`
    Password string `json:"password"`
}

// LoginRequest represents login request payload
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}

// RegisterRequest represents registration request payload
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SyncRequest represents data sync request
type SyncRequest struct {
	Chats         string `json:"chats"`
	LinkPreviews  string `json:"link_previews"`
	CurrentChatID string `json:"current_chat_id"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	UserID        string    `json:"user_id"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	CreatedAt     time.Time `json:"created_at"`
	Chats         string    `json:"chats,omitempty"`
	LinkPreviews  string    `json:"link_previews,omitempty"`
	CurrentChatID string    `json:"current_chat_id,omitempty"`
}

// findUserByEmail finds a user by email
func FindUserByEmail(email string) (*store.User, bool) {
	store.Storage.Mu.RLock()
	defer store.Storage.Mu.RUnlock()

	for _, user := range store.Storage.Users {
		if user.Email == email {
			return &user, true
		}
	}
	return nil, false
}

// findUserByUsername finds a user by username
func findUserByUsername(username string) (*store.User, bool) {
	store.Storage.Mu.RLock()
	defer store.Storage.Mu.RUnlock()

	for _, user := range store.Storage.Users {
		if user.Username == username {
			return &user, true
		}
	}
	return nil, false
}

// registerHandler handles user registration
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Validate required fields
	if req.Username == "" || req.Email == "" || req.Password == "" {
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Username, email, and password are required",
		})
		return
	}

	// Check if user already exists
	if _, exists := FindUserByEmail(req.Email); exists {
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User with this email already exists",
		})
		return
	}

	if _, exists := findUserByUsername(req.Username); exists {
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User with this username already exists",
		})
		return
	}

	// Create new user
	userID := encrypt.GenerateUserID()
	passwordHash := encrypt.HashCredentials(req.Username, req.Email, req.Password)
	
	user := store.User{
		ID:           userID,
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Create empty user data record
	userData := store.UserData{
		UserID:        userID,
		Chats:         "[]",
		LinkPreviews:  "{}",
		CurrentChatID: "",
		UpdatedAt:     time.Now(),
	}

	// Store in memory and save to file
	store.Storage.Mu.Lock()
	store.Storage.Users[userID] = user
	store.Storage.UserData[userID] = userData
	store.Storage.Mu.Unlock()

	store.SaveStorage()

	// Return response with user data
	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "User registered successfully",
		Data: AuthResponse{
			UserID:        user.ID,
			Username:      user.Username,
			Email:         user.Email,
			CreatedAt:     user.CreatedAt,
			Chats:         userData.Chats,
			LinkPreviews:  userData.LinkPreviews,
			CurrentChatID: userData.CurrentChatID,
		},
	})
}

// loginHandler handles user login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" || req.Username == "" {
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Username, email, and password are required",
		})
		return
	}

	// Find user by email
	user, exists := FindUserByEmail(req.Email)
	if !exists {
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid credentials",
		})
		return
	}

	// Verify credentials hash
	expectedHash := encrypt.HashCredentials(req.Username, req.Email, req.Password)
	if user.PasswordHash != expectedHash {
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid credentials",
		})
		return
	}

	// Get user data
	store.Storage.Mu.RLock()
	userData, hasData := store.Storage.UserData[user.ID]
	store.Storage.Mu.RUnlock()

	if !hasData {
		// Create empty user data if doesn't exist
		userData = store.UserData{
			UserID:        user.ID,
			Chats:         "[]",
			LinkPreviews:  "{}",
			CurrentChatID: "",
			UpdatedAt:     time.Now(),
		}
		
		store.Storage.Mu.Lock()
		store.Storage.UserData[user.ID] = userData
		store.Storage.Mu.Unlock()
		
		store.SaveStorage()
	}

	// Return response with user data
	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Login successful",
		Data: AuthResponse{
			UserID:        user.ID,
			Username:      user.Username,
			Email:         user.Email,
			CreatedAt:     user.CreatedAt,
			Chats:         userData.Chats,
			LinkPreviews:  userData.LinkPreviews,
			CurrentChatID: userData.CurrentChatID,
		},
	})
}

// syncHandler handles data synchronization
func SyncHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User ID header required",
		})
		return
	}

	// Verify user exists
	store.Storage.Mu.RLock()
	_, userExists := store.Storage.Users[userID]
	store.Storage.Mu.RUnlock()

	if !userExists {
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User not found",
		})
		return
	}

	switch r.Method {
	case "GET":
		// Get user data
		store.Storage.Mu.RLock()
		userData, exists := store.Storage.UserData[userID]
		store.Storage.Mu.RUnlock()

		if !exists {
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: "User data not found",
			})
			return
		}

		json.NewEncoder(w).Encode(store.Response{
			Success: true,
			Message: "Data retrieved successfully",
			Data: map[string]interface{}{
				"chats":           userData.Chats,
				"link_previews":   userData.LinkPreviews,
				"current_chat_id": userData.CurrentChatID,
				"updated_at":      userData.UpdatedAt,
			},
		})

	case "POST":
		// Update user data
		var req SyncRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: "Invalid request body",
			})
			return
		}

		userData := store.UserData{
			UserID:        userID,
			Chats:         req.Chats,
			LinkPreviews:  req.LinkPreviews,
			CurrentChatID: req.CurrentChatID,
			UpdatedAt:     time.Now(),
		}

		store.Storage.Mu.Lock()
		store.Storage.UserData[userID] = userData
		store.Storage.Mu.Unlock()

		store.SaveStorage()

		json.NewEncoder(w).Encode(store.Response{
			Success: true,
			Message: "Data synchronized successfully",
			Data: map[string]interface{}{
				"updated_at": userData.UpdatedAt,
			},
		})

	default:
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Method not allowed",
		})
	}
}

// Health check handler
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	store.Storage.Mu.RLock()
	userCount := len(store.Storage.Users)
	transactionCount := len(store.Storage.Transactions)
	orderCount := len(store.Storage.Orders)
	store.Storage.Mu.RUnlock()

	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Server is healthy",
		Data: map[string]interface{}{
			"timestamp":         time.Now(),
			"user_count":        userCount,
			"transaction_count": transactionCount,
			"order_count":       orderCount,
		},
	})
}

// DeleteAccountHandler handles account deletion securely
func DeleteAccountHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    // Only allow DELETE method
    if r.Method != http.MethodDelete {
        w.WriteHeader(http.StatusMethodNotAllowed)
        json.NewEncoder(w).Encode(store.Response{
            Success: false,
            Message: "Method not allowed",
        })
        return
    }

    // Get user ID from header
    userID := r.Header.Get("X-User-ID")
    if userID == "" {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(store.Response{
            Success: false,
            Message: "User ID header required",
        })
        return
    }

    // Parse request body
    var req DeleteAccountRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(store.Response{
            Success: false,
            Message: "Invalid request body",
        })
        return
    }

    // Validate input
    if req.Email == "" || req.Username == "" || req.Password == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(store.Response{
            Success: false,
            Message: "Username, email, and password are required",
        })
        return
    }

    // Verify user exists and userID matches
    store.Storage.Mu.RLock()
    user, userExists := store.Storage.Users[userID]
    store.Storage.Mu.RUnlock()

    if !userExists {
        w.WriteHeader(http.StatusNotFound)
        json.NewEncoder(w).Encode(store.Response{
            Success: false,
            Message: "User not found",
        })
        return
    }

    // Verify the email matches the user ID (additional security check)
    if user.Email != req.Email {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(store.Response{
            Success: false,
            Message: "Invalid credentials",
        })
        return
    }

    // Verify username matches
    if user.Username != req.Username {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(store.Response{
            Success: false,
            Message: "Invalid credentials",
        })
        return
    }

    // Verify password
    expectedHash := encrypt.HashCredentials(req.Username, req.Email, req.Password)
    if user.PasswordHash != expectedHash {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(store.Response{
            Success: false,
            Message: "Invalid credentials",
        })
        return
    }

    // Log the deletion attempt (for audit purposes)
    slog.Info("Account deletion requested", 
        "user_id", userID,
        "username", user.Username,
        "email", user.Email,
        "timestamp", time.Now(),
    )

    // Delete user and associated data
    store.Storage.Mu.Lock()
    delete(store.Storage.Users, userID)
    delete(store.Storage.UserData, userID)
    store.Storage.Mu.Unlock()

    // Save changes to persistent storage
    if err := store.SaveStorage(); err != nil {
        slog.Error("Failed to save storage after account deletion", 
            "user_id", userID, 
            "error", err,
        )
        // Note: At this point the user is deleted from memory but not from disk
        // You might want to handle this differently based on your requirements
    }

    json.NewEncoder(w).Encode(store.Response{
        Success: true,
        Message: "Account deleted successfully",
    })
}
