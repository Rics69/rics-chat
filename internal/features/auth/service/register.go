package auth_service

import (
	"context"
	"fmt"

	"github.com/Rics69/rics-chat/internal/core/domain"
)

func (s *AuthService) Register(
	ctx context.Context,
	login string,
	password string,
) (domain.User, string, error) {
	login = domain.NormalizeLogin(login)

	// валидируем до bcrypt: хэширование специально медленное, не тратим его на мусор
	if err := domain.ValidateLogin(login); err != nil {
		return domain.User{}, "", fmt.Errorf("validate login: %w", err)
	}

	if err := domain.ValidatePassword(password); err != nil {
		return domain.User{}, "", fmt.Errorf("validate password: %w", err)
	}

	passwordHash, err := hashPassword(password)
	if err != nil {
		return domain.User{}, "", fmt.Errorf("hash password: %w", err)
	}

	user := domain.NewUserUnitialized(login, passwordHash)
	if err := user.Validate(); err != nil {
		return domain.User{}, "", fmt.Errorf("validate user domain: %w", err)
	}

	user, err = s.usersRepository.CreateUser(ctx, user)
	if err != nil {
		return domain.User{}, "", fmt.Errorf("create user: %w", err)
	}

	token, err := s.tokenManager.NewToken(user.ID)
	if err != nil {
		return domain.User{}, "", fmt.Errorf("new token: %w", err)
	}

	return user, token, nil
}
