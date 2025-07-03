package errors

import (
	"fmt"
	"net/http"
)

// AppError представляет ошибку приложения
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	HTTPStatus int       `json:"-"`
	Cause      error     `json:"-"`
}

// Error реализует интерфейс error
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap возвращает причину ошибки
func (e *AppError) Unwrap() error {
	return e.Cause
}

// NewAppError создает новую ошибку приложения
func NewAppError(code ErrorCode, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// NewAppErrorWithCause создает новую ошибку приложения с причиной
func NewAppErrorWithCause(code ErrorCode, message string, httpStatus int, cause error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Cause:      cause,
	}
}

// Предопределенные ошибки
var (
	ErrInternal        = NewAppError(ErrCodeInternal, "Internal server error", http.StatusInternalServerError)
	ErrValidation      = NewAppError(ErrCodeValidation, "Validation error", http.StatusBadRequest)
	ErrNotFound        = NewAppError(ErrCodeNotFound, "Resource not found", http.StatusNotFound)
	ErrUnauthorized    = NewAppError(ErrCodeUnauthorized, "Unauthorized", http.StatusUnauthorized)
	ErrForbidden       = NewAppError(ErrCodeForbidden, "Forbidden", http.StatusForbidden)
	ErrConflict        = NewAppError(ErrCodeConflict, "Conflict", http.StatusConflict)
	ErrTooManyRequests = NewAppError(ErrCodeTooManyRequest, "Too many requests", http.StatusTooManyRequests)

	// Пользователи
	ErrUserNotFound    = NewAppError(ErrCodeUserNotFound, "User not found", http.StatusNotFound)
	ErrUserExists      = NewAppError(ErrCodeUserExists, "User already exists", http.StatusConflict)
	ErrInvalidPassword = NewAppError(ErrCodeInvalidPassword, "Invalid password", http.StatusBadRequest)

	// Аутентификация
	ErrInvalidToken  = NewAppError(ErrCodeInvalidToken, "Invalid token", http.StatusUnauthorized)
	ErrExpiredToken  = NewAppError(ErrCodeExpiredToken, "Token expired", http.StatusUnauthorized)
	ErrInvalidLogin  = NewAppError(ErrCodeInvalidLogin, "Invalid login credentials", http.StatusUnauthorized)
	ErrLoginRequired = NewAppError(ErrCodeLoginRequired, "Login required", http.StatusUnauthorized)

	// База данных
	ErrDatabase    = NewAppError(ErrCodeDatabase, "Database error", http.StatusInternalServerError)
	ErrRedis       = NewAppError(ErrCodeRedis, "Redis error", http.StatusInternalServerError)
	ErrTransaction = NewAppError(ErrCodeTransaction, "Transaction error", http.StatusInternalServerError)

	// Внешние сервисы
	ErrKafka = NewAppError(ErrCodeKafka, "Kafka error", http.StatusInternalServerError)
	ErrEmail = NewAppError(ErrCodeEmail, "Email service error", http.StatusInternalServerError)
)
