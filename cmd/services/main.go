package main

import (
	"Auth/config"
	"Auth/internal/delivery/http/registration"
	"Auth/internal/delivery/http/registration/auth"
	"Auth/internal/entity"
	"Auth/internal/infrastructure/middleware"
	"Auth/internal/infrastructure/postgres/connect"
	"Auth/internal/repository"
	"Auth/internal/repository/postgres"
	"Auth/internal/repository/redis"
	"Auth/internal/usecase"
	"Auth/pkg/logger"
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load(".env")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Логгер
	logInstance, err := logger.NewSimpleLogger("info")
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logInstance.Close()

	// PostgreSQL Connection
	postgresDB, err := connect.NewPostgresDB(ctx, cfg.Database.DSN)
	if err != nil {
		logInstance.Fatalw("Failed to connect PostgreSQL", "error", err)
	}
	defer postgresDB.Close()

	// Redis Connection
	redisClient := config.NewRedisClient(cfg.Redis)
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		logInstance.Fatalw("Failed to connect Redis", "error", err)
	}
	defer redisClient.Close()

	// Repository setup
	postgresRepo := postgres.NewPostgresRepository(postgresDB.GetConn())
	redisRepo := redis.NewRedisRepository(redisClient)
	mainRepo := repository.NewRepository(postgresRepo, redisRepo, logInstance)

	// JWT Token Manager
	tokenManager := usecase.NewJWTTokenManager(cfg.JWT)

	// Handlers initialization
	authHandler := auth.NewAuthHandler(tokenManager, cfg.JWT, mainRepo)
	registrationHandler := registration.NewRegistrationHandler(mainRepo, tokenManager, cfg.JWT)

	r := chi.NewRouter()

	// Авторизация и регистрация
	r.Post("/login", authHandler.Login)
	r.Post("/refresh-token", authHandler.RefreshToken)
	r.Post("/register", registrationHandler.Register)

	// Защищённые роуты (JWT middleware)
	r.Route("/user", func(r chi.Router) {
		r.Use(middleware.JWTAuthMiddleware(tokenManager))

		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			var user entity.User
			if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
				http.Error(w, "Некорректный запрос", http.StatusBadRequest)
				return
			}

			if err := mainRepo.CreateUser(r.Context(), &user); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(user)
		})
	})

	// Запуск сервера
	logInstance.Infow("Starting server", "port", 8080)
	if err := http.ListenAndServe(":8080", r); err != nil {
		logInstance.Fatalw("Server failed", "error", err)
	}
}
