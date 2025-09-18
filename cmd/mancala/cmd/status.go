package cmd

import (
	"github.com/laerson/mancala/internal/mancala"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show connection and login status",
	Long:  `Display the current connection status and login information.`,
	Run: func(cmd *cobra.Command, args []string) {
		mancala.DisplayConnectionStatus(clientState)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
