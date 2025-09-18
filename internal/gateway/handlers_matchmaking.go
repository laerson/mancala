package gateway

import (
	"net/http"

	"github.com/gin-gonic/gin"
	matchmakingpb "github.com/laerson/mancala/proto/matchmaking"
)

// MatchmakingHandlers handles matchmaking related endpoints
type MatchmakingHandlers struct {
	clients *ServiceClients
}

// NewMatchmakingHandlers creates new matchmaking handlers
func NewMatchmakingHandlers(clients *ServiceClients) *MatchmakingHandlers {
	return &MatchmakingHandlers{clients: clients}
}

// EnqueueRequest represents a queue enrollment request
type EnqueueRequest struct {
	PlayerID   string `json:"player_id" binding:"required"`
	PlayerName string `json:"player_name" binding:"required"`
}

// Enqueue handles player queue enrollment
func (h *MatchmakingHandlers) Enqueue(c *gin.Context) {
	var req EnqueueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call Matchmaking service
	resp, err := h.clients.Matchmaking.Enqueue(addGRPCContext(c), &matchmakingpb.EnqueueRequest{
		Player: &matchmakingpb.Player{
			Id:   req.PlayerID,
			Name: req.PlayerName,
		},
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue player"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  resp.Success,
		"queue_id": resp.QueueId,
		"message":  resp.Message,
	})
}

// CancelQueue handles queue cancellation
func (h *MatchmakingHandlers) CancelQueue(c *gin.Context) {
	playerID := c.Param("player_id")
	if playerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Player ID required"})
		return
	}

	// Call Matchmaking service
	resp, err := h.clients.Matchmaking.CancelQueue(addGRPCContext(c), &matchmakingpb.CancelQueueRequest{
		PlayerId: playerID,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel queue"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": resp.Success,
		"message": resp.Message,
	})
}

// GetQueueStatus handles queue status requests
func (h *MatchmakingHandlers) GetQueueStatus(c *gin.Context) {
	playerID := c.Param("player_id")
	if playerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Player ID required"})
		return
	}

	// Call Matchmaking service
	resp, err := h.clients.Matchmaking.GetQueueStatus(addGRPCContext(c), &matchmakingpb.GetQueueStatusRequest{
		PlayerId: playerID,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get queue status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":         resp.Status.String(),
		"queue_position": resp.QueuePosition,
	})
}
