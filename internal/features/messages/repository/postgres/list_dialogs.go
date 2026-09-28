package messages_postgres_repository

import (
	"context"
	"fmt"

	"github.com/Rics69/rics-chat/internal/core/domain"
)

func (r *MessagesRepository) ListDialogs(ctx context.Context, userID int64, limit int) ([]domain.Dialog, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	// DISTINCT ON оставляет первую строку в каждой группе (peer_low, peer_high),
	// а ORDER BY ... id DESC делает первой самое свежее сообщение диалога
	query := `
	SELECT
		u.id, u.login, u.created_at,
		m.id, m.sender_id, m.recipient_id, m.text, m.created_at
	FROM (
		SELECT DISTINCT ON (peer_low, peer_high)
			id, sender_id, recipient_id, text, created_at
		FROM chat.messages
		WHERE sender_id=$1 OR recipient_id=$1
		ORDER BY peer_low, peer_high, id DESC
	) m
	JOIN chat.users u
		ON u.id = CASE WHEN m.sender_id=$1 THEN m.recipient_id ELSE m.sender_id END
	ORDER BY m.id DESC
	LIMIT $2;
	`

	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("select dialogs: %w", err)
	}
	defer rows.Close()

	var dialogs []domain.Dialog
	for rows.Next() {
		var (
			userModel    UserModel
			messageModel MessageModel
		)

		err := rows.Scan(
			&userModel.ID,
			&userModel.Login,
			&userModel.CreatedAt,
			&messageModel.ID,
			&messageModel.SenderID,
			&messageModel.RecipientID,
			&messageModel.Text,
			&messageModel.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan dialogs: %w", err)
		}

		dialogs = append(dialogs, domain.NewDialog(
			userDomainFromModel(userModel),
			messageDomainFromModel(messageModel),
		))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("next rows: %w", err)
	}

	return dialogs, nil
}
