package middleware

import (
	"Auth/internal/usecase"
	"context"
	"net/http"
)

type contextKey string

const CtxUserKey contextKey = "user"

func JWTAuthMiddleware(manager usecase.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("access-token")
			if err != nil || cookie.Value == "" {
				http.Error(w, "Unauthorized - no token", http.StatusUnauthorized)
				return
			}

			userClaims, err := manager.ValidateAccessToken(cookie.Value)
			if err != nil {
				http.Error(w, "Unauthorized - invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), CtxUserKey, userClaims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
