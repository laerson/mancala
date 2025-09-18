package cmd

import (
	"fmt"
	"os"

	"github.com/laerson/mancala/internal/mancala"
	"github.com/spf13/cobra"
)

var (
	clientState   *mancala.ClientState
	apiClient     *mancala.APIClient
	currentGameID string
	inGame        bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mancala",
	Short: "Mancala Game CLI Client",
	Long: `A command line client for the Mancala game.

Connect to a Mancala game server, create an account, and play matches
against other players online.`,
	Run: func(cmd *cobra.Command, args []string) {
		mancala.DisplayWelcome()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Initialize client state
	var err error
	clientState, err = mancala.NewClientState()
	if err != nil {
		fmt.Printf("Error initializing client state: %v\n", err)
		os.Exit(1)
	}

	// Initialize API client if connected
	config := clientState.GetConfig()
	if config.ServerURL != "" {
		apiClient = mancala.NewAPIClient(config.ServerURL)
		if config.AccessToken != "" {
			apiClient.SetToken(config.AccessToken)
		}
	}
}
