package service

import (
	"context"
	"fmt"
	"github.com/PianyCoder/test_file_service/infrastructure/logger"
	"github.com/PianyCoder/test_file_service/internal/storage"

	entity "github.com/PianyCoder/test_file_service/internal/entity"
	"golang.org/x/sync/semaphore"
	"io"
	"path/filepath"
)

const DefaultChunkSize = 1024 * 1024

type fileService struct {
	storage         storage.FileStorage
	chunkSize       int
	uploadLimiter   *semaphore.Weighted
	downloadLimiter *semaphore.Weighted
	listLimiter     *semaphore.Weighted
}

func NewFileService(storage storage.FileStorage, uploadLimit, downloadLimit, listLimit int64, chunkSize int) FileService {
	if chunkSize <= 0 {
		chunkSize = DefaultChunkSize
	}

	if uploadLimit <= 0 {
		uploadLimit = 10
	}
	if downloadLimit <= 0 {
		downloadLimit = 10
	}
	if listLimit <= 0 {
		listLimit = 100
	}

	return &fileService{
		storage:         storage,
		chunkSize:       chunkSize,
		uploadLimiter:   semaphore.NewWeighted(uploadLimit),
		downloadLimiter: semaphore.NewWeighted(downloadLimit),
		listLimiter:     semaphore.NewWeighted(listLimit),
	}
}

func (fs *fileService) UploadFile(ctx context.Context, filename string, reader io.Reader) error {
	l := logger.FromContext(ctx)
	l.Infow("service.UploadFile called", "filename", filename)
	if err := fs.uploadLimiter.Acquire(ctx, 1); err != nil {
		l.Errorw("failed to acquire upload semaphore", "error", err)
		return fmt.Errorf("failed to acquire upload semaphore: %w", err)
	}
	defer fs.uploadLimiter.Release(1)

	if filename == "" {
		l.Error("filename cannot be empty")
		return fmt.Errorf("filename cannot be empty")
	}

	if filepath.Clean(filename) != filepath.Base(filename) {
		l.Errorw("invalid filename (possible traversal)", "filename", filename)
		return fmt.Errorf("invalid filename (possible traversal): %s", filename)
	}

	if err := fs.storage.SaveFile(ctx, filename, reader, -1); err != nil {
		l.Errorw("storage error on save", "error", err, "filename", filename)
		return fmt.Errorf("storage error on save: %w", err)
	}
	l.Infow("service.UploadFile finished", "filename", filename)
	return nil
}

func (fs *fileService) DownloadFile(ctx context.Context, filename string, writer io.Writer) error {
	l := logger.FromContext(ctx)
	l.Infow("service.DownloadFile called", "filename", filename)
	if err := fs.downloadLimiter.Acquire(ctx, 1); err != nil {
		l.Errorw("failed to acquire download semaphore", "error", err)
		return fmt.Errorf("failed to acquire download semaphore: %w", err)
	}
	defer fs.downloadLimiter.Release(1)

	if filename == "" {
		l.Error("filename cannot be empty")
		return fmt.Errorf("filename cannot be empty")
	}
	if filepath.Clean(filename) != filepath.Base(filename) {
		l.Errorw("invalid filename (possible traversal)", "filename", filename)
		return fmt.Errorf("invalid filename (possible traversal): %s", filename)
	}

	r, err := fs.storage.GetFileReader(ctx, filename)
	if err != nil {
		l.Errorw("storage error on get", "error", err, "filename", filename)
		return fmt.Errorf("storage error on get: %w", err)
	}
	defer r.Close()

	n, err := io.Copy(writer, r)
	if err != nil {
		l.Errorw("failed to stream file to writer", "error", err, "bytes_copied", n, "filename", filename)
		return fmt.Errorf("failed to stream file to writer: %w", err)
	}
	l.Infow("service.DownloadFile finished", "filename", filename, "bytes", n)
	return nil
}

func (fs *fileService) ListFiles(ctx context.Context) ([]entity.FileMetadata, error) {
	l := logger.FromContext(ctx)
	l.Infow("service.ListFiles called")
	if err := fs.listLimiter.Acquire(ctx, 1); err != nil {
		l.Errorw("failed to acquire list semaphore", "error", err)
		return nil, fmt.Errorf("failed to acquire list semaphore: %w", err)
	}
	defer fs.listLimiter.Release(1)

	metadata, err := fs.storage.ListAllFilesMetadata(ctx)
	if err != nil {
		l.Errorw("storage error list", "error", err)
		return nil, fmt.Errorf("storage error list: %w", err)
	}
	l.Infow("service.ListFiles finished", "count", len(metadata))
	return metadata, nil
}
