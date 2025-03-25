package main

import (
	"Auth/config"
	"Auth/db/migrations/manager"
	"Auth/internal/infrastructure/postgres/connect"
	"Auth/pkg/appcontext"
	"Auth/pkg/logger"
	"context"
	"flag"
	"log"
)

func main() {
	// Определение флагов командной строки
	var (
		up             = flag.Bool("up", false, "Запустить миграции вверх")
		down           = flag.Bool("down", false, "Откатить последнюю миграцию")
		postgres       = flag.Bool("postgres", false, "Применить только PostgreSQL миграции")
		redisFlag      = flag.Bool("redis", false, "Применить только Redis миграции")
		migrationsPath = flag.String("path", "db/migrations", "Путь к файлам миграций")
	)
	flag.Parse()

	ctx := context.Background()

	// Инициализация логгера
	logInstance, err := logger.NewSimpleLogger("info")
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
			logInstance.Fatalw("Failed to load config:", "error", err)
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
			postgresDB.Close()
			return
		}

		// Устанавливаем глобальный контекст БД
		appcontext.SetInstance(postgresDB, redisClient)
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
		dbContext.RedisClient,
		logInstance,
		*migrationsPath,
	)
	if err != nil {
		logInstance.Fatalw("Failed to create migration manager", "error", err)
	}
	defer migrationMgr.Close()

	// Выполнение команд миграции
	if *up {
		if *postgres {
			// Только PostgreSQL миграции вверх
			logInstance.Infow("Запуск PostgreSQL миграций...")
			if err := migrationMgr.MigratePostgresUp(); err != nil {
				logInstance.Fatalw("Failed to apply PostgreSQL migrations", "error", err)
			}
			logInstance.Infow("PostgreSQL миграции успешно применены")
		} else if *redisFlag {
			// Только Redis миграции вверх
			logInstance.Infow("Запуск Redis миграций...")
			if err := migrationMgr.MigrateRedisUp(ctx); err != nil {
				logInstance.Fatalw("Failed to apply Redis migrations", "error", err)
			}
			logInstance.Infow("Redis миграции успешно применены")
		} else {
			// Все миграции вверх
			logInstance.Infow("Запуск всех миграций...")
			if err := migrationMgr.MigrateUp(ctx); err != nil {
				logInstance.Fatalw("Failed to apply migrations", "error", err)
			}
			logInstance.Infow("Все миграции успешно применены")
		}
	} else if *down {
		if *postgres {

			logInstance.Infow("Откат PostgreSQL миграций...")
			if err := migrationMgr.MigratePostgresDown(); err != nil {
				logInstance.Fatalw("Failed to rollback PostgreSQL migrations", "error", err)
			}
			logInstance.Infow("PostgreSQL миграции успешно откачены")
		} else if *redisFlag {
			// Только Redis миграции вниз
			logInstance.Infow("Откат Redis миграций...")
			if err := migrationMgr.MigrateRedisDown(ctx); err != nil {
				logInstance.Fatalw("Failed to rollback Redis migrations", "error", err)
			}
			logInstance.Infow("Redis миграции успешно откачены")
		} else {
			// Все миграции вниз
			logInstance.Infow("Откат всех миграций...")
			if err := migrationMgr.MigrateDown(ctx); err != nil {
				logInstance.Fatalw("Failed to rollback migrations", "error", err)
			}
			logInstance.Infow("Все миграции успешно откачены")
		}
	} else {
		logInstance.Infow("Не указано действие для миграций. Используйте флаги -up или -down.")
	}
}
