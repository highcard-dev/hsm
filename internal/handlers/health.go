package handlers

import (
	"net/http"

	"hsm/api"
	"hsm/internal/utils"
)

// GetHealth returns the health status
// (GET /health)
func (s *Server) GetHealth(w http.ResponseWriter, r *http.Request) {
	response := api.HealthResponse{
		Status:  "healthy",
		Version: s.version,
	}
	utils.WriteJSON(w, http.StatusOK, response)
}
