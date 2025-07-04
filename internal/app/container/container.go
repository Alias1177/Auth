package container

import (
	"context"

	"github.com/Alias1177/Auth/db/migrations/manager"
	"github.com/Alias1177/Auth/internal/config"
	"github.com/Alias1177/Auth/internal/handler/auth"
	"github.com/Alias1177/Auth/internal/handler/user"
	"github.com/Alias1177/Auth/internal/repository"
	"github.com/Alias1177/Auth/internal/repository/postgres"
	"github.com/Alias1177/Auth/internal/repository/redis"
	"github.com/Alias1177/Auth/internal/service"
	"github.com/Alias1177/Auth/pkg/appcontext"
	"github.com/Alias1177/Auth/pkg/database/connect"
	"github.com/Alias1177/Auth/pkg/jwt"
	"github.com/Alias1177/Auth/pkg/kafka"
	"github.com/Alias1177/Auth/pkg/logger"
	"github.com/Alias1177/Auth/pkg/validator"
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
	tokenManager  service.TokenManager
	kafkaProducer *kafka.Producer

	// Handlers
	authHandler          *auth.AuthHandler
	registrationHandler  *auth.RegistrationHandler
	userHandler          *user.UserHandler
	passwordResetHandler *auth.PasswordResetHandler
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

	if err := container.initHandlers(); err != nil {
		return nil, err
	}

	return container, nil
}

// initDatabase инициализирует подключения к базам данных
func (c *Container) initDatabase(ctx context.Context) error {
	// Подключение к PostgreSQL
	postgresDB, err := connect.NewPostgresDB(ctx, c.config.Database.DSN)
	if err != nil {
		c.logger.Fatalw("Failed to connect PostgreSQL:", "error", err)
		return err
	}

	// Подключение к Redis
	redisClient := config.NewRedisClient(c.config.Redis)
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		c.logger.Fatalw("Failed to connect Redis", "error", err)
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
	emailService := service.NewEmailService(c.config, c.logger)
	passwordResetService := service.NewPasswordResetService(
		c.mainRepo,
		c.redisRepo,
		c.logger,
		emailService,
	)

	// Валидатор
	validator := validator.New()

	c.passwordResetHandler = auth.NewPasswordResetHandler(
		passwordResetService,
		emailService,
		validator,
		c.logger,
	)

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
