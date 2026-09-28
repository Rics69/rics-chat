package auth_transport_grpc

import (
	"context"

	chatv1 "github.com/Rics69/rics-chat/pkg/api/chat/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type AuthGRPCHandler struct {
	// заглушка: все методы отвечают codes.Unimplemented, пока мы их не реализуем.
	// Встраивается по значению — этого требует protoc-gen-go-grpc (иначе паника при регистрации)
	chatv1.UnimplementedAuthServiceServer
}

func NewAuthGRPCHandler() *AuthGRPCHandler {
	return &AuthGRPCHandler{}
}

// RegisterGRPC — аналог Routes() из http-транспорта: регистрируем сервис на gRPC-сервере
func (h *AuthGRPCHandler) RegisterGRPC(server *grpc.Server) {
	chatv1.RegisterAuthServiceServer(server, h)
}

// RegisterGateway — REST-ручки из option (google.api.http) в proto.
// gateway сам ходит в наш же gRPC-сервер по endpoint, как обычный клиент
func (h *AuthGRPCHandler) RegisterGateway(
	ctx context.Context,
	mux *runtime.ServeMux,
	endpoint string,
	opts []grpc.DialOption,
) error {
	return chatv1.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)
}
