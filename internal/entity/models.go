package entity

type User struct {
	ID       int    `db:"id" json:"id"`
	UserName string `db:"username" json:"username,omitempty"`
	Email    string `db:"email" json:"email,omitempty"`
	Password string `db:"password" json:"password"` // Пароль не уходит в JSON
}

type UserClaims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email,omitempty"`
	ExpiresAt int64  `json:"exp,omitempty"` // Убираем из JSON, если пустое
}
