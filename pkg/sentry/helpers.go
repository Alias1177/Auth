package sentry

import (
	"context"
	"net/http"

	"github.com/getsentry/sentry-go"
)

// CaptureError отправляет ошибку в Sentry с дополнительным контекстом
func CaptureError(ctx context.Context, err error, req *http.Request) *sentry.EventID {
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub()
	}

	// Добавляем информацию о запросе
	hub.ConfigureScope(func(scope *sentry.Scope) {
		if req != nil {
			scope.SetTag("http.method", req.Method)
			scope.SetTag("http.url", req.URL.String())
			scope.SetTag("http.user_agent", req.UserAgent())
			scope.SetTag("http.remote_addr", req.RemoteAddr)
		}
		scope.SetLevel(sentry.LevelError)
	})

	return hub.CaptureException(err)
}

// CaptureMessageWithContext отправляет сообщение в Sentry с дополнительным контекстом
func CaptureMessageWithContext(ctx context.Context, message string, req *http.Request) *sentry.EventID {
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub()
	}

	// Добавляем информацию о запросе
	hub.ConfigureScope(func(scope *sentry.Scope) {
		if req != nil {
			scope.SetTag("http.method", req.Method)
			scope.SetTag("http.url", req.URL.String())
			scope.SetTag("http.user_agent", req.UserAgent())
			scope.SetTag("http.remote_addr", req.RemoteAddr)
		}
		scope.SetLevel(sentry.LevelInfo)
	})

	return hub.CaptureMessage(message)
}

// CaptureWarning отправляет предупреждение в Sentry
func CaptureWarning(ctx context.Context, message string, req *http.Request) *sentry.EventID {
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub()
	}

	hub.ConfigureScope(func(scope *sentry.Scope) {
		if req != nil {
			scope.SetTag("http.method", req.Method)
			scope.SetTag("http.url", req.URL.String())
			scope.SetTag("http.user_agent", req.UserAgent())
			scope.SetTag("http.remote_addr", req.RemoteAddr)
		}
		scope.SetLevel(sentry.LevelWarning)
	})

	return hub.CaptureMessage(message)
}

// AddUserInfo добавляет информацию о пользователе в Sentry
func AddUserInfo(ctx context.Context, userID string, email string) {
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub()
	}

	hub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{
			ID:    userID,
			Email: email,
		})
	})
}

// AddTag добавляет тег в Sentry
func AddTag(ctx context.Context, key, value string) {
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub()
	}

	hub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag(key, value)
	})
}

// AddContext добавляет контекст в Sentry
func AddContext(ctx context.Context, name string, data map[string]interface{}) {
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub()
	}

	hub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetContext(name, data)
	})
}

// GetHubFromContext получает hub из контекста
func GetHubFromContext(ctx context.Context) *sentry.Hub {
	return sentry.GetHubFromContext(ctx)
}
