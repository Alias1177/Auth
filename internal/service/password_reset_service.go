package service

import (
	"context"

	"github.com/Alias1177/Auth/internal/errors"
	"github.com/Alias1177/Auth/pkg/kafka"
	"github.com/Alias1177/Auth/pkg/logger"
	"github.com/Alias1177/Auth/pkg/notification"
	crypto "github.com/Alias1177/Auth/pkg/security"
)

// PasswordResetService интерфейс для работы со сбросом пароля
type PasswordResetService interface {
	RequestReset(ctx context.Context, email string) error
	ConfirmReset(ctx context.Context, email, code, newPassword string) error
	ValidateCode(ctx context.Context, email, code string) (bool, error)
}

// PasswordResetServiceImpl реализация сервиса сброса пароля
type PasswordResetServiceImpl struct {
	userRepo            UserRepository
	userCache           UserCache
	logger              *logger.Logger
	kafkaProducer       *kafka.Producer
	notificationClient  *notification.NotificationClient
}

// NewPasswordResetService создает новый экземпляр сервиса сброса пароля
func NewPasswordResetService(
	userRepo UserRepository,
	userCache UserCache,
	logger *logger.Logger,
	kafkaProducer *kafka.Producer,
	notificationClient *notification.NotificationClient,
) *PasswordResetServiceImpl {
	return &PasswordResetServiceImpl{
		userRepo:           userRepo,
		userCache:          userCache,
		logger:             logger,
		kafkaProducer:      kafkaProducer,
		notificationClient: notificationClient,
	}
}

// RequestReset запрашивает сброс пароля для указанного email
func (s *PasswordResetServiceImpl) RequestReset(ctx context.Context, email string) error {
	// Проверяем существование пользователя
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		s.logger.Debugw("Password reset requested for non-existent email", "email", email)
		// Возвращаем специальную ошибку для несуществующего пользователя
		return errors.ErrUserNotFound
	}

	// Отправляем запрос в Notification Service через Kafka
	err = s.kafkaProducer.SendPasswordResetRequest(ctx, email)
	if err != nil {
		s.logger.Errorw("Failed to send password reset request to Kafka", "email", email, "error", err)
		return errors.ErrInternal
	}

	s.logger.Infow("Password reset request sent to Notification Service", "email", email, "user_id", user.ID)
	return nil
}

// ConfirmReset подтверждает сброс пароля с кодом
func (s *PasswordResetServiceImpl) ConfirmReset(ctx context.Context, email, code, newPassword string) error {
	// Валидируем код через Notification Service
	isValid, err := s.ValidateCode(ctx, email, code)
	if err != nil {
		return err
	}

	if !isValid {
		return errors.ErrInvalidToken
	}

	// Получаем пользователя
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		s.logger.Errorw("Failed to get user for password reset", "email", email, "error", err)
		return errors.ErrUserNotFound
	}

	// Хешируем новый пароль
	hashedPassword, err := crypto.HashPassword(newPassword)
	if err != nil {
		s.logger.Errorw("Failed to hash new password", "email", email, "error", err)
		return errors.ErrInternal
	}

	// Обновляем пароль
	user.Password = hashedPassword
	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		s.logger.Errorw("Failed to update user password", "email", email, "error", err)
		return errors.ErrInternal
	}

	s.logger.Infow("Password successfully reset", "email", email, "user_id", user.ID)
	return nil
}

// ValidateCode проверяет код подтверждения через Notification Service
func (s *PasswordResetServiceImpl) ValidateCode(ctx context.Context, email, code string) (bool, error) {
	// Валидируем код через HTTP API Notification Service
	isValid, err := s.notificationClient.ValidatePasswordResetCode(email, code)
	if err != nil {
		s.logger.Errorw("Failed to validate code with Notification Service", "email", email, "error", err)
		return false, errors.ErrInternal
	}

	if !isValid {
		s.logger.Warnw("Invalid reset code", "email", email)
		return false, nil
	}

	return true, nil
}
