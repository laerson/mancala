package games

import (
	"testing"

	enginepb "github.com/laerson/mancala/proto/engine"
)

func TestNewGame(t *testing.T) {
	player1ID := "player1"
	player2ID := "player2"

	game := NewGame(player1ID, player2ID)

	if game.Id == "" {
		t.Error("Game ID should not be empty")
	}
	if game.Player1Id != player1ID {
		t.Errorf("Player1 ID should be %q, got %q", player1ID, game.Player1Id)
	}
	if game.Player2Id != player2ID {
		t.Errorf("Player2 ID should be %q, got %q", player2ID, game.Player2Id)
	}

	if game.State == nil {
		t.Fatal("Game state should not be nil")
	}
	if game.State.Board == nil {
		t.Fatal("Board should not be nil")
	}
	if game.State.CurrentPlayer != enginepb.Player_PLAYER_ONE {
		t.Errorf("Current player should be PLAYER_ONE, got %v", game.State.CurrentPlayer)
	}

	expectedPits := []uint32{4, 4, 4, 4, 4, 4, 0, 4, 4, 4, 4, 4, 4, 0}
	if len(game.State.Board.Pits) != len(expectedPits) {
		t.Errorf("Board should have %d pits, got %d", len(expectedPits), len(game.State.Board.Pits))
	}
	for i, expected := range expectedPits {
		if game.State.Board.Pits[i] != expected {
			t.Errorf("Pit %d should have %d seeds, got %d", i, expected, game.State.Board.Pits[i])
		}
	}
}

func TestGenerateGameID(t *testing.T) {
	id1 := generateGameID()
	id2 := generateGameID()

	if id1 == "" {
		t.Error("Game ID should not be empty")
	}
	if id2 == "" {
		t.Error("Game ID should not be empty")
	}
	if id1 == id2 {
		t.Error("Game IDs should be unique")
	}
	if len(id1) != 32 {
		t.Errorf("Game ID should be 32 characters long, got %d", len(id1))
	}
}

func TestIsPlayerInGame(t *testing.T) {
	game := NewGame("player1", "player2")

	tests := []struct {
		name     string
		playerID string
		want     bool
	}{
		{
			name:     "Player1 is in game",
			playerID: "player1",
			want:     true,
		},
		{
			name:     "Player2 is in game",
			playerID: "player2",
			want:     true,
		},
		{
			name:     "Player not in game",
			playerID: "player3",
			want:     false,
		},
		{
			name:     "Empty player ID",
			playerID: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsPlayerInGame(game, tt.playerID)
			if got != tt.want {
				t.Errorf("IsPlayerInGame() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPlayerFromID(t *testing.T) {
	game := NewGame("player1", "player2")

	tests := []struct {
		name     string
		playerID string
		want     enginepb.Player
	}{
		{
			name:     "Player1 returns PLAYER_ONE",
			playerID: "player1",
			want:     enginepb.Player_PLAYER_ONE,
		},
		{
			name:     "Player2 returns PLAYER_TWO",
			playerID: "player2",
			want:     enginepb.Player_PLAYER_TWO,
		},
		{
			name:     "Unknown player returns PLAYER_TWO (default)",
			playerID: "player3",
			want:     enginepb.Player_PLAYER_TWO,
		},
		{
			name:     "Empty player ID returns PLAYER_TWO (default)",
			playerID: "",
			want:     enginepb.Player_PLAYER_TWO,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetPlayerFromID(tt.playerID, game)
			if got != tt.want {
				t.Errorf("GetPlayerFromID() = %v, want %v", got, tt.want)
			}
		})
	}
}
