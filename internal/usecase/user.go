package usecase

import (
	"Auth/internal/entity"
	"context"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *entity.User) error
	GetUserByID(ctx context.Context, id int) (*entity.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	UpdateUser(ctx context.Context, user *entity.User) error
}
