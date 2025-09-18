package auth

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	authpb "github.com/laerson/mancala/proto/auth"
)

// AuthInterceptor provides JWT authentication for gRPC services
type AuthInterceptor struct {
	authClient authpb.AuthClient
	jwtManager *JWTManager
}

// NewAuthInterceptor creates a new auth interceptor
func NewAuthInterceptor(authClient authpb.AuthClient, jwtSecret string) *AuthInterceptor {
	return &AuthInterceptor{
		authClient: authClient,
		jwtManager: NewJWTManager(jwtSecret, 24*3600, 7*24*3600), // Same config as auth service
	}
}

// UnaryInterceptor returns a gRPC unary server interceptor for JWT authentication
func (interceptor *AuthInterceptor) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Skip authentication for health checks and internal service calls
		if interceptor.isExemptMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		// Extract and validate JWT token
		userID, err := interceptor.validateTokenFromContext(ctx)
		if err != nil {
			return nil, err
		}

		// Add user ID to context for use in handlers
		ctx = context.WithValue(ctx, "user_id", userID)
		return handler(ctx, req)
	}
}

// StreamInterceptor returns a gRPC stream server interceptor for JWT authentication
func (interceptor *AuthInterceptor) StreamInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Skip authentication for internal service calls
		if interceptor.isExemptMethod(info.FullMethod) {
			return handler(srv, stream)
		}

		// Extract and validate JWT token
		userID, err := interceptor.validateTokenFromContext(stream.Context())
		if err != nil {
			return err
		}

		// Create a new context with user ID
		ctx := context.WithValue(stream.Context(), "user_id", userID)
		wrappedStream := &wrappedServerStream{
			ServerStream: stream,
			ctx:          ctx,
		}

		return handler(srv, wrappedStream)
	}
}

// validateTokenFromContext extracts and validates JWT token from gRPC metadata
func (interceptor *AuthInterceptor) validateTokenFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	// Extract authorization header
	authHeader := md["authorization"]
	if len(authHeader) == 0 {
		return "", status.Errorf(codes.Unauthenticated, "missing authorization header")
	}

	// Extract token from "Bearer <token>" format
	token := authHeader[0]
	if !strings.HasPrefix(token, "Bearer ") {
		return "", status.Errorf(codes.Unauthenticated, "invalid authorization header format")
	}
	token = strings.TrimPrefix(token, "Bearer ")

	// Validate token locally first (faster)
	claims, err := interceptor.jwtManager.ValidateAccessToken(token)
	if err != nil {
		// If local validation fails, try auth service (handles revoked tokens, etc.)
		return interceptor.validateWithAuthService(ctx, token)
	}

	return claims.UserID, nil
}

// validateWithAuthService validates token via auth service
func (interceptor *AuthInterceptor) validateWithAuthService(ctx context.Context, token string) (string, error) {
	if interceptor.authClient == nil {
		return "", status.Errorf(codes.Unauthenticated, "invalid token")
	}

	resp, err := interceptor.authClient.ValidateToken(ctx, &authpb.ValidateTokenRequest{
		AccessToken: token,
	})
	if err != nil {
		return "", status.Errorf(codes.Unauthenticated, "token validation failed: %v", err)
	}

	if !resp.Valid {
		return "", status.Errorf(codes.Unauthenticated, "invalid token: %s", resp.Message)
	}

	return resp.User.UserId, nil
}

// isExemptMethod checks if a method should skip authentication
func (interceptor *AuthInterceptor) isExemptMethod(method string) bool {
	exemptMethods := []string{
		// Health checks
		"/grpc.health.v1.Health/Check",
		"/grpc.health.v1.Health/Watch",

		// Games service - Create is called by matchmaking service
		"/proto.games.Games/Create",

		// Auth service methods (obviously don't need auth)
		"/proto.auth.Auth/Register",
		"/proto.auth.Auth/Login",
		"/proto.auth.Auth/ValidateToken",
		"/proto.auth.Auth/RefreshToken",
	}

	for _, exempt := range exemptMethods {
		if method == exempt {
			return true
		}
	}
	return false
}

// wrappedServerStream wraps grpc.ServerStream with custom context
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// GetUserIDFromContext extracts user ID from context (helper for handlers)
func GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return "", fmt.Errorf("user ID not found in context")
	}
	return userID, nil
}

// ValidatePlayerOwnership checks if the authenticated user owns the specified player ID
func ValidatePlayerOwnership(ctx context.Context, playerID string) error {
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "authentication required")
	}

	if userID != playerID {
		return status.Errorf(codes.PermissionDenied, "player ID does not match authenticated user")
	}

	return nil
}
