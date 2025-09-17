package matchmaking

import (
	"context"
	"fmt"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	gamespb "github.com/laerson/mancala/proto/games"
	matchmakingpb "github.com/laerson/mancala/proto/matchmaking"
)

// Mock Games client for testing
type mockGamesClient struct {
	createGameFunc func(ctx context.Context, req *gamespb.CreateGameRequest) (*gamespb.CreateGameResponse, error)
}

func (m *mockGamesClient) Create(ctx context.Context, req *gamespb.CreateGameRequest, opts ...grpc.CallOption) (*gamespb.CreateGameResponse, error) {
	if m.createGameFunc != nil {
		return m.createGameFunc(ctx, req)
	}
	// Default implementation
	return &gamespb.CreateGameResponse{
		Game: &gamespb.Game{
			Id:        "test-game-id",
			Player1Id: req.Player1Id,
			Player2Id: req.Player2Id,
		},
	}, nil
}

func (m *mockGamesClient) Move(ctx context.Context, req *gamespb.MakeGameMoveRequest, opts ...grpc.CallOption) (*gamespb.MakeGameMoveResponse, error) {
	// Not needed for matchmaking tests
	return nil, nil
}

func TestServer_Enqueue(t *testing.T) {
	server := NewServer(&mockGamesClient{}, "redis:6379")

	tests := []struct {
		name    string
		req     *matchmakingpb.EnqueueRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "valid enqueue",
			req: &matchmakingpb.EnqueueRequest{
				Player: &matchmakingpb.Player{
					Id:   "player1",
					Name: "Alice",
				},
			},
			wantErr: false,
		},
		{
			name: "nil player",
			req: &matchmakingpb.EnqueueRequest{
				Player: nil,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "empty player ID",
			req: &matchmakingpb.EnqueueRequest{
				Player: &matchmakingpb.Player{
					Id:   "",
					Name: "Alice",
				},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "empty player name",
			req: &matchmakingpb.EnqueueRequest{
				Player: &matchmakingpb.Player{
					Id:   "player1",
					Name: "",
				},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := server.Enqueue(context.Background(), tt.req)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if st, ok := status.FromError(err); ok {
					if st.Code() != tt.errCode {
						t.Errorf("Expected error code %v, got %v", tt.errCode, st.Code())
					}
				} else {
					t.Error("Expected gRPC status error")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if resp == nil {
					t.Error("Expected response but got nil")
				}
				if !resp.Success {
					t.Error("Expected successful enqueue")
				}
				if resp.QueueId == "" {
					t.Error("Expected non-empty queue ID")
				}
			}
		})
	}
}

func TestServer_CancelQueue(t *testing.T) {
	server := NewServer(&mockGamesClient{}, "redis:6379")

	// First enqueue a player
	enqueueReq := &matchmakingpb.EnqueueRequest{
		Player: &matchmakingpb.Player{
			Id:   "player1",
			Name: "Alice",
		},
	}
	enqueueResp, err := server.Enqueue(context.Background(), enqueueReq)
	if err != nil {
		t.Fatalf("Failed to enqueue player: %v", err)
	}

	tests := []struct {
		name      string
		req       *matchmakingpb.CancelQueueRequest
		wantErr   bool
		errCode   codes.Code
		wantFound bool
	}{
		{
			name: "valid cancel",
			req: &matchmakingpb.CancelQueueRequest{
				PlayerId: "player1",
				QueueId:  enqueueResp.QueueId,
			},
			wantErr:   false,
			wantFound: true,
		},
		{
			name: "cancel non-existent player",
			req: &matchmakingpb.CancelQueueRequest{
				PlayerId: "nonexistent",
				QueueId:  "queue123",
			},
			wantErr:   false,
			wantFound: false,
		},
		{
			name: "empty player ID",
			req: &matchmakingpb.CancelQueueRequest{
				PlayerId: "",
				QueueId:  "queue123",
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := server.CancelQueue(context.Background(), tt.req)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if st, ok := status.FromError(err); ok {
					if st.Code() != tt.errCode {
						t.Errorf("Expected error code %v, got %v", tt.errCode, st.Code())
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if resp == nil {
					t.Error("Expected response but got nil")
				}
				if resp.Success != tt.wantFound {
					t.Errorf("Expected success %v, got %v", tt.wantFound, resp.Success)
				}
			}
		})
	}
}

func TestServer_GetQueueStatus(t *testing.T) {
	server := NewServer(&mockGamesClient{}, "redis:6379")

	// Enqueue a player
	enqueueReq := &matchmakingpb.EnqueueRequest{
		Player: &matchmakingpb.Player{
			Id:   "player1",
			Name: "Alice",
		},
	}
	_, err := server.Enqueue(context.Background(), enqueueReq)
	if err != nil {
		t.Fatalf("Failed to enqueue player: %v", err)
	}

	tests := []struct {
		name         string
		req          *matchmakingpb.GetQueueStatusRequest
		wantErr      bool
		errCode      codes.Code
		wantStatus   matchmakingpb.QueueStatus
		wantPosition int32
	}{
		{
			name: "valid status check",
			req: &matchmakingpb.GetQueueStatusRequest{
				PlayerId: "player1",
			},
			wantErr:      false,
			wantStatus:   matchmakingpb.QueueStatus_QUEUED,
			wantPosition: 1,
		},
		{
			name: "non-existent player",
			req: &matchmakingpb.GetQueueStatusRequest{
				PlayerId: "nonexistent",
			},
			wantErr:      false,
			wantStatus:   matchmakingpb.QueueStatus_CANCELLED,
			wantPosition: 0,
		},
		{
			name: "empty player ID",
			req: &matchmakingpb.GetQueueStatusRequest{
				PlayerId: "",
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := server.GetQueueStatus(context.Background(), tt.req)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if st, ok := status.FromError(err); ok {
					if st.Code() != tt.errCode {
						t.Errorf("Expected error code %v, got %v", tt.errCode, st.Code())
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if resp == nil {
					t.Error("Expected response but got nil")
				}
				if resp.Status != tt.wantStatus {
					t.Errorf("Expected status %v, got %v", tt.wantStatus, resp.Status)
				}
				if resp.QueuePosition != tt.wantPosition {
					t.Errorf("Expected position %d, got %d", tt.wantPosition, resp.QueuePosition)
				}
			}
		})
	}
}

func TestServer_AutoMatching(t *testing.T) {
	// Create a mock games client that tracks game creation
	var createdGames []string
	mockClient := &mockGamesClient{
		createGameFunc: func(ctx context.Context, req *gamespb.CreateGameRequest) (*gamespb.CreateGameResponse, error) {
			gameID := "game-" + req.Player1Id + "-" + req.Player2Id
			createdGames = append(createdGames, gameID)
			return &gamespb.CreateGameResponse{
				Game: &gamespb.Game{
					Id:        gameID,
					Player1Id: req.Player1Id,
					Player2Id: req.Player2Id,
				},
			}, nil
		},
	}

	server := NewServer(mockClient, "redis:6379")

	// Enqueue two players
	players := []*matchmakingpb.Player{
		{Id: "player1", Name: "Alice"},
		{Id: "player2", Name: "Bob"},
	}

	for _, player := range players {
		req := &matchmakingpb.EnqueueRequest{Player: player}
		_, err := server.Enqueue(context.Background(), req)
		if err != nil {
			t.Fatalf("Failed to enqueue player %s: %v", player.Id, err)
		}
	}

	// Wait a bit for the background matching process
	time.Sleep(2 * time.Second)

	// Check that a game was created
	if len(createdGames) != 1 {
		t.Errorf("Expected 1 game to be created, got %d", len(createdGames))
	}

	// Check that both players are no longer in queue
	for _, player := range players {
		req := &matchmakingpb.GetQueueStatusRequest{PlayerId: player.Id}
		resp, err := server.GetQueueStatus(context.Background(), req)
		if err != nil {
			t.Errorf("Failed to get status for player %s: %v", player.Id, err)
			continue
		}
		if resp.Status == matchmakingpb.QueueStatus_QUEUED {
			t.Errorf("Player %s should not be queued after matching", player.Id)
		}
	}
}

func TestServer_EnqueueMultiplePlayers(t *testing.T) {
	server := NewServer(&mockGamesClient{}, "redis:6379")

	// Enqueue multiple players
	playerCount := 5
	for i := 0; i < playerCount; i++ {
		req := &matchmakingpb.EnqueueRequest{
			Player: &matchmakingpb.Player{
				Id:   fmt.Sprintf("player%d", i),
				Name: fmt.Sprintf("Player%d", i),
			},
		}
		_, err := server.Enqueue(context.Background(), req)
		if err != nil {
			t.Fatalf("Failed to enqueue player%d: %v", i, err)
		}
	}

	// Check queue positions
	for i := 0; i < playerCount; i++ {
		req := &matchmakingpb.GetQueueStatusRequest{
			PlayerId: fmt.Sprintf("player%d", i),
		}
		resp, err := server.GetQueueStatus(context.Background(), req)
		if err != nil {
			t.Errorf("Failed to get status for player%d: %v", i, err)
			continue
		}

		expectedPosition := int32(i + 1)
		if resp.QueuePosition != expectedPosition {
			t.Errorf("Expected position %d for player%d, got %d", expectedPosition, i, resp.QueuePosition)
		}
	}
}

func TestServer_ReenqueueSamePlayer(t *testing.T) {
	server := NewServer(&mockGamesClient{}, "redis:6379")

	player := &matchmakingpb.Player{
		Id:   "player1",
		Name: "Alice",
	}

	// Enqueue the player twice
	req := &matchmakingpb.EnqueueRequest{Player: player}

	resp1, err := server.Enqueue(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed first enqueue: %v", err)
	}

	resp2, err := server.Enqueue(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed second enqueue: %v", err)
	}

	// Queue IDs should be different
	if resp1.QueueId == resp2.QueueId {
		t.Error("Expected different queue IDs for re-enqueue")
	}

	// Player should still be at position 1 (only one instance in queue)
	statusReq := &matchmakingpb.GetQueueStatusRequest{PlayerId: "player1"}
	statusResp, err := server.GetQueueStatus(context.Background(), statusReq)
	if err != nil {
		t.Fatalf("Failed to get queue status: %v", err)
	}

	if statusResp.QueuePosition != 1 {
		t.Errorf("Expected position 1 after re-enqueue, got %d", statusResp.QueuePosition)
	}
}
