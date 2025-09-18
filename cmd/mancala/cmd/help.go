package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var helpCmd = &cobra.Command{
	Use:   "quickstart",
	Short: "Show quick start guide",
	Long:  `Display a quick start guide for new users.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(`
🎮 MANCALA CLI QUICK START GUIDE

1️⃣  CONNECT TO SERVER
   mancala connect <server-ip>
   Example: mancala connect 192.168.1.100

2️⃣  CREATE ACCOUNT OR LOGIN
   mancala register    (for new account)
   mancala login       (for existing account)

3️⃣  JOIN A GAME
   mancala play
   ⚠️  Keep this terminal open for notifications!

4️⃣  MAKE MOVES (in another terminal)
   mancala move <pit-number>
   Example: mancala move 3

   Pit numbers: 0, 1, 2, 3, 4, 5

📊 CHECK STATUS
   mancala status

🚪 LOGOUT
   mancala logout

💡 TIPS:
   • Use two terminals: one for 'play', one for 'move'
   • Your moves are pit numbers 0-5
   • Game updates appear in real-time
   • Press Ctrl+C in 'play' to leave queue

🎯 GAME BOARD LAYOUT:
    Player 2's side
  [ ][ ][ ][ ][ ][ ]
[ ]               [ ]  ← Mancalas
  [ ][ ][ ][ ][ ][ ]
    Player 1's side
     0  1  2  3  4  5  ← Your pit numbers

📚 For full documentation: see docs/CLI_CLIENT.md
`)
	},
}

func init() {
	rootCmd.AddCommand(helpCmd)
}
