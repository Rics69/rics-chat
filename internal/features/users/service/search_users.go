package users_service

import (
	"context"
	"fmt"
	"regexp"

	"github.com/Rics69/rics-chat/internal/core/domain"
	core_errors "github.com/Rics69/rics-chat/internal/core/errors"
)

const (
	defaultSearchLimit = 20
	maxSearchLimit     = 50
)

var searchQueryRegexp = regexp.MustCompile(`^[a-z0-9_]{1,32}$`)

func (s *UsersService) SearchUsers(
	ctx context.Context,
	currentUserID int64,
	query string,
	limit int,
) ([]domain.User, error) {
	query = domain.NormalizeLogin(query)

	if !searchQueryRegexp.MatchString(query) {
		return nil, fmt.Errorf(
			"invalid search query='%s': must match %s: %w",
			query,
			searchQueryRegexp.String(),
			core_errors.ErrInvalidArgument,
		)
	}

	if limit < 0 {
		return nil, fmt.Errorf("limit must be non-negative: %w", core_errors.ErrInvalidArgument)
	}

	if limit == 0 {
		limit = defaultSearchLimit
	}

	limit = min(limit, maxSearchLimit)

	users, err := s.usersRepository.SearchUsers(ctx, query, currentUserID, limit)
	if err != nil {
		return nil, fmt.Errorf("search users in repository: %w", err)
	}

	return users, nil
}
