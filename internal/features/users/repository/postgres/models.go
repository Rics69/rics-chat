package users_postgres_repository

import (
	"time"

	"github.com/Rics69/rics-chat/internal/core/domain"
)

// password_hash тут не выбираем: этой фиче он не нужен
type UserModel struct {
	ID        int64
	Login     string
	CreatedAt time.Time
}

func userDomainFromModel(user UserModel) domain.User {
	return domain.NewUser(user.ID, user.Login, "", user.CreatedAt)
}

func userDomainsFromModels(users []UserModel) []domain.User {
	userDomains := make([]domain.User, len(users))
	for i, user := range users {
		userDomains[i] = userDomainFromModel(user)
	}

	return userDomains
}
