package middleware

import (
	"context"
	"net/http"
)

// ключ контекста для хранения пути
type pathContextKey string

const PathKey pathContextKey = "metric_path"

// PathMiddleware явно устанавливает путь в контекст для метрик
func PathMiddleware(path string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Добавляем информацию о пути в контекст
			ctx := context.WithValue(r.Context(), PathKey, path)
			// Вызываем следующий обработчик с обновленным контекстом
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
