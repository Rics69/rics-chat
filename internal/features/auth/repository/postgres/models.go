package auth_postgres_repository

import (
	"time"

	"github.com/Rics69/rics-chat/internal/core/domain"
)

type UserModel struct {
	ID           int64
	Login        string
	PasswordHash string
	CreatedAt    time.Time
}

func userDomainFromModel(user UserModel) domain.User {
	return domain.NewUser(
		user.ID,
		user.Login,
		user.PasswordHash,
		user.CreatedAt,
	)
}
