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
)

func Start(ctx context.Context) error {
	zapLog, err := logger.Init()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer func() { _ = zapLog.Sync() }()

	ctx = logger.WithContext(ctx, zapLog)
	l := logger.FromContext(ctx)

	l.Info("starting application")

	cfg, err := config.Load()
	if err != nil {
		l.Errorw("failed to load config", "error", err)
		return fmt.Errorf("failed to load config: %w", err)
	}
	l.Infow("config loaded", "server_addr", cfg.ServerConfig.Addr, "minio_bucket", cfg.MinioConfig.BucketName)

	minioCli, err := minio_cli.NewClient(ctx, cfg)
	if err != nil {
		l.Errorw("failed to initialize minio", "error", err)
		return fmt.Errorf("failed to initialize minio: %w", err)
	}
	l.Info("minio client initialized")

	stg := storage.NewMinioStorage(minioCli, cfg.MinioConfig.BucketName)
	l.Infow("storage initialized", "bucket", cfg.MinioConfig.BucketName)

	cfgS := cfg.ServiceConfig
	svc := service.NewFileService(stg, cfgS.UploadLimit, cfgS.DownloadLimit, cfgS.ListLimit, cfgS.ChunkSize)
	l.Infow("service initialized", "chunk_size", cfgS.ChunkSize)

	ctlr := controller.NewFileServiceHandler(svc)

	grpcServer, err := server.NewGRPCServer(cfg.ServerConfig.Addr, ctlr, zapLog)
	if err != nil {
		l.Errorw("failed to create gRPC server", "error", err)
		return fmt.Errorf("failed to create grpc server: %w", err)
	}
	l.Infow("gRPC server created", "addr", cfg.ServerConfig.Addr)

	if err := grpcServer.StartServer(ctx); err != nil {
		l.Errorw("grpc server stopped with error", "error", err)
		return fmt.Errorf("grpc server stopped: %w", err)
	}

	l.Info("Server stopped gracefully")
	return nil
}
