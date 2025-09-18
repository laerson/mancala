package client

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds the client configuration
type Config struct {
	ServerURL    string `json:"server_url"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Username     string `json:"username"`
	UserID       string `json:"user_id"`
}

// ClientState manages the client's persistent state
type ClientState struct {
	config     *Config
	configPath string
}

// NewClientState creates a new client state manager
func NewClientState() (*ClientState, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(homeDir, ".mancala")
	configPath := filepath.Join(configDir, "config.json")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	state := &ClientState{
		configPath: configPath,
		config:     &Config{},
	}

	// Load existing config if it exists
	state.Load()

	return state, nil
}

// Load loads the configuration from disk
func (cs *ClientState) Load() error {
	data, err := os.ReadFile(cs.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Config file doesn't exist, use default config
			cs.config = &Config{}
			return nil
		}
		return err
	}

	return json.Unmarshal(data, cs.config)
}

// Save saves the configuration to disk
func (cs *ClientState) Save() error {
	data, err := json.MarshalIndent(cs.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cs.configPath, data, 0600)
}

// GetConfig returns a copy of the current configuration
func (cs *ClientState) GetConfig() Config {
	return *cs.config
}

// SetServerURL sets the server URL and saves the config
func (cs *ClientState) SetServerURL(url string) error {
	cs.config.ServerURL = url
	return cs.Save()
}

// SetAuth sets the authentication tokens and user info
func (cs *ClientState) SetAuth(accessToken, refreshToken, username, userID string) error {
	cs.config.AccessToken = accessToken
	cs.config.RefreshToken = refreshToken
	cs.config.Username = username
	cs.config.UserID = userID
	return cs.Save()
}

// ClearAuth clears the authentication information
func (cs *ClientState) ClearAuth() error {
	cs.config.AccessToken = ""
	cs.config.RefreshToken = ""
	cs.config.Username = ""
	cs.config.UserID = ""
	return cs.Save()
}

// IsConnected checks if the client is connected to a server
func (cs *ClientState) IsConnected() bool {
	return cs.config.ServerURL != ""
}

// IsLoggedIn checks if the client is logged in
func (cs *ClientState) IsLoggedIn() bool {
	return cs.config.AccessToken != ""
}
