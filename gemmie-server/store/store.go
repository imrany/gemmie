package store

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/imrany/gemmie/gemmie-server/internal/encrypt"
)

// Response represents API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type StorageType struct {
	Users        map[string]User        `json:"users"`         // key: user_id
	UserData     map[string]UserData    `json:"user_data"`     // key: user_id
	Transactions map[string]Transaction `json:"transactions"` // key: transaction_id
	Mu           sync.RWMutex           `json:"-"`
}

type Modes string

const (
	ModesLightResponse Modes = "light-response"
	ModesWebSearch     Modes = "web-search"
	ModesDeepSearch    Modes = "deep-search"
)

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Preferences  string    `json:"preferences,omitempty"`
	WorkFunction string    `json:"work_function,omitempty"`
	Theme        string    `json:"theme,omitempty"`
	SyncEnabled  bool      `json:"sync_enabled"`
	Plan         string    `json:"plan,omitempty"`
	PlanName     string    `json:"plan_name,omitempty"`
	Amount       int       `json:"amount,omitempty"`
	Duration     string    `json:"duration,omitempty"`
	PhoneNumber  string    `json:"phone_number,omitempty"`
	ExpiryTimestamp int64  `json:"expiry_timestamp,omitempty"`
	ExpireDuration int64   `json:"expire_duration,omitempty"`
	Price        string    `json:"price,omitempty"`
	ResponseMode Modes    `json:"response_mode,omitempty"`
	AgreeToTerms bool      `json:"agree_to_terms"`
	
	EmailVerified   bool      `json:"email_verified"`      // Whether email is verified
	EmailSubscribed bool      `json:"email_subscribed"`    // Whether user subscribed to promotional emails
	VerificationToken string  `json:"verification_token,omitempty"` // Token for email verification
	VerificationTokenExpiry time.Time `json:"verification_token_expiry,omitempty"` // Token expiry
	UnsubscribeToken string   `json:"unsubscribe_token,omitempty"` // Secure token for unsubscribe
}

type UserData struct {
	UserID        string    `json:"user_id"`
	Chats         string    `json:"chats"`
	LinkPreviews  string    `json:"link_previews"`
	CurrentChatID string    `json:"current_chat_id"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Transaction represents a payment transaction
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
	Storage     *StorageType
	storageFile string
)

// InitStorage initializes the storage system with the given file path.
func InitStorage(filePath string) {
	storageFile = filePath

	// Initialize storage
	Storage = &StorageType{
		Users:        make(map[string]User),
		UserData:     make(map[string]UserData),
		Transactions: make(map[string]Transaction),
		Mu:           sync.RWMutex{},
	}

	loadStorage()
	slog.Info("JSON file storage initialized", "path", storageFile)
}

func loadStorage() {
	if _, err := os.Stat(storageFile); os.IsNotExist(err) {
		slog.Warn("Storage file does not exist, creating new one", "path", storageFile)
		if err := SaveStorage(); err != nil {
			slog.Error("Failed to create initial storage file", "error", err)
			os.Exit(1)
		}
		return
	}

	data, err := os.ReadFile(storageFile)
	if err != nil {
		slog.Error("Error reading storage file", "error", err)
		return
	}

	Storage.Mu.Lock()
	defer Storage.Mu.Unlock()

	if err := json.Unmarshal(data, Storage); err != nil {
		slog.Error("Error unmarshaling storage data", "error", err)
		// Create backup of corrupted file
		backupFile := storageFile + ".corrupted." + time.Now().Format("20060102_150405")
		if err := os.WriteFile(backupFile, data, 0644); err != nil {
			slog.Error("Failed to create backup of corrupted file", "error", err)
		}
		
		// Reset storage to empty state
		Storage.Users = make(map[string]User)
		Storage.UserData = make(map[string]UserData)
		Storage.Transactions = make(map[string]Transaction)
		
		slog.Warn("Storage reset due to unmarshaling error, backup created", "backup", backupFile)
		return
	}

	// Migration: Set default values for new fields for existing users
	needsSave := false
	for userID, user := range Storage.Users {
		// Check if this user needs migration
		// Migrate new fields for User struct
		updated := false

		if user.ResponseMode == "" {
			user.ResponseMode = "light-response"
			updated = true
		}

		if user.UnsubscribeToken == "" {
			user.UnsubscribeToken = encrypt.GenerateUnsubscribeToken(userID)
			updated = true
		}
		if !user.EmailSubscribed {
			user.EmailSubscribed = true
			updated = true
		}

		// Add more migrations for any new fields here, e.g.:
		// if user.NewField == "" { user.NewField = "defaultValue"; updated = true }

		if updated {
			slog.Debug("Migrating user", "user_id", userID, "email", user.Email)
			Storage.Users[userID] = user
			needsSave = true
		}
	}

	if needsSave {
		slog.Info("Migrated users with new email fields", "count", len(Storage.Users))
		if err := saveStorageInternal(); err != nil {
			slog.Error("Failed to save storage after migration", "error", err)
		}
	}

	slog.Info("Storage loaded",
		"users", len(Storage.Users),
		"user_data_records", len(Storage.UserData),
		"transactions", len(Storage.Transactions),
	)
}

// Helper function to save without acquiring lock (caller must hold lock)
func saveStorageInternal() error {
	// Create backup before saving
	if err := createBackup(); err != nil {
		slog.Warn("Failed to create backup", "error", err)
	}

	data, err := json.MarshalIndent(Storage, "", "  ")
	if err != nil {
		slog.Error("Error marshaling storage data", "error", err)
		return err
	}

	// Write to temporary file first for atomic operation
	tempFile := storageFile + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		slog.Error("Error writing temporary storage file", "error", err)
		return err
	}

	// Atomically replace the old file
	if err := os.Rename(tempFile, storageFile); err != nil {
		slog.Error("Error replacing storage file", "error", err)
		return err
	}

	slog.Debug("Storage saved", "path", storageFile)
	return nil
}

// createBackup creates a backup of the current storage file
func createBackup() error {
	if _, err := os.Stat(storageFile); os.IsNotExist(err) {
		return nil // No file to backup
	}

	// Use filepath to safely handle file extensions
	ext := filepath.Ext(storageFile)
	base := storageFile[:len(storageFile)-len(ext)]
	backupFile := base + ".backup" + ext

	data, err := os.ReadFile(storageFile)
	if err != nil {
		return err
	}

	return os.WriteFile(backupFile, data, 0644)
}

func SaveStorage() error {
	Storage.Mu.Lock()
	defer Storage.Mu.Unlock()

	return saveStorageInternal()
}

func GetVersion() string{
	return "v0.4.22"
}