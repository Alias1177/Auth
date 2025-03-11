package entity

type User struct {
	ID       int    `db:"id"`
	UserName string `db:"username,omitempty"`
	Email    string `db:"email,omitempty"`
	Password string `db:"password,omitempty"`
}
