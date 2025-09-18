package cmd

import (
	"fmt"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to your account",
	Long:  `Login to your existing user account on the connected Mancala server.`,
	Run: func(cmd *cobra.Command, args []string) {
		if !clientState.IsConnected() {
			fmt.Println("❌ Not connected to a server. Use 'mancala connect <server-ip>' first.")
			return
		}

		if apiClient == nil {
			fmt.Println("❌ API client not initialized. Please reconnect.")
			return
		}

		fmt.Println("=== LOGIN ===")

		// Get username
		fmt.Print("Username: ")
		var username string
		fmt.Scanln(&username)

		if username == "" {
			fmt.Println("❌ Username cannot be empty.")
			return
		}

		// Get password (hidden input)
		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Printf("\n❌ Error reading password: %v\n", err)
			return
		}
		password := string(passwordBytes)
		fmt.Println() // New line after hidden password input

		if password == "" {
			fmt.Println("❌ Password cannot be empty.")
			return
		}

		fmt.Println("Logging in...")

		// Login
		resp, err := apiClient.Login(username, password)
		if err != nil {
			fmt.Printf("❌ Login failed: %v\n", err)
			return
		}

		if !resp.Success {
			fmt.Printf("❌ Login failed: %s\n", resp.Message)
			return
		}

		// Save authentication info
		err = clientState.SetAuth(resp.AccessToken, resp.RefreshToken, resp.User.Username, resp.User.UserID)
		if err != nil {
			fmt.Printf("⚠️ Login successful but failed to save login info: %v\n", err)
			fmt.Println("You may need to login again next time.")
		}

		// Update API client token
		apiClient.SetToken(resp.AccessToken)

		fmt.Printf("✅ Login successful!\n")
		fmt.Printf("Welcome back, %s!\n", resp.User.Username)
		fmt.Println("\nUse 'mancala play' to join a game!")
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
