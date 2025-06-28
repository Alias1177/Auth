package main

import (
	"context"
	"flag"
	"log"

	"github.com/Alias1177/Auth/config"
	"github.com/Alias1177/Auth/db/migrations/manager"
	"github.com/Alias1177/Auth/internal/infrastructure/postgres/connect"
	"github.com/Alias1177/Auth/pkg/appcontext"
	"github.com/Alias1177/Auth/pkg/logger"
)

func main() {
	// Определение флагов командной строки
	var (
		up             = flag.Bool("up", false, "Запустить миграции вверх")
		down           = flag.Bool("down", false, "Откатить последнюю миграцию")
		migrationsPath = flag.String("path", "db/migrations", "Путь к файлам миграций")
	)
	flag.Parse()

	ctx := context.Background()

	// Инициализация логгера
	logInstance, err := logger.New("info")
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logInstance.Close()

	// Проверяем, существуют ли уже установленные соединения с БД
	dbContext := appcontext.GetInstance()

	// Если соединения нет, создаем новое
	if dbContext == nil {
		logInstance.Infow("Подключения к БД отсутствуют, устанавливаем новые...")

		cfg, err := config.Load(".env")
		if err != nil {
			logInstance.Fatalw("Failed to load configs:", "error", err)
		}

		// Подключение к PostgreSQL
		postgresDB, err := connect.NewPostgresDB(ctx, cfg.Database.DSN)
		if err != nil {
			logInstance.Fatalw("Failed to connect PostgreSQL:", "error", err)
		}

		// Устанавливаем глобальный контекст БД
		appcontext.SetInstance(postgresDB, nil)
		dbContext = appcontext.GetInstance()
	} else {
		logInstance.Infow("Используем существующие подключения к БД")
	}

	// Настраиваем отложенное закрытие соединений, только если мы их создали в этом процессе
	if dbContext != nil && appcontext.GetInstance() == dbContext {
		defer dbContext.Close()
	}

	// Инициализация менеджера миграций
	migrationMgr, err := manager.NewMigrationManager(
		dbContext.PostgresDB.GetConn(),
		logInstance,
		*migrationsPath,
	)
	if err != nil {
		logInstance.Fatalw("Failed to create migration manager", "error", err)
	}
	defer migrationMgr.Close()

	// Выполнение команд миграции
	if *up {
		logInstance.Infow("Запуск миграций PostgreSQL...")
		if err := migrationMgr.MigrateUp(ctx); err != nil {
			logInstance.Fatalw("Failed to apply migrations", "error", err)
		}
		logInstance.Infow("Миграции PostgreSQL успешно применены")
	} else if *down {
		logInstance.Infow("Откат миграций PostgreSQL...")
		if err := migrationMgr.MigrateDown(ctx); err != nil {
			logInstance.Fatalw("Failed to rollback migrations", "error", err)
		}
		logInstance.Infow("Миграции PostgreSQL успешно откачены")
	} else {
		logInstance.Infow("Не указано действие для миграций. Используйте флаги -up или -down.")
	}
}
