package encrypt

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// generateUserID generates a unique user ID
func GenerateUserID() string {
	return fmt.Sprintf("user_%d", time.Now().UnixNano())
}

// hashCredentials creates a SHA-256 hash of username + email + password
func HashCredentials(username, email, password string) string {
	combined := username + email + password
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}