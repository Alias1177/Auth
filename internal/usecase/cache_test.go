package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/Alias1177/Auth/internal/entity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Мок UserCache
type MockUserCache struct {
	mock.Mock
}

func (m *MockUserCache) GetUser(ctx context.Context, id int) (*entity.User, error) {
	args := m.Called(ctx, id)
	if user := args.Get(0); user != nil {
		return user.(*entity.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserCache) SaveUser(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func TestUserCache_GetUser(t *testing.T) {
	ctx := context.Background()
	mockCache := new(MockUserCache)

	expectedUser := &entity.User{
		ID:        1,
		UserName:  "vladimir",
		Email:     "vladimir@example.com",
		Password:  "securepass123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockCache.On("GetUser", ctx, 1).Return(expectedUser, nil)

	user, err := mockCache.GetUser(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	mockCache.AssertExpectations(t)
}

func TestUserCache_SaveUser(t *testing.T) {
	ctx := context.Background()
	mockCache := new(MockUserCache)

	user := &entity.User{
		ID:        2,
		UserName:  "karl",
		Email:     "karl@example.com",
		Password:  "anotherpass456",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockCache.On("SaveUser", ctx, user).Return(nil)

	err := mockCache.SaveUser(ctx, user)

	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}
