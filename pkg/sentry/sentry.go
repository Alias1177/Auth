package sentry

import (
	"context"
	"log"
	"time"

	"github.com/Alias1177/Auth/internal/config"
	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
)

// Init инициализирует Sentry
func Init(cfg *config.SentryConfig, logger *zap.Logger) error {
	// Если DSN не указан, пропускаем инициализацию
	if cfg.DSN == "" {
		logger.Info("Sentry DSN не указан, мониторинг ошибок отключен")
		return nil
	}

	// Настройка опций Sentry
	options := sentry.ClientOptions{
		Dsn:              cfg.DSN,
		Environment:      cfg.Environment,
		Debug:            cfg.Debug,
		TracesSampleRate: cfg.TracesSampleRate,
		EnableTracing:    cfg.EnableTracing,
		// Настройка интеграций
		Integrations: func(integrations []sentry.Integration) []sentry.Integration {
			// Добавляем кастомные интеграции если нужно
			return integrations
		},
		// Настройка beforeSend для фильтрации событий
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			// Логируем отправку события в Sentry
			logger.Debug("Отправка события в Sentry",
				zap.String("event_id", string(event.EventID)),
				zap.String("level", string(event.Level)),
				zap.String("message", event.Message),
			)
			return event
		},
		// Настройка beforeSendTransaction для фильтрации транзакций
		BeforeSendTransaction: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			logger.Debug("Отправка транзакции в Sentry",
				zap.String("event_id", string(event.EventID)),
				zap.String("transaction", event.Transaction),
			)
			return event
		},
	}

	// Инициализация Sentry
	err := sentry.Init(options)
	if err != nil {
		log.Printf("Ошибка инициализации Sentry: %v", err)
		return err
	}

	logger.Info("Sentry успешно инициализирован",
		zap.String("environment", cfg.Environment),
		zap.Bool("debug", cfg.Debug),
		zap.Bool("tracing_enabled", cfg.EnableTracing),
	)

	return nil
}

// Flush ожидает отправки всех событий в Sentry
func Flush(timeout time.Duration) {
	sentry.Flush(timeout)
}

// CaptureException отправляет исключение в Sentry
func CaptureException(err error) *sentry.EventID {
	return sentry.CaptureException(err)
}

// CaptureMessage отправляет сообщение в Sentry
func CaptureMessage(message string) *sentry.EventID {
	return sentry.CaptureMessage(message)
}

// CaptureEvent отправляет кастомное событие в Sentry
func CaptureEvent(event *sentry.Event) *sentry.EventID {
	return sentry.CaptureEvent(event)
}

// StartSpan создает новый span для трейсинга
func StartSpan(ctx context.Context, operation string) *sentry.Span {
	return sentry.StartSpan(ctx, operation)
}

// GetHub возвращает текущий hub Sentry
func GetHub() *sentry.Hub {
	return sentry.CurrentHub()
}
