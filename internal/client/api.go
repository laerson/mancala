package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// APIClient handles communication with the API Gateway
type APIClient struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

// NewAPIClient creates a new API client
func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetToken sets the authentication token
func (c *APIClient) SetToken(token string) {
	c.token = token
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterResponse represents a registration response
type RegisterResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

// User represents user information
type User struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	CreatedAt   int64  `json:"created_at"`
	LastLogin   int64  `json:"last_login"`
}

// EnqueueRequest represents a matchmaking enqueue request
type EnqueueRequest struct {
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name"`
}

// EnqueueResponse represents a matchmaking enqueue response
type EnqueueResponse struct {
	Success bool   `json:"success"`
	QueueID string `json:"queue_id"`
	Message string `json:"message"`
}

// QueueStatusResponse represents a queue status response
type QueueStatusResponse struct {
	Status        string `json:"status"`
	QueuePosition int32  `json:"queue_position"`
}

// MakeMoveRequest represents a game move request
type MakeMoveRequest struct {
	PlayerID string `json:"player_id"`
	PitIndex uint32 `json:"pit_index"`
}

// MakeMoveResponse represents a game move response
type MakeMoveResponse struct {
	Success bool        `json:"success"`
	Result  interface{} `json:"result,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Register registers a new user account
func (c *APIClient) Register(username, password string) (*RegisterResponse, error) {
	req := RegisterRequest{
		Username: username,
		Password: password,
	}

	resp, err := c.makeRequest("POST", "/api/v1/auth/register", req, false)
	if err != nil {
		return nil, err
	}

	var result RegisterResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Login authenticates a user
func (c *APIClient) Login(username, password string) (*LoginResponse, error) {
	req := LoginRequest{
		Username: username,
		Password: password,
	}

	resp, err := c.makeRequest("POST", "/api/v1/auth/login", req, false)
	if err != nil {
		return nil, err
	}

	var result LoginResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Enqueue adds a player to the matchmaking queue
func (c *APIClient) Enqueue(playerID, playerName string) (*EnqueueResponse, error) {
	req := EnqueueRequest{
		PlayerID:   playerID,
		PlayerName: playerName,
	}

	resp, err := c.makeRequest("POST", "/api/v1/matchmaking/enqueue", req, true)
	if err != nil {
		return nil, err
	}

	var result EnqueueResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CancelQueue removes a player from the matchmaking queue
func (c *APIClient) CancelQueue(playerID string) error {
	_, err := c.makeRequest("DELETE", fmt.Sprintf("/api/v1/matchmaking/queue/%s", playerID), nil, true)
	return err
}

// GetQueueStatus gets the current queue status for a player
func (c *APIClient) GetQueueStatus(playerID string) (*QueueStatusResponse, error) {
	resp, err := c.makeRequest("GET", fmt.Sprintf("/api/v1/matchmaking/queue/%s/status", playerID), nil, true)
	if err != nil {
		return nil, err
	}

	var result QueueStatusResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// MakeMove makes a move in a game
func (c *APIClient) MakeMove(gameID, playerID string, pitIndex uint32) (*MakeMoveResponse, error) {
	req := MakeMoveRequest{
		PlayerID: playerID,
		PitIndex: pitIndex,
	}

	resp, err := c.makeRequest("POST", fmt.Sprintf("/api/v1/games/%s/move", gameID), req, true)
	if err != nil {
		return nil, err
	}

	var result MakeMoveResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// TestConnection tests if the server is reachable
func (c *APIClient) TestConnection() error {
	_, err := c.makeRequest("GET", "/health", nil, false)
	return err
}

// makeRequest makes an HTTP request to the API
func (c *APIClient) makeRequest(method, endpoint string, body interface{}, requireAuth bool) ([]byte, error) {
	url := c.baseURL + endpoint

	var reqBody io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(bodyBytes)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if requireAuth && c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(responseBody))
	}

	return responseBody, nil
}
