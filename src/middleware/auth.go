package middleware

import (
	"context"
	"net/http"
	"strings"

	"project-manager/src/utils"
)

type contextKey string

const UserIDKey contextKey = "userID"

// enforce valid JWT presence and expose the user ID to downstream handlers
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.WriteError(w, http.StatusUnauthorized, "Authorization header missing")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.WriteError(w, http.StatusUnauthorized, "Authorization header format must be Bearer {token}")
			return
		}

		tokenString := parts[1]

		userID, err := utils.ValidateToken(tokenString)
		if err != nil {
			utils.WriteError(w, http.StatusUnauthorized, err.Error())
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
