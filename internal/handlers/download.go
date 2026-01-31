package handlers

import (
	"net/http"

	"hsm/api"
	"hsm/internal/services"
	"hsm/internal/utils"
)

// GetDownloadURL returns the download URL as JSON
// (GET /api/v1/download)
func (s *Server) GetDownloadURL(w http.ResponseWriter, r *http.Request, params api.GetDownloadURLParams) {
	patchline := services.PatchlineRelease
	if params.Patchline != nil && *params.Patchline != "" {
		patchline = *params.Patchline
	}

	url, version, err := s.downloadService.GetDownloadURL(patchline)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, api.ErrorResponse{Error: err.Error()})
		return
	}

	utils.WriteJSON(w, http.StatusOK, api.DownloadResponse{Url: url, Version: version})
}

// GetDownloadURLPlain returns the download URL as plain text
// (GET /download)
func (s *Server) GetDownloadURLPlain(w http.ResponseWriter, r *http.Request, params api.GetDownloadURLPlainParams) {
	patchline := services.PatchlineRelease
	if params.Patchline != nil && *params.Patchline != "" {
		patchline = *params.Patchline
	}

	url, _, err := s.downloadService.GetDownloadURL(patchline)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(url))
}

// GetVersionPlain returns the version as plain text
// (GET /version)
func (s *Server) GetVersionPlain(w http.ResponseWriter, r *http.Request, params api.GetVersionPlainParams) {
	patchline := services.PatchlineRelease
	if params.Patchline != nil && *params.Patchline != "" {
		patchline = *params.Patchline
	}

	_, version, err := s.downloadService.GetDownloadURL(patchline)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(version))
}
