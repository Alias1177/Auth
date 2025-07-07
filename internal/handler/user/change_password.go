package user

import (
	"net/http"

	"github.com/Alias1177/Auth/internal/dto"
	"github.com/Alias1177/Auth/pkg/errors"
	"github.com/Alias1177/Auth/pkg/httputil"
	crypto "github.com/Alias1177/Auth/pkg/security"
)

// ResetPasswordHandler обработчик для сброса пароля по email
func (h *UserHandler) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {

	// Декодирование JSON запроса
	var req dto.ResetPasswordByEmailRequest
	if err := httputil.DecodeJSON(r, &req, h.logger); err != nil {
		httputil.JSONErrorWithID(w, http.StatusBadRequest, dto.MsgInvalidRequest)
		return
	}

	// Проверяем, что email и пароль заполнены
	if req.Email == "" || req.Password == "" {
		h.logger.Warnw("Missing email or password in reset request", "email", req.Email)
		httputil.JSONErrorWithID(w, http.StatusBadRequest, dto.MsgMissingEmailOrPassword)
		return
	}

	// Находим пользователя по email
	user, err := h.userRepository.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		errors.HandleDatabaseError(w, err, h.logger, "get user by email for password reset")
		return
	}

	// Хешируем новый пароль
	hashedPassword, err := crypto.HashPassword(req.Password)
	if err != nil {
		errors.HandleInternalError(w, err, h.logger, "hash new password")
		return
	}

	// Обновляем пароль пользователя
	user.Password = hashedPassword

	// Обновляем пользователя в базе данных
	if err := h.userRepository.UpdateUser(r.Context(), user); err != nil {
		errors.HandleInternalError(w, err, h.logger, "update user password")
		return
	}

	// Отправка успешного ответа
	if err := httputil.JSONSuccessWithID(w, http.StatusOK, dto.MsgSuccessPasswordChanged, nil); err != nil {
		errors.HandleInternalError(w, err, h.logger, "encode response")
	}

}
