package storage

import (
	"context"
	"fmt"
	"github.com/PianyCoder/test_file_service/infrastructure/logger"
	"github.com/PianyCoder/test_file_service/internal/entity"
	"github.com/minio/minio-go/v7"
	"io"
	"os"
)

type MinioStorage struct {
	client *minio.Client
	bucket string
}

func NewMinioStorage(client *minio.Client, bucket string) *MinioStorage {
	return &MinioStorage{client: client, bucket: bucket}
}

func (ms *MinioStorage) SaveFile(ctx context.Context, filename string, r io.Reader, size int64) error {
	l := logger.FromContext(ctx)
	l.Infow("storage.SaveFile called", "filename", filename, "size", size)

	var reader io.Reader = r
	var cleanup func()
	var objectSize int64 = size

	if size < 0 {
		l.Debugw("size unknown, buffering to temp file", "filename", filename)
		f, err := os.CreateTemp("", "upload-*")
		if err != nil {
			l.Errorw("failed to create temp file", "error", err)
			return fmt.Errorf("failed to create temp file: %w", err)
		}
		tempName := f.Name()
		cleanup = func() { _ = os.Remove(tempName); _ = f.Close() }

		n, err := io.Copy(f, r)
		if err != nil {
			_ = f.Close()
			_ = os.Remove(tempName)
			l.Errorw("failed to write temp file", "temp", tempName, "error", err)
			return fmt.Errorf("failed to write temp file: %w", err)
		}
		l.Infow("temp file written", "temp", tempName, "bytes", n)

		stat, err := f.Stat()
		if err != nil {
			_ = f.Close()
			_ = os.Remove(tempName)
			l.Errorw("failed to stat temp file", "temp", tempName, "error", err)
			return fmt.Errorf("failed to stat temp file: %w", err)
		}
		objectSize = stat.Size()

		if _, err := f.Seek(0, io.SeekStart); err != nil {
			_ = f.Close()
			_ = os.Remove(tempName)
			l.Errorw("failed to seek temp file", "temp", tempName, "error", err)
			return fmt.Errorf("failed to seek temp file: %w", err)
		}
		reader = f
	} else {
		reader = r
		cleanup = func() {}
	}

	l.Infow("putting object to minio", "bucket", ms.bucket, "object", filename, "size", objectSize)
	_, err := ms.client.PutObject(ctx, ms.bucket, filename, reader, objectSize, minio.PutObjectOptions{})
	if cleanup != nil {
		cleanup()
	}
	if err != nil {
		l.Errorw("failed to upload file to minio", "bucket", ms.bucket, "object", filename, "error", err)
		return fmt.Errorf("failed to upload file to minio: %w", err)
	}
	l.Infow("file uploaded to minio", "bucket", ms.bucket, "object", filename, "size", objectSize)
	return nil
}

func (ms *MinioStorage) GetFileReader(ctx context.Context, filename string) (io.ReadCloser, error) {
	l := logger.FromContext(ctx)
	l.Infow("GetFileReader called", "bucket", ms.bucket, "object", filename)
	obj, err := ms.client.GetObject(ctx, ms.bucket, filename, minio.GetObjectOptions{})
	if err != nil {
		l.Errorw("failed to get object from minio", "bucket", ms.bucket, "object", filename, "error", err)
		return nil, fmt.Errorf("failed to get object from minio: %w", err)
	}
	return obj, nil
}

func (ms *MinioStorage) ListAllFilesMetadata(ctx context.Context) ([]entity.FileMetadata, error) {
	l := logger.FromContext(ctx)
	l.Infow("ListAllFilesMetadata called", "bucket", ms.bucket)
	var metadata []entity.FileMetadata
	opts := minio.ListObjectsOptions{
		Recursive: true,
	}

	objectCh := ms.client.ListObjects(ctx, ms.bucket, opts)
	count := 0
	for obj := range objectCh {
		if obj.Err != nil {
			l.Errorw("minio: list objects error", "error", obj.Err)
			return nil, fmt.Errorf("minio: list objects error: %w", obj.Err)
		}
		metadata = append(metadata, entity.FileMetadata{
			Name:      obj.Key,
			CreatedAt: obj.LastModified,
			UpdatedAt: obj.LastModified,
		})
		count++
	}
	l.Infow("ListAllFilesMetadata finished", "bucket", ms.bucket, "files", count)
	return metadata, nil
}
