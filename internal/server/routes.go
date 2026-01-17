package server

import (
	"net/http"

	"hsm/internal/handlers"
	"hsm/internal/middleware"
	"hsm/internal/services"
)

const version = "1.0.0"

// SetupRoutes configures all HTTP routes
func SetupRoutes(sessionService *services.SessionService, downloadService *services.DownloadService, jwksEndpoint string) http.Handler {
	mux := http.NewServeMux()

	healthHandler := handlers.NewHealthHandler(version)
	downloadHandler := handlers.NewDownloadHandler(downloadService)

	// Create appropriate session handler based on mode
	var sessionHandler *handlers.SessionHandler
	if jwksEndpoint != "" {
		// Multi-user mode: use UserSessionService for subject tracking
		userSessionService := services.NewUserSessionService(sessionService)
		sessionHandler = handlers.NewMultiUserSessionHandler(sessionService, userSessionService)
	} else {
		// Single-user mode
		sessionHandler = handlers.NewSessionHandler(sessionService)
	}

	mux.HandleFunc("GET /health", healthHandler.Check)

	// Non-versioned routes (plain text responses)
	mux.HandleFunc("POST /game-session", sessionHandler.CreateEnv)
	mux.HandleFunc("GET /download", downloadHandler.GetDownloadURLPlain)
	mux.HandleFunc("GET /version", downloadHandler.GetVersionPlain)

	// Versioned API routes (JSON responses)
	mux.HandleFunc("GET /api/v1/session", sessionHandler.Get)
	mux.HandleFunc("POST /api/v1/session", sessionHandler.Create)
	mux.HandleFunc("POST /api/v1/session/refresh", sessionHandler.Refresh)
	mux.HandleFunc("DELETE /api/v1/session", sessionHandler.Delete)

	mux.HandleFunc("GET /api/v1/download/version", downloadHandler.GetVersion)
	mux.HandleFunc("GET /api/v1/download/url", downloadHandler.GetDownloadURL)

	var handler http.Handler = mux

	// Apply JWT authentication middleware if JWKS endpoint is configured
	if jwksEndpoint != "" {
		handler = middleware.JWTAuth(jwksEndpoint)(handler)
	}

	handler = middleware.Logging(handler)

	return handler
}
