package service

import (
	"context"

	"github.com/Alias1177/Auth/internal/domain"
)

type UserServiceImpl struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserServiceImpl {
	return &UserServiceImpl{repo: repo}
}

func (s *UserServiceImpl) GetByID(ctx context.Context, id int) (*domain.User, error) {
	return s.repo.GetUserByID(ctx, id)
}

func (s *UserServiceImpl) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.repo.GetUserByEmail(ctx, email)
}

func (s *UserServiceImpl) Create(ctx context.Context, user *domain.User) error {
	return s.repo.CreateUser(ctx, user)
}

func (s *UserServiceImpl) Update(ctx context.Context, user *domain.User) error {
	return s.repo.UpdateUser(ctx, user)
}

func (s *UserServiceImpl) ResetPassword(ctx context.Context, user *domain.User) error {
	return s.repo.ResetPassword(ctx, user)
}
