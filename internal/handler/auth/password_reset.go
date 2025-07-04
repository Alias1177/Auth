package auth

import (
	"net/http"

	"github.com/Alias1177/Auth/internal/dto"
	apperrors "github.com/Alias1177/Auth/internal/errors"
	"github.com/Alias1177/Auth/internal/service"
	"github.com/Alias1177/Auth/pkg/errors"
	"github.com/Alias1177/Auth/pkg/httputil"
	"github.com/Alias1177/Auth/pkg/logger"
	"github.com/Alias1177/Auth/pkg/validator"
)

// PasswordResetHandler обработчик для сброса пароля
type PasswordResetHandler struct {
	passwordResetService service.PasswordResetService
	emailService         service.EmailService
	validator            *validator.Validator
	logger               *logger.Logger
}

// NewPasswordResetHandler создает новый обработчик сброса пароля
func NewPasswordResetHandler(
	passwordResetService service.PasswordResetService,
	emailService service.EmailService,
	validator *validator.Validator,
	logger *logger.Logger,
) *PasswordResetHandler {
	return &PasswordResetHandler{
		passwordResetService: passwordResetService,
		emailService:         emailService,
		validator:            validator,
		logger:               logger,
	}
}

// RequestPasswordReset обрабатывает запрос на сброс пароля
func (h *PasswordResetHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req dto.RequestPasswordResetRequest

	// Декодирование JSON запроса
	if err := httputil.DecodeJSON(r, &req, h.logger); err != nil {
		httputil.JSONError(w, http.StatusBadRequest, "Некорректный запрос")
		return
	}

	// Валидация запроса
	if err := h.validator.Validate(req); err != nil {
		h.logger.Warnw("Validation failed for password reset request", "email", req.Email, "error", err)
		httputil.JSONError(w, http.StatusBadRequest, "Некорректные данные запроса")
		return
	}

	// Запрашиваем сброс пароля
	if err := h.passwordResetService.RequestReset(r.Context(), req.Email); err != nil {
		h.logger.Errorw("Failed to request password reset", "email", req.Email, "error", err)
		errors.HandleInternalError(w, err, h.logger, "request password reset")
		return
	}

	// Отправляем успешный ответ
	response := dto.RequestPasswordResetResponse{
		Message: "Если указанный email существует в системе, на него будет отправлен код подтверждения",
	}

	// В режиме разработки также отправляем код для тестирования
	if code, err := h.emailService.SendPasswordResetCode(r.Context(), req.Email, ""); err == nil && code != "" {
		response.Code = code
	}

	if err := httputil.JSONResponse(w, http.StatusOK, response); err != nil {
		errors.HandleInternalError(w, err, h.logger, "encode response")
	}
}

// ConfirmPasswordReset обрабатывает подтверждение сброса пароля
func (h *PasswordResetHandler) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req dto.ConfirmPasswordResetRequest

	// Декодирование JSON запроса
	if err := httputil.DecodeJSON(r, &req, h.logger); err != nil {
		httputil.JSONError(w, http.StatusBadRequest, "Некорректный запрос")
		return
	}

	// Валидация запроса
	if err := h.validator.Validate(req); err != nil {
		h.logger.Warnw("Validation failed for password reset confirmation", "email", req.Email, "error", err)
		httputil.JSONError(w, http.StatusBadRequest, "Некорректные данные запроса")
		return
	}

	// Подтверждаем сброс пароля
	if err := h.passwordResetService.ConfirmReset(r.Context(), req.Email, req.Code, req.Password); err != nil {
		h.logger.Errorw("Failed to confirm password reset", "email", req.Email, "error", err)

		// Обрабатываем различные типы ошибок
		switch err {
		case apperrors.ErrInvalidToken:
			httputil.JSONError(w, http.StatusBadRequest, "Неверный код подтверждения")
		case apperrors.ErrExpiredToken:
			httputil.JSONError(w, http.StatusBadRequest, "Код подтверждения истек")
		case apperrors.ErrTooManyRequests:
			httputil.JSONError(w, http.StatusTooManyRequests, "Превышен лимит попыток ввода кода")
		case apperrors.ErrInvalidPassword:
			httputil.JSONError(w, http.StatusBadRequest, "Пароль не соответствует требованиям безопасности")
		case apperrors.ErrUserNotFound:
			httputil.JSONError(w, http.StatusNotFound, "Пользователь не найден")
		default:
			errors.HandleInternalError(w, err, h.logger, "confirm password reset")
		}
		return
	}

	// Отправляем успешный ответ
	response := dto.ConfirmPasswordResetResponse{
		Message: "Пароль успешно изменен",
	}

	if err := httputil.JSONResponse(w, http.StatusOK, response); err != nil {
		errors.HandleInternalError(w, err, h.logger, "encode response")
	}
}
