package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

// StorageInterface defines the storage operations for authentication
type StorageInterface interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByID(ctx context.Context, userID string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	UpdateLastLogin(ctx context.Context, userID string) error
	StoreRefreshToken(ctx context.Context, refreshToken *RefreshToken) error
	GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, token string) error
	Close() error
}

// Storage handles user data persistence using PostgreSQL and Redis
type Storage struct {
	db          *sql.DB
	redisClient *redis.Client
}

// NewStorage creates a new storage instance
func NewStorage(dbURL, redisAddr string) (*Storage, error) {
	// Connect to PostgreSQL
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	// Connect to Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	storage := &Storage{
		db:          db,
		redisClient: redisClient,
	}

	// Initialize database schema
	if err := storage.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize database schema: %w", err)
	}

	return storage, nil
}

// initSchema creates the necessary database tables
func (s *Storage) initSchema() error {
	queries := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`,
		`CREATE TABLE IF NOT EXISTS users (
			user_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			username VARCHAR(30) UNIQUE NOT NULL,
			display_name VARCHAR(100) NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			last_login TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			CONSTRAINT username_length CHECK (LENGTH(username) >= 3),
			CONSTRAINT username_format CHECK (username ~ '^[a-zA-Z0-9_-]+$')
		);`,
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);`,
		`CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute schema query: %w", err)
		}
	}

	return nil
}

// CreateUser stores a new user in PostgreSQL
func (s *Storage) CreateUser(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (user_id, username, display_name, password_hash, created_at, last_login)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := s.db.ExecContext(ctx, query,
		user.UserID,
		user.Username,
		user.DisplayName,
		user.PasswordHash,
		user.CreatedAt,
		user.LastLogin,
	)

	if err != nil {
		// Check for unique constraint violation (username already exists)
		if err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"` {
			return fmt.Errorf("username already exists")
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUserByID retrieves a user by their UUID from PostgreSQL
func (s *Storage) GetUserByID(ctx context.Context, userID string) (*User, error) {
	query := `
		SELECT user_id, username, display_name, password_hash, created_at, last_login
		FROM users
		WHERE user_id = $1
	`

	var user User
	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&user.UserID,
		&user.Username,
		&user.DisplayName,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.LastLogin,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUserByUsername retrieves a user by their username from PostgreSQL
func (s *Storage) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	query := `
		SELECT user_id, username, display_name, password_hash, created_at, last_login
		FROM users
		WHERE username = $1
	`

	var user User
	err := s.db.QueryRowContext(ctx, query, username).Scan(
		&user.UserID,
		&user.Username,
		&user.DisplayName,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.LastLogin,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// UpdateLastLogin updates the user's last login timestamp in PostgreSQL
func (s *Storage) UpdateLastLogin(ctx context.Context, userID string) error {
	query := `
		UPDATE users
		SET last_login = NOW()
		WHERE user_id = $1
	`

	result, err := s.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// StoreRefreshToken stores a refresh token in Redis with expiration
func (s *Storage) StoreRefreshToken(ctx context.Context, refreshToken *RefreshToken) error {
	tokenData, err := json.Marshal(refreshToken)
	if err != nil {
		return fmt.Errorf("failed to marshal refresh token: %w", err)
	}

	// Store with TTL based on expiration
	ttl := time.Until(refreshToken.ExpiresAt)
	if ttl <= 0 {
		return fmt.Errorf("refresh token already expired")
	}

	pipe := s.redisClient.TxPipeline()

	// Store token by token ID
	pipe.Set(ctx, fmt.Sprintf("refresh_token:%s", refreshToken.TokenID), tokenData, ttl)

	// Store mapping from token string to token ID
	pipe.Set(ctx, fmt.Sprintf("refresh_token_lookup:%s", refreshToken.Token), refreshToken.TokenID, ttl)

	// Add to user's active tokens set
	pipe.SAdd(ctx, fmt.Sprintf("user_refresh_tokens:%s", refreshToken.UserID), refreshToken.TokenID)
	pipe.Expire(ctx, fmt.Sprintf("user_refresh_tokens:%s", refreshToken.UserID), ttl)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}

	return nil
}

// GetRefreshToken retrieves a refresh token by token string from Redis
func (s *Storage) GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error) {
	// Get token ID from lookup
	tokenID, err := s.redisClient.Get(ctx, fmt.Sprintf("refresh_token_lookup:%s", token)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("refresh token not found")
		}
		return nil, fmt.Errorf("failed to lookup refresh token: %w", err)
	}

	// Get token data
	data, err := s.redisClient.Get(ctx, fmt.Sprintf("refresh_token:%s", tokenID)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("refresh token not found")
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	var refreshToken RefreshToken
	if err := json.Unmarshal([]byte(data), &refreshToken); err != nil {
		return nil, fmt.Errorf("failed to unmarshal refresh token: %w", err)
	}

	return &refreshToken, nil
}

// DeleteRefreshToken removes a refresh token from Redis
func (s *Storage) DeleteRefreshToken(ctx context.Context, token string) error {
	refreshToken, err := s.GetRefreshToken(ctx, token)
	if err != nil {
		return err
	}

	pipe := s.redisClient.TxPipeline()

	// Remove token data
	pipe.Del(ctx, fmt.Sprintf("refresh_token:%s", refreshToken.TokenID))

	// Remove lookup mapping
	pipe.Del(ctx, fmt.Sprintf("refresh_token_lookup:%s", token))

	// Remove from user's active tokens set
	pipe.SRem(ctx, fmt.Sprintf("user_refresh_tokens:%s", refreshToken.UserID), refreshToken.TokenID)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	return nil
}

// Close closes both PostgreSQL and Redis connections
func (s *Storage) Close() error {
	var dbErr, redisErr error

	if s.db != nil {
		dbErr = s.db.Close()
	}

	if s.redisClient != nil {
		redisErr = s.redisClient.Close()
	}

	if dbErr != nil {
		return fmt.Errorf("failed to close database: %w", dbErr)
	}

	if redisErr != nil {
		return fmt.Errorf("failed to close redis: %w", redisErr)
	}

	return nil
}
