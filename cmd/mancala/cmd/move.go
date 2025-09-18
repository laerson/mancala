package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var moveCmd = &cobra.Command{
	Use:   "move <pit-number>",
	Short: "Make a move in the current game",
	Long: `Make a move in the current game by selecting a pit number (0-5).

Pit numbers for Player 1:
  0  1  2  3  4  5

Example:
  mancala move 3`,
	Args: cobra.ExactArgs(1),
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

		if currentGameID == "" {
			fmt.Println("❌ No active game. Use 'mancala play' to join a game first.")
			return
		}

		pitStr := args[0]
		pitIndex, err := strconv.Atoi(pitStr)
		if err != nil {
			fmt.Printf("❌ Invalid pit number: %s. Must be a number between 0-5.\n", pitStr)
			return
		}

		if pitIndex < 0 || pitIndex > 5 {
			fmt.Printf("❌ Invalid pit number: %d. Must be between 0-5.\n", pitIndex)
			return
		}

		config := clientState.GetConfig()

		fmt.Printf("🎲 Making move: pit %d...\n", pitIndex)

		// Make the move
		resp, err := apiClient.MakeMove(currentGameID, config.UserID, uint32(pitIndex))
		if err != nil {
			fmt.Printf("❌ Failed to make move: %v\n", err)
			return
		}

		if !resp.Success {
			fmt.Printf("❌ Move failed: %s\n", resp.Error)
			return
		}

		fmt.Printf("✅ Move successful!\n")
		fmt.Println("⏳ Waiting for opponent's move...")

		// The game state will be updated via notifications in the play command
	},
}

func init() {
	rootCmd.AddCommand(moveCmd)
}
