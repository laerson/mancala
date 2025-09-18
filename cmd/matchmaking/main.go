package main

import (
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/laerson/mancala/internal/auth"
	"github.com/laerson/mancala/internal/matchmaking"
	authpb "github.com/laerson/mancala/proto/auth"
	gamespb "github.com/laerson/mancala/proto/games"
	matchmakingpb "github.com/laerson/mancala/proto/matchmaking"
)

func main() {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50054"
	}

	// Create gRPC connections
	gamesAddr := os.Getenv("GAMES_ADDR")
	if gamesAddr == "" {
		gamesAddr = "games:50052"
	}

	authAddr := os.Getenv("AUTH_ADDR")
	if authAddr == "" {
		authAddr = "auth:50055"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "mancala-jwt-secret-key-change-in-production"
		log.Println("Warning: Using default JWT secret. Set JWT_SECRET environment variable in production!")
	}

	// Connect to Games service
	gamesConn, err := grpc.NewClient(gamesAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to games service: %v", err)
	}
	defer gamesConn.Close()

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

	// Create server
	server := matchmaking.NewServer(
		gamespb.NewGamesClient(gamesConn),
		redisAddr,
	)

	// Create auth interceptor
	authInterceptor := auth.NewAuthInterceptor(authClient, jwtSecret)

	// Start gRPC server
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.UnaryInterceptor()),
		grpc.StreamInterceptor(authInterceptor.StreamInterceptor()),
	)
	matchmakingpb.RegisterMatchmakingServer(grpcServer, server)

	log.Printf("Matchmaking service listening on port %s", port)
	log.Printf("Connected to Games service at %s", gamesAddr)
	log.Printf("Connected to Auth service at %s", authAddr)
	log.Printf("Connected to Redis at %s", redisAddr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
