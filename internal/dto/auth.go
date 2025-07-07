package dto

// LoginRequest DTO для запроса входа
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginResponse DTO для ответа входа
type LoginResponse struct {
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
	User         UserDTO `json:"user"`
}

// RegisterRequest DTO для запроса регистрации
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// RegisterResponse DTO для ответа регистрации
type RegisterResponse struct {
	User UserDTO `json:"user"`
}

// RefreshRequest DTO для обновления токена
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshResponse DTO для ответа обновления токена
type RefreshResponse struct {
	AccessToken string `json:"access_token"`
}

// RequestPasswordResetRequest DTO для запроса сброса пароля
type RequestPasswordResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// RequestPasswordResetResponse DTO для ответа запроса сброса пароля
type RequestPasswordResetResponse struct {
	Code string `json:"code,omitempty"`
}

// ConfirmPasswordResetRequest DTO для подтверждения сброса пароля
type ConfirmPasswordResetRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Code     string `json:"code" validate:"required,min=6,max=6"`
	Password string `json:"password" validate:"required,min=8"`
}

// ConfirmPasswordResetResponse DTO для ответа подтверждения сброса пароля
type ConfirmPasswordResetResponse struct {
}
