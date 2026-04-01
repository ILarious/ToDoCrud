package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"todo_crud/internal/app/http/handler"
	"todo_crud/internal/domain/service"
)

func AuthMiddleware(auth service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
			if authHeader == "" {
				writeUnauthorized(w, "authorization header is required")
				return
			}

			token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
			if token == authHeader || token == "" {
				writeUnauthorized(w, "invalid authorization header")
				return
			}

			userID, err := auth.ParseToken(token)
			if err != nil {
				writeUnauthorized(w, "invalid token")
				return
			}

			next.ServeHTTP(w, r.WithContext(handler.WithUserID(r.Context(), userID)))
		})
	}
}

func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
