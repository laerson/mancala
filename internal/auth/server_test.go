package auth

import (
	"context"
	"fmt"
	"testing"
	"time"

	authpb "github.com/laerson/mancala/proto/auth"
)

// Mock storage for testing
type mockStorage struct {
	users         map[string]*User
	usersByName   map[string]string
	refreshTokens map[string]*RefreshToken
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		users:         make(map[string]*User),
		usersByName:   make(map[string]string),
		refreshTokens: make(map[string]*RefreshToken),
	}
}

func (m *mockStorage) CreateUser(ctx context.Context, user *User) error {
	if _, exists := m.usersByName[user.Username]; exists {
		return fmt.Errorf("username already exists")
	}
	m.users[user.UserID] = user
	m.usersByName[user.Username] = user.UserID
	return nil
}

func (m *mockStorage) GetUserByID(ctx context.Context, userID string) (*User, error) {
	if user, exists := m.users[userID]; exists {
		return user, nil
	}
	return nil, fmt.Errorf("user not found")
}

func (m *mockStorage) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	if userID, exists := m.usersByName[username]; exists {
		return m.GetUserByID(ctx, userID)
	}
	return nil, fmt.Errorf("user not found")
}

func (m *mockStorage) UpdateLastLogin(ctx context.Context, userID string) error {
	return nil // Mock implementation
}

func (m *mockStorage) StoreRefreshToken(ctx context.Context, token *RefreshToken) error {
	m.refreshTokens[token.Token] = token
	return nil
}

func (m *mockStorage) GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error) {
	if refreshToken, exists := m.refreshTokens[token]; exists {
		return refreshToken, nil
	}
	return nil, fmt.Errorf("refresh token not found")
}

func (m *mockStorage) DeleteRefreshToken(ctx context.Context, token string) error {
	delete(m.refreshTokens, token)
	return nil
}

func (m *mockStorage) Close() error {
	return nil
}

func TestServer_Register(t *testing.T) {
	server := &Server{
		storage:    nil, // We'll override methods
		jwtManager: NewJWTManager("test-secret", time.Hour, 24*time.Hour),
	}

	tests := []struct {
		name        string
		req         *authpb.RegisterRequest
		wantErr     bool
		wantSuccess bool
	}{
		{
			name: "valid registration",
			req: &authpb.RegisterRequest{
				Username: "testuser",
				Password: "password123",
			},
			wantErr:     false,
			wantSuccess: true,
		},
		{
			name: "invalid username",
			req: &authpb.RegisterRequest{
				Username: "ab", // too short
				Password: "password123",
			},
			wantErr:     false,
			wantSuccess: false,
		},
		{
			name: "invalid password",
			req: &authpb.RegisterRequest{
				Username: "testuser",
				Password: "pass", // too short
			},
			wantErr:     false,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use mock storage for each test
			server.storage = newMockStorage()

			resp, err := server.Register(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if resp.Success != tt.wantSuccess {
				t.Errorf("Register() success = %v, wantSuccess %v", resp.Success, tt.wantSuccess)
			}

			if tt.wantSuccess && resp.AccessToken == "" {
				t.Error("Expected access token in successful registration")
			}

			if tt.wantSuccess && resp.User == nil {
				t.Error("Expected user info in successful registration")
			}
		})
	}
}

func TestServer_Login(t *testing.T) {
	server := &Server{
		storage:    newMockStorage(),
		jwtManager: NewJWTManager("test-secret", time.Hour, 24*time.Hour),
	}

	// Create a test user first
	hashedPassword, _ := HashPassword("password123")
	testUser := &User{
		UserID:       "test-user-id",
		Username:     "testuser",
		DisplayName:  "Test User",
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
		LastLogin:    time.Now(),
	}
	server.storage.(*mockStorage).users[testUser.UserID] = testUser
	server.storage.(*mockStorage).usersByName[testUser.Username] = testUser.UserID

	tests := []struct {
		name        string
		req         *authpb.LoginRequest
		wantSuccess bool
	}{
		{
			name: "valid login",
			req: &authpb.LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			wantSuccess: true,
		},
		{
			name: "wrong password",
			req: &authpb.LoginRequest{
				Username: "testuser",
				Password: "wrongpassword",
			},
			wantSuccess: false,
		},
		{
			name: "user not found",
			req: &authpb.LoginRequest{
				Username: "nonexistent",
				Password: "password123",
			},
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := server.Login(context.Background(), tt.req)

			if err != nil {
				t.Errorf("Login() error = %v", err)
				return
			}

			if resp.Success != tt.wantSuccess {
				t.Errorf("Login() success = %v, wantSuccess %v", resp.Success, tt.wantSuccess)
			}

			if tt.wantSuccess && resp.AccessToken == "" {
				t.Error("Expected access token in successful login")
			}
		})
	}
}

func TestServer_ValidateToken(t *testing.T) {
	server := &Server{
		storage:    newMockStorage(),
		jwtManager: NewJWTManager("test-secret", time.Hour, 24*time.Hour),
	}

	// Create a test user
	testUser := &User{
		UserID:      "test-user-id",
		Username:    "testuser",
		DisplayName: "Test User",
		CreatedAt:   time.Now(),
		LastLogin:   time.Now(),
	}
	server.storage.(*mockStorage).users[testUser.UserID] = testUser

	// Generate a valid token
	validToken, err := server.jwtManager.GenerateAccessToken(testUser.UserID, testUser.Username)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	tests := []struct {
		name      string
		token     string
		wantValid bool
	}{
		{
			name:      "valid token",
			token:     validToken,
			wantValid: true,
		},
		{
			name:      "invalid token",
			token:     "invalid-token",
			wantValid: false,
		},
		{
			name:      "empty token",
			token:     "",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &authpb.ValidateTokenRequest{
				AccessToken: tt.token,
			}

			resp, err := server.ValidateToken(context.Background(), req)

			if err != nil {
				t.Errorf("ValidateToken() error = %v", err)
				return
			}

			if resp.Valid != tt.wantValid {
				t.Errorf("ValidateToken() valid = %v, wantValid %v", resp.Valid, tt.wantValid)
			}

			if tt.wantValid && resp.User == nil {
				t.Error("Expected user info for valid token")
			}
		})
	}
}
