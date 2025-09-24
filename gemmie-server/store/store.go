package store

import (
	"encoding/json"
	"log/slog"
	"os"
	"sync"
	"time"
)

// Response represents API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type StorageType struct {
	Users        map[string]User                 `json:"users"`         // key: user_id
	UserData     map[string]UserData             `json:"user_data"`     // key: user_id
	Transactions map[string]Transaction `json:"transactions"` // key: transaction_id
	Mu           sync.RWMutex                    `json:"-"`
}

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
	Plan 	   string    `json:"plan,omitempty"`
	PlanName	 string    `json:"plan_name,omitempty"`
	Amount       int       `json:"amount,omitempty"`
	Duration     string    `json:"duration,omitempty"`
	PhoneNumber        string    `json:"phone_number,omitempty"`
	ExpiryTimestamp int64  `json:"expiry_timestamp,omitempty"`
	ExpireDuration int64   `json:"expire_duration,omitempty"`
	Price        string    `json:"price,omitempty"`
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
	ResultCode         string    `json:"ResultCode"`
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
		Storage = &StorageType{
			Users:        make(map[string]User),
			UserData:     make(map[string]UserData),
			Transactions: make(map[string]Transaction),
			Mu:           sync.RWMutex{},
		}

		slog.Warn("Storage reset due to unmarshaling error")
		return
	}

	slog.Info("Storage loaded",
		"users", len(Storage.Users),
		"user_data_records", len(Storage.UserData),
		"transactions", len(Storage.Transactions),
	)
}

func SaveStorage() error {
	Storage.Mu.Lock()
	defer Storage.Mu.Unlock()

	data, err := json.MarshalIndent(Storage, "", "  ")
	if err != nil {
		slog.Error("Error marshaling storage data", "error", err)
		return err
	}

	if err := os.WriteFile(storageFile, data, 0644); err != nil {
		slog.Error("Error writing storage file", "error", err)
		return err
	}

	slog.Debug("Storage saved", "path", storageFile)
	return nil
}