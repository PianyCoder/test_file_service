package storage

import (
	"bytes"
	"context"
	"fmt"
	"github.com/PianyCoder/test_file_service/internal/entity"
	"github.com/minio/minio-go/v7"
)

type MinioStorage struct {
	client *minio.Client
	bucket string
}

func NewMinioStorage(client *minio.Client, bucket string) *MinioStorage {
	return &MinioStorage{client: client, bucket: bucket}
}

func (ms *MinioStorage) SaveFile(ctx context.Context, filename string, data []byte) error {
	_, err := ms.client.PutObject(ctx, ms.bucket, filename, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to upload file to minio: %w", err)
	}
	return nil
}

func (ms *MinioStorage) GetFileContent(ctx context.Context, filename string) ([]byte, error) {
	obj, err := ms.client.GetObject(ctx, ms.bucket, filename, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	defer func() {
		_ = obj.Close()
	}()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(obj); err != nil {
		return nil, fmt.Errorf("failed to read object: %w", err)
	}
	return buf.Bytes(), nil
}

func (ms *MinioStorage) ListAllFilesMetadata(ctx context.Context) ([]entity.FileMetadata, error) {
	var metadata []entity.FileMetadata
	opts := minio.ListObjectsOptions{
		Recursive: true,
	}

	objectCh := ms.client.ListObjects(ctx, ms.bucket, opts)
	for obj := range objectCh {
		if obj.Err != nil {
			return nil, fmt.Errorf("minio: list objects error: %w", obj.Err)
		}
		metadata = append(metadata, entity.FileMetadata{
			Name:      obj.Key,
			CreatedAt: obj.LastModified,
			UpdatedAt: obj.LastModified,
		})
	}
	return metadata, nil
}
