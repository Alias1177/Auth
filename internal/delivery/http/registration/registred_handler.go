package registration

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Alias1177/Auth/config"
	"github.com/Alias1177/Auth/internal/entity"
	"github.com/Alias1177/Auth/internal/usecase"
	"github.com/Alias1177/Auth/pkg/kafka"
	"github.com/Alias1177/Auth/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

type RegistrationHandler struct {
	userRepository usecase.UserRepository
	tokenManager   usecase.TokenManager
	jwtConfig      config.JWTConfig
	logger         logger.Logger
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
		logger:         *log,
		kafkaProducer:  producer,
	}
}

func (h *RegistrationHandler) setTokenCookie(w http.ResponseWriter, cookieName, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})
}

func (h *RegistrationHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Errorw("error while decoding body", "error", err)
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	_, err := h.userRepository.GetUserByEmail(r.Context(), req.Email)
	if err == nil {
		h.logger.Errorw("User already exists", "email", req.Email, "username", req.Username)
		http.Error(w, "Пользователь с таким email уже существует", http.StatusConflict)
		return
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		h.logger.Errorw("error while fetching user", "error", err)
		http.Error(w, "Ошибка проверки пользователя", http.StatusInternalServerError)
		return
	}

	// 🔑 Хешируем пароль перед записью в БД
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.logger.Errorw("error while hashing password", "error", err)
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
		h.logger.Errorw("error while creating user", "error", err)
		http.Error(w, "Ошибка создания пользователя", http.StatusInternalServerError)
		return
	}

	// Отправляем информацию о регистрации в Kafka
	if h.kafkaProducer != nil {
		if err := h.kafkaProducer.SendEmailRegistration(r.Context(), req.Email, req.Username); err != nil {
			// Логируем ошибку, но не прерываем процесс регистрации
			h.logger.Errorw("Failed to send registration to Kafka", "error", err, "email", req.Email)
		}
	}

	claims := entity.UserClaims{
		UserID: strconv.Itoa(newUser.ID),
		Email:  newUser.Email,
	}

	accessToken, err := h.tokenManager.GenerateAccessToken(claims)
	if err != nil {
		h.logger.Errorw("error while generating access token", "error", err)
		http.Error(w, "Ошибка генерации access token", http.StatusInternalServerError)
		return
	}

	h.setTokenCookie(w, "access-token", accessToken)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Пользователь успешно зарегистрирован",
	})
}
