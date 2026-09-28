package users_postgres_repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/Rics69/rics-chat/internal/core/domain"
)

// '_' и '%' в LIKE — шаблоны: без экранирования запрос "a_b" найдёт и "axb"
var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

func (r *UsersRepository) SearchUsers(
	ctx context.Context,
	loginPrefix string,
	excludeUserID int64,
	limit int,
) ([]domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT id, login, created_at
	FROM chat.users
	WHERE login LIKE $1 AND id<>$2
	ORDER BY login ASC
	LIMIT $3;
	`

	pattern := likeEscaper.Replace(loginPrefix) + "%"

	rows, err := r.pool.Query(ctx, query, pattern, excludeUserID, limit)
	if err != nil {
		return nil, fmt.Errorf("select users: %w", err)
	}
	defer rows.Close()

	var usersModels []UserModel
	for rows.Next() {
		var userModel UserModel
		err := rows.Scan(
			&userModel.ID,
			&userModel.Login,
			&userModel.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan users: %w", err)
		}

		usersModels = append(usersModels, userModel)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("next rows: %w", err)
	}

	return userDomainsFromModels(usersModels), nil
}
