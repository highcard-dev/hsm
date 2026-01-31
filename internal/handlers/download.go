package handlers

import (
	"hsm/internal/services"
	"hsm/internal/utils"
	"net/http"
)

type DownloadHandler struct {
	downloadService *services.DownloadService
}

func NewDownloadHandler(downloadService *services.DownloadService) *DownloadHandler {
	return &DownloadHandler{downloadService: downloadService}
}

// GetDownloadURL returns the signed download URL for a patchline
func (h *DownloadHandler) GetDownloadURL(w http.ResponseWriter, r *http.Request) {
	patchline := r.URL.Query().Get("patchline")
	if patchline == "" {
		patchline = services.PatchlineRelease
	}

	url, version, err := h.downloadService.GetDownloadURL(patchline)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"url": url, "version": version})
}

// GetVersionPlain returns the latest version as plain text
func (h *DownloadHandler) GetVersionPlain(w http.ResponseWriter, r *http.Request) {
	patchline := r.URL.Query().Get("patchline")
	if patchline == "" {
		patchline = services.PatchlineRelease
	}

	_, version, err := h.downloadService.GetDownloadURL(patchline)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(version))
}

// GetDownloadURLPlain returns the signed download URL as plain text
func (h *DownloadHandler) GetDownloadURLPlain(w http.ResponseWriter, r *http.Request) {
	patchline := r.URL.Query().Get("patchline")
	if patchline == "" {
		patchline = services.PatchlineRelease
	}

	url, _, err := h.downloadService.GetDownloadURL(patchline)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(url))
}
