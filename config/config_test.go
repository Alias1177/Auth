package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfigFromFile(t *testing.T) {
	// Создаем временный конфигурационный файл
	content := `
database:
  dsn: "postgres://user:password@localhost:5432/dbname"
redis:
  addr: "localhost:6379"
  password: "secret"
  db: 1
jwt:
  secret: "jwt_secret"
`
	tmpFile, err := os.CreateTemp("", "configs.*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	assert.NoError(t, err)
	tmpFile.Close()

	// Загружаем конфигурацию
	cfg, err := Load(tmpFile.Name())
	assert.NoError(t, err)

	// Проверяем значения
	assert.Equal(t, "postgres://user:password@localhost:5432/dbname", cfg.Database.DSN)
	assert.Equal(t, "localhost:6379", cfg.Redis.Addr)
	assert.Equal(t, "secret", cfg.Redis.Password)
	assert.Equal(t, 1, cfg.Redis.DB)
	assert.Equal(t, "jwt_secret", cfg.JWT.Secret)
}

func TestLoadConfigFromEnv(t *testing.T) {
	// Устанавливаем переменные окружения
	os.Setenv("DATABASE_DSN", "postgres://user:password@localhost:5432/dbname")
	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("REDIS_PASSWORD", "secret")
	os.Setenv("REDIS_DB", "1")
	os.Setenv("JWT_SECRET", "jwt_secret")
	defer func() {
		os.Unsetenv("DATABASE_DSN")
		os.Unsetenv("REDIS_ADDR")
		os.Unsetenv("REDIS_PASSWORD")
		os.Unsetenv("REDIS_DB")
		os.Unsetenv("JWT_SECRET")
	}()

	// Загружаем конфигурацию
	cfg, err := Load("")
	assert.NoError(t, err)

	// Проверяем значения
	assert.Equal(t, "postgres://user:password@localhost:5432/dbname", cfg.Database.DSN)
	assert.Equal(t, "localhost:6379", cfg.Redis.Addr)
	assert.Equal(t, "secret", cfg.Redis.Password)
	assert.Equal(t, 1, cfg.Redis.DB)
	assert.Equal(t, "jwt_secret", cfg.JWT.Secret)
}

func TestNewRedisClient(t *testing.T) {
	cfg := RedisConfig{
		Addr:     "localhost:6379",
		Password: "secret",
		DB:       1,
	}

	client := NewRedisClient(cfg)
	assert.NotNil(t, client)

	// Проверяем, что клиент создан с правильными параметрами
	options := client.Options()
	assert.Equal(t, "localhost:6379", options.Addr)
	assert.Equal(t, "secret", options.Password)
	assert.Equal(t, 1, options.DB)
}

func TestLoadConfigWithMissingRequiredField(t *testing.T) {
	// Устанавливаем переменные окружения, но не указываем DATABASE_DSN
	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("REDIS_PASSWORD", "secret")
	os.Setenv("REDIS_DB", "1")
	os.Setenv("JWT_SECRET", "jwt_secret")
	defer func() {
		os.Unsetenv("REDIS_ADDR")
		os.Unsetenv("REDIS_PASSWORD")
		os.Unsetenv("REDIS_DB")
		os.Unsetenv("JWT_SECRET")
	}()

	// Загружаем конфигурацию
	_, err := Load("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field \"DSN\" is required") // Обновляем проверку
}
