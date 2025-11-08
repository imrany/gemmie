package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/imrany/gemmie/gemmie-server/pkg/mailer"
)

// SendEmailRequest represents the request body for sending emails
type SendEmailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	HTML    string `json:"html"`
	IsHTML  bool   `json:"is_html"`
}

// SendEmailResponse represents the response for email sending
type SendEmailResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// SendEmailHandler handles sending emails via SMTP
func SendEmailHandler(w http.ResponseWriter, r *http.Request, smtpConfig mailer.SMTPConfig) {
	w.Header().Set("Content-Type", "application/json")

	// Only allow POST requests
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(SendEmailResponse{
			Success: false,
			Message: "Method not allowed",
		})
		return
	}

	// Parse request body
	var req SendEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode email request", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SendEmailResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Validate required fields
	if req.To == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SendEmailResponse{
			Success: false,
			Message: "Recipient email (to) is required",
		})
		return
	}

	if req.Subject == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SendEmailResponse{
			Success: false,
			Message: "Email subject is required",
		})
		return
	}

	if req.HTML == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SendEmailResponse{
			Success: false,
			Message: "Email HTML content is required",
		})
		return
	}

	// Check if SMTP is configured
	if smtpConfig.Host == "" || smtpConfig.Username == "" || smtpConfig.Password == "" {
		slog.Error("SMTP not configured")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(SendEmailResponse{
			Success: false,
			Message: "Email service is not configured",
		})
		return
	}

	emailData := mailer.EmailData{
		To:      []string{req.To},
		Subject: req.Subject,
		Body:    req.HTML,
		IsHTML:  req.IsHTML,
	}
	// Send email using mailer package
	err := mailer.SendEmail(
		emailData,
		smtpConfig,
	)

	if err != nil {
		slog.Error("Failed to send email",
			"to", req.To,
			"subject", req.Subject,
			"error", err,
		)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(SendEmailResponse{
			Success: false,
			Message: "Failed to send email",
		})
		return
	}

	slog.Info("Email sent successfully",
		"to", req.To,
		"subject", req.Subject,
	)

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SendEmailResponse{
		Success: true,
		Message: "Email sent successfully",
	})
}
