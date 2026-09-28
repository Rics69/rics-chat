package auth_transport_grpc

import (
	"github.com/Rics69/rics-chat/internal/core/domain"
	chatv1 "github.com/Rics69/rics-chat/pkg/api/chat/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func userToProto(user domain.User) *chatv1.User {
	return &chatv1.User{
		Id:        user.ID,
		Login:     user.Login,
		CreatedAt: timestamppb.New(user.CreatedAt),
	}
}
