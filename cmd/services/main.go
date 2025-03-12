package main

import (
	"Auth/config"
	"Auth/internal/entity"
	"Auth/internal/infrastructure/postgres/connect"
	"Auth/internal/repository"
	"Auth/internal/repository/postgres"
	"Auth/internal/repository/redis"
	"Auth/pkg/logger"
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func main() {
	ctx := context.Background()

	// Загружаем конфигурацию
	cfg, err := config.Load(".env")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Инициализируем логгер
	logInstance, err := logger.NewSimpleLogger("info")
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logInstance.Close()

	// Подключаемся к PostgreSQL используя конфигурацию
	postgresDB, err := connect.NewPostgresDB(ctx, cfg.Database.DSN)
	if err != nil {
		logInstance.Fatalw("Failed to connect to PostgreSQL", "error", err)
	}
	defer postgresDB.Close()

	// Инициализируем Redis используя конфигурацию
	redisClient := config.NewRedisClient(cfg.Redis)
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		logInstance.Fatalw("Failed to connect to Redis", "error", err)
	}
	defer redisClient.Close()

	// Создаем репозитории
	postgresRepo := postgres.NewPostgresRepository(postgresDB.GetConn())
	redisRepo := redis.NewRedisRepository(redisClient)
	mainRepo := repository.NewRepository(postgresRepo, redisRepo, logInstance)

	// Настраиваем маршрутизатор
	r := chi.NewRouter()

	//Todo заглушка
	r.Post("/user", func(w http.ResponseWriter, r *http.Request) {
		user := &entity.User{
			UserName: "testuser1",
			Email:    "test@exampl1e.com",
			Password: "passwor1d123",
		} //Todo заглушка

		err := mainRepo.CreateUser(r.Context(), user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	})

	//TODO сделать graceful shutdown
	// Запускаем сервер
	logInstance.Infow("Starting server", "port", 8080)
	if err := http.ListenAndServe(":8080", r); err != nil {
		logInstance.Fatalw("Server failed", "error", err)
	}
}
