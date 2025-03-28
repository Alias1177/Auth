package main

import (
	"Auth/config"
	"Auth/db/migrations/manager"
	"Auth/internal/delivery/http/registration"
	"Auth/internal/delivery/http/registration/auth"
	"Auth/internal/delivery/http/user"
	"Auth/internal/infrastructure/middleware"
	"Auth/internal/infrastructure/postgres/connect"
	"Auth/internal/repository"
	"Auth/internal/repository/postgres"
	"Auth/internal/repository/redis"
	"Auth/pkg/appcontext"
	"Auth/pkg/jwt"
	"Auth/pkg/kafka"
	"Auth/pkg/logger"
	"context"
	"flag"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

func main() {
	var runMigrations = flag.Bool("migrate", false, "Запустить миграции при старте приложения")
	flag.Parse()

	ctx := context.Background()

	logInstance, err := logger.NewSimpleLogger("info")
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}

	defer logInstance.Close()

	r := chi.NewRouter()

	loggerMiddleware := middleware.NewLoggerMiddleware(logInstance)
	metrics := middleware.NewMetricsMiddleware("auth_service")

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{}, // Пустой список (разрешим динамически)
		AllowOriginFunc: func(r *http.Request, origin string) bool {
			return true
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	cfg, err := config.Load(".env")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Подключение к PostgreSQL
	postgresDB, err := connect.NewPostgresDB(ctx, cfg.Database.DSN)
	if err != nil {
		logInstance.Fatalw("Failed to connect PostgreSQL:", "error", err)
	}

	// Подключение к Redis
	redisClient := config.NewRedisClient(cfg.Redis)
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		logInstance.Fatalw("Failed to connect Redis", "error", err)
		postgresDB.Close() // Закрываем PostgreSQL, если Redis не доступен
		return
	}

	// Устанавливаем глобальный контекст БД
	appcontext.SetInstance(postgresDB, redisClient)

	// Получаем контекст БД и настраиваем отложенное закрытие соединений
	dbContext := appcontext.GetInstance()
	defer dbContext.Close()

	// Инициализация Kafka Producer
	kafkaProducer := kafka.NewProducer(cfg.Kafka.BrokerAddress, cfg.Kafka.EmailTopic, logInstance)
	defer kafkaProducer.Close()

	// Запуск миграций если указан флаг
	if *runMigrations {
		logInstance.Infow("Запуск миграций...")

		// Создаем менеджер миграций
		migrationMgr, err := manager.NewMigrationManager(
			dbContext.PostgresDB.GetConn(),
			dbContext.RedisClient,
			logInstance,
			"db/migrations",
		)
		if err != nil {
			logInstance.Fatalw("Не удалось создать менеджер миграций", "error", err)
		}
		defer migrationMgr.Close()

		// Запускаем миграции
		if err := migrationMgr.MigrateUp(ctx); err != nil {
			logInstance.Fatalw("Ошибка при применении миграций", "error", err)
		}

		logInstance.Infow("Миграции успешно применены")
	}

	postgresRepo := postgres.NewPostgresRepository(
		dbContext.PostgresDB.GetConn(),
		redis.NewRedisRepository(dbContext.RedisClient, logInstance),
		logInstance,
	)
	redisRepo := redis.NewRedisRepository(dbContext.RedisClient, logInstance)
	mainRepo := repository.NewRepository(postgresRepo, redisRepo, logInstance)

	// JWT Token Manager
	tokenManager := jwt.NewJWTTokenManager(cfg.JWT)

	// Инициализация хэндлеров
	authHandler := auth.NewAuthHandler(tokenManager, cfg.JWT, mainRepo, logInstance)
	registrationHandler := registration.NewRegistrationHandler(mainRepo, tokenManager, cfg.JWT, logInstance, kafkaProducer)
	userHandler := user.NewUserHandler(mainRepo, logInstance)
	userGet := user.NewUserHandler(mainRepo, logInstance)

	r.Use(loggerMiddleware.Handler)
	r.Use(metrics.Middleware)

	// Маршруты
	r.Post("/login", authHandler.Login)
	r.Post("/register", registrationHandler.Register)
	r.Handle("/metrics", promhttp.Handler())
	// Добавьте эту строку в вашем main.go в разделе регистрации маршрутов
	r.Post("/reset-password", user.ResetPasswordHandler(mainRepo, logInstance))

	// Защищённые маршруты
	r.Route("/user", func(r chi.Router) {
		r.Use(middleware.JWTAuthMiddleware(tokenManager))
		r.Patch("/{id}", userHandler.UpdateUser)
		r.Get("/me", userGet.GetUserInfoHandler)
	})

	logInstance.Infow("Starting server", "port", 8080)
	if err := http.ListenAndServe(":8080", r); err != nil {
		logInstance.Fatalw("Server failed", "error", err)
	}
}
