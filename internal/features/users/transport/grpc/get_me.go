package users_transport_grpc

import (
	"context"

	core_auth "github.com/Rics69/rics-chat/internal/core/auth"
	core_logger "github.com/Rics69/rics-chat/internal/core/logger"
	core_grpc_response "github.com/Rics69/rics-chat/internal/core/transport/grpc/response"
	chatv1 "github.com/Rics69/rics-chat/pkg/api/chat/v1"
)

func (h *UsersGRPCHandler) GetMe(
	ctx context.Context,
	_ *chatv1.GetMeRequest,
) (*chatv1.GetMeResponse, error) {
	log := core_logger.FromContext(ctx)
	responseHandler := core_grpc_response.NewGRPCResponseHandler(log)

	userID, err := core_auth.UserIDFromContext(ctx)
	if err != nil {
		return nil, responseHandler.ErrorResponse(err, "failed to get current user id")
	}

	user, err := h.usersService.GetUser(ctx, userID)
	if err != nil {
		return nil, responseHandler.ErrorResponse(err, "failed to get current user")
	}

	return &chatv1.GetMeResponse{
		User: userToProto(user),
	}, nil
}
