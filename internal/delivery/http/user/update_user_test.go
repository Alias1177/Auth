package user_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Alias1177/Auth/internal/delivery/http/user"
	"github.com/Alias1177/Auth/internal/entity"
	"github.com/Alias1177/Auth/pkg/logger"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository для тестирования
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id int) (*entity.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*entity.User), args.Error(1)
}

func TestUpdateUser_Success(t *testing.T) {
	// Инициализация моков
	mockRepo := new(MockUserRepository)
	mockLogger := &logger.Logger{} // Используем реальный логгер или можно создать мок

	// Настройка ожидаемого вызова
	testUser := &entity.User{
		ID:       1,
		UserName: "testuser",
		Email:    "test@example.com",
	}
	mockRepo.On("UpdateUser", mock.Anything, testUser).Return(nil)

	// Создание handler
	handler := user.NewUserHandler(mockRepo, mockLogger)

	// Создание тестового запроса
	userJSON, _ := json.Marshal(testUser)
	req := httptest.NewRequest("PUT", "/users/1", bytes.NewBuffer(userJSON))
	req.Header.Set("Content-Type", "application/json")

	// Создание роутера chi
	r := chi.NewRouter()
	r.Put("/users/{id}", handler.UpdateUser)

	// Запись ответа
	w := httptest.NewRecorder()

	// Выполнение запроса
	r.ServeHTTP(w, req)

	// Проверки
	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestUpdateUser_InvalidID(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockLogger := &logger.Logger{}
	handler := user.NewUserHandler(mockRepo, mockLogger)

	req := httptest.NewRequest("PUT", "/users/invalid", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Put("/users/{id}", handler.UpdateUser)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
