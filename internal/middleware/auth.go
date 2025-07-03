package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Alias1177/Auth/internal/service"
)

type contextKey string

const CtxUserKey contextKey = "user"

func JWTAuthMiddleware(manager service.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var token string

			// 1️⃣ Сначала проверяем Authorization-заголовок
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			}

			// 2️⃣ Если в заголовке нет, проверяем cookie
			if token == "" {
				cookie, err := r.Cookie("access-token")
				if err == nil {
					token = cookie.Value
				}
			}

			// 3️⃣ Если токен так и не нашли — отправляем ошибку
			if token == "" {
				http.Error(w, "Unauthorized - no token", http.StatusUnauthorized)
				return
			}

			// 4️⃣ Проверяем токен
			userClaims, err := manager.ValidateAccessToken(token)
			if err != nil {
				fmt.Println("Ошибка валидации токена:", err)
				http.Error(w, "Unauthorized - invalid token", http.StatusUnauthorized)
				return
			}

			// 5️⃣ Добавляем данные пользователя в контекст
			ctx := context.WithValue(r.Context(), CtxUserKey, userClaims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
