package messages_service

import (
	"context"
	"fmt"

	"github.com/Rics69/rics-chat/internal/core/domain"
	core_errors "github.com/Rics69/rics-chat/internal/core/errors"
)

type MessagesService struct {
	messagesRepository MessagesRepository
}

type MessagesRepository interface {
	CreateMessage(ctx context.Context, message domain.Message) (domain.Message, error)
	ListMessages(
		ctx context.Context,
		userID int64,
		peerID int64,
		beforeID *int64,
		limit int,
	) ([]domain.Message, error)
	ListDialogs(ctx context.Context, userID int64, limit int) ([]domain.Dialog, error)
}

func NewMessagesService(messagesRepository MessagesRepository) *MessagesService {
	return &MessagesService{
		messagesRepository: messagesRepository,
	}
}

func normalizeLimit(limit int, defaultLimit int, maxLimit int) (int, error) {
	if limit < 0 {
		return 0, fmt.Errorf("limit must be non-negative: %w", core_errors.ErrInvalidArgument)
	}

	if limit == 0 {
		return defaultLimit, nil
	}

	return min(limit, maxLimit), nil
}
