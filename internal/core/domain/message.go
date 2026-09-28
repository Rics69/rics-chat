package domain

import (
	"fmt"
	"strings"
	"time"

	core_errors "github.com/Rics69/rics-chat/internal/core/errors"
)

type Message struct {
	ID int64

	SenderID    int64
	RecipientID int64
	Text        string
	CreatedAt   time.Time
}

func NewMessage(
	id int64,
	senderID int64,
	recipientID int64,
	text string,
	createdAt time.Time,
) Message {
	return Message{
		ID:          id,
		SenderID:    senderID,
		RecipientID: recipientID,
		Text:        text,
		CreatedAt:   createdAt,
	}
}

func NewMessageUnitialized(senderID int64, recipientID int64, text string) Message {
	return NewMessage(UnitializedID, senderID, recipientID, text, time.Now())
}

func (m *Message) Validate() error {
	if strings.TrimSpace(m.Text) == "" {
		return fmt.Errorf("'Text' can't be empty: %w", core_errors.ErrInvalidArgument)
	}

	textLen := len([]rune(m.Text))
	if textLen > 4096 {
		return fmt.Errorf("invalid 'Text' len: %d: %w", textLen, core_errors.ErrInvalidArgument)
	}

	if m.SenderID == m.RecipientID {
		return fmt.Errorf("can't send message to yourself: %w", core_errors.ErrInvalidArgument)
	}

	return nil
}

func (m *Message) PeerID(userID int64) int64 {
	if m.SenderID == userID {
		return m.RecipientID
	}

	return m.SenderID
}

type Dialog struct {
	Peer        User
	LastMessage Message
}

func NewDialog(peer User, lastMessage Message) Dialog {
	return Dialog{
		Peer:        peer,
		LastMessage: lastMessage,
	}
}
