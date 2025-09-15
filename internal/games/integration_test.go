package games

import (
	"context"
	"net"
	"testing"

	enginepb "github.com/laerson/mancala/proto/engine"
	gamespb "github.com/laerson/mancala/proto/games"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

func setupIntegrationTest(t *testing.T) (gamespb.GamesClient, func()) {
	ctx := context.Background()

	redisContainer, err := redis.Run(ctx, "redis:7-alpine")
	if err != nil {
		t.Fatalf("failed to start redis container: %v", err)
	}

	host, err := redisContainer.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get redis host: %v", err)
	}

	port, err := redisContainer.MappedPort(ctx, "6379")
	if err != nil {
		t.Fatalf("failed to get redis port: %v", err)
	}

	addr := host + ":" + port.Port()
	storage := NewRedisStorage(addr, "", 0)

	mockEngineClient := NewMockEngineClient()
	mockEngineClient.SetMoveResponse(&enginepb.MoveResponse{
		Result: &enginepb.MoveResponse_MoveResult{
			MoveResult: &enginepb.MoveResult{
				Board: &enginepb.Board{
					Pits: []uint32{0, 5, 5, 5, 5, 4, 0, 4, 4, 4, 4, 4, 4, 0},
				},
				CurrentPlayer: enginepb.Player_PLAYER_TWO,
				IsFinished:    false,
				Winner:        enginepb.Winner_NO_WINNER,
			},
		},
	})

	server := NewServer(storage, mockEngineClient)

	lis := bufconn.Listen(bufSize)
	grpcServer := grpc.NewServer()
	gamespb.RegisterGamesServer(grpcServer, server)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Errorf("Server exited with error: %v", err)
		}
	}()

	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	conn, err := grpc.NewClient("passthrough:///bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	client := gamespb.NewGamesClient(conn)

	cleanup := func() {
		conn.Close()
		grpcServer.Stop()
		if err := testcontainers.TerminateContainer(redisContainer); err != nil {
			t.Logf("failed to terminate redis container: %v", err)
		}
	}

	return client, cleanup
}

func TestIntegration_CreateAndMoveGame(t *testing.T) {
	client, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	createReq := &gamespb.CreateGameRequest{
		Player1Id: "player1",
		Player2Id: "player2",
	}

	createResp, err := client.Create(ctx, createReq)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if createResp.Game == nil {
		t.Fatal("Create() returned nil game")
	}

	gameID := createResp.Game.Id
	if gameID == "" {
		t.Fatal("Create() returned empty game ID")
	}

	moveReq := &gamespb.MakeGameMoveRequest{
		PlayerId: "player1",
		GameId:   gameID,
		PitIndex: 0,
	}

	moveResp, err := client.Move(ctx, moveReq)
	if err != nil {
		t.Fatalf("Move() error = %v", err)
	}

	moveResult, ok := moveResp.Result.(*gamespb.MakeGameMoveResponse_MoveResult)
	if !ok {
		t.Fatal("Move() should return MoveResult")
	}

	if moveResult.MoveResult.CurrentPlayer != enginepb.Player_PLAYER_TWO {
		t.Errorf("Move() CurrentPlayer = %v, want %v", moveResult.MoveResult.CurrentPlayer, enginepb.Player_PLAYER_TWO)
	}
}

func TestIntegration_CreateMultipleGames(t *testing.T) {
	client, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	game1Req := &gamespb.CreateGameRequest{
		Player1Id: "alice",
		Player2Id: "bob",
	}

	game2Req := &gamespb.CreateGameRequest{
		Player1Id: "charlie",
		Player2Id: "diana",
	}

	game1Resp, err := client.Create(ctx, game1Req)
	if err != nil {
		t.Fatalf("Create() game1 error = %v", err)
	}

	game2Resp, err := client.Create(ctx, game2Req)
	if err != nil {
		t.Fatalf("Create() game2 error = %v", err)
	}

	if game1Resp.Game.Id == game2Resp.Game.Id {
		t.Error("Create() should generate unique game IDs")
	}

	if game1Resp.Game.Player1Id != "alice" {
		t.Errorf("Game1 Player1Id = %v, want alice", game1Resp.Game.Player1Id)
	}
	if game2Resp.Game.Player1Id != "charlie" {
		t.Errorf("Game2 Player1Id = %v, want charlie", game2Resp.Game.Player1Id)
	}
}

func TestIntegration_InvalidMoveRequests(t *testing.T) {
	client, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	createReq := &gamespb.CreateGameRequest{
		Player1Id: "player1",
		Player2Id: "player2",
	}

	createResp, err := client.Create(ctx, createReq)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	gameID := createResp.Game.Id

	tests := []struct {
		name       string
		request    *gamespb.MakeGameMoveRequest
		wantErrMsg string
	}{
		{
			name: "Empty player ID",
			request: &gamespb.MakeGameMoveRequest{
				PlayerId: "",
				GameId:   gameID,
				PitIndex: 0,
			},
			wantErrMsg: "player ID and game ID are required",
		},
		{
			name: "Invalid game ID",
			request: &gamespb.MakeGameMoveRequest{
				PlayerId: "player1",
				GameId:   "invalid-game-id",
				PitIndex: 0,
			},
			wantErrMsg: "game not found",
		},
		{
			name: "Player not in game",
			request: &gamespb.MakeGameMoveRequest{
				PlayerId: "player3",
				GameId:   gameID,
				PitIndex: 0,
			},
			wantErrMsg: "player is not part of this game",
		},
		{
			name: "Wrong turn",
			request: &gamespb.MakeGameMoveRequest{
				PlayerId: "player2",
				GameId:   gameID,
				PitIndex: 7,
			},
			wantErrMsg: "it's not your turn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			moveResp, err := client.Move(ctx, tt.request)
			if err != nil {
				t.Fatalf("Move() error = %v", err)
			}

			errorResult, ok := moveResp.Result.(*gamespb.MakeGameMoveResponse_Error)
			if !ok {
				t.Fatal("Move() should return Error")
			}

			if errorResult.Error.Message != tt.wantErrMsg {
				t.Errorf("Move() error message = %v, want %v", errorResult.Error.Message, tt.wantErrMsg)
			}
		})
	}
}
