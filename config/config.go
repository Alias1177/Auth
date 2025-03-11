package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
)

type DatabaseConfig struct {
	DSN string `env:"DATABASE_DSN" env-required:"true"`
}

type JWTConfig struct {
	Secret string `env:"JWT_SECRET" env-required:"true"`
}

type Config struct {
	Database DatabaseConfig
	JWT      JWTConfig
}

func Load(path string) Config {
	var cfg Config

	err := cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		log.Fatalf("Unable to read config: %v", err)
	}

	err = cleanenv.ReadEnv(&cfg)
	if err != nil {

		log.Fatalf("Unable to read environment variables: %v", err)
	}

	return cfg
}
