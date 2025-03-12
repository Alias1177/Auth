package entity

import "github.com/golang-jwt/jwt/v5"

type User struct {
	ID       int    `db:"id"`
	UserName string `db:"username,omitempty"`
	Email    string `db:"email,omitempty"`
	Password string `db:"password,omitempty"`
}
type UserClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email,omitempty"`
	jwt.RegisteredClaims
}
