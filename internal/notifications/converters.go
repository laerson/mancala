package notifications

import (
	"github.com/laerson/mancala/internal/events"
	enginepb "github.com/laerson/mancala/proto/engine"
	notificationspb "github.com/laerson/mancala/proto/notifications"
)

// createMatchFoundNotification creates a match found notification from event data
func createMatchFoundNotification(event events.Event, data events.MatchFoundData) *notificationspb.Notification {
	return &notificationspb.Notification{
		Id:        event.ID,
		Type:      notificationspb.NotificationType_NOTIFICATION_TYPE_MATCH_FOUND,
		GameId:    event.GameID,
		Timestamp: event.Timestamp,
		Data: &notificationspb.Notification_MatchFound{
			MatchFound: &notificationspb.MatchFoundNotification{
				MatchId:     data.MatchID,
				Player1Id:   data.Player1ID,
				Player1Name: data.Player1Name,
				Player2Id:   data.Player2ID,
				Player2Name: data.Player2Name,
			},
		},
	}
}

// createMoveMadeNotification creates a move made notification from event data
func createMoveMadeNotification(event events.Event, data events.MoveMadeData) *notificationspb.Notification {
	return &notificationspb.Notification{
		Id:        event.ID,
		Type:      notificationspb.NotificationType_NOTIFICATION_TYPE_MOVE_MADE,
		GameId:    event.GameID,
		Timestamp: event.Timestamp,
		Data: &notificationspb.Notification_MoveMade{
			MoveMade: &notificationspb.MoveMadeNotification{
				PlayerId:   data.PlayerID,
				PitIndex:   data.PitIndex,
				GameState:  convertMapToGameState(data.GameState),
				MoveResult: convertMapToMoveResult(data.MoveResult),
			},
		},
	}
}

// createGameOverNotification creates a game over notification from event data
func createGameOverNotification(event events.Event, data events.GameOverData) *notificationspb.Notification {
	return &notificationspb.Notification{
		Id:        event.ID,
		Type:      notificationspb.NotificationType_NOTIFICATION_TYPE_GAME_OVER,
		GameId:    event.GameID,
		Timestamp: event.Timestamp,
		Data: &notificationspb.Notification_GameOver{
			GameOver: &notificationspb.GameOverNotification{
				FinalState: convertMapToGameState(data.FinalState),
				WinnerId:   data.WinnerID,
				IsDraw:     data.IsDraw,
			},
		},
	}
}

// convertMapToGameState converts a map to a GameState proto message
func convertMapToGameState(stateMap map[string]interface{}) *enginepb.GameState {
	if stateMap == nil {
		return nil
	}

	gameState := &enginepb.GameState{}

	// Convert board
	if boardInterface, exists := stateMap["board"]; exists {
		if boardSlice, ok := boardInterface.([]interface{}); ok {
			board := &enginepb.Board{
				Pits: make([]uint32, len(boardSlice)),
			}
			for i, pit := range boardSlice {
				if pitValue, ok := pit.(float64); ok {
					board.Pits[i] = uint32(pitValue)
				}
			}
			gameState.Board = board
		}
	}

	// Convert current player
	if currentPlayerInterface, exists := stateMap["current_player"]; exists {
		if currentPlayer, ok := currentPlayerInterface.(float64); ok {
			gameState.CurrentPlayer = enginepb.Player(currentPlayer)
		}
	}

	return gameState
}

// convertMapToMoveResult converts a map to a MoveResult proto message
func convertMapToMoveResult(resultMap map[string]interface{}) *enginepb.MoveResult {
	if resultMap == nil {
		return nil
	}

	moveResult := &enginepb.MoveResult{}

	// Convert board
	if boardInterface, exists := resultMap["board"]; exists {
		if boardSlice, ok := boardInterface.([]interface{}); ok {
			board := &enginepb.Board{
				Pits: make([]uint32, len(boardSlice)),
			}
			for i, pit := range boardSlice {
				if pitValue, ok := pit.(float64); ok {
					board.Pits[i] = uint32(pitValue)
				}
			}
			moveResult.Board = board
		}
	}

	// Convert current player
	if currentPlayerInterface, exists := resultMap["current_player"]; exists {
		if currentPlayer, ok := currentPlayerInterface.(float64); ok {
			moveResult.CurrentPlayer = enginepb.Player(currentPlayer)
		}
	}

	// Convert is finished
	if isFinishedInterface, exists := resultMap["is_finished"]; exists {
		if isFinished, ok := isFinishedInterface.(bool); ok {
			moveResult.IsFinished = isFinished
		}
	}

	// Convert winner
	if winnerInterface, exists := resultMap["winner"]; exists {
		if winner, ok := winnerInterface.(float64); ok {
			moveResult.Winner = enginepb.Winner(winner)
		}
	}

	return moveResult
}
