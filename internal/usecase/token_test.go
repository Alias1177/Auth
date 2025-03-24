package usecase_test

import (
	"Auth/internal/entity"
	"Auth/internal/usecase/mocks_token"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGenerateAccessToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTokenManager := mocks_token.NewMockTokenManager(ctrl)

	claims := entity.UserClaims{
		UserID:    "12345",
		Email:     "karl@example.com",
		ExpiresAt: 1700000000,
	}
	expectedToken := "mocked.jwt.token"

	mockTokenManager.EXPECT().
		GenerateAccessToken(claims).
		Return(expectedToken, nil)

	token, err := mockTokenManager.GenerateAccessToken(claims)

	assert.NoError(t, err)
	assert.Equal(t, expectedToken, token)
}

func TestValidateAccessToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTokenManager := mocks_token.NewMockTokenManager(ctrl)

	token := "mocked.jwt.token"
	expectedClaims := &entity.UserClaims{
		UserID:    "67890",
		Email:     "vladimir@example.com",
		ExpiresAt: 1800000000,
	}

	mockTokenManager.EXPECT().
		ValidateAccessToken(token).
		Return(expectedClaims, nil)

	claims, err := mockTokenManager.ValidateAccessToken(token)

	assert.NoError(t, err)
	assert.Equal(t, expectedClaims, claims)
}
