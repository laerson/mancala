package notifications

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/laerson/mancala/internal/events"
)

// EventSubscriber subscribes to Redis streams and distributes events to clients
type EventSubscriber struct {
	redisClient   *redis.Client
	clientManager *ClientManager
	consumerGroup string
	consumerName  string
	mu            sync.RWMutex
	running       bool
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewEventSubscriber creates a new event subscriber
func NewEventSubscriber(redisAddr string, clientManager *ClientManager) *EventSubscriber {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	ctx, cancel := context.WithCancel(context.Background())

	return &EventSubscriber{
		redisClient:   rdb,
		clientManager: clientManager,
		consumerGroup: "notifications-service",
		consumerName:  "notification-consumer-1",
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start begins consuming events from Redis streams
func (es *EventSubscriber) Start() error {
	es.mu.Lock()
	defer es.mu.Unlock()

	if es.running {
		return nil
	}

	// Create consumer group if it doesn't exist
	err := es.redisClient.XGroupCreateMkStream(es.ctx, events.EventsStreamKey, es.consumerGroup, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		log.Printf("Failed to create consumer group: %v", err)
		return err
	}

	es.running = true
	go es.consume()

	log.Printf("Event subscriber started, consuming from group: %s", es.consumerGroup)
	return nil
}

// Stop stops the event subscriber
func (es *EventSubscriber) Stop() {
	es.mu.Lock()
	defer es.mu.Unlock()

	if !es.running {
		return
	}

	es.cancel()
	es.running = false
	log.Println("Event subscriber stopped")
}

// consume continuously reads events from Redis streams
func (es *EventSubscriber) consume() {
	for {
		select {
		case <-es.ctx.Done():
			return
		default:
			// Read messages from the stream
			streams, err := es.redisClient.XReadGroup(es.ctx, &redis.XReadGroupArgs{
				Group:    es.consumerGroup,
				Consumer: es.consumerName,
				Streams:  []string{events.EventsStreamKey, ">"},
				Count:    10,
				Block:    1 * time.Second,
			}).Result()

			if err != nil {
				if err != redis.Nil {
					log.Printf("Error reading from stream: %v", err)
				}
				continue
			}

			// Process each stream
			for _, stream := range streams {
				for _, message := range stream.Messages {
					es.processMessage(message)

					// Acknowledge the message
					es.redisClient.XAck(es.ctx, events.EventsStreamKey, es.consumerGroup, message.ID)
				}
			}
		}
	}
}

// processMessage processes a single Redis stream message
func (es *EventSubscriber) processMessage(message redis.XMessage) {
	eventDataStr, ok := message.Values["data"].(string)
	if !ok {
		log.Printf("Invalid event data format in message %s", message.ID)
		return
	}

	var event events.Event
	if err := json.Unmarshal([]byte(eventDataStr), &event); err != nil {
		log.Printf("Failed to unmarshal event %s: %v", message.ID, err)
		return
	}

	log.Printf("Processing event %s (%s) for game %s", event.ID, event.Type, event.GameID)

	// Route event to appropriate handler
	switch event.Type {
	case events.EventTypeMatchFound:
		es.handleMatchFound(event)
	case events.EventTypeMoveMade:
		es.handleMoveMade(event)
	case events.EventTypeGameOver:
		es.handleGameOver(event)
	default:
		log.Printf("Unknown event type: %s", event.Type)
	}
}

// handleMatchFound processes match found events
func (es *EventSubscriber) handleMatchFound(event events.Event) {
	var data events.MatchFoundData
	if err := mapToStruct(event.Data, &data); err != nil {
		log.Printf("Failed to parse match found data: %v", err)
		return
	}

	// Notify both players about the match
	players := []string{data.Player1ID, data.Player2ID}
	notification := createMatchFoundNotification(event, data)

	for _, playerID := range players {
		es.clientManager.NotifyPlayer(playerID, notification)
	}
}

// handleMoveMade processes move made events
func (es *EventSubscriber) handleMoveMade(event events.Event) {
	var data events.MoveMadeData
	if err := mapToStruct(event.Data, &data); err != nil {
		log.Printf("Failed to parse move made data: %v", err)
		return
	}

	// Get game participants to notify the opponent
	gameParticipants := es.clientManager.GetGameParticipants(event.GameID)
	notification := createMoveMadeNotification(event, data)

	for _, playerID := range gameParticipants {
		// Don't notify the player who made the move
		if playerID != data.PlayerID {
			es.clientManager.NotifyPlayer(playerID, notification)
		}
	}
}

// handleGameOver processes game over events
func (es *EventSubscriber) handleGameOver(event events.Event) {
	var data events.GameOverData
	if err := mapToStruct(event.Data, &data); err != nil {
		log.Printf("Failed to parse game over data: %v", err)
		return
	}

	// Notify all game participants
	gameParticipants := es.clientManager.GetGameParticipants(event.GameID)
	notification := createGameOverNotification(event, data)

	for _, playerID := range gameParticipants {
		es.clientManager.NotifyPlayer(playerID, notification)
	}

	// Clean up game participants tracking
	es.clientManager.RemoveGameParticipants(event.GameID)
}

// Helper function to convert map to struct
func mapToStruct(input map[string]interface{}, output interface{}) error {
	bytes, err := json.Marshal(input)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, output)
}
