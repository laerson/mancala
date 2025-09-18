package gateway

import "time"

// ServiceConfig holds configuration for backend services
type ServiceConfig struct {
	AuthAddr          string
	GamesAddr         string
	MatchmakingAddr   string
	NotificationsAddr string
	EngineAddr        string
}

// GatewayConfig holds configuration for the API gateway
type GatewayConfig struct {
	Port         string
	Services     ServiceConfig
	JWTSecret    string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DefaultConfig returns a default gateway configuration
func DefaultConfig() *GatewayConfig {
	return &GatewayConfig{
		Port: "8080",
		Services: ServiceConfig{
			AuthAddr:          "auth:50055",
			GamesAddr:         "games:50052",
			MatchmakingAddr:   "matchmaking:50054",
			NotificationsAddr: "notifications:50056",
			EngineAddr:        "engine:50051",
		},
		JWTSecret:    "mancala-jwt-secret-key-change-in-production",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}
