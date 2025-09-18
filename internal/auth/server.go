package auth

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authpb "github.com/laerson/mancala/proto/auth"
)

// Server implements the Auth service
type Server struct {
	authpb.UnimplementedAuthServer
	storage    StorageInterface
	jwtManager *JWTManager
}

// NewServer creates a new auth server
func NewServer(dbURL, redisAddr, jwtSecret string) (*Server, error) {
	storage, err := NewStorage(dbURL, redisAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	return &Server{
		storage:    storage,
		jwtManager: NewJWTManager(jwtSecret, 24*time.Hour, 7*24*time.Hour), // 1 day access, 7 days refresh
	}, nil
}

// Register creates a new user account
func (s *Server) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	// Validate input
	if !IsValidUsername(req.Username) {
		return &authpb.RegisterResponse{
			Success: false,
			Message: "Username must be 3-30 characters and contain only letters, numbers, underscore, or hyphen",
		}, nil
	}

	if !IsValidPassword(req.Password) {
		return &authpb.RegisterResponse{
			Success: false,
			Message: "Password must be at least 8 characters long",
		}, nil
	}

	// Hash password
	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to process password")
	}

	// Create user
	userID := uuid.New().String()
	displayName := req.DisplayName
	if displayName == "" {
		displayName = req.Username
	}

	user := &User{
		UserID:       userID,
		Username:     req.Username,
		DisplayName:  displayName,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		LastLogin:    time.Now(),
	}

	// Store user
	err = s.storage.CreateUser(ctx, user)
	if err != nil {
		if err.Error() == "username already exists" {
			return &authpb.RegisterResponse{
				Success: false,
				Message: "Username already exists",
			}, nil
		}
		log.Printf("Failed to create user: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to create user")
	}

	// Generate tokens
	accessToken, err := s.jwtManager.GenerateAccessToken(user.UserID, user.Username)
	if err != nil {
		log.Printf("Failed to generate access token: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to generate tokens")
	}

	refreshTokenString, err := s.jwtManager.GenerateRefreshToken(user.UserID)
	if err != nil {
		log.Printf("Failed to generate refresh token: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to generate tokens")
	}

	// Store refresh token
	refreshToken := &RefreshToken{
		TokenID:   uuid.New().String(),
		UserID:    user.UserID,
		Token:     refreshTokenString,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now(),
	}

	err = s.storage.StoreRefreshToken(ctx, refreshToken)
	if err != nil {
		log.Printf("Failed to store refresh token: %v", err)
		// Continue anyway, user can login again
	}

	log.Printf("User registered: %s (%s)", user.Username, user.UserID)

	return &authpb.RegisterResponse{
		Success:      true,
		Message:      "User registered successfully",
		User:         s.userToProto(user),
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
	}, nil
}

// Login authenticates a user and returns JWT tokens
func (s *Server) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	// Get user by username
	user, err := s.storage.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return &authpb.LoginResponse{
			Success: false,
			Message: "Invalid credentials",
		}, nil
	}

	// Check password
	if !CheckPassword(req.Password, user.PasswordHash) {
		return &authpb.LoginResponse{
			Success: false,
			Message: "Invalid credentials",
		}, nil
	}

	// Update last login
	err = s.storage.UpdateLastLogin(ctx, user.UserID)
	if err != nil {
		log.Printf("Failed to update last login: %v", err)
		// Continue anyway
	}

	// Generate tokens
	accessToken, err := s.jwtManager.GenerateAccessToken(user.UserID, user.Username)
	if err != nil {
		log.Printf("Failed to generate access token: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to generate tokens")
	}

	refreshTokenString, err := s.jwtManager.GenerateRefreshToken(user.UserID)
	if err != nil {
		log.Printf("Failed to generate refresh token: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to generate tokens")
	}

	// Store refresh token
	refreshToken := &RefreshToken{
		TokenID:   uuid.New().String(),
		UserID:    user.UserID,
		Token:     refreshTokenString,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now(),
	}

	err = s.storage.StoreRefreshToken(ctx, refreshToken)
	if err != nil {
		log.Printf("Failed to store refresh token: %v", err)
		// Continue anyway
	}

	log.Printf("User logged in: %s (%s)", user.Username, user.UserID)

	return &authpb.LoginResponse{
		Success:      true,
		Message:      "Login successful",
		User:         s.userToProto(user),
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
	}, nil
}

// ValidateToken validates a JWT access token
func (s *Server) ValidateToken(ctx context.Context, req *authpb.ValidateTokenRequest) (*authpb.ValidateTokenResponse, error) {
	claims, err := s.jwtManager.ValidateAccessToken(req.AccessToken)
	if err != nil {
		return &authpb.ValidateTokenResponse{
			Valid:   false,
			Message: "Invalid or expired token",
		}, nil
	}

	// Get user info
	user, err := s.storage.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return &authpb.ValidateTokenResponse{
			Valid:   false,
			Message: "User not found",
		}, nil
	}

	return &authpb.ValidateTokenResponse{
		Valid:     true,
		Message:   "Token is valid",
		User:      s.userToProto(user),
		ExpiresAt: claims.ExpiresAt.Unix(),
	}, nil
}

// RefreshToken generates new tokens using a refresh token
func (s *Server) RefreshToken(ctx context.Context, req *authpb.RefreshTokenRequest) (*authpb.RefreshTokenResponse, error) {
	// Validate refresh token
	claims, err := s.jwtManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return &authpb.RefreshTokenResponse{
			Success: false,
			Message: "Invalid or expired refresh token",
		}, nil
	}

	// Check if token exists in storage
	_, err = s.storage.GetRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return &authpb.RefreshTokenResponse{
			Success: false,
			Message: "Refresh token not found",
		}, nil
	}

	// Get user
	user, err := s.storage.GetUserByID(ctx, claims.Subject)
	if err != nil {
		return &authpb.RefreshTokenResponse{
			Success: false,
			Message: "User not found",
		}, nil
	}

	// Generate new tokens
	accessToken, err := s.jwtManager.GenerateAccessToken(user.UserID, user.Username)
	if err != nil {
		log.Printf("Failed to generate access token: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to generate tokens")
	}

	newRefreshTokenString, err := s.jwtManager.GenerateRefreshToken(user.UserID)
	if err != nil {
		log.Printf("Failed to generate refresh token: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to generate tokens")
	}

	// Delete old refresh token
	err = s.storage.DeleteRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		log.Printf("Failed to delete old refresh token: %v", err)
		// Continue anyway
	}

	// Store new refresh token
	newRefreshToken := &RefreshToken{
		TokenID:   uuid.New().String(),
		UserID:    user.UserID,
		Token:     newRefreshTokenString,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now(),
	}

	err = s.storage.StoreRefreshToken(ctx, newRefreshToken)
	if err != nil {
		log.Printf("Failed to store new refresh token: %v", err)
		// Continue anyway
	}

	return &authpb.RefreshTokenResponse{
		Success:      true,
		Message:      "Tokens refreshed successfully",
		AccessToken:  accessToken,
		RefreshToken: newRefreshTokenString,
	}, nil
}

// GetProfile retrieves user profile information
func (s *Server) GetProfile(ctx context.Context, req *authpb.GetProfileRequest) (*authpb.GetProfileResponse, error) {
	user, err := s.storage.GetUserByID(ctx, req.UserId)
	if err != nil {
		return &authpb.GetProfileResponse{
			Success: false,
			Message: "User not found",
		}, nil
	}

	return &authpb.GetProfileResponse{
		Success: true,
		Message: "Profile retrieved successfully",
		User:    s.userToProto(user),
	}, nil
}

// userToProto converts internal User to protobuf User
func (s *Server) userToProto(user *User) *authpb.User {
	return &authpb.User{
		UserId:      user.UserID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		CreatedAt:   user.CreatedAt.Unix(),
		LastLogin:   user.LastLogin.Unix(),
	}
}
