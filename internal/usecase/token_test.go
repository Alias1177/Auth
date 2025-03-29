package usecase_test

import (
	"Auth/internal/entity"
	"Auth/internal/usecase"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Моковая реализация TokenManager
type mockTokenManager struct {
	mock.Mock
}

func (m *mockTokenManager) GenerateAccessToken(claims entity.UserClaims) (string, error) {
	args := m.Called(claims)
	return args.String(0), args.Error(1)
}

func (m *mockTokenManager) ValidateAccessToken(token string) (*entity.UserClaims, error) {
	args := m.Called(token)
	return args.Get(0).(*entity.UserClaims), args.Error(1)
}

// Пример использования токенов (можешь заменить на свой реальный код)
func doTokenStuff(tman usecase.TokenManager) (string, *entity.UserClaims, error) {
	claims := entity.UserClaims{
		UserID: "123",
		Email:  "test@example.com",
	}
	token, err := tman.GenerateAccessToken(claims)
	if err != nil {
		return "", nil, err
	}

	parsedClaims, err := tman.ValidateAccessToken(token)
	return token, parsedClaims, err
}

func TestTokenManager(t *testing.T) {
	mockToken := new(mockTokenManager)

	testClaims := entity.UserClaims{
		UserID: "123",
		Email:  "test@example.com",
	}

	mockToken.On("GenerateAccessToken", testClaims).Return("mocked-token", nil)
	mockToken.On("ValidateAccessToken", "mocked-token").Return(&testClaims, nil)

	token, claims, err := doTokenStuff(mockToken)

	assert.NoError(t, err)
	assert.Equal(t, "mocked-token", token)
	assert.Equal(t, testClaims.UserID, claims.UserID)
	assert.Equal(t, testClaims.Email, claims.Email)

	mockToken.AssertExpectations(t)
}
