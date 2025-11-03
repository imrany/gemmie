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
	Email        string `json:"email"`
	Password     string `json:"password"`
	AgreeToTerms bool   `json:"agree_to_terms"`
	Username     string `json:"username"`
	UserAgent    string `json:"user_agent"`
}

// RegisterRequest represents registration request payload
type RegisterRequest struct {
	Username     string `json:"username"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	AgreeToTerms bool   `json:"agree_to_terms"`
	UserAgent    string `json:"user_agent"`
}

// SyncRequest represents data sync request
type SyncRequest struct {
	Preferences       string             `json:"preferences,omitempty"`
	WorkFunction      string             `json:"work_function,omitempty"`
	Theme             string             `json:"theme,omitempty"`
	SyncEnabled       bool               `json:"sync_enabled"`
	Username          string             `json:"username"`
	Plan              string             `json:"plan,omitempty"`
	PlanName          string             `json:"plan_name,omitempty"`
	ResponseMode      store.Modes        `json:"response_mode,omitempty"`
	Amount            int                `json:"amount,omitempty"`
	Duration          string             `json:"duration,omitempty"`
	PhoneNumber       string             `json:"phone_number,omitempty"` // Consistent naming
	ExpiryTimestamp   int64              `json:"expiry_timestamp,omitempty"`
	ExpireDuration    int64              `json:"expire_duration,omitempty"`
	Price             string             `json:"price,omitempty"`
	ExternalReference string             `json:"external_reference,omitempty"` // For payment tracking
	EmailVerified     bool               `json:"email_verified"`               // Whether email is verified
	EmailSubscribed   bool               `json:"email_subscribed"`             // Whether user subscribed to promotional emails
	RequestCount      store.RequestCount `json:"request_count"`
	Chats             []store.Chat       `json:"chats,omitempty"`
}

// Update AuthResponse struct
type AuthResponse struct {
	UserID          string             `json:"user_id"`
	Username        string             `json:"username"`
	Email           string             `json:"email"`
	CreatedAt       time.Time          `json:"created_at"`
	Chats           []store.Chat       `json:"chats,omitempty"`
	Preferences     string             `json:"preferences,omitempty"`
	WorkFunction    string             `json:"work_function,omitempty"`
	Theme           string             `json:"theme,omitempty"`
	SyncEnabled     bool               `json:"sync_enabled"`
	Plan            string             `json:"plan,omitempty"`
	PlanName        string             `json:"plan_name,omitempty"`
	Amount          int                `json:"amount,omitempty"`
	ResponseMode    store.Modes        `json:"response_mode,omitempty"`
	Duration        string             `json:"duration,omitempty"`
	PhoneNumber     string             `json:"phone_number,omitempty"`
	ExpiryTimestamp int64              `json:"expiry_timestamp,omitempty"`
	ExpireDuration  int64              `json:"expire_duration,omitempty"`
	Price           string             `json:"price,omitempty"`
	EmailVerified   bool               `json:"email_verified"`
	EmailSubscribed bool               `json:"email_subscribed"`
	RequestCount    store.RequestCount `json:"request_count"`
}

type ProfileUpdateRequest struct {
	Username     string      `json:"username,omitempty"`
	WorkFunction string      `json:"workFunction,omitempty"`
	Preferences  string      `json:"preferences,omitempty"`
	Theme        string      `json:"theme,omitempty"`
	SyncEnabled  *bool       `json:"sync_enabled,omitempty"` // Pointer to handle explicit false
	ResponseMode store.Modes `json:"response_mode,omitempty"`
	PhoneNumber  string      `json:"phone_number,omitempty"` // Added phone number support
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
	user, err := store.GetUserByEmail(email)
	if err != nil {
		slog.Error("Error finding user by email", "email", email, "error", err)
		return nil, false
	}
	if user == nil {
		return nil, false
	}
	return user, true
}

// findUserByUsername finds a user by username
func findUserByUsername(username string) (*store.User, bool) {
	user, err := store.GetUserByUsername(username)
	if err != nil {
		slog.Error("Error finding user by username", "username", username, "error", err)
		return nil, false
	}
	if user == nil {
		return nil, false
	}
	return user, true
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

	// Validation
	if req.Username == "" || req.Email == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Username, email, and password are required",
		})
		return
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid email format",
		})
		return
	}

	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]{3,30}$`)
	if !usernameRegex.MatchString(req.Username) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Username must be 3-30 characters and contain only letters, numbers, underscores, or hyphens",
		})
		return
	}

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

	// Create new user
	userID := encrypt.GenerateUserID()
	passwordHash := encrypt.HashCredentials(req.Username, req.Email, req.Password)
	unsubscribeToken := encrypt.GenerateUnsubscribeToken(userID)

	user := store.User{
		ID:               userID,
		Username:         req.Username,
		Email:            req.Email,
		PasswordHash:     passwordHash,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		Theme:            "system",
		SyncEnabled:      true,
		Plan:             "free",
		PlanName:         "Free",
		ResponseMode:     "light-response",
		AgreeToTerms:     req.AgreeToTerms,
		EmailVerified:    false,
		EmailSubscribed:  true,
		UnsubscribeToken: unsubscribeToken,
		RequestCount: store.RequestCount{
			Count:     0,
			Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
		},
	}

	// Save to database
	if err := store.CreateUser(user); err != nil {
		slog.Error("Failed to create user", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to create user",
		})
		return
	}

	// Return response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "User registered successfully. Please check your email to verify your account.",
		Data: AuthResponse{
			UserID:          user.ID,
			Username:        user.Username,
			Email:           user.Email,
			CreatedAt:       user.CreatedAt,
			Preferences:     user.Preferences,
			WorkFunction:    user.WorkFunction,
			Theme:           user.Theme,
			SyncEnabled:     user.SyncEnabled,
			PhoneNumber:     user.PhoneNumber,
			Plan:            user.Plan,
			PlanName:        user.PlanName,
			Amount:          user.Amount,
			Duration:        user.Duration,
			Price:           user.Price,
			ResponseMode:    user.ResponseMode,
			ExpiryTimestamp: user.ExpiryTimestamp,
			ExpireDuration:  user.ExpireDuration,
			EmailVerified:   user.EmailVerified,
			EmailSubscribed: user.EmailSubscribed,
			RequestCount:    user.RequestCount,
		},
	})
}

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

	if !req.AgreeToTerms {
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

	// Verify credentials
	expectedHash := encrypt.HashCredentials(req.Username, req.Email, req.Password)
	if user.PasswordHash != expectedHash {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid credentials",
		})
		return
	}

	//  Get user chats from database
	userChats, err := store.GetChatsByUserId(user.ID)
	if err != nil {
		slog.Error("Failed to get user chats", "user_id", user.ID, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to retrieve user chats",
		})
		return
	}

	// Update terms acceptance if needed
	if user.AgreeToTerms != req.AgreeToTerms {
		user.AgreeToTerms = req.AgreeToTerms
		user.UpdatedAt = time.Now()

		if err := store.UpdateUser(*user); err != nil {
			slog.Error("Failed to update terms acceptance", "user_id", user.ID, "error", err)
		}
	}

	// update user agent if need
	if user.UserAgent != req.UserAgent {
		user.UserAgent = req.UserAgent
		user.UpdatedAt = time.Now()

		if err := store.UpdateUser(*user); err != nil {
			slog.Error("Failed to update user agent", "user_id", user.ID, "error", err)
		}
	}

	// Return response
	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Login successful",
		Data: AuthResponse{
			UserID:          user.ID,
			Username:        user.Username,
			Email:           user.Email,
			CreatedAt:       user.CreatedAt,
			Chats:           userChats,
			Preferences:     user.Preferences,
			WorkFunction:    user.WorkFunction,
			Theme:           user.Theme,
			SyncEnabled:     user.SyncEnabled,
			Plan:            user.Plan,
			PlanName:        user.PlanName,
			Amount:          user.Amount,
			Duration:        user.Duration,
			PhoneNumber:     user.PhoneNumber,
			ExpiryTimestamp: user.ExpiryTimestamp,
			ExpireDuration:  user.ExpireDuration,
			Price:           user.Price,
			ResponseMode:    user.ResponseMode,
			EmailVerified:   user.EmailVerified,
			EmailSubscribed: user.EmailSubscribed,
			RequestCount:    user.RequestCount,
		},
	})
}

// SyncHandler
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

	switch r.Method {
	case "GET":
		userTranx, err := store.GetUserTransactions(user.PhoneNumber)
		if err != nil {
			slog.Error("Failed to get user transactions", "user_id", user.ID, "error", err)
		}

		userChats, err := store.GetChatsByUserId(user.ID)
		if err != nil {
			slog.Error("Failed to get user chats", "user_id", user.ID, "error", err)
		}

		json.NewEncoder(w).Encode(store.Response{
			Success: true,
			Message: "Data retrieved successfully",
			Data: map[string]any{
				"preferences":       user.Preferences,
				"work_function":     user.WorkFunction,
				"theme":             user.Theme,
				"sync_enabled":      user.SyncEnabled,
				"username":          user.Username,
				"plan":              user.Plan,
				"plan_name":         user.PlanName,
				"amount":            user.Amount,
				"duration":          user.Duration,
				"phone_number":      user.PhoneNumber,
				"expiry_timestamp":  user.ExpiryTimestamp,
				"expire_duration":   user.ExpireDuration,
				"price":             user.Price,
				"response_mode":     user.ResponseMode,
				"email_verified":    user.EmailVerified,
				"email_subscribed":  user.EmailSubscribed,
				"request_count":     user.RequestCount,
				"user_transactions": userTranx,
				"chats":             userChats,
			},
		})

	case "POST":
		var req SyncRequest
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
		req.Plan = sanitizeString(req.Plan)
		req.PlanName = sanitizeString(req.PlanName)

		// Validate phone number
		if req.PhoneNumber != "" && !validatePhoneNumber(req.PhoneNumber) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: "Invalid phone number format",
			})
			return
		}

		// Update user data in database
		user.Preferences = req.Preferences
		user.WorkFunction = req.WorkFunction
		user.Theme = req.Theme
		user.SyncEnabled = req.SyncEnabled
		user.Username = req.Username
		user.Plan = req.Plan
		user.PlanName = req.PlanName
		user.ResponseMode = req.ResponseMode
		user.Amount = req.Amount
		user.Duration = req.Duration
		user.PhoneNumber = req.PhoneNumber
		user.ExpiryTimestamp = req.ExpiryTimestamp
		user.ExpireDuration = req.ExpireDuration
		user.Price = req.Price
		user.EmailVerified = req.EmailVerified
		user.EmailSubscribed = req.EmailSubscribed
		user.RequestCount = req.RequestCount

		if err := store.UpdateUser(*user); err != nil {
			slog.Error("Failed to update user data", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: "Failed to save data",
			})
			return
		}

		// Update user profile fields
		if req.Username != "" {
			user.Username = req.Username
		}
		if req.Preferences != "" {
			user.Preferences = req.Preferences
		}
		if req.WorkFunction != "" {
			user.WorkFunction = req.WorkFunction
		}
		if req.Theme != "" {
			user.Theme = req.Theme
		}

		if user.SyncEnabled != req.SyncEnabled {
			user.SyncEnabled = req.SyncEnabled
		}

		if req.Amount != 0 {
			user.Amount = req.Amount
		}
		if req.Plan != "" {
			user.Plan = req.Plan
		}
		if req.PlanName != "" {
			user.PlanName = req.PlanName
		}
		if req.Duration != "" {
			user.Duration = req.Duration
		}
		if req.PhoneNumber != "" {
			user.PhoneNumber = req.PhoneNumber
		}
		if req.ExpiryTimestamp != 0 {
			user.ExpiryTimestamp = req.ExpiryTimestamp
		}
		if req.ExpireDuration != 0 {
			user.ExpireDuration = req.ExpireDuration
		}
		if req.Price != "" {
			user.Price = req.Price
		}
		if req.ResponseMode != "" {
			user.ResponseMode = req.ResponseMode
		}

		if req.RequestCount.Count != 0 || req.RequestCount.Timestamp != 0 {
			incomingTime := time.Unix(0, req.RequestCount.Timestamp*int64(time.Millisecond))
			existingTime := time.Unix(0, user.RequestCount.Timestamp*int64(time.Millisecond))

			if incomingTime.After(existingTime) || req.RequestCount.Count != user.RequestCount.Count {
				user.RequestCount = req.RequestCount
			}
		}

		user.UpdatedAt = time.Now()

		if err := store.UpdateUser(*user); err != nil {
			slog.Error("Failed to update user", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: "Failed to save user data",
			})
			return
		}

		json.NewEncoder(w).Encode(store.Response{
			Success: true,
			Message: "Data synchronized successfully",
			Data: map[string]any{
				"updated_at": user.UpdatedAt,
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

// ProfileHandler
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

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

	// Get user from database
	existingUser, err := store.GetUserByID(userID)
	if err != nil {
		slog.Error("Failed to get user", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to retrieve user",
		})
		return
	}

	if existingUser == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User not found",
		})
		return
	}

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

	// Validate phone number
	if req.PhoneNumber != "" && !validatePhoneNumber(req.PhoneNumber) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid phone number format. Please use format: +254XXXXXXXXX or 07XXXXXXXX",
		})
		return
	}

	// Check username uniqueness if changing
	if req.Username != "" && req.Username != existingUser.Username {
		if otherUser, _ := findUserByUsername(req.Username); otherUser != nil {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: "Username is already taken",
			})
			return
		}
		existingUser.Username = req.Username
	}

	// Update fields
	if req.WorkFunction != "" {
		existingUser.WorkFunction = req.WorkFunction
	}
	if req.Preferences != "" {
		existingUser.Preferences = req.Preferences
	}
	if req.Theme != "" {
		existingUser.Theme = req.Theme
	}
	if req.SyncEnabled != nil {
		existingUser.SyncEnabled = *req.SyncEnabled
	}
	if req.PhoneNumber != "" {
		existingUser.PhoneNumber = req.PhoneNumber
	}
	if req.ResponseMode != "" {
		existingUser.ResponseMode = req.ResponseMode
	}

	existingUser.UpdatedAt = time.Now()

	// Save to database
	if err := store.UpdateUser(*existingUser); err != nil {
		slog.Error("Failed to update user", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to save profile changes",
		})
		return
	}

	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Profile updated successfully",
		Data: map[string]interface{}{
			"username":      existingUser.Username,
			"work_function": existingUser.WorkFunction,
			"preferences":   existingUser.Preferences,
			"theme":         existingUser.Theme,
			"sync_enabled":  existingUser.SyncEnabled,
			"phone_number":  existingUser.PhoneNumber,
			"updated_at":    existingUser.UpdatedAt,
		},
	})
}

// HealthHandler
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	version := store.GetVersion()

	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Server is healthy",
		Data: map[string]interface{}{
			"timestamp": time.Now(),
			"version":   version,
			"database":  "connected",
		},
	})
}

// DeleteAccountHandler
func DeleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodDelete {
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

	var req DeleteAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	req.Email = sanitizeString(req.Email)
	req.Username = sanitizeString(req.Username)

	if req.Email == "" || req.Username == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Username, email, and password are required",
		})
		return
	}

	// Get user from database
	user, err := store.GetUserByID(userID)
	if err != nil {
		slog.Error("Failed to get user", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to retrieve user",
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

	// Verify credentials
	if user.Email != req.Email || user.Username != req.Username {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid credentials",
		})
		return
	}

	expectedHash := encrypt.HashCredentials(user.Username, user.Email, req.Password)
	if user.PasswordHash != expectedHash {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid credentials",
		})
		return
	}

	slog.Info("Account deletion requested",
		"user_id", userID,
		"username", user.Username,
		"email", user.Email,
		"timestamp", time.Now(),
	)

	// Delete from database (CASCADE will delete user_data)
	if err := store.DeleteUser(userID); err != nil {
		slog.Error("Failed to delete user", "user_id", userID, "error", err)
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
