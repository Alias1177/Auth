package middleware

import (
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
)

// SentryMiddleware создает middleware для интеграции с Sentry
func SentryMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Создаем span для трейсинга
			span := sentry.StartSpan(r.Context(), "http.request")
			defer span.Finish()

			// Добавляем информацию о запросе в span
			span.SetTag("http.method", r.Method)
			span.SetTag("http.url", r.URL.String())
			span.SetTag("http.user_agent", r.UserAgent())
			span.SetTag("http.remote_addr", r.RemoteAddr)

			// Создаем hub для этого запроса
			hub := sentry.CurrentHub().Clone()
			hub.ConfigureScope(func(scope *sentry.Scope) {
				scope.SetTag("http.method", r.Method)
				scope.SetTag("http.url", r.URL.String())
				scope.SetTag("http.user_agent", r.UserAgent())
				scope.SetTag("http.remote_addr", r.RemoteAddr)
				scope.SetContext("http", map[string]interface{}{
					"method":      r.Method,
					"url":         r.URL.String(),
					"user_agent":  r.UserAgent(),
					"remote_addr": r.RemoteAddr,
				})
			})

			// Создаем новый контекст с hub
			ctx := sentry.SetHubOnContext(r.Context(), hub)

			// Обертываем ResponseWriter для отслеживания статуса
			wrappedWriter := &sentryResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Засекаем время начала
			start := time.Now()

			// Выполняем следующий обработчик
			next.ServeHTTP(wrappedWriter, r.WithContext(ctx))

			// Добавляем информацию о ответе
			duration := time.Since(start)
			span.SetTag("http.status_code", string(rune(wrappedWriter.statusCode)))
			span.SetTag("http.duration_ms", string(rune(duration.Milliseconds())))

			hub.ConfigureScope(func(scope *sentry.Scope) {
				scope.SetTag("http.status_code", string(rune(wrappedWriter.statusCode)))
				scope.SetTag("http.duration_ms", string(rune(duration.Milliseconds())))
			})

			// Логируем запрос
			logger.Debug("HTTP request processed",
				zap.String("method", r.Method),
				zap.String("url", r.URL.String()),
				zap.Int("status_code", wrappedWriter.statusCode),
				zap.Duration("duration", duration),
			)

			// Если произошла ошибка (4xx или 5xx), отправляем в Sentry
			if wrappedWriter.statusCode >= 400 {
				event := sentry.NewEvent()
				event.Level = sentry.LevelError
				event.Message = "HTTP request failed"
				event.Tags = map[string]string{
					"http.method":      r.Method,
					"http.url":         r.URL.String(),
					"http.status_code": string(rune(wrappedWriter.statusCode)),
				}
				event.Contexts = map[string]sentry.Context{
					"http": {
						"method":      r.Method,
						"url":         r.URL.String(),
						"status_code": wrappedWriter.statusCode,
						"duration_ms": duration.Milliseconds(),
					},
				}
				hub.CaptureEvent(event)
			}
		})
	}
}

// sentryResponseWriter обертка для ResponseWriter для отслеживания статуса
type sentryResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *sentryResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *sentryResponseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}
