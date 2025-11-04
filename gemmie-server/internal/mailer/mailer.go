package mailer

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"time"
)

// OtpPurpose represents the purpose of an OTP
type OtpPurpose string

const (
	OtpPurposeLogin         OtpPurpose = "login"
	OtpPurposePasswordReset OtpPurpose = "password_reset"
	OtpPurposeVerification  OtpPurpose = "verification"
	OtpPurposeRegistration  OtpPurpose = "registration"
)

// SMTPConfig holds the SMTP server configuration
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Email    string
}

// EmailData represents an email message
type EmailData struct {
	To      []string
	Subject string
	Body    string
	IsHTML  bool
}

// OTPData represents OTP information
type OTPData struct {
	Code      string
	ExpiresAt time.Time
	Purpose   OtpPurpose
}

// ValidateConfig checks if the SMTP configuration is valid
func ValidateConfig(config SMTPConfig) error {
	if config.Host == "" {
		return fmt.Errorf("SMTP host is required")
	}
	if config.Port == 0 {
		return fmt.Errorf("SMTP port is required")
	}
	if config.Username == "" {
		return fmt.Errorf("SMTP username is required")
	}
	if config.Password == "" {
		return fmt.Errorf("SMTP password is required")
	}
	if config.Email == "" {
		return fmt.Errorf("SMTP email is required")
	}
	return nil
}

// SendEmail sends a generic email
func SendEmail(emailData EmailData, config SMTPConfig) error {
	if err := ValidateConfig(config); err != nil {
		return fmt.Errorf("SMTP configuration error: %w", err)
	}

	// Validate email data
	if len(emailData.To) == 0 {
		return fmt.Errorf("recipient email address is required")
	}
	if emailData.Subject == "" {
		return fmt.Errorf("email subject is required")
	}
	if emailData.Body == "" {
		return fmt.Errorf("email body is required")
	}

	// Create authentication
	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)

	message := buildMessage(emailData, config)

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	return sendWithTLS(addr, auth, config.Email, emailData.To, []byte(message), config)
}

// SendOTP generates and sends an OTP via email
func SendOTP(email string, purpose OtpPurpose, otp string, config SMTPConfig) (string, error) {
	if email == "" {
		return "", fmt.Errorf("email address is required")
	}

	subject := getOTPSubject(purpose)
	body := getOTPBody(otp, purpose)

	emailData := EmailData{
		To:      []string{email},
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	}

	err := SendEmail(emailData, config)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	log.Printf("OTP sent successfully to %s for purpose: %s", email, purpose)
	return otp, nil
}

// SendOTPWithCustomTemplate sends OTP with custom email template
func SendOTPWithCustomTemplate(email string, purpose OtpPurpose, otp string, subject, htmlTemplate, textTemplate string, config SMTPConfig) (string, error) {
	if email == "" {
		return "", fmt.Errorf("email address is required")
	}

	// Replace placeholders in templates
	htmlBody := strings.ReplaceAll(htmlTemplate, "{{OTP}}", otp)
	htmlBody = strings.ReplaceAll(htmlBody, "{{PURPOSE}}", string(purpose))

	textBody := strings.ReplaceAll(textTemplate, "{{OTP}}", otp)
	textBody = strings.ReplaceAll(textBody, "{{PURPOSE}}", string(purpose))

	// Create multipart message
	body := buildMultipartMessage(htmlBody, textBody)

	emailData := EmailData{
		To:      []string{email},
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	}

	err := SendEmail(emailData, config)
	if err != nil {
		return "", fmt.Errorf("failed to send OTP email: %w", err)
	}

	log.Printf("OTP with custom template sent successfully to %s for purpose: %s", email, purpose)
	return otp, nil
}

// SendOTPWithExistingCode sends an existing OTP via email (for resend functionality)
func SendOTPWithExistingCode(email string, purpose OtpPurpose, otp string, config SMTPConfig) error {
	if email == "" {
		return fmt.Errorf("email address is required")
	}
	if otp == "" {
		return fmt.Errorf("OTP is required")
	}

	subject := getOTPSubject(purpose)
	body := getOTPBody(otp, purpose)

	emailData := EmailData{
		To:      []string{email},
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	}

	err := SendEmail(emailData, config)
	if err != nil {
		return fmt.Errorf("failed to resend OTP email: %w", err)
	}

	log.Printf("OTP resent successfully to %s for purpose: %s", email, purpose)
	return nil
}

// Helper functions
func buildMessage(emailData EmailData, config SMTPConfig) string {
	var message strings.Builder

	// Headers
	message.WriteString(fmt.Sprintf("From: %s\r\n", config.Email))
	message.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(emailData.To, ", ")))
	message.WriteString(fmt.Sprintf("Subject: %s\r\n", emailData.Subject))

	if emailData.IsHTML {
		message.WriteString("MIME-Version: 1.0\r\n")
		message.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	} else {
		message.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	}

	message.WriteString("\r\n")
	message.WriteString(emailData.Body)

	return message.String()
}

func buildMultipartMessage(htmlBody, textBody string) string {
	boundary := fmt.Sprintf("boundary_%d", time.Now().Unix())

	var message strings.Builder
	message.WriteString("MIME-Version: 1.0\r\n")
	message.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\r\n", boundary))
	message.WriteString("\r\n")

	// Text part
	message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	message.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	message.WriteString("\r\n")
	message.WriteString(textBody)
	message.WriteString("\r\n")

	// HTML part
	message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	message.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	message.WriteString("\r\n")
	message.WriteString(htmlBody)
	message.WriteString("\r\n")

	message.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	return message.String()
}

func sendWithTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte, config SMTPConfig) error {
	// Connect to server
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			log.Printf("Error closing SMTP client: %v", closeErr)
		}
	}()

	// Start TLS
	tlsConfig := &tls.Config{
		ServerName:         config.Host,
		InsecureSkipVerify: false, // Set to true only for testing
	}

	if err = client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("failed to start TLS: %w", err)
	}

	// Authenticate
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	// Set sender
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipients
	for _, recipient := range to {
		if err = client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
		}
	}

	// Send message
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	_, err = writer.Write(msg)
	if err != nil {
		writer.Close() // Try to close on error
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return client.Quit()
}

// getOTPSubject creates OTP Email Subject based on purpose
func getOTPSubject(purpose OtpPurpose) string {
	switch purpose {
	case OtpPurposeLogin:
		return "Your Login Verification Code"
	case OtpPurposePasswordReset:
		return "Password Reset Verification Code"
	case OtpPurposeVerification:
		return "Account Verification Code"
	case OtpPurposeRegistration:
		return "Registration Verification Code"
	default:
		return "Your Verification Code"
	}
}

// getOTPBody creates OTP Email body based on purpose
func getOTPBody(otp string, purpose OtpPurpose) string {
	switch purpose {
	case OtpPurposeLogin:
		return fmt.Sprintf(`
			<html>
			<head>
				<meta charset="UTF-8">
				<title>Login Verification</title>
			</head>
			<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
				<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
					<h2 style="color: #2c5aa0;">Login Verification</h2>
					<p>Your login verification code is:</p>
					<div style="background-color: #f4f4f4; padding: 20px; text-align: center; font-size: 24px; font-weight: bold; letter-spacing: 3px; margin: 20px 0;">
						%s
					</div>
					<p><strong>This code will expire in 10 minutes.</strong></p>
					<p style="color: #666;">If you didn't request this code, please ignore this email and ensure your account is secure.</p>
					<hr style="border: 1px solid #eee; margin: 20px 0;">
					<p style="font-size: 12px; color: #999;">This is an automated message, please do not reply to this email.</p>
				</div>
			</body>
			</html>
		`, otp)
	case OtpPurposePasswordReset:
		return fmt.Sprintf(`
			<html>
			<head>
				<meta charset="UTF-8">
				<title>Password Reset</title>
			</head>
			<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
				<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
					<h2 style="color: #d9534f;">Password Reset</h2>
					<p>Your password reset verification code is:</p>
					<div style="background-color: #f4f4f4; padding: 20px; text-align: center; font-size: 24px; font-weight: bold; letter-spacing: 3px; margin: 20px 0;">
						%s
					</div>
					<p><strong>This code will expire in 15 minutes.</strong></p>
					<p style="color: #666;">If you didn't request a password reset, please ignore this email and ensure your account is secure.</p>
					<hr style="border: 1px solid #eee; margin: 20px 0;">
					<p style="font-size: 12px; color: #999;">This is an automated message, please do not reply to this email.</p>
				</div>
			</body>
			</html>
		`, otp)
	case OtpPurposeRegistration:
		return fmt.Sprintf(`
			<html>
			<head>
				<meta charset="UTF-8">
				<title>Registration Verification</title>
			</head>
			<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
				<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
					<h2 style="color: #5cb85c;">Welcome! Complete Your Registration</h2>
					<p>Thank you for registering! Please use the following verification code to complete your account setup:</p>
					<div style="background-color: #f4f4f4; padding: 20px; text-align: center; font-size: 24px; font-weight: bold; letter-spacing: 3px; margin: 20px 0;">
						%s
					</div>
					<p><strong>This code will expire in 10 minutes.</strong></p>
					<p style="color: #666;">If you didn't create an account, please ignore this email.</p>
					<hr style="border: 1px solid #eee; margin: 20px 0;">
					<p style="font-size: 12px; color: #999;">This is an automated message, please do not reply to this email.</p>
				</div>
			</body>
			</html>
		`, otp)
	default:
		return fmt.Sprintf(`
			<html>
			<head>
				<meta charset="UTF-8">
				<title>Verification Code</title>
			</head>
			<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
				<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
					<h2 style="color: #337ab7;">Verification Code</h2>
					<p>Your verification code is:</p>
					<div style="background-color: #f4f4f4; padding: 20px; text-align: center; font-size: 24px; font-weight: bold; letter-spacing: 3px; margin: 20px 0;">
						%s
					</div>
					<p><strong>This code will expire in 10 minutes.</strong></p>
					<hr style="border: 1px solid #eee; margin: 20px 0;">
					<p style="font-size: 12px; color: #999;">This is an automated message, please do not reply to this email.</p>
				</div>
			</body>
			</html>
		`, otp)
	}
}

// GetOTPExpirationDuration returns the appropriate expiration duration for each purpose
func GetOTPExpirationDuration(purpose OtpPurpose) time.Duration {
	switch purpose {
	case OtpPurposePasswordReset:
		return 15 * time.Minute
	case OtpPurposeLogin, OtpPurposeVerification, OtpPurposeRegistration:
		return 10 * time.Minute
	default:
		return 10 * time.Minute
	}
}

// CreateOTPData creates an OTPData struct with appropriate expiration
func CreateOTPData(code string, purpose OtpPurpose) OTPData {
	return OTPData{
		Code:      code,
		ExpiresAt: time.Now().Add(GetOTPExpirationDuration(purpose)),
		Purpose:   purpose,
	}
}

// IsOTPExpired checks if an OTP has expired
func (o *OTPData) IsExpired() bool {
	return time.Now().After(o.ExpiresAt)
}

// IsValidForPurpose checks if an OTP is valid for a specific purpose
func (o *OTPData) IsValidForPurpose(purpose OtpPurpose) bool {
	return o.Purpose == purpose && !o.IsExpired()
}
