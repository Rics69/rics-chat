package auth_service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Rics69/rics-chat/internal/core/domain"
	core_errors "github.com/Rics69/rics-chat/internal/core/errors"
)

func (s *AuthService) Login(
	ctx context.Context,
	login string,
	password string,
) (domain.User, string, error) {
	// одна и та же ошибка для "нет логина" и "неверный пароль" — не подсказываем, какие логины существуют
	errInvalidCredentials := fmt.Errorf("invalid login or password: %w", core_errors.ErrUnauthenticated)

	user, err := s.usersRepository.GetUserByLogin(ctx, domain.NormalizeLogin(login))
	if err != nil {
		if errors.Is(err, core_errors.ErrNotFound) {
			// bcrypt всё равно гоняем, иначе по времени ответа видно, что логина нет
			_ = comparePassword(dummyPasswordHash, password)

			return domain.User{}, "", errInvalidCredentials
		}

		return domain.User{}, "", fmt.Errorf("get user by login: %w", err)
	}

	if err := comparePassword(user.PasswordHash, password); err != nil {
		if errors.Is(err, core_errors.ErrUnauthenticated) {
			return domain.User{}, "", errInvalidCredentials
		}

		return domain.User{}, "", fmt.Errorf("compare password: %w", err)
	}

	token, err := s.tokenManager.NewToken(user.ID)
	if err != nil {
		return domain.User{}, "", fmt.Errorf("new token: %w", err)
	}

	return user, token, nil
}
