package usecase

import "Auth/internal/entity"

// TokenManager interface реализация по JWT
type TokenManager interface {
	GenerateAccessToken(userClaims entity.UserClaims) (string, error)
	ValidateAccessToken(token string) (*entity.UserClaims, error)
}
