package games

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/redis"
)

func setupRedisContainer(t *testing.T) (*redis.RedisContainer, *RedisStorage) {
	ctx := context.Background()

	redisContainer, err := redis.Run(ctx, "redis:7-alpine")
	if err != nil {
		t.Fatalf("failed to start redis container: %v", err)
	}

	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(redisContainer); err != nil {
			t.Logf("failed to terminate redis container: %v", err)
		}
	})

	host, err := redisContainer.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get redis host: %v", err)
	}

	port, err := redisContainer.MappedPort(ctx, "6379")
	if err != nil {
		t.Fatalf("failed to get redis port: %v", err)
	}

	addr := host + ":" + port.Port()
	storage := NewRedisStorage(addr, "", 0)

	return redisContainer, storage
}

func TestRedisStorage_SaveGame(t *testing.T) {
	_, storage := setupRedisContainer(t)
	ctx := context.Background()

	game := NewGame("player1", "player2")

	err := storage.SaveGame(ctx, game)
	if err != nil {
		t.Errorf("SaveGame() error = %v, want nil", err)
	}
}

func TestRedisStorage_GetGame(t *testing.T) {
	_, storage := setupRedisContainer(t)
	ctx := context.Background()

	originalGame := NewGame("player1", "player2")
	err := storage.SaveGame(ctx, originalGame)
	if err != nil {
		t.Fatalf("SaveGame() error = %v", err)
	}

	retrievedGame, err := storage.GetGame(ctx, originalGame.Id)
	if err != nil {
		t.Errorf("GetGame() error = %v, want nil", err)
	}

	if retrievedGame.Id != originalGame.Id {
		t.Errorf("GetGame() ID = %v, want %v", retrievedGame.Id, originalGame.Id)
	}
	if retrievedGame.Player1Id != originalGame.Player1Id {
		t.Errorf("GetGame() Player1Id = %v, want %v", retrievedGame.Player1Id, originalGame.Player1Id)
	}
	if retrievedGame.Player2Id != originalGame.Player2Id {
		t.Errorf("GetGame() Player2Id = %v, want %v", retrievedGame.Player2Id, originalGame.Player2Id)
	}
	if retrievedGame.State.CurrentPlayer != originalGame.State.CurrentPlayer {
		t.Errorf("GetGame() CurrentPlayer = %v, want %v", retrievedGame.State.CurrentPlayer, originalGame.State.CurrentPlayer)
	}
}

func TestRedisStorage_GetGame_NotFound(t *testing.T) {
	_, storage := setupRedisContainer(t)
	ctx := context.Background()

	_, err := storage.GetGame(ctx, "nonexistent-game-id")
	if err == nil {
		t.Error("GetGame() error = nil, want error for nonexistent game")
	}
	if err.Error() != "game not found" {
		t.Errorf("GetGame() error = %v, want 'game not found'", err)
	}
}

func TestRedisStorage_DeleteGame(t *testing.T) {
	_, storage := setupRedisContainer(t)
	ctx := context.Background()

	game := NewGame("player1", "player2")
	err := storage.SaveGame(ctx, game)
	if err != nil {
		t.Fatalf("SaveGame() error = %v", err)
	}

	err = storage.DeleteGame(ctx, game.Id)
	if err != nil {
		t.Errorf("DeleteGame() error = %v, want nil", err)
	}

	_, err = storage.GetGame(ctx, game.Id)
	if err == nil {
		t.Error("GetGame() after delete should return error")
	}
}

func TestRedisStorage_SaveAndRetrieveComplexGame(t *testing.T) {
	_, storage := setupRedisContainer(t)
	ctx := context.Background()

	game := NewGame("player1", "player2")
	game.State.Board.Pits = []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14}

	err := storage.SaveGame(ctx, game)
	if err != nil {
		t.Fatalf("SaveGame() error = %v", err)
	}

	retrievedGame, err := storage.GetGame(ctx, game.Id)
	if err != nil {
		t.Fatalf("GetGame() error = %v", err)
	}

	if len(retrievedGame.State.Board.Pits) != len(game.State.Board.Pits) {
		t.Errorf("Board pits length = %v, want %v", len(retrievedGame.State.Board.Pits), len(game.State.Board.Pits))
	}

	for i, expected := range game.State.Board.Pits {
		if retrievedGame.State.Board.Pits[i] != expected {
			t.Errorf("Pit %d = %v, want %v", i, retrievedGame.State.Board.Pits[i], expected)
		}
	}
}

func TestGameKey(t *testing.T) {
	gameID := "test-game-id"
	expected := "game:test-game-id"

	result := gameKey(gameID)
	if result != expected {
		t.Errorf("gameKey() = %v, want %v", result, expected)
	}
}
