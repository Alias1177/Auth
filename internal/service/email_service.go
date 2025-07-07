package service

import (
	"context"

	"github.com/Alias1177/Auth/internal/config"
	"github.com/Alias1177/Auth/pkg/logger"
)

// EmailService интерфейс для отправки email уведомлений
type EmailService interface {
	SendPasswordResetCode(ctx context.Context, email, code string) (string, error)
	SendEmailRegistration(ctx context.Context, email, username string) error
}

// DevelopmentEmailService сервис для режима разработки
type DevelopmentEmailService struct {
	config *config.Config
	logger *logger.Logger
}

func NewDevelopmentEmailService(cfg *config.Config, log *logger.Logger) *DevelopmentEmailService {
	return &DevelopmentEmailService{
		config: cfg,
		logger: log,
	}
}

func (s *DevelopmentEmailService) SendPasswordResetCode(ctx context.Context, email, code string) (string, error) {
	// В режиме разработки возвращаем код для тестирования
	s.logger.Infow("Password reset code generated (development mode)",
		"email", email,
		"code", code,
		"message", "В продакшене код будет отправлен на email")

	return code, nil
}

func (s *DevelopmentEmailService) SendEmailRegistration(ctx context.Context, email, username string) error {
	s.logger.Infow("Email registration notification (development mode)",
		"email", email,
		"username", username,
		"message", "В продакшене уведомление будет отправлено на email")
	return nil
}

// ProductionEmailService сервис для продакшена (заглушка для будущей интеграции)
type ProductionEmailService struct {
	config *config.Config
	logger *logger.Logger
}

func NewProductionEmailService(cfg *config.Config, log *logger.Logger) *ProductionEmailService {
	return &ProductionEmailService{
		config: cfg,
		logger: log,
	}
}

func (s *ProductionEmailService) SendPasswordResetCode(ctx context.Context, email, code string) (string, error) {
	// TODO: Интеграция с микросервисом нотификаций
	s.logger.Infow("Password reset code sent via notification service",
		"email", email,
		"code", code)
	// Всегда возвращаем code, независимо от окружения
	return code, nil
}

func (s *ProductionEmailService) SendEmailRegistration(ctx context.Context, email, username string) error {
	// TODO: Интеграция с микросервисом нотификаций
	s.logger.Infow("Email registration sent via notification service",
		"email", email,
		"username", username)
	return nil
}

// NewEmailService создает EmailService в зависимости от окружения
func NewEmailService(cfg *config.Config, log *logger.Logger) EmailService {
	if cfg.App.Environment == "production" {
		return NewProductionEmailService(cfg, log)
	}
	return NewDevelopmentEmailService(cfg, log)
}
