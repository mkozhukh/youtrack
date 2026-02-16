package mcp

import (
	"context"
	"net/http"
	"strings"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// AuthTokenKey is the context key for the auth token
	AuthTokenKey contextKey = "youtrack_auth_token"
)

// GetAuthToken extracts the auth token from context
func GetAuthToken(ctx context.Context) string {
	if token, ok := ctx.Value(AuthTokenKey).(string); ok {
		return token
	}
	return ""
}

// WithAuthToken adds an auth token to the context
func WithAuthToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, AuthTokenKey, token)
}

// AuthMiddleware extracts the Authorization header and adds it to the request context
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		// Support both "Bearer <token>" and raw token formats
		token := authHeader
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// Add token to context if present
		if token != "" {
			ctx := WithAuthToken(r.Context(), token)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}
