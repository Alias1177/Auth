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
	"time"
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
func (h *AuthHandler) setTokenCookie(w http.ResponseWriter, cookieName, token string, tokenTTL time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(tokenTTL),
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

	// 🔐 Проверка пароля тут:
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "Пароль неверный", http.StatusUnauthorized)
		return
	}

	// Логика генерации токенов и установки cookie ниже:
	claims := entity.UserClaims{
		UserID: strconv.Itoa(user.ID),
		Email:  user.Email,
	}

	accessToken, err := h.tokenManager.GenerateAccessToken(claims)
	if err != nil {
		http.Error(w, "Не удалось создать access token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := h.tokenManager.GenerateRefreshToken(claims)
	if err != nil {
		http.Error(w, "Не удалось создать refresh token", http.StatusInternalServerError)
		return
	}

	h.setTokenCookie(w, "access-token", accessToken, h.jwtConfig.AccessTokenTTL)
	h.setTokenCookie(w, "refresh-token", refreshToken, h.jwtConfig.RefreshTokenTTL)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Вы успешно вошли в систему",
	})
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh-token")
	if err != nil {
		http.Error(w, "Refresh токен не найден", http.StatusUnauthorized)
		return
	}

	refreshToken := cookie.Value

	claims, err := h.tokenManager.ParseRefreshToken(refreshToken)
	if err != nil {
		http.Error(w, "Неверный refresh токен", http.StatusUnauthorized)
		return
	}

	// Генерация новых токенов
	newClaims := entity.UserClaims{
		UserID: claims.UserID,
		Email:  claims.Email,
	}

	newAccessToken, err := h.tokenManager.GenerateAccessToken(newClaims)
	if err != nil {
		http.Error(w, "Не удалось создать access token", http.StatusInternalServerError)
		return
	}

	newRefreshToken, err := h.tokenManager.GenerateRefreshToken(newClaims)
	if err != nil {
		http.Error(w, "Не удалось создать refresh token", http.StatusInternalServerError)
		return
	}

	// Установка новых токенов в Cookie
	h.setTokenCookie(w, "access-token", newAccessToken, h.jwtConfig.AccessTokenTTL)
	h.setTokenCookie(w, "refresh-token", newRefreshToken, h.jwtConfig.RefreshTokenTTL)

	response := map[string]string{
		"message": "Токены успешно обновлены",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
