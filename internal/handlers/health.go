package handlers

import (
	"hsm/internal/utils"
	"net/http"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// HealthHandler handles health check requests
type HealthHandler struct {
	version string
}

// NewHealthHandler creates a new HealthHandler
func NewHealthHandler(version string) *HealthHandler {
	return &HealthHandler{
		version: version,
	}
}

// Check returns the health status
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:  "healthy",
		Version: h.version,
	}
	utils.WriteJSON(w, http.StatusOK, response)
}

// Ready returns the readiness status
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:  "ready",
		Version: h.version,
	}
	utils.WriteJSON(w, http.StatusOK, response)
}
