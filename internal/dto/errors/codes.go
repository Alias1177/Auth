package errors

// ErrorCode представляет код ошибки
type ErrorCode string

const (
	// Общие ошибки
	ErrCodeInternal       ErrorCode = "INTERNAL_ERROR"
	ErrCodeValidation     ErrorCode = "VALIDATION_ERROR"
	ErrCodeNotFound       ErrorCode = "NOT_FOUND"
	ErrCodeUnauthorized   ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden      ErrorCode = "FORBIDDEN"
	ErrCodeConflict       ErrorCode = "CONFLICT"
	ErrCodeTooManyRequest ErrorCode = "TOO_MANY_REQUESTS"

	// Пользователи
	ErrCodeUserNotFound    ErrorCode = "USER_NOT_FOUND"
	ErrCodeUserExists      ErrorCode = "USER_EXISTS"
	ErrCodeInvalidPassword ErrorCode = "INVALID_PASSWORD"

	// Аутентификация
	ErrCodeInvalidToken  ErrorCode = "INVALID_TOKEN"
	ErrCodeExpiredToken  ErrorCode = "EXPIRED_TOKEN"
	ErrCodeInvalidLogin  ErrorCode = "INVALID_LOGIN"
	ErrCodeLoginRequired ErrorCode = "LOGIN_REQUIRED"

	// База данных
	ErrCodeDatabase    ErrorCode = "DATABASE_ERROR"
	ErrCodeRedis       ErrorCode = "REDIS_ERROR"
	ErrCodeTransaction ErrorCode = "TRANSACTION_ERROR"

	// Внешние сервисы
	ErrCodeKafka ErrorCode = "KAFKA_ERROR"
	ErrCodeEmail ErrorCode = "EMAIL_ERROR"
)
