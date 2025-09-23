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
	Orders       map[string]Order         `json:"orders"`        // key: order_id
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

// Order represents an order
type Order struct {
	ID                string    `json:"id"`
	ExternalReference string    `json:"external_reference"`
	Username      string    `json:"username"`
	Email     string    `json:"email"`
	PhoneNumber     string    `json:"phone_number"`
	Amount            int       `json:"amount"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// Transaction represents a payment transaction
type Transaction struct {
	ID                 string    `json:"id"`
	ExternalReference  string    `json:"external_reference"`
	MpesaReceiptNumber string    `json:"mpesa_receipt_number"`
	CheckoutRequestID  string    `json:"checkout_request_id"`
	MerchantRequestID  string    `json:"merchant_request_id"`
	Amount             int       `json:"amount"`
	PhoneNumber        string    `json:"phone_number"`
	ResultCode         string    `json:"result_code"`
	ResultDescription  string    `json:"result_description"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
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
		Orders:       make(map[string]Order),
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
			Orders:       make(map[string]Order),
			Mu:           sync.RWMutex{},
		}

		slog.Warn("Storage reset due to unmarshaling error")
		return
	}

	slog.Info("Storage loaded",
		"users", len(Storage.Users),
		"user_data_records", len(Storage.UserData),
		"orders", len(Storage.Orders),
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