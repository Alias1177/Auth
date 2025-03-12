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
	"Auth/pkg/jwt"
	"Auth/pkg/logger"
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"strconv"
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
	tokenManager := jwt.NewJWTTokenManager(cfg.JWT)

	// Handlers initialization
	authHandler := auth.NewAuthHandler(tokenManager, cfg.JWT, mainRepo)
	registrationHandler := registration.NewRegistrationHandler(mainRepo, tokenManager, cfg.JWT)

	r := chi.NewRouter()

	// Авторизация и регистрация
	r.Post("/login", authHandler.Login)
	r.Post("/register", registrationHandler.Register)

	// Защищённые роуты (JWT middleware)
	r.Route("/user", func(r chi.Router) {
		r.Use(middleware.JWTAuthMiddleware(tokenManager))

		r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
			// Получаем информацию о пользователе из контекста (добавленную middleware)
			userClaims, ok := r.Context().Value(middleware.CtxUserKey).(*entity.UserClaims)
			if !ok {
				http.Error(w, "Ошибка получения информации о пользователе", http.StatusInternalServerError)
				return
			}

			// Преобразуем ID из строки в int
			userID, err := strconv.Atoi(userClaims.UserID)
			if err != nil {
				http.Error(w, "Некорректный ID пользователя", http.StatusBadRequest)
				return
			}

			// Получаем полную информацию о пользователе
			user, err := mainRepo.GetUser(r.Context(), userID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Отправляем информацию о пользователе
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
