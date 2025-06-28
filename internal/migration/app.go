package migration

import (
	"context"
	"flag"

	"github.com/Alias1177/Auth/config"
	"github.com/Alias1177/Auth/db/migrations/manager"
	"github.com/Alias1177/Auth/internal/infrastructure/postgres/connect"
	"github.com/Alias1177/Auth/pkg/appcontext"
	"github.com/Alias1177/Auth/pkg/logger"
)

// App представляет приложение для миграций
type App struct {
	logger *logger.Logger
	config *config.Config
}

// New создает новое приложение для миграций
func New() *App {
	return &App{}
}

// Run запускает приложение миграций
func (m *App) Run() error {
	// Определение флагов командной строки
	var (
		up             = flag.Bool("up", false, "Запустить миграции вверх")
		down           = flag.Bool("down", false, "Откатить последнюю миграцию")
		migrationsPath = flag.String("path", "db/migrations", "Путь к файлам миграций")
	)
	flag.Parse()

	ctx := context.Background()

	// Инициализация логгера
	if err := m.initLogger(); err != nil {
		return err
	}
	defer m.logger.Close()

	// Загрузка конфигурации
	if err := m.loadConfig(); err != nil {
		return err
	}

	// Инициализация базы данных
	if err := m.initDatabase(ctx); err != nil {
		return err
	}
	defer m.closeDatabase()

	// Выполнение миграций
	return m.runMigrations(ctx, *up, *down, *migrationsPath)
}

// initLogger инициализирует логгер
func (m *App) initLogger() error {
	logger, err := logger.New("info")
	if err != nil {
		return err
	}
	m.logger = logger
	return nil
}

// loadConfig загружает конфигурацию
func (m *App) loadConfig() error {
	cfg, err := config.Load(".env")
	if err != nil {
		m.logger.Fatalw("Failed to load config:", "error", err)
		return err
	}
	m.config = cfg
	return nil
}

// initDatabase инициализирует базу данных
func (m *App) initDatabase(ctx context.Context) error {
	// Проверяем, существуют ли уже установленные соединения с БД
	dbContext := appcontext.GetInstance()

	// Если соединения нет, создаем новое
	if dbContext == nil {
		m.logger.Infow("Подключения к БД отсутствуют, устанавливаем новые...")

		// Подключение к PostgreSQL
		postgresDB, err := connect.NewPostgresDB(ctx, m.config.Database.DSN)
		if err != nil {
			m.logger.Fatalw("Failed to connect PostgreSQL:", "error", err)
			return err
		}

		// Устанавливаем глобальный контекст БД (без Redis для миграций)
		appcontext.SetInstance(postgresDB, nil)
	} else {
		m.logger.Infow("Используем существующие подключения к БД")
	}

	return nil
}

// closeDatabase закрывает соединения с базой данных
func (m *App) closeDatabase() {
	dbContext := appcontext.GetInstance()
	if dbContext != nil {
		dbContext.Close()
	}
}

// runMigrations выполняет миграции
func (m *App) runMigrations(ctx context.Context, up, down bool, migrationsPath string) error {
	dbContext := appcontext.GetInstance()

	// Инициализация менеджера миграций
	migrationMgr, err := manager.NewMigrationManager(
		dbContext.PostgresDB.GetConn(),
		m.logger,
		migrationsPath,
	)
	if err != nil {
		m.logger.Fatalw("Failed to create migration manager", "error", err)
		return err
	}
	defer migrationMgr.Close()

	// Выполнение команд миграции
	if up {
		m.logger.Infow("Запуск миграций PostgreSQL...")
		if err := migrationMgr.MigrateUp(ctx); err != nil {
			m.logger.Fatalw("Failed to apply migrations", "error", err)
			return err
		}
		m.logger.Infow("Миграции PostgreSQL успешно применены")
	} else if down {
		m.logger.Infow("Откат миграций PostgreSQL...")
		if err := migrationMgr.MigrateDown(ctx); err != nil {
			m.logger.Fatalw("Failed to rollback migrations", "error", err)
			return err
		}
		m.logger.Infow("Миграции PostgreSQL успешно откачены")
	} else {
		m.logger.Infow("Не указано действие для миграций. Используйте флаги -up или -down.")
	}

	return nil
}
