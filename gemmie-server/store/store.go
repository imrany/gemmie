package store

import (
	"encoding/json"
	"log/slog"
	"os"
	"sync"
	"time"
)

type StorageType struct {
	Users    map[string]User     `json:"users"`     // key: user_id
	UserData map[string]UserData `json:"user_data"` // key: user_id
	Mu       sync.RWMutex        `json:"-"`
}

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserData struct {
	UserID        string    `json:"user_id"`
	Chats         string    `json:"chats"`
	LinkPreviews  string    `json:"link_previews"`
	CurrentChatID string    `json:"current_chat_id"`
	UpdatedAt     time.Time `json:"updated_at"`
}

var Storage *StorageType
var storageFile string

// InitStorage initializes the storage system with the given file path.
func InitStorage(filePath string) {
	storageFile = filePath

	Storage = &StorageType{
		Users:    make(map[string]User),
		UserData: make(map[string]UserData),
		Mu:       sync.RWMutex{},
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
		return
	}

	slog.Info("Storage loaded",
		"users", len(Storage.Users),
		"user_data_records", len(Storage.UserData),
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
