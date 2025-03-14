package entity

import "time"

// User - модель пользователя с валидацией
type User struct {
	ID        int       `db:"id" json:"id"`
	UserName  string    `db:"username" json:"username,omitempty" validate:"required"`
	Email     string    `db:"email" json:"email,omitempty" validate:"required,email"`
	Password  string    `db:"password" json:"password" validate:"required"`
	CreatedAt time.Time `db:"created_at" json:"created_at,omitempty"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at,omitempty"`

}

// UserClaims - модель токена с валидацией
type UserClaims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email,omitempty"`
	ExpiresAt int64  `json:"exp,omitempty"`

