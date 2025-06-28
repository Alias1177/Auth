package usecase_test

import (
	"github.com/Alias1177/Auth/internal/entity"
	//"Auth/internal/usecase"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Моковая реализация UserRepository
type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) CreateUser(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) GetUserByID(ctx context.Context, id int) (*entity.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *mockUserRepo) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *mockUserRepo) UpdateUser(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func TestUserRepository(t *testing.T) {
	mockRepo := new(mockUserRepo)
	ctx := context.TODO()

	user := &entity.User{
		ID:        1,
		UserName:  "vladimir",
		Email:     "vladimir@example.com",
		Password:  "hashed_password",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Ожидания моков
	mockRepo.On("CreateUser", ctx, user).Return(nil)
	mockRepo.On("GetUserByID", ctx, 1).Return(user, nil)
	mockRepo.On("GetUserByEmail", ctx, "vladimir@example.com").Return(user, nil)
	mockRepo.On("UpdateUser", ctx, user).Return(nil)

	// Тестирование методов
	err := mockRepo.CreateUser(ctx, user)
	assert.NoError(t, err)

	foundByID, err := mockRepo.GetUserByID(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, foundByID.ID)

	foundByEmail, err := mockRepo.GetUserByEmail(ctx, "vladimir@example.com")
	assert.NoError(t, err)
	assert.Equal(t, user.Email, foundByEmail.Email)

	err = mockRepo.UpdateUser(ctx, user)
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}
