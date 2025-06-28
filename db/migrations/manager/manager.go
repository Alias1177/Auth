package manager

import (
	"context"
	"errors"
	"fmt"

	"github.com/Alias1177/Auth/pkg/logger"
	"github.com/Alias1177/Auth/pkg/migration"
	"github.com/Alias1177/Auth/pkg/migration/postgres"
	redis_migration "github.com/Alias1177/Auth/pkg/migration/redis"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// MigrationManager предоставляет интерфейс для управления миграциями PostgreSQL и Redis
type MigrationManager struct {
	pgMigrator    *postgres.Migrator
	redisMigrator *redis_migration.RedisMigrator
	logger        *logger.Logger
}

// NewMigrationManager создаёт новый менеджер миграций
func NewMigrationManager(db *sqlx.DB, redisClient *redis.Client, log *logger.Logger, migrationsPath string) (*MigrationManager, error) {
	// Создаём мигратор PostgreSQL
	pgMigrator, err := postgres.NewMigrator(db, migrationsPath, log)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать мигратор PostgreSQL: %w", err)
	}

	// Создаём мигратор Redis
	redisMigrator := redis_migration.NewRedisMigrator(redisClient, log)

	return &MigrationManager{
		pgMigrator:    pgMigrator,
		redisMigrator: redisMigrator,
		logger:        log,
	}, nil
}

// MigrateUp применяет миграции для обеих баз данных
func (m *MigrationManager) MigrateUp(ctx context.Context) error {
	// Применяем миграции PostgreSQL
	m.logger.Infow("Запуск миграций PostgreSQL...")
	if err := m.pgMigrator.Up(); err != nil {
		if errors.Is(err, migration.ErrNoChange) {
			m.logger.Infow("Нет новых миграций PostgreSQL")
		} else {
			return fmt.Errorf("ошибка при применении миграций PostgreSQL: %w", err)
		}
	} else {
		m.logger.Infow("Миграции PostgreSQL успешно применены")
	}

	// Применяем миграции Redis
	m.logger.Infow("Запуск миграций Redis...")
	if err := m.redisMigrator.Up(ctx); err != nil {
		return fmt.Errorf("ошибка при применении миграций Redis: %w", err)
	}
	m.logger.Infow("Миграции Redis успешно применены")

	return nil
}

// MigrateDown откатывает последнюю миграцию для обеих баз данных
func (m *MigrationManager) MigrateDown(ctx context.Context) error {
	// Откатываем миграцию PostgreSQL
	m.logger.Infow("Откат последней миграции PostgreSQL...")
	if err := m.pgMigrator.Down(); err != nil {
		return fmt.Errorf("ошибка при откате миграции PostgreSQL: %w", err)
	}
	m.logger.Infow("Миграция PostgreSQL успешно откачена")

	// Откатываем миграцию Redis
	m.logger.Infow("Откат миграций Redis...")
	if err := m.redisMigrator.Down(ctx); err != nil {
		return fmt.Errorf("ошибка при откате миграций Redis: %w", err)
	}
	m.logger.Infow("Миграции Redis успешно откачены")

	return nil
}

// MigratePostgresUp применяет только миграции PostgreSQL
func (m *MigrationManager) MigratePostgresUp() error {
	m.logger.Infow("Запуск миграций PostgreSQL...")
	if err := m.pgMigrator.Up(); err != nil {
		if errors.Is(err, migration.ErrNoChange) {
			m.logger.Infow("Нет новых миграций PostgreSQL")
			return nil
		}
		return fmt.Errorf("ошибка при применении миграций PostgreSQL: %w", err)
	}
	m.logger.Infow("Миграции PostgreSQL успешно применены")
	return nil
}

// MigratePostgresDown откатывает только миграции PostgreSQL
func (m *MigrationManager) MigratePostgresDown() error {
	m.logger.Infow("Откат последней миграции PostgreSQL...")
	if err := m.pgMigrator.Down(); err != nil {
		return fmt.Errorf("ошибка при откате миграции PostgreSQL: %w", err)
	}
	m.logger.Infow("Миграция PostgreSQL успешно откачена")
	return nil
}

// MigrateRedisUp применяет только миграции Redis
func (m *MigrationManager) MigrateRedisUp(ctx context.Context) error {
	m.logger.Infow("Запуск миграций Redis...")
	if err := m.redisMigrator.Up(ctx); err != nil {
		return fmt.Errorf("ошибка при применении миграций Redis: %w", err)
	}
	m.logger.Infow("Миграции Redis успешно применены")
	return nil
}

// MigrateRedisDown откатывает только миграции Redis
func (m *MigrationManager) MigrateRedisDown(ctx context.Context) error {
	m.logger.Infow("Откат миграций Redis...")
	if err := m.redisMigrator.Down(ctx); err != nil {
		return fmt.Errorf("ошибка при откате миграций Redis: %w", err)
	}
	m.logger.Infow("Миграции Redis успешно откачены")
	return nil
}

// Close закрывает ресурсы миграторов
func (m *MigrationManager) Close() {
	if m.pgMigrator != nil {
		m.pgMigrator.Close()
	}
}
