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
	"Auth/pkg/jwt"
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
	// Флаг для запуска миграций при старте приложения
	var (
		runMigrations        = flag.Bool("migrate", false, "Запустить все миграции при старте приложения")
		runPostgresMigration = flag.Bool("migrate-pg", false, "Запустить только миграции PostgreSQL")
		runRedisMigration    = flag.Bool("migrate-redis", false, "Запустить только миграции Redis")
		migrationsPath       = flag.String("migrations-path", config.DefaultMigrationsPath, "Путь к файлам миграций")
	)
	flag.Parse()

	ctx := context.Background()

	logInstance, err := logger.NewSimpleLogger("info")
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}

	defer logInstance.Close()

	r := chi.NewRouter()

	loggerMiddleware := middleware.NewLoggerMiddleware(logInstance)

	metrics := middleware.NewMetricsMiddleware(config.ServiceName)

	// Настройки CORS
	corsOptions := cors.Options{

		AllowedOrigins: []string{}, // Пустой список (разрешим динамически)
		AllowOriginFunc: func(r *http.Request, origin string) bool {
			return true
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           300,
	}


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

	// Подключение к Redis
	redisClient := config.NewRedisClient(cfg.Redis)
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		logInstance.Fatalw("Failed to connect Redis", "error", err)
	}
	defer redisClient.Close()


	// Создаем менеджер миграций независимо от флагов,
	// чтобы не дублировать код и иметь единую точку управления миграциями
	migrationMgr, err := manager.NewMigrationManager(
		postgresDB.GetConn(),
		redisClient,
		logInstance,
		*migrationsPath,
	)
	if err != nil {
		logInstance.Fatalw("Не удалось создать менеджер миграций", "error", err)
	}
	defer migrationMgr.Close()

	// Обработка различных вариантов запуска миграций
	switch {
	case *runMigrations:
		logInstance.Infow("Запуск всех миграций...")
		if err := migrationMgr.MigrateUp(ctx); err != nil {
			logInstance.Fatalw("Ошибка при применении миграций", "error", err)
		}
		logInstance.Infow("Миграции успешно применены")

	case *runPostgresMigration:
		logInstance.Infow("Запуск миграций PostgreSQL...")
		if err := migrationMgr.MigratePostgresUp(); err != nil {
			logInstance.Fatalw("Ошибка при применении миграций PostgreSQL", "error", err)
		}
		logInstance.Infow("Миграции PostgreSQL успешно применены")

	case *runRedisMigration:
		logInstance.Infow("Запуск миграций Redis...")
		if err := migrationMgr.MigrateRedisUp(ctx); err != nil {
			logInstance.Fatalw("Ошибка при применении миграций Redis", "error", err)
		}
		logInstance.Infow("Миграции Redis успешно применены")

	}

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

	// Базовые middleware
	r.Use(cors.Handler(corsOptions))
	r.Use(loggerMiddleware.Handler)
	r.Use(metrics.Middleware) // Применяем метрики ко всем запросам

	// Эндпоинт для метрик - всегда используем PathMiddleware для правильного определения маршрута
	r.With(middleware.PathMiddleware("/metrics")).
		Handle("/metrics", promhttp.Handler())

	// Аутентификация и регистрация - с явным указанием путей
	r.With(middleware.PathMiddleware("/login")).
		Post("/login", authHandler.Login)

	r.With(middleware.PathMiddleware("/register")).
		Post("/register", registrationHandler.Register)

	// Защищённые маршруты пользователя
	r.Route("/user", func(r chi.Router) {
		r.Use(middleware.JWTAuthMiddleware(tokenManager))

		// Явно устанавливаем пути для каждого маршрута
		r.With(middleware.PathMiddleware("/user/{id}")).
			Put("/{id}", userHandler.UpdateUser)

		r.With(middleware.PathMiddleware("/user/me")).
			Get("/me", userHandler.GetUserInfoHandler)
	})

	// Запуск сервера
	logInstance.Infow("Starting server", "port", 8080)
	if err := http.ListenAndServe(":8080", r); err != nil {
		logInstance.Fatalw("Server failed", "error", err)
	}
}
