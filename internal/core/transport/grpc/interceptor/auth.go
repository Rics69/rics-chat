package core_grpc_interceptor

import (
	"context"
	"fmt"
	"strings"

	core_auth "github.com/Rics69/rics-chat/internal/core/auth"
	core_errors "github.com/Rics69/rics-chat/internal/core/errors"
	core_logger "github.com/Rics69/rics-chat/internal/core/logger"
	core_grpc_response "github.com/Rics69/rics-chat/internal/core/transport/grpc/response"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const authorizationMetadataKey = "authorization"

type TokenParser interface {
	ParseToken(token string) (int64, error)
}

func Auth(tokenParser TokenParser, publicMethods ...string) grpc.UnaryServerInterceptor {
	public := make(map[string]struct{}, len(publicMethods))
	for _, method := range publicMethods {
		public[method] = struct{}{}
	}

	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if _, ok := public[info.FullMethod]; ok {
			return handler(ctx, req)
		}

		log := core_logger.FromContext(ctx)
		responseHandler := core_grpc_response.NewGRPCResponseHandler(log)

		md, _ := metadata.FromIncomingContext(ctx)

		token, err := bearerToken(firstValue(md, authorizationMetadataKey))
		if err != nil {
			return nil, responseHandler.ErrorResponse(err, "failed to get bearer token")
		}

		userID, err := tokenParser.ParseToken(token)
		if err != nil {
			return nil, responseHandler.ErrorResponse(err, "failed to parse token")
		}

		ctx = core_auth.ToContext(ctx, userID)
		ctx = core_logger.ToContext(ctx, log.With(zap.Int64("user_id", userID)))

		return handler(ctx, req)
	}
}

func bearerToken(header string) (string, error) {
	const prefix = "Bearer "

	if len(header) < len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", fmt.Errorf("missing 'Bearer' authorization: %w", core_errors.ErrUnauthenticated)
	}

	token := strings.TrimSpace(header[len(prefix):])
	if token == "" {
		return "", fmt.Errorf("empty bearer token: %w", core_errors.ErrUnauthenticated)
	}

	return token, nil
}
