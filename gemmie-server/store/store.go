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
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type Modes string

const (
	ModesLightResponse Modes = "light-response"
	ModesWebSearch     Modes = "web-search"
	ModesDeepSearch    Modes = "deep-search"
)

type User struct {
	ID                      string       `json:"id"`
	Username                string       `json:"username"`
	Email                   string       `json:"email"`
	PasswordHash            string       `json:"password_hash"`
	CreatedAt               time.Time    `json:"created_at"`
	UpdatedAt               time.Time    `json:"updated_at"`
	Preferences             string       `json:"preferences,omitempty"`
	WorkFunction            string       `json:"work_function,omitempty"`
	Theme                   string       `json:"theme,omitempty"`
	SyncEnabled             bool         `json:"sync_enabled"`
	Plan                    string       `json:"plan,omitempty"`
	PlanName                string       `json:"plan_name,omitempty"`
	Amount                  int          `json:"amount,omitempty"`
	Duration                string       `json:"duration,omitempty"`
	PhoneNumber             string       `json:"phone_number,omitempty"`
	ExpiryTimestamp         int64        `json:"expiry_timestamp,omitempty"`
	ExpireDuration          int64        `json:"expire_duration,omitempty"`
	Price                   string       `json:"price,omitempty"`
	ResponseMode            Modes        `json:"response_mode,omitempty"`
	AgreeToTerms            bool         `json:"agree_to_terms"`
	RequestCount            RequestCount `json:"request_count,omitempty"`
	EmailVerified           bool         `json:"email_verified"`
	EmailSubscribed         bool         `json:"email_subscribed"`
	VerificationToken       string       `json:"verification_token,omitempty"`
	VerificationTokenExpiry time.Time    `json:"verification_token_expiry,omitempty"`
	UnsubscribeToken        string       `json:"unsubscribe_token,omitempty"`
}

type RequestCount struct {
	Count     int   `json:"count"`
	Timestamp int64 `json:"timestamp"`
}

type UserData struct {
	UserID        string    `json:"user_id"`
	Chats         string    `json:"chats"`
	LinkPreviews  string    `json:"link_previews"`
	CurrentChatID string    `json:"current_chat_id"`
	UpdatedAt     time.Time `json:"updated_at"`
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

var (
	DB *sql.DB
)

// InitStorage initializes the PostgreSQL database connection
func InitStorage(connString string) error {
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

	// Create tables
	if err := createTables(); err != nil {
		slog.Error("Failed to create tables", "error", err)
		return err
	}

	slog.Info("PostgreSQL storage initialized")
	return nil
}

// createTables creates all necessary tables
func createTables() error {
	ctx := context.Background()

	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			preferences TEXT,
			work_function TEXT,
			theme TEXT,
			sync_enabled BOOLEAN DEFAULT false,
			plan TEXT,
			plan_name TEXT,
			amount INTEGER DEFAULT 0,
			duration TEXT,
			phone_number TEXT,
			expiry_timestamp BIGINT DEFAULT 0,
			expire_duration BIGINT DEFAULT 0,
			price TEXT,
			response_mode TEXT DEFAULT 'light-response',
			agree_to_terms BOOLEAN DEFAULT false,
			request_count_value INTEGER DEFAULT 0,
			request_count_timestamp BIGINT DEFAULT 0,
			email_verified BOOLEAN DEFAULT false,
			email_subscribed BOOLEAN DEFAULT true,
			verification_token TEXT,
			verification_token_expiry TIMESTAMP,
			unsubscribe_token TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`,
		`CREATE TABLE IF NOT EXISTS user_data (
			user_id TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			chats TEXT,
			link_previews TEXT,
			current_chat_id TEXT,
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS transactions (
			id TEXT PRIMARY KEY,
			external_reference TEXT,
			mpesa_receipt_number TEXT,
			checkout_request_id TEXT,
			merchant_request_id TEXT,
			amount INTEGER,
			phone_number TEXT,
			result_code INTEGER,
			result_description TEXT,
			status TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_external_ref ON transactions(external_reference)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_phone ON transactions(phone_number)`,
	}

	for _, query := range queries {
		if _, err := DB.ExecContext(ctx, query); err != nil {
			return err
		}
	}

	slog.Info("Database tables created successfully")
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
	return "v0.6.0"
}