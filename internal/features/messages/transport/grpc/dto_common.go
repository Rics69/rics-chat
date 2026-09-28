package messages_transport_grpc

import (
	"github.com/Rics69/rics-chat/internal/core/domain"
	chatv1 "github.com/Rics69/rics-chat/pkg/api/chat/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func messageToProto(message domain.Message) *chatv1.Message {
	return &chatv1.Message{
		Id:          message.ID,
		SenderId:    message.SenderID,
		RecipientId: message.RecipientID,
		Text:        message.Text,
		CreatedAt:   timestamppb.New(message.CreatedAt),
	}
}

func messagesToProto(messages []domain.Message) []*chatv1.Message {
	messagesProto := make([]*chatv1.Message, len(messages))
	for i, message := range messages {
		messagesProto[i] = messageToProto(message)
	}

	return messagesProto
}

func userToProto(user domain.User) *chatv1.User {
	return &chatv1.User{
		Id:        user.ID,
		Login:     user.Login,
		CreatedAt: timestamppb.New(user.CreatedAt),
	}
}

func dialogsToProto(dialogs []domain.Dialog) []*chatv1.Dialog {
	dialogsProto := make([]*chatv1.Dialog, len(dialogs))
	for i, dialog := range dialogs {
		dialogsProto[i] = &chatv1.Dialog{
			Peer:        userToProto(dialog.Peer),
			LastMessage: messageToProto(dialog.LastMessage),
		}
	}

	return dialogsProto
}
