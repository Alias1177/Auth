package user

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Alias1177/Auth/internal/entity"
	"github.com/Alias1177/Auth/internal/usecase"
	"github.com/Alias1177/Auth/pkg/logger"
	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

// UserHandler управляет запросами, связанными с пользователями.
type UserHandler struct {
	userRepository usecase.UserRepository
	logger         *logger.Logger // Изменили на *logger.Logger для консистентности
}

func NewUserHandler(userRepo usecase.UserRepository, log *logger.Logger) *UserHandler {
	return &UserHandler{
		userRepository: userRepo,
		logger:         log,
	}
}

// UpdateUser обновляет данные пользователя.
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// Получаем ID пользователя из URL параметра
	userIDStr := chi.URLParam(r, "id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		if h.logger != nil {
			h.logger.Errorw("userHandler", "user_id", userIDStr, "err", err)
		}
		http.Error(w, "Некорректный ID пользователя", http.StatusBadRequest)
		return
	}

	// Разбираем тело запроса
	var user entity.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		if h.logger != nil {
			h.logger.Errorw("userHandler", "err", err)
		}
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	// Устанавливаем ID пользователя
	user.ID = userID

	// Если был передан новый пароль, хэшируем его
	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			if h.logger != nil {
				h.logger.Errorw("userHandler", "err", err)
			}
			http.Error(w, "Не удалось хэшировать пароль", http.StatusInternalServerError)
			return
		}
		user.Password = string(hashedPassword)
	}

	// Обновляем пользователя
	err = h.userRepository.UpdateUser(r.Context(), &user)
	if err != nil {
		if h.logger != nil {
			h.logger.Errorw("userHandler", "err", err)
		}
		http.Error(w, "Не удалось обновить данные пользователя", http.StatusInternalServerError)
		return
	}

	// Ответ клиенту
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Пользователь успешно обновлён",
	})
}
