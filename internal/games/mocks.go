package games

import (
	"context"

	enginepb "github.com/laerson/mancala/proto/engine"
	gamespb "github.com/laerson/mancala/proto/games"
	"google.golang.org/grpc"
)

type MockStorage struct {
	games map[string]*gamespb.Game
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		games: make(map[string]*gamespb.Game),
	}
}

func (m *MockStorage) SaveGame(ctx context.Context, game *gamespb.Game) error {
	m.games[game.Id] = game
	return nil
}

func (m *MockStorage) GetGame(ctx context.Context, gameID string) (*gamespb.Game, error) {
	game, exists := m.games[gameID]
	if !exists {
		return nil, &gameNotFoundError{}
	}
	return game, nil
}

func (m *MockStorage) DeleteGame(ctx context.Context, gameID string) error {
	delete(m.games, gameID)
	return nil
}

type gameNotFoundError struct{}

func (e *gameNotFoundError) Error() string {
	return "game not found"
}

type MockEngineClient struct {
	moveResponse *enginepb.MoveResponse
	moveError    error
}

func NewMockEngineClient() *MockEngineClient {
	return &MockEngineClient{}
}

func (m *MockEngineClient) SetMoveResponse(response *enginepb.MoveResponse) {
	m.moveResponse = response
}

func (m *MockEngineClient) SetMoveError(err error) {
	m.moveError = err
}

func (m *MockEngineClient) Move(ctx context.Context, req *enginepb.MoveRequest, opts ...grpc.CallOption) (*enginepb.MoveResponse, error) {
	if m.moveError != nil {
		return nil, m.moveError
	}
	return m.moveResponse, nil
}
