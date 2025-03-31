package user

import (
	"Auth/internal/usecase"
	"Auth/pkg/logger"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

// ResetPasswordRequest структура для запроса обновления пароля
type ResetPasswordRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,containsAny=1234567890!@#$%"`
}

func ValidateRequest(v interface{}) error {
	validate := validator.New()
	return validate.Struct(v)
}

// ResetPasswordHandler обработчик для сброса пароля по email
func ResetPasswordHandler(userRepo usecase.UserRepository, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Разбираем тело запроса
		var req ResetPasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Errorw("Ошибка при разборе запроса на сброс пароля", "err", err)
			http.Error(w, "Некорректный запрос", http.StatusBadRequest)
			return
		}

		// Проверяем, что email и пароль заполнены
		if req.Email == "" || req.Password == "" {
			log.Errorw("Отсутствует email или пароль в запросе", "email", req.Email)
			http.Error(w, "Email и пароль должны быть заполнены", http.StatusBadRequest)
			return
		}

		if err := ValidateRequest(req); err != nil {
			log.Errorw("error while validating request", "error", err)
			http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		}

		// Находим пользователя по email
		user, err := userRepo.GetUserByEmail(r.Context(), req.Email)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				log.Errorw("Пользователь не найден", "email", req.Email)
				http.Error(w, "Пользователь с таким email не найден", http.StatusNotFound)
				return
			}
			log.Errorw("Ошибка при поиске пользователя", "email", req.Email, "err", err)
			http.Error(w, "Ошибка при поиске пользователя", http.StatusInternalServerError)
			return
		}

		// Хэшируем новый пароль
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Errorw("Ошибка при хэшировании пароля", "err", err)
			http.Error(w, "Не удалось хэшировать пароль", http.StatusInternalServerError)
			return
		}

		// Обновляем пароль пользователя
		user.Password = string(hashedPassword)

		// Обновляем пользователя в базе данных
		if err := userRepo.UpdateUser(r.Context(), user); err != nil {
			log.Errorw("Ошибка при обновлении пароля", "email", req.Email, "err", err)
			http.Error(w, "Не удалось обновить пароль", http.StatusInternalServerError)
			return
		}

		// Отвечаем успехом
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Пароль успешно обновлён",
		})
	}
}
