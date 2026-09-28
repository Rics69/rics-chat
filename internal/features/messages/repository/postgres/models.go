package messages_postgres_repository

import (
	"time"

	"github.com/Rics69/rics-chat/internal/core/domain"
)

type MessageModel struct {
	ID          int64
	SenderID    int64
	RecipientID int64
	Text        string
	CreatedAt   time.Time
}

type UserModel struct {
	ID        int64
	Login     string
	CreatedAt time.Time
}

func messageDomainFromModel(message MessageModel) domain.Message {
	return domain.NewMessage(
		message.ID,
		message.SenderID,
		message.RecipientID,
		message.Text,
		message.CreatedAt,
	)
}

func messageDomainsFromModels(messages []MessageModel) []domain.Message {
	messageDomains := make([]domain.Message, len(messages))
	for i, message := range messages {
		messageDomains[i] = messageDomainFromModel(message)
	}

	return messageDomains
}

func userDomainFromModel(user UserModel) domain.User {
	return domain.NewUser(user.ID, user.Login, "", user.CreatedAt)
}
