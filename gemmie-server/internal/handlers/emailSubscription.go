package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/imrany/gemmie/gemmie-server/internal/encrypt"
	"github.com/imrany/gemmie/gemmie-server/internal/mailer"
	"github.com/imrany/gemmie/gemmie-server/store"
)

// UnsubscribeRequest represents unsubscribe request payload
type UnsubscribeRequest struct {
	Email string `json:"email"`
	Token string `json:"token"` // Can be user ID or unsubscribe token
}

// SubscriptionUpdateRequest for managing email subscription preferences
type SubscriptionUpdateRequest struct {
	EmailSubscribed bool `json:"email_subscribed"`
}

// VerifyEmailRequest for email verification
type VerifyEmailRequest struct {
	Token string `json:"token"`
}

// UnsubscribeHandler
func UnsubscribeHandler(w http.ResponseWriter, r *http.Request) {
	// Support both GET (from email link) and POST (from API)
	var req UnsubscribeRequest
	isGetRequest := r.Method == http.MethodGet

	switch r.Method {
	case http.MethodGet:
		req.Email = r.URL.Query().Get("email")
		req.Token = r.URL.Query().Get("token")
	case http.MethodPost:
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: "Invalid request body",
			})
			return
		}
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Method not allowed",
		})
		return
	}

	// Validate inputs
	req.Email = sanitizeString(req.Email)
	req.Token = sanitizeString(req.Token)

	if req.Email == "" || req.Token == "" {
		if isGetRequest {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "%s", getErrorHTML("Missing Information", "Email and token are required"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Email and token are required",
		})
		return
	}

	// Find user by email
	user, exists := FindUserByEmail(req.Email)
	if !exists {
		if isGetRequest {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "%s", getErrorHTML("User Not Found", "We couldn't find an account with that email address"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User not found",
		})
		return
	}

	// Verify token (unsubscribe token)
	if req.Token != user.UnsubscribeToken {
		if isGetRequest {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "%s", getErrorHTML("Invalid Token", "The unsubscribe link is invalid or has expired"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid unsubscribe token",
		})
		return
	}

	// Check if already unsubscribed
	if !user.EmailSubscribed {
		if isGetRequest {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%s", getAlreadyUnsubscribedHTML(user.Email, user.UnsubscribeToken))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(store.Response{
			Success: true,
			Message: "You are already unsubscribed from promotional emails",
		})
		return
	}

	// Update subscription status
	var u store.User
	u.EmailSubscribed = false
	u.EmailVerified = true
	u.UpdatedAt = time.Now()
	if err := store.UpdateUser(u); err != nil {
		slog.Error("Error updating user subscription status", "user_id", user.ID, "error", err)
		if isGetRequest {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "%s", getErrorHTML("Update Failed", "Failed to update subscription status. Please try again later."))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to update subscription status",
		})
		return
	}

	slog.Info("User unsubscribed from promotional emails",
		"user_id", user.ID,
		"email", user.Email,
	)

	// Return success response
	if isGetRequest {
		// For email links, return HTML page
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s", getUnsubscribeSuccessHTML(user.Email, user.UnsubscribeToken))
	} else {
		// For API calls, return JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(store.Response{
			Success: true,
			Message: "Successfully unsubscribed from promotional emails",
		})
	}
}

// Helper functions for HTML responses
func getUnsubscribeSuccessHTML(email, token string) string {
	return `
<!DOCTYPE html>
<html>
<head>
	<title>Unsubscribed</title>
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<style>
		body { font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; text-align: center; background-color: #f4f4f4; }
		.container { background: white; padding: 40px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
		.success { color: #5cb85c; font-size: 48px; }
		h1 { color: #333; margin: 20px 0; }
		p { color: #666; line-height: 1.6; margin: 15px 0; }
		.resubscribe { margin-top: 30px; padding: 20px; background-color: #f8f9fa; border-radius: 5px; }
		.button { display: inline-block; background: #667eea; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; margin-top: 10px; transition: background 0.3s; }
		.button:hover { background: #5568d3; }
	</style>
</head>
<body>
	<div class="container">
		<div class="success">✓</div>
		<h1>Successfully Unsubscribed</h1>
		<p>You have been unsubscribed from Gemmie promotional emails.</p>
		<p>You will no longer receive upgrade notifications and marketing emails.</p>
		<div class="resubscribe">
			<p><strong>Changed your mind?</strong></p>
			<a href="https://gemmie.villebz.com/resubscribe?email=` + email + `&token=` + token + `" class="button">Click here to resubscribe</a>
		</div>
	</div>
</body>
</html>
	`
}

func getAlreadyUnsubscribedHTML(email, token string) string {
	return `
<!DOCTYPE html>
<html>
<head>
	<title>Already Unsubscribed</title>
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<style>
		body { font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; text-align: center; background-color: #f4f4f4; }
		.container { background: white; padding: 40px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
		.info { color: #0275d8; font-size: 48px; }
		h1 { color: #333; margin: 20px 0; }
		p { color: #666; line-height: 1.6; margin: 15px 0; }
		.button { display: inline-block; background: #667eea; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; margin-top: 20px; transition: background 0.3s; }
		.button:hover { background: #5568d3; }
	</style>
</head>
<body>
	<div class="container">
		<div class="info">ℹ</div>
		<h1>Already Unsubscribed</h1>
		<p>You are already unsubscribed from promotional emails.</p>
		<p>Want to receive updates again?</p>
		<a href="https://gemmie.villebz.com/resubscribe?email=` + email + `&token=` + token + `" class="button">Resubscribe</a>
	</div>
</body>
</html>
	`
}

func getErrorHTML(title, message string) string {
	return `
<!DOCTYPE html>
<html>
<head>
	<title>` + title + `</title>
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<style>
		body { font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; text-align: center; background-color: #f4f4f4; }
		.container { background: white; padding: 40px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
		.error { color: #d9534f; font-size: 48px; }
		h1 { color: #333; margin: 20px 0; }
		p { color: #666; line-height: 1.6; margin: 15px 0; }
		.button { display: inline-block; background: #667eea; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; margin-top: 20px; transition: background 0.3s; }
		.button:hover { background: #5568d3; }
	</style>
</head>
<body>
	<div class="container">
		<div class="error">✗</div>
		<h1>` + title + `</h1>
		<p>` + message + `</p>
		<a href="https://gemmie-ai.web.app" class="button">Go to Gemmie</a>
	</div>
</body>
</html>
	`
}

// ResubscribeHandler handles resubscribing users to promotional emails
func ResubscribeHandler(w http.ResponseWriter, r *http.Request) {
	// DON'T set Content-Type here
	
	var req struct {
		Email string `json:"email"`
		Token string `json:"token"`
	}
	isGetRequest := r.Method == http.MethodGet

	switch r.Method {
	case http.MethodPost:
		w.Header().Set("Content-Type", "application/json")
		// Handle JSON body for API calls
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: "Invalid request body",
			})
			return
		}
	case http.MethodGet:
		// Handle query parameters for email links
		req.Email = r.URL.Query().Get("email")
		req.Token = r.URL.Query().Get("token")
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Method not allowed",
		})
		return
	}

	// Validate inputs
	req.Email = sanitizeString(req.Email)
	req.Token = sanitizeString(req.Token)

	if req.Email == "" || req.Token == "" {
		if isGetRequest {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "%s", getErrorHTML("Missing Information", "Email and token are required"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Email and token are required",
		})
		return
	}

	// Find user by email
	var userID string
	var user store.User
	var userFound bool

	users, _ := store.GetUsers()
	for _, u := range users {
		if u.Email == req.Email {
			userID = u.ID
			user = u
			userFound = true
			break
		}
	}

	if !userFound {
		if isGetRequest {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "%s", getErrorHTML("User Not Found", "We couldn't find an account with that email address"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User not found",
		})
		return
	}

	// Verify unsubscribe token matches
	if user.UnsubscribeToken != req.Token {
		if isGetRequest {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "%s", getErrorHTML("Invalid Token", "The resubscribe link is invalid or has expired"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid or expired token",
		})
		return
	}

	// Check if already subscribed
	if user.EmailSubscribed {
		if isGetRequest {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%s", getAlreadySubscribedHTML())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(store.Response{
			Success: true,
			Message: "You are already subscribed to promotional emails",
		})
		return
	}

	// Update subscription status
	var u store.User
	u.EmailSubscribed = true
	u.UpdatedAt = time.Now()
	if err := store.UpdateUser(u); err != nil {
		slog.Error("Failed to save storage after resubscribe",
			"user_id", userID,
			"email", req.Email,
			"error", err,
		)
		if isGetRequest {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "%s", getErrorHTML("Update Failed", "Failed to update subscription status. Please try again later."))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to update subscription status",
		})
		return
	}

	slog.Info("User resubscribed to promotional emails",
		"user_id", userID,
		"email", req.Email,
	)

	// Return appropriate response based on method
	if isGetRequest {
		// For email links, return HTML page
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s", getResubscribeSuccessHTML())
	} else {
		// For API calls, return JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(store.Response{
			Success: true,
			Message: "Successfully resubscribed to promotional emails",
		})
	}
}

// VerifyEmailHandler verifies user's email with token
func VerifyEmailHandler(w http.ResponseWriter, r *http.Request) {
	// DON'T set Content-Type here
	
	// Support both GET (from email link) and POST (from API)
	var token string
	isGetRequest := r.Method == http.MethodGet

	switch r.Method {
	case http.MethodGet:
		token = r.URL.Query().Get("token")
	case http.MethodPost:
		w.Header().Set("Content-Type", "application/json")
		var req VerifyEmailRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: "Invalid request body",
			})
			return
		}
		token = req.Token
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Method not allowed",
		})
		return
	}

	token = sanitizeString(token)
	if token == "" {
		if isGetRequest {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "%s", getErrorHTML("Missing Token", "Verification token is required"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Verification token is required",
		})
		return
	}

	// Find user with this token
	var foundUser *store.User
	var foundUserID string

	users, _ := store.GetUsers()
	for _, user := range users {
		if user.VerificationToken == token {
			userCopy := user
			foundUser = &userCopy
			foundUserID = user.ID
			break
		}
	}

	if foundUser == nil {
		if isGetRequest {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "%s", getErrorHTML("Invalid Token", "The verification link is invalid or has already been used"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid verification token",
		})
		return
	}

	// Check if token expired
	if time.Now().After(foundUser.VerificationTokenExpiry) {
		if isGetRequest {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "%s", getErrorHTML("Token Expired", "This verification link has expired. Please request a new one."))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Verification token has expired",
		})
		return
	}

	// Check if already verified
	if foundUser.EmailVerified {
		if isGetRequest {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%s", getAlreadyVerifiedHTML())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(store.Response{
			Success: true,
			Message: "Email is already verified",
		})
		return
	}

	// Mark email as verified
	var u store.User
	u.EmailVerified = true
	u.VerificationToken = "" // Clear token after use
	u.UpdatedAt = time.Now()
	if err := store.UpdateUser(u); err != nil {
		slog.Error("Failed to save email verification",
			"user_id", foundUserID,
			"error", err,
		)
		if isGetRequest {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "%s", getErrorHTML("Verification Failed", "Failed to verify email. Please try again later."))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to verify email",
		})
		return
	}

	slog.Info("Email verified successfully",
		"user_id", foundUserID,
		"email", foundUser.Email,
	)

	// Return response based on request method
	if isGetRequest {
		// For email links, return HTML page
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s", getVerifySuccessHTML())
	} else {
		// For API calls, return JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(store.Response{
			Success: true,
			Message: "Email verified successfully",
			Data: map[string]interface{}{
				"email_verified": true,
			},
		})
	}
}

// Additional HTML helper functions
func getResubscribeSuccessHTML() string {
	return `
<!DOCTYPE html>
<html>
<head>
    <title>Resubscribed</title>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; text-align: center; background-color: #f4f4f4; }
        .container { background: white; padding: 40px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .success { color: #5cb85c; font-size: 48px; }
        h1 { color: #333; margin: 20px 0; }
        p { color: #666; line-height: 1.6; margin: 15px 0; }
        .button { display: inline-block; background: #667eea; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; margin-top: 20px; transition: background 0.3s; }
        .button:hover { background: #5568d3; }
    </style>
</head>
<body>
    <div class="container">
        <div class="success">✓</div>
        <h1>Successfully Resubscribed</h1>
        <p>You have been resubscribed to Gemmie promotional emails.</p>
        <p>You will now receive upgrade notifications and marketing emails.</p>
        <a href="https://gemmie-ai.web.app" class="button">Return to Gemmie</a>
    </div>
</body>
</html>
	`
}

func getAlreadySubscribedHTML() string {
	return `
<!DOCTYPE html>
<html>
<head>
    <title>Already Subscribed</title>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; text-align: center; background-color: #f4f4f4; }
        .container { background: white; padding: 40px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .info { color: #0275d8; font-size: 48px; }
        h1 { color: #333; margin: 20px 0; }
        p { color: #666; line-height: 1.6; margin: 15px 0; }
        .button { display: inline-block; background: #667eea; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; margin-top: 20px; transition: background 0.3s; }
        .button:hover { background: #5568d3; }
    </style>
</head>
<body>
    <div class="container">
        <div class="info">ℹ</div>
        <h1>Already Subscribed</h1>
        <p>You are already subscribed to promotional emails.</p>
        <p>You'll continue receiving updates from Gemmie.</p>
        <a href="https://gemmie-ai.web.app" class="button">Return to Gemmie</a>
    </div>
</body>
</html>
	`
}

func getVerifySuccessHTML() string {
	return `
<!DOCTYPE html>
<html>
<head>
    <title>Email Verified</title>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; text-align: center; background-color: #f4f4f4; }
        .container { background: white; padding: 40px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .success { color: #5cb85c; font-size: 48px; }
        h1 { color: #333; margin: 20px 0; }
        p { color: #666; line-height: 1.6; margin: 15px 0; }
        .button { display: inline-block; background: #667eea; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; margin-top: 20px; transition: background 0.3s; }
        .button:hover { background: #5568d3; }
    </style>
</head>
<body>
    <div class="container">
        <div class="success">✓</div>
        <h1>Email Verified Successfully!</h1>
        <p>Your email has been verified. You can now enjoy all features of Gemmie.</p>
        <a href="https://gemmie-ai.web.app" class="button">Go to Gemmie</a>
    </div>
</body>
</html>
	`
}

func getAlreadyVerifiedHTML() string {
	return `
<!DOCTYPE html>
<html>
<head>
    <title>Already Verified</title>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; text-align: center; background-color: #f4f4f4; }
        .container { background: white; padding: 40px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .info { color: #0275d8; font-size: 48px; }
        h1 { color: #333; margin: 20px 0; }
        p { color: #666; line-height: 1.6; margin: 15px 0; }
        .button { display: inline-block; background: #667eea; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; margin-top: 20px; transition: background 0.3s; }
        .button:hover { background: #5568d3; }
    </style>
</head>
<body>
    <div class="container">
        <div class="info">ℹ</div>
        <h1>Email Already Verified</h1>
        <p>Your email address is already verified.</p>
        <p>You're all set to use Gemmie!</p>
        <a href="https://gemmie-ai.web.app" class="button">Go to Gemmie</a>
    </div>
</body>
</html>
	`
}

// UpdateEmailSubscriptionHandler updates user's email subscription preference
func UpdateEmailSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPut {
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
	var req SubscriptionUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Verify user exists
	user, _ := store.GetUserByID(userID)

	if user == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User not found",
		})
		return
	}

	// Update subscription status
	var u store.User
	u.EmailSubscribed = req.EmailSubscribed
	u.UpdatedAt = time.Now()
	if err := store.UpdateUser(u); err != nil{
		slog.Error("Failed to save storage after subscription update",
			"user_id", userID,
			"error", err,
		)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to update subscription status",
		})
		return
	}

	action := "unsubscribed from"
	if req.EmailSubscribed {
		action = "subscribed to"
	}

	slog.Info("Email subscription updated",
		"user_id", userID,
		"email", user.Email,
		"subscribed", req.EmailSubscribed,
	)

	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: fmt.Sprintf("Successfully %s promotional emails", action),
		Data: map[string]interface{}{
			"email_subscribed": user.EmailSubscribed,
		},
	})
}

// SendVerificationEmailHandler sends email verification link
func SendVerificationEmailHandler(w http.ResponseWriter, r *http.Request, smtpConfig mailer.SMTPConfig) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
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

	// Verify user exists
	user, _ := store.GetUserByID(userID)

	if user == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User not found",
		})
		return
	}

	// Check if already verified
	if user.EmailVerified {
		json.NewEncoder(w).Encode(store.Response{
			Success: true,
			Message: "Email is already verified",
		})
		return
	}

	// Generate verification token
	token := encrypt.GenerateVerificationToken(user.Email, userID)
	expiry := time.Now().Add(24 * time.Hour) // Token valid for 24 hours

	// Update user with token
	var u store.User
	u.VerificationToken = token
	u.VerificationTokenExpiry = expiry
	if err := store.UpdateUser(u); err != nil{
		slog.Error("Failed to save verification token",
			"user_id", userID,
			"error", err,
		)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to generate verification token",
		})
		return
	}


	// Send verification email
	verifyURL := fmt.Sprintf("https://gemmie-ai.web.app/verify-email?token=%s", token)
	emailBody := buildVerificationEmailBody(user.Username, verifyURL)

	emailData := mailer.EmailData{
		To:      []string{user.Email},
		Subject: "Verify Your Email - Gemmie",
		Body:    emailBody,
		IsHTML:  true,
	}

	if err := mailer.SendEmail(emailData, smtpConfig); err != nil {
		slog.Error("Failed to send verification email",
			"user_id", userID,
			"error", err,
		)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to send verification email",
		})
		return
	}

	slog.Info("Verification email sent",
		"user_id", userID,
		"email", user.Email,
	)

	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Verification email sent successfully",
	})
}

func buildVerificationEmailBody(username, verifyURL string) string {
	return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Verify Your Email - Gemmie</title>
    <style>
        body { 
            font-family: Arial, sans-serif; 
            line-height: 1.6; 
            color: #333; 
            margin: 0;
            padding: 0;
            background-color: #f4f4f4;
        }
        .container { 
            max-width: 600px; 
            margin: 0 auto; 
            padding: 20px; 
            background-color: #ffffff;
        }
        .header { 
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); 
            color: white; 
            padding: 30px; 
            text-align: center; 
            border-radius: 10px 10px 0 0; 
        }
        .header h1 {
            margin: 0;
            font-size: 24px;
        }
        .content { 
            background: #ffffff; 
            padding: 30px; 
            border: 1px solid #e0e0e0; 
            border-top: none;
        }
        .verify-button { 
            display: inline-block; 
            background: #667eea; 
            color: white !important; 
            padding: 15px 30px; 
            text-decoration: none; 
            border-radius: 5px; 
            margin: 20px 0; 
            font-weight: bold;
            text-align: center;
        }
        .verify-button:hover {
            background: #5568d3;
        }
        .footer { 
            text-align: center; 
            padding: 20px; 
            color: #666; 
            font-size: 12px; 
            background-color: #f8f9fa;
            border-radius: 0 0 10px 10px;
        }
        .link-text { 
            word-break: break-all; 
            color: #666; 
            font-size: 12px; 
            margin-top: 20px;
            padding: 15px;
            background-color: #f8f9fa;
            border-radius: 5px;
        }
        .warning {
            color: #999; 
            font-size: 14px;
            margin-top: 30px;
            padding: 15px;
            background-color: #fff3cd;
            border-left: 4px solid #ffc107;
            border-radius: 4px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎉 Welcome to Gemmie, ` + username + `!</h1>
            <p style="margin: 10px 0 0 0;">Just one more step to get started</p>
        </div>
        
        <div class="content">
            <h2 style="color: #333; margin-top: 0;">Verify Your Email Address</h2>
            <p>Thanks for signing up! Please verify your email address to activate your account and access all features.</p>
            
            <center>
                <a href="` + verifyURL + `" class="verify-button">Verify Email Address →</a>
            </center>
            
            <p style="margin-top: 30px; text-align: center;">
                <strong>⏱️ This verification link will expire in 24 hours.</strong>
            </p>
            
            <div class="link-text">
                <strong>Button not working?</strong><br>
                Copy and paste this link into your browser:<br>
                <span style="color: #667eea;">` + verifyURL + `</span>
            </div>
            
            <div class="warning">
                <strong>⚠️ Didn't create a Gemmie account?</strong><br>
                If you didn't sign up for Gemmie, you can safely ignore this email.
            </div>
        </div>
        
        <div class="footer">
            <p style="margin: 5px 0;"><strong>Thanks for joining Gemmie!</strong></p>
            <p style="margin: 5px 0;">Questions? Reply to this email or visit our support center.</p>
            <p style="margin: 15px 0 5px 0; font-size: 11px; color: #999;">
                © 2025 Gemmie. All rights reserved.
            </p>
        </div>
    </div>
</body>
</html>
`
}
