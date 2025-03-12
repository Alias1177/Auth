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
	"context"
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
		AllowedHeaders:   []string{"Content-Type", "Authorization"}, // Добавлен заголовок Authorization
		AllowCredentials: true,                                      // Разрешаем отправку куки
		MaxAge:           300,                                       // Максимальное время кэширования CORS заголовков в браузере
	}))

	cfg, err := config.Load(".env")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

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
	postgresRepo := postgres.NewPostgresRepository(postgresDB.GetConn(), redis.NewRedisRepository(redisClient))
	redisRepo := redis.NewRedisRepository(redisClient)
	mainRepo := repository.NewRepository(postgresRepo, redisRepo, logInstance)

	// JWT Token Manager
	tokenManager := jwt.NewJWTTokenManager(cfg.JWT)

	// Handlers initialization
	authHandler := auth.NewAuthHandler(tokenManager, cfg.JWT, mainRepo)
	registrationHandler := registration.NewRegistrationHandler(mainRepo, tokenManager, cfg.JWT)
	userHandler := user.NewUserHandler(mainRepo)
	userGet := user.NewUserHandler(mainRepo)

	// Авторизация и регистрация
	r.Post("/login", authHandler.Login)
	r.Post("/register", registrationHandler.Register)

	// Защищённые роуты (JWT middleware)
	r.Route("/user", func(r chi.Router) {
		r.Use(middleware.JWTAuthMiddleware(tokenManager))

		r.Put("/{id}", userHandler.UpdateUser)
		r.Get("/me", userGet.GetUserInfoHandler)
	})

	// Запуск сервера
	//TODO graceful shutdown
	logInstance.Infow("Starting server", "port", 8080)
	if err := http.ListenAndServe(":8080", r); err != nil {
		logInstance.Fatalw("Server failed", "error", err)
	}
}
