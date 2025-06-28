// internal/repository/repository.go
package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Alias1177/Auth/internal/entity"
	"github.com/Alias1177/Auth/internal/usecase"
)

// Repository представляет собой агрегатор репозиториев PostgreSQL и Redis.
type Repository struct {
	postgres usecase.UserRepository
	redis    usecase.UserCache
	log      usecase.Logger
}

func NewRepository(pg usecase.UserRepository, redis usecase.UserCache, log usecase.Logger) *Repository {
	return &Repository{
		postgres: pg,
		redis:    redis,
		log:      log,
	}
}

// GetUser получает пользователя по его ID, используя кэш Redis и базу PostgreSQL.
func (r *Repository) GetUser(ctx context.Context, id int) (*entity.User, error) {
	// Сначала пробуем получить пользователя из Redis.
	user, err := r.redis.GetUser(ctx, id)
	if err == nil {
		return user, nil
	}

	// Если в Redis нет пользователя, получаем его из PostgreSQL.
	user, err = r.postgres.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Сохраняем данные пользователя в Redis для будущих запросов.
	if err := r.redis.SaveUser(ctx, user); err != nil {
		// Логируем ошибку сохранения в Redis, но не прерываем выполнение.
		slog.Error("failed to save user to Redis", "error", err)
	}

	return user, nil
}

// CreateUser создает нового пользователя в PostgreSQL и кэше Redis.
func (r *Repository) CreateUser(ctx context.Context, user *entity.User) error {
	// Проверяем, существует ли пользователь с таким email
	_, err := r.postgres.GetUserByEmail(ctx, user.Email)
	if err == nil {
		return fmt.Errorf("user with email %s already exists", user.Email)
	}

	// Создаем транзакцию в PostgreSQL
	if err := r.postgres.CreateUser(ctx, user); err != nil {
		return fmt.Errorf("failed to create user in database: %w", err)
	}

	// Сохраняем в Redis
	if err := r.redis.SaveUser(ctx, user); err != nil {
		// Если не удалось сохранить в Redis, это критическая ошибка,
		// так как мы должны получать данные только из Redis
		r.log.Warnw("Failed to save user to Redis cache", "error", err, "user_id", user.ID)
	}

	return nil
}

// GetUserByID получение пользователя по ID (прямая прокси без изменения логики)
func (r *Repository) GetUserByID(ctx context.Context, id int) (*entity.User, error) {
	return r.postgres.GetUserByID(ctx, id)
}

// GetUserByEmail получение пользователя по Email (прямая прокси без изменения логики)
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	return r.postgres.GetUserByEmail(ctx, email)
}

// UpdateUser обновление данных пользователя (прямая прокси без изменения логики)
func (r *Repository) UpdateUser(ctx context.Context, user *entity.User) error {
	return r.postgres.UpdateUser(ctx, user)
}
