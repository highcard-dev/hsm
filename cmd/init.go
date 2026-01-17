package cmd

import (
	"context"
	"fmt"
	"hsm/internal/client"
	"hsm/internal/services"
	"hsm/internal/utils"
	"os"

	"github.com/spf13/cobra"
)

var stdoutFlag bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Start the HTTP server",
	Long:  "Start the HSM HTTP server on the specified port.",
	RunE: func(cmd *cobra.Command, args []string) error {
		deviceFlow := services.NewDeviceFlowService(client.New())
		session, err := deviceFlow.Flow(context.Background())
		if err != nil {
			return fmt.Errorf("failed to initiate device flow: %w", err)
		}

		if stdoutFlag {
			fmt.Fprintln(os.Stdout, "")
			fmt.Fprintln(os.Stdout, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Fprintln(os.Stdout, "                        Session Created")
			fmt.Fprintln(os.Stdout, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Fprintln(os.Stdout, "")
			fmt.Fprintf(os.Stdout, "  Access Token:   %s\n", session.AccessToken)
			if session.RefreshToken != "" {
				fmt.Fprintf(os.Stdout, "  Refresh Token:  %s\n", session.RefreshToken)
			}
			if session.TokenType != "" {
				fmt.Fprintf(os.Stdout, "  Token Type:     %s\n", session.TokenType)
			}
			if session.Scope != "" {
				fmt.Fprintf(os.Stdout, "  Scope:          %s\n", session.Scope)
			}
			if !session.ExpiresAt.IsZero() {
				fmt.Fprintf(os.Stdout, "  Expires At:     %s\n", session.ExpiresAt.Format("2006-01-02 15:04:05 MST"))
			}
			fmt.Fprintln(os.Stdout, "")
			fmt.Fprintln(os.Stdout, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Fprintln(os.Stdout, "")
		} else {
			// Save session to JSON file by default
			if err := utils.SaveSessionToFile(services.SessionFileName, session); err != nil {
				return fmt.Errorf("failed to save session: %w", err)
			}

			fmt.Printf("Session saved to %s\n", services.SessionFileName)
		}

		return nil
	},
}

func init() {
	initCmd.Flags().BoolVar(&stdoutFlag, "stdout", false, "Output session token to stdout instead of saving to file")
}
