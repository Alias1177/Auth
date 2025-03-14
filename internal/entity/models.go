package entity

// User - модель пользователя с валидацией
type User struct {
	ID       int    `db:"id" json:"id"`
	UserName string `db:"username" json:"username,omitempty" validate:"required,min=2,max=50"`
	Email    string `db:"email" json:"email,omitempty" validate:"required,email_custom"`
	Password string `db:"password" json:"-" validate:"required,min=6,max=200"`
}

// UserClaims - модель токена с валидацией
type UserClaims struct {
	UserID    string `json:"user_id" validate:"required"`
	Email     string `json:"email,omitempty" validate:"omitempty,email"`
	ExpiresAt int64  `json:"exp,omitempty" validate:"required"`
}
