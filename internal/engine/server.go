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

	board := req.GetGameState().GetBoard().GetPits()
	player := req.GetGameState().GetCurrentPlayer()
	pitIndex := req.GetPitIndex()
	winner := enginepb.Winner_NO_WINNER

	seeds := board[pitIndex]
	board[pitIndex] = 0

	// distribution logic
	for seeds > 0 {
		pitIndex = (pitIndex + 1) % 14
		if player == enginepb.Player_PLAYER_ONE && pitIndex == 13 ||
			player == enginepb.Player_PLAYER_TWO && pitIndex == 6 {
			pitIndex = (pitIndex + 1) % 14 // Skip opponent's store
		}
		board[pitIndex] += 1
		seeds -= 1
	}
	// Check Capture
	if board[pitIndex] == 1 && board[12-pitIndex] > 0 && isPlayablePit(pitIndex, player) {
		switch player {
		case enginepb.Player_PLAYER_ONE:
			board[6] += 1 + board[12-pitIndex]
			board[12-pitIndex] = 0
			board[pitIndex] = 0
		case enginepb.Player_PLAYER_TWO:
			board[7] += 1 + board[12-pitIndex]
			board[12-pitIndex] = 0
			board[pitIndex] = 0
		}
	}

	// Check if the game is finished
	isFinished := true
	switch player {
	case enginepb.Player_PLAYER_ONE:
		for _, v := range board[:6] {
			if v > 0 {
				isFinished = false
				break
			}
		}
	case enginepb.Player_PLAYER_TWO:
		for _, v := range board[7:13] {
			if v > 0 {
				isFinished = false
				break
			}
		}
	}
	if isFinished {
		// Move remaining seeds to store
		switch player {
		case enginepb.Player_PLAYER_ONE:
			for i, v := range board[7:13] {
				board[13] += v
				board[i+7] = 0
			}
		case enginepb.Player_PLAYER_TWO:
			for i, v := range board[0:6] {
				board[6] += v
				board[i] = 0
			}
		}
		// Get Winner
		if board[6] > board[13] {
			winner = enginepb.Winner_WINNER_PLAYER_ONE
		} else if board[13] > board[6] {
			winner = enginepb.Winner_WINNER_PLAYER_TWO
		} else {
			winner = enginepb.Winner_DRAW
		}
	}
	repeatTurn := pitIndex == 6 || pitIndex == 13
	if !repeatTurn {
		switch player {
		case enginepb.Player_PLAYER_ONE:
			player = enginepb.Player_PLAYER_TWO
		case enginepb.Player_PLAYER_TWO:
			player = enginepb.Player_PLAYER_ONE
		}
	}
	next := &enginepb.MoveResult{
		Board:         &enginepb.Board{Pits: board},
		CurrentPlayer: player,
		IsFinished:    isFinished,
		Winner:        winner,
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
