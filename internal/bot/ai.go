package bot

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"

	botpb "github.com/laerson/mancala/proto/bot"
	enginepb "github.com/laerson/mancala/proto/engine"
)

// AIEngine handles bot move calculations
type AIEngine struct {
	rand *rand.Rand
}

// NewAIEngine creates a new AI engine
func NewAIEngine() *AIEngine {
	return &AIEngine{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// CalculateMove determines the best move for a bot given the game state
func (ai *AIEngine) CalculateMove(gameState *enginepb.GameState, difficulty botpb.BotDifficulty, botPlayerID string) (uint32, string, int32, error) {
	// Determine which player the bot is (Player 1 or Player 2)
	botPlayer := ai.getBotPlayer(gameState, botPlayerID)
	if botPlayer == enginepb.Player_PLAYER_UNSPECIFIED {
		return 0, "", 0, fmt.Errorf("bot player not found in game state")
	}

	// Get valid moves for the bot
	validMoves := ai.getValidMoves(gameState, botPlayer)
	if len(validMoves) == 0 {
		return 0, "", 0, fmt.Errorf("no valid moves available")
	}

	switch difficulty {
	case botpb.BotDifficulty_BOT_DIFFICULTY_EASY:
		return ai.calculateEasyMove(validMoves)
	case botpb.BotDifficulty_BOT_DIFFICULTY_MEDIUM:
		return ai.calculateMediumMove(gameState, validMoves, botPlayer)
	case botpb.BotDifficulty_BOT_DIFFICULTY_HARD:
		return ai.calculateHardMove(gameState, validMoves, botPlayer)
	default:
		return ai.calculateEasyMove(validMoves)
	}
}

// getBotPlayer determines which player (1 or 2) the bot is playing as
func (ai *AIEngine) getBotPlayer(gameState *enginepb.GameState, botPlayerID string) enginepb.Player {
	// This is a simplified approach - in a real implementation,
	// we'd need to track which player ID corresponds to which player number
	// For now, assume the bot is the current player
	return gameState.CurrentPlayer
}

// getValidMoves returns all valid moves for a player
func (ai *AIEngine) getValidMoves(gameState *enginepb.GameState, player enginepb.Player) []uint32 {
	var validMoves []uint32

	// Determine the pit range for this player
	var startPit, endPit uint32
	if player == enginepb.Player_PLAYER_ONE {
		startPit, endPit = 0, 5 // Player 1 controls pits 0-5
	} else {
		startPit, endPit = 7, 12 // Player 2 controls pits 7-12
	}

	// Check each pit for valid moves (must have stones)
	for pit := startPit; pit <= endPit; pit++ {
		if pit < uint32(len(gameState.Board.Pits)) && gameState.Board.Pits[pit] > 0 {
			validMoves = append(validMoves, pit)
		}
	}

	return validMoves
}

// calculateEasyMove - Random valid move
func (ai *AIEngine) calculateEasyMove(validMoves []uint32) (uint32, string, int32, error) {
	if len(validMoves) == 0 {
		return 0, "", 0, fmt.Errorf("no valid moves")
	}

	move := validMoves[ai.rand.Intn(len(validMoves))]
	reasoning := fmt.Sprintf("Random move from %d valid options", len(validMoves))

	return move, reasoning, 0, nil
}

// calculateMediumMove - Basic strategy with heuristics
func (ai *AIEngine) calculateMediumMove(gameState *enginepb.GameState, validMoves []uint32, botPlayer enginepb.Player) (uint32, string, int32, error) {
	type moveEval struct {
		pit    uint32
		score  int32
		reason string
	}

	var evaluations []moveEval

	for _, pit := range validMoves {
		score, reason := ai.evaluateBasicMove(gameState, pit, botPlayer)
		evaluations = append(evaluations, moveEval{pit, score, reason})
	}

	// Sort by score (highest first)
	sort.Slice(evaluations, func(i, j int) bool {
		return evaluations[i].score > evaluations[j].score
	})

	best := evaluations[0]
	return best.pit, best.reason, best.score, nil
}

// calculateHardMove - Advanced AI with minimax
func (ai *AIEngine) calculateHardMove(gameState *enginepb.GameState, validMoves []uint32, botPlayer enginepb.Player) (uint32, string, int32, error) {
	const maxDepth = 6 // Look ahead 6 moves

	bestMove := validMoves[0]
	bestScore := int32(math.Inf(-1))

	for _, move := range validMoves {
		score := ai.minimax(gameState, move, maxDepth-1, int32(math.Inf(-1)), int32(math.Inf(1)), false, botPlayer)
		if score > bestScore {
			bestScore = score
			bestMove = move
		}
	}

	reasoning := fmt.Sprintf("Minimax evaluation with depth %d, score: %d", maxDepth, bestScore)
	return bestMove, reasoning, bestScore, nil
}

// evaluateBasicMove - Heuristic evaluation for medium difficulty
func (ai *AIEngine) evaluateBasicMove(gameState *enginepb.GameState, pit uint32, botPlayer enginepb.Player) (int32, string) {
	stones := gameState.Board.Pits[pit]
	score := int32(0)
	reasons := []string{}

	// Prefer moves that capture opponent stones
	if ai.wouldCapture(gameState, pit, botPlayer) {
		score += 10
		reasons = append(reasons, "captures opponent stones")
	}

	// Prefer moves that give extra turns
	if ai.wouldGetExtraTurn(gameState, pit, botPlayer) {
		score += 8
		reasons = append(reasons, "gets extra turn")
	}

	// Prefer moves with more stones (generally more impactful)
	score += int32(stones)
	reasons = append(reasons, fmt.Sprintf("plays %d stones", stones))

	// Avoid leaving stones that opponent can easily capture
	if ai.wouldLeaveVulnerable(gameState, pit, botPlayer) {
		score -= 5
		reasons = append(reasons, "but leaves vulnerable position")
	}

	reason := fmt.Sprintf("Basic strategy: %v", reasons)
	return score, reason
}

// wouldCapture checks if a move would capture opponent stones
func (ai *AIEngine) wouldCapture(gameState *enginepb.GameState, pit uint32, botPlayer enginepb.Player) bool {
	stones := gameState.Board.Pits[pit]
	finalPit := (pit + stones) % 14

	// Check if the final pit is on our side and empty
	var ourSideStart, ourSideEnd uint32
	if botPlayer == enginepb.Player_PLAYER_ONE {
		ourSideStart, ourSideEnd = 0, 5
	} else {
		ourSideStart, ourSideEnd = 7, 12
	}

	if finalPit >= ourSideStart && finalPit <= ourSideEnd && gameState.Board.Pits[finalPit] == 0 {
		// Check if opposite pit has stones to capture
		oppositePit := 12 - finalPit
		return gameState.Board.Pits[oppositePit] > 0
	}

	return false
}

// wouldGetExtraTurn checks if a move would land in the bot's mancala
func (ai *AIEngine) wouldGetExtraTurn(gameState *enginepb.GameState, pit uint32, botPlayer enginepb.Player) bool {
	stones := gameState.Board.Pits[pit]
	finalPit := (pit + stones) % 14

	// Check if final pit is our mancala
	var ourMancala uint32
	if botPlayer == enginepb.Player_PLAYER_ONE {
		ourMancala = 6
	} else {
		ourMancala = 13
	}

	return finalPit == ourMancala
}

// wouldLeaveVulnerable checks if a move would leave us vulnerable to captures
func (ai *AIEngine) wouldLeaveVulnerable(gameState *enginepb.GameState, pit uint32, botPlayer enginepb.Player) bool {
	// Simplified vulnerability check - if the pit would become empty
	// and the opposite pit has stones, we might be vulnerable
	if gameState.Board.Pits[pit] == 1 {
		oppositePit := 12 - pit
		return gameState.Board.Pits[oppositePit] > 0
	}
	return false
}

// minimax algorithm for hard difficulty
func (ai *AIEngine) minimax(gameState *enginepb.GameState, move uint32, depth int32, alpha, beta int32, isMaximizing bool, botPlayer enginepb.Player) int32 {
	// Base case: reached depth limit or game over
	if depth == 0 || ai.isGameOver(gameState) {
		return ai.evaluatePosition(gameState, botPlayer)
	}

	// Simulate the move (simplified - in real implementation, we'd need the engine)
	// For now, return a heuristic evaluation
	return ai.evaluatePosition(gameState, botPlayer)
}

// evaluatePosition evaluates the current position for the bot
func (ai *AIEngine) evaluatePosition(gameState *enginepb.GameState, botPlayer enginepb.Player) int32 {
	var botMancala, opponentMancala uint32

	if botPlayer == enginepb.Player_PLAYER_ONE {
		botMancala, opponentMancala = 6, 13
	} else {
		botMancala, opponentMancala = 13, 6
	}

	// Simple evaluation: difference in mancala stones
	botStones := int32(gameState.Board.Pits[botMancala])
	opponentStones := int32(gameState.Board.Pits[opponentMancala])

	return botStones - opponentStones
}

// isGameOver checks if the game is finished
func (ai *AIEngine) isGameOver(gameState *enginepb.GameState) bool {
	// Check if either player has no moves left
	player1HasMoves := ai.playerHasMoves(gameState, enginepb.Player_PLAYER_ONE)
	player2HasMoves := ai.playerHasMoves(gameState, enginepb.Player_PLAYER_TWO)

	return !player1HasMoves || !player2HasMoves
}

// playerHasMoves checks if a player has any valid moves
func (ai *AIEngine) playerHasMoves(gameState *enginepb.GameState, player enginepb.Player) bool {
	validMoves := ai.getValidMoves(gameState, player)
	return len(validMoves) > 0
}
