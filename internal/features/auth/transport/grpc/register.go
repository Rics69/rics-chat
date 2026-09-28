package auth_transport_grpc

import (
	"context"

	core_logger "github.com/Rics69/rics-chat/internal/core/logger"
	core_grpc_response "github.com/Rics69/rics-chat/internal/core/transport/grpc/response"
	chatv1 "github.com/Rics69/rics-chat/pkg/api/chat/v1"
)

func (h *AuthGRPCHandler) Register(
	ctx context.Context,
	request *chatv1.RegisterRequest,
) (*chatv1.RegisterResponse, error) {
	log := core_logger.FromContext(ctx)
	responseHandler := core_grpc_response.NewGRPCResponseHandler(log)

	user, token, err := h.authService.Register(ctx, request.GetLogin(), request.GetPassword())
	if err != nil {
		return nil, responseHandler.ErrorResponse(err, "failed to register user")
	}

	return &chatv1.RegisterResponse{
		User:        userToProto(user),
		AccessToken: token,
	}, nil
}
