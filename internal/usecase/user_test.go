package usecase_test

import (
	"Auth/internal/entity"
	"Auth/internal/usecase/mocks_user"
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks_user.NewMockUserRepository(ctrl)
	ctx := context.Background()

	newUser := &entity.User{ID: 1, UserName: "karl", Email: "karl@example.com"}

	// Ожидаем, что метод CreateUser будет вызван с аргументами ctx и newUser
	mockRepo.EXPECT().
		CreateUser(ctx, newUser).
		Return(nil) // Ошибка не ожидается

	// Вызов метода
	err := mockRepo.CreateUser(ctx, newUser)

	// Проверка, что ошибка не возникла
	assert.NoError(t, err)
}

func TestGetUserByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks_user.NewMockUserRepository(ctrl)
	ctx := context.Background()

	// Создание ожидаемого пользователя
	expectedUser := &entity.User{ID: 1, UserName: "karl", Email: "karl@example.com"}

	// Ожидаем, что метод GetUserByID будет вызван с аргументом ctx и id = 1
	mockRepo.EXPECT().
		GetUserByID(ctx, 1).
		Return(expectedUser, nil) // Ожидаем возвращение пользователя и отсутствие ошибки

	// Вызов метода
	user, err := mockRepo.GetUserByID(ctx, 1)

	// Проверка, что ошибка не возникла и результат соответствует ожиданиям
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
}

func TestGetUserByEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks_user.NewMockUserRepository(ctrl)
	ctx := context.Background()

	// Создание ожидаемого пользователя
	expectedUser := &entity.User{ID: 1, UserName: "karl", Email: "karl@example.com"}

	// Ожидаем, что метод GetUserByEmail будет вызван с аргументом ctx и email = "karl@example.com"
	mockRepo.EXPECT().
		GetUserByEmail(ctx, "karl@example.com").
		Return(expectedUser, nil) // Ожидаем возвращение пользователя и отсутствие ошибки

	// Вызов метода
	user, err := mockRepo.GetUserByEmail(ctx, "karl@example.com")

	// Проверка, что ошибка не возникла и результат соответствует ожиданиям
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
}

func TestUpdateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks_user.NewMockUserRepository(ctrl)
	ctx := context.Background()

	// Создание обновленного пользователя
	updatedUser := &entity.User{ID: 1, UserName: "karl", Email: "karl_updated@example.com"}

	// Ожидаем, что метод UpdateUser будет вызван с аргументами ctx и updatedUser
	mockRepo.EXPECT().
		UpdateUser(ctx, updatedUser).
		Return(nil) // Ошибка не ожидается

	// Вызов метода
	err := mockRepo.UpdateUser(ctx, updatedUser)

	// Проверка, что ошибка не возникла
	assert.NoError(t, err)
}
