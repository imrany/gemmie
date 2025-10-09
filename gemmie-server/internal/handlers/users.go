package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
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
	AgreeToTerms bool		`json:"agree_to_terms"`
	Username string `json:"username"`
}

// RegisterRequest represents registration request payload
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	AgreeToTerms bool		`json:"agree_to_terms"`
}

// SyncRequest represents data sync request - FIXED field naming consistency
type SyncRequest struct {
	Chats           string `json:"chats"`
	LinkPreviews    string `json:"link_previews"`
	CurrentChatID   string `json:"current_chat_id"`
	Preferences     string `json:"preferences,omitempty"`
	WorkFunction    string `json:"work_function,omitempty"`
	Theme           string `json:"theme,omitempty"`
	SyncEnabled     bool   `json:"sync_enabled"`
	Username        string `json:"username"`
	Plan            string `json:"plan,omitempty"`
	PlanName        string `json:"plan_name,omitempty"`
	Amount          int    `json:"amount,omitempty"`
	Duration        string `json:"duration,omitempty"`
	PhoneNumber     string `json:"phone_number,omitempty"`     // Consistent naming
	ExpiryTimestamp int64  `json:"expiry_timestamp,omitempty"`
	ExpireDuration  int64  `json:"expire_duration,omitempty"`
	Price           string `json:"price,omitempty"`
	ExternalReference string `json:"external_reference,omitempty"` // For payment tracking
}

// Update AuthResponse struct
type AuthResponse struct {
	UserID          string    `json:"user_id"`
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	CreatedAt       time.Time `json:"created_at"`
	Chats           string    `json:"chats,omitempty"`
	LinkPreviews    string    `json:"link_previews,omitempty"`
	CurrentChatID   string    `json:"current_chat_id,omitempty"`
	Preferences     string    `json:"preferences,omitempty"`
	WorkFunction    string    `json:"work_function,omitempty"`
	Theme           string    `json:"theme,omitempty"`
	SyncEnabled     bool      `json:"sync_enabled"`
	Plan            string    `json:"plan,omitempty"`
	PlanName        string    `json:"plan_name,omitempty"`
	Amount          int       `json:"amount,omitempty"`
	Duration        string    `json:"duration,omitempty"`
	PhoneNumber     string    `json:"phone_number,omitempty"`
	ExpiryTimestamp int64     `json:"expiry_timestamp,omitempty"`
	ExpireDuration  int64     `json:"expire_duration,omitempty"`
	Price           string    `json:"price,omitempty"`
	EmailVerified   bool      `json:"email_verified"`
	EmailSubscribed bool      `json:"email_subscribed"`
}

type ProfileUpdateRequest struct {
	Username     string `json:"username,omitempty"`
	WorkFunction string `json:"workFunction,omitempty"`
	Preferences  string `json:"preferences,omitempty"`
	Theme        string `json:"theme,omitempty"`
	SyncEnabled  *bool  `json:"sync_enabled,omitempty"` // Pointer to handle explicit false
	PhoneNumber  string `json:"phone_number,omitempty"`  // Added phone number support
}

// Phone number validation function
func validatePhoneNumber(phone string) bool {
	if phone == "" {
		return true // Optional field
	}
	
	// Kenyan phone number regex - supports Safaricom, Airtel, Telkom
	// Formats: +254XXXXXXXXX, 254XXXXXXXXX, 07XXXXXXXX, 01XXXXXXXX
	phoneRegex := regexp.MustCompile(`^(\+254|254|0)(7[0-9]{8}|1[0-9]{8})$`)
	return phoneRegex.MatchString(strings.TrimSpace(phone))
}

// Input sanitization function
func sanitizeString(input string) string {
	return strings.TrimSpace(input)
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

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Sanitize inputs
	req.Username = sanitizeString(req.Username)
	req.Email = sanitizeString(req.Email)

	// Enhanced validation (existing validation code...)
	if req.Username == "" || req.Email == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Username, email, and password are required",
		})
		return
	}

	// Validate email format
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid email format",
		})
		return
	}

	// Validate username
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]{3,30}$`)
	if !usernameRegex.MatchString(req.Username) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Username must be 3-30 characters and contain only letters, numbers, underscores, or hyphens",
		})
		return
	}

	// Validate password strength
	if len(req.Password) < 8 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Password must be at least 8 characters long",
		})
		return
	}

	if !req.AgreeToTerms {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Accept our terms of service and privacy policy",
		})
		return
	}

	// Check if user already exists
	if _, exists := FindUserByEmail(req.Email); exists {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User with this email already exists",
		})
		return
	}

	if _, exists := findUserByUsername(req.Username); exists {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User with this username already exists",
		})
		return
	}

	// Create new user with email fields
	userID := encrypt.GenerateUserID()
	passwordHash := encrypt.HashCredentials(req.Username, req.Email, req.Password)
	unsubscribeToken := encrypt.GenerateUnsubscribeToken(userID)
	
	user := store.User{
		ID:              userID,
		Username:        req.Username,
		Email:           req.Email,
		PasswordHash:    passwordHash,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Preferences:     "",
		WorkFunction:    "",
		Theme:           "system",
		SyncEnabled:     true,
		Plan:            "free",
		PlanName:        "Free",
		Amount:          0,
		Duration:        "",
		PhoneNumber:     "",
		ExpiryTimestamp: 0,
		ExpireDuration:  0,
		Price:           "",
		AgreeToTerms:    req.AgreeToTerms,
		
		EmailVerified:       false,  // Will be verified after clicking link
		EmailSubscribed:     true,   // Default: subscribed to promotional emails
		UnsubscribeToken:    unsubscribeToken,
		VerificationToken:   "",     // Will be set when verification email is sent
		VerificationTokenExpiry: time.Time{},
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

	if err := store.SaveStorage(); err != nil {
		slog.Error("Failed to save storage after registration", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to save user data",
		})
		return
	}

	// Return response with user data (including new fields)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "User registered successfully. Please check your email to verify your account.",
		Data: AuthResponse{
			UserID:        user.ID,
			Username:      user.Username,
			Email:         user.Email,
			CreatedAt:     user.CreatedAt,
			Chats:         userData.Chats,
			LinkPreviews:  userData.LinkPreviews,
			CurrentChatID: userData.CurrentChatID,
			Preferences:   user.Preferences,
			WorkFunction:  user.WorkFunction,
			Theme:         user.Theme,
			SyncEnabled:   user.SyncEnabled,
			PhoneNumber:   user.PhoneNumber,
			Plan:          user.Plan,
			PlanName:      user.PlanName,
			Amount:        user.Amount,
			Duration:      user.Duration,
			Price:         user.Price,
			ExpiryTimestamp: user.ExpiryTimestamp,
			ExpireDuration:  user.ExpireDuration,
		},
	})
}

// loginHandler handles user login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Sanitize inputs
	req.Email = sanitizeString(req.Email)
	req.Username = sanitizeString(req.Username)

	// Validate required fields
	if req.Email == "" || req.Password == "" || req.Username == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Username, email, and password are required",
		})
		return
	}

	if !req.AgreeToTerms{
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Accept our terms of service and privacy policy",
		})
		return
	}

	// Find user by email
	user, exists := FindUserByEmail(req.Email)
	if !exists {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid credentials",
		})
		return
	}

	// Verify credentials hash
	expectedHash := encrypt.HashCredentials(req.Username, req.Email, req.Password)
	if user.PasswordHash != expectedHash {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid credentials",
		})
		return
	}

	// Get user data
	store.Storage.Mu.RLock()
	userData, hasData := store.Storage.UserData[user.ID]
	existingUser :=store.Storage.Users[user.ID]
	store.Storage.Mu.RUnlock()

	if !existingUser.AgreeToTerms {
		store.Storage.Mu.Lock()
		existingUser.AgreeToTerms = req.AgreeToTerms
		store.Storage.Mu.Unlock()
		
		store.SaveStorage()
	}


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

	// Return response with user data - FIXED consistent naming
	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Login successful",
		Data: AuthResponse{
			UserID:          user.ID,
			Username:        user.Username,
			Email:           user.Email,
			CreatedAt:       user.CreatedAt,
			Chats:           userData.Chats,
			LinkPreviews:    userData.LinkPreviews,
			CurrentChatID:   userData.CurrentChatID,
			Preferences:     user.Preferences,
			WorkFunction:    user.WorkFunction,
			Theme:           user.Theme,
			SyncEnabled:     user.SyncEnabled,
			Plan:            user.Plan,
			PlanName:        user.PlanName,
			Amount:          user.Amount,
			Duration:        user.Duration,
			PhoneNumber:     user.PhoneNumber,     // Consistent naming
			ExpiryTimestamp: user.ExpiryTimestamp,
			ExpireDuration:  user.ExpireDuration,
			Price:           user.Price,
		},
	})
}

// syncHandler handles data synchronization
func SyncHandler(w http.ResponseWriter, r *http.Request) {
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
	store.Storage.Mu.RLock()
	_, userExists := store.Storage.Users[userID]
	store.Storage.Mu.RUnlock()

	if !userExists {
		w.WriteHeader(http.StatusNotFound)
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
		user := store.Storage.Users[userID]
		store.Storage.Mu.RUnlock()

		if !exists {
			w.WriteHeader(http.StatusNotFound)
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
				"chats":             userData.Chats,
				"link_previews":     userData.LinkPreviews,
				"current_chat_id":   userData.CurrentChatID,
				"updated_at":        userData.UpdatedAt,
				"preferences":       user.Preferences,
				"work_function":     user.WorkFunction,
				"theme":             user.Theme,
				"sync_enabled":      user.SyncEnabled,
				"username":          user.Username,
				"plan":              user.Plan,
				"plan_name":         user.PlanName,
				"amount":            user.Amount,
				"duration":          user.Duration,
				"phone_number":      user.PhoneNumber,      // Consistent naming
				"expiry_timestamp":  user.ExpiryTimestamp,
				"expire_duration":   user.ExpireDuration,
				"price":             user.Price,
			},
		})

	case "POST":
		// Update user data
		var req SyncRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: "Invalid request body",
			})
			return
		}

		// Sanitize string inputs
		req.Username = sanitizeString(req.Username)
		req.PhoneNumber = sanitizeString(req.PhoneNumber)
		req.Plan = sanitizeString(req.Plan)
		req.PlanName = sanitizeString(req.PlanName)

		// Validate phone number if provided
		if req.PhoneNumber != "" && !validatePhoneNumber(req.PhoneNumber) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: "Invalid phone number format",
			})
			return
		}

		// Update UserData
		userData := store.UserData{
			UserID:        userID,
			Chats:         req.Chats,
			LinkPreviews:  req.LinkPreviews,
			CurrentChatID: req.CurrentChatID,
			UpdatedAt:     time.Now(),
		}

		// Get existing user and update only profile fields
		store.Storage.Mu.Lock()
		
		// Update UserData
		store.Storage.UserData[userID] = userData
		
		
		// Update User profile fields while preserving existing data
		if existingUser, exists := store.Storage.Users[userID]; exists {
			// Only update fields if they're provided in the request
			if req.Username != "" {
				existingUser.Username = req.Username
			}
			if req.Preferences != "" {
				existingUser.Preferences = req.Preferences
			}
			if req.WorkFunction != "" {
				existingUser.WorkFunction = req.WorkFunction
			}
			if req.Theme != "" {
				existingUser.Theme = req.Theme
			}
			if req.SyncEnabled != existingUser.SyncEnabled {
				existingUser.SyncEnabled = req.SyncEnabled
			}
			if req.Amount != 0 {
				existingUser.Amount = req.Amount
			}
			if req.Plan != "" {
				existingUser.Plan = req.Plan
			}
			if req.PlanName != "" {
				existingUser.PlanName = req.PlanName
			}
			if req.Duration != "" {
				existingUser.Duration = req.Duration
			}
			if req.PhoneNumber != "" {
				existingUser.PhoneNumber = req.PhoneNumber // Consistent naming
			}
			if req.ExpiryTimestamp != 0 {
				existingUser.ExpiryTimestamp = req.ExpiryTimestamp
			}
			if req.ExpireDuration != 0 {
				existingUser.ExpireDuration = req.ExpireDuration
			}
			if req.Price != "" {
				existingUser.Price = req.Price
			}
			// Always update the timestamp
			existingUser.UpdatedAt = time.Now()
			
			// Save the updated user back to storage
			store.Storage.Users[userID] = existingUser
		}
		
		store.Storage.Mu.Unlock()

		// Save to persistent storage
		if err := store.SaveStorage(); err != nil {
			slog.Error("Failed to save storage after sync", 
				"user_id", userID, 
				"error", err,
			)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: "Failed to save data",
			})
			return
		}

		json.NewEncoder(w).Encode(store.Response{
			Success: true,
			Message: "Data synchronized successfully",
			Data: map[string]interface{}{
				"updated_at":        userData.UpdatedAt,
			},
		})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Method not allowed",
		})
	}
}

// ProfileHandler handles profile updates
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Only allow PUT method
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Method not allowed",
		})
		return
	}

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
	store.Storage.Mu.RLock()
	existingUser, userExists := store.Storage.Users[userID]
	store.Storage.Mu.RUnlock()

	if !userExists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User not found",
		})
		return
	}

	// Parse request body
	var req ProfileUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Sanitize inputs
	req.Username = sanitizeString(req.Username)
	req.PhoneNumber = sanitizeString(req.PhoneNumber)

	// Validate phone number if provided
	if req.PhoneNumber != "" && !validatePhoneNumber(req.PhoneNumber) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid phone number format. Please use format: +254XXXXXXXXX or 07XXXXXXXX",
		})
		return
	}

	// Update user profile fields
	store.Storage.Mu.Lock()
	updatedUser := existingUser // Copy existing user data
	
	if req.Username != "" {
		// Check if username is already taken by another user
		for otherUserID, otherUser := range store.Storage.Users {
			if otherUserID != userID && otherUser.Username == req.Username {
				store.Storage.Mu.Unlock()
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(store.Response{
					Success: false,
					Message: "Username is already taken",
				})
				return
			}
		}
		updatedUser.Username = req.Username
	}
	
	if req.WorkFunction != "" {
		updatedUser.WorkFunction = req.WorkFunction
	}
	
	if req.Preferences != "" {
		updatedUser.Preferences = req.Preferences
	}
	
	if req.Theme != "" {
		updatedUser.Theme = req.Theme
	}
	
	if req.SyncEnabled != nil {
		updatedUser.SyncEnabled = *req.SyncEnabled
	}

	if req.PhoneNumber != "" {
		updatedUser.PhoneNumber = req.PhoneNumber // Consistent naming
	}
	
	// Always update the timestamp
	updatedUser.UpdatedAt = time.Now()
	
	// Save back to storage
	store.Storage.Users[userID] = updatedUser
	store.Storage.Mu.Unlock()

	// Save to persistent storage
	if err := store.SaveStorage(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to save profile changes",
		})
		return
	}

	// Return success response with consistent naming
	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Profile updated successfully",
		Data: map[string]interface{}{
			"username":      updatedUser.Username,
			"work_function": updatedUser.WorkFunction,
			"preferences":   updatedUser.Preferences,
			"theme":         updatedUser.Theme,
			"sync_enabled":  updatedUser.SyncEnabled,
			"phone_number":  updatedUser.PhoneNumber, // Consistent naming
			"updated_at":    updatedUser.UpdatedAt,
		},
	})
}

// Health check handler
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	store.Storage.Mu.RLock()
	userCount := len(store.Storage.Users)
	transactionCount := len(store.Storage.Transactions)
	store.Storage.Mu.RUnlock()

	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Server is healthy",
		Data: map[string]interface{}{
			"timestamp":         time.Now(),
			"user_count":        userCount,
			"transaction_count": transactionCount,
			"version":           "1.0.0", // Add version info
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

    // Sanitize inputs
    req.Email = sanitizeString(req.Email)
    req.Username = sanitizeString(req.Username)

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
	expectedHash := encrypt.HashCredentials(user.Username, user.Email, req.Password)
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
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(store.Response{
            Success: false,
            Message: "Failed to complete account deletion",
        })
        return
    }

    json.NewEncoder(w).Encode(store.Response{
        Success: true,
        Message: "Account deleted successfully",
    })
}