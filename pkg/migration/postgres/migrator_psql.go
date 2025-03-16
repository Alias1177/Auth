package postgres

import (
	"Auth/pkg/logger"
	"Auth/pkg/migration"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"sync"
)

// Migrator handles database migrations
type Migrator struct {
	m    *migrate.Migrate
	log  *logger.Logger
	once sync.Once
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *sqlx.DB, s string, log *logger.Logger) (*Migrator, error) {
	// Передаём чистый *sql.DB из sqlx.DB
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		log.Errorw("Could not create migration driver", "error", err)
		return nil, fmt.Errorf("could not create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(

		fmt.Sprintf("file://%s", s),
		"postgres",
		driver,
	)
	if err != nil {
		log.Errorw("Could not create migrator", "error", err)
		return nil, fmt.Errorf("could not create migrator: %w", err)
	}

	return &Migrator{m: m, log: log}, nil
}

// Up runs all pending migrations
func (m *Migrator) Up() error {
	err := m.m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			m.log.Infow("No new migrations to apply")
			return migration.ErrNoChange
		}
		m.log.Errorw("Failed to apply migrations", "error", err)
		return fmt.Errorf("%w: %v", migration.ErrUpFailed, err)
	}
	m.log.Infow("Migrations applied successfully")
	return nil
}

// Down rolls back the last migration
func (m *Migrator) Down() error {
	err := m.m.Down()
	if err != nil {
		m.log.Errorw("Failed to rollback migration", "error", err)
		return fmt.Errorf("%w: %v", migration.ErrDownFailed, err)
	}
	m.log.Infow("Migration rolled back successfully")
	return nil
}

// Close closes the migrator
func (m *Migrator) Close() {
	m.once.Do(func() {
		err1, err2 := m.m.Close()
		if err1 != nil || err2 != nil {
			m.log.Errorw("Failed to close migrations", "errors", errors.Join(err1, err2))
		} else {
			m.log.Infow("Migrator closed successfully")
		}
	})
}
