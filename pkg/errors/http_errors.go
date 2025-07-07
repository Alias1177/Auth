package errors

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/Alias1177/Auth/internal/dto"
	"github.com/Alias1177/Auth/pkg/httputil"
	"github.com/Alias1177/Auth/pkg/logger"
)

// HTTPError представляет HTTP ошибку
type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *HTTPError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

// NewHTTPError создает новую HTTP ошибку
func NewHTTPError(code int, message string, err error) *HTTPError {
	return &HTTPError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// HandleDatabaseError обрабатывает ошибки базы данных
func HandleDatabaseError(w http.ResponseWriter, err error, log *logger.Logger, operation string) {
	if errors.Is(err, sql.ErrNoRows) {
		log.Warnw("Resource not found", "operation", operation, "error", err)
		httputil.JSONErrorWithID(w, http.StatusNotFound, dto.MsgResourceNotFound)
		return
	}

	log.Errorw("Database error", "operation", operation, "error", err)
	httputil.JSONErrorWithID(w, http.StatusInternalServerError, dto.MsgDatabaseError)
}

// HandleValidationError обрабатывает ошибки валидации
func HandleValidationError(w http.ResponseWriter, err error, log *logger.Logger) {
	log.Warnw("Validation error", "error", err)
	httputil.JSONErrorWithID(w, http.StatusBadRequest, dto.MsgInvalidRequestData)
}

// HandleInternalError обрабатывает внутренние ошибки
func HandleInternalError(w http.ResponseWriter, err error, log *logger.Logger, operation string) {
	log.Errorw("Internal error", "operation", operation, "error", err)
	httputil.JSONErrorWithID(w, http.StatusInternalServerError, dto.MsgInternalError)
}

// HandleUnauthorizedError обрабатывает ошибки авторизации
func HandleUnauthorizedError(w http.ResponseWriter, message string, log *logger.Logger) {
	log.Warnw("Unauthorized access", "message", message)
	httputil.JSONErrorWithID(w, http.StatusUnauthorized, dto.MsgUnauthorized)
}
