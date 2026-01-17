package handlers

import (
	"hsm/internal/client"
	"hsm/internal/middleware"
	"hsm/internal/services"
	"hsm/internal/utils"
	"log"
	"net/http"
)

type SessionHandler struct {
	sessionService     *services.SessionService
	userSessionService *services.UserSessionService
}

func NewSessionHandler(sessionService *services.SessionService) *SessionHandler {
	return &SessionHandler{sessionService: sessionService}
}

func NewMultiUserSessionHandler(sessionService *services.SessionService, userSessionService *services.UserSessionService) *SessionHandler {
	return &SessionHandler{
		sessionService:     sessionService,
		userSessionService: userSessionService,
	}
}

func (h *SessionHandler) isMultiUser() bool {
	return h.userSessionService != nil
}

// Create creates a new session
func (h *SessionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var session *client.GameSession
	var err error

	if h.isMultiUser() {
		subject, ok := middleware.GetSubjectFromContext(r.Context())
		if !ok {
			utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		session, err = h.userSessionService.GetOrCreateSession(subject)
	} else {
		session, err = h.sessionService.CreateGameSession()
	}

	if err != nil {
		log.Printf("failed to create game session: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	utils.WriteJSON(w, http.StatusOK, session)
}

// CreateEnv creates a new session and returns it in env format
func (h *SessionHandler) CreateEnv(w http.ResponseWriter, r *http.Request) {
	var session *client.GameSession
	var err error

	if h.isMultiUser() {
		subject, ok := middleware.GetSubjectFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		session, err = h.userSessionService.GetOrCreateSession(subject)
	} else {
		session, err = h.sessionService.CreateGameSession()
	}

	if err != nil {
		log.Printf("failed to create game session: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("HYTALE_SERVER_SESSION_TOKEN=\"" + session.SessionToken + "\"\n"))
	w.Write([]byte("HYTALE_SERVER_IDENTITY_TOKEN=\"" + session.IdentityToken + "\"\n"))
}

// Get returns the current session (multi-user only)
func (h *SessionHandler) Get(w http.ResponseWriter, r *http.Request) {
	if !h.isMultiUser() {
		utils.WriteJSON(w, http.StatusNotImplemented, map[string]string{"error": "not available in single-user mode"})
		return
	}

	subject, ok := middleware.GetSubjectFromContext(r.Context())
	if !ok {
		utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	session := h.userSessionService.GetSession(subject)
	if session == nil {
		utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "no session"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, session)
}

// Delete deletes a session
func (h *SessionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if h.isMultiUser() {
		subject, ok := middleware.GetSubjectFromContext(r.Context())
		if !ok {
			utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if err := h.userSessionService.DeleteSession(subject); err != nil {
			log.Printf("failed to delete game session: %v", err)
			utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	} else {
		token := r.URL.Query().Get("token")
		if token == "" {
			utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "token required"})
			return
		}
		if err := h.sessionService.DeleteGameSession(token); err != nil {
			log.Printf("failed to delete game session: %v", err)
			utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// Refresh refreshes a session
func (h *SessionHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var session *client.GameSession
	var err error

	if h.isMultiUser() {
		subject, ok := middleware.GetSubjectFromContext(r.Context())
		if !ok {
			utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		session, err = h.userSessionService.RefreshSession(subject)
	} else {
		token := r.URL.Query().Get("token")
		if token == "" {
			utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "token required"})
			return
		}
		session, err = h.sessionService.RefreshGameSession(token)
	}

	if err != nil {
		log.Printf("failed to refresh game session: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if session == nil {
		utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "no session"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, session)
}
