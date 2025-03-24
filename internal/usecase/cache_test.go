package usecase_test

import (
	"Auth/internal/entity"
	"Auth/internal/usecase/mocks_cache"
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUserCache_GetUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCache := mocks_cache.NewMockUserCache(ctrl)
	ctx := context.Background()

	expectedUser := &entity.User{ID: 1, UserName: "karl"}

	// Установка ожидания на мок
	mockCache.EXPECT().
		GetUser(ctx, 1).
		Return(expectedUser, nil)

	// Вызов
	user, err := mockCache.GetUser(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
}

func TestUserCache_SaveUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCache := mocks_cache.NewMockUserCache(ctrl)
	ctx := context.Background()

	testUser := &entity.User{ID: 2, UserName: "vladimir"}

	mockCache.EXPECT().
		SaveUser(ctx, testUser).
		Return(nil)

	err := mockCache.SaveUser(ctx, testUser)

	assert.NoError(t, err)
}
