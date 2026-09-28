package messages_transport_grpc

import (
	"context"

	core_auth "github.com/Rics69/rics-chat/internal/core/auth"
	core_logger "github.com/Rics69/rics-chat/internal/core/logger"
	core_grpc_response "github.com/Rics69/rics-chat/internal/core/transport/grpc/response"
	chatv1 "github.com/Rics69/rics-chat/pkg/api/chat/v1"
)

func (h *MessagesGRPCHandler) ListDialogs(
	ctx context.Context,
	request *chatv1.ListDialogsRequest,
) (*chatv1.ListDialogsResponse, error) {
	log := core_logger.FromContext(ctx)
	responseHandler := core_grpc_response.NewGRPCResponseHandler(log)

	userID, err := core_auth.UserIDFromContext(ctx)
	if err != nil {
		return nil, responseHandler.ErrorResponse(err, "failed to get current user id")
	}

	dialogs, err := h.messagesService.ListDialogs(ctx, userID, int(request.GetLimit()))
	if err != nil {
		return nil, responseHandler.ErrorResponse(err, "failed to list dialogs")
	}

	return &chatv1.ListDialogsResponse{
		Dialogs: dialogsToProto(dialogs),
	}, nil
}
