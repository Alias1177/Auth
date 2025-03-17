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
	postgres2 "Auth/pkg/migration/postgres"
	redis_migration "Auth/pkg/migration/redis"
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
	//metrics := middleware.NewMetricsMiddleware("auth_service")
	// Настройка CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{}, // Пустой список (разрешим динамически)
		AllowOriginFunc: func(r *http.Request, origin string) bool {
			return true
		},
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

	// Запуск миграций Postgres
	migrator, err := postgres2.NewMigrator(postgresDB.GetConn(), "db/migrations/postgres", logInstance)
	if err != nil {
		logInstance.Fatalw("Failed to create migrator", "error", err)
	}
	defer migrator.Close()

	if err := migrator.Up(); err != nil {
		if errors.Is(err, migration.ErrNoChange) {
			logInstance.Infow("No new PostgreSQL migrations, skipping")
		} else {
			logInstance.Fatalw("Failed to run PostgreSQL migrations", "error", err)
		}
	} else {
		logInstance.Infow("PostgreSQL migrations completed successfully")
	}

	//Откат миграции (по необходимости)
	//if err := migrator.Down(); err != nil {
	//	logInstance.Errorf("Rollback failed: %v", err)
	//}

	// Подключение к Redis
	redisClient := config.NewRedisClient(cfg.Redis)
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		logInstance.Fatalw("Failed to connect Redis", "error", err)
	}
	defer redisClient.Close()

	// Запуск миграции Redis
	redisMigrator := redis_migration.NewRedisMigrator(redisClient, logInstance)
	if err := redisMigrator.Up(ctx); err != nil {
		logInstance.Fatalw("Failed to run Redis migrations", "error", err)
	} else {
		logInstance.Infow("Redis migrations completed successfully")
	}

	// Откат миграции (по необходимости)
	// if err := migrator.Down(ctx); err != nil {
	//     logInstance.Fatalw("Failed to rollback Redis migrations", "error", err)
	// }

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

	//middlewares
	r.Use(loggerMiddleware.Handler)
	//r.Use(metrics.Middleware)

	// Маршруты
	r.Post("/login", authHandler.Login)
	r.Post("/register", registrationHandler.Register)
	//r.Handle("/metrics", promhttp.Handler())

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
