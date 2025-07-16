package oauth

import (
	"github.com/Alias1177/Auth/internal/config"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

const (
	key    = "randomstring"
	MaxAge = 60 * 60 * 24 * 30
	isProd = false
)

func NewOAuth(cfg *config.Config) {
	googleClientID := cfg.Google.ClientID
	googleClientSecret := cfg.Google.ClientSecret
	googleRedirectURL := cfg.Google.RedirectURL

	// Если redirect URL не задан, используем дефолтный
	if googleRedirectURL == "" {
		googleRedirectURL = "http://localhost:8080/auth/google/callback"
	}

	store := sessions.NewCookieStore([]byte(key))
	store.MaxAge(MaxAge)

	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = isProd

	gothic.Store = store

	goth.UseProviders(
		google.New(googleClientID, googleClientSecret, googleRedirectURL),
	)
}
