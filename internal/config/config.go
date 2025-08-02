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
	DSN string `env:"DATABASE_DSN"`
}

// KafkaConfig конфигурация для Kafka
type KafkaConfig struct {
	BrokerAddress string `env:"KAFKA_BROKER_ADDRESS" env-default:"127.0.0.1:9092"`
	EmailTopic    string `env:"KAFKA_EMAIL_TOPIC" env-default:"notifications"`
}

// NotificationConfig конфигурация для Notification Service
type NotificationConfig struct {
	ServiceURL string `env:"NOTIFICATION_SERVICE_URL" env-default:"http://31.97.76.108:8080"`
}

// SentryConfig конфигурация для Sentry
type SentryConfig struct {
	DSN              string  `env:"SENTRY_DSN" env-default:""`
	Environment      string  `env:"SENTRY_ENVIRONMENT" env-default:"development"`
	Debug            bool    `env:"SENTRY_DEBUG" env-default:"false"`
	TracesSampleRate float64 `env:"SENTRY_TRACES_SAMPLE_RATE" env-default:"1.0"`
	EnableTracing    bool    `env:"SENTRY_ENABLE_TRACING" env-default:"true"`
}

// AppConfig конфигурация приложения
type AppConfig struct {
	Environment string `env:"APP_ENV" env-default:"development"`
	Debug       bool   `env:"APP_DEBUG" env-default:"true"`
}

type GoogleConfig struct {
	ClientID     string `env:"GOOGLE_CLIENT_ID"`
	ClientSecret string `env:"GOOGLE_CLIENT_SECRET"`
	RedirectURL  string `env:"GOOGLE_REDIRECT_URL"`
}

// Config общая конфигурация приложения
type Config struct {
	App          AppConfig
	Database     DatabaseConfig
	Redis        RedisConfig
	JWT          JWTConfig
	Kafka        KafkaConfig
	Notification NotificationConfig
	Sentry       SentryConfig
	Google       GoogleConfig
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

	// Логируем загруженную конфигурацию для диагностики
	log.Printf("Loaded Kafka config - BrokerAddress: %s, Topic: %s", cfg.Kafka.BrokerAddress, cfg.Kafka.EmailTopic)
	log.Printf("Loaded Notification config - ServiceURL: %s", cfg.Notification.ServiceURL)

	return cfg, nil
}

// LoadFromEnv загружает конфигурацию только из переменных окружения
func LoadFromEnv(cfg *Config) error {
	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		return err
	}

	// Логируем загруженную конфигурацию для диагностики
	log.Printf("Loaded from ENV - Kafka BrokerAddress: %s, Topic: %s", cfg.Kafka.BrokerAddress, cfg.Kafka.EmailTopic)
	log.Printf("Loaded from ENV - Notification ServiceURL: %s", cfg.Notification.ServiceURL)

	return nil
}
