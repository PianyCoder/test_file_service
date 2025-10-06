package minio_cli

import (
	"context"
	"fmt"
	"github.com/PianyCoder/test_file_service/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func NewClient(ctx context.Context, cfg config.Config) (*minio.Client, error) {
	client, err := minio.New(cfg.MinioConfig.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioConfig.AccessKey, cfg.MinioConfig.SecretKey, ""),
		Secure: cfg.MinioConfig.UseSSL,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to init minio client: %w", err)
	}

	exists, err := client.BucketExists(ctx, cfg.MinioConfig.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket: %w", err)
	}

	if !exists {
		if err := client.MakeBucket(ctx, cfg.MinioConfig.Bucket, minio.MakeBucketOptions{Region: cfg.MinioConfig.Location}); err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return client, nil
}
