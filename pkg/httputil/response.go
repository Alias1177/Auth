package httputil

import (
	"encoding/json"
	"net/http"

	"github.com/Alias1177/Auth/pkg/logger"
)

// JSONResponse отправляет JSON ответ
func JSONResponse(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

// JSONError отправляет JSON ошибку
func JSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

// JSONErrorWithID отправляет JSON ошибку с id_message
func JSONErrorWithID(w http.ResponseWriter, statusCode int, idMessage int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":      "error",
		"id_message": idMessage,
		"code":       statusCode,
	})
}

// JSONSuccessWithID отправляет успешный JSON ответ с id_message
func JSONSuccessWithID(w http.ResponseWriter, statusCode int, idMessage int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(map[string]interface{}{
		"id_message": idMessage,
		"data":       data,
	})
}

// DecodeJSON декодирует JSON из запроса
func DecodeJSON(r *http.Request, dst interface{}, log *logger.Logger) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		if log != nil {
			log.Errorw("Failed to decode JSON request", "error", err)
		}
		return err
	}
	return nil
}

// SetTokenCookie устанавливает JWT токен в куки
func SetTokenCookie(w http.ResponseWriter, cookieName, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})
}

// SuccessResponse создает стандартный успешный ответ
func SuccessResponse(message string) map[string]string {
	return map[string]string{
		"message": message,
	}
}
