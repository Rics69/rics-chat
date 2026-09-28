package users_service

import (
	"context"

	"github.com/Rics69/rics-chat/internal/core/domain"
)

type UsersService struct {
	usersRepository UsersRepository
}

type UsersRepository interface {
	GetUser(ctx context.Context, id int64) (domain.User, error)
	SearchUsers(
		ctx context.Context,
		loginPrefix string,
		excludeUserID int64,
		limit int,
	) ([]domain.User, error)
}

func NewUsersService(usersRepository UsersRepository) *UsersService {
	return &UsersService{
		usersRepository: usersRepository,
	}
}
