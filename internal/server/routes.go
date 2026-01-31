package server

import (
	"net/http"

	"hsm/api"
	"hsm/internal/handlers"
	"hsm/internal/middleware"
	"hsm/internal/services"
)

const version = "1.0.0"

// SetupRoutes configures all HTTP routes using oapi-codegen generated handlers
func SetupRoutes(sessionService *services.SessionService, downloadService *services.DownloadService, jwksEndpoint string, jwksCACert string, jwksJWTToken string) http.Handler {
	// Create server with appropriate configuration
	var opts []handlers.ServerOption
	if jwksEndpoint != "" {
		// Multi-user mode: use UserSessionService for subject tracking
		userSessionService := services.NewUserSessionService(sessionService)
		opts = append(opts, handlers.WithUserSessionService(userSessionService))
	}

	server := handlers.NewServer(version, sessionService, downloadService, opts...)

	// Create the base handler with all routes
	baseHandler := api.HandlerWithOptions(server, api.StdHTTPServerOptions{})

	// Apply JWT authentication middleware if JWKS endpoint is configured
	var handler = baseHandler
	if jwksEndpoint != "" {
		handler = middleware.JWTAuthWithPublicPaths(jwksEndpoint, jwksCACert, jwksJWTToken, []string{"/health"})(handler)
	}

	return middleware.Logging(handler)
}
