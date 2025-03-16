package migration

import "errors"

// TODO детальная обработка ошибок
// Определяем кастомные ошибки
var (
	ErrNoChange   = errors.New("no change in migration")
	ErrUpFailed   = errors.New("migration up failed")
	ErrDownFailed = errors.New("migration down failed")
)

const (
	ErrCheckKeys   = "Failed to check existing keys"
	ErrInsertUser  = "Failed to insert initial user"
	ErrDeleteUser  = "Failed to delete test user"
	ErrMarshalUser = "Failed to marshal user data"
)
