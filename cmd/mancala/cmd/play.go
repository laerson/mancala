package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/laerson/mancala/internal/mancala"
	"github.com/spf13/cobra"
)

var (
	currentGameID string
	inGame        bool
)

var playCmd = &cobra.Command{
	Use:   "play",
	Short: "Join the matchmaking queue",
	Long:  `Join the matchmaking queue to be paired with another player for a game.`,
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

		fmt.Printf("🎮 Joining matchmaking queue as %s...\n", config.Username)

		// Enqueue for matchmaking
		resp, err := apiClient.Enqueue(config.UserID, config.Username)
		if err != nil {
			fmt.Printf("❌ Failed to join queue: %v\n", err)
			return
		}

		if !resp.Success {
			fmt.Printf("❌ Failed to join queue: %s\n", resp.Message)
			return
		}

		fmt.Printf("✅ %s\n", resp.Message)
		fmt.Println("⏳ Waiting for an opponent...")
		fmt.Println("Press Ctrl+C to cancel and leave the queue.")

		// Set up context for cancellation
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Handle Ctrl+C
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		go func() {
			<-sigChan
			fmt.Println("\n🚫 Cancelling queue...")
			if err := apiClient.CancelQueue(config.UserID); err != nil {
				fmt.Printf("⚠️ Error leaving queue: %v\n", err)
			} else {
				fmt.Println("✅ Left matchmaking queue.")
			}
			cancel()
		}()

		// Subscribe to notifications
		notificationClient := mancala.NewNotificationClient(config.ServerURL, config.AccessToken)

		err = notificationClient.Subscribe(ctx, config.UserID, func(notification mancala.Notification) {
			switch notification.Type {
			case "NOTIFICATION_TYPE_MATCH_FOUND":
				fmt.Println("\n🎯 MATCH FOUND!")
				mancala.DisplayMatchFound(notification.Data)

				// Extract game ID from notification
				if notification.GameID != "" {
					currentGameID = notification.GameID
					inGame = true
					fmt.Printf("\n📝 Game ID: %s\n", notification.GameID)
					fmt.Println("Use 'mancala move <pit>' to make moves (in a new terminal)")
				}

			case "NOTIFICATION_TYPE_MOVE_MADE":
				if inGame {
					mancala.DisplayMoveResult(notification.Data)
					fmt.Print("\nWaiting for your move (use 'mancala move <pit>' in a new terminal)...")
				}

			case "NOTIFICATION_TYPE_GAME_OVER":
				if inGame {
					mancala.DisplayGameOver(notification.Data)
					inGame = false
					currentGameID = ""
					cancel() // End the notification subscription
				}
			}
		})

		if err != nil && err != context.Canceled {
			fmt.Printf("❌ Notification error: %v\n", err)
		}

		fmt.Println("\n👋 Session ended.")
	},
}

func init() {
	rootCmd.AddCommand(playCmd)
}
