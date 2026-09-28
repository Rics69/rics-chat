package messages_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Rics69/rics-chat/internal/core/domain"
	core_errors "github.com/Rics69/rics-chat/internal/core/errors"
	core_postgres_pool "github.com/Rics69/rics-chat/internal/core/repository/postgres/pool"
)

func (r *MessagesRepository) CreateMessage(ctx context.Context, message domain.Message) (domain.Message, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	INSERT INTO chat.messages (sender_id, recipient_id, text, created_at)
	VALUES ($1, $2, $3, $4)
	RETURNING id, sender_id, recipient_id, text, created_at;
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		message.SenderID,
		message.RecipientID,
		message.Text,
		message.CreatedAt,
	)

	var messageModel MessageModel
	err := row.Scan(
		&messageModel.ID,
		&messageModel.SenderID,
		&messageModel.RecipientID,
		&messageModel.Text,
		&messageModel.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrViolatesForeignKey) {
			return domain.Message{}, fmt.Errorf(
				"recipient with id='%d': %w",
				message.RecipientID,
				core_errors.ErrNotFound,
			)
		}

		return domain.Message{}, fmt.Errorf("scan error: %w", err)
	}

	return messageDomainFromModel(messageModel), nil
}
