package server

import (
	"context"
	"net/http"

	"github.com/Alias1177/Auth/internal/app/container"
	"github.com/Alias1177/Auth/internal/infrastructure/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server представляет HTTP сервер
type Server struct {
	router    chi.Router
	container *container.Container
	server    *http.Server
}

// New создает новый HTTP сервер
func New(container *container.Container) (*Server, error) {
	s := &Server{
		container: container,
		router:    chi.NewRouter(),
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s, nil
}

// setupMiddleware настраивает middleware
func (s *Server) setupMiddleware() {
	logger := s.container.GetLogger()

	// CORS middleware
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{}, // Пустой список (разрешим динамически)
		AllowOriginFunc: func(r *http.Request, origin string) bool {
			return true
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Logger middleware
	loggerMiddleware := middleware.NewLoggerMiddleware(logger)
	s.router.Use(loggerMiddleware.Handler)

	// Metrics middleware
	metrics := middleware.NewMetricsMiddleware("auth_service")
	s.router.Use(metrics.Middleware)
}

// setupRoutes настраивает маршруты
func (s *Server) setupRoutes() {
	authHandler := s.container.GetAuthHandler()
	registrationHandler := s.container.GetRegistrationHandler()
	userHandler := s.container.GetUserHandler()
	tokenManager := s.container.GetTokenManager()

	// Публичные маршруты
	s.router.Post("/login", authHandler.Login)
	s.router.Post("/register", registrationHandler.Register)
	s.router.Handle("/metrics", promhttp.Handler())
	s.router.Post("/refresh-token", authHandler.Refresh)

	// Защищённые маршруты
	s.router.Route("/user", func(r chi.Router) {
		r.Use(middleware.JWTAuthMiddleware(tokenManager))
		r.Patch("/{id}", userHandler.UpdateUser)
		r.Get("/me", userHandler.GetUserInfoHandler)
	})
}

// Start запускает HTTP сервер
func (s *Server) Start(addr string) error {
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	return s.server.ListenAndServe()
}

// Shutdown gracefully останавливает сервер
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}
