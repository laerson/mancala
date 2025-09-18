package notifications

import (
	"log"
	"sync"

	notificationspb "github.com/laerson/mancala/proto/notifications"
)

// ClientConnection represents a connected client
type ClientConnection struct {
	playerID string
	stream   notificationspb.Notifications_SubscribeServer
	done     chan bool
}

// ClientManager manages client connections and notifications
type ClientManager struct {
	mu               sync.RWMutex
	clients          map[string]*ClientConnection // playerID -> connection
	gameParticipants map[string][]string          // gameID -> []playerID
}

// NewClientManager creates a new client manager
func NewClientManager() *ClientManager {
	return &ClientManager{
		clients:          make(map[string]*ClientConnection),
		gameParticipants: make(map[string][]string),
	}
}

// AddClient adds a new client connection
func (cm *ClientManager) AddClient(playerID string, stream notificationspb.Notifications_SubscribeServer) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Close existing connection if any
	if existingClient, exists := cm.clients[playerID]; exists {
		close(existingClient.done)
	}

	client := &ClientConnection{
		playerID: playerID,
		stream:   stream,
		done:     make(chan bool),
	}

	cm.clients[playerID] = client
	log.Printf("Client connected: %s", playerID)
}

// RemoveClient removes a client connection
func (cm *ClientManager) RemoveClient(playerID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if client, exists := cm.clients[playerID]; exists {
		close(client.done)
		delete(cm.clients, playerID)
		log.Printf("Client disconnected: %s", playerID)
	}
}

// NotifyPlayer sends a notification to a specific player
func (cm *ClientManager) NotifyPlayer(playerID string, notification *notificationspb.Notification) {
	cm.mu.RLock()
	client, exists := cm.clients[playerID]
	cm.mu.RUnlock()

	if !exists {
		log.Printf("Player %s not connected, skipping notification", playerID)
		return
	}

	select {
	case <-client.done:
		// Client disconnected, remove it
		cm.RemoveClient(playerID)
		return
	default:
		// Send notification
		if err := client.stream.Send(notification); err != nil {
			log.Printf("Failed to send notification to player %s: %v", playerID, err)
			cm.RemoveClient(playerID)
		} else {
			log.Printf("Sent %s notification to player %s", notification.Type, playerID)
		}
	}
}

// AddGameParticipants adds players to a game for notification tracking
func (cm *ClientManager) AddGameParticipants(gameID string, playerIDs []string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.gameParticipants[gameID] = playerIDs
	log.Printf("Added game participants for game %s: %v", gameID, playerIDs)
}

// GetGameParticipants gets the players in a game
func (cm *ClientManager) GetGameParticipants(gameID string) []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if participants, exists := cm.gameParticipants[gameID]; exists {
		return participants
	}
	return []string{}
}

// RemoveGameParticipants removes game participants tracking
func (cm *ClientManager) RemoveGameParticipants(gameID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	delete(cm.gameParticipants, gameID)
	log.Printf("Removed game participants for game %s", gameID)
}

// GetConnectedClientsCount returns the number of connected clients
func (cm *ClientManager) GetConnectedClientsCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return len(cm.clients)
}

// IsPlayerConnected checks if a player is currently connected
func (cm *ClientManager) IsPlayerConnected(playerID string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	_, exists := cm.clients[playerID]
	return exists
}
