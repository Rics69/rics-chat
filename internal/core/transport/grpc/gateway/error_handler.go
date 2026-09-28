package core_grpc_gateway

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	rkgrpc "github.com/rookie-ninja/rk-grpc/v2/boot"
)

// ErrorHandler — обёртка над rk: он копирует в HTTP-заголовки все метаданные gRPC-ответа,
// включая content-type: application/grpc, и в ответе оказывается два Content-Type
func ErrorHandler(
	ctx context.Context,
	mux *runtime.ServeMux,
	marshaler runtime.Marshaler,
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	if md, ok := runtime.ServerMetadataFromContext(ctx); ok {
		delete(md.HeaderMD, "content-type")
	}

	rkgrpc.HttpErrorHandler(ctx, mux, marshaler, w, r, err)
}
