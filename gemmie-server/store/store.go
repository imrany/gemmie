package store

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	_ "github.com/lib/pq"
)

// Response represents API response
type Response struct {
	Success bool   `json:"success"`
	Status  int    `json:"status,omitempty"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type Modes string

const (
	ModesLightResponse Modes = "light-response"
	ModesWebSearch     Modes = "web-search"
	ModesDeepSearch    Modes = "deep-search"
)

type User struct {
	ID                      string        `json:"id"`
	Username                string        `json:"username"`
	Email                   string        `json:"email"`
	PasswordHash            string        `json:"password_hash"`
	CreatedAt               time.Time     `json:"created_at"`
	UpdatedAt               time.Time     `json:"updated_at"`
	Preferences             string        `json:"preferences,omitempty"`
	WorkFunction            string        `json:"work_function,omitempty"`
	Theme                   string        `json:"theme,omitempty"`
	SyncEnabled             bool          `json:"sync_enabled"`
	Plan                    string        `json:"plan,omitempty"`
	PlanName                string        `json:"plan_name,omitempty"`
	Amount                  int           `json:"amount,omitempty"`
	Duration                string        `json:"duration,omitempty"`
	PhoneNumber             string        `json:"phone_number,omitempty"`
	ExpiryTimestamp         int64         `json:"expiry_timestamp,omitempty"`
	ExpireDuration          int64         `json:"expire_duration,omitempty"`
	Price                   string        `json:"price,omitempty"`
	ResponseMode            Modes         `json:"response_mode,omitempty"`
	AgreeToTerms            bool          `json:"agree_to_terms"`
	RequestCount            RequestCount  `json:"request_count"`
	EmailVerified           bool          `json:"email_verified"`
	EmailSubscribed         bool          `json:"email_subscribed"`
	VerificationToken       string        `json:"verification_token,omitempty"`
	VerificationTokenExpiry time.Time     `json:"verification_token_expiry"`
	UnsubscribeToken        string        `json:"unsubscribe_token,omitempty"`
	UserTransactions        []Transaction `json:"user_transactions,omitempty"`
	UserAgent               string        `json:"user_agent"`
}

type RequestCount struct {
	Count     int   `json:"count"`
	Timestamp int64 `json:"timestamp"`
}

type Transaction struct {
	ID                 string    `json:"id"`
	ExternalReference  string    `json:"ExternalReference"`
	MpesaReceiptNumber string    `json:"MpesaReceiptNumber"`
	CheckoutRequestID  string    `json:"CheckoutRequestID"`
	MerchantRequestID  string    `json:"MerchantRequestID"`
	Amount             int       `json:"Amount"`
	PhoneNumber        string    `json:"Phone"`
	ResultCode         int       `json:"ResultCode"`
	ResultDescription  string    `json:"ResultDesc"`
	Status             string    `json:"Status"`
	CreatedAt          time.Time `json:"CreatedAt"`
	UpdatedAt          time.Time `json:"UpdatedAt"`
}

type PlatformError struct {
	ID          string    `json:"id"`
	Message     string    `json:"message"`
	Description string    `json:"description,omitempty"`
	Action      string    `json:"action"`
	Status      string    `json:"status,omitempty"`
	UserId      string    `json:"user_id"`
	Context     string    `json:"context,omitempty"`
	Severity    string    `json:"severity"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   time.Time `json:"UpdatedAt"`
}

type Chat struct {
	ID            string    `json:"id"`
	UserId        string    `json:"user_id"`
	Title         string    `json:"title"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	IsArchived    bool      `json:"is_archived"`
	MessageCount  int       `json:"message_count"`
	LastMessageAt time.Time `json:"last_message_at"`
	IsPrivate     bool      `json:"is_private"`
	Messages      []Message `json:"messages"`
	IsReadOnly    bool      `json:"is_read_only"`
}

type Message struct {
	ID         string    `json:"id"`
	ChatId     string    `json:"chat_id"`
	Prompt     string    `json:"prompt,omitempty"`
	Response   string    `json:"response"`
	CreatedAt  time.Time `json:"created_at"`
	Model      string    `json:"model,omitempty"`
	References []string  `json:"references,omitempty"`
}

type Arcade struct {
	ID          string    `json:"id,omitempty"`
	UserId      string    `json:"user_id"`
	Code        string    `json:"code"`
	Label       string    `json:"label"`
	Description string    `json:"description"`
	CodeType    string    `json:"code_type"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PushSubscription represents a stored subscription
type PushSubscription struct {
	UserID    string    `json:"user_id"`
	Endpoint  string    `json:"endpoint"`
	P256dhKey string    `json:"p256dh_key"`
	AuthKey   string    `json:"auth_key"`
	UserAgent string    `json:"user_agent,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SubscriptionRequest from frontend
type SubscriptionRequest struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

// NotificationPayload represents the notification data
type NotificationPayload struct {
	Title              string         `json:"title"`
	Body               string         `json:"body"`
	Icon               string         `json:"icon,omitempty"`
	Badge              string         `json:"badge,omitempty"`
	Image              string         `json:"image,omitempty"`
	Data               map[string]any `json:"data,omitempty"`
	Tag                string         `json:"tag,omitempty"`
	RequireInteraction bool           `json:"requireInteraction,omitempty"`
}

// SendNotificationRequest for sending to specific users
type SendNotificationRequest struct {
	UserIDs []string            `json:"user_ids,omitempty"` // If empty, sends to all
	Payload NotificationPayload `json:"payload"`
}

var (
	DB *sql.DB
)

// InitStorage initializes the PostgreSQL database connection and runs migrations
func InitStorage(connString string) error {
	return initStorage(connString, true)
}

// InitStorageWithoutMigration initializes connection without running migrations
func InitStorageWithoutMigration(connString string) error {
	return initStorage(connString, false)
}

// initStorage is the internal initialization function
func initStorage(connString string, runMigrations bool) error {
	var err error
	DB, err = sql.Open("postgres", connString)
	if err != nil {
		slog.Error("Failed to open database connection", "error", err)
		return err
	}

	// Set connection pool settings
	DB.SetMaxOpenConns(100) // Based on server capacity
	DB.SetMaxIdleConns(25)  // Keep connections ready
	DB.SetConnMaxLifetime(5 * time.Minute)
	DB.SetConnMaxIdleTime(10 * time.Minute)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := DB.PingContext(ctx); err != nil {
		slog.Error("Failed to ping database", "error", err)
		return err
	}

	// Run migrations if requested
	if runMigrations {
		if err := RunMigrations(); err != nil {
			slog.Error("Failed to run migrations", "error", err)
			return err
		}
	}

	slog.Info("PostgreSQL storage initialized")
	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

func GetVersion() string {
	return "v0.25.5"
}
