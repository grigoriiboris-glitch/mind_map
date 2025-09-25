package userreq

import (
    "strings"
    "github.com/mymindmap/api/pkg/core/validator"
)
type Register struct {
    Name     string `json:"name" validate:"required,min=2,max=100"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,strongpassword"`
}

func (r *Register) Validate() error {
    r.Email = strings.ToLower(strings.TrimSpace(r.Email))
    r.Password = strings.TrimSpace(r.Password)
    r.Name = strings.TrimSpace(r.Name)

    return validator.Validate(r)
}

// // validatePassword проверяет сложность пароля
// // Требования: минимум 8 символов, цифры, буквы в разных регистрах, спецсимволы
// func (s *AuthService) validatePassword(password string) error {
// 	if len(password) < MinPasswordLength {
// 		return fmt.Errorf("password must be at least %d characters long", MinPasswordLength)
// 	}

// 	if len(password) > MaxPasswordLength {
// 		return fmt.Errorf("password must not exceed %d characters", MaxPasswordLength)
// 	}

// 	// Проверка сложности пароля
// 	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)           // Есть цифры
// 	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)           // Есть строчные буквы
// 	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)           // Есть заглавные буквы
// 	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~` + "`" + `]`).MatchString(password) // Есть спецсимволы

// 	if !hasDigit || !hasLower || !hasUpper || !hasSpecial {
// 		return ErrWeakPassword
// 	}

// 	return nil
// }