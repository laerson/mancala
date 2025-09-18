package games

import (
	"context"
	"fmt"

	"github.com/laerson/mancala/internal/events"
	enginepb "github.com/laerson/mancala/proto/engine"
	gamespb "github.com/laerson/mancala/proto/games"
	"google.golang.org/grpc"
)

type EngineClient interface {
	Move(ctx context.Context, req *enginepb.MoveRequest, opts ...grpc.CallOption) (*enginepb.MoveResponse, error)
}

type Server struct {
	gamespb.UnimplementedGamesServer
	storage        Storage
	engineClient   EngineClient
	eventPublisher *events.EventPublisher
}

func NewServer(storage Storage, engineClient EngineClient, redisAddr string) *Server {
	return &Server{
		storage:        storage,
		engineClient:   engineClient,
		eventPublisher: events.NewEventPublisher(redisAddr),
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

		// Publish MOVE_MADE event
		gameStateMap := gameStateToMap(game.State)
		moveResultMap := moveResultToMap(result.MoveResult)
		err = s.eventPublisher.PublishMoveMade(ctx, req.GameId, req.PlayerId, req.PitIndex, gameStateMap, moveResultMap)
		if err != nil {
			// Log error but don't fail the game operation
			fmt.Printf("Failed to publish move made event: %v", err)
		}

		if result.MoveResult.IsFinished {
			// Publish GAME_OVER event
			winnerID := determineWinner(result.MoveResult, game)
			isDraw := winnerID == ""
			err = s.eventPublisher.PublishGameOver(ctx, req.GameId, winnerID, isDraw, gameStateMap)
			if err != nil {
				// Log error but don't fail the game operation
				fmt.Printf("Failed to publish game over event: %v", err)
			}

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

// Helper function to convert GameState to map for event publishing
func gameStateToMap(state *enginepb.GameState) map[string]interface{} {
	if state == nil {
		return nil
	}

	boardSlice := make([]interface{}, len(state.Board.Pits))
	for i, v := range state.Board.Pits {
		boardSlice[i] = v
	}

	return map[string]interface{}{
		"board":          boardSlice,
		"current_player": int(state.CurrentPlayer),
	}
}

// Helper function to convert MoveResult to map for event publishing
func moveResultToMap(moveResult *enginepb.MoveResult) map[string]interface{} {
	if moveResult == nil {
		return nil
	}

	boardSlice := make([]interface{}, len(moveResult.Board.Pits))
	for i, v := range moveResult.Board.Pits {
		boardSlice[i] = v
	}

	return map[string]interface{}{
		"board":          boardSlice,
		"current_player": int(moveResult.CurrentPlayer),
		"is_finished":    moveResult.IsFinished,
		"winner":         int(moveResult.Winner),
	}
}

// Helper function to determine winner from MoveResult and game context
func determineWinner(moveResult *enginepb.MoveResult, game *gamespb.Game) string {
	if !moveResult.IsFinished {
		return ""
	}

	switch moveResult.Winner {
	case enginepb.Winner_WINNER_PLAYER_ONE:
		return game.Player1Id
	case enginepb.Winner_WINNER_PLAYER_TWO:
		return game.Player2Id
	default:
		// Draw or no winner
		return ""
	}
}
