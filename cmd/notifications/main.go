package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/laerson/mancala/internal/auth"
	"github.com/laerson/mancala/internal/notifications"
	authpb "github.com/laerson/mancala/proto/auth"
	notificationspb "github.com/laerson/mancala/proto/notifications"
)

func main() {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50056"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
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

	// Create notification server
	notificationServer := notifications.NewServer(redisAddr)

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
	notificationspb.RegisterNotificationsServer(grpcServer, notificationServer)

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Shutting down notification service...")
		notificationServer.Stop()
		grpcServer.GracefulStop()
	}()

	log.Printf("Notification service listening on port %s", port)
	log.Printf("Connected to Redis at %s", redisAddr)
	log.Printf("Connected to Auth service at %s", authAddr)
	log.Printf("Ready to serve notifications")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
