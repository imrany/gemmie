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

type UserData struct {
	UserID       string    `json:"user_id"`
	Chats        string    `json:"chats"`
	LinkPreviews string    `json:"link_previews"`
	UpdatedAt    time.Time `json:"updated_at"`
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
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)

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
	return "v0.8.0"
}
