package connect

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log/slog"
)

type PostgresDB struct {
	conn *sqlx.DB
}

func NewPostgresDB(ctx context.Context, dsn string) (*PostgresDB, error) {
	db, err := sqlx.ConnectContext(ctx, "postgres", dsn)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		slog.Error("failed to ping database", "error", err)
		return nil, fmt.Errorf("ping database: %w", err)
	}

	slog.Info("successfully connected to database")
	return &PostgresDB{conn: db}, nil
}

func (db *PostgresDB) GetConn() *sqlx.DB {
	return db.conn
}

func (db *PostgresDB) Close() error {
	if db.conn != nil {
		slog.Info("closing database connection")
		if err := db.conn.Close(); err != nil {
			slog.Error("failed to close connection", "error", err)
			return fmt.Errorf("close connection: %w", err)
		}
	}
	return nil
}
