package gateway

import (
	"context"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	notificationspb "github.com/laerson/mancala/proto/notifications"
)

// NotificationsHandlers handles notification related endpoints
type NotificationsHandlers struct {
	clients *ServiceClients
}

// NewNotificationsHandlers creates new notifications handlers
func NewNotificationsHandlers(clients *ServiceClients) *NotificationsHandlers {
	return &NotificationsHandlers{clients: clients}
}

// SubscribeToNotifications handles SSE connections for real-time notifications
func (h *NotificationsHandlers) SubscribeToNotifications(c *gin.Context) {
	playerID := c.Param("player_id")
	if playerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Player ID required"})
		return
	}

	// Set headers for Server-Sent Events
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Call Notifications service
	stream, err := h.clients.Notifications.Subscribe(addGRPCContext(c), &notificationspb.SubscribeRequest{
		PlayerId: playerID,
	})

	if err != nil {
		log.Printf("Failed to subscribe to notifications: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to subscribe to notifications"})
		return
	}

	// Create a context with cancellation for cleanup
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	// Start receiving notifications from gRPC stream
	go func() {
		defer cancel()
		for {
			notification, err := stream.Recv()
			if err == io.EOF {
				log.Printf("Notification stream ended for player %s", playerID)
				return
			}
			if err != nil {
				log.Printf("Error receiving notification for player %s: %v", playerID, err)
				return
			}

			// Format notification as SSE
			select {
			case <-ctx.Done():
				return
			default:
				// Send notification as SSE event
				data := formatNotificationForSSE(notification)
				c.SSEvent("notification", data)
				c.Writer.Flush()
			}
		}
	}()

	// Send initial connection confirmation
	c.SSEvent("connected", gin.H{"player_id": playerID, "timestamp": time.Now().Unix()})
	c.Writer.Flush()

	// Keep connection alive until client disconnects
	<-ctx.Done()
	log.Printf("Notification subscription ended for player %s", playerID)
}

// formatNotificationForSSE formats a notification for Server-Sent Events
func formatNotificationForSSE(notification *notificationspb.Notification) gin.H {
	data := gin.H{
		"id":        notification.Id,
		"type":      notification.Type.String(),
		"game_id":   notification.GameId,
		"timestamp": notification.Timestamp,
	}

	// Add specific data based on notification type
	switch notification.Type {
	case notificationspb.NotificationType_NOTIFICATION_TYPE_MATCH_FOUND:
		if matchFound := notification.GetMatchFound(); matchFound != nil {
			data["data"] = gin.H{
				"match_id":     matchFound.MatchId,
				"player1_id":   matchFound.Player1Id,
				"player1_name": matchFound.Player1Name,
				"player2_id":   matchFound.Player2Id,
				"player2_name": matchFound.Player2Name,
			}
		}
	case notificationspb.NotificationType_NOTIFICATION_TYPE_MOVE_MADE:
		if moveMade := notification.GetMoveMade(); moveMade != nil {
			data["data"] = gin.H{
				"player_id":   moveMade.PlayerId,
				"pit_index":   moveMade.PitIndex,
				"game_state":  moveMade.GameState,
				"move_result": moveMade.MoveResult,
			}
		}
	case notificationspb.NotificationType_NOTIFICATION_TYPE_GAME_OVER:
		if gameOver := notification.GetGameOver(); gameOver != nil {
			data["data"] = gin.H{
				"final_state": gameOver.FinalState,
				"winner_id":   gameOver.WinnerId,
				"is_draw":     gameOver.IsDraw,
			}
		}
	}

	return data
}
