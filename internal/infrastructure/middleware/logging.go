package middleware

import (
	"Auth/pkg/logger"
	"bufio"
	"context"
	"errors"
	"github.com/google/uuid"
	"net"
	"net/http"
	"sync"
	"time"
)

// RequestIDKey is the key for request ID in the context.
const RequestIDKey = "request_id"

// LoggerMiddleware is a middleware that logs HTTP requests.
type LoggerMiddleware struct {
	log *logger.Logger
}

// NewLoggerMiddleware creates a new logger middleware.
func NewLoggerMiddleware(log *logger.Logger) *LoggerMiddleware {
	return &LoggerMiddleware{
		log: log,
	}
}

// Handler wraps an HTTP handler with request logging.
func (m *LoggerMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := uuid.New().String()

		// Add request ID to context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		r = r.WithContext(ctx)

		// Wrap response writer to capture status and size
		rw := newResponseWriter(w)

		// Pre-request logging
		m.log.Infow("Request started",
			"request_id", requestID,
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)

		// Handle panic recovery
		defer func() {
			if rec := recover(); rec != nil {
				m.log.Errorw("Request panicked",
					"request_id", requestID,
					"error", rec,
				)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Post-request logging
		duration := time.Since(start)
		logFunc := m.getLogFunc(rw.statusCode)
		logFunc("Request completed",
			"request_id", requestID,
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.statusCode,
			"duration_ms", duration.Milliseconds(),
			"size", rw.size,
		)
	})
}

// getLogFunc selects the appropriate log function based on status code.
func (m *LoggerMiddleware) getLogFunc(statusCode int) func(string, ...interface{}) {
	if statusCode >= 500 {
		return m.log.Errorw
	} else if statusCode >= 400 {
		return m.log.Warnw
	}
	return m.log.Infow
}

// responseWriter is a wrapper for http.ResponseWriter that captures the status code and size.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
	mu         sync.Mutex // Protect concurrent writes
}

// newResponseWriter creates a new responseWriter.
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // Default to 200 OK
	}
}

// WriteHeader captures the status code.
func (rw *responseWriter) WriteHeader(code int) {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the response size and handles errors.
func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	size, err := rw.ResponseWriter.Write(b)
	if err == nil {
		rw.size += size
	}
	return size, err
}

// Hijack supports the Hijacker interface if the underlying writer supports it.
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	if hj, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, errors.New("hijack not supported")
}
