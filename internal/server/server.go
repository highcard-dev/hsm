package server

import (
	"log"
	"net/http"

	"hsm/internal/client"
	"hsm/internal/services"
)

// Config holds server configuration
type Config struct {
	Port         string
	JWKSEndpoint string
}

// Start initializes and starts the HTTP server
func Start(config Config) error {
	c := client.New()
	sessionService, err := services.NewSessionService(c)
	if err != nil {
		log.Fatalf("failed to create session service: %v", err)
	}
	downloadService := services.NewDownloadService(c)

	handler := SetupRoutes(sessionService, downloadService, config.JWKSEndpoint)

	if config.JWKSEndpoint != "" {
		log.Printf("Multi-user mode enabled with JWKS endpoint: %s", config.JWKSEndpoint)
	} else {
		log.Println("Single-user mode (no JWT validation)")
	}

	addr := ":" + config.Port
	log.Printf("Starting server on %s", addr)

	return http.ListenAndServe(addr, handler)
}
