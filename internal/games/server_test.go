package games

import (
	"context"
	"errors"
	"testing"

	enginepb "github.com/laerson/mancala/proto/engine"
	gamespb "github.com/laerson/mancala/proto/games"
)

func TestServer_Create(t *testing.T) {
	storage := NewMockStorage()
	engineClient := NewMockEngineClient()
	server := NewServer(storage, engineClient, "localhost:6379")

	tests := []struct {
		name    string
		request *gamespb.CreateGameRequest
		wantErr bool
	}{
		{
			name: "Valid create request",
			request: &gamespb.CreateGameRequest{
				Player1Id: "player1",
				Player2Id: "player2",
			},
			wantErr: false,
		},
		{
			name: "Empty player1 ID",
			request: &gamespb.CreateGameRequest{
				Player1Id: "",
				Player2Id: "player2",
			},
			wantErr: true,
		},
		{
			name: "Empty player2 ID",
			request: &gamespb.CreateGameRequest{
				Player1Id: "player1",
				Player2Id: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := server.Create(context.Background(), tt.request)

			if tt.wantErr {
				if err == nil {
					t.Error("Create() error = nil, want error")
				}
				return
			}

			if err != nil {
				t.Errorf("Create() error = %v, want nil", err)
				return
			}

			if response.Game == nil {
				t.Error("Create() response.Game = nil, want game")
				return
			}

			if response.Game.Player1Id != tt.request.Player1Id {
				t.Errorf("Create() Player1Id = %v, want %v", response.Game.Player1Id, tt.request.Player1Id)
			}
			if response.Game.Player2Id != tt.request.Player2Id {
				t.Errorf("Create() Player2Id = %v, want %v", response.Game.Player2Id, tt.request.Player2Id)
			}
		})
	}
}

func TestServer_Move_ValidMove(t *testing.T) {
	storage := NewMockStorage()
	engineClient := NewMockEngineClient()
	server := NewServer(storage, engineClient, "localhost:6379")

	game := NewGame("player1", "player2")
	storage.SaveGame(context.Background(), game)

	engineClient.SetMoveResponse(&enginepb.MoveResponse{
		Result: &enginepb.MoveResponse_MoveResult{
			MoveResult: &enginepb.MoveResult{
				Board: &enginepb.Board{
					Pits: []uint32{0, 5, 5, 5, 5, 4, 0, 4, 4, 4, 4, 4, 4, 0},
				},
				CurrentPlayer: enginepb.Player_PLAYER_TWO,
				IsFinished:    false,
				Winner:        enginepb.Winner_NO_WINNER,
			},
		},
	})

	request := &gamespb.MakeGameMoveRequest{
		PlayerId: "player1",
		GameId:   game.Id,
		PitIndex: 0,
	}

	response, err := server.Move(context.Background(), request)
	if err != nil {
		t.Errorf("Move() error = %v, want nil", err)
	}

	if response.Result == nil {
		t.Fatal("Move() response.Result = nil, want result")
	}

	moveResult, ok := response.Result.(*gamespb.MakeGameMoveResponse_MoveResult)
	if !ok {
		t.Fatal("Move() response should contain MoveResult")
	}

	if moveResult.MoveResult.CurrentPlayer != enginepb.Player_PLAYER_TWO {
		t.Errorf("Move() CurrentPlayer = %v, want %v", moveResult.MoveResult.CurrentPlayer, enginepb.Player_PLAYER_TWO)
	}
}

func TestServer_Move_GameNotFound(t *testing.T) {
	storage := NewMockStorage()
	engineClient := NewMockEngineClient()
	server := NewServer(storage, engineClient, "localhost:6379")

	request := &gamespb.MakeGameMoveRequest{
		PlayerId: "player1",
		GameId:   "nonexistent",
		PitIndex: 0,
	}

	response, err := server.Move(context.Background(), request)
	if err != nil {
		t.Errorf("Move() error = %v, want nil", err)
	}

	errorResult, ok := response.Result.(*gamespb.MakeGameMoveResponse_Error)
	if !ok {
		t.Fatal("Move() response should contain Error")
	}

	if errorResult.Error.Message != "game not found" {
		t.Errorf("Move() error message = %v, want 'game not found'", errorResult.Error.Message)
	}
}

func TestServer_Move_PlayerNotInGame(t *testing.T) {
	storage := NewMockStorage()
	engineClient := NewMockEngineClient()
	server := NewServer(storage, engineClient, "localhost:6379")

	game := NewGame("player1", "player2")
	storage.SaveGame(context.Background(), game)

	request := &gamespb.MakeGameMoveRequest{
		PlayerId: "player3",
		GameId:   game.Id,
		PitIndex: 0,
	}

	response, err := server.Move(context.Background(), request)
	if err != nil {
		t.Errorf("Move() error = %v, want nil", err)
	}

	errorResult, ok := response.Result.(*gamespb.MakeGameMoveResponse_Error)
	if !ok {
		t.Fatal("Move() response should contain Error")
	}

	if errorResult.Error.Message != "player is not part of this game" {
		t.Errorf("Move() error message = %v, want 'player is not part of this game'", errorResult.Error.Message)
	}
}

func TestServer_Move_NotPlayerTurn(t *testing.T) {
	storage := NewMockStorage()
	engineClient := NewMockEngineClient()
	server := NewServer(storage, engineClient, "localhost:6379")

	game := NewGame("player1", "player2")
	storage.SaveGame(context.Background(), game)

	request := &gamespb.MakeGameMoveRequest{
		PlayerId: "player2",
		GameId:   game.Id,
		PitIndex: 7,
	}

	response, err := server.Move(context.Background(), request)
	if err != nil {
		t.Errorf("Move() error = %v, want nil", err)
	}

	errorResult, ok := response.Result.(*gamespb.MakeGameMoveResponse_Error)
	if !ok {
		t.Fatal("Move() response should contain Error")
	}

	if errorResult.Error.Message != "it's not your turn" {
		t.Errorf("Move() error message = %v, want 'it's not your turn'", errorResult.Error.Message)
	}
}

func TestServer_Move_EngineError(t *testing.T) {
	storage := NewMockStorage()
	engineClient := NewMockEngineClient()
	server := NewServer(storage, engineClient, "localhost:6379")

	game := NewGame("player1", "player2")
	storage.SaveGame(context.Background(), game)

	engineClient.SetMoveError(errors.New("engine service unavailable"))

	request := &gamespb.MakeGameMoveRequest{
		PlayerId: "player1",
		GameId:   game.Id,
		PitIndex: 0,
	}

	response, err := server.Move(context.Background(), request)
	if err != nil {
		t.Errorf("Move() error = %v, want nil", err)
	}

	errorResult, ok := response.Result.(*gamespb.MakeGameMoveResponse_Error)
	if !ok {
		t.Fatal("Move() response should contain Error")
	}

	expectedMsg := "engine error: engine service unavailable"
	if errorResult.Error.Message != expectedMsg {
		t.Errorf("Move() error message = %v, want %v", errorResult.Error.Message, expectedMsg)
	}
}

func TestServer_Move_EngineReturnedError(t *testing.T) {
	storage := NewMockStorage()
	engineClient := NewMockEngineClient()
	server := NewServer(storage, engineClient, "localhost:6379")

	game := NewGame("player1", "player2")
	storage.SaveGame(context.Background(), game)

	engineClient.SetMoveResponse(&enginepb.MoveResponse{
		Result: &enginepb.MoveResponse_Error{
			Error: &enginepb.Error{Message: "invalid pit index"},
		},
	})

	request := &gamespb.MakeGameMoveRequest{
		PlayerId: "player1",
		GameId:   game.Id,
		PitIndex: 0,
	}

	response, err := server.Move(context.Background(), request)
	if err != nil {
		t.Errorf("Move() error = %v, want nil", err)
	}

	errorResult, ok := response.Result.(*gamespb.MakeGameMoveResponse_Error)
	if !ok {
		t.Fatal("Move() response should contain Error")
	}

	if errorResult.Error.Message != "invalid pit index" {
		t.Errorf("Move() error message = %v, want 'invalid pit index'", errorResult.Error.Message)
	}
}

func TestServer_Move_GameFinished(t *testing.T) {
	storage := NewMockStorage()
	engineClient := NewMockEngineClient()
	server := NewServer(storage, engineClient, "localhost:6379")

	game := NewGame("player1", "player2")
	storage.SaveGame(context.Background(), game)

	engineClient.SetMoveResponse(&enginepb.MoveResponse{
		Result: &enginepb.MoveResponse_MoveResult{
			MoveResult: &enginepb.MoveResult{
				Board: &enginepb.Board{
					Pits: []uint32{0, 0, 0, 0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 24},
				},
				CurrentPlayer: enginepb.Player_PLAYER_ONE,
				IsFinished:    true,
				Winner:        enginepb.Winner_DRAW,
			},
		},
	})

	request := &gamespb.MakeGameMoveRequest{
		PlayerId: "player1",
		GameId:   game.Id,
		PitIndex: 0,
	}

	response, err := server.Move(context.Background(), request)
	if err != nil {
		t.Errorf("Move() error = %v, want nil", err)
	}

	moveResult, ok := response.Result.(*gamespb.MakeGameMoveResponse_MoveResult)
	if !ok {
		t.Fatal("Move() response should contain MoveResult")
	}

	if !moveResult.MoveResult.IsFinished {
		t.Error("Move() IsFinished = false, want true")
	}
	if moveResult.MoveResult.Winner != enginepb.Winner_DRAW {
		t.Errorf("Move() Winner = %v, want %v", moveResult.MoveResult.Winner, enginepb.Winner_DRAW)
	}

	_, err = storage.GetGame(context.Background(), game.Id)
	if err == nil {
		t.Error("Game should be deleted after finishing")
	}
}

func TestServer_Move_EmptyPlayerID(t *testing.T) {
	storage := NewMockStorage()
	engineClient := NewMockEngineClient()
	server := NewServer(storage, engineClient, "localhost:6379")

	request := &gamespb.MakeGameMoveRequest{
		PlayerId: "",
		GameId:   "some-game",
		PitIndex: 0,
	}

	response, err := server.Move(context.Background(), request)
	if err != nil {
		t.Errorf("Move() error = %v, want nil", err)
	}

	errorResult, ok := response.Result.(*gamespb.MakeGameMoveResponse_Error)
	if !ok {
		t.Fatal("Move() response should contain Error")
	}

	if errorResult.Error.Message != "player ID and game ID are required" {
		t.Errorf("Move() error message = %v, want 'player ID and game ID are required'", errorResult.Error.Message)
	}
}

func TestServer_Move_EmptyGameID(t *testing.T) {
	storage := NewMockStorage()
	engineClient := NewMockEngineClient()
	server := NewServer(storage, engineClient, "localhost:6379")

	request := &gamespb.MakeGameMoveRequest{
		PlayerId: "player1",
		GameId:   "",
		PitIndex: 0,
	}

	response, err := server.Move(context.Background(), request)
	if err != nil {
		t.Errorf("Move() error = %v, want nil", err)
	}

	errorResult, ok := response.Result.(*gamespb.MakeGameMoveResponse_Error)
	if !ok {
		t.Fatal("Move() response should contain Error")
	}

	if errorResult.Error.Message != "player ID and game ID are required" {
		t.Errorf("Move() error message = %v, want 'player ID and game ID are required'", errorResult.Error.Message)
	}
}
