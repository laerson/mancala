package auth

import (
	"testing"
	"time"
)

func TestJWTManager_GenerateAndValidateAccessToken(t *testing.T) {
	manager := NewJWTManager("test-secret", time.Hour, 24*time.Hour)
	userID := "test-user-id"
	username := "testuser"

	// Generate token
	token, err := manager.GenerateAccessToken(userID, username)
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	if token == "" {
		t.Error("Token should not be empty")
	}

	// Validate token
	claims, err := manager.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("Failed to validate access token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, claims.UserID)
	}

	if claims.Username != username {
		t.Errorf("Expected username %s, got %s", username, claims.Username)
	}

	if claims.ExpiresAt.Before(time.Now()) {
		t.Error("Token should not be expired")
	}
}

func TestJWTManager_GenerateAndValidateRefreshToken(t *testing.T) {
	manager := NewJWTManager("test-secret", time.Hour, 24*time.Hour)
	userID := "test-user-id"

	// Generate token
	token, err := manager.GenerateRefreshToken(userID)
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	if token == "" {
		t.Error("Token should not be empty")
	}

	// Validate token
	claims, err := manager.ValidateRefreshToken(token)
	if err != nil {
		t.Fatalf("Failed to validate refresh token: %v", err)
	}

	if claims.Subject != userID {
		t.Errorf("Expected subject %s, got %s", userID, claims.Subject)
	}

	if claims.ExpiresAt.Before(time.Now()) {
		t.Error("Token should not be expired")
	}
}

func TestJWTManager_ExpiredToken(t *testing.T) {
	// Create manager with very short expiration
	manager := NewJWTManager("test-secret", -time.Hour, -time.Hour)
	userID := "test-user-id"
	username := "testuser"

	// Generate expired token
	token, err := manager.GenerateAccessToken(userID, username)
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	// Try to validate expired token
	_, err = manager.ValidateAccessToken(token)
	if err == nil {
		t.Error("Expected error for expired token")
	}
}

func TestJWTManager_InvalidToken(t *testing.T) {
	manager := NewJWTManager("test-secret", time.Hour, 24*time.Hour)

	// Test invalid token
	_, err := manager.ValidateAccessToken("invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token")
	}

	// Test token with wrong secret
	wrongManager := NewJWTManager("wrong-secret", time.Hour, 24*time.Hour)
	token, _ := manager.GenerateAccessToken("user", "username")

	_, err = wrongManager.ValidateAccessToken(token)
	if err == nil {
		t.Error("Expected error for token with wrong secret")
	}
}
