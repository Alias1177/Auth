package usecase

import "github.com/Alias1177/Auth/internal/entity"

// TokenManager interface реализация по JWT
type TokenManager interface {
	GenerateAccessToken(userClaims entity.UserClaims) (string, error)
	ValidateAccessToken(token string) (*entity.UserClaims, error)
	GenerateRefreshToken(userClaims entity.UserClaims) (string, error)
	RefreshTokens(refreshToken string) (string, string, error)
}
