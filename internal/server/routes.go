package server

import (
	"net/http"

	"hsm/internal/handlers"
	"hsm/internal/middleware"
	"hsm/internal/services"
)

const version = "1.0.0"

// SetupRoutes configures all HTTP routes
func SetupRoutes(sessionService *services.SessionService, downloadService *services.DownloadService, jwksEndpoint string, jwksCACert string) http.Handler {
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

	// Protected routes (require auth when JWKS is configured)
	protectedMux := http.NewServeMux()

	// Non-versioned routes (plain text responses)
	protectedMux.HandleFunc("POST /game-session", sessionHandler.CreateEnv)
	protectedMux.HandleFunc("GET /download", downloadHandler.GetDownloadURLPlain)
	protectedMux.HandleFunc("GET /version", downloadHandler.GetVersionPlain)

	// Versioned API routes (JSON responses)
	protectedMux.HandleFunc("GET /api/v1/session", sessionHandler.Get)
	protectedMux.HandleFunc("POST /api/v1/session", sessionHandler.Create)
	protectedMux.HandleFunc("POST /api/v1/session/refresh", sessionHandler.Refresh)
	protectedMux.HandleFunc("DELETE /api/v1/session", sessionHandler.Delete)

	protectedMux.HandleFunc("GET /api/v1/download/version", downloadHandler.GetVersion)
	protectedMux.HandleFunc("GET /api/v1/download/url", downloadHandler.GetDownloadURL)

	var protectedHandler http.Handler = protectedMux

	// Apply JWT authentication middleware if JWKS endpoint is configured
	if jwksEndpoint != "" {
		protectedHandler = middleware.JWTAuth(jwksEndpoint, jwksCACert)(protectedHandler)
	}

	// Main mux combines public and protected routes
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler.Check)
	mux.Handle("/", protectedHandler)

	return middleware.Logging(mux)
}
