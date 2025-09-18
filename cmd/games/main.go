package main

import (
	"log"
	"net"
	"os"

	"github.com/laerson/mancala/internal/auth"
	"github.com/laerson/mancala/internal/games"
	authpb "github.com/laerson/mancala/proto/auth"
	enginepb "github.com/laerson/mancala/proto/engine"
	gamespb "github.com/laerson/mancala/proto/games"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Get configuration from environment variables
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	engineAddr := os.Getenv("ENGINE_ADDR")
	if engineAddr == "" {
		engineAddr = "localhost:50051"
	}

	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50052"
	}

	authAddr := os.Getenv("AUTH_ADDR")
	if authAddr == "" {
		authAddr = "auth:50055"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "mancala-jwt-secret-key-change-in-production"
		log.Println("Warning: Using default JWT secret. Set JWT_SECRET environment variable in production!")
	}

	storage := games.NewRedisStorage(redisAddr, "", 0)

	engineConn, err := grpc.NewClient(engineAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to engine service: %v", err)
	}
	defer engineConn.Close()

	engineClient := enginepb.NewEngineClient(engineConn)

	// Connect to Auth service
	authConn, err := grpc.NewClient(authAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Warning: Failed to connect to auth service: %v", err)
		log.Println("Auth service validation will be skipped")
	}
	var authClient authpb.AuthClient
	if authConn != nil {
		authClient = authpb.NewAuthClient(authConn)
		defer authConn.Close()
	}

	gamesServer := games.NewServer(storage, engineClient, redisAddr)

	// Create auth interceptor
	authInterceptor := auth.NewAuthInterceptor(authClient, jwtSecret)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.UnaryInterceptor()),
	)
	gamespb.RegisterGamesServer(grpcServer, gamesServer)

	log.Printf("Starting games service on port %s", port)
	log.Printf("Connected to Redis at %s for event publishing", redisAddr)
	log.Printf("Connected to Engine service at %s", engineAddr)
	log.Printf("Connected to Auth service at %s", authAddr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
