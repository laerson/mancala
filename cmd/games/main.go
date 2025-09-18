package main

import (
	"log"
	"net"
	"os"

	"github.com/laerson/mancala/internal/games"
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

	storage := games.NewRedisStorage(redisAddr, "", 0)

	engineConn, err := grpc.NewClient(engineAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to engine service: %v", err)
	}
	defer engineConn.Close()

	engineClient := enginepb.NewEngineClient(engineConn)

	gamesServer := games.NewServer(storage, engineClient, redisAddr)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	gamespb.RegisterGamesServer(grpcServer, gamesServer)

	log.Printf("Starting games service on port %s", port)
	log.Printf("Connected to Redis at %s for event publishing", redisAddr)
	log.Printf("Connected to Engine service at %s", engineAddr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
