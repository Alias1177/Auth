package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/Alias1177/Auth/internal/errors"
	"github.com/Alias1177/Auth/pkg/logger"
	crypto "github.com/Alias1177/Auth/pkg/security"
)

// PasswordResetData структура для хранения данных сброса пароля в Redis
type PasswordResetData struct {
	Code      string    `json:"code"`
	Attempts  int       `json:"attempts"`
	ExpiresAt time.Time `json:"expires_at"`
}

// PasswordResetService интерфейс для работы со сбросом пароля
type PasswordResetService interface {
	RequestReset(ctx context.Context, email string) error
	ConfirmReset(ctx context.Context, email, code, newPassword string) error
	ValidateCode(ctx context.Context, email, code string) error
}

// PasswordResetServiceImpl реализация сервиса сброса пароля
type PasswordResetServiceImpl struct {
	userRepo     UserRepository
	userCache    UserCache
	logger       *logger.Logger
	emailService EmailService
}

// NewPasswordResetService создает новый экземпляр сервиса сброса пароля
func NewPasswordResetService(
	userRepo UserRepository,
	userCache UserCache,
	logger *logger.Logger,
	emailService EmailService,
) *PasswordResetServiceImpl {
	return &PasswordResetServiceImpl{
		userRepo:     userRepo,
		userCache:    userCache,
		logger:       logger,
		emailService: emailService,
	}
}

// RequestReset запрашивает сброс пароля для указанного email
func (s *PasswordResetServiceImpl) RequestReset(ctx context.Context, email string) error {
	// Проверяем существование пользователя
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		s.logger.Warnw("Password reset requested for non-existent email", "email", email)
		// Не раскрываем информацию о существовании пользователя
		return nil
	}

	// Генерируем код подтверждения
	code, err := s.generateResetCode()
	if err != nil {
		s.logger.Errorw("Failed to generate reset code", "email", email, "error", err)
		return errors.ErrInternal
	}

	// Создаем данные для сброса пароля
	resetData := PasswordResetData{
		Code:      code,
		Attempts:  0,
		ExpiresAt: time.Now().Add(15 * time.Minute), // 15 минут на ввод кода
	}

	// Сохраняем в Redis
	key := fmt.Sprintf("password_reset:%s", email)
	data, err := json.Marshal(resetData)
	if err != nil {
		s.logger.Errorw("Failed to marshal reset data", "email", email, "error", err)
		return errors.ErrInternal
	}

	// Сохраняем с TTL 15 минут
	if err := s.userCache.SetWithTTL(ctx, key, string(data), 15*time.Minute); err != nil {
		s.logger.Errorw("Failed to save reset code to cache", "email", email, "error", err)
		return errors.ErrInternal
	}

	// Отправляем код на email
	sentCode, err := s.emailService.SendPasswordResetCode(ctx, email, code)
	if err != nil {
		s.logger.Errorw("Failed to send reset code email", "email", email, "error", err)
		// Не возвращаем ошибку, чтобы не раскрывать информацию о существовании пользователя
		return nil
	}

	// Сохраняем отправленный код для возврата в ответе (только в режиме разработки)
	if sentCode != "" {
		// В режиме разработки код будет возвращен в ответе
		s.logger.Infow("Reset code will be returned in response (development mode)", "email", email)
	}

	s.logger.Infow("Password reset code sent", "email", email, "user_id", user.ID)
	return nil
}

// ConfirmReset подтверждает сброс пароля с кодом
func (s *PasswordResetServiceImpl) ConfirmReset(ctx context.Context, email, code, newPassword string) error {
	// Валидируем новый пароль
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}

	// Проверяем код
	if err := s.ValidateCode(ctx, email, code); err != nil {
		return err
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

	// Удаляем код из Redis
	key := fmt.Sprintf("password_reset:%s", email)
	if err := s.userCache.Delete(ctx, key); err != nil {
		s.logger.Warnw("Failed to delete reset code from cache", "email", email, "error", err)
		// Не критично, код и так истечет
	}

	s.logger.Infow("Password successfully reset", "email", email, "user_id", user.ID)
	return nil
}

// ValidateCode проверяет код подтверждения
func (s *PasswordResetServiceImpl) ValidateCode(ctx context.Context, email, code string) error {
	key := fmt.Sprintf("password_reset:%s", email)

	// Получаем данные из Redis
	dataStr, err := s.userCache.Get(ctx, key)
	if err != nil {
		s.logger.Warnw("Reset code not found or expired", "email", email)
		return errors.ErrExpiredToken
	}

	var resetData PasswordResetData
	if err := json.Unmarshal([]byte(dataStr), &resetData); err != nil {
		s.logger.Errorw("Failed to unmarshal reset data", "email", email, "error", err)
		return errors.ErrInternal
	}

	// Проверяем срок действия
	if time.Now().After(resetData.ExpiresAt) {
		s.logger.Warnw("Reset code expired", "email", email)
		return errors.ErrExpiredToken
	}

	// Проверяем количество попыток
	if resetData.Attempts >= 5 {
		s.logger.Warnw("Too many reset code attempts", "email", email, "attempts", resetData.Attempts)
		return errors.ErrTooManyRequests
	}

	// Увеличиваем счетчик попыток
	resetData.Attempts++

	// Проверяем код
	if resetData.Code != code {
		// Сохраняем обновленные данные
		data, _ := json.Marshal(resetData)
		s.userCache.SetWithTTL(ctx, key, string(data), 15*time.Minute)

		s.logger.Warnw("Invalid reset code", "email", email, "attempts", resetData.Attempts)
		return errors.ErrInvalidToken
	}

	return nil
}

// generateResetCode генерирует 6-значный код подтверждения
func (s *PasswordResetServiceImpl) generateResetCode() (string, error) {
	code := ""
	for i := 0; i < 6; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code += fmt.Sprintf("%d", num.Int64())
	}
	return code, nil
}

// validatePassword проверяет сложность пароля
func (s *PasswordResetServiceImpl) validatePassword(password string) error {
	if len(password) < 8 {
		return errors.ErrInvalidPassword
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasNumber = true
		case char == '!' || char == '@' || char == '#' || char == '$' || char == '%' ||
			char == '^' || char == '&' || char == '*' || char == '(' || char == ')' ||
			char == '-' || char == '_' || char == '+' || char == '=' || char == '[' ||
			char == ']' || char == '{' || char == '}' || char == '|' || char == '\\' ||
			char == ':' || char == ';' || char == '"' || char == '\'' || char == '<' ||
			char == '>' || char == ',' || char == '.' || char == '?' || char == '/':
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return errors.ErrInvalidPassword
	}

	return nil
}
