package registration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Alias1177/Auth/config"
	"github.com/Alias1177/Auth/internal/entity"
	"github.com/Alias1177/Auth/pkg/kafka"
	"github.com/Alias1177/Auth/pkg/logger"

	"github.com/stretchr/testify/assert"
)

// MockUserRepository реализует usecase.UserRepository
type MockUserRepository struct{}

func (r *MockUserRepository) CreateUser(ctx context.Context, user *entity.User) error {
	return nil
}

func (r *MockUserRepository) GetUserByID(ctx context.Context, id int) (*entity.User, error) {
	return nil, nil
}

func (r *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	return nil, nil
}

func (r *MockUserRepository) UpdateUser(ctx context.Context, user *entity.User) error {
	return nil
}

// MockTokenManager реализует usecase.TokenManager
type MockTokenManager struct{}

func (m *MockTokenManager) GenerateAccessToken(userClaims entity.UserClaims) (string, error) {
	return "test-token", nil
}

func (m *MockTokenManager) ValidateAccessToken(token string) (*entity.UserClaims, error) {
	return nil, nil
}

// MockLogger реализует logger.Logger
type MockLogger struct {
	logger *logger.Logger
}

func NewMockLogger() *logger.Logger {
	return &logger.Logger{} // Возвращаем реальный тип *logger.Logger
}

func (l *MockLogger) Infow(msg string, keysAndValues ...any)  {}
func (l *MockLogger) Errorw(msg string, keysAndValues ...any) {}
func (l *MockLogger) Warnw(msg string, keysAndValues ...any)  {}
func (l *MockLogger) Debugw(msg string, keysAndValues ...any) {}
func (l *MockLogger) Fatalw(msg string, keysAndValues ...any) {}
func (l *MockLogger) Close() error                            { return nil }

// MockKafkaProducer реализует kafka.Producer
type MockKafkaProducer struct {
	producer *kafka.Producer
}

func NewMockKafkaProducer() *kafka.Producer {
	return &kafka.Producer{} // Возвращаем реальный тип *kafka.Producer
}

func (p *MockKafkaProducer) Send(ctx context.Context, topic string, message any) error {
	return nil
}

func (p *MockKafkaProducer) SendEmailRegistration(ctx context.Context, email, username string) error {
	return nil
}

func (p *MockKafkaProducer) Close() error {
	return nil
}

func TestRegistrationHandler_Register_Success(t *testing.T) {
	// Инициализация тестовых зависимостей
	handler := NewRegistrationHandler(
		&MockUserRepository{},
		&MockTokenManager{},
		config.JWTConfig{Secret: "test-secret"},
		NewMockLogger(),        // Возвращаем правильный тип
		NewMockKafkaProducer(), // Возвращаем правильный тип
	)

	// Подготовка запроса
	reqBody := map[string]string{
		"email":    "test@example.com",
		"username": "testuser",
		"password": "securepassword",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Вызов метода
	handler.Register(w, req)

	// Проверки
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Пользователь успешно зарегистрирован", response["message"])
}

func TestRegistrationHandler_Register_InvalidRequest(t *testing.T) {
	handler := NewRegistrationHandler(
		nil,
		nil,
		config.JWTConfig{},
		NewMockLogger(), // Возвращаем правильный тип
		nil,
	)

	// Неправильный JSON
	req := httptest.NewRequest("POST", "/register", bytes.NewReader([]byte("{invalid}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Некорректный запрос")
}
