package services

import (
	"hsm/internal/client"
	"sync"
	"time"
)

// UserSessionService manages game sessions for multi-user mode.
// Each user (JWT subject) can only have one active session at a time.
type UserSessionService struct {
	sessionService *SessionService
	sessions       map[string]*client.GameSession // subject -> session
	mu             sync.RWMutex
}

func NewUserSessionService(sessionService *SessionService) *UserSessionService {
	return &UserSessionService{
		sessionService: sessionService,
		sessions:       make(map[string]*client.GameSession),
	}
}

// GetSession returns the session for a subject
func (s *UserSessionService) GetSession(subject string) *client.GameSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[subject]
}

// GetOrCreateSession returns an existing active session or creates a new one.
// If an active session exists, it refreshes and returns it.
func (s *UserSessionService) GetOrCreateSession(subject string) (*client.GameSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for existing session
	if session, exists := s.sessions[subject]; exists {
		if time.Now().Before(session.ExpiresAt) {
			// Refresh and return
			if refreshed, err := s.sessionService.RefreshGameSession(session.SessionToken); err == nil {
				s.sessions[subject] = refreshed
				return refreshed, nil
			}
		}
		// Session expired/invalid, clean up
		delete(s.sessions, subject)
	}

	// Create new session
	session, err := s.sessionService.CreateGameSession()
	if err != nil {
		return nil, err
	}

	s.sessions[subject] = session
	return session, nil
}

// DeleteSession deletes the session for a subject
func (s *UserSessionService) DeleteSession(subject string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[subject]
	if !exists {
		return nil
	}

	if err := s.sessionService.DeleteGameSession(session.SessionToken); err != nil {
		return err
	}

	delete(s.sessions, subject)
	return nil
}

// RefreshSession refreshes the session for a subject
func (s *UserSessionService) RefreshSession(subject string) (*client.GameSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[subject]
	if !exists {
		return nil, nil
	}

	refreshed, err := s.sessionService.RefreshGameSession(session.SessionToken)
	if err != nil {
		return nil, err
	}

	s.sessions[subject] = refreshed
	return refreshed, nil
}
