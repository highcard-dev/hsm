package services

import (
	"context"
	"fmt"
	"hsm/internal/client"
	"hsm/internal/utils"
	"log"
	"sync"
	"time"
)

const (
	SessionFileName      = "session.json"
	RefreshThreshold     = 5 * time.Minute
	RefreshCheckInterval = 1 * time.Minute
)

// SessionService is a stateless service that wraps the Hytale API client.
// It handles OAuth token management but does not track game sessions.
type SessionService struct {
	client      *client.Client
	profileId   string
	sessionPath string
	session     *client.Session
	mu          sync.RWMutex
	cancel      context.CancelFunc
}

func NewSessionService(c *client.Client) (*SessionService, error) {
	svc := &SessionService{
		client:      c,
		sessionPath: SessionFileName,
	}

	if err := svc.loadAndRefreshSession(); err != nil {
		return nil, err
	}

	c.WithToken(svc.session.Token)

	profiles, err := c.GetProfiles()
	if err != nil {
		return nil, fmt.Errorf("failed to get profiles: %w", err)
	}
	if len(profiles.Profiles) == 0 {
		return nil, fmt.Errorf("no profiles found")
	}
	svc.profileId = profiles.Profiles[0].UUID

	ctx, cancel := context.WithCancel(context.Background())
	svc.cancel = cancel
	go svc.keepSessionFresh(ctx)

	return svc, nil
}

func (s *SessionService) loadAndRefreshSession() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, err := utils.ReadSessionFromFile(s.sessionPath)
	if err != nil {
		return err
	}
	if session.Token == "" {
		return fmt.Errorf("invalid session: missing token")
	}

	if session.NeedsRefresh(RefreshThreshold) {
		if session.RefreshToken == "" {
			return fmt.Errorf("session expired and no refresh token available")
		}
		newSession, err := s.client.RefreshAccessToken(session.RefreshToken)
		if err != nil {
			return fmt.Errorf("failed to refresh token: %w", err)
		}
		if newSession.RefreshToken == "" {
			newSession.RefreshToken = session.RefreshToken
		}
		if err := utils.SaveSessionToFile(s.sessionPath, newSession); err != nil {
			return fmt.Errorf("failed to save session: %w", err)
		}
		session = newSession
	}

	s.session = session
	s.client.WithToken(session.Token)

	log.Println("Session loaded and refreshed")
	return nil
}

func (s *SessionService) keepSessionFresh(ctx context.Context) {
	ticker := time.NewTicker(RefreshCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.mu.RLock()
			needsRefresh := s.session != nil && s.session.NeedsRefresh(RefreshThreshold)
			s.mu.RUnlock()

			if needsRefresh {
				log.Println("Refreshing OAuth session...")
				if err := s.loadAndRefreshSession(); err != nil {
					log.Printf("Failed to refresh session: %v", err)
				}
			}
		}
	}
}

func (s *SessionService) Close() {
	if s.cancel != nil {
		s.cancel()
	}
}

// CreateGameSession creates a new game session via the API
func (s *SessionService) CreateGameSession() (*client.GameSession, error) {
	return s.client.CreateGameSession(s.profileId)
}

// DeleteGameSession terminates a game session via the API
func (s *SessionService) DeleteGameSession(sessionToken string) error {
	return s.client.TerminateGameSession(sessionToken)
}

// RefreshGameSession refreshes a game session via the API
func (s *SessionService) RefreshGameSession(sessionToken string) (*client.GameSession, error) {
	return s.client.RefreshGameSession(sessionToken)
}
