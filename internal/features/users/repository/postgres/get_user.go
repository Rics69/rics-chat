package users_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Rics69/rics-chat/internal/core/domain"
	core_errors "github.com/Rics69/rics-chat/internal/core/errors"
	core_postgres_pool "github.com/Rics69/rics-chat/internal/core/repository/postgres/pool"
)

func (r *UsersRepository) GetUser(ctx context.Context, id int64) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT id, login, created_at
	FROM chat.users
	WHERE id=$1;
	`

	row := r.pool.QueryRow(ctx, query, id)

	var userModel UserModel
	err := row.Scan(
		&userModel.ID,
		&userModel.Login,
		&userModel.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return domain.User{}, fmt.Errorf(
				"user with id='%d': %w",
				id,
				core_errors.ErrNotFound,
			)
		}

		return domain.User{}, fmt.Errorf("scan error: %w", err)
	}

	return userDomainFromModel(userModel), nil
}
