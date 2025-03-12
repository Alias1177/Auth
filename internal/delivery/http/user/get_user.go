package user

import (
	"Auth/internal/entity"
	"Auth/internal/infrastructure/middleware"
	"encoding/json"
	"net/http"
	"strconv"
)

// GetUserInfoHandler отвечает за получение информации о пользователе.
func (h *UserHandler) GetUserInfoHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем информацию о пользователе из контекста (добавленную middleware)
	userClaims, ok := r.Context().Value(middleware.CtxUserKey).(*entity.UserClaims)
	if !ok {
		http.Error(w, "Ошибка получения информации о пользователе", http.StatusInternalServerError)
		return
	}

	// Преобразуем ID из строки в int
	userID, err := strconv.Atoi(userClaims.UserID)
	if err != nil {
		http.Error(w, "Некорректный ID пользователя", http.StatusBadRequest)
		return
	}

	// Получаем полную информацию о пользователе
	user, err := h.userRepository.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем информацию о пользователе
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
