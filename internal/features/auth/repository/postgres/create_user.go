package auth_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Rics69/rics-chat/internal/core/domain"
	core_errors "github.com/Rics69/rics-chat/internal/core/errors"
	core_postgres_pool "github.com/Rics69/rics-chat/internal/core/repository/postgres/pool"
)

func (r *AuthRepository) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	INSERT INTO chat.users (login, password_hash, created_at)
	VALUES ($1, $2, $3)
	RETURNING id, login, password_hash, created_at;
	`

	row := r.pool.QueryRow(ctx, query, user.Login, user.PasswordHash, user.CreatedAt)

	var userModel UserModel
	err := row.Scan(
		&userModel.ID,
		&userModel.Login,
		&userModel.PasswordHash,
		&userModel.CreatedAt,
	)

	if err != nil {
		// уникальность проверяет сама БД: SELECT перед INSERT не спасёт от гонки двух регистраций
		if errors.Is(err, core_postgres_pool.ErrViolatesUnique) {
			return domain.User{}, fmt.Errorf(
				"user with login='%s' already exists: %w",
				user.Login,
				core_errors.ErrConflict,
			)
		}

		return domain.User{}, fmt.Errorf("scan error: %w", err)
	}

	return userDomainFromModel(userModel), nil
}
