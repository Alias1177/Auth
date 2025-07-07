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
		httputil.JSONErrorWithID(w, http.StatusBadRequest, dto.MsgInvalidRequest)
		return
	}

	// Валидация запроса
	if err := h.validator.Validate(req); err != nil {
		h.logger.Warnw("Validation failed for password reset request", "email", req.Email, "error", err)
		httputil.JSONErrorWithID(w, http.StatusBadRequest, dto.MsgInvalidRequestData)
		return
	}

	// Запрашиваем сброс пароля
	if err := h.passwordResetService.RequestReset(r.Context(), req.Email); err != nil {
		h.logger.Errorw("Failed to request password reset", "email", req.Email, "error", err)
		errors.HandleInternalError(w, err, h.logger, "request password reset")
		return
	}

	// Отправляем успешный ответ
	response := dto.RequestPasswordResetResponse{}

	// В режиме разработки также отправляем код для тестирования
	if code, err := h.emailService.SendPasswordResetCode(r.Context(), req.Email, ""); err == nil && code != "" {
		response.Code = code
	}

	if err := httputil.JSONSuccessWithID(w, http.StatusOK, dto.MsgSuccessPasswordResetRequested, response); err != nil {
		errors.HandleInternalError(w, err, h.logger, "encode response")
	}
}

// ConfirmPasswordReset обрабатывает подтверждение сброса пароля
func (h *PasswordResetHandler) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req dto.ConfirmPasswordResetRequest

	// Декодирование JSON запроса
	if err := httputil.DecodeJSON(r, &req, h.logger); err != nil {
		httputil.JSONErrorWithID(w, http.StatusBadRequest, dto.MsgInvalidRequest)
		return
	}

	// Валидация запроса
	if err := h.validator.Validate(req); err != nil {
		h.logger.Warnw("Validation failed for password reset confirmation", "email", req.Email, "error", err)
		httputil.JSONErrorWithID(w, http.StatusBadRequest, dto.MsgInvalidRequestData)
		return
	}

	// Подтверждаем сброс пароля
	if err := h.passwordResetService.ConfirmReset(r.Context(), req.Email, req.Code, req.Password); err != nil {
		h.logger.Errorw("Failed to confirm password reset", "email", req.Email, "error", err)

		// Обрабатываем различные типы ошибок
		switch err {
		case apperrors.ErrInvalidToken:
			httputil.JSONErrorWithID(w, http.StatusBadRequest, dto.MsgInvalidResetCode)
		case apperrors.ErrExpiredToken:
			httputil.JSONErrorWithID(w, http.StatusBadRequest, dto.MsgResetCodeExpired)
		case apperrors.ErrTooManyRequests:
			httputil.JSONErrorWithID(w, http.StatusTooManyRequests, dto.MsgTooManyResetAttempts)
		case apperrors.ErrInvalidPassword:
			httputil.JSONErrorWithID(w, http.StatusBadRequest, dto.MsgPasswordTooWeak)
		case apperrors.ErrUserNotFound:
			httputil.JSONErrorWithID(w, http.StatusNotFound, dto.MsgUserNotFound)
		default:
			errors.HandleInternalError(w, err, h.logger, "confirm password reset")
		}
		return
	}

	// Отправляем успешный ответ
	response := dto.ConfirmPasswordResetResponse{}

	if err := httputil.JSONSuccessWithID(w, http.StatusOK, dto.MsgSuccessPasswordResetConfirmed, response); err != nil {
		errors.HandleInternalError(w, err, h.logger, "encode response")
	}
}
