package service

import (
	"context"

	"github.com/Alias1177/Auth/internal/domain"
	"github.com/Alias1177/Auth/internal/errors"
	crypto "github.com/Alias1177/Auth/pkg/security"
)

type AuthServiceImpl struct {
	userRepo     UserRepository
	tokenManager TokenManager
}

func NewAuthService(userRepo UserRepository, tokenManager TokenManager) *AuthServiceImpl {
	return &AuthServiceImpl{
		userRepo:     userRepo,
		tokenManager: tokenManager,
	}
}

func (s *AuthServiceImpl) Login(ctx context.Context, email, password string) (*domain.User, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if err := crypto.VerifyPassword(user.Password, password); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthServiceImpl) Register(ctx context.Context, user *domain.User) error {
	existing, err := s.userRepo.GetUserByEmail(ctx, user.Email)
	if err == nil && existing != nil {
		return errors.ErrUserExists
	}
	hashedPassword, err := crypto.HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword
	return s.userRepo.CreateUser(ctx, user)
}

func (s *AuthServiceImpl) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	claims, err := s.tokenManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", err
	}
	accessToken, err := s.tokenManager.GenerateAccessToken(*claims)
	if err != nil {
		return "", err
	}
	return accessToken, nil
}

func (s *AuthServiceImpl) Logout(ctx context.Context, userID string) error {
	// Здесь может быть логика blacklist токенов или удаление refresh токена
	return nil // no-op
}
