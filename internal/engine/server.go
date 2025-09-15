package engine

import (
	"context"
	"fmt"

	enginepb "github.com/laerson/mancala/proto/engine"
)

var errInvalidBoard = fmt.Errorf("board must have exactly 14 pits")
var errInvalidPitIndex = fmt.Errorf("invalid pit index for current player")
var errEmptyPit = fmt.Errorf("pit cannot be empty")

type Server struct {
	enginepb.UnimplementedEngineServer
}

// Move tries to apply a move to the given game state and returns either an error (in-message)
// or the updated game state. Only transport failures should be returned as Go errors.
func (s *Server) Move(ctx context.Context, req *enginepb.MoveRequest) (*enginepb.MoveResponse, error) {

	err := validateMoveRequest(req)
	if err != nil {
		return errResp(err.Error()), nil
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

func validateMoveRequest(req *enginepb.MoveRequest) error {
	board := req.GetGameState().GetBoard().GetPits()
	if len(board) != 14 {
		return errInvalidBoard
	}

	pitIndex := req.GetPitIndex()
	player := req.GetGameState().GetCurrentPlayer()
	if !isPlayablePit(pitIndex, player) {
		return errInvalidPitIndex
	}

	if board[pitIndex] == 0 {
		return errEmptyPit
	}

	return nil
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
