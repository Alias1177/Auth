package registration

import (
	"net/http"
	"strconv"

	"github.com/Alias1177/Auth/config"
	"github.com/Alias1177/Auth/internal/entity"
	"github.com/Alias1177/Auth/internal/usecase"
	"github.com/Alias1177/Auth/pkg/crypto"
	"github.com/Alias1177/Auth/pkg/errors"
	"github.com/Alias1177/Auth/pkg/httputil"
	"github.com/Alias1177/Auth/pkg/kafka"
	"github.com/Alias1177/Auth/pkg/logger"
)

type RegistrationHandler struct {
	userRepository usecase.UserRepository
	tokenManager   usecase.TokenManager
	jwtConfig      config.JWTConfig
	logger         *logger.Logger
	kafkaProducer  *kafka.Producer
}

func NewRegistrationHandler(
	repo usecase.UserRepository,
	manager usecase.TokenManager,
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
		httputil.JSONError(w, http.StatusBadRequest, "Некорректный запрос")
		return
	}

	// Проверка существования пользователя
	_, err := h.userRepository.GetUserByEmail(r.Context(), req.Email)
	if err == nil {
		h.logger.Warnw("User already exists", "email", req.Email, "username", req.Username)
		httputil.JSONError(w, http.StatusConflict, "Пользователь с таким email уже существует")
		return
	}

	// Хеширование пароля
	hashedPassword, err := crypto.HashPassword(req.Password)
	if err != nil {
		errors.HandleInternalError(w, err, h.logger, "hash password")
		return
	}

	// Создание пользователя
	newUser := entity.User{
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
	claims := entity.UserClaims{
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
	response := httputil.SuccessResponse("Пользователь успешно зарегистрирован")
	if err := httputil.JSONResponse(w, http.StatusCreated, response); err != nil {
		errors.HandleInternalError(w, err, h.logger, "encode response")
	}
}
