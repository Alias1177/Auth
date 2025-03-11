// internal/repository/redis/redis.go
package redis

import (
	"Auth/internal/entity"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	client *redis.Client // Redis клиент для взаимодействия с сервером Redis.
}

// NewRedisRepository создает новый экземпляр RedisRepository.
func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{client: client}
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
		return nil, fmt.Errorf("failed to unmarshal user data: %w", err)
	}
	return &user, nil
}

// SaveUser сохраняет данные пользователя в Redis.
func (r *RedisRepository) SaveUser(ctx context.Context, user *entity.User) error {
	// Сериализуем данные пользователя в JSON.
	jsonData, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %w", err)
	}

	// Формируем ключ для хранения данных.
	key := fmt.Sprintf("user:%d", user.ID)

	// Сохраняем данные в Redis.
	return r.client.Set(ctx, key, jsonData, 0).Err()
}
