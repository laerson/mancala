package games

import (
	"context"
	"encoding/json"
	"fmt"

	gamespb "github.com/laerson/mancala/proto/games"
	"github.com/redis/go-redis/v9"
)

type Storage interface {
	SaveGame(ctx context.Context, game *gamespb.Game) error
	GetGame(ctx context.Context, gameID string) (*gamespb.Game, error)
	DeleteGame(ctx context.Context, gameID string) error
}

type RedisStorage struct {
	client *redis.Client
}

func NewRedisStorage(addr, password string, db int) *RedisStorage {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisStorage{
		client: rdb,
	}
}

func (r *RedisStorage) SaveGame(ctx context.Context, game *gamespb.Game) error {
	gameJSON, err := json.Marshal(game)
	if err != nil {
		return fmt.Errorf("failed to marshal game: %w", err)
	}

	err = r.client.Set(ctx, gameKey(game.Id), gameJSON, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to save game to redis: %w", err)
	}

	return nil
}

func (r *RedisStorage) GetGame(ctx context.Context, gameID string) (*gamespb.Game, error) {
	gameJSON, err := r.client.Get(ctx, gameKey(gameID)).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("game not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game from redis: %w", err)
	}

	var game gamespb.Game
	err = json.Unmarshal([]byte(gameJSON), &game)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal game: %w", err)
	}

	return &game, nil
}

func (r *RedisStorage) DeleteGame(ctx context.Context, gameID string) error {
	err := r.client.Del(ctx, gameKey(gameID)).Err()
	if err != nil {
		return fmt.Errorf("failed to delete game from redis: %w", err)
	}

	return nil
}

func gameKey(gameID string) string {
	return fmt.Sprintf("game:%s", gameID)
}
