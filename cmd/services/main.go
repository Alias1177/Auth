package main

import (
	"Auth/config"
	"Auth/internal/delivery/http/registration"
	"Auth/internal/delivery/http/registration/auth"
	"Auth/internal/delivery/http/user"
	"Auth/internal/infrastructure/middleware"
	"Auth/internal/infrastructure/postgres/connect"
	"Auth/internal/repository"
	"Auth/internal/repository/postgres"
	"Auth/internal/repository/redis"
	"Auth/pkg/jwt"
	"Auth/pkg/logger"
	"Auth/pkg/migration"
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"log"
	"net/http"
)

func main() {
	ctx := context.Background()

	// Логгер
	logInstance, err := logger.NewSimpleLogger("info")
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logInstance.Close()

	r := chi.NewRouter()

	loggerMiddleware := middleware.NewLoggerMiddleware(logInstance)
	r.Use(loggerMiddleware.Handler)

	// Настройка CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"}, // TODO Укажите точный домен вашего фронтенда
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Загрузка конфига
	cfg, err := config.Load(".env")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Подключение к PostgreSQL
	postgresDB, err := connect.NewPostgresDB(ctx, cfg.Database.DSN)
	if err != nil {
		logInstance.Fatalw("Failed to connect PostgreSQL", "error", err)
	}
	defer postgresDB.Close()

	// Запуск миграций
	migrator, err := migration.NewMigrator(postgresDB.GetConn(), "db/migrations", logInstance)
	if err != nil {
		logInstance.Fatalw("Failed to create migrator", "error", err)
	}
	defer migrator.Close()

	if err := migrator.Up(); err != nil {
		if errors.Is(err, migration.ErrNoChange) {
			logInstance.Infow("No new migrations, skipping")
		} else {
			logInstance.Fatalw("Failed to run migrations", "error", err)
		}
	} else {
		logInstance.Infow("Database migrations completed successfully")
	}

	// Подключение к Redis
	redisClient := config.NewRedisClient(cfg.Redis)
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		logInstance.Fatalw("Failed to connect Redis", "error", err)
	}
	defer redisClient.Close()

	// Создание репозиториев
	postgresRepo := postgres.NewPostgresRepository(postgresDB.GetConn(), redis.NewRedisRepository(redisClient, logInstance), logInstance)
	redisRepo := redis.NewRedisRepository(redisClient, logInstance)
	mainRepo := repository.NewRepository(postgresRepo, redisRepo, logInstance)

	// JWT Token Manager
	tokenManager := jwt.NewJWTTokenManager(cfg.JWT)

	// Инициализация хэндлеров
	authHandler := auth.NewAuthHandler(tokenManager, cfg.JWT, mainRepo, logInstance)
	registrationHandler := registration.NewRegistrationHandler(mainRepo, tokenManager, cfg.JWT, logInstance)

	userHandler := user.NewUserHandler(mainRepo, logInstance)
	userGet := user.NewUserHandler(mainRepo, logInstance)

	// Маршруты
	r.Post("/login", authHandler.Login)
	r.Post("/register", registrationHandler.Register)

	// Защищённые маршруты
	r.Route("/user", func(r chi.Router) {
		r.Use(middleware.JWTAuthMiddleware(tokenManager))
		r.Put("/{id}", userHandler.UpdateUser)
		r.Get("/me", userGet.GetUserInfoHandler)
	})

	// Запуск сервера
	logInstance.Infow("Starting server", "port", 8080)
	if err := http.ListenAndServe(":8080", r); err != nil {
		logInstance.Fatalw("Server failed", "error", err)
	}
}
