package entity

type User struct {
	ID       int    `db:"id"`
	Email    string `db:"email,omitempty"`
	Password string `db:"password,omitempty"`
}
