// internal/repository/redis/redis.go
package redis

import (
	"Auth/internal/entity"
	"Auth/pkg/logger"
	"context"
	"encoding/json"
	"fmt"
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
func (r *RedisRepository) GetUser(ctx context.Context, id int) (*entity.User, error) {
	// Формируем ключ для хранения данных в Redis.
	key := fmt.Sprintf("user:%d", id)

	// Получаем данные из Redis.
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var user entity.User
	// Десериализуем данные из JSON.
	if err := json.Unmarshal(data, &user); err != nil {
		r.log.Errorw("Unmarshal err", "err", err)
		return nil, fmt.Errorf("failed to unmarshal user data: %w", err)
	}
	return &user, nil
}

// SaveUser сохраняет данные пользователя в Redis.
func (r *RedisRepository) SaveUser(ctx context.Context, user *entity.User) error {
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

func (r *RedisRepository) SetUser(ctx context.Context, user *entity.User) error {
	key := fmt.Sprintf("user:%d", user.ID)
	value, err := json.Marshal(user)
	if err != nil {
		r.log.Errorw("Marshal err", "err", err)
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	err = r.client.Set(ctx, key, value, 0).Err()
	if err != nil {
		r.log.Errorw("Set err", "err", err)
		return fmt.Errorf("failed to set user in redis: %w", err)
	}

	return nil
}
