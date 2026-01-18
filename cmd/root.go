package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var sessionLocation string

var rootCmd = &cobra.Command{
	Use:   "hsm",
	Short: "HSM service",
	Long:  "HSM is a service that provides various functionalities.",
}

func init() {
	rootCmd.PersistentFlags().StringVar(&sessionLocation, "session-location", "", "Path to session.json file (default: ~/.config/hsm/session.json)")
}

// GetSessionLocation returns the session file path, using ~/.config/hsm/session.json as default
func GetSessionLocation() string {
	if sessionLocation != "" {
		return sessionLocation
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home dir cannot be determined
		return "session.json"
	}
	return filepath.Join(homeDir, ".config", "hsm", "session.json")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
