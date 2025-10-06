package app

import (
	"context"
	"fmt"
	"github.com/PianyCoder/test_file_service/infrastructure/logger"
	"github.com/PianyCoder/test_file_service/infrastructure/minio_cli"
	"github.com/PianyCoder/test_file_service/internal/config"
	"github.com/PianyCoder/test_file_service/internal/controller"
	"github.com/PianyCoder/test_file_service/internal/server"
	"github.com/PianyCoder/test_file_service/internal/service"
	"github.com/PianyCoder/test_file_service/internal/storage"
	"go.uber.org/zap"
)

func Start(ctx context.Context) error {
	zapLog, err := logger.Init()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	ctx = logger.WithContext(ctx, zapLog)
	l := logger.FromContext(ctx)

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	minioCli, err := minio_cli.NewClient(ctx, *cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize minio: %w", err)
	}

	stg := storage.NewMinioStorage(minioCli, cfg.MinioConfig.BucketName)

	cfgS := cfg.ServiceConfig
	svc := service.NewFileService(stg, cfgS.UploadLimit, cfgS.DownloadLimit, cfgS.ListLimit, cfgS.ChunkSize)

	ctlr := controller.NewFileServiceHandler(svc)

	grpcServer, err := server.NewGRPCServer(cfg.ServerConfig.Addr, ctlr)
	if err != nil {
		l.Error("failed to create gRPC server", zap.Error(err))
		return fmt.Errorf("%w", err)
	}

	if err := grpcServer.StartServer(ctx); err != nil {
		l.Error("failed to start gRPC server", zap.Error(err))
		return fmt.Errorf("%w", err)
	}

	l.Info("Server stopped gracefully")

	return nil
}
