package usecase

import (
	"Auth/config"
	"Auth/internal/entity"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type JWTTokenManager struct {
	secret                     string
	accessTokenTTL, refreshTTL time.Duration
}

func NewJWTTokenManager(cfg config.JWTConfig) *JWTTokenManager {
	return &JWTTokenManager{
		secret:         cfg.Secret,
		accessTokenTTL: cfg.AccessTokenTTL,
		refreshTTL:     cfg.RefreshTokenTTL,
	}
}

func (j *JWTTokenManager) GenerateAccessToken(userClaims entity.UserClaims) (string, error) {
	tokenClaims := jwt.MapClaims{
		"sub":   userClaims.UserID,
		"email": userClaims.Email,
		"exp":   time.Now().Add(j.accessTokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	return token.SignedString([]byte(j.secret))
}

func (j *JWTTokenManager) GenerateRefreshToken(userClaims entity.UserClaims) (string, error) {
	tokenClaims := jwt.MapClaims{
		"sub": userClaims.UserID,
		"exp": time.Now().Add(j.refreshTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	return token.SignedString([]byte(j.secret))
}

func (j *JWTTokenManager) ValidateAccessToken(token string) (*entity.UserClaims, error) {
	return j.validateToken(token, true)
}

func (j *JWTTokenManager) ValidateRefreshToken(token string) (*entity.UserClaims, error) {
	return j.validateToken(token, false)
}

func (j *JWTTokenManager) validateToken(token string, isAccess bool) (*entity.UserClaims, error) {
	fmt.Println("Получен токен:", token)
	fmt.Println("Текущее время сервера:", time.Now().Unix())

	// Для jwt/v5 используем другой подход
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		fmt.Println("Проверяем метод подписи:", token.Method.Alg())
		return []byte(j.secret), nil
	}, jwt.WithoutClaimsValidation()) // Этот параметр отключает проверку claims включая exp

	if err != nil {
		fmt.Println("Ошибка парсинга токена:", err)
		return nil, errors.New("token invalid")
	}

	if !parsedToken.Valid {
		fmt.Println("Токен недействителен!")
		return nil, errors.New("token invalid")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	fmt.Println("Расшифрованные claims:", claims)

	// В jwt/v5 преобразование к строке может потребовать дополнительной проверки
	userID, ok := claims["sub"].(string)
	if !ok {
		return nil, errors.New("invalid subject claim")
	}

	email, ok := claims["email"].(string)
	if !ok {
		return nil, errors.New("invalid email claim")
	}

	return &entity.UserClaims{
		UserID: userID,
		Email:  email,
	}, nil
}
