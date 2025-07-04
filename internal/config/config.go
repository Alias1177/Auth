package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/redis/go-redis/v9"
)

// RedisConfig конфигурация для Redis
type RedisConfig struct {
	Addr     string `env:"REDIS_ADDR" env-default:"localhost:6379"`
	Password string `env:"REDIS_PASSWORD" env-default:""`
	DB       int    `env:"REDIS_DB" env-default:"0"`
}

type JWTConfig struct {
	Secret string `env:"JWT_SECRET"`
}

// DatabaseConfig конфигурация для PostgreSQL
type DatabaseConfig struct {
	DSN string `env:"DATABASE_DSN" env-required:"true"`
}

// KafkaConfig конфигурация для Kafka
type KafkaConfig struct {
	BrokerAddress string `env:"KAFKA_BROKER_ADDRESS" env-default:"localhost:29092"`
	EmailTopic    string `env:"KAFKA_EMAIL_TOPIC" env-default:"user_registrations"`
}

// AppConfig конфигурация приложения
type AppConfig struct {
	Environment string `env:"APP_ENV" env-default:"development"`
	Debug       bool   `env:"APP_DEBUG" env-default:"true"`
}

// Config общая конфигурация приложения
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Kafka    KafkaConfig
}

// NewRedisClient создает новый клиент Redis на основе конфигурации
func NewRedisClient(cfg RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
}

// Load загружает конфигурацию из файла и переменных окружения
func Load(path string) (*Config, error) {
	cfg := &Config{}

	// Читаем конфигурацию из файла
	err := cleanenv.ReadConfig(path, cfg)
	if err != nil {
		log.Printf("Warning: не удалось прочитать конфигурационный файл: %v", err)
	}

	// Читаем переменные окружения
	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
