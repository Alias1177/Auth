// internal/repository/repository.go
package repository

import (
	"Auth/internal/entity"
	"Auth/internal/repository/postgres"
	"Auth/internal/repository/redis"
	"Auth/pkg/logger"
	"context"
	"fmt"
	"log/slog"
)

// Repository представляет собой агрегатор репозиториев PostgreSQL и Redis.
type Repository struct {
	postgres *postgres.PostgresRepository // Репозиторий для работы с PostgreSQL.
	redis    *redis.RedisRepository
	log      *logger.Logger
}

// NewRepository создает новый экземпляр агрегированного репозитория.
func NewRepository(pg *postgres.PostgresRepository, rd *redis.RedisRepository, log *logger.Logger) *Repository {
	return &Repository{
		postgres: pg,
		redis:    rd,
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
		return fmt.Errorf("failed to save user to cache: %w", err)
	}

	return nil
}
