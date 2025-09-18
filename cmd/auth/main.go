package main

import (
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	"github.com/laerson/mancala/internal/auth"
	authpb "github.com/laerson/mancala/proto/auth"
)

func main() {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50055"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@postgres:5432/mancala?sslmode=disable"
		log.Println("Warning: Using default database URL. Set DATABASE_URL environment variable in production!")
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

	// Create server
	server, err := auth.NewServer(dbURL, redisAddr, jwtSecret)
	if err != nil {
		log.Fatalf("Failed to create auth server: %v", err)
	}

	// Start gRPC server
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	authpb.RegisterAuthServer(grpcServer, server)

	log.Printf("Auth service listening on port %s", port)
	log.Printf("Connected to PostgreSQL at %s", dbURL)
	log.Printf("Connected to Redis at %s", redisAddr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
