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
	query := `SELECT id, username, email, password FROM UsersLog WHERE id = $1`
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from db: %w", err)
	}
	return &user, nil
}

// GetUserByEmail получает пользователя из базы данных по email.
func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	query := `SELECT id, username, email, password FROM UsersLog WHERE email = $1`
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

// CreateUser создает нового пользователя в базе данных.
func (r *PostgresRepository) CreateUser(ctx context.Context, user *entity.User) error {
	query := `INSERT INTO UsersLog (username, email, password) 
             VALUES ($1, $2, $3) 
             RETURNING id`
	return r.db.QueryRowxContext(ctx, query, user.UserName, user.Email, user.Password).Scan(&user.ID)
}

// UpdateUser обновляет данные существующего пользователя в базе данных.
func (r *PostgresRepository) UpdateUser(ctx context.Context, user *entity.User) error {
	query := `UPDATE UsersLog 
             SET username = $1, email = $2, password = $3 
             WHERE id = $4`
	result, err := r.db.ExecContext(ctx, query, user.UserName, user.Email, user.Password, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}
