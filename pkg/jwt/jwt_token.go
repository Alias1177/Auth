package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/Alias1177/Auth/internal/config"
	"github.com/Alias1177/Auth/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

type JWTTokenManager struct {
	secret string
}

type RefreshTokenStruct struct {
	Token string `json:"refresh_token"`
}

func NewJWTTokenManager(cfg config.JWTConfig) *JWTTokenManager {
	return &JWTTokenManager{
		secret: cfg.Secret,
	}
}

func (j *JWTTokenManager) GenerateAccessToken(userClaims domain.UserClaims) (string, error) {
	exp := time.Now().Add(15 * time.Minute).Unix()
	tokenClaims := jwt.MapClaims{
		"sub":   userClaims.UserID,
		"email": userClaims.Email,
		"exp":   exp,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	return token.SignedString([]byte(j.secret))
}

func (j *JWTTokenManager) ValidateAccessToken(token string) (*domain.UserClaims, error) {
	return j.validateToken(token, true)
}

func (j *JWTTokenManager) validateToken(token string, isAccess bool) (*domain.UserClaims, error) {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(j.secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token format: %w", err)
	}
	if !parsedToken.Valid {
		return nil, errors.New("invalid token")
	}
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims format")
	}

	// Проверка exp
	if expRaw, ok := claims["exp"]; ok {
		switch exp := expRaw.(type) {
		case float64:
			if int64(exp) < time.Now().Unix() {
				return nil, errors.New("token expired")
			}
		case int64:
			if exp < time.Now().Unix() {
				return nil, errors.New("token expired")
			}
		}
	}
	// Проверка типа для refresh
	if !isAccess {
		t, ok := claims["type"].(string)
		if !ok || t != "refresh" {
			return nil, errors.New("not a refresh token")
		}
	}
	userID, ok := claims["sub"].(string)
	if !ok {
		return nil, errors.New("missing or invalid 'sub' claim")
	}
	email, ok := claims["email"].(string)
	if !ok {
		return nil, errors.New("missing or invalid 'email' claim")
	}

	return &domain.UserClaims{
		UserID: userID,
		Email:  email,
	}, nil
}

func (j *JWTTokenManager) GenerateRefreshToken(userClaims domain.UserClaims) (string, error) {
	exp := time.Now().Add(7 * 24 * time.Hour).Unix()
	claims := jwt.MapClaims{
		"sub":   userClaims.UserID,
		"email": userClaims.Email,
		"type":  "refresh",
		"exp":   exp,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secret))
}

func (j *JWTTokenManager) RefreshTokens(refreshToken string) (string, string, error) {
	claims, err := j.validateToken(refreshToken, false)
	if err != nil {
		return "", "", err
	}

	newAccessToken, err := j.GenerateAccessToken(*claims)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := j.GenerateRefreshToken(*claims)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

func (j *JWTTokenManager) ValidateRefreshToken(token string) (*domain.UserClaims, error) {
	return j.validateToken(token, false)
}
