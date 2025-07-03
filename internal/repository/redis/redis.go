package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Alias1177/Auth/internal/domain"
	"github.com/Alias1177/Auth/pkg/logger"
	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	client *redis.Client // Redis клиент для взаимодействия с сервером Redis.
	log    *logger.Logger
}

// NewRedisRepository создает новый экземпляр RedisRepository.
func NewRedisRepository(client *redis.Client, log *logger.Logger) *RedisRepository {
	return &RedisRepository{
		client: client,
		log:    log,
	}
}

// GetUser получает данные пользователя из Redis по ID.
func (r *RedisRepository) GetUser(ctx context.Context, id int) (*domain.User, error) {
	// Формируем ключ для хранения данных в Redis.
	key := fmt.Sprintf("user:%d", id)

	// Получаем данные из Redis.
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var user domain.User
	// Десериализуем данные из JSON.
	if err := json.Unmarshal(data, &user); err != nil {
		r.log.Errorw("Unmarshal err", "err", err)
		return nil, fmt.Errorf("failed to unmarshal user data: %w", err)
	}
	return &user, nil
}

// SaveUser сохраняет данные пользователя в Redis.
func (r *RedisRepository) SaveUser(ctx context.Context, user *domain.User) error {
	// Сериализуем данные пользователя в JSON.
	jsonData, err := json.Marshal(user)
	if err != nil {
		r.log.Errorw("Marshal err", "err", err)
		return fmt.Errorf("failed to marshal user data: %w", err)
	}

	// Формируем ключ для хранения данных.
	key := fmt.Sprintf("user:%d", user.ID)

	// Сохраняем данные в Redis.
	return r.client.Set(ctx, key, jsonData, 0).Err()
}

// SetUser сохраняет данные пользователя в Redis (альтернативный метод).
func (r *RedisRepository) SetUser(ctx context.Context, user *domain.User) error {
	key := fmt.Sprintf("user:%d", user.ID)
	value, err := json.Marshal(user)
	if err != nil {
		r.log.Errorw("Marshal err", "err", err)
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	err = r.client.Set(ctx, key, value, 0).Err()
	if err != nil {
		// Check if it's a read-only error
		if strings.Contains(err.Error(), "READONLY") {
			r.log.Warnw("Redis is in read-only mode, skipping cache update",
				"error", err,
				"user_id", user.ID)
			return nil // Return nil to allow the operation to continue
		}
		r.log.Errorw("Set err", "err", err)
		return fmt.Errorf("failed to set user in redis: %w", err)
	}

	return nil
}

// Close закрывает соединение с Redis.
func (r *RedisRepository) Close() error {
	if r.client == nil {
		return nil
	}

	err := r.client.Close()
	if err != nil {
		r.log.Errorw("Failed to close Redis connection", "err", err)
		return fmt.Errorf("failed to close Redis client: %w", err)
	}

	r.log.Infow("Redis connection closed successfully")
	return nil
}
