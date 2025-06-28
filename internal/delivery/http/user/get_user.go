package user

import (
	"net/http"
	"strconv"

	"github.com/Alias1177/Auth/internal/entity"
	"github.com/Alias1177/Auth/internal/infrastructure/middleware"
	"github.com/Alias1177/Auth/pkg/errors"
	"github.com/Alias1177/Auth/pkg/httputil"
)

// GetUserInfoHandler отвечает за получение информации о пользователе.
func (h *UserHandler) GetUserInfoHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем информацию о пользователе из контекста (добавленную middleware)
	userClaims, ok := r.Context().Value(middleware.CtxUserKey).(*entity.UserClaims)
	if !ok {
		errors.HandleInternalError(w, nil, h.logger, "get user claims from context")
		return
	}

	// Преобразуем ID из строки в int
	userID, err := strconv.Atoi(userClaims.UserID)
	if err != nil {
		h.logger.Errorw("Invalid user ID in claims", "user_id", userClaims.UserID, "error", err)
		httputil.JSONError(w, http.StatusBadRequest, "Некорректный ID пользователя")
		return
	}

	// Получаем полную информацию о пользователе
	user, err := h.userRepository.GetUserByID(r.Context(), userID)
	if err != nil {
		errors.HandleDatabaseError(w, err, h.logger, "get user by ID")
		return
	}

	// Отправляем информацию о пользователе
	if err := httputil.JSONResponse(w, http.StatusOK, user); err != nil {
		errors.HandleInternalError(w, err, h.logger, "encode user response")
	}
}
