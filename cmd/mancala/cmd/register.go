package cmd

import (
	"fmt"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new user account",
	Long:  `Create a new user account on the connected Mancala server.`,
	Run: func(cmd *cobra.Command, args []string) {
		if !clientState.IsConnected() {
			fmt.Println("❌ Not connected to a server. Use 'mancala connect <server-ip>' first.")
			return
		}

		if apiClient == nil {
			fmt.Println("❌ API client not initialized. Please reconnect.")
			return
		}

		fmt.Println("=== REGISTER NEW ACCOUNT ===")

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

		// Confirm password
		fmt.Print("Confirm Password: ")
		confirmPasswordBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Printf("\n❌ Error reading password confirmation: %v\n", err)
			return
		}
		confirmPassword := string(confirmPasswordBytes)
		fmt.Println()

		if password != confirmPassword {
			fmt.Println("❌ Passwords do not match.")
			return
		}

		fmt.Println("Creating account...")

		// Register user
		resp, err := apiClient.Register(username, password)
		if err != nil {
			fmt.Printf("❌ Registration failed: %v\n", err)
			return
		}

		if !resp.Success {
			fmt.Printf("❌ Registration failed: %s\n", resp.Message)
			return
		}

		// Save authentication info
		err = clientState.SetAuth(resp.AccessToken, resp.RefreshToken, resp.User.Username, resp.User.UserID)
		if err != nil {
			fmt.Printf("⚠️ Registration successful but failed to save login info: %v\n", err)
			fmt.Println("You may need to login manually.")
			return
		}

		// Update API client token
		apiClient.SetToken(resp.AccessToken)

		fmt.Printf("✅ Account created successfully!\n")
		fmt.Printf("Welcome, %s!\n", resp.User.Username)
		fmt.Println("\nYou are now logged in.")
		fmt.Println("Use 'mancala play' to join a game!")
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)
}
