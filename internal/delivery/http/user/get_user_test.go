package user

import (
	"Auth/internal/entity"
	"Auth/internal/infrastructure/middleware"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockUserRepository реализует UserRepository для тестов
type MockUserRepository struct {
	mockGetUserByID func(context.Context, int) (*entity.User, error)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id int) (*entity.User, error) {
	return m.mockGetUserByID(ctx, id)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	return nil, nil
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *entity.User) error {
	return nil
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, user *entity.User) error {
	return nil
}

func TestUserHandler_GetUserInfoHandler_Success(t *testing.T) {
	// Подготовка тестовых данных
	userID := 123
	testUser := &entity.User{
		ID:       userID,
		Email:    "test@example.com",
		UserName: "testuser",
	}

	// Настройка моков
	userRepo := &MockUserRepository{
		mockGetUserByID: func(ctx context.Context, id int) (*entity.User, error) {
			assert.Equal(t, userID, id)
			return testUser, nil
		},
	}

	handler := &UserHandler{
		userRepository: userRepo,
	}

	// Создание тестового запроса с контекстом
	req := httptest.NewRequest("GET", "/user", nil)
	ctx := context.WithValue(req.Context(), middleware.CtxUserKey, &entity.UserClaims{
		UserID: strconv.Itoa(userID),
		Email:  testUser.Email,
	})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Вызов метода
	handler.GetUserInfoHandler(w, req)

	// Проверки
	assert.Equal(t, http.StatusOK, w.Code)

	var responseUser entity.User
	err := json.NewDecoder(w.Body).Decode(&responseUser)
	assert.NoError(t, err)
	assert.Equal(t, *testUser, responseUser)
}

func TestUserHandler_GetUserInfoHandler_NoUserInContext(t *testing.T) {
	handler := &UserHandler{
		userRepository: &MockUserRepository{},
	}

	// Запрос без пользователя в контексте
	req := httptest.NewRequest("GET", "/user", nil)
	w := httptest.NewRecorder()

	handler.GetUserInfoHandler(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Ошибка получения информации о пользователе")
}

func TestUserHandler_GetUserInfoHandler_InvalidUserID(t *testing.T) {
	handler := &UserHandler{
		userRepository: &MockUserRepository{},
	}

	// Запрос с некорректным ID пользователя
	req := httptest.NewRequest("GET", "/user", nil)
	ctx := context.WithValue(req.Context(), middleware.CtxUserKey, &entity.UserClaims{
		UserID: "invalid_id",
	})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetUserInfoHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Некорректный ID пользователя")
}

func TestUserHandler_GetUserInfoHandler_UserNotFound(t *testing.T) {
	userID := 123

	// Настройка моков
	userRepo := &MockUserRepository{
		mockGetUserByID: func(ctx context.Context, id int) (*entity.User, error) {
			return nil, errors.New("user not found")
		},
	}

	handler := &UserHandler{
		userRepository: userRepo,
	}

	// Создание тестового запроса
	req := httptest.NewRequest("GET", "/user", nil)
	ctx := context.WithValue(req.Context(), middleware.CtxUserKey, &entity.UserClaims{
		UserID: strconv.Itoa(userID),
	})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetUserInfoHandler(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "user not found")
}
