package container

import (
	"context"
	"time"

	"github.com/Alias1177/Auth/db/migrations/manager"
	"github.com/Alias1177/Auth/internal/config"
	"github.com/Alias1177/Auth/internal/handler/auth"
	"github.com/Alias1177/Auth/internal/handler/user"
	"github.com/Alias1177/Auth/internal/oauth"
	"github.com/Alias1177/Auth/internal/repository"
	"github.com/Alias1177/Auth/internal/repository/postgres"
	"github.com/Alias1177/Auth/internal/repository/redis"
	"github.com/Alias1177/Auth/internal/service"
	"github.com/Alias1177/Auth/pkg/appcontext"
	"github.com/Alias1177/Auth/pkg/database/connect"
	"github.com/Alias1177/Auth/pkg/jwt"
	"github.com/Alias1177/Auth/pkg/kafka"
	"github.com/Alias1177/Auth/pkg/logger"
	"github.com/Alias1177/Auth/pkg/notification"
	"github.com/Alias1177/Auth/pkg/validator"
	redisClient "github.com/redis/go-redis/v9"
)

// Container содержит все зависимости приложения
type Container struct {
	config *config.Config
	logger *logger.Logger

	// Repositories
	postgresRepo *postgres.PostgresRepository
	redisRepo    *redis.RedisRepository
	mainRepo     *repository.Repository

	// Services
	tokenManager       service.TokenManager
	kafkaProducer      *kafka.Producer
	notificationClient *notification.NotificationClient

	// Handlers
	authHandler          *auth.AuthHandler
	registrationHandler  *auth.RegistrationHandler
	userHandler          *user.UserHandler
	passwordResetHandler *auth.PasswordResetHandler
	oauthHandler         *auth.OAuthHandler
}

// New создает новый контейнер зависимостей
func New(ctx context.Context, cfg *config.Config, log *logger.Logger) (*Container, error) {
	container := &Container{
		config: cfg,
		logger: log,
	}

	// Инициализация в правильном порядке
	if err := container.initDatabase(ctx); err != nil {
		return nil, err
	}

	if err := container.initRepositories(); err != nil {
		return nil, err
	}

	if err := container.initServices(); err != nil {
		return nil, err
	}

	// Инициализация OAuth
	oauth.NewOAuth(cfg)

	if err := container.initHandlers(); err != nil {
		return nil, err
	}

	return container, nil
}

// initDatabase инициализирует подключения к базам данных
func (c *Container) initDatabase(ctx context.Context) error {
	// Подключение к PostgreSQL с retry логикой
	var postgresDB *connect.PostgresDB
	var err error

	maxRetries := 10
	retryDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		c.logger.Infow("Attempting to connect to PostgreSQL", "attempt", i+1, "max_attempts", maxRetries)

		postgresDB, err = connect.NewPostgresDB(ctx, c.config.Database.DSN)
		if err == nil {
			c.logger.Infow("Successfully connected to PostgreSQL")
			break
		}

		c.logger.Warnw("Failed to connect to PostgreSQL, retrying",
			"attempt", i+1, "error", err, "next_retry_in", retryDelay)

		if i < maxRetries-1 {
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
		}
	}

	if err != nil {
		c.logger.Fatalw("Failed to connect PostgreSQL after all retries:", "error", err)
		return err
	}

	// Подключение к Redis с retry логикой
	var redisClient *redisClient.Client

	for i := 0; i < maxRetries; i++ {
		c.logger.Infow("Attempting to connect to Redis", "attempt", i+1, "max_attempts", maxRetries)

		redisClient = config.NewRedisClient(c.config.Redis)
		if _, err := redisClient.Ping(ctx).Result(); err == nil {
			c.logger.Infow("Successfully connected to Redis")
			break
		}

		c.logger.Warnw("Failed to connect to Redis, retrying",
			"attempt", i+1, "error", err, "next_retry_in", retryDelay)

		if i < maxRetries-1 {
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
		}
	}

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		c.logger.Fatalw("Failed to connect Redis after all retries:", "error", err)
		postgresDB.Close()
		return err
	}

	// Устанавливаем глобальный контекст БД
	appcontext.SetInstance(postgresDB, redisClient)

	return nil
}

// initRepositories инициализирует репозитории
func (c *Container) initRepositories() error {
	dbContext := appcontext.GetInstance()

	c.redisRepo = redis.NewRedisRepository(dbContext.RedisClient, c.logger)
	c.postgresRepo = postgres.NewPostgresRepository(
		dbContext.PostgresDB.GetConn(),
		c.redisRepo,
		c.logger,
	)
	c.mainRepo = repository.NewRepository(c.postgresRepo, c.redisRepo, c.logger)

	return nil
}

// initServices инициализирует сервисы
func (c *Container) initServices() error {
	// JWT Token Manager
	c.tokenManager = jwt.NewJWTTokenManager(c.config.JWT)

	// Kafka Producer
	c.kafkaProducer = kafka.NewProducer(
		c.config.Kafka.BrokerAddress,
		c.config.Kafka.EmailTopic,
		c.logger,
	)

	// Notification Client
	c.notificationClient = notification.NewNotificationClient(c.config.Notification.ServiceURL)

	return nil
}

// initHandlers инициализирует HTTP хендлеры
func (c *Container) initHandlers() error {
	c.authHandler = auth.NewAuthHandler(
		c.tokenManager,
		c.config.JWT,
		c.mainRepo,
		c.logger,
	)

	c.registrationHandler = auth.NewRegistrationHandler(
		c.mainRepo,
		c.tokenManager,
		c.config.JWT,
		c.logger,
		c.kafkaProducer,
	)

	c.userHandler = user.NewUserHandler(c.mainRepo, c.logger)

	// Инициализация сервиса сброса пароля
	passwordResetService := service.NewPasswordResetService(
		c.mainRepo,
		c.redisRepo,
		c.logger,
		c.kafkaProducer,
		c.notificationClient,
	)

	// Валидатор
	validator := validator.New()

	c.passwordResetHandler = auth.NewPasswordResetHandler(
		passwordResetService,
		validator,
		c.logger,
	)

	// Инициализация OAuth handler
	c.oauthHandler = auth.NewOAuthService(c.logger, c.tokenManager, c.mainRepo)

	return nil
}

// RunMigrations запускает миграции базы данных
func (c *Container) RunMigrations(ctx context.Context) error {
	dbContext := appcontext.GetInstance()

	migrationMgr, err := manager.NewMigrationManager(
		dbContext.PostgresDB.GetConn(),
		c.logger,
		"db/migrations",
	)
	if err != nil {
		c.logger.Fatalw("Не удалось создать менеджер миграций", "error", err)
		return err
	}
	defer migrationMgr.Close()

	if err := migrationMgr.MigrateUp(ctx); err != nil {
		c.logger.Fatalw("Ошибка при применении миграций", "error", err)
		return err
	}

	c.logger.Infow("Миграции успешно применены")
	return nil
}

// Getters для доступа к зависимостям
func (c *Container) GetAuthHandler() *auth.AuthHandler {
	return c.authHandler
}

func (c *Container) GetRegistrationHandler() *auth.RegistrationHandler {
	return c.registrationHandler
}

func (c *Container) GetUserHandler() *user.UserHandler {
	return c.userHandler
}

func (c *Container) GetPasswordResetHandler() *auth.PasswordResetHandler {
	return c.passwordResetHandler
}

func (c *Container) GetTokenManager() service.TokenManager {
	return c.tokenManager
}

func (c *Container) GetLogger() *logger.Logger {
	return c.logger
}

func (c *Container) GetOAuthHandler() *auth.OAuthHandler {
	return c.oauthHandler
}

// Close закрывает все соединения
func (c *Container) Close() {
	if c.kafkaProducer != nil {
		c.kafkaProducer.Close()
	}

	dbContext := appcontext.GetInstance()
	if dbContext != nil {
		dbContext.Close()
	}
}
