package user

import (
	"net/http"

	"github.com/Alias1177/Auth/internal/service"
	"github.com/Alias1177/Auth/pkg/errors"
	"github.com/Alias1177/Auth/pkg/httputil"
	"github.com/Alias1177/Auth/pkg/logger"
	crypto "github.com/Alias1177/Auth/pkg/security"
)

// ResetPasswordRequest структура для запроса обновления пароля
type ResetPasswordRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ResetPasswordHandler обработчик для сброса пароля по email
func ResetPasswordHandler(userRepo service.UserRepository, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Декодирование JSON запроса
		var req ResetPasswordRequest
		if err := httputil.DecodeJSON(r, &req, log); err != nil {
			httputil.JSONError(w, http.StatusBadRequest, "Некорректный запрос")
			return
		}

		// Проверяем, что email и пароль заполнены
		if req.Email == "" || req.Password == "" {
			log.Warnw("Missing email or password in reset request", "email", req.Email)
			httputil.JSONError(w, http.StatusBadRequest, "Email и пароль должны быть заполнены")
			return
		}

		// Находим пользователя по email
		user, err := userRepo.GetUserByEmail(r.Context(), req.Email)
		if err != nil {
			errors.HandleDatabaseError(w, err, log, "get user by email for password reset")
			return
		}

		// Хешируем новый пароль
		hashedPassword, err := crypto.HashPassword(req.Password)
		if err != nil {
			errors.HandleInternalError(w, err, log, "hash new password")
			return
		}

		// Обновляем пароль пользователя
		user.Password = hashedPassword

		// Обновляем пользователя в базе данных
		if err := userRepo.UpdateUser(r.Context(), user); err != nil {
			errors.HandleInternalError(w, err, log, "update user password")
			return
		}

		// Отправка успешного ответа
		response := httputil.SuccessResponse("Пароль успешно обновлён")
		if err := httputil.JSONResponse(w, http.StatusOK, response); err != nil {
			errors.HandleInternalError(w, err, log, "encode response")
		}
	}
}
