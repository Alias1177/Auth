package jwt

import (
	"Auth/config"
	"Auth/internal/entity"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
)

type JWTTokenManager struct {
	secret string
}

func NewJWTTokenManager(cfg config.JWTConfig) *JWTTokenManager {
	return &JWTTokenManager{
		secret: cfg.Secret,
	}
}

func (j *JWTTokenManager) GenerateAccessToken(userClaims entity.UserClaims) (string, error) {
	tokenClaims := jwt.MapClaims{
		"sub":   userClaims.UserID,
		"email": userClaims.Email,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	return token.SignedString([]byte(j.secret))
}

func (j *JWTTokenManager) ValidateAccessToken(token string) (*entity.UserClaims, error) {
	return j.validateToken(token, true)
}

func (j *JWTTokenManager) validateToken(token string, isAccess bool) (*entity.UserClaims, error) {
	// Парсинг токена
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Проверяем, что метод подписи соответствует HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(j.secret), nil
	})

	// Если возникла ошибка при парсинге токена
	if err != nil {
		return nil, fmt.Errorf("invalid token format: %w", err)
	}

	// Проверяем, является ли токен валидным
	if !parsedToken.Valid {
		return nil, errors.New("invalid token")
	}

	// Извлекаем claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims format")
	}

	// Логирование всех claims для отладки
	fmt.Println("Расшифрованные claims:", claims)

	// Извлечение нужных данных из claims
	userID, ok := claims["sub"].(string)
	if !ok {
		return nil, errors.New("missing or invalid 'sub' claim")
	}

	email, ok := claims["email"].(string)
	if !ok {
		return nil, errors.New("missing or invalid 'email' claim")
	}

	// Если все проверки прошли успешно, возвращаем данные о пользователе
	return &entity.UserClaims{
		UserID: userID,
		Email:  email,
	}, nil
}
