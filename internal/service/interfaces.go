package service

import (
	"context"

	"github.com/Alias1177/Auth/internal/domain"
)

// UserService интерфейс для работы с пользователями
type UserService interface {
	GetByID(ctx context.Context, id int) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	ResetPassword(ctx context.Context, email, password string) error
}

// AuthService интерфейс для аутентификации
type AuthService interface {
	Login(ctx context.Context, email, password string) (*domain.User, error)
	Register(ctx context.Context, user *domain.User) error
	RefreshToken(ctx context.Context, refreshToken string) (string, error)
	Logout(ctx context.Context, userID string) error
}

// TokenService интерфейс для работы с токенами
type TokenService interface {
	GenerateAccessToken(userClaims domain.UserClaims) (string, error)
	GenerateRefreshToken(userClaims domain.UserClaims) (string, error)
	ValidateAccessToken(token string) (*domain.UserClaims, error)
	ValidateRefreshToken(token string) (*domain.UserClaims, error)
}

// PasswordService интерфейс для работы с паролями
type PasswordService interface {
	HashPassword(password string) (string, error)
	VerifyPassword(hashedPassword, password string) error
}
