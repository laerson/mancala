package gateway

import (
	"net/http"

	"github.com/gin-gonic/gin"
	gamespb "github.com/laerson/mancala/proto/games"
)

// GamesHandlers handles game related endpoints
type GamesHandlers struct {
	clients *ServiceClients
}

// NewGamesHandlers creates new games handlers
func NewGamesHandlers(clients *ServiceClients) *GamesHandlers {
	return &GamesHandlers{clients: clients}
}

// CreateGameRequest represents a game creation request
type CreateGameRequest struct {
	Player1ID string `json:"player1_id" binding:"required"`
	Player2ID string `json:"player2_id" binding:"required"`
}

// MakeMoveRequest represents a move request
type MakeMoveRequest struct {
	PlayerID string `json:"player_id" binding:"required"`
	PitIndex uint32 `json:"pit_index" binding:"required"`
}

// CreateGame handles game creation
func (h *GamesHandlers) CreateGame(c *gin.Context) {
	var req CreateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call Games service
	resp, err := h.clients.Games.Create(addGRPCContext(c), &gamespb.CreateGameRequest{
		Player1Id: req.Player1ID,
		Player2Id: req.Player2ID,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create game"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"game": gin.H{
			"id":         resp.Game.Id,
			"player1_id": resp.Game.Player1Id,
			"player2_id": resp.Game.Player2Id,
			"state":      resp.Game.State,
		},
	})
}

// MakeMove handles game moves
func (h *GamesHandlers) MakeMove(c *gin.Context) {
	gameID := c.Param("game_id")
	if gameID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Game ID required"})
		return
	}

	var req MakeMoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call Games service
	resp, err := h.clients.Games.Move(addGRPCContext(c), &gamespb.MakeGameMoveRequest{
		PlayerId: req.PlayerID,
		GameId:   gameID,
		PitIndex: req.PitIndex,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to make move"})
		return
	}

	// Handle response based on result type
	switch result := resp.Result.(type) {
	case *gamespb.MakeGameMoveResponse_MoveResult:
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"result":  result.MoveResult,
		})
	case *gamespb.MakeGameMoveResponse_Error:
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   result.Error.Message,
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Unexpected response format",
		})
	}
}
