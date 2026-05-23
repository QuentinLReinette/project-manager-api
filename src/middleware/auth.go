package middleware

import (
	"context"
	"net/http"

	"project-manager/src/utils"
)

type contextKey string

const UserIDKey contextKey = "userID"

// enforce valid JWT presence and expose the user ID to downstream handlers
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			utils.WriteError(w, http.StatusUnauthorized, "Authentication cookie missing or expired")
			return
		}

		tokenString := cookie.Value

		userID, err := utils.ValidateToken(tokenString)
		if err != nil {
			utils.WriteError(w, http.StatusUnauthorized, err.Error())
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
