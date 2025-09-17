package main

import (
	"log"
	"net"

	"github.com/laerson/mancala/internal/engine"
	enginepb "github.com/laerson/mancala/proto/engine"
	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	enginepb.RegisterEngineServer(grpcServer, &engine.Server{})

	log.Println("Starting server on port 50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
