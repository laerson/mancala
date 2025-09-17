package events

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

const (
	EventsStreamKey = "mancala:events"
)

// Event types
type EventType string

const (
	EventTypeMoveMade   EventType = "MOVE_MADE"
	EventTypeGameOver   EventType = "GAME_OVER"
	EventTypeMatchFound EventType = "MATCH_FOUND"
)

// Base event structure
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	GameID    string                 `json:"game_id,omitempty"`
	Timestamp int64                  `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// Move made event data
type MoveMadeData struct {
	PlayerID   string                 `json:"player_id"`
	PitIndex   uint32                 `json:"pit_index"`
	GameState  map[string]interface{} `json:"game_state"`
	MoveResult map[string]interface{} `json:"move_result"`
}

// Game over event data
type GameOverData struct {
	FinalState map[string]interface{} `json:"final_state"`
	WinnerID   string                 `json:"winner_id,omitempty"`
	IsDraw     bool                   `json:"is_draw"`
}

// Match found event data
type MatchFoundData struct {
	MatchID     string `json:"match_id"`
	Player1ID   string `json:"player1_id"`
	Player1Name string `json:"player1_name"`
	Player2ID   string `json:"player2_id"`
	Player2Name string `json:"player2_name"`
}

// EventPublisher handles publishing events to Redis Streams
type EventPublisher struct {
	redisClient *redis.Client
}

// NewEventPublisher creates a new event publisher
func NewEventPublisher(redisAddr string) *EventPublisher {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Test connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Failed to connect to Redis for events: %v", err)
	}

	return &EventPublisher{
		redisClient: rdb,
	}
}

// PublishMoveMade publishes a move made event
func (ep *EventPublisher) PublishMoveMade(ctx context.Context, gameID, playerID string, pitIndex uint32, gameState, moveResult map[string]interface{}) error {
	data := MoveMadeData{
		PlayerID:   playerID,
		PitIndex:   pitIndex,
		GameState:  gameState,
		MoveResult: moveResult,
	}

	event := Event{
		ID:        uuid.New().String(),
		Type:      EventTypeMoveMade,
		GameID:    gameID,
		Timestamp: time.Now().Unix(),
		Data:      structToMap(data),
	}

	return ep.publishEvent(ctx, event)
}

// PublishGameOver publishes a game over event
func (ep *EventPublisher) PublishGameOver(ctx context.Context, gameID, winnerID string, isDraw bool, finalState map[string]interface{}) error {
	data := GameOverData{
		FinalState: finalState,
		WinnerID:   winnerID,
		IsDraw:     isDraw,
	}

	event := Event{
		ID:        uuid.New().String(),
		Type:      EventTypeGameOver,
		GameID:    gameID,
		Timestamp: time.Now().Unix(),
		Data:      structToMap(data),
	}

	return ep.publishEvent(ctx, event)
}

// PublishMatchFound publishes a match found event
func (ep *EventPublisher) PublishMatchFound(ctx context.Context, gameID, matchID, player1ID, player1Name, player2ID, player2Name string) error {
	data := MatchFoundData{
		MatchID:     matchID,
		Player1ID:   player1ID,
		Player1Name: player1Name,
		Player2ID:   player2ID,
		Player2Name: player2Name,
	}

	event := Event{
		ID:        uuid.New().String(),
		Type:      EventTypeMatchFound,
		GameID:    gameID,
		Timestamp: time.Now().Unix(),
		Data:      structToMap(data),
	}

	return ep.publishEvent(ctx, event)
}

// publishEvent publishes an event to Redis Stream
func (ep *EventPublisher) publishEvent(ctx context.Context, event Event) error {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	fields := map[string]interface{}{
		"event_id": event.ID,
		"type":     string(event.Type),
		"game_id":  event.GameID,
		"data":     string(eventJSON),
	}

	_, err = ep.redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: EventsStreamKey,
		Values: fields,
	}).Result()

	if err != nil {
		log.Printf("Failed to publish event %s: %v", event.ID, err)
		return err
	}

	log.Printf("Published event %s (%s) for game %s", event.ID, event.Type, event.GameID)
	return nil
}

// Helper function to convert struct to map
func structToMap(data interface{}) map[string]interface{} {
	bytes, _ := json.Marshal(data)
	var result map[string]interface{}
	json.Unmarshal(bytes, &result)
	return result
}
