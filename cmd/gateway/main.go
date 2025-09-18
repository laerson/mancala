package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/laerson/mancala/internal/gateway"
)

func main() {
	// Create configuration
	config := gateway.DefaultConfig()

	// Override with environment variables
	if port := os.Getenv("HTTP_PORT"); port != "" {
		config.Port = port
	}

	if authAddr := os.Getenv("AUTH_ADDR"); authAddr != "" {
		config.Services.AuthAddr = authAddr
	}

	if gamesAddr := os.Getenv("GAMES_ADDR"); gamesAddr != "" {
		config.Services.GamesAddr = gamesAddr
	}

	if matchmakingAddr := os.Getenv("MATCHMAKING_ADDR"); matchmakingAddr != "" {
		config.Services.MatchmakingAddr = matchmakingAddr
	}

	if notificationsAddr := os.Getenv("NOTIFICATIONS_ADDR"); notificationsAddr != "" {
		config.Services.NotificationsAddr = notificationsAddr
	}

	if engineAddr := os.Getenv("ENGINE_ADDR"); engineAddr != "" {
		config.Services.EngineAddr = engineAddr
	}

	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		config.JWTSecret = jwtSecret
	} else {
		log.Println("Warning: Using default JWT secret. Set JWT_SECRET environment variable in production!")
	}

	// Create and start server
	server, err := gateway.NewServer(config)
	if err != nil {
		log.Fatalf("Failed to create gateway server: %v", err)
	}

	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Printf("API Gateway listening on port %s", config.Port)
		log.Printf("Backend services:")
		log.Printf("  - Auth: %s", config.Services.AuthAddr)
		log.Printf("  - Games: %s", config.Services.GamesAddr)
		log.Printf("  - Matchmaking: %s", config.Services.MatchmakingAddr)
		log.Printf("  - Notifications: %s", config.Services.NotificationsAddr)
		log.Printf("  - Engine: %s", config.Services.EngineAddr)
		log.Printf("API Gateway ready to serve requests")

		if err := server.Start(); err != nil {
			log.Printf("Gateway server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-c
	log.Println("Received shutdown signal")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Gracefully shutdown the server
	if err := server.Stop(ctx); err != nil {
		log.Printf("Gateway server shutdown error: %v", err)
	}

	log.Println("Gateway server shut down complete")
}
