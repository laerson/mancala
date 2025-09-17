package games

import (
	"crypto/rand"
	"encoding/hex"

	enginepb "github.com/laerson/mancala/proto/engine"
	gamespb "github.com/laerson/mancala/proto/games"
)

func NewGame(player1ID, player2ID string) *gamespb.Game {
	gameID := generateGameID()

	initialBoard := &enginepb.Board{
		Pits: []uint32{4, 4, 4, 4, 4, 4, 0, 4, 4, 4, 4, 4, 4, 0},
	}

	gameState := &enginepb.GameState{
		Board:         initialBoard,
		CurrentPlayer: enginepb.Player_PLAYER_ONE,
	}

	return &gamespb.Game{
		Id:        gameID,
		State:     gameState,
		Player1Id: player1ID,
		Player2Id: player2ID,
	}
}

func generateGameID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func IsPlayerInGame(game *gamespb.Game, playerID string) bool {
	return game.Player1Id == playerID || game.Player2Id == playerID
}

func GetPlayerFromID(playerID string, game *gamespb.Game) enginepb.Player {
	if game.Player1Id == playerID {
		return enginepb.Player_PLAYER_ONE
	}
	return enginepb.Player_PLAYER_TWO
}
