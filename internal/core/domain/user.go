package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	core_errors "github.com/Rics69/rics-chat/internal/core/errors"
)

var loginRegexp = regexp.MustCompile(`^[a-z0-9_]{3,32}$`)

type User struct {
	ID int64

	Login        string
	PasswordHash string
	CreatedAt    time.Time
}

func NewUser(id int64, login string, passwordHash string, createdAt time.Time) User {
	return User{
		ID:           id,
		Login:        login,
		PasswordHash: passwordHash,
		CreatedAt:    createdAt,
	}
}

func NewUserUnitialized(login string, passwordHash string) User {
	return NewUser(UnitializedID, login, passwordHash, time.Now())
}

func (u *User) Validate() error {
	if err := ValidateLogin(u.Login); err != nil {
		return err
	}

	if u.PasswordHash == "" {
		return fmt.Errorf("'PasswordHash' can't be empty: %w", core_errors.ErrInvalidArgument)
	}

	return nil
}

func NormalizeLogin(login string) string {
	return strings.ToLower(strings.TrimSpace(login))
}

func ValidateLogin(login string) error {
	if !loginRegexp.MatchString(login) {
		return fmt.Errorf(
			"invalid 'Login' format: must match %s: %w",
			loginRegexp.String(),
			core_errors.ErrInvalidArgument,
		)
	}

	return nil
}

func ValidatePassword(password string) error {
	passwordLen := len(password)
	if passwordLen < 8 || passwordLen > 72 {
		return fmt.Errorf(
			"invalid 'Password' len: %d, must be between 8 and 72 bytes: %w",
			passwordLen,
			core_errors.ErrInvalidArgument,
		)
	}

	return nil
}
