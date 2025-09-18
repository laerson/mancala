package mancala

import (
	"fmt"
	"strings"
)

// GameBoard represents a Mancala game board
type GameBoard struct {
	Pits          []uint32 `json:"pits"`
	CurrentPlayer int      `json:"current_player"`
}

// DisplayBoard displays the Mancala board in ASCII art
func DisplayBoard(board GameBoard) {
	if len(board.Pits) != 14 {
		fmt.Println("Invalid board: expected 14 pits")
		return
	}

	fmt.Println()
	fmt.Println("    MANCALA BOARD")
	fmt.Println("  Player 2's side")
	fmt.Println()

	// Top row (Player 2's pits) - indices 7-12 (reverse order for display)
	fmt.Print("  ")
	for i := 12; i >= 7; i-- {
		fmt.Printf("[ %2d ]", board.Pits[i])
	}
	fmt.Println()

	// Mancalas (Player 2's mancala on left, Player 1's on right)
	fmt.Printf("[ %2d ]", board.Pits[13]) // Player 2's mancala
	fmt.Print(strings.Repeat("      ", 6))
	fmt.Printf("[ %2d ]", board.Pits[6]) // Player 1's mancala
	fmt.Println()

	// Bottom row (Player 1's pits) - indices 0-5
	fmt.Print("  ")
	for i := 0; i <= 5; i++ {
		fmt.Printf("[ %2d ]", board.Pits[i])
	}
	fmt.Println()

	fmt.Println()
	fmt.Println("  Player 1's side")

	// Show pit numbers for reference
	fmt.Println("\n  Pit numbers (Player 1):")
	fmt.Print("  ")
	for i := 0; i <= 5; i++ {
		fmt.Printf("  %2d  ", i)
	}
	fmt.Println()

	// Current player indicator
	if board.CurrentPlayer == 0 {
		fmt.Println("\n  >>> Player 1's turn <<<")
	} else {
		fmt.Println("\n  >>> Player 2's turn <<<")
	}
	fmt.Println()
}

// DisplayWelcome displays a welcome message
func DisplayWelcome() {
	fmt.Println(`
╔═══════════════════════════════════════════════════════════════╗
║                      MANCALA GAME CLIENT                     ║
║                                                               ║
║  Welcome to the Mancala CLI client!                          ║
║                                                               ║
║  Commands:                                                    ║
║    mancala connect <server-ip>  - Connect to game server     ║
║    mancala register             - Create new account         ║
║    mancala login                - Login to existing account  ║
║    mancala play                 - Join matchmaking queue     ║
║    mancala status               - Check connection status    ║
║    mancala logout               - Logout from current account║
║                                                               ║
╚═══════════════════════════════════════════════════════════════╝
`)
}

// DisplayConnectionStatus displays the current connection status
func DisplayConnectionStatus(state *ClientState) {
	config := state.GetConfig()

	fmt.Println("\n=== CONNECTION STATUS ===")
	if config.ServerURL != "" {
		fmt.Printf("Server: %s\n", config.ServerURL)
		fmt.Printf("Connected: ✓\n")
	} else {
		fmt.Printf("Server: Not connected\n")
		fmt.Printf("Connected: ✗\n")
	}

	if config.Username != "" {
		fmt.Printf("Username: %s\n", config.Username)
		fmt.Printf("Logged in: ✓\n")
	} else {
		fmt.Printf("Username: Not logged in\n")
		fmt.Printf("Logged in: ✗\n")
	}
	fmt.Println()
}

// DisplayMatchFound displays when a match is found
func DisplayMatchFound(data map[string]interface{}) {
	fmt.Println("\n🎯 MATCH FOUND! 🎯")
	fmt.Println("==================")

	if player1ID, ok := data["player1_id"].(string); ok {
		if player1Name, ok := data["player1_name"].(string); ok {
			fmt.Printf("Player 1: %s (%s)\n", player1Name, player1ID)
		}
	}

	if player2ID, ok := data["player2_id"].(string); ok {
		if player2Name, ok := data["player2_name"].(string); ok {
			fmt.Printf("Player 2: %s (%s)\n", player2Name, player2ID)
		}
	}

	fmt.Println("\nGame is starting...")
	fmt.Println("Use 'mancala move <pit>' to make moves")
	fmt.Println()
}

// DisplayMoveResult displays the result of a move
func DisplayMoveResult(data map[string]interface{}) {
	fmt.Println("\n📱 MOVE MADE")
	fmt.Println("=============")

	if playerID, ok := data["player_id"].(string); ok {
		fmt.Printf("Player: %s\n", playerID)
	}

	if pitIndex, ok := data["pit_index"].(float64); ok {
		fmt.Printf("Pit: %d\n", int(pitIndex))
	}

	// Display updated board if available
	if gameState, ok := data["game_state"].(map[string]interface{}); ok {
		if boardData, ok := gameState["board"].([]interface{}); ok {
			board := GameBoard{
				Pits: make([]uint32, len(boardData)),
			}

			for i, pit := range boardData {
				if pitValue, ok := pit.(float64); ok {
					board.Pits[i] = uint32(pitValue)
				}
			}

			if currentPlayer, ok := gameState["current_player"].(float64); ok {
				board.CurrentPlayer = int(currentPlayer)
			}

			DisplayBoard(board)
		}
	}
}

// DisplayGameOver displays game over information
func DisplayGameOver(data map[string]interface{}) {
	fmt.Println("\n🏁 GAME OVER! 🏁")
	fmt.Println("=================")

	if winnerID, ok := data["winner_id"].(string); ok {
		if winnerID != "" {
			fmt.Printf("Winner: %s\n", winnerID)
		} else {
			fmt.Println("Result: Draw!")
		}
	}

	if isDraw, ok := data["is_draw"].(bool); ok && isDraw {
		fmt.Println("Result: It's a draw!")
	}

	// Display final board if available
	if finalState, ok := data["final_state"].(map[string]interface{}); ok {
		if boardData, ok := finalState["board"].([]interface{}); ok {
			board := GameBoard{
				Pits: make([]uint32, len(boardData)),
			}

			for i, pit := range boardData {
				if pitValue, ok := pit.(float64); ok {
					board.Pits[i] = uint32(pitValue)
				}
			}

			fmt.Println("\nFinal Board:")
			DisplayBoard(board)
		}
	}

	fmt.Println("Thanks for playing!")
	fmt.Println()
}
