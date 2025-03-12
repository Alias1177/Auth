package usecase

import (
	"Auth/internal/entity"
	"errors"
	"github.com/golang-jwt/jwt/v5"
)

// ParseRefreshToken распарсивает и валидирует Refresh JWT токен.
func (j *JWTTokenManager) ParseRefreshToken(tokenStr string) (entity.UserClaims, error) {
	parsedToken, err := jwt.ParseWithClaims(tokenStr, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.secret), nil
	})
	if err != nil {
		return entity.UserClaims{}, err
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || !parsedToken.Valid {
		return entity.UserClaims{}, errors.New("invalid refresh token")
	}

	// Парсим user_id (sub) и email
	userID, ok := claims["sub"].(string)
	if !ok {
		return entity.UserClaims{}, errors.New("user_id missing")
	}

	email, _ := claims["email"].(string) // email может быть пустым, поэтому не проверяем ok

	return entity.UserClaims{
		UserID: userID,
		Email:  email,
	}, nil
}
