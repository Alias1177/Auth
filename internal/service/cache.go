package service

import (
	"context"
	"time"

	"github.com/Alias1177/Auth/internal/domain"
)

type UserCache interface {
	GetUser(ctx context.Context, id int) (*domain.User, error)
	SaveUser(ctx context.Context, user *domain.User) error
	// Методы для работы с произвольными ключами и TTL
	Get(ctx context.Context, key string) (string, error)
	SetWithTTL(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}
