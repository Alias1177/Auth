package service

import (
	"context"

	"github.com/Alias1177/Auth/internal/domain"
)

type UserCache interface {
	GetUser(ctx context.Context, id int) (*domain.User, error)
	SaveUser(ctx context.Context, user *domain.User) error
}
