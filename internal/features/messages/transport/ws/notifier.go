package messages_transport_ws

import (
	"context"

	"github.com/Rics69/rics-chat/internal/core/domain"
	core_logger "github.com/Rics69/rics-chat/internal/core/logger"
	chatv1 "github.com/Rics69/rics-chat/pkg/api/chat/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Hub interface {
	SendToUser(userID int64, payload []byte)
}

type MessagesWSNotifier struct {
	hub Hub
}

func NewMessagesWSNotifier(hub Hub) *MessagesWSNotifier {
	return &MessagesWSNotifier{
		hub: hub,
	}
}

// тот же protojson с proto-именами, что и у gateway: json в сокете совпадает с REST
var marshalOptions = protojson.MarshalOptions{UseProtoNames: true}

func (n *MessagesWSNotifier) NotifyNewMessage(ctx context.Context, message domain.Message) {
	log := core_logger.FromContext(ctx)

	event := &chatv1.ServerEvent{
		Event: &chatv1.ServerEvent_NewMessage{
			NewMessage: messageToProto(message),
		},
	}

	payload, err := marshalOptions.Marshal(event)
	if err != nil {
		log.Error("marshal new message event", zap.Error(err))
		return
	}

	// отправителю тоже: у него могут быть открыты другие вкладки
	n.hub.SendToUser(message.RecipientID, payload)
	n.hub.SendToUser(message.SenderID, payload)
}

func messageToProto(message domain.Message) *chatv1.Message {
	return &chatv1.Message{
		Id:          message.ID,
		SenderId:    message.SenderID,
		RecipientId: message.RecipientID,
		Text:        message.Text,
		CreatedAt:   timestamppb.New(message.CreatedAt),
	}
}
