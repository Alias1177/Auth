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

// MessageEnByID возвращает английский текст по id_message
func MessageEnByID(id int) string {
	switch id {
	case 1000:
		return "You have successfully logged in"
	case 1001:
		return "User successfully registered"
	case 1002:
		return "You have successfully logged out"
	case 1003:
		return "Password successfully reset"
	case 1004:
		return "User successfully updated"
	case 1005:
		return "Token successfully refreshed"
	case 1006:
		return "Password successfully changed"
	case 1007:
		return "Password reset request sent"
	case 1008:
		return "Password reset confirmed"
	case 1009:
		return "User information retrieved"
	case 2000:
		return "Invalid email"
	case 2001:
		return "Invalid password"
	case 2002:
		return "Invalid token"
	case 2003:
		return "Invalid request"
	case 2004:
		return "Invalid reset code"
	case 2005:
		return "Password too weak"
	case 2006:
		return "User with this email already exists"
	case 2007:
		return "Invalid user ID"
	case 2008:
		return "Email and password must be filled"
	case 2009:
		return "Invalid request data"
	case 3000:
		return "Wrong password"
	case 3001:
		return "User not found"
	case 3002:
		return "Token expired"
	case 3003:
		return "Invalid token"
	case 3004:
		return "Unauthorized access"
	case 3005:
		return "Reset code expired"
	case 3006:
		return "Invalid reset code"
	case 3007:
		return "Resource not found"
	case 3008:
		return "Too many reset attempts"
	case 4000:
		return "Internal server error"
	case 4001:
		return "Database error"
	case 4002:
		return "Email send error"
	case 4003:
		return "Token generation error"
	case 4004:
		return "Password hashing error"
	case 4005:
		return "Response encoding error"
	case 4006:
		return "User claims error"
	default:
		return "Unknown message"
	}
}

// JSONErrorWithID отправляет JSON ошибку с id_message и message_en
func JSONErrorWithID(w http.ResponseWriter, statusCode int, idMessage int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id_message": idMessage,
		"message_en": MessageEnByID(idMessage),
	})
}

// JSONSuccessWithID отправляет успешный JSON ответ с id_message и message_en
func JSONSuccessWithID(w http.ResponseWriter, statusCode int, idMessage int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(map[string]interface{}{
		"id_message": idMessage,
		"message_en": MessageEnByID(idMessage),
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
