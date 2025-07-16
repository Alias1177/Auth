package auth

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"text/template"

	"github.com/Alias1177/Auth/internal/domain"
	"github.com/Alias1177/Auth/internal/service"
	"github.com/Alias1177/Auth/pkg/logger"
	"github.com/go-chi/chi/v5"
	"github.com/markbates/goth/gothic"
)

const userTemplate = `
<p><a href="/logout/{{.Provider}}">logout</a></p>
<p>Name: {{.Name}} [{{.LastName}}, {{.FirstName}}]</p>
<p>Email: {{.Email}}</p>
<p>NickName: {{.NickName}}</p>
<p>Location: {{.Location}}</p>
<p>AvatarURL: {{.AvatarURL}} <img src="{{.AvatarURL}}"></p>
<p>Description: {{.Description}}</p>
<p>UserID: {{.UserID}}</p>
<p>AccessToken: {{.AccessToken}}</p>
<p>ExpiresAt: {{.ExpiresAt}}</p>
<p>RefreshToken: {{.RefreshToken}}</p>
`

type OAuthHandler struct {
	logger       *logger.Logger
	tokenManager service.TokenManager
	userRepo     service.UserRepository
}

func NewOAuthService(
	logger *logger.Logger,
	tokenManager service.TokenManager,
	userRepo service.UserRepository,
) *OAuthHandler {
	return &OAuthHandler{
		logger:       logger,
		tokenManager: tokenManager,
		userRepo:     userRepo,
	}
}

func (s *OAuthHandler) GetCallback(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")

	r = r.WithContext(context.WithValue(r.Context(), "provider", provider))

	gothUser, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		s.logger.Errorw("OAuth callback error", "error", err, "provider", provider)
		http.Error(w, "OAuth authentication failed", http.StatusInternalServerError)
		return
	}

	// Проверяем, существует ли пользователь в нашей БД
	existingUser, err := s.userRepo.GetUserByEmail(r.Context(), gothUser.Email)
	if err != nil {
		// Пользователь не существует, создаем нового
		newUser := &domain.User{
			Email:    gothUser.Email,
			UserName: gothUser.NickName,
			// Пароль не нужен для OAuth пользователей
		}

		if err := s.userRepo.CreateUser(r.Context(), newUser); err != nil {
			s.logger.Errorw("Failed to create OAuth user", "error", err, "email", gothUser.Email)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
		existingUser = newUser
	}

	// Генерируем JWT токены
	claims := domain.UserClaims{
		UserID: strconv.Itoa(existingUser.ID),
		Email:  existingUser.Email,
	}

	accessToken, err := s.tokenManager.GenerateAccessToken(claims)
	if err != nil {
		s.logger.Errorw("Failed to generate access token", "error", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := s.tokenManager.GenerateRefreshToken(claims)
	if err != nil {
		s.logger.Errorw("Failed to generate refresh token", "error", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Возвращаем JSON ответ с токенами
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{
		"access_token": "%s",
		"refresh_token": "%s",
		"user": {
			"id": %d,
			"email": "%s",
			"username": "%s"
		}
	}`, accessToken, refreshToken, existingUser.ID, existingUser.Email, existingUser.UserName)
}

func (s *OAuthHandler) GetLogout(w http.ResponseWriter, r *http.Request) {
	gothic.Logout(w, r)
	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (s *OAuthHandler) GetAuth(w http.ResponseWriter, r *http.Request) {
	if gothUser, err := gothic.CompleteUserAuth(w, r); err == nil {
		t, _ := template.New("foo").Parse(userTemplate)
		t.Execute(w, gothUser)
	} else {
		gothic.BeginAuthHandler(w, r)
	}
}
