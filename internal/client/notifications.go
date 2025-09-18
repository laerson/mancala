package client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// NotificationClient handles Server-Sent Events from the notifications service
type NotificationClient struct {
	baseURL string
	token   string
}

// Notification represents a notification from the server
type Notification struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	GameID    string                 `json:"game_id"`
	Timestamp int64                  `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// NewNotificationClient creates a new notification client
func NewNotificationClient(baseURL, token string) *NotificationClient {
	return &NotificationClient{
		baseURL: baseURL,
		token:   token,
	}
}

// Subscribe subscribes to notifications for a player
func (nc *NotificationClient) Subscribe(ctx context.Context, playerID string, callback func(Notification)) error {
	url := fmt.Sprintf("%s/api/v1/notifications/subscribe/%s", nc.baseURL, playerID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+nc.token)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to subscribe to notifications: %s", resp.Status)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			line := scanner.Text()
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")

				var notification Notification
				if err := json.Unmarshal([]byte(data), &notification); err != nil {
					// Skip malformed notifications
					continue
				}

				callback(notification)
			}
		}
	}

	return scanner.Err()
}
