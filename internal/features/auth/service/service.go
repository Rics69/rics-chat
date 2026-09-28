package auth_service

import (
	"context"

	"github.com/Rics69/rics-chat/internal/core/domain"
)

type AuthService struct {
	usersRepository UsersRepository
	tokenManager    TokenManager
}

type UsersRepository interface {
	CreateUser(ctx context.Context, user domain.User) (domain.User, error)
	GetUserByLogin(ctx context.Context, login string) (domain.User, error)
}

type TokenManager interface {
	NewToken(userID int64) (string, error)
}

func NewAuthService(usersRepository UsersRepository, tokenManager TokenManager) *AuthService {
	return &AuthService{
		usersRepository: usersRepository,
		tokenManager:    tokenManager,
	}
}
