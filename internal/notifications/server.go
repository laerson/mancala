package notifications

import (
	"log"

	"github.com/laerson/mancala/internal/auth"
	notificationspb "github.com/laerson/mancala/proto/notifications"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements the Notifications gRPC service
type Server struct {
	notificationspb.UnimplementedNotificationsServer
	clientManager   *ClientManager
	eventSubscriber *EventSubscriber
}

// NewServer creates a new notification server
func NewServer(redisAddr string) *Server {
	clientManager := NewClientManager()
	eventSubscriber := NewEventSubscriber(redisAddr, clientManager)

	server := &Server{
		clientManager:   clientManager,
		eventSubscriber: eventSubscriber,
	}

	// Start the event subscriber
	if err := eventSubscriber.Start(); err != nil {
		log.Printf("Failed to start event subscriber: %v", err)
	}

	return server
}

// Subscribe handles client subscription requests
func (s *Server) Subscribe(req *notificationspb.SubscribeRequest, stream notificationspb.Notifications_SubscribeServer) error {
	// Validate that the authenticated user owns this player ID
	if err := auth.ValidatePlayerOwnership(stream.Context(), req.PlayerId); err != nil {
		return status.Errorf(codes.Unauthenticated, "unauthorized: player ID does not match authenticated user")
	}

	if req.PlayerId == "" {
		return status.Errorf(codes.InvalidArgument, "player_id is required")
	}

	log.Printf("Player %s subscribing to notifications", req.PlayerId)

	// Add client to manager
	s.clientManager.AddClient(req.PlayerId, stream)

	// Keep the connection alive
	<-stream.Context().Done()

	// Clean up when client disconnects
	s.clientManager.RemoveClient(req.PlayerId)
	log.Printf("Player %s unsubscribed from notifications", req.PlayerId)

	return nil
}

// Stop gracefully shuts down the notification server
func (s *Server) Stop() {
	s.eventSubscriber.Stop()
}

// NotifyMatchFound is a helper method for external services to notify about matches
func (s *Server) NotifyMatchFound(gameID string, player1ID, player2ID string) {
	// Add game participants for tracking
	s.clientManager.AddGameParticipants(gameID, []string{player1ID, player2ID})
}

// GetConnectedClientsCount returns the number of connected clients (for monitoring)
func (s *Server) GetConnectedClientsCount() int {
	return s.clientManager.GetConnectedClientsCount()
}

// IsPlayerConnected checks if a player is currently connected (for other services)
func (s *Server) IsPlayerConnected(playerID string) bool {
	return s.clientManager.IsPlayerConnected(playerID)
}
