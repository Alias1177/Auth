package user

import (
	"Auth/internal/entity"
	"Auth/internal/usecase"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strconv"
)

// UserHandler управляет запросами, связанными с пользователями.
type UserHandler struct {
	userRepository usecase.UserRepository
}

func NewUserHandler(userRepo usecase.UserRepository) *UserHandler {
	return &UserHandler{
		userRepository: userRepo,
	}
}

// UpdateUser обновляет данные пользователя.
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// Получаем ID пользователя из URL параметра
	userIDStr := chi.URLParam(r, "id") // Получаем ID пользователя из параметров URL
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Некорректный ID пользователя", http.StatusBadRequest)
		return
	}

	// Разбираем тело запроса
	var user entity.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	// Устанавливаем ID пользователя
	user.ID = userID

	// Если был передан новый пароль, хэшируем его
	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Не удалось хэшировать пароль", http.StatusInternalServerError)
			return
		}
		// Заменяем пароль на хэшированный
		user.Password = string(hashedPassword)
	}

	// Обновляем пользователя
	err = h.userRepository.UpdateUser(r.Context(), &user)
	if err != nil {
		http.Error(w, "Не удалось обновить данные пользователя", http.StatusInternalServerError)
		return
	}

	// Ответ клиенту
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Пользователь успешно обновлён",
	})
}
