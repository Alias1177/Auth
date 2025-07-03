package service

import (
	"github.com/Alias1177/Auth/internal/domain"
	"github.com/Alias1177/Auth/pkg/jwt"
)

type TokenServiceImpl struct {
	jwtManager *jwt.JWTTokenManager
}

func NewTokenService(jwtManager *jwt.JWTTokenManager) *TokenServiceImpl {
	return &TokenServiceImpl{jwtManager: jwtManager}
}

func (s *TokenServiceImpl) GenerateAccessToken(userClaims domain.UserClaims) (string, error) {
	return s.jwtManager.GenerateAccessToken(userClaims)
}

func (s *TokenServiceImpl) GenerateRefreshToken(userClaims domain.UserClaims) (string, error) {
	return s.jwtManager.GenerateRefreshToken(userClaims)
}

func (s *TokenServiceImpl) ValidateAccessToken(token string) (*domain.UserClaims, error) {
	return s.jwtManager.ValidateAccessToken(token)
}

func (s *TokenServiceImpl) ValidateRefreshToken(token string) (*domain.UserClaims, error) {
	return s.jwtManager.ValidateRefreshToken(token)
}
