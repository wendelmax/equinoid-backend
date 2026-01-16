package auth

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
)

// hashPassword gera um hash bcrypt da senha
func hashPassword(password string, cost int) (string, error) {
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// checkPassword verifica se a senha confere com o hash
func checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePasswordStrength valida a força da senha
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}

	if len(password) > 128 {
		return ErrPasswordTooLong
	}

	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case 'a' <= char && char <= 'z':
			hasLower = true
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case '0' <= char && char <= '9':
			hasDigit = true
		case char >= 32 && char <= 126: // ASCII printable characters
			if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
				hasSpecial = true
			}
		}
	}

	if !hasLower {
		return ErrPasswordNeedsLowercase
	}

	if !hasUpper {
		return ErrPasswordNeedsUppercase
	}

	if !hasDigit {
		return ErrPasswordNeedsDigit
	}

	if !hasSpecial {
		return ErrPasswordNeedsSpecialChar
	}

	return nil
}

// Erros de validação de senha
var (
	ErrPasswordTooShort         = errors.New("password must be at least 8 characters long")
	ErrPasswordTooLong          = errors.New("password must be at most 128 characters long")
	ErrPasswordNeedsLowercase   = errors.New("password must contain at least one lowercase letter")
	ErrPasswordNeedsUppercase   = errors.New("password must contain at least one uppercase letter")
	ErrPasswordNeedsDigit       = errors.New("password must contain at least one digit")
	ErrPasswordNeedsSpecialChar = errors.New("password must contain at least one special character")
)
