package messages_transport_grpc

import (
	"context"

	core_auth "github.com/Rics69/rics-chat/internal/core/auth"
	core_logger "github.com/Rics69/rics-chat/internal/core/logger"
	core_grpc_response "github.com/Rics69/rics-chat/internal/core/transport/grpc/response"
	chatv1 "github.com/Rics69/rics-chat/pkg/api/chat/v1"
)

func (h *MessagesGRPCHandler) ListMessages(
	ctx context.Context,
	request *chatv1.ListMessagesRequest,
) (*chatv1.ListMessagesResponse, error) {
	log := core_logger.FromContext(ctx)
	responseHandler := core_grpc_response.NewGRPCResponseHandler(log)

	userID, err := core_auth.UserIDFromContext(ctx)
	if err != nil {
		return nil, responseHandler.ErrorResponse(err, "failed to get current user id")
	}

	messages, nextBeforeID, err := h.messagesService.ListMessages(
		ctx,
		userID,
		request.GetPeerId(),
		request.BeforeId,
		int(request.GetLimit()),
	)
	if err != nil {
		return nil, responseHandler.ErrorResponse(err, "failed to list messages")
	}

	return &chatv1.ListMessagesResponse{
		Messages:     messagesToProto(messages),
		NextBeforeId: nextBeforeID,
	}, nil
}
