package core_grpc_response

import (
	"errors"
	"fmt"

	core_errors "github.com/Rics69/rics-chat/internal/core/errors"
	core_logger "github.com/Rics69/rics-chat/internal/core/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCResponseHandler struct {
	log *core_logger.Logger
}

func NewGRPCResponseHandler(log *core_logger.Logger) *GRPCResponseHandler {
	return &GRPCResponseHandler{
		log: log,
	}
}

func (h *GRPCResponseHandler) ErrorResponse(err error, msg string) error {
	var (
		code    codes.Code
		logFunc func(string, ...zap.Field)
	)

	switch {
	case errors.Is(err, core_errors.ErrInvalidArgument):
		code = codes.InvalidArgument
		logFunc = h.log.Warn
	case errors.Is(err, core_errors.ErrNotFound):
		code = codes.NotFound
		logFunc = h.log.Debug
	case errors.Is(err, core_errors.ErrConflict):
		code = codes.AlreadyExists
		logFunc = h.log.Warn
	case errors.Is(err, core_errors.ErrUnauthenticated):
		code = codes.Unauthenticated
		logFunc = h.log.Debug
	default:
		code = codes.Internal
		logFunc = h.log.Error
	}

	logFunc(msg, zap.Error(err))

	return h.errorResponse(code, err, msg)
}

func (h *GRPCResponseHandler) PanicResponse(p any, msg string) error {
	err := fmt.Errorf("unexpected panic: %v", p)

	h.log.Error(msg, zap.Error(err))

	return h.errorResponse(codes.Internal, err, msg)
}

func (h *GRPCResponseHandler) errorResponse(code codes.Code, err error, msg string) error {
	if code == codes.Internal {
		return status.Error(code, msg)
	}

	return status.Errorf(code, "%s: %v", msg, err)
}
