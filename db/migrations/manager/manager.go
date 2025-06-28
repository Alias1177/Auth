package manager

import (
	"Auth/pkg/logger"
	"Auth/pkg/migration"
	"Auth/pkg/migration/postgres"
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// MigrationManager предоставляет интерфейс для управления миграциями PostgreSQL
// (Redis миграции убраны)
type MigrationManager struct {
	pgMigrator *postgres.Migrator
	logger     *logger.Logger
}

// NewMigrationManager создаёт новый менеджер миграций только для PostgreSQL
func NewMigrationManager(db *sqlx.DB, log *logger.Logger, migrationsPath string) (*MigrationManager, error) {
	pgMigrator, err := postgres.NewMigrator(db, migrationsPath, log)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать мигратор PostgreSQL: %w", err)
	}
	return &MigrationManager{
		pgMigrator: pgMigrator,
		logger:     log,
	}, nil
}

// MigrateUp применяет миграции только для PostgreSQL
func (m *MigrationManager) MigrateUp(ctx context.Context) error {
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
	return nil
}

// MigrateDown откатывает последнюю миграцию только для PostgreSQL
func (m *MigrationManager) MigrateDown(ctx context.Context) error {
	m.logger.Infow("Откат последней миграции PostgreSQL...")
	if err := m.pgMigrator.Down(); err != nil {
		return fmt.Errorf("ошибка при откате миграции PostgreSQL: %w", err)
	}
	m.logger.Infow("Миграция PostgreSQL успешно откатчена")
	return nil
}

// Close закрывает ресурсы мигратора
func (m *MigrationManager) Close() {
	if m.pgMigrator != nil {
		m.pgMigrator.Close()
	}
}
