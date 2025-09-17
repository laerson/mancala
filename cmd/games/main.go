package main

import (
	"log"
	"net"

	"github.com/laerson/mancala/internal/games"
	enginepb "github.com/laerson/mancala/proto/engine"
	gamespb "github.com/laerson/mancala/proto/games"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	storage := games.NewRedisStorage("localhost:6379", "", 0)

	engineConn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to engine service: %v", err)
	}
	defer engineConn.Close()

	engineClient := enginepb.NewEngineClient(engineConn)

	gamesServer := games.NewServer(storage, engineClient)

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	gamespb.RegisterGamesServer(grpcServer, gamesServer)

	log.Println("Starting games service on port 50052")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
