package registration

import (
	"Auth/config"
	"Auth/internal/entity"
	"Auth/internal/usecase"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type RegistrationHandler struct {
	userRepository usecase.UserRepository
	tokenManager   usecase.TokenManager
	jwtConfig      config.JWTConfig
}

func NewRegistrationHandler(repo usecase.UserRepository, manager usecase.TokenManager, cfg config.JWTConfig) *RegistrationHandler {
	return &RegistrationHandler{
		userRepository: repo,
		tokenManager:   manager,
		jwtConfig:      cfg,
	}
}

func (h *RegistrationHandler) setTokenCookie(w http.ResponseWriter, cookieName, token string, tokenTTL time.Duration) {
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

func (h *RegistrationHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	_, err := h.userRepository.GetUserByEmail(r.Context(), req.Email)
	if err == nil {
		http.Error(w, "Пользователь с таким email уже существует", http.StatusConflict)
		return
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Ошибка проверки пользователя", http.StatusInternalServerError)
		return
	}

	// 🔑 Хешируем пароль перед записью в БД
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Ошибка при обработке пароля", http.StatusInternalServerError)
		return
	}

	// 🎯 Создание пользователя с хешированным паролем
	newUser := entity.User{
		Email:    req.Email,
		UserName: req.Username,
		Password: string(hashedPassword), // ✅ сохраняем именно хеш!
	}

	if err := h.userRepository.CreateUser(r.Context(), &newUser); err != nil {
		http.Error(w, "Ошибка создания пользователя", http.StatusInternalServerError)
		return
	}

	claims := entity.UserClaims{
		UserID: strconv.Itoa(newUser.ID),
		Email:  newUser.Email,
	}

	accessToken, err := h.tokenManager.GenerateAccessToken(claims)
	if err != nil {
		http.Error(w, "Ошибка генерации access token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := h.tokenManager.GenerateRefreshToken(claims)
	if err != nil {
		http.Error(w, "Ошибка генерации refresh token", http.StatusInternalServerError)
		return
	}

	h.setTokenCookie(w, "access-token", accessToken, h.jwtConfig.AccessTokenTTL)
	h.setTokenCookie(w, "refresh-token", refreshToken, h.jwtConfig.RefreshTokenTTL)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Пользователь успешно зарегистрирован",
	})
}
