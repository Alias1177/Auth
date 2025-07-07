package user

import (
	"net/http"
	"strconv"

	"github.com/Alias1177/Auth/internal/domain"
	"github.com/Alias1177/Auth/internal/dto"
	"github.com/Alias1177/Auth/internal/service"
	"github.com/Alias1177/Auth/pkg/errors"
	"github.com/Alias1177/Auth/pkg/httputil"
	"github.com/Alias1177/Auth/pkg/logger"
	crypto "github.com/Alias1177/Auth/pkg/security"
	"github.com/go-chi/chi/v5"
)

// UserHandler управляет запросами, связанными с пользователями.
type UserHandler struct {
	userRepository service.UserRepository
	logger         *logger.Logger
}

func NewUserHandler(userRepo service.UserRepository, log *logger.Logger) *UserHandler {
	return &UserHandler{
		userRepository: userRepo,
		logger:         log,
	}
}

// UpdateUser обновляет данные пользователя.
func (h *UserHandler) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем ID пользователя из URL параметра
	userIDStr := chi.URLParam(r, "id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		h.logger.Errorw("Invalid user ID", "user_id", userIDStr, "error", err)
		httputil.JSONErrorWithID(w, http.StatusBadRequest, dto.MsgInvalidUserID)
		return
	}

	// Декодирование JSON запроса
	var user domain.User
	if err := httputil.DecodeJSON(r, &user, h.logger); err != nil {
		httputil.JSONErrorWithID(w, http.StatusBadRequest, dto.MsgInvalidRequest)
		return
	}

	// Устанавливаем ID пользователя
	user.ID = userID

	// Если был передан новый пароль, хешируем его
	if user.Password != "" {
		hashedPassword, err := crypto.HashPassword(user.Password)
		if err != nil {
			errors.HandleInternalError(w, err, h.logger, "hash password")
			return
		}
		user.Password = hashedPassword
	}

	// Обновляем пользователя
	if err := h.userRepository.UpdateUser(r.Context(), &user); err != nil {
		errors.HandleInternalError(w, err, h.logger, "update user")
		return
	}

	// Отправка успешного ответа
	if err := httputil.JSONSuccessWithID(w, http.StatusOK, dto.MsgSuccessUserUpdated, nil); err != nil {
		errors.HandleInternalError(w, err, h.logger, "encode response")
	}
}
