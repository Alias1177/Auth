package auth

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/Alias1177/Auth/config"
	"github.com/Alias1177/Auth/internal/entity"
	"github.com/Alias1177/Auth/pkg/logger"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// MockTokenManager реализует usecase.TokenManager
type MockTokenManager struct {
	mockGenerateAccessToken func(entity.UserClaims) (string, error)
}

func (m *MockTokenManager) GenerateAccessToken(claims entity.UserClaims) (string, error) {
	return m.mockGenerateAccessToken(claims)
}

func (m *MockTokenManager) ValidateAccessToken(token string) (*entity.UserClaims, error) {
	return nil, nil
}

// MockUserRepository реализует usecase.UserRepository
type MockUserRepository struct {
	mockGetUserByEmail func(context.Context, string) (*entity.User, error)
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *entity.User) error {
	return nil
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id int) (*entity.User, error) {
	return nil, nil
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	return m.mockGetUserByEmail(ctx, email)
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, user *entity.User) error {
	return nil
}

// TestLoggerWrapper реализует *logger.Logger
type TestLoggerWrapper struct {
	logger *logger.Logger
}

func NewTestLoggerWrapper() *logger.Logger {
	return &logger.Logger{} // Возвращаем реальный тип *logger.Logger
}

func (l *TestLoggerWrapper) Infow(msg string, keysAndValues ...any)  {}
func (l *TestLoggerWrapper) Errorw(msg string, keysAndValues ...any) {}
func (l *TestLoggerWrapper) Warnw(msg string, keysAndValues ...any)  {}
func (l *TestLoggerWrapper) Debugw(msg string, keysAndValues ...any) {}
func (l *TestLoggerWrapper) Fatalw(msg string, keysAndValues ...any) {}
func (l *TestLoggerWrapper) Close() error                            { return nil }

func TestAuthHandler_Login_Success(t *testing.T) {
	// Подготовка тестовых данных
	email := "test@example.com"
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	userID := 1

	// Настройка моков
	tokenMgr := &MockTokenManager{
		mockGenerateAccessToken: func(claims entity.UserClaims) (string, error) {
			assert.Equal(t, strconv.Itoa(userID), claims.UserID)
			assert.Equal(t, email, claims.Email)
			return "test-token", nil
		},
	}

	userRepo := &MockUserRepository{
		mockGetUserByEmail: func(ctx context.Context, e string) (*entity.User, error) {
			assert.Equal(t, email, e)
			return &entity.User{
				ID:       userID,
				Email:    email,
				Password: string(hashedPassword),
			}, nil
		},
	}

	// Создание обработчика с правильным типом логгера
	handler := NewAuthHandler(
		tokenMgr,
		config.JWTConfig{Secret: "test-secret"},
		userRepo,
		NewTestLoggerWrapper(), // Используем обертку
	)

	// Подготовка запроса
	reqBody := map[string]string{
		"email":    email,
		"password": password,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Вызов метода
	handler.Login(w, req)

	// Проверки
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Вы успешно вошли в систему", response["message"])
	assert.Equal(t, "test-token", response["access_token"])

	// Проверка cookies
	cookies := w.Result().Cookies()
	assert.Len(t, cookies, 1)
	assert.Equal(t, "access-token", cookies[0].Name)
	assert.Equal(t, "test-token", cookies[0].Value)
}

func TestAuthHandler_Login_InvalidRequest(t *testing.T) {
	// Создание обработчика с правильным типом логгера
	handler := NewAuthHandler(
		&MockTokenManager{},
		config.JWTConfig{},
		&MockUserRepository{},
		NewTestLoggerWrapper(),
	)

	// Неправильный JSON
	req := httptest.NewRequest("POST", "/login", bytes.NewReader([]byte("{invalid}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Некорректный запрос")
}

func TestAuthHandler_Login_UserNotFound(t *testing.T) {
	email := "notfound@example.com"

	// Настройка моков
	userRepo := &MockUserRepository{
		mockGetUserByEmail: func(ctx context.Context, e string) (*entity.User, error) {
			assert.Equal(t, email, e)
			return nil, sql.ErrNoRows
		},
	}

	// Создание обработчика
	handler := NewAuthHandler(
		&MockTokenManager{},
		config.JWTConfig{},
		userRepo,
		NewTestLoggerWrapper(),
	)

	// Подготовка запроса
	reqBody := map[string]string{
		"email":    email,
		"password": "password",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Пользователь не найден")
}

func TestAuthHandler_Login_WrongPassword(t *testing.T) {
	email := "user@example.com"
	password := "wrongpassword"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)

	// Настройка моков
	userRepo := &MockUserRepository{
		mockGetUserByEmail: func(ctx context.Context, e string) (*entity.User, error) {
			return &entity.User{
				Email:    email,
				Password: string(hashedPassword),
			}, nil
		},
	}

	// Создание обработчика
	handler := NewAuthHandler(
		&MockTokenManager{},
		config.JWTConfig{},
		userRepo,
		NewTestLoggerWrapper(),
	)

	// Подготовка запроса
	reqBody := map[string]string{
		"email":    email,
		"password": password,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Пароль неверный")
}

func TestAuthHandler_Login_TokenGenerationError(t *testing.T) {
	email := "user@example.com"
	password := "password"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	// Настройка моков
	tokenMgr := &MockTokenManager{
		mockGenerateAccessToken: func(claims entity.UserClaims) (string, error) {
			return "", errors.New("token generation failed")
		},
	}

	userRepo := &MockUserRepository{
		mockGetUserByEmail: func(ctx context.Context, e string) (*entity.User, error) {
			return &entity.User{
				Email:    email,
				Password: string(hashedPassword),
			}, nil
		},
	}

	// Создание обработчика
	handler := NewAuthHandler(
		tokenMgr,
		config.JWTConfig{},
		userRepo,
		NewTestLoggerWrapper(),
	)

	// Подготовка запроса
	reqBody := map[string]string{
		"email":    email,
		"password": password,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Не удалось создать access token")
}
