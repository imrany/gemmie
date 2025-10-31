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

func GenerateID(prefix *string) string {
	Fprefix := ""
	if prefix != nil {
		Fprefix = fmt.Sprintf("%s_", *prefix)
	}
	return fmt.Sprintf("%s%d", Fprefix, time.Now().UnixNano())
}

// hashCredentials creates a SHA-256 hash of username + email + password
func HashCredentials(username, email, password string) string {
	combined := username + email + password
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

// Helper function to generate unsubscribe token
func GenerateUnsubscribeToken(userID string) string {
	data := fmt.Sprintf("%s:%d:%s", userID, time.Now().Unix(), "unsubscribe")
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func GenerateVerificationToken(email, userID string) string {
	data := fmt.Sprintf("%s:%s:%d", email, userID, time.Now().Unix())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
