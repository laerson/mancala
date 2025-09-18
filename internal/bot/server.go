package bot

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	botpb "github.com/laerson/mancala/proto/bot"
	enginepb "github.com/laerson/mancala/proto/engine"
)

// Server implements the Bot gRPC service
type Server struct {
	botpb.UnimplementedBotServer
	aiEngine *AIEngine
}

// NewServer creates a new Bot service server
func NewServer() *Server {
	return &Server{
		aiEngine: NewAIEngine(),
	}
}

// GetMove calculates the next move for a bot
func (s *Server) GetMove(ctx context.Context, req *botpb.GetMoveRequest) (*botpb.GetMoveResponse, error) {
	if req.GameState == nil {
		return &botpb.GetMoveResponse{
			Result: &botpb.GetMoveResponse_Error{
				Error: &botpb.Error{
					Message: "game state is required",
					Code:    "INVALID_REQUEST",
				},
			},
		}, nil
	}

	if req.BotId == "" {
		return &botpb.GetMoveResponse{
			Result: &botpb.GetMoveResponse_Error{
				Error: &botpb.Error{
					Message: "bot ID is required",
					Code:    "INVALID_REQUEST",
				},
			},
		}, nil
	}

	// Calculate the bot's move
	pitIndex, reasoning, evaluationScore, err := s.aiEngine.CalculateMove(
		req.GameState,
		req.Difficulty,
		req.BotId,
	)

	if err != nil {
		return &botpb.GetMoveResponse{
			Result: &botpb.GetMoveResponse_Error{
				Error: &botpb.Error{
					Message: err.Error(),
					Code:    "CALCULATION_ERROR",
				},
			},
		}, nil
	}

	return &botpb.GetMoveResponse{
		Result: &botpb.GetMoveResponse_Move{
			Move: &botpb.MoveResult{
				PitIndex:        pitIndex,
				Reasoning:       reasoning,
				EvaluationScore: evaluationScore,
			},
		},
	}, nil
}

// ListBots returns all available bot profiles
func (s *Server) ListBots(ctx context.Context, req *botpb.ListBotsRequest) (*botpb.ListBotsResponse, error) {
	bots := []*botpb.BotProfile{
		{
			Id:          "easy-bot",
			Name:        "Novice Bot",
			Difficulty:  botpb.BotDifficulty_BOT_DIFFICULTY_EASY,
			Description: "Makes random moves - perfect for beginners",
			Wins:        150,
			Losses:      350,
		},
		{
			Id:          "medium-bot",
			Name:        "Strategic Bot",
			Difficulty:  botpb.BotDifficulty_BOT_DIFFICULTY_MEDIUM,
			Description: "Uses basic strategy - looks for captures and extra turns",
			Wins:        280,
			Losses:      220,
		},
		{
			Id:          "hard-bot",
			Name:        "Master Bot",
			Difficulty:  botpb.BotDifficulty_BOT_DIFFICULTY_HARD,
			Description: "Advanced AI with minimax algorithm - challenging opponent",
			Wins:        420,
			Losses:      80,
		},
	}

	return &botpb.ListBotsResponse{
		Bots: bots,
	}, nil
}

// CreateBot creates a new bot instance for a game
func (s *Server) CreateBot(ctx context.Context, req *botpb.CreateBotRequest) (*botpb.CreateBotResponse, error) {
	if req.Difficulty == botpb.BotDifficulty_BOT_DIFFICULTY_UNSPECIFIED {
		req.Difficulty = botpb.BotDifficulty_BOT_DIFFICULTY_MEDIUM // Default to medium
	}

	// Generate a unique bot ID
	botID := fmt.Sprintf("bot-%s", uuid.New().String()[:8])

	// Determine bot name based on difficulty
	var botName string
	var description string
	var wins, losses int32

	switch req.Difficulty {
	case botpb.BotDifficulty_BOT_DIFFICULTY_EASY:
		botName = "Novice Bot"
		description = "Makes random moves - perfect for beginners"
		wins, losses = 150, 350
	case botpb.BotDifficulty_BOT_DIFFICULTY_MEDIUM:
		botName = "Strategic Bot"
		description = "Uses basic strategy - looks for captures and extra turns"
		wins, losses = 280, 220
	case botpb.BotDifficulty_BOT_DIFFICULTY_HARD:
		botName = "Master Bot"
		description = "Advanced AI with minimax algorithm - challenging opponent"
		wins, losses = 420, 80
	}

	// Add custom suffix if provided
	if req.NameSuffix != "" {
		botName += " " + req.NameSuffix
	}

	bot := &botpb.BotProfile{
		Id:          botID,
		Name:        botName,
		Difficulty:  req.Difficulty,
		Description: description,
		Wins:        wins,
		Losses:      losses,
	}

	return &botpb.CreateBotResponse{
		Bot: bot,
	}, nil
}
