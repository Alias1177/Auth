package redis_migration

import (
	"context"
	"fmt"

	"github.com/Alias1177/Auth/pkg/logger"
	"github.com/Alias1177/Auth/pkg/migration"
	"github.com/redis/go-redis/v9"
)

// RedisMigrator управляет миграциями Redis.
type RedisMigrator struct {
	client *redis.Client
	log    *logger.Logger
}

// NewRedisMigrator создаёт новый мигратор для Redis.
func NewRedisMigrator(client *redis.Client, log *logger.Logger) *RedisMigrator {
	return &RedisMigrator{
		client: client,
		log:    log,
	}
}

// Up выполняет миграции Redis.
func (r *RedisMigrator) Up(ctx context.Context) error {
	r.log.Infow("Starting Redis migrations...")

	// Проверяем, есть ли уже данные
	exists, err := r.client.Exists(ctx, "user:1").Result()
	if err != nil {
		r.log.Errorw(migration.ErrCheckKeys, "error", err)
		return fmt.Errorf("%s: %w", migration.ErrCheckKeys, err)
	}

	// Если данных нет, добавляем тестового пользователя
	if exists == 0 {
		r.log.Infow("No users found in Redis, seeding initial data...")

		// Пример пользователя
		userJSON := `{"ID":1,"UserName":"testName","Email":"test@example.com","Password":"$2a$10$784zAhPwnidm32M4joFIb.TLer2jLe0iT84szRKiIRELyo0PxxwrG"}`

		if err := r.client.Set(ctx, "user:1", userJSON, 0).Err(); err != nil {
			r.log.Errorw(migration.ErrInsertUser, "error", err)
			return fmt.Errorf("%s: %w", migration.ErrInsertUser, err)
		}
		r.log.Infow("Initial user added successfully")
	} else {
		r.log.Infow("Users already exist in Redis, skipping seeding")
	}

	return nil
}

// Down откатывает миграции (удаляет тестовые данные).
func (r *RedisMigrator) Down(ctx context.Context) error {
	r.log.Infow("Rolling back Redis migrations...")

	// Удаляем тестовые данные
	if err := r.client.Del(ctx, "user:1").Err(); err != nil {
		r.log.Errorw(migration.ErrDeleteUser, "error", err)
		return fmt.Errorf("%s: %w", migration.ErrDeleteUser, err)
	}

	r.log.Infow("Test user deleted successfully")
	return nil
}
