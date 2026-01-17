package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client provides methods to interact with the HSM API
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string // Optional authentication token
}

// New creates a new HSM client
func New() *Client {
	return &Client{
		baseURL: "https://oauth.accounts.hytale.com",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// WithToken sets an authentication token for the client
func (c *Client) WithToken(token string) *Client {
	c.token = token
	return c
}

// WithHTTPClient allows setting a custom HTTP client
func (c *Client) WithHTTPClient(httpClient *http.Client) *Client {
	c.httpClient = httpClient
	return c
}

// Session represents a session/token response
type Session struct {
	ID           string    `json:"id,omitempty"`
	Token        string    `json:"token,omitempty"`
	AccessToken  string    `json:"access_token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type,omitempty"`
	Scope        string    `json:"scope,omitempty"`
	ExpiresIn    int       `json:"expires_in,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
}

// NeedsRefresh returns true if the token is expired or will expire within the threshold
func (s *Session) NeedsRefresh(threshold time.Duration) bool {
	if s.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().Add(threshold).After(s.ExpiresAt)
}

// CreateSessionRequest represents a request to create a new session
type CreateSessionRequest struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	TTL      int    `json:"ttl,omitempty"` // Time to live in seconds
}

// DeviceAuthorizationResponse represents the response from the device authorization endpoint
type DeviceAuthorizationResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete,omitempty"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval,omitempty"`
}

// CreateDeviceFlow initiates the OAuth2 device authorization flow
func (c *Client) CreateDeviceFlow() (*DeviceAuthorizationResponse, error) {
	// Prepare form data
	data := url.Values{}
	data.Set("client_id", "hytale-server")
	data.Set("scope", "openid offline auth:server")

	body := bytes.NewBufferString(data.Encode())

	httpReq, err := http.NewRequest("POST", c.baseURL+"/oauth2/device/auth", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var deviceAuth DeviceAuthorizationResponse
	if err := json.NewDecoder(resp.Body).Decode(&deviceAuth); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &deviceAuth, nil
}

var (
	// ErrAuthorizationPending indicates the user hasn't authorized the device yet
	ErrAuthorizationPending = errors.New("authorization_pending")
)

// ExchangeDeviceCodeForToken makes a single request to exchange a device code for tokens.
// Returns a Session on success, or ErrAuthorizationPending/ErrSlowDown for special conditions.
func (c *Client) ExchangeDeviceCodeForToken(ctx context.Context, deviceCode string) (*Session, error) {
	// Prepare form data for token exchange
	data := url.Values{}
	data.Set("client_id", "hytale-server")
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	data.Set("device_code", deviceCode)

	body := bytes.NewBufferString(data.Encode())

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/oauth2/token", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		var errorResp struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		if json.Unmarshal(bodyBytes, &errorResp) == nil {
			if errorResp.Error == "authorization_pending" {
				return nil, ErrAuthorizationPending
			}
		}
		return nil, fmt.Errorf("token request failed: %s", string(bodyBytes))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Success response
	session := &Session{
		CreatedAt: time.Now(),
	}

	if err := json.Unmarshal(bodyBytes, session); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Set Token field to match AccessToken for backward compatibility
	session.Token = session.AccessToken

	// Calculate ExpiresAt from ExpiresIn if provided
	if session.ExpiresIn > 0 {
		session.ExpiresAt = time.Now().Add(time.Duration(session.ExpiresIn) * time.Second)
	}

	return session, nil
}

// CreateSession creates a new session and returns authentication tokens
func (c *Client) CreateSession(req *CreateSessionRequest) (*Session, error) {
	var body io.Reader
	if req != nil {
		jsonData, err := json.Marshal(req)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/api/v1/session", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var session Session
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &session, nil
}

// ListSessions retrieves all sessions for the authenticated user
func (c *Client) ListSessions() ([]Session, error) {
	httpReq, err := http.NewRequest("GET", c.baseURL+"/api/v1/session", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var sessions []Session
	if err := json.NewDecoder(resp.Body).Decode(&sessions); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return sessions, nil
}

// DeleteSession deletes a session by ID
func (c *Client) DeleteSession(sessionID string) error {
	url := c.baseURL + "/api/v1/session"
	if sessionID != "" {
		url += "?id=" + sessionID
	}

	httpReq, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if c.token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// RefreshAccessToken refreshes an access token using a refresh token via OAuth2
func (c *Client) RefreshAccessToken(refreshToken string) (*Session, error) {
	// Use OAuth2 refresh_token grant type
	data := url.Values{}
	data.Set("client_id", "hytale-server")
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	body := bytes.NewBufferString(data.Encode())

	httpReq, err := http.NewRequest("POST", c.baseURL+"/oauth2/token", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	session := &Session{
		CreatedAt: time.Now(),
	}

	if err := json.Unmarshal(bodyBytes, session); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Set Token field to match AccessToken for backward compatibility
	session.Token = session.AccessToken

	// Calculate ExpiresAt from ExpiresIn if provided
	if session.ExpiresIn > 0 {
		session.ExpiresAt = time.Now().Add(time.Duration(session.ExpiresIn) * time.Second)
	}

	return session, nil
}

// Profile represents a Hytale game profile
type Profile struct {
	UUID     string `json:"uuid"`
	Username string `json:"username"`
}

// ProfilesResponse represents the response from get-profiles
type ProfilesResponse struct {
	Owner    string    `json:"owner"`
	Profiles []Profile `json:"profiles"`
}

// GetProfiles retrieves available game profiles for the authenticated user
func (c *Client) GetProfiles() (*ProfilesResponse, error) {
	httpReq, err := http.NewRequest("GET", "https://account-data.hytale.com/my-account/get-profiles", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var profiles ProfilesResponse
	if err := json.NewDecoder(resp.Body).Decode(&profiles); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &profiles, nil
}

// GameSession represents a Hytale game session
type GameSession struct {
	SessionToken  string    `json:"sessionToken"`
	IdentityToken string    `json:"identityToken"`
	ExpiresAt     time.Time `json:"expiresAt"`
}

// CreateGameSession creates a new game session for a specific profile UUID
func (c *Client) CreateGameSession(profileUUID string) (*GameSession, error) {
	payload := map[string]string{"uuid": profileUUID}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", "https://sessions.hytale.com/game-session/new", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var gameSession GameSession
	if err := json.NewDecoder(resp.Body).Decode(&gameSession); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &gameSession, nil
}

// RefreshGameSession refreshes an existing game session using its session token
func (c *Client) RefreshGameSession(sessionToken string) (*GameSession, error) {
	httpReq, err := http.NewRequest("POST", "https://sessions.hytale.com/game-session/refresh", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+sessionToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var gameSession GameSession
	if err := json.NewDecoder(resp.Body).Decode(&gameSession); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &gameSession, nil
}

// TerminateGameSession terminates a game session using its session token
func (c *Client) TerminateGameSession(sessionToken string) error {
	httpReq, err := http.NewRequest("DELETE", "https://sessions.hytale.com/game-session", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+sessionToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// SignedURLResponse represents the response containing a signed download URL
type SignedURLResponse struct {
	URL string `json:"url"`
}

// GetSignedURL fetches the signed download URL for a patchline
func (c *Client) GetSignedURL(patchline string) (*SignedURLResponse, error) {
	url := fmt.Sprintf("https://account-data.hytale.com/game-assets/version/%s.json", patchline)

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result SignedURLResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
