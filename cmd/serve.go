package cmd

import (
	"hsm/internal/server"

	"github.com/spf13/cobra"
)

var (
	port         string
	jwksEndpoint string
	jwksCACert   string
	jwksJWTToken string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server",
	Long:  "Start the HSM HTTP server on the specified port.",
	RunE: func(cmd *cobra.Command, args []string) error {
		config := server.Config{
			Port:         port,
			JWKSEndpoint: jwksEndpoint,
			JWKSCACert:   jwksCACert,
			JWKSJWTToken: jwksJWTToken,
			SessionPath:  GetSessionLocation(),
		}
		return server.Start(config)
	},
}

func init() {
	serveCmd.Flags().StringVarP(&port, "port", "p", "8080", "Port to listen on")
	serveCmd.Flags().StringVar(&jwksEndpoint, "jwks-endpoint", "", "JWKS endpoint URL for JWT validation (optional, enables multi-user mode)")
	serveCmd.Flags().StringVar(&jwksCACert, "jwks-ca-cert", "", "CA certificate file for JWKS endpoint TLS verification (optional, for Kubernetes)")
	serveCmd.Flags().StringVar(&jwksJWTToken, "jwks-jwt-token-file", "", "Path to JWT token file for JWKS endpoint authentication (optional, for Kubernetes service account)")
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(loginCmd)
}
