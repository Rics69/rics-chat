package messages_service

import (
	"context"
	"fmt"

	"github.com/Rics69/rics-chat/internal/core/domain"
)

const (
	defaultDialogsLimit = 50
	maxDialogsLimit     = 100
)

func (s *MessagesService) ListDialogs(ctx context.Context, userID int64, limit int) ([]domain.Dialog, error) {
	limit, err := normalizeLimit(limit, defaultDialogsLimit, maxDialogsLimit)
	if err != nil {
		return nil, fmt.Errorf("normalize limit: %w", err)
	}

	dialogs, err := s.messagesRepository.ListDialogs(ctx, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list dialogs from repository: %w", err)
	}

	return dialogs, nil
}
