package redis_migration

import (
	"Auth/pkg/logger"
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func setupTestRedis() (*redis.Client, func()) {
	// Подключение к локальному Redis (или mock-серверу)
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1, // Используем отдельную тестовую БД
	})

	// Очистка перед и после тестов
	client.FlushDB(context.Background())

	// Возвращаем клиент и функцию для очистки после тестов
	return client, func() {
		client.FlushDB(context.Background())
		client.Close()
	}
}

func TestRedisMigrator_Up(t *testing.T) {
	ctx := context.Background()
	client, cleanup := setupTestRedis()
	defer cleanup()

	logInstance, _ := logger.NewSimpleLogger("info")
	migrator := NewRedisMigrator(client, logInstance)

	// 1. Проверяем, что ключа еще нет
	exists, err := client.Exists(ctx, "user:1").Result()
	require.NoError(t, err)
	assert.Equal(t, int64(0), exists)

	// 2. Запускаем миграцию
	err = migrator.Up(ctx)
	require.NoError(t, err)

	// 3. Проверяем, что ключ появился
	exists, err = client.Exists(ctx, "user:1").Result()
	require.NoError(t, err)
	assert.Equal(t, int64(1), exists)

	// 4. Повторный запуск Up() не должен добавить дубликаты
	err = migrator.Up(ctx)
	require.NoError(t, err)

	// Данные не должны измениться
	exists, err = client.Exists(ctx, "user:1").Result()
	require.NoError(t, err)
	assert.Equal(t, int64(1), exists)
}

func TestRedisMigrator_Down(t *testing.T) {
	ctx := context.Background()
	client, cleanup := setupTestRedis()
	defer cleanup()

	logInstance, _ := logger.NewSimpleLogger("info")
	migrator := NewRedisMigrator(client, logInstance)

	// Предварительно заполняем Redis
	err := client.Set(ctx, "user:1", `{"ID":1,"UserName":"testName"}`, 0).Err()
	require.NoError(t, err)

	// Проверяем, что данные есть
	exists, err := client.Exists(ctx, "user:1").Result()
	require.NoError(t, err)
	assert.Equal(t, int64(1), exists)

	// Выполняем откат миграции
	err = migrator.Down(ctx)
	require.NoError(t, err)

	// Проверяем, что данные удалены
	exists, err = client.Exists(ctx, "user:1").Result()
	require.NoError(t, err)
	assert.Equal(t, int64(0), exists)
}
