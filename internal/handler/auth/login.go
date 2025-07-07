package auth

import (
	"net/http"
	"strconv"

	"github.com/Alias1177/Auth/internal/config"
	"github.com/Alias1177/Auth/internal/domain"
	"github.com/Alias1177/Auth/internal/dto"
	"github.com/Alias1177/Auth/internal/service"
	"github.com/Alias1177/Auth/pkg/errors"
	"github.com/Alias1177/Auth/pkg/httputil"
	"github.com/Alias1177/Auth/pkg/logger"
	crypto "github.com/Alias1177/Auth/pkg/security"
)

type AuthHandler struct {
	tokenManager   service.TokenManager
	jwtConfig      config.JWTConfig
	userRepository service.UserRepository
	logger         *logger.Logger
}

func NewAuthHandler(manager service.TokenManager, cfg config.JWTConfig, repo service.UserRepository, log *logger.Logger) *AuthHandler {
	return &AuthHandler{
		tokenManager:   manager,
		jwtConfig:      cfg,
		userRepository: repo,
		logger:         log,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Декодирование JSON запроса
	if err := httputil.DecodeJSON(r, &req, h.logger); err != nil {
		httputil.JSONErrorWithID(w, http.StatusBadRequest, dto.MsgInvalidRequest)
		return
	}

	// Получение пользователя по email
	user, err := h.userRepository.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		errors.HandleDatabaseError(w, err, h.logger, "get user by email")
		return
	}

	// Проверка пароля
	if err := crypto.VerifyPassword(user.Password, req.Password); err != nil {
		httputil.JSONErrorWithID(w, http.StatusUnauthorized, dto.MsgWrongPassword)
		return
	}

	// Генерация JWT токена
	claims := domain.UserClaims{
		UserID: strconv.Itoa(user.ID),
		Email:  user.Email,
	}

	accessToken, err := h.tokenManager.GenerateAccessToken(claims)
	if err != nil {
		errors.HandleInternalError(w, err, h.logger, "generate access token")
		return
	}
	refreshToken, err := h.tokenManager.GenerateRefreshToken(claims)
	if err != nil {
		errors.HandleInternalError(w, err, h.logger, "generate refresh token")
		return
	}

	// Установка токена в куки
	httputil.SetTokenCookie(w, "access-token", accessToken)

	// Отправка успешного ответа
	response := map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}

	if err := httputil.JSONSuccessWithID(w, http.StatusOK, dto.MsgSuccessLogin, response); err != nil {
		errors.HandleInternalError(w, err, h.logger, "encode response")
	}
}
