package handlers

import (
	"log"
	"net/http"

	"hsm/api"
	"hsm/internal/client"
	"hsm/internal/middleware"
	"hsm/internal/utils"
)

// GetSession returns the current session (multi-user only)
// (GET /api/v1/session)
func (s *Server) GetSession(w http.ResponseWriter, r *http.Request) {
	if !s.isMultiUser() {
		utils.WriteJSON(w, http.StatusNotImplemented, api.ErrorResponse{Error: "not available in single-user mode"})
		return
	}

	subject, ok := middleware.GetSubjectFromContext(r.Context())
	if !ok {
		utils.WriteJSON(w, http.StatusUnauthorized, api.ErrorResponse{Error: "unauthorized"})
		return
	}

	session := s.userSessionService.GetSession(subject)
	if session == nil {
		utils.WriteJSON(w, http.StatusNotFound, api.ErrorResponse{Error: "no session"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, toAPIGameSession(session))
}

// CreateSession creates a new session
// (POST /api/v1/session)
func (s *Server) CreateSession(w http.ResponseWriter, r *http.Request) {
	var session *client.GameSession
	var err error

	if s.isMultiUser() {
		subject, ok := middleware.GetSubjectFromContext(r.Context())
		if !ok {
			utils.WriteJSON(w, http.StatusUnauthorized, api.ErrorResponse{Error: "unauthorized"})
			return
		}
		session, err = s.userSessionService.GetOrCreateSession(subject)
	} else {
		session, err = s.sessionService.CreateGameSession()
	}

	if err != nil {
		log.Printf("failed to create game session: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, api.ErrorResponse{Error: err.Error()})
		return
	}

	utils.WriteJSON(w, http.StatusOK, toAPIGameSession(session))
}

// DeleteSession deletes a session
// (DELETE /api/v1/session)
func (s *Server) DeleteSession(w http.ResponseWriter, r *http.Request, params api.DeleteSessionParams) {
	if s.isMultiUser() {
		subject, ok := middleware.GetSubjectFromContext(r.Context())
		if !ok {
			utils.WriteJSON(w, http.StatusUnauthorized, api.ErrorResponse{Error: "unauthorized"})
			return
		}
		if err := s.userSessionService.DeleteSession(subject); err != nil {
			log.Printf("failed to delete game session: %v", err)
			utils.WriteJSON(w, http.StatusInternalServerError, api.ErrorResponse{Error: err.Error()})
			return
		}
	} else {
		if params.Token == nil || *params.Token == "" {
			utils.WriteJSON(w, http.StatusBadRequest, api.ErrorResponse{Error: "token required"})
			return
		}
		if err := s.sessionService.DeleteGameSession(*params.Token); err != nil {
			log.Printf("failed to delete game session: %v", err)
			utils.WriteJSON(w, http.StatusInternalServerError, api.ErrorResponse{Error: err.Error()})
			return
		}
	}

	utils.WriteJSON(w, http.StatusOK, api.MessageResponse{Message: "deleted"})
}

// RefreshSession refreshes a session
// (POST /api/v1/session/refresh)
func (s *Server) RefreshSession(w http.ResponseWriter, r *http.Request, params api.RefreshSessionParams) {
	var session *client.GameSession
	var err error

	if s.isMultiUser() {
		subject, ok := middleware.GetSubjectFromContext(r.Context())
		if !ok {
			utils.WriteJSON(w, http.StatusUnauthorized, api.ErrorResponse{Error: "unauthorized"})
			return
		}
		session, err = s.userSessionService.RefreshSession(subject)
	} else {
		if params.Token == nil || *params.Token == "" {
			utils.WriteJSON(w, http.StatusBadRequest, api.ErrorResponse{Error: "token required"})
			return
		}
		session, err = s.sessionService.RefreshGameSession(*params.Token)
	}

	if err != nil {
		log.Printf("failed to refresh game session: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, api.ErrorResponse{Error: err.Error()})
		return
	}

	if session == nil {
		utils.WriteJSON(w, http.StatusNotFound, api.ErrorResponse{Error: "no session"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, toAPIGameSession(session))
}

// CreateGameSessionEnv creates a session and returns it in env format
// (POST /game-session)
func (s *Server) CreateGameSessionEnv(w http.ResponseWriter, r *http.Request) {
	var session *client.GameSession
	var err error

	if s.isMultiUser() {
		subject, ok := middleware.GetSubjectFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		session, err = s.userSessionService.GetOrCreateSession(subject)
	} else {
		session, err = s.sessionService.CreateGameSession()
	}

	if err != nil {
		log.Printf("failed to create game session: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("HYTALE_SERVER_SESSION_TOKEN=\"" + session.SessionToken + "\"\n"))
	_, _ = w.Write([]byte("HYTALE_SERVER_IDENTITY_TOKEN=\"" + session.IdentityToken + "\"\n"))
}

// toAPIGameSession converts a client.GameSession to api.GameSession
func toAPIGameSession(session *client.GameSession) api.GameSession {
	return api.GameSession{
		SessionToken:  session.SessionToken,
		IdentityToken: session.IdentityToken,
		ExpiresAt:     session.ExpiresAt,
	}
}
