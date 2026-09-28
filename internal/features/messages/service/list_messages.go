package messages_service

import (
	"context"
	"fmt"

	"github.com/Rics69/rics-chat/internal/core/domain"
)

const (
	defaultMessagesLimit = 50
	maxMessagesLimit     = 100
)

func (s *MessagesService) ListMessages(
	ctx context.Context,
	userID int64,
	peerID int64,
	beforeID *int64,
	limit int,
) ([]domain.Message, *int64, error) {
	limit, err := normalizeLimit(limit, defaultMessagesLimit, maxMessagesLimit)
	if err != nil {
		return nil, nil, fmt.Errorf("normalize limit: %w", err)
	}

	// просим на одно больше: если пришло limit+1, значит есть ещё страница
	messages, err := s.messagesRepository.ListMessages(ctx, userID, peerID, beforeID, limit+1)
	if err != nil {
		return nil, nil, fmt.Errorf("list messages from repository: %w", err)
	}

	if len(messages) <= limit {
		return messages, nil, nil
	}

	messages = messages[:limit]
	nextBeforeID := messages[len(messages)-1].ID

	return messages, &nextBeforeID, nil
}
