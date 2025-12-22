package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/myestatia/myestatia-go/internal/infrastructure/security"
)

type contextKey string

const (
	AgentIDKey   contextKey = "agent_id"
	CompanyIDKey contextKey = "company_id"
	RoleKey      contextKey = "role"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		claims, err := security.ValidateToken(bearerToken[1])
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), AgentIDKey, claims.AgentID)
		ctx = context.WithValue(ctx, CompanyIDKey, claims.CompanyID)
		ctx = context.WithValue(ctx, RoleKey, claims.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
