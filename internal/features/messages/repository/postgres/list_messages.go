package messages_postgres_repository

import (
	"context"
	"fmt"

	"github.com/Rics69/rics-chat/internal/core/domain"
)

func (r *MessagesRepository) ListMessages(
	ctx context.Context,
	userID int64,
	peerID int64,
	beforeID *int64,
	limit int,
) ([]domain.Message, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	// COALESCE вместо "$3 IS NULL OR id < $3": так условие остаётся диапазоном по индексу
	query := `
	SELECT id, sender_id, recipient_id, text, created_at
	FROM chat.messages
	WHERE peer_low=LEAST($1::BIGINT, $2::BIGINT)
		AND peer_high=GREATEST($1::BIGINT, $2::BIGINT)
		AND id < COALESCE($3, 9223372036854775807)
	ORDER BY id DESC
	LIMIT $4;
	`

	rows, err := r.pool.Query(ctx, query, userID, peerID, beforeID, limit)
	if err != nil {
		return nil, fmt.Errorf("select messages: %w", err)
	}
	defer rows.Close()

	var messageModels []MessageModel
	for rows.Next() {
		var messageModel MessageModel
		err := rows.Scan(
			&messageModel.ID,
			&messageModel.SenderID,
			&messageModel.RecipientID,
			&messageModel.Text,
			&messageModel.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan messages: %w", err)
		}

		messageModels = append(messageModels, messageModel)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("next rows: %w", err)
	}

	return messageDomainsFromModels(messageModels), nil
}
