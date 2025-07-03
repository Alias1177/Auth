package service

import (
	"github.com/Alias1177/Auth/internal/domain"
)

// TokenManager interface реализация по JWT
type TokenManager interface {
	GenerateAccessToken(userClaims domain.UserClaims) (string, error)
	ValidateAccessToken(token string) (*domain.UserClaims, error)
	GenerateRefreshToken(userClaims domain.UserClaims) (string, error)
	RefreshTokens(refreshToken string) (string, string, error)
	ValidateRefreshToken(token string) (*domain.UserClaims, error)
}
