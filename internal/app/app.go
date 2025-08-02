package app

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Alias1177/Auth/internal/app/container"
	"github.com/Alias1177/Auth/internal/config"
	"github.com/Alias1177/Auth/internal/server"
	"github.com/Alias1177/Auth/pkg/logger"
	"github.com/Alias1177/Auth/pkg/sentry"
)

// App представляет основное приложение
type App struct {
	config    *config.Config
	logger    *logger.Logger
	server    *server.Server
	container *container.Container
	ctx       context.Context
	cancel    context.CancelFunc
}

// New создает новый экземпляр приложения
func New() *App {
	ctx, cancel := context.WithCancel(context.Background())
	return &App{
		ctx:    ctx,
		cancel: cancel,
	}
}

// Run запускает приложение
func (a *App) Run() error {
	// Парсинг флагов
	var (
		runMigrations = flag.Bool("migrate", false, "Запустить миграции при старте приложения")
		migrationOnly = flag.Bool("migration-only", false, "Запустить только миграции без сервера")
		migrationUp   = flag.Bool("migration-up", false, "Запустить миграции вверх")
		migrationDown = flag.Bool("migration-down", false, "Откатить миграции")
	)
	flag.Parse()

	// Если нужны только миграции, запускаем migration app
	if *migrationOnly || *migrationUp || *migrationDown {
		return a.runMigrationMode()
	}

	// Инициализация логгера
	if err := a.initLogger(); err != nil {
		return err
	}
	defer a.logger.Close()

	// Загрузка конфигурации
	if err := a.loadConfig(); err != nil {
		return err
	}

	// Инициализация Sentry
	if err := sentry.Init(&a.config.Sentry, a.logger.GetZapLogger()); err != nil {
		a.logger.Errorw("Ошибка инициализации Sentry", "error", err)
		// Не прерываем запуск приложения, если Sentry не удалось инициализировать
	}
	defer sentry.Flush(2 * time.Second)

	// Инициализация контейнера зависимостей
	if err := a.initContainer(); err != nil {
		return err
	}
	defer a.container.Close()

	// Запуск миграций если нужно
	if *runMigrations {
		if err := a.runMigrations(); err != nil {
			return err
		}
	}

	// Инициализация сервера
	if err := a.initServer(); err != nil {
		return err
	}

	// Graceful shutdown
	return a.runWithGracefulShutdown()
}

// runMigrationMode запускает приложение в режиме миграций
func (a *App) runMigrationMode() error {
	migrationApp := NewMigrationApp()
	return migrationApp.Run()
}

// initLogger инициализирует логгер
func (a *App) initLogger() error {
	logger, err := logger.New("info")
	if err != nil {
		return err
	}
	a.logger = logger
	return nil
}

// loadConfig загружает конфигурацию
func (a *App) loadConfig() error {
	// Пытаемся загрузить из .env файла (для локальной разработки и Docker)
	cfg, err := config.Load(".env")
	if err != nil {
		a.logger.Warnw("Failed to load .env file, using environment variables only:", "error", err)
		// Если файл не найден, загружаем только из переменных окружения
		cfg = &config.Config{}
		err = config.LoadFromEnv(cfg)
		if err != nil {
			a.logger.Fatalw("Failed to load config from environment:", "error", err)
			return err
		}
	} else {
		a.logger.Infow("Successfully loaded config from .env file")
	}
	a.config = cfg
	return nil
}

// initContainer инициализирует контейнер зависимостей
func (a *App) initContainer() error {
	container, err := container.New(a.ctx, a.config, a.logger)
	if err != nil {
		return err
	}
	a.container = container
	return nil
}

// runMigrations запускает миграции
func (a *App) runMigrations() error {
	a.logger.Infow("Запуск миграций...")
	return a.container.RunMigrations(a.ctx)
}

// initServer инициализирует HTTP сервер
func (a *App) initServer() error {
	server, err := server.New(a.container)
	if err != nil {
		return err
	}
	a.server = server
	return nil
}

// runWithGracefulShutdown запускает сервер с graceful shutdown
func (a *App) runWithGracefulShutdown() error {
	// Проверяем доступность порта перед запуском
	if err := a.checkPortAvailability(":8080"); err != nil {
		return fmt.Errorf("port 8080 is not available: %w", err)
	}

	// Канал для получения сигналов
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Запуск сервера в горутине
	serverErr := make(chan error, 1)
	go func() {
		a.logger.Infow("Starting server", "port", 8080)
		serverErr <- a.server.Start(":8080")
	}()

	// Ожидание сигнала или ошибки
	select {
	case err := <-serverErr:
		return err
	case sig := <-quit:
		a.logger.Infow("Received shutdown signal", "signal", sig)

		// Graceful shutdown с таймаутом
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		return a.server.Shutdown(shutdownCtx)
	}
}

// checkPortAvailability проверяет доступность порта
func (a *App) checkPortAvailability(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	ln.Close()
	return nil
}
