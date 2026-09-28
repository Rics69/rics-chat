package auth_service

import (
	"errors"
	"fmt"

	core_errors "github.com/Rics69/rics-chat/internal/core/errors"
	"golang.org/x/crypto/bcrypt"
)

// заранее посчитанный bcrypt-хэш произвольной строки, нужен только для выравнивания времени в Login
const dummyPasswordHash = "$2a$10$XiSzhz9bElVkG6lrVI4dy.sRKROrXWxUC5hhaqrC0md3xR93V4U9u"

func hashPassword(password string) (string, error) {
	// соль генерируется внутри и хранится прямо в хэше, отдельная колонка не нужна
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("bcrypt generate: %w", err)
	}

	return string(hash), nil
}

func comparePassword(hash string, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return fmt.Errorf("password mismatch: %w", core_errors.ErrUnauthenticated)
		}

		return fmt.Errorf("bcrypt compare: %w", err)
	}

	return nil
}
