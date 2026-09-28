package messages_service

import (
	"context"
	"fmt"

	"github.com/Rics69/rics-chat/internal/core/domain"
)

func (s *MessagesService) SendMessage(
	ctx context.Context,
	senderID int64,
	recipientID int64,
	text string,
) (domain.Message, error) {
	message := domain.NewMessageUnitialized(senderID, recipientID, text)
	if err := message.Validate(); err != nil {
		return domain.Message{}, fmt.Errorf("validate message domain: %w", err)
	}

	message, err := s.messagesRepository.CreateMessage(ctx, message)
	if err != nil {
		return domain.Message{}, fmt.Errorf("create message: %w", err)
	}

	return message, nil
}
