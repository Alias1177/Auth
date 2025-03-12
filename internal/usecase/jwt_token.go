package usecase

import (
	"Auth/config"
	"Auth/internal/entity"
	"errors"
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
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.secret), nil
	})
	if err != nil || !parsedToken.Valid {
		return nil, errors.New("token invalid")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return nil, errors.New("user_id missing")
	}

	userClaims := &entity.UserClaims{
		UserID: userID,
	}

	if email, found := claims["email"].(string); found && isAccess {
		userClaims.Email = email
	}

	return userClaims, nil
}
