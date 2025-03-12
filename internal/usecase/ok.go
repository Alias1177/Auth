package usecase

import (
	"Auth/internal/entity"
	"errors"
	"github.com/golang-jwt/jwt/v5"
)

// ParseRefreshToken распарсивает и валидирует Refresh JWT токен.
func (j *JWTTokenManager) ParseRefreshToken(tokenStr string) (entity.UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &entity.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.secret), nil
	})
	if err != nil {
		return entity.UserClaims{}, err
	}

	// Проверка валидности токена и его типа
	claims, ok := token.Claims.(*entity.UserClaims)
	if !ok || !token.Valid {
		return entity.UserClaims{}, errors.New("invalid refresh token")
	}

	return *claims, nil
}
