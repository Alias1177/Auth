package main

import (
	"Auth/pkg/logger"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

// AuthService - сервис аутентификации.
type AuthService struct {
	log *logger.Logger
}

// NewAuthService создает новый AuthService.
func NewAuthService(log *logger.Logger) *AuthService {
	return &AuthService{log: log}
}

// LoginRequest - структура для парсинга JSON-запроса.
type LoginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

// LoginHandler обрабатывает запрос на вход.
func (s *AuthService) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	// Декодируем JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.log.Errorw("Failed to parse JSON", "error", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Проверка входных данных
	if req.Name == "" {
		s.log.Warnw("User not found", "user_name", req.Name)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if req.Password == "" {
		s.log.Warnw("Password is nil", "user_name", req.Name)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	s.log.Infow("User logged in", "username", req.Name)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Login successful!"})
}

// SetupRouter инициализирует маршрутизатор.
func SetupRouter(authService *AuthService) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/login", authService.LoginHandler) // Подключаем обработчик
	return r
}

func main() {
	// Инициализируем логгер.
	logInstance, err := logger.NewSimpleLogger("info")
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logInstance.Close()

	// Создаем сервис аутентификации.
	authService := NewAuthService(logInstance)

	// Создаем маршрутизатор.
	router := SetupRouter(authService)

	// Запускаем сервер.
	logInstance.Infow("Starting server", "port", 8080)
	http.ListenAndServe(":8080", router)
}
