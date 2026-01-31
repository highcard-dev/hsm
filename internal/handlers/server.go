package handlers

import (
	"hsm/internal/services"
)

// Server implements the api.ServerInterface
type Server struct {
	version            string
	sessionService     *services.SessionService
	userSessionService *services.UserSessionService
	downloadService    *services.DownloadService
}

// ServerOption is a functional option for configuring the Server
type ServerOption func(*Server)

// WithUserSessionService enables multi-user mode
func WithUserSessionService(userSessionService *services.UserSessionService) ServerOption {
	return func(s *Server) {
		s.userSessionService = userSessionService
	}
}

// NewServer creates a new Server that implements api.ServerInterface
func NewServer(
	version string,
	sessionService *services.SessionService,
	downloadService *services.DownloadService,
	opts ...ServerOption,
) *Server {
	s := &Server{
		version:         version,
		sessionService:  sessionService,
		downloadService: downloadService,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Server) isMultiUser() bool {
	return s.userSessionService != nil
}
