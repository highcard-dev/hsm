package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"

	"hsm/internal/client"
	"hsm/internal/services"

	"github.com/spf13/cobra"
)

var additionalArgs []string
var additionalJavaArgs []string

// hasFlag checks if a flag is present in the args slice
func hasFlag(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag {
			return true
		}
	}
	return false
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Hytale server with a game session",
	Long:  "Create a game session and start the Hytale server with the session tokens set as environment variables.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create session service (handles session loading, refresh, and profile fetching)
		sessionPath := GetSessionLocation()
		sessionService, err := services.NewSessionService(client.New(), sessionPath)
		if err != nil {
			return fmt.Errorf("failed to initialize session: %w (run 'hsm login' first)", err)
		}
		defer sessionService.Close()

		// Create game session
		fmt.Println("Creating game session...")
		gameSession, err := sessionService.CreateGameSession()
		if err != nil {
			return fmt.Errorf("failed to create game session: %w", err)
		}

		fmt.Println("Game session created successfully!")

		// Build command args: Java args first, then -jar, then server args
		var cmdArgs []string

		// 1. Additional Java args (e.g., -Xmx4G, -XX:+UseG1GC)
		cmdArgs = append(cmdArgs, additionalJavaArgs...)

		// 2. -jar argument
		if !hasFlag(additionalJavaArgs, "-jar") {
			cmdArgs = append(cmdArgs, "-jar", "Server/HytaleServer.jar")
		}

		// 3. Server args (after -jar)
		if !hasFlag(additionalArgs, "--assets") {
			cmdArgs = append(cmdArgs, "--assets", "Assets.zip")
		}
		cmdArgs = append(cmdArgs, additionalArgs...)

		// Create the command
		serverCmd := exec.Command("java", cmdArgs...)
		serverCmd.Stdout = os.Stdout
		serverCmd.Stderr = os.Stderr
		serverCmd.Stdin = os.Stdin

		// Set environment variables
		serverCmd.Env = append(os.Environ(),
			"HYTALE_SERVER_SESSION_TOKEN="+gameSession.SessionToken,
			"HYTALE_SERVER_IDENTITY_TOKEN="+gameSession.IdentityToken,
		)

		// Start the server
		if err := serverCmd.Start(); err != nil {
			return fmt.Errorf("failed to start server: %w", err)
		}

		// Handle interrupt signals to forward them to the server
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			for sig := range sigChan {
				fmt.Printf("\nReceived signal %v, stopping server...\n", sig)
				if serverCmd.Process != nil {
					if err := serverCmd.Process.Signal(sig); err != nil {
						_ = serverCmd.Process.Kill()
					}
				}
			}
		}()

		// Wait for the server to exit
		waitErr := serverCmd.Wait()

		// Stop signal handling
		signal.Stop(sigChan)
		close(sigChan)
		wg.Wait()

		if waitErr != nil {
			fmt.Printf("Server exited with error: %v\n", waitErr)
		} else {
			fmt.Println("Server exited successfully")
		}

		// Clean up the game session
		fmt.Println("Terminating game session...")
		if err := sessionService.DeleteGameSession(gameSession.SessionToken); err != nil {
			fmt.Printf("Warning: failed to terminate game session: %v\n", err)
		} else {
			fmt.Println("Game session terminated successfully")
		}

		return nil
	},
}

func init() {
	startCmd.Flags().StringArrayVar(&additionalJavaArgs, "additional-java-args", nil, "Additional Java arguments (e.g., --additional-java-args '-Xmx4G' --additional-java-args '-XX:+UseG1GC')")
	startCmd.Flags().StringArrayVar(&additionalArgs, "additional-args", nil, "Additional arguments to pass to the server (can override defaults, e.g., --additional-args '--assets Custom.zip')")
	rootCmd.AddCommand(startCmd)
}
