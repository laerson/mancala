package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from the current account",
	Long:  `Logout from the current account and clear saved authentication information.`,
	Run: func(cmd *cobra.Command, args []string) {
		if !clientState.IsLoggedIn() {
			fmt.Println("❌ Not logged in.")
			return
		}

		config := clientState.GetConfig()
		username := config.Username

		err := clientState.ClearAuth()
		if err != nil {
			fmt.Printf("❌ Failed to logout: %v\n", err)
			return
		}

		// Clear API client token
		if apiClient != nil {
			apiClient.SetToken("")
		}

		fmt.Printf("✅ Successfully logged out %s.\n", username)
		fmt.Println("Use 'mancala login' or 'mancala register' to log back in.")
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
