package usecase

import "Auth/internal/entity"

// TokenManager interface реализация по JWT
type TokenManager interface {
	GenerateAccessToken(userClaims entity.UserClaims) (string, error)
	GenerateRefreshToken(userClaims entity.UserClaims) (string, error)
	ParseRefreshToken(tokenStr string) (entity.UserClaims, error)
	ValidateAccessToken(token string) (*entity.UserClaims, error)
	ValidateRefreshToken(token string) (*entity.UserClaims, error)
}
