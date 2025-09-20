package store

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

// Storage represents the JSON file storage
type StorageType struct {
	Users    map[string]User     `json:"users"`     // key: user_id
	UserData map[string]UserData `json:"user_data"` // key: user_id
	Mu       sync.RWMutex        `json:"-"`
}

// User represents a user in the system
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserData represents the synced data for a user
type UserData struct {
	UserID        string    `json:"user_id"`
	Chats         string    `json:"chats"`
	LinkPreviews  string    `json:"link_previews"`
	CurrentChatID string    `json:"current_chat_id"`
	UpdatedAt     time.Time `json:"updated_at"`
}

var Storage *StorageType
const storageFile = "gemmie_data.json"

func init() {
	Storage = &StorageType{
		Users:    make(map[string]User),
		UserData: make(map[string]UserData),
		Mu:       sync.RWMutex{},
	}
	loadStorage()
	log.Println("JSON file storage initialized")
}

// loadStorage loads data from JSON file
func loadStorage() {
	if _, err := os.Stat(storageFile); os.IsNotExist(err) {
		log.Println("Storage file does not exist, creating new one")
		SaveStorage()
		return
	}

	data, err := os.ReadFile(storageFile)
	if err != nil {
		log.Printf("Error reading storage file: %v", err)
		return
	}

	Storage.Mu.Lock()
	defer Storage.Mu.Unlock()

	if err := json.Unmarshal(data, Storage); err != nil {
		log.Printf("Error unmarshaling storage data: %v", err)
		return
	}

	log.Printf("Loaded %d users and %d user data records from storage", len(Storage.Users), len(Storage.UserData))
}

// saveStorage saves data to JSON file
func SaveStorage() {
	Storage.Mu.RLock()
	defer Storage.Mu.RUnlock()

	data, err := json.MarshalIndent(Storage, "", "  ")
	if err != nil {
		log.Printf("Error marshaling storage data: %v", err)
		return
	}

	if err := os.WriteFile(storageFile, data, 0644); err != nil {
		log.Printf("Error writing storage file: %v", err)
	}
}
