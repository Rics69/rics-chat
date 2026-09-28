package core_auth

import (
	"context"
	"fmt"

	core_errors "github.com/Rics69/rics-chat/internal/core/errors"
)

type userIDContextKey struct{}

var (
	key = userIDContextKey{}
)

func ToContext(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, key, userID)
}

func UserIDFromContext(ctx context.Context) (int64, error) {
	userID, ok := ctx.Value(key).(int64)
	if !ok {
		return 0, fmt.Errorf("no user id in context: %w", core_errors.ErrUnauthenticated)
	}

	return userID, nil
}
