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

// JWTAuth creates JWT authentication middleware using JWKS
// caCertFile is optional - if provided, it will be used for TLS verification when fetching JWKS
func JWTAuth(jwksURL string, caCertFile string) func(http.Handler) http.Handler {
	var jwks keyfunc.Keyfunc
	var err error

	if caCertFile != "" {
		// Create custom HTTP client with CA cert
		caCert, err := os.ReadFile(caCertFile)
		if err != nil {
			panic("failed to read CA cert file: " + err.Error())
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			panic("failed to parse CA cert file")
		}

		httpClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: caCertPool,
				},
			},
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
