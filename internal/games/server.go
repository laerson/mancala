package games

import (
	"context"
	"fmt"

	enginepb "github.com/laerson/mancala/proto/engine"
	gamespb "github.com/laerson/mancala/proto/games"
	"google.golang.org/grpc"
)

type EngineClient interface {
	Move(ctx context.Context, req *enginepb.MoveRequest, opts ...grpc.CallOption) (*enginepb.MoveResponse, error)
}

type Server struct {
	gamespb.UnimplementedGamesServer
	storage      Storage
	engineClient EngineClient
}

func NewServer(storage Storage, engineClient EngineClient) *Server {
	return &Server{
		storage:      storage,
		engineClient: engineClient,
	}
}

func (s *Server) Create(ctx context.Context, req *gamespb.CreateGameRequest) (*gamespb.CreateGameResponse, error) {
	if req.Player1Id == "" || req.Player2Id == "" {
		return nil, fmt.Errorf("both player IDs are required")
	}

	game := NewGame(req.Player1Id, req.Player2Id)

	err := s.storage.SaveGame(ctx, game)
	if err != nil {
		return nil, fmt.Errorf("failed to save game: %w", err)
	}

	return &gamespb.CreateGameResponse{
		Game: game,
	}, nil
}

func (s *Server) Move(ctx context.Context, req *gamespb.MakeGameMoveRequest) (*gamespb.MakeGameMoveResponse, error) {
	if req.PlayerId == "" || req.GameId == "" {
		return &gamespb.MakeGameMoveResponse{
			Result: &gamespb.MakeGameMoveResponse_Error{
				Error: &gamespb.Error{Message: "player ID and game ID are required"},
			},
		}, nil
	}

	game, err := s.storage.GetGame(ctx, req.GameId)
	if err != nil {
		return &gamespb.MakeGameMoveResponse{
			Result: &gamespb.MakeGameMoveResponse_Error{
				Error: &gamespb.Error{Message: "game not found"},
			},
		}, nil
	}

	if !IsPlayerInGame(game, req.PlayerId) {
		return &gamespb.MakeGameMoveResponse{
			Result: &gamespb.MakeGameMoveResponse_Error{
				Error: &gamespb.Error{Message: "player is not part of this game"},
			},
		}, nil
	}

	currentPlayer := GetPlayerFromID(req.PlayerId, game)
	if game.State.CurrentPlayer != currentPlayer {
		return &gamespb.MakeGameMoveResponse{
			Result: &gamespb.MakeGameMoveResponse_Error{
				Error: &gamespb.Error{Message: "it's not your turn"},
			},
		}, nil
	}

	moveRequest := &enginepb.MoveRequest{
		GameState: game.State,
		PitIndex:  req.PitIndex,
	}

	moveResponse, err := s.engineClient.Move(ctx, moveRequest)
	if err != nil {
		return &gamespb.MakeGameMoveResponse{
			Result: &gamespb.MakeGameMoveResponse_Error{
				Error: &gamespb.Error{Message: fmt.Sprintf("engine error: %v", err)},
			},
		}, nil
	}

	switch result := moveResponse.Result.(type) {
	case *enginepb.MoveResponse_Error:
		return &gamespb.MakeGameMoveResponse{
			Result: &gamespb.MakeGameMoveResponse_Error{
				Error: &gamespb.Error{Message: result.Error.Message},
			},
		}, nil
	case *enginepb.MoveResponse_MoveResult:
		game.State.Board = result.MoveResult.Board
		game.State.CurrentPlayer = result.MoveResult.CurrentPlayer

		if result.MoveResult.IsFinished {
			err = s.storage.DeleteGame(ctx, req.GameId)
			if err != nil {
				return &gamespb.MakeGameMoveResponse{
					Result: &gamespb.MakeGameMoveResponse_Error{
						Error: &gamespb.Error{Message: "failed to clean up finished game"},
					},
				}, nil
			}
		} else {
			err = s.storage.SaveGame(ctx, game)
			if err != nil {
				return &gamespb.MakeGameMoveResponse{
					Result: &gamespb.MakeGameMoveResponse_Error{
						Error: &gamespb.Error{Message: "failed to save game state"},
					},
				}, nil
			}
		}

		return &gamespb.MakeGameMoveResponse{
			Result: &gamespb.MakeGameMoveResponse_MoveResult{
				MoveResult: result.MoveResult,
			},
		}, nil
	default:
		return &gamespb.MakeGameMoveResponse{
			Result: &gamespb.MakeGameMoveResponse_Error{
				Error: &gamespb.Error{Message: "unexpected engine response"},
			},
		}, nil
	}
}
