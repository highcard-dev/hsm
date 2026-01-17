package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"hsm/internal/client"
)

// DeviceFlowService handles OAuth2 device flow business logic
type DeviceFlowService struct {
	client *client.Client
}

// NewDeviceFlowService creates a new DeviceFlowService with a client
func NewDeviceFlowService(client *client.Client) *DeviceFlowService {
	return &DeviceFlowService{
		client: client,
	}
}

// Flow orchestrates the complete OAuth2 device flow:
// 1. Initiates device authorization
// 2. Polls for token exchange until authorized or expired
// Returns the session with tokens when complete
func (d *DeviceFlowService) Flow(ctx context.Context) (*client.Session, error) {
	// Step 1: Initiate device authorization
	deviceAuth, err := d.client.CreateDeviceFlow()
	if err != nil {
		return nil, fmt.Errorf("failed to initiate device flow: %w", err)
	}

	// Step 2: Poll for token exchange
	// Default interval is 5 seconds if not specified
	pollInterval := time.Duration(deviceAuth.Interval) * time.Second
	if deviceAuth.Interval == 0 {
		pollInterval = 5 * time.Second
	}

	// Create a context with timeout based on device authorization expiration
	pollCtx := ctx
	if deviceAuth.ExpiresIn > 0 {
		var cancel context.CancelFunc
		pollCtx, cancel = context.WithTimeout(ctx, time.Duration(deviceAuth.ExpiresIn)*time.Second)
		defer cancel()
	}

	log.Printf("Please visit %s to authorize the device", deviceAuth.VerificationURIComplete)

	// Poll for token with the polling logic in the service
	session, err := d.pollForToken(pollCtx, deviceAuth.DeviceCode, pollInterval)
	if err != nil {
		return nil, fmt.Errorf("failed to poll for token: %w", err)
	}

	return session, nil
}

// pollForToken implements the polling logic for device flow token exchange.
// It calls the client's ExchangeDeviceCodeForToken method repeatedly until
// the user authorizes or the context is cancelled.
func (d *DeviceFlowService) pollForToken(ctx context.Context, deviceCode string, interval time.Duration) (*client.Session, error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Make initial request immediately, then poll at intervals
	for {
		session, err := d.client.ExchangeDeviceCodeForToken(ctx, deviceCode)
		if err == nil {
			// Success!
			return session, nil
		}

		// Check for specific device flow errors
		if errors.Is(err, client.ErrAuthorizationPending) {
			// Wait for next poll interval
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-ticker.C:
				continue
			}
		}

		// For any other error, return it
		return nil, err
	}
}
