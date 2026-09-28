package auth_transport_grpc

import (
	"context"

	"github.com/Rics69/rics-chat/internal/core/domain"
	chatv1 "github.com/Rics69/rics-chat/pkg/api/chat/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type AuthGRPCHandler struct {
	chatv1.UnimplementedAuthServiceServer

	authService AuthService
}

type AuthService interface {
	Register(
		ctx context.Context,
		login string,
		password string,
	) (domain.User, string, error)
	Login(
		ctx context.Context,
		login string,
		password string,
	) (domain.User, string, error)
}

func NewAuthGRPCHandler(authService AuthService) *AuthGRPCHandler {
	return &AuthGRPCHandler{
		authService: authService,
	}
}

func (h *AuthGRPCHandler) RegisterGRPC(server *grpc.Server) {
	chatv1.RegisterAuthServiceServer(server, h)
}

func (h *AuthGRPCHandler) RegisterGateway(
	ctx context.Context,
	mux *runtime.ServeMux,
	endpoint string,
	opts []grpc.DialOption,
) error {
	return chatv1.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)
}
