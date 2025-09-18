package auth

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user account in the database
type User struct {
	UserID       string    `json:"user_id" redis:"user_id"`
	Username     string    `json:"username" redis:"username"`
	DisplayName  string    `json:"display_name" redis:"display_name"`
	PasswordHash string    `json:"password_hash" redis:"password_hash"`
	CreatedAt    time.Time `json:"created_at" redis:"created_at"`
	LastLogin    time.Time `json:"last_login" redis:"last_login"`
}

// RefreshToken represents a refresh token in the database
type RefreshToken struct {
	TokenID   string    `json:"token_id" redis:"token_id"`
	UserID    string    `json:"user_id" redis:"user_id"`
	Token     string    `json:"token" redis:"token"`
	ExpiresAt time.Time `json:"expires_at" redis:"expires_at"`
	CreatedAt time.Time `json:"created_at" redis:"created_at"`
}

// HashPassword hashes a plain text password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword compares a plain text password with a hashed password
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// IsValidUsername checks if username meets requirements
func IsValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 30 {
		return false
	}

	// Check for valid characters (alphanumeric, underscore, hyphen)
	for _, char := range username {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '-') {
			return false
		}
	}

	return true
}

// IsValidPassword checks if password meets requirements
func IsValidPassword(password string) bool {
	return len(password) >= 8
}
