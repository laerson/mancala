package main

import (
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/laerson/mancala/internal/matchmaking"
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

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}

	// Connect to Games service
	gamesConn, err := grpc.NewClient(gamesAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to games service: %v", err)
	}
	defer gamesConn.Close()

	// Create server
	server := matchmaking.NewServer(
		gamespb.NewGamesClient(gamesConn),
		redisAddr,
	)

	// Start gRPC server
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	matchmakingpb.RegisterMatchmakingServer(grpcServer, server)

	log.Printf("Matchmaking service listening on port %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
