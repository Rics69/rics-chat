package users_transport_grpc

import (
	"context"

	core_auth "github.com/Rics69/rics-chat/internal/core/auth"
	core_logger "github.com/Rics69/rics-chat/internal/core/logger"
	core_grpc_response "github.com/Rics69/rics-chat/internal/core/transport/grpc/response"
	chatv1 "github.com/Rics69/rics-chat/pkg/api/chat/v1"
)

func (h *UsersGRPCHandler) SearchUsers(
	ctx context.Context,
	request *chatv1.SearchUsersRequest,
) (*chatv1.SearchUsersResponse, error) {
	log := core_logger.FromContext(ctx)
	responseHandler := core_grpc_response.NewGRPCResponseHandler(log)

	userID, err := core_auth.UserIDFromContext(ctx)
	if err != nil {
		return nil, responseHandler.ErrorResponse(err, "failed to get current user id")
	}

	users, err := h.usersService.SearchUsers(
		ctx,
		userID,
		request.GetQuery(),
		int(request.GetLimit()),
	)
	if err != nil {
		return nil, responseHandler.ErrorResponse(err, "failed to search users")
	}

	return &chatv1.SearchUsersResponse{
		Users: usersToProto(users),
	}, nil
}
