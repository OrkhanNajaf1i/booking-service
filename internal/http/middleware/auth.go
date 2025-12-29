package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	authDomain "github.com/OrkhanNajaf1i/booking-service/internal/domain/auth"
)

type contextKey string

const (
	UserKey     contextKey = "user"
	UserIDKey   contextKey = "user_id"
	RoleKey     contextKey = "role"
	BusinessKey contextKey = "business_id"
)

func AuthMiddleware(tokenManager authDomain.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				sendError(w, http.StatusUnauthorized, "NO_TOKEN", "Authorization header tələb olunur")
				return
			}
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				sendError(w, http.StatusUnauthorized, "INVALID_TOKEN_FORMAT", "Token formatı yanlışdır")
				return
			}
			token := parts[1]
			claims, err := tokenManager.ValidateAccessToken(token)
			if err != nil {
				sendError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Token etibarsızdır")
				return
			}
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, RoleKey, string(claims.Role))
			ctx = context.WithValue(ctx, BusinessKey, claims.BusinessID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roleVal := r.Context().Value(RoleKey)
			if roleVal != nil {
				sendError(w, http.StatusForbidden, "NO_ROLE", "Rol məlumatı tapılmadı")
				return
			}
			userRole := roleVal.(string)
			allowed := false
			for _, role := range allowedRoles {
				if userRole == role {
					allowed = true
					break
				}
			}
			if !allowed {
				sendError(w, http.StatusForbidden, "FORBIDDEN", "İcazəsiz giriş")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func sendError(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"code":    code,
		"message": msg,
	})
}
