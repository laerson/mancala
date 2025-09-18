package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	botDifficulty string
)

var botCmd = &cobra.Command{
	Use:   "bot [difficulty]",
	Short: "Play against a bot opponent",
	Long: `Play against an AI bot opponent. You can specify the difficulty level:
- easy: Random moves, perfect for beginners
- medium: Basic strategy with captures and extra turns
- hard: Advanced AI with minimax algorithm

If no difficulty is specified, medium difficulty will be used.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if !clientState.IsConnected() {
			fmt.Println("❌ Not connected to a server. Use 'mancala connect <server-ip>' first.")
			return
		}

		if !clientState.IsLoggedIn() {
			fmt.Println("❌ Not logged in. Use 'mancala login' or 'mancala register' first.")
			return
		}

		if apiClient == nil {
			fmt.Println("❌ API client not initialized. Please reconnect.")
			return
		}

		config := clientState.GetConfig()

		// Set difficulty from argument or flag
		difficulty := "medium" // default
		if len(args) > 0 {
			difficulty = args[0]
		} else if botDifficulty != "" {
			difficulty = botDifficulty
		}

		// Validate difficulty
		if difficulty != "easy" && difficulty != "medium" && difficulty != "hard" {
			fmt.Printf("❌ Invalid difficulty '%s'. Use 'easy', 'medium', or 'hard'\n", difficulty)
			return
		}

		fmt.Printf("🤖 Creating bot match (%s difficulty) for %s...\n", difficulty, config.Username)

		// Create bot match
		resp, err := apiClient.BotMatch(config.UserID, config.Username, difficulty)
		if err != nil {
			fmt.Printf("❌ Failed to create bot match: %v\n", err)
			return
		}

		if !resp.Success {
			fmt.Printf("❌ Failed to create bot match: %s\n", resp.Message)
			return
		}

		fmt.Printf("✅ %s\n", resp.Message)
		fmt.Printf("🎮 Game ID: %s\n", resp.GameID)
		fmt.Printf("🤖 Bot Opponent: %s\n", resp.BotName)
		fmt.Println("\n🎯 Your game is ready! Use 'mancala move <pit>' to make your first move.")
		fmt.Println("Use 'mancala status' to see the current game state.")

		// Update current game state
		currentGameID = resp.GameID
		inGame = true
	},
}

func init() {
	rootCmd.AddCommand(botCmd)

	// Add difficulty flag as alternative to positional argument
	botCmd.Flags().StringVarP(&botDifficulty, "difficulty", "d", "", "Bot difficulty (easy, medium, hard)")
}
