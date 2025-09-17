package matchmaking

import (
	"fmt"
	"testing"
	"time"

	matchmakingpb "github.com/laerson/mancala/proto/matchmaking"
)

func TestPlayerQueue_Enqueue(t *testing.T) {
	queue := NewPlayerQueue()
	player := &matchmakingpb.Player{
		Id:   "player1",
		Name: "Alice",
	}
	queueID := "queue123"

	queue.Enqueue(player, queueID)

	if queue.GetQueueLength() != 1 {
		t.Errorf("Expected queue length 1, got %d", queue.GetQueueLength())
	}

	queuedPlayer, position := queue.GetPlayerStatus("player1")
	if queuedPlayer == nil {
		t.Fatal("Expected player to be in queue")
	}

	if queuedPlayer.Player.Id != "player1" {
		t.Errorf("Expected player ID 'player1', got '%s'", queuedPlayer.Player.Id)
	}

	if position != 1 {
		t.Errorf("Expected position 1, got %d", position)
	}
}

func TestPlayerQueue_EnqueueDuplicate(t *testing.T) {
	queue := NewPlayerQueue()
	player := &matchmakingpb.Player{
		Id:   "player1",
		Name: "Alice",
	}

	// Enqueue the same player twice
	queue.Enqueue(player, "queue1")
	queue.Enqueue(player, "queue2")

	// Should only have one entry
	if queue.GetQueueLength() != 1 {
		t.Errorf("Expected queue length 1 after duplicate enqueue, got %d", queue.GetQueueLength())
	}

	queuedPlayer, _ := queue.GetPlayerStatus("player1")
	if queuedPlayer == nil {
		t.Fatal("Expected player to be in queue")
	}

	// Should have the latest queue ID
	if queuedPlayer.QueueID != "queue2" {
		t.Errorf("Expected queue ID 'queue2', got '%s'", queuedPlayer.QueueID)
	}
}

func TestPlayerQueue_RemovePlayer(t *testing.T) {
	queue := NewPlayerQueue()
	player := &matchmakingpb.Player{
		Id:   "player1",
		Name: "Alice",
	}

	queue.Enqueue(player, "queue1")

	if queue.GetQueueLength() != 1 {
		t.Errorf("Expected queue length 1, got %d", queue.GetQueueLength())
	}

	removed := queue.RemovePlayer("player1")
	if !removed {
		t.Error("Expected player to be removed")
	}

	if queue.GetQueueLength() != 0 {
		t.Errorf("Expected queue length 0 after removal, got %d", queue.GetQueueLength())
	}

	queuedPlayer, position := queue.GetPlayerStatus("player1")
	if queuedPlayer != nil {
		t.Error("Expected player to not be in queue after removal")
	}
	if position != -1 {
		t.Errorf("Expected position -1 for removed player, got %d", position)
	}
}

func TestPlayerQueue_RemoveNonExistentPlayer(t *testing.T) {
	queue := NewPlayerQueue()

	removed := queue.RemovePlayer("nonexistent")
	if removed {
		t.Error("Expected removal of non-existent player to return false")
	}
}

func TestPlayerQueue_TryMatchPlayers(t *testing.T) {
	queue := NewPlayerQueue()

	// Test with empty queue
	player1, player2 := queue.TryMatchPlayers()
	if player1 != nil || player2 != nil {
		t.Error("Expected no match with empty queue")
	}

	// Test with one player
	queue.Enqueue(&matchmakingpb.Player{Id: "player1", Name: "Alice"}, "queue1")
	player1, player2 = queue.TryMatchPlayers()
	if player1 != nil || player2 != nil {
		t.Error("Expected no match with only one player")
	}

	// Test with two players
	queue.Enqueue(&matchmakingpb.Player{Id: "player2", Name: "Bob"}, "queue2")
	player1, player2 = queue.TryMatchPlayers()

	if player1 == nil || player2 == nil {
		t.Fatal("Expected match with two players")
	}

	if player1.Player.Id != "player1" {
		t.Errorf("Expected first player to be 'player1', got '%s'", player1.Player.Id)
	}

	if player2.Player.Id != "player2" {
		t.Errorf("Expected second player to be 'player2', got '%s'", player2.Player.Id)
	}

	// Queue should be empty after matching
	if queue.GetQueueLength() != 0 {
		t.Errorf("Expected empty queue after matching, got length %d", queue.GetQueueLength())
	}
}

func TestPlayerQueue_QueueOrder(t *testing.T) {
	queue := NewPlayerQueue()

	// Enqueue players in order
	players := []*matchmakingpb.Player{
		{Id: "player1", Name: "Alice"},
		{Id: "player2", Name: "Bob"},
		{Id: "player3", Name: "Charlie"},
	}

	for i, player := range players {
		queue.Enqueue(player, fmt.Sprintf("queue%d", i+1))
	}

	// Check positions
	for i, player := range players {
		_, position := queue.GetPlayerStatus(player.Id)
		expectedPosition := int32(i + 1)
		if position != expectedPosition {
			t.Errorf("Expected position %d for %s, got %d", expectedPosition, player.Id, position)
		}
	}

	// Match first two players
	player1, player2 := queue.TryMatchPlayers()
	if player1.Player.Id != "player1" || player2.Player.Id != "player2" {
		t.Error("Expected to match first two players in queue order")
	}

	// Player3 should now be at position 1
	_, position := queue.GetPlayerStatus("player3")
	if position != 1 {
		t.Errorf("Expected player3 to be at position 1 after match, got %d", position)
	}
}

func TestPlayerQueue_QueueTime(t *testing.T) {
	queue := NewPlayerQueue()
	player := &matchmakingpb.Player{
		Id:   "player1",
		Name: "Alice",
	}

	beforeEnqueue := time.Now()
	queue.Enqueue(player, "queue1")
	afterEnqueue := time.Now()

	queuedPlayer, _ := queue.GetPlayerStatus("player1")
	if queuedPlayer == nil {
		t.Fatal("Expected player to be in queue")
	}

	queueTime := queuedPlayer.QueueTime
	if queueTime.Before(beforeEnqueue) || queueTime.After(afterEnqueue) {
		t.Error("Queue time should be between before and after enqueue times")
	}
}

func TestPlayerQueue_ConcurrentAccess(t *testing.T) {
	queue := NewPlayerQueue()
	done := make(chan bool, 2)

	// Simulate concurrent enqueuing
	go func() {
		for i := 0; i < 100; i++ {
			player := &matchmakingpb.Player{
				Id:   fmt.Sprintf("player%d", i),
				Name: fmt.Sprintf("Player%d", i),
			}
			queue.Enqueue(player, fmt.Sprintf("queue%d", i))
		}
		done <- true
	}()

	// Simulate concurrent matching
	go func() {
		for i := 0; i < 50; i++ {
			queue.TryMatchPlayers()
			time.Sleep(time.Millisecond)
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Queue should be consistent (no crashes/panics)
	length := queue.GetQueueLength()
	if length < 0 || length > 100 {
		t.Errorf("Unexpected queue length after concurrent access: %d", length)
	}
}
