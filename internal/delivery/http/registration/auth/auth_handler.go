package auth

import (
	"Auth/config"
	"Auth/internal/entity"
	"Auth/internal/usecase"
	"database/sql"
	"encoding/json"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strconv"
)

type AuthHandler struct {
	tokenManager   usecase.TokenManager
	jwtConfig      config.JWTConfig
	userRepository usecase.UserRepository
}

func NewAuthHandler(manager usecase.TokenManager, cfg config.JWTConfig, repo usecase.UserRepository) *AuthHandler {
	return &AuthHandler{
		tokenManager:   manager,
		jwtConfig:      cfg,
		userRepository: repo,
	}
}

// setTokenCookie устанавливает JWT токен в куки с необходимыми параметрами безопасности
func (h *AuthHandler) setTokenCookie(w http.ResponseWriter, cookieName, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	user, err := h.userRepository.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Пользователь не найден", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Ошибка запроса пользователя", http.StatusInternalServerError)
		return
	}

	// 🔐 Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "Пароль неверный", http.StatusUnauthorized)
		return
	}

	// Генерация JWT-токенов
	claims := entity.UserClaims{
		UserID: strconv.Itoa(user.ID),
		Email:  user.Email,
	}

	accessToken, err := h.tokenManager.GenerateAccessToken(claims)
	if err != nil {
		http.Error(w, "Не удалось создать access token", http.StatusInternalServerError)
		return
	}

	// Установка токенов в куки
	h.setTokenCookie(w, "access-token", accessToken)

	// 👇 Возвращаем токены в JSON-ответе
	response := map[string]string{
		"message":      "Вы успешно вошли в систему",
		"access_token": accessToken,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Ошибка формирования JSON ответа", http.StatusInternalServerError)
	}
}
