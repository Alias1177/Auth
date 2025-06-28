package usecase

import (
	"context"

	"github.com/Alias1177/Auth/internal/entity"
)

type UserCache interface {
	GetUser(ctx context.Context, id int) (*entity.User, error)
	SaveUser(ctx context.Context, user *entity.User) error
}
