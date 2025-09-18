package cmd

import (
	"fmt"
	"strings"

	"github.com/laerson/mancala/internal/mancala"
	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:   "connect <server-ip>",
	Short: "Connect to a Mancala game server",
	Long: `Connect to a Mancala game server by providing the server's IP address.

Examples:
  mancala connect 192.168.1.100
  mancala connect localhost
  mancala connect mancala.example.com`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serverIP := args[0]

		// Build server URL
		serverURL := serverIP
		if !strings.HasPrefix(serverURL, "http://") && !strings.HasPrefix(serverURL, "https://") {
			serverURL = "http://" + serverURL
		}

		// Add port if not specified
		if !strings.Contains(serverURL[7:], ":") { // Skip the "http://" part
			serverURL += ":8080"
		}

		fmt.Printf("Connecting to server: %s\n", serverURL)

		// Create API client and test connection
		testClient := mancala.NewAPIClient(serverURL)
		if err := testClient.TestConnection(); err != nil {
			fmt.Printf("❌ Failed to connect to server: %v\n", err)
			fmt.Println("\nPlease check:")
			fmt.Println("- Server IP address is correct")
			fmt.Println("- Server is running and accessible")
			fmt.Println("- Network connectivity")
			return
		}

		// Save connection
		if err := clientState.SetServerURL(serverURL); err != nil {
			fmt.Printf("❌ Failed to save connection: %v\n", err)
			return
		}

		// Update global API client
		apiClient = testClient

		fmt.Printf("✅ Successfully connected to %s\n", serverURL)
		fmt.Println("\nNext steps:")
		fmt.Println("- Run 'mancala register' to create a new account")
		fmt.Println("- Run 'mancala login' to login with existing account")
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)
}
