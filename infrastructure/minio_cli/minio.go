package minio_cli

import (
	"context"
	"fmt"
	"github.com/PianyCoder/test_file_service/infrastructure/logger"
	"github.com/PianyCoder/test_file_service/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"net"
	"strconv"
)

func NewClient(ctx context.Context, cfg *config.Config) (*minio.Client, error) {
	l := logger.FromContext(ctx)
	ep := cfg.MinioConfig.Endpoint

	if host, portStr, err := net.SplitHostPort(ep); err == nil && host != "" && portStr != "" {
		l.Debugw("minio endpoint contains port", "endpoint", ep)
	} else {
		l.Debugw("minio endpoint missing port; adding configured port", "endpoint", ep, "port", cfg.MinioConfig.Port)
		ep = ep + ":" + strconv.Itoa(cfg.MinioConfig.Port)
	}

	l.Infow("initializing minio client", "endpoint", ep, "secure", cfg.MinioConfig.UseSSL)

	client, err := minio.New(ep, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioConfig.AccessKey, cfg.MinioConfig.SecretKey, ""),
		Secure: cfg.MinioConfig.UseSSL,
	})
	if err != nil {
		l.Errorw("failed to init minio client", "error", err)
		return nil, fmt.Errorf("failed to init minio client: %w", err)
	}

	l.Infow("checking bucket existence", "bucket", cfg.MinioConfig.BucketName)
	exists, err := client.BucketExists(ctx, cfg.MinioConfig.BucketName)
	if err != nil {
		l.Errorw("failed to check bucket existence", "error", err, "bucket", cfg.MinioConfig.BucketName)
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}
	if !exists {
		l.Infow("bucket does not exist; creating", "bucket", cfg.MinioConfig.BucketName)
		if err := client.MakeBucket(ctx, cfg.MinioConfig.BucketName, minio.MakeBucketOptions{Region: cfg.MinioConfig.Location}); err != nil {
			l.Errorw("failed to create bucket", "bucket", cfg.MinioConfig.BucketName, "error", err)
			return nil, fmt.Errorf("failed to create bucket %s: %w", cfg.MinioConfig.BucketName, err)
		}
		l.Infow("bucket created", "bucket", cfg.MinioConfig.BucketName)
	} else {
		l.Infow("bucket exists", "bucket", cfg.MinioConfig.BucketName)
	}

	return client, nil
}
