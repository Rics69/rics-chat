package users_transport_grpc

import (
	"context"

	"github.com/Rics69/rics-chat/internal/core/domain"
	chatv1 "github.com/Rics69/rics-chat/pkg/api/chat/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type UsersGRPCHandler struct {
	chatv1.UnimplementedUsersServiceServer

	usersService UsersService
}

type UsersService interface {
	GetUser(
		ctx context.Context,
		id int64,
	) (domain.User, error)
	SearchUsers(
		ctx context.Context,
		currentUserID int64,
		query string,
		limit int,
	) ([]domain.User, error)
}

func NewUsersGRPCHandler(usersService UsersService) *UsersGRPCHandler {
	return &UsersGRPCHandler{
		usersService: usersService,
	}
}

func (h *UsersGRPCHandler) RegisterGRPC(server *grpc.Server) {
	chatv1.RegisterUsersServiceServer(server, h)
}

func (h *UsersGRPCHandler) RegisterGateway(
	ctx context.Context,
	mux *runtime.ServeMux,
	endpoint string,
	opts []grpc.DialOption,
) error {
	return chatv1.RegisterUsersServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)
}
