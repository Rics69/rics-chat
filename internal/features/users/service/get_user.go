package users_service

import (
	"context"
	"fmt"

	"github.com/Rics69/rics-chat/internal/core/domain"
)

func (s *UsersService) GetUser(ctx context.Context, id int64) (domain.User, error) {
	user, err := s.usersRepository.GetUser(ctx, id)
	if err != nil {
		return domain.User{}, fmt.Errorf("get user from repository: %w", err)
	}

	return user, nil
}
