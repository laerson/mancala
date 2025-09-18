package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "testpassword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash == "" {
		t.Error("Hash should not be empty")
	}

	if hash == password {
		t.Error("Hash should not equal original password")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "testpassword123"
	wrongPassword := "wrongpassword"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Test correct password
	if !CheckPassword(password, hash) {
		t.Error("CheckPassword should return true for correct password")
	}

	// Test wrong password
	if CheckPassword(wrongPassword, hash) {
		t.Error("CheckPassword should return false for wrong password")
	}
}

func TestIsValidUsername(t *testing.T) {
	tests := []struct {
		username string
		valid    bool
	}{
		{"user", true},
		{"user123", true},
		{"user_name", true},
		{"user-name", true},
		{"User123", true},
		{"us", false}, // too short
		{"", false},   // empty
		{"a", false},  // too short
		{"verylongusernamethatexceedsthirtycharacterslimit", false}, // too long
		{"user@name", false}, // invalid character
		{"user name", false}, // space
		{"user.name", false}, // dot
		{"user#name", false}, // hash
	}

	for _, test := range tests {
		result := IsValidUsername(test.username)
		if result != test.valid {
			t.Errorf("IsValidUsername(%q) = %v, want %v", test.username, result, test.valid)
		}
	}
}

func TestIsValidPassword(t *testing.T) {
	tests := []struct {
		password string
		valid    bool
	}{
		{"password", true},
		{"12345678", true},
		{"complex123", true},
		{"a", false},       // too short
		{"", false},        // empty
		{"1234567", false}, // too short (7 chars)
	}

	for _, test := range tests {
		result := IsValidPassword(test.password)
		if result != test.valid {
			t.Errorf("IsValidPassword(%q) = %v, want %v", test.password, result, test.valid)
		}
	}
}
