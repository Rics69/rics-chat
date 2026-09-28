package messages_transport_grpc

import (
	"context"

	"github.com/Rics69/rics-chat/internal/core/domain"
	chatv1 "github.com/Rics69/rics-chat/pkg/api/chat/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type MessagesGRPCHandler struct {
	chatv1.UnimplementedMessagesServiceServer

	messagesService MessagesService
}

type MessagesService interface {
	SendMessage(
		ctx context.Context,
		senderID int64,
		recipientID int64,
		text string,
	) (domain.Message, error)
	ListMessages(
		ctx context.Context,
		userID int64,
		peerID int64,
		beforeID *int64,
		limit int,
	) ([]domain.Message, *int64, error)
	ListDialogs(
		ctx context.Context,
		userID int64,
		limit int,
	) ([]domain.Dialog, error)
}

func NewMessagesGRPCHandler(messagesService MessagesService) *MessagesGRPCHandler {
	return &MessagesGRPCHandler{
		messagesService: messagesService,
	}
}

func (h *MessagesGRPCHandler) RegisterGRPC(server *grpc.Server) {
	chatv1.RegisterMessagesServiceServer(server, h)
}

func (h *MessagesGRPCHandler) RegisterGateway(
	ctx context.Context,
	mux *runtime.ServeMux,
	endpoint string,
	opts []grpc.DialOption,
) error {
	return chatv1.RegisterMessagesServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)
}
