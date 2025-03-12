package usecase

import (
	"Auth/internal/entity"
	"context"
)

type UserCache interface {
	GetUser(ctx context.Context, id int) (*entity.User, error)
	SaveUser(ctx context.Context, user *entity.User) error
}
