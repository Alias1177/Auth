package auth

import (
	"net/http"
	"strconv"

	"github.com/Alias1177/Auth/config"
	"github.com/Alias1177/Auth/internal/entity"
	"github.com/Alias1177/Auth/internal/usecase"
	"github.com/Alias1177/Auth/pkg/crypto"
	"github.com/Alias1177/Auth/pkg/errors"
	"github.com/Alias1177/Auth/pkg/httputil"
	"github.com/Alias1177/Auth/pkg/logger"
)

type AuthHandler struct {
	tokenManager   usecase.TokenManager
	jwtConfig      config.JWTConfig
	userRepository usecase.UserRepository
	logger         *logger.Logger
}

func NewAuthHandler(manager usecase.TokenManager, cfg config.JWTConfig, repo usecase.UserRepository, log *logger.Logger) *AuthHandler {
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
		httputil.JSONError(w, http.StatusBadRequest, "Некорректный запрос")
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
		errors.HandleUnauthorizedError(w, "Неверный пароль", h.logger)
		return
	}

	// Генерация JWT токена
	claims := entity.UserClaims{
		UserID: strconv.Itoa(user.ID),
		Email:  user.Email,
	}

	accessToken, err := h.tokenManager.GenerateAccessToken(claims)
	if err != nil {
		errors.HandleInternalError(w, err, h.logger, "generate access token")
		return
	}

	// Установка токена в куки
	httputil.SetTokenCookie(w, "access-token", accessToken)

	// Отправка успешного ответа
	response := map[string]string{
		"message":      "Вы успешно вошли в систему",
		"access_token": accessToken,
	}

	if err := httputil.JSONResponse(w, http.StatusOK, response); err != nil {
		errors.HandleInternalError(w, err, h.logger, "encode response")
	}
}
