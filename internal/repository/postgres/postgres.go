package postgres

import (
	"Auth/internal/entity"
	"Auth/internal/repository/redis"
	"Auth/pkg/logger"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// PostgresRepository предоставляет методы для работы с PostgreSQL.
type PostgresRepository struct {
	db        *sqlx.DB
	redisRepo *redis.RedisRepository // Добавляем Redis репозиторий
	log       *logger.Logger
}

func NewPostgresRepository(db *sqlx.DB, redisRepo *redis.RedisRepository, log *logger.Logger) *PostgresRepository {
	return &PostgresRepository{
		db:        db,
		redisRepo: redisRepo,
		log:       log,
	}
}

// GetUserByID получает пользователя из базы данных по ID.
func (r *PostgresRepository) GetUserByID(ctx context.Context, id int) (*entity.User, error) {
	var user entity.User
	query := `SELECT id, username, email, password FROM UsersLog WHERE id = $1`
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		r.log.Errorw("Get err", "err", err)
		return nil, fmt.Errorf("failed to get user from db: %w", err)
	}
	return &user, nil
}

// GetUserByEmail получает пользователя из базы данных по email.
func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := "SELECT id, username, email, password FROM UsersLog WHERE email = $1"
	user := entity.User{}

	err := r.db.QueryRowContext(ctx, query, email).
		Scan(&user.ID, &user.UserName, &user.Email, &user.Password)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log.Errorw("Get err", "err", err)
			return nil, sql.ErrNoRows
		}
		r.log.Errorw("Get err", "err", err)
		return nil, err // другая ошибка (например, ошибка соединения или запроса)
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
		r.log.Errorw("Update err", "err", err)
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.Errorw("failed to get rows", "err", err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		r.log.Errorw("no rows updated", "id", user.ID)
		return fmt.Errorf("user not found")
	}

	err = r.redisRepo.SetUser(ctx, user)
	if err != nil {
		r.log.Errorw("redis set err", "err", err)
		return fmt.Errorf("failed to update user in redis: %w", err)
	}

	return nil
}
