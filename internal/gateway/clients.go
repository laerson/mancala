package gateway

import (
	"log"

	authpb "github.com/laerson/mancala/proto/auth"
	gamespb "github.com/laerson/mancala/proto/games"
	matchmakingpb "github.com/laerson/mancala/proto/matchmaking"
	notificationspb "github.com/laerson/mancala/proto/notifications"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ServiceClients holds all gRPC service clients
type ServiceClients struct {
	Auth          authpb.AuthClient
	Games         gamespb.GamesClient
	Matchmaking   matchmakingpb.MatchmakingClient
	Notifications notificationspb.NotificationsClient
}

// NewServiceClients creates and initializes all gRPC service clients
func NewServiceClients(config ServiceConfig) (*ServiceClients, func(), error) {
	clients := &ServiceClients{}
	var closers []func()

	// Connect to Auth service
	authConn, err := grpc.NewClient(config.AuthAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	clients.Auth = authpb.NewAuthClient(authConn)
	closers = append(closers, func() { authConn.Close() })

	// Connect to Games service
	gamesConn, err := grpc.NewClient(config.GamesAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	clients.Games = gamespb.NewGamesClient(gamesConn)
	closers = append(closers, func() { gamesConn.Close() })

	// Connect to Matchmaking service
	matchmakingConn, err := grpc.NewClient(config.MatchmakingAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	clients.Matchmaking = matchmakingpb.NewMatchmakingClient(matchmakingConn)
	closers = append(closers, func() { matchmakingConn.Close() })

	// Connect to Notifications service
	notificationsConn, err := grpc.NewClient(config.NotificationsAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	clients.Notifications = notificationspb.NewNotificationsClient(notificationsConn)
	closers = append(closers, func() { notificationsConn.Close() })

	// Create cleanup function
	cleanup := func() {
		for _, closer := range closers {
			closer()
		}
	}

	log.Printf("Connected to all backend services")
	return clients, cleanup, nil
}
