package main

import (
	"context"
	"log"
	"net"

	enginepb "github.com/laerson/mancala/proto/engine"
	"google.golang.org/grpc"
)

type server struct {
	enginepb.UnimplementedEngineServer
}

// Move tries to apply a move to the given game state and returns either an error (in-message)
// or the updated game state. Only transport failures should be returned as Go errors.
func (s *server) Move(ctx context.Context, req *enginepb.MoveRequest) (*enginepb.MoveResponse, error) {
	// Validate Board
	// Board must have exactly 14 pits
	board := req.GetGameState().GetBoard().GetPits()
	if len(board) != 14 {
		return errResp("Board must have exactly 14 pits"), nil
	}
	// Validate PitIndex
	// PitIndex must be between 0 and 5 or 7 and 12
	pitIndex := req.GetPitIndex()
	player := req.GetGameState().GetCurrentPlayer()
	if !isPlayablePit(pitIndex, player) {
		return errResp("PitIndex must be between 0 and 5 or 7 and 12"), nil
	}

	// PitIndex cannot be empty
	if board[pitIndex] == 0 {
		return errResp("PitIndex cannot be empty"), nil
	}

	// TODO: Implement move logic
	// The current code only echoes the state back
	next := &enginepb.MoveResult{
		Board:         req.GetGameState().GetBoard(),
		CurrentPlayer: req.GetGameState().GetCurrentPlayer(),
		IsFinished:    false,
		Winner:        enginepb.Winner_NO_WINNER,
	}

	return &enginepb.MoveResponse{
		Result: &enginepb.MoveResponse_MoveResult{
			MoveResult: next,
		},
	}, nil

}

// errResp creates a MoveResponse containing an error message.
func errResp(msg string) *enginepb.MoveResponse {
	return &enginepb.MoveResponse{
		Result: &enginepb.MoveResponse_Error{
			Error: &enginepb.Error{Message: msg},
		},
	}
}

// isPlayablePit returns true if the given pit is playable by the given player.
func isPlayablePit(p uint32, player enginepb.Player) bool {
	return (p <= 5) && (player == enginepb.Player_PLAYER_ONE) || (p >= 7 && p <= 12) && (player == enginepb.Player_PLAYER_TWO)
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	enginepb.RegisterEngineServer(grpcServer, &server{})

	log.Println("Starting server on port 50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
