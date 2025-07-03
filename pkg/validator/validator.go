package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator представляет валидатор
type Validator struct {
	validate *validator.Validate
}

// New создает новый валидатор
func New() *Validator {
	v := validator.New()

	// Регистрируем пользовательские валидаторы
	v.RegisterValidation("password", validatePassword)

	// Регистрируем функцию для получения имен полей из JSON тегов
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{validate: v}
}

// Validate валидирует структуру
func (v *Validator) Validate(i interface{}) error {
	err := v.validate.Struct(i)
	if err != nil {
		return v.formatValidationError(err)
	}
	return nil
}

// formatValidationError форматирует ошибку валидации
func (v *Validator) formatValidationError(err error) error {
	var errors []string

	for _, err := range err.(validator.ValidationErrors) {
		switch err.Tag() {
		case "required":
			errors = append(errors, fmt.Sprintf("Field %s is required", err.Field()))
		case "email":
			errors = append(errors, fmt.Sprintf("Field %s must be a valid email", err.Field()))
		case "min":
			errors = append(errors, fmt.Sprintf("Field %s must be at least %s characters long", err.Field(), err.Param()))
		case "max":
			errors = append(errors, fmt.Sprintf("Field %s must be at most %s characters long", err.Field(), err.Param()))
		case "password":
			errors = append(errors, fmt.Sprintf("Field %s must contain at least one uppercase letter, one lowercase letter, one digit, and one special character", err.Field()))
		default:
			errors = append(errors, fmt.Sprintf("Field %s is invalid", err.Field()))
		}
	}

	return fmt.Errorf("validation failed: %s", strings.Join(errors, ", "))
}

// validatePassword пользовательский валидатор для пароля
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasNumber = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}
