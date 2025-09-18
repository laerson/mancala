package gateway

import (
	"net/http"

	"github.com/gin-gonic/gin"
	authpb "github.com/laerson/mancala/proto/auth"
)

// AuthHandlers handles authentication related endpoints
type AuthHandlers struct {
	clients *ServiceClients
}

// NewAuthHandlers creates new auth handlers
func NewAuthHandlers(clients *ServiceClients) *AuthHandlers {
	return &AuthHandlers{clients: clients}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login handles user login
func (h *AuthHandlers) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call Auth service
	resp, err := h.clients.Auth.Login(addGRPCContext(c), &authpb.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       resp.Success,
		"message":       resp.Message,
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
		"user":          resp.User,
	})
}

// Register handles user registration
func (h *AuthHandlers) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call Auth service
	resp, err := h.clients.Auth.Register(addGRPCContext(c), &authpb.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Registration failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":       resp.Success,
		"message":       resp.Message,
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
		"user":          resp.User,
	})
}

// ValidateToken handles token validation
func (h *AuthHandlers) ValidateToken(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token parameter required"})
		return
	}

	// Call Auth service
	resp, err := h.clients.Auth.ValidateToken(addGRPCContext(c), &authpb.ValidateTokenRequest{
		AccessToken: token,
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":      resp.Valid,
		"message":    resp.Message,
		"user":       resp.User,
		"expires_at": resp.ExpiresAt,
	})
}
