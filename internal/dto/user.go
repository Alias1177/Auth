package dto

import "time"

// UserDTO DTO для пользователя
type UserDTO struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateUserRequest DTO для обновления пользователя
type UpdateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
}

// UserResponse DTO для ответа с пользователем
type UserResponse struct {
	User UserDTO `json:"user"`
}

// ResetPasswordRequest DTO для сброса пароля
type ResetPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ChangePasswordRequest DTO для смены пароля
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

// ResetPasswordByEmailRequest DTO для сброса пароля по email
type ResetPasswordByEmailRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}
