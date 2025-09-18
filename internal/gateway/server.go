package gateway

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Server represents the API Gateway server
type Server struct {
	config  *GatewayConfig
	clients *ServiceClients
	cleanup func()
	router  *gin.Engine
	server  *http.Server
}

// NewServer creates a new API Gateway server
func NewServer(config *GatewayConfig) (*Server, error) {
	// Initialize service clients
	clients, cleanup, err := NewServiceClients(config.Services)
	if err != nil {
		return nil, err
	}

	// Create server instance
	server := &Server{
		config:  config,
		clients: clients,
		cleanup: cleanup,
	}

	// Setup routes
	server.setupRoutes()

	// Create HTTP server
	server.server = &http.Server{
		Addr:         ":" + config.Port,
		Handler:      server.router,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	return server, nil
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Set gin mode
	gin.SetMode(gin.ReleaseMode)

	s.router = gin.New()

	// Add middleware
	s.router.Use(CORSMiddleware())
	s.router.Use(LoggingMiddleware())
	s.router.Use(gin.Recovery())

	// Create handlers
	authHandlers := NewAuthHandlers(s.clients)
	matchmakingHandlers := NewMatchmakingHandlers(s.clients)
	gamesHandlers := NewGamesHandlers(s.clients)
	notificationsHandlers := NewNotificationsHandlers(s.clients)

	// JWT middleware
	jwtMiddleware := NewJWTMiddleware(s.config.JWTSecret, s.clients)

	// Health check endpoint
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"service":   "mancala-gateway",
		})
	})

	// API version 1 routes
	v1 := s.router.Group("/api/v1")

	// Authentication routes (public)
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/login", authHandlers.Login)
		authGroup.POST("/register", authHandlers.Register)
		authGroup.GET("/validate", authHandlers.ValidateToken)
	}

	// Protected routes (require authentication)
	protected := v1.Group("/")
	protected.Use(jwtMiddleware.RequireAuth())

	// Matchmaking routes
	matchmakingGroup := protected.Group("/matchmaking")
	{
		matchmakingGroup.POST("/enqueue", matchmakingHandlers.Enqueue)
		matchmakingGroup.POST("/bot", matchmakingHandlers.BotMatch)
		matchmakingGroup.DELETE("/queue/:player_id", matchmakingHandlers.CancelQueue)
		matchmakingGroup.GET("/queue/:player_id/status", matchmakingHandlers.GetQueueStatus)
	}

	// Games routes
	gamesGroup := protected.Group("/games")
	{
		gamesGroup.POST("/", gamesHandlers.CreateGame)
		gamesGroup.POST("/:game_id/move", gamesHandlers.MakeMove)
	}

	// Notifications routes (Server-Sent Events)
	notificationsGroup := protected.Group("/notifications")
	{
		notificationsGroup.GET("/subscribe/:player_id", notificationsHandlers.SubscribeToNotifications)
	}

	log.Printf("API Gateway routes configured")
}

// Start starts the API Gateway server
func (s *Server) Start() error {
	log.Printf("Starting API Gateway on port %s", s.config.Port)
	return s.server.ListenAndServe()
}

// Stop gracefully stops the API Gateway server
func (s *Server) Stop(ctx context.Context) error {
	log.Println("Stopping API Gateway...")

	// Shutdown HTTP server
	if err := s.server.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down HTTP server: %v", err)
	}

	// Cleanup gRPC connections
	if s.cleanup != nil {
		s.cleanup()
	}

	log.Println("API Gateway stopped")
	return nil
}
