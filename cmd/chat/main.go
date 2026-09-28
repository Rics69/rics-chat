package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"time"

	"github.com/Rics69/rics-chat/api/openapi"
	core_config "github.com/Rics69/rics-chat/internal/core/config"
	core_logger "github.com/Rics69/rics-chat/internal/core/logger"
	core_pgx_pool "github.com/Rics69/rics-chat/internal/core/repository/postgres/pool/pgx"
	core_grpc_interceptor "github.com/Rics69/rics-chat/internal/core/transport/grpc/interceptor"
	auth_transport_grpc "github.com/Rics69/rics-chat/internal/features/auth/transport/grpc"
	rkboot "github.com/rookie-ninja/rk-boot/v2"
	rkentry "github.com/rookie-ninja/rk-entry/v2/entry"
	rkgrpc "github.com/rookie-ninja/rk-grpc/v2/boot"
	"go.uber.org/zap"
)

// имя должно совпадать с grpc[].name в boot.yaml
const grpcEntryName = "rics-chat"

// boot.yaml вшиваем в бинарь — не нужно таскать его рядом с exe в docker
//
//go:embed boot.yaml
var bootConfig []byte

func main() {
	cfg := core_config.NewConfigMust()
	time.Local = cfg.TimeZone

	// SIGINT/SIGTERM ловит сам rk-boot в WaitForShutdownSig,
	// поэтому signal.NotifyContext не нужен
	ctx := context.Background()

	logger, err := core_logger.NewLogger(core_logger.NewConfigMust())
	if err != nil {
		fmt.Println("failed to init application logger:", err)
		os.Exit(1)
	}

	defer logger.Close()

	logger.Debug("application time zone", zap.Any("zone", time.Local))

	logger.Debug("initializing postgres connection pool")
	pool, err := core_pgx_pool.NewPool(ctx, core_pgx_pool.NewConfigMust())
	if err != nil {
		logger.Fatal("failed to init postgres connection pool", zap.Error(err))
	}

	defer pool.Close()

	logger.Debug("initializing feature", zap.String("feature", "auth"))

	authTransportGRPC := auth_transport_grpc.NewAuthGRPCHandler()

	logger.Debug("initializing rk-boot")

	// embed FS нужно отдать rk до NewBoot — entries создаются прямо в нём
	rkentry.GlobalAppCtx.AddEmbedFS(rkentry.SWEntryType, grpcEntryName, &openapi.SwaggerFS)

	boot := rkboot.NewBoot(rkboot.WithBootConfigRaw(bootConfig))

	grpcEntry := rkgrpc.GetGrpcEntry(grpcEntryName)
	if grpcEntry == nil {
		logger.Fatal("grpc entry not found in boot.yaml", zap.String("name", grpcEntryName))
	}

	grpcEntry.AddUnaryInterceptors(
		core_grpc_interceptor.RequestID(),
		core_grpc_interceptor.Logger(logger),
		core_grpc_interceptor.Trace(),
		core_grpc_interceptor.Panic(),
	)

	grpcEntry.AddRegFuncGrpc(authTransportGRPC.RegisterGRPC)
	grpcEntry.AddRegFuncGw(authTransportGRPC.RegisterGateway)

	boot.Bootstrap(ctx)

	// блокируется до сигнала, потом гасит gRPC/HTTP сервер.
	// defer'ы (pool.Close, logger.Close) выполнятся уже после остановки сервера
	boot.WaitForShutdownSig(ctx)
}
