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
	"github.com/Alias1177/Auth/pkg/kafka"
	"github.com/Alias1177/Auth/pkg/logger"
	crypto "github.com/Alias1177/Auth/pkg/security"
)

type RegistrationHandler struct {
	userRepository service.UserRepository
	tokenManager   service.TokenManager
	jwtConfig      config.JWTConfig
	logger         *logger.Logger
	kafkaProducer  *kafka.Producer
}

func NewRegistrationHandler(
	repo service.UserRepository,
	manager service.TokenManager,
	cfg config.JWTConfig,
	log *logger.Logger,
	producer *kafka.Producer,
) *RegistrationHandler {
	return &RegistrationHandler{
		userRepository: repo,
		tokenManager:   manager,
		jwtConfig:      cfg,
		logger:         log,
		kafkaProducer:  producer,
	}
}

func (h *RegistrationHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// Декодирование JSON запроса
	if err := httputil.DecodeJSON(r, &req, h.logger); err != nil {
		httputil.JSONErrorWithID(w, http.StatusBadRequest, dto.MsgInvalidRequest)
		return
	}

	// Проверка существования пользователя
	_, err := h.userRepository.GetUserByEmail(r.Context(), req.Email)
	if err == nil {
		h.logger.Warnw("User already exists", "email", req.Email, "username", req.Username)
		httputil.JSONErrorWithID(w, http.StatusConflict, dto.MsgEmailAlreadyExists)
		return
	}

	// Хеширование пароля
	hashedPassword, err := crypto.HashPassword(req.Password)
	if err != nil {
		errors.HandleInternalError(w, err, h.logger, "hash password")
		return
	}

	// Создание пользователя
	newUser := domain.User{
		Email:    req.Email,
		UserName: req.Username,
		Password: hashedPassword,
	}

	if err := h.userRepository.CreateUser(r.Context(), &newUser); err != nil {
		errors.HandleInternalError(w, err, h.logger, "create user")
		return
	}

	// Отправка в Kafka (не критично для успеха регистрации)
	if h.kafkaProducer != nil {
		if err := h.kafkaProducer.SendEmailRegistration(r.Context(), req.Email, req.Username); err != nil {
			h.logger.Errorw("Failed to send registration to Kafka", "error", err, "email", req.Email)
		}
	}

	// Генерация JWT токена
	claims := domain.UserClaims{
		UserID: strconv.Itoa(newUser.ID),
		Email:  newUser.Email,
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
		"access_token": accessToken,
	}

	if err := httputil.JSONSuccessWithID(w, http.StatusCreated, dto.MsgSuccessRegister, response); err != nil {
		errors.HandleInternalError(w, err, h.logger, "encode response")
	}
}
