package middleware

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const SubjectContextKey contextKey = "jwt_subject"

// bearerTokenTransport wraps an http.RoundTripper to add a bearer token to requests
type bearerTokenTransport struct {
	base      http.RoundTripper
	tokenFile string
}

func (t *bearerTokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Read token fresh each time (Kubernetes rotates tokens)
	token, err := os.ReadFile(t.tokenFile)
	if err != nil {
		return nil, err
	}
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(string(token)))
	return t.base.RoundTrip(req)
}

// JWTAuth creates JWT authentication middleware using JWKS
// caCertFile is optional - if provided, it will be used for TLS verification when fetching JWKS
// jwtTokenFile is optional - if provided, it will be used as bearer token when fetching JWKS (for Kubernetes API)
func JWTAuth(jwksURL string, caCertFile string, jwtTokenFile string) func(http.Handler) http.Handler {
	var jwks keyfunc.Keyfunc
	var err error

	// Build custom transport if CA cert or JWT token is provided
	var transport http.RoundTripper = http.DefaultTransport

	if caCertFile != "" {
		// Create TLS config with custom CA cert
		caCert, readErr := os.ReadFile(caCertFile)
		if readErr != nil {
			panic("failed to read CA cert file: " + readErr.Error())
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			panic("failed to parse CA cert file")
		}

		transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		}
	}

	if jwtTokenFile != "" {
		// Wrap transport with bearer token auth
		transport = &bearerTokenTransport{
			base:      transport,
			tokenFile: jwtTokenFile,
		}
	}

	if caCertFile != "" || jwtTokenFile != "" {
		httpClient := &http.Client{
			Transport: transport,
		}

		override := keyfunc.Override{
			Client: httpClient,
		}
		jwks, err = keyfunc.NewDefaultOverrideCtx(context.Background(), []string{jwksURL}, override)
	} else {
		jwks, err = keyfunc.NewDefaultCtx(context.Background(), []string{jwksURL})
	}

	if err != nil {
		panic("failed to create JWKS keyfunc: " + err.Error())
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract bearer token
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				http.Error(w, `{"error":"missing or invalid authorization header"}`, http.StatusUnauthorized)
				return
			}

			// Parse and validate token
			token, err := jwt.Parse(strings.TrimPrefix(auth, "Bearer "), jwks.Keyfunc)
			if err != nil || !token.Valid {
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}

			// Extract subject
			claims, _ := token.Claims.(jwt.MapClaims)
			subject, _ := claims["sub"].(string)
			if subject == "" {
				http.Error(w, `{"error":"missing subject"}`, http.StatusUnauthorized)
				return
			}

			// Add subject to context
			ctx := context.WithValue(r.Context(), SubjectContextKey, subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetSubjectFromContext extracts the JWT subject from the request context
func GetSubjectFromContext(ctx context.Context) (string, bool) {
	subject, ok := ctx.Value(SubjectContextKey).(string)
	return subject, ok
}
