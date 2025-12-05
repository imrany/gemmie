package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/imrany/gemmie/gemmie-server/store"
	"github.com/imrany/whats-email/pkg/mailer"
)

// EmailSchedulerConfig holds the configuration for email scheduling
type EmailSchedulerConfig struct {
	SMTPConfig      mailer.SMTPConfig
	SendInterval    time.Duration // How often to send emails (e.g., 7 days)
	EnableScheduler bool          // Master switch to enable/disable scheduler
}

// StartEmailScheduler starts the background scheduler for sending upgrade emails
func StartEmailScheduler(config EmailSchedulerConfig) {
	if !config.EnableScheduler {
		slog.Info("Email scheduler is disabled")
		return
	}

	slog.Info("Starting email scheduler",
		"interval", config.SendInterval.String(),
	)

	// Run immediately on startup
	go sendUpgradeEmails(config.SMTPConfig)

	// Schedule periodic sends
	ticker := time.NewTicker(config.SendInterval)
	go func() {
		for range ticker.C {
			sendUpgradeEmails(config.SMTPConfig)
		}
	}()
}

// sendUpgradeEmails sends upgrade emails to eligible users
func sendUpgradeEmails(smtpConfig mailer.SMTPConfig) {
	ctx := context.Background()
	slog.Info("Starting upgrade email batch send", "timestamp", time.Now())

	users, err := store.GetUsers()
	if err != nil {
		slog.Error("Error getting users", "error", err)
	}

	sentCount := 0
	failedCount := 0
	skippedCount := 0

	for _, user := range users {
		// Check if user is eligible for upgrade email
		if !isEligibleForUpgradeEmail(user) {
			skippedCount++
			continue
		}

		// Send email
		if err := sendUpgradeEmail(user, smtpConfig); err != nil {
			slog.Error("Failed to send upgrade email",
				"user_id", user.ID,
				"email", user.Email,
				"error", err,
			)
			failedCount++
		} else {
			slog.Info("Upgrade email sent successfully",
				"user_id", user.ID,
				"email", user.Email,
				"plan", user.Plan,
			)
			sentCount++
		}

		// Get subscriptions
		subscriptions, err := store.GetSubscriptionsByUserIDs(ctx, []string{user.ID})
		if err != nil {
			slog.Error("Failed to get subscriptions", "Error", err)
		}

		if len(subscriptions) == 0 {
			slog.Info("No subscriptions found for user", "user_id", user.ID)
		}

		sub := subscriptions[0]
		payload := store.NotificationPayload{
			Title:              "Unlock Premium Features",
			Body:               "Upgrade your Gemmie Plan and enjoy exclusive features! 🚀,\n click to upgrade your account",
			Data:               map[string]any{"user_id": user.ID, "email": user.Email, "plan": user.Plan, "url": "/upgrade"},
			Tag:                "upgrade-email",
			RequireInteraction: true,
		}

		data, err := json.Marshal(payload)
		if err != nil {
			slog.Error("Failed to marshal notification payload", "Error", err)
		}

		resp, err := webpush.SendNotification(data, &webpush.Subscription{
			Endpoint: sub.Endpoint,
			Keys: webpush.Keys{
				Auth:   sub.AuthKey,
				P256dh: sub.P256dhKey,
			},
		}, &webpush.Options{
			Subscriber:      VapidEmail,
			VAPIDPublicKey:  VapidPublicKey,
			VAPIDPrivateKey: VapidPrivateKey,
			TTL:             30,
		})

		if err != nil {
			slog.Error("Failed to send notification", "Endpoint", sub.Endpoint, "Error", err)

			// Delete invalid subscriptions (410 Gone or 404 Not Found)
			if resp != nil && (resp.StatusCode == 410 || resp.StatusCode == 404) {
				store.DeleteSubscription(ctx, sub.Endpoint)
				slog.Info("Deleted invalid subscription", "Endpoint", sub.Endpoint)
			}
		} else {
			resp.Body.Close()
			if resp.StatusCode <= 200 && resp.StatusCode > 300 {
				slog.Info("Push failed", "Status", resp.StatusCode, "Endpoint", sub.Endpoint)
			}
		}

		// Add small delay between emails to avoid rate limiting
		time.Sleep(100 * time.Millisecond)
	}

	slog.Info("Upgrade email batch completed",
		"sent", sentCount,
		"failed", failedCount,
		"skipped", skippedCount,
		"total", len(users),
	)
}

// isEligibleForUpgradeEmail checks if a user should receive an upgrade email
func isEligibleForUpgradeEmail(user store.User) bool {
	// Check if user is subscribed to emails
	if !user.EmailSubscribed {
		return false
	}

	// TODO: In future, also check email verification
	// if !user.EmailVerified {
	//     return false
	// }

	now := time.Now().Unix()

	// Free plan users are always eligible
	if user.Plan == "free" || user.Plan == "" {
		return true
	}

	// Users with expired plans are eligible
	if user.ExpiryTimestamp > 0 && user.ExpiryTimestamp < now {
		return true
	}

	// Users with plans expiring soon (within 7 days) are eligible
	sevenDaysFromNow := now + (7 * 24 * 60 * 60)
	if user.ExpiryTimestamp > 0 && user.ExpiryTimestamp < sevenDaysFromNow {
		return true
	}

	return false
}

// sendUpgradeEmail sends an upgrade email to a specific user
func sendUpgradeEmail(user store.User, smtpConfig mailer.SMTPConfig) error {
	var subject = "Unlock Premium Features with Gemmie Plans! 🚀"

	// Determine email content based on user status
	var body string
	if user.Plan == "free" || user.Plan == "" {
		body = buildFreeUserEmailBody(user)
	} else {
		subject = "Upgrade your Gemmie " + user.PlanName
		body = buildExpiredUserEmailBody(user)
	}

	emailData := mailer.EmailData{
		To:      []string{user.Email},
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	}

	return mailer.SendEmail(emailData, smtpConfig)
}

// buildFreeUserEmailBody creates HTML email body for free plan users
func buildFreeUserEmailBody(user store.User) string {
	return `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #ffffff; padding: 30px; border: 1px solid #e0e0e0; }
        .feature { margin: 15px 0; padding: 15px; background: #f8f9fa; border-radius: 5px; }
        .feature-icon { font-size: 24px; margin-right: 10px; }
        .cta-button { display: inline-block; background: #667eea; color: #ffff; padding: 15px 30px; text-decoration: none; border-radius: 5px; margin: 20px 0; font-weight: bold; }
        .footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>✨ Hey ` + user.Username + `!</h1>
            <p>Ready to supercharge your Gemmie experience?</p>
        </div>

        <div class="content">
            <h2>Upgrade your Gemmie plan Today!</h2>
            <p>You've been enjoying Gemmie's free features, but there's so much more waiting for you:</p>

            <div class="feature">
                <span class="feature-icon">⚡</span>
                <strong>Unlimited Conversations</strong> - No limits on your AI interactions
            </div>

            <div class="feature">
                <span class="feature-icon">🔄</span>
                <strong>Advanced Sync</strong> - Seamless sync across all your devices
            </div>

            <div class="feature">
                <span class="feature-icon">🎨</span>
                <strong>Premium Themes</strong> - Customize your experience
            </div>

            <div class="feature">
                <span class="feature-icon">🚀</span>
                <strong>Priority Support</strong> - Get help when you need it
            </div>

            <div class="feature">
                <span class="feature-icon">💾</span>
                <strong>Increased Storage</strong> - Store more conversations and data
            </div>

            <center>
                <a href="https://gemmie-ai.web.app/upgrade" class="cta-button">Upgrade Now →</a>
            </center>

            <p style="margin-top: 30px;">
                <strong>Special Offer:</strong> Upgrade this week and get 20% off your first month!
            </p>
        </div>

        <div class="footer">
            <p>Thanks for being part of the Gemmie community!</p>
            <p>Questions? Reply to this email or visit our support center.</p>
            <p><a href="https://gemmie.villebiz.com/unsubscribe?email=` + user.Email + `&token=` + user.UnsubscribeToken + `">Unsubscribe from promotional emails</a></p>
        </div>
    </div>
</body>
</html>
`
}

// buildExpiredUserEmailBody creates HTML email body for expired/expiring plan users
func buildExpiredUserEmailBody(user store.User) string {
	now := time.Now().Unix()
	isExpired := user.ExpiryTimestamp < now

	statusMessage := "Your Gemmie " + user.PlanName + " subscription expired"
	if !isExpired {
		daysRemaining := (user.ExpiryTimestamp - now) / (24 * 60 * 60)
		statusMessage = "Your Gemmie " + user.PlanName + " subscription expires in " + string(rune(daysRemaining)) + " days"
	}

	return `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #ffffff; padding: 30px; border: 1px solid #e0e0e0; }
        .alert-box { background: #fff3cd; border: 1px solid #ffc107; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .cta-button { display: inline-block; background: #f5576c; color: #ffff; padding: 15px 30px; text-decoration: none; border-radius: 5px; margin: 20px 0; font-weight: bold; }
        .footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>⏰ Don't Miss Out, ` + user.Username + `!</h1>
            <p>` + statusMessage + `</p>
        </div>

        <div class="content">
            <div class="alert-box">
                <strong>⚠️ Action Required:</strong> Renew your subscription to keep enjoying premium features!
            </div>

            <h2>What You'll Lose:</h2>
            <ul>
                <li>Unlimited AI conversations</li>
                <li>Advanced sync capabilities</li>
                <li>Premium themes and customization</li>
                <li>Priority customer support</li>
                <li>Extended storage capacity</li>
            </ul>

            <p>Renew now and pick up right where you left off!</p>

            <center>
                <a href="https://gemmie-ai.web.app/upgrade/` + user.Plan + `" class="cta-button">Renew Subscription →</a>
            </center>

            <p style="margin-top: 30px;">
                <strong>Loyal Customer Bonus:</strong> Renew within 7 days and get 15% off!
            </p>
        </div>

        <div class="footer">
            <p>We'd love to have you back!</p>
            <p>Questions? Reply to this email or contact support.</p>
            <p><a href="https://gemmie.villebiz.com/unsubscribe?email=` + user.Email + `&token=` + user.UnsubscribeToken + `">Unsubscribe from promotional emails</a></p>
        </div>
    </div>
</body>
</html>
`
}
