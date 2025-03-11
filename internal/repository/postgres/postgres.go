// internal/repository/postgres/postgres.go
package postgres

import (
	"Auth/internal/entity"
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// PostgresRepository предоставляет методы для работы с PostgreSQL.
type PostgresRepository struct {
	db *sqlx.DB
}

// NewPostgresRepository создает новый экземпляр PostgresRepository.
func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// GetUserByID получает пользователя из базы данных по ID.
func (r *PostgresRepository) GetUserByID(ctx context.Context, id int) (*entity.User, error) {
	var user entity.User
	query := `SELECT id, email, password FROM users WHERE id = $1`
	// Выполняем запрос и сканируем результат в объект user.
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from db: %w", err)
	}
	return &user, nil
}

// GetUserByEmail получает пользователя из базы данных по email.
func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	query := `SELECT id, email, password FROM users WHERE email = $1`
	// Выполняем запрос и сканируем результат в объект user.
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

// CreateUser создает нового пользователя в базе данных.
func (r *PostgresRepository) CreateUser(ctx context.Context, user *entity.User) error {
	query := `INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id`
	// Выполняем запрос на вставку и получаем созданный ID пользователя.
	row := r.db.QueryRowxContext(ctx, query, user.Email, user.Password)
	return row.Scan(&user.ID)
}

// UpdateUser обновляет данные существующего пользователя в базе данных.
func (r *PostgresRepository) UpdateUser(ctx context.Context, user *entity.User) error {
	query := `UPDATE users SET email = $1, password = $2 WHERE id = $3`
	// Выполняем запрос на обновление.
	result, err := r.db.ExecContext(ctx, query, user.Email, user.Password, user.ID)
	if err != nil {
		return err
	}

	// Проверяем количество затронутых строк.
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}
