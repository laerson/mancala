package matchmaking

import (
	"sync"
	"time"

	matchmakingpb "github.com/laerson/mancala/proto/matchmaking"
)

type QueuedPlayer struct {
	Player    *matchmakingpb.Player
	QueueID   string
	QueueTime time.Time
	Stream    matchmakingpb.Matchmaking_StreamUpdatesServer // Optional streaming connection
}

type PlayerQueue struct {
	mu      sync.Mutex
	players map[string]*QueuedPlayer // key: player_id
	queue   []*QueuedPlayer          // ordered queue
}

func NewPlayerQueue() *PlayerQueue {
	return &PlayerQueue{
		players: make(map[string]*QueuedPlayer),
		queue:   make([]*QueuedPlayer, 0),
	}
}

func (pq *PlayerQueue) Enqueue(player *matchmakingpb.Player, queueID string) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	// Remove player if already queued
	pq.removePlayerLocked(player.Id)

	queuedPlayer := &QueuedPlayer{
		Player:    player,
		QueueID:   queueID,
		QueueTime: time.Now(),
	}

	pq.players[player.Id] = queuedPlayer
	pq.queue = append(pq.queue, queuedPlayer)
}

func (pq *PlayerQueue) RemovePlayer(playerID string) bool {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	return pq.removePlayerLocked(playerID)
}

func (pq *PlayerQueue) removePlayerLocked(playerID string) bool {
	queuedPlayer, exists := pq.players[playerID]
	if !exists {
		return false
	}

	// Remove from map
	delete(pq.players, playerID)

	// Remove from queue slice
	for i, p := range pq.queue {
		if p.Player.Id == playerID {
			pq.queue = append(pq.queue[:i], pq.queue[i+1:]...)
			break
		}
	}

	// Notify via stream if connected
	if queuedPlayer.Stream != nil {
		queuedPlayer.Stream.Send(&matchmakingpb.MatchmakingUpdate{
			QueueId: queuedPlayer.QueueID,
			Status:  matchmakingpb.QueueStatus_CANCELLED,
			Update: &matchmakingpb.MatchmakingUpdate_QueueCancelled{
				QueueCancelled: &matchmakingpb.QueueCancelled{
					Reason: "Queue cancelled by player",
				},
			},
		})
	}

	return true
}

func (pq *PlayerQueue) GetPlayerStatus(playerID string) (*QueuedPlayer, int32) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	queuedPlayer, exists := pq.players[playerID]
	if !exists {
		return nil, -1
	}

	// Find position in queue
	position := int32(-1)
	for i, p := range pq.queue {
		if p.Player.Id == playerID {
			position = int32(i + 1) // 1-based position
			break
		}
	}

	return queuedPlayer, position
}

func (pq *PlayerQueue) TryMatchPlayers() (*QueuedPlayer, *QueuedPlayer) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if len(pq.queue) < 2 {
		return nil, nil
	}

	// Take first two players
	player1 := pq.queue[0]
	player2 := pq.queue[1]

	// Remove both players from queue
	pq.removePlayerLocked(player1.Player.Id)
	pq.removePlayerLocked(player2.Player.Id)

	return player1, player2
}

func (pq *PlayerQueue) SetPlayerStream(playerID string, stream matchmakingpb.Matchmaking_StreamUpdatesServer) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if queuedPlayer, exists := pq.players[playerID]; exists {
		queuedPlayer.Stream = stream
	}
}

func (pq *PlayerQueue) GetQueueLength() int {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	return len(pq.queue)
}
