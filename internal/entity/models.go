package entity

type User struct {
	ID       int    `db:"id"`
	UserName string `db:"username,omitempty"`
	Email    string `db:"email,omitempty"`
	Password string `db:"password,omitempty"`
}
type UserClaims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email,omitempty"`
	ExpiresAt int64  `json:"exp"` // Явно указываем exp
}
