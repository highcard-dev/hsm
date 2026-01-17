package client_test

import (
	"fmt"
	"hsm/internal/client"
)

func ExampleClient_CreateSession() {
	// Create a new client
	cl := client.New()

	// Create a new session with credentials
	session, err := cl.CreateSession(&client.CreateSessionRequest{
		Username: "user@example.com",
		Password: "password123",
		TTL:      3600, // 1 hour
	})
	if err != nil {
		fmt.Printf("Error creating session: %v\n", err)
		return
	}

	fmt.Printf("Session created! Access Token: %s\n", session.AccessToken)
	fmt.Printf("Refresh Token: %s\n", session.RefreshToken)
}

func ExampleClient_ListSessions() {
	// Create a client with authentication token
	cl := client.New().WithToken("your-access-token")

	// List all sessions
	sessions, err := cl.ListSessions()
	if err != nil {
		fmt.Printf("Error listing sessions: %v\n", err)
		return
	}

	fmt.Printf("Found %d sessions\n", len(sessions))
	for _, session := range sessions {
		fmt.Printf("Session ID: %s, Created: %v\n", session.ID, session.CreatedAt)
	}
}

func ExampleClient_DeleteSession() {
	// Create a client with authentication token
	cl := client.New().WithToken("your-access-token")

	// Delete a session by ID
	err := cl.DeleteSession("session-id-123")
	if err != nil {
		fmt.Printf("Error deleting session: %v\n", err)
		return
	}

	fmt.Println("Session deleted successfully")
}

func ExampleClient_RefreshAccessToken() {
	cl := client.New()

	// Refresh an access token using a refresh token (OAuth2)
	session, err := cl.RefreshAccessToken("your-refresh-token")
	if err != nil {
		fmt.Printf("Error refreshing token: %v\n", err)
		return
	}

	fmt.Printf("Token refreshed! New Access Token: %s\n", session.AccessToken)
}
