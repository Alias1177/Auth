package service

import crypto "github.com/Alias1177/Auth/pkg/security"

type PasswordServiceImpl struct{}

func NewPasswordService() *PasswordServiceImpl {
	return &PasswordServiceImpl{}
}

func (s *PasswordServiceImpl) HashPassword(password string) (string, error) {
	return crypto.HashPassword(password)
}

func (s *PasswordServiceImpl) VerifyPassword(hashedPassword, password string) error {
	return crypto.VerifyPassword(hashedPassword, password)
}
