// Example integration in cmd/service/main.go
package main

import (
	"Auth/config"
	"Auth/internal/delivery/http/registration"
	"Auth/internal/delivery/http/registration/auth"
	"Auth/internal/delivery/http/user"
	"Auth/internal/infrastructure/middleware"
	"Auth/internal/infrastructure/middleware/ratelimiter"
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
	"time"
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

	// Create rate limiters with different limits for different endpoints
	authRateLimiter := ratelimiter.NewRateLimiter(30, time.Minute, logInstance) // 30 reqs/min for auth endpoints
	apiRateLimiter := ratelimiter.NewRateLimiter(300, time.Minute, logInstance) // 300 reqs/min for general API

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

	// Apply common middleware
	r.Use(loggerMiddleware.Handler)
	r.Use(metrics.Middleware)

	cfg, err := config.Load(".env")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Database connections
	postgresDB, err := connect.NewPostgresDB(ctx, cfg.Database.DSN)
	if err != nil {
		logInstance.Fatalw("Failed to connect PostgreSQL:", "error", err)
	}

	redisClient := config.NewRedisClient(cfg.Redis)
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		logInstance.Fatalw("Failed to connect Redis", "error", err)
		postgresDB.Close()
		return
	}

	appcontext.SetInstance(postgresDB, redisClient)
	dbContext := appcontext.GetInstance()
	defer dbContext.Close()

	// Kafka Producer
	kafkaProducer := kafka.NewProducer(cfg.Kafka.BrokerAddress, cfg.Kafka.EmailTopic, logInstance)
	defer kafkaProducer.Close()

	// Migrations if needed
	if *runMigrations {
		// ... migration code (unchanged)
	}

	// Repositories
	postgresRepo := postgres.NewPostgresRepository(
		dbContext.PostgresDB.GetConn(),
		redis.NewRedisRepository(dbContext.RedisClient, logInstance),
		logInstance,
	)
	redisRepo := redis.NewRedisRepository(dbContext.RedisClient, logInstance)
	mainRepo := repository.NewRepository(postgresRepo, redisRepo, logInstance)

	// JWT Token Manager
	tokenManager := jwt.NewJWTTokenManager(cfg.JWT)

	// Handlers
	authHandler := auth.NewAuthHandler(tokenManager, cfg.JWT, mainRepo, logInstance)
	registrationHandler := registration.NewRegistrationHandler(mainRepo, tokenManager, cfg.JWT, logInstance, kafkaProducer)
	userHandler := user.NewUserHandler(mainRepo, logInstance)
	userGet := user.NewUserHandler(mainRepo, logInstance)

	// Public routes with authentication rate limiter
	r.Group(func(r chi.Router) {
		r.Use(authRateLimiter.Middleware) // Apply rate limiting to auth endpoints
		r.Post("/login", authHandler.Login)
		r.Post("/register", registrationHandler.Register)
		r.Post("/reset-password", user.ResetPasswordHandler(mainRepo, logInstance))
	})

	// Metrics endpoint
	r.Handle("/metrics", promhttp.Handler())

	// Protected routes with API rate limiter
	r.Route("/user", func(r chi.Router) {
		r.Use(middleware.JWTAuthMiddleware(tokenManager))
		r.Use(apiRateLimiter.Middleware) // Apply rate limiting to API endpoints
		r.Patch("/{id}", userHandler.UpdateUser)
		r.Get("/me", userGet.GetUserInfoHandler)
	})

	logInstance.Infow("Starting server", "port", 8080)
	if err := http.ListenAndServe(":8080", r); err != nil {
		logInstance.Fatalw("Server failed", "error", err)
	}
}
