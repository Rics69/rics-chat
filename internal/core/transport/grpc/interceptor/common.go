package core_grpc_interceptor

import (
	"context"
	"time"

	core_logger "github.com/Rics69/rics-chat/internal/core/logger"
	core_grpc_response "github.com/Rics69/rics-chat/internal/core/transport/grpc/response"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ключи metadata в gRPC всегда в нижнем регистре,
// gateway (с enableRkGwOption) прокидывает сюда HTTP-заголовок X-Request-ID
const requestIDMetadataKey = "x-request-id"

// Порядок в ChainUnaryInterceptor такой же, как у http-middleware:
// первый в списке — самый внешний.

func RequestID() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		// metadata — это map, копируем, чтобы не менять чужую
		md = md.Copy()

		requestID := firstValue(md, requestIDMetadataKey)
		if requestID == "" {
			requestID = uuid.NewString()
		}

		md.Set(requestIDMetadataKey, requestID)
		ctx = metadata.NewIncomingContext(ctx, md)

		// отдаём id клиенту в заголовках ответа (аналог w.Header().Set)
		_ = grpc.SetHeader(ctx, metadata.Pairs(requestIDMetadataKey, requestID))

		return handler(ctx, req)
	}
}

func Logger(log *core_logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		md, _ := metadata.FromIncomingContext(ctx)

		l := log.With(
			zap.String("request_id", firstValue(md, requestIDMetadataKey)),
			zap.String("grpc_method", info.FullMethod),
		)

		ctx = core_logger.ToContext(ctx, l)

		return handler(ctx, req)
	}
}

func Trace() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		log := core_logger.FromContext(ctx)

		before := time.Now()

		log.Debug(">>> incoming gRPC request", zap.Time("time", before.UTC()))

		resp, err := handler(ctx, req)

		// status.Code(nil) == codes.OK
		log.Debug(
			"<<< done gRPC request",
			zap.String("status_code", status.Code(err).String()),
			zap.Duration("latency", time.Since(before)),
		)

		return resp, err
	}
}

func Panic() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		log := core_logger.FromContext(ctx)
		responseHandler := core_grpc_response.NewGRPCResponseHandler(log)

		defer func() {
			if p := recover(); p != nil {
				resp = nil
				err = responseHandler.PanicResponse(p, "during handle gRPC request got unexpected panic")
			}
		}()

		return handler(ctx, req)
	}
}

func firstValue(md metadata.MD, key string) string {
	values := md.Get(key)
	if len(values) == 0 {
		return ""
	}

	return values[0]
}
