package matchmaking

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/laerson/mancala/internal/auth"
	"github.com/laerson/mancala/internal/events"
	gamespb "github.com/laerson/mancala/proto/games"
	matchmakingpb "github.com/laerson/mancala/proto/matchmaking"
)

type Server struct {
	matchmakingpb.UnimplementedMatchmakingServer
	queue          *PlayerQueue
	gamesClient    gamespb.GamesClient
	eventPublisher *events.EventPublisher
}

func NewServer(gamesClient gamespb.GamesClient, redisAddr string) *Server {
	server := &Server{
		queue:          NewPlayerQueue(),
		gamesClient:    gamesClient,
		eventPublisher: events.NewEventPublisher(redisAddr),
	}

	// Start background matchmaking process
	go server.processMatchmaking()

	return server
}

func (s *Server) Enqueue(ctx context.Context, req *matchmakingpb.EnqueueRequest) (*matchmakingpb.EnqueueResponse, error) {
	if req.Player == nil {
		return nil, status.Errorf(codes.InvalidArgument, "player is required")
	}

	if req.Player.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "player ID is required")
	}

	if req.Player.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "player name is required")
	}

	// Validate that the authenticated user owns this player ID
	if err := auth.ValidatePlayerOwnership(ctx, req.Player.Id); err != nil {
		return nil, err
	}

	queueID := uuid.New().String()
	s.queue.Enqueue(req.Player, queueID)

	log.Printf("Player %s (%s) enqueued with queue ID %s", req.Player.Id, req.Player.Name, queueID)

	return &matchmakingpb.EnqueueResponse{
		Success: true,
		QueueId: queueID,
		Message: fmt.Sprintf("Player %s successfully enqueued", req.Player.Name),
	}, nil
}

func (s *Server) CancelQueue(ctx context.Context, req *matchmakingpb.CancelQueueRequest) (*matchmakingpb.CancelQueueResponse, error) {
	if req.PlayerId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "player ID is required")
	}

	// Validate that the authenticated user owns this player ID
	if err := auth.ValidatePlayerOwnership(ctx, req.PlayerId); err != nil {
		return nil, err
	}

	removed := s.queue.RemovePlayer(req.PlayerId)
	if !removed {
		return &matchmakingpb.CancelQueueResponse{
			Success: false,
			Message: "Player not found in queue",
		}, nil
	}

	log.Printf("Player %s removed from queue", req.PlayerId)

	return &matchmakingpb.CancelQueueResponse{
		Success: true,
		Message: "Successfully removed from queue",
	}, nil
}

func (s *Server) GetQueueStatus(ctx context.Context, req *matchmakingpb.GetQueueStatusRequest) (*matchmakingpb.GetQueueStatusResponse, error) {
	if req.PlayerId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "player ID is required")
	}

	// Validate that the authenticated user owns this player ID
	if err := auth.ValidatePlayerOwnership(ctx, req.PlayerId); err != nil {
		return nil, err
	}

	queuedPlayer, position := s.queue.GetPlayerStatus(req.PlayerId)
	if queuedPlayer == nil {
		return &matchmakingpb.GetQueueStatusResponse{
			Status:        matchmakingpb.QueueStatus_CANCELLED,
			QueuePosition: 0,
		}, nil
	}

	return &matchmakingpb.GetQueueStatusResponse{
		Status:        matchmakingpb.QueueStatus_QUEUED,
		QueuePosition: position,
	}, nil
}

func (s *Server) StreamUpdates(req *matchmakingpb.StreamUpdatesRequest, stream matchmakingpb.Matchmaking_StreamUpdatesServer) error {
	if req.PlayerId == "" {
		return status.Errorf(codes.InvalidArgument, "player ID is required")
	}

	// Validate that the authenticated user owns this player ID
	if err := auth.ValidatePlayerOwnership(stream.Context(), req.PlayerId); err != nil {
		return err
	}

	// Set the stream for this player
	s.queue.SetPlayerStream(req.PlayerId, stream)

	// Send initial queue position
	queuedPlayer, position := s.queue.GetPlayerStatus(req.PlayerId)
	if queuedPlayer != nil {
		stream.Send(&matchmakingpb.MatchmakingUpdate{
			QueueId: queuedPlayer.QueueID,
			Status:  matchmakingpb.QueueStatus_QUEUED,
			Update: &matchmakingpb.MatchmakingUpdate_QueuePosition{
				QueuePosition: &matchmakingpb.QueuePositionUpdate{
					Position: position,
				},
			},
		})
	}

	// Keep connection alive until context is done
	<-stream.Context().Done()
	return nil
}

func (s *Server) processMatchmaking() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		player1, player2 := s.queue.TryMatchPlayers()
		if player1 != nil && player2 != nil {
			go s.createMatch(player1, player2)
		}
	}
}

func (s *Server) createMatch(player1, player2 *QueuedPlayer) {
	log.Printf("Creating match between %s and %s", player1.Player.Name, player2.Player.Name)

	// Create game via Games service
	gameReq := &gamespb.CreateGameRequest{
		Player1Id: player1.Player.Id,
		Player2Id: player2.Player.Id,
	}

	gameResp, err := s.gamesClient.Create(context.Background(), gameReq)
	if err != nil {
		log.Printf("Failed to create game: %v", err)
		// TODO: Re-queue players or notify them of failure
		return
	}

	log.Printf("Game created: %s", gameResp.Game.Id)

	// Publish MatchFound event
	matchID := uuid.New().String()
	err = s.eventPublisher.PublishMatchFound(
		context.Background(),
		gameResp.Game.Id,
		matchID,
		player1.Player.Id,
		player1.Player.Name,
		player2.Player.Id,
		player2.Player.Name,
	)
	if err != nil {
		log.Printf("Failed to publish match found event: %v", err)
	}

	// Notify players via streams
	s.notifyPlayerMatch(player1, player2, gameResp.Game, matchID)
	s.notifyPlayerMatch(player2, player1, gameResp.Game, matchID)
}

func (s *Server) notifyPlayerMatch(player, opponent *QueuedPlayer, game *gamespb.Game, matchID string) {
	if player.Stream != nil {
		player.Stream.Send(&matchmakingpb.MatchmakingUpdate{
			QueueId: player.QueueID,
			Status:  matchmakingpb.QueueStatus_GAME_CREATED,
			Update: &matchmakingpb.MatchmakingUpdate_GameCreated{
				GameCreated: &matchmakingpb.GameCreated{
					GameId: game.Id,
					Game:   game,
				},
			},
		})
	}
}
