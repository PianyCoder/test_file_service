package service

import (
	"context"
	"fmt"
	"github.com/PianyCoder/test_file_service/internal/storage"

	"github.com/PianyCoder/test_file_service/internal/entity"
	"golang.org/x/sync/semaphore"
	"io"
	"path/filepath"
)

const DefaultChunkSize = 1024 * 1024

type FileService interface {
	UploadFile(ctx context.Context, filename string, reader io.Reader) error
	DownloadFile(ctx context.Context, filename string, writer io.Writer) error
	ListFiles(ctx context.Context) ([]entity.FileMetadata, error)
}

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
	return &fileService{
		storage:         storage,
		chunkSize:       chunkSize,
		uploadLimiter:   semaphore.NewWeighted(uploadLimit),
		downloadLimiter: semaphore.NewWeighted(downloadLimit),
		listLimiter:     semaphore.NewWeighted(listLimit),
	}
}

func (fs *fileService) UploadFile(ctx context.Context, filename string, reader io.Reader) error {
	// Acquire with ctx so cancellation/deadline works.
	if err := fs.uploadLimiter.Acquire(ctx, 1); err != nil {
		return fmt.Errorf("failed to acquire upload semaphore: %w", err)
	}
	defer fs.uploadLimiter.Release(1)

	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}
	if filepath.Clean(filename) != filepath.Base(filename) {
		return fmt.Errorf("invalid filename (possible traversal): %s", filename)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read upload data: %w", err)
	}

	if err := fs.storage.SaveFile(ctx, filename, data); err != nil {
		return fmt.Errorf("storage error on save: %w", err)
	}
	return nil
}

func (fs *fileService) DownloadFile(ctx context.Context, filename string, writer io.Writer) error {
	if err := fs.downloadLimiter.Acquire(ctx, 1); err != nil {
		return fmt.Errorf("failed to acquire download semaphore: %w", err)
	}
	defer fs.downloadLimiter.Release(1)

	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}
	if filepath.Clean(filename) != filepath.Base(filename) {
		return fmt.Errorf("invalid filename (possible traversal): %s", filename)
	}

	data, err := fs.storage.GetFileContent(ctx, filename)
	if err != nil {
		return fmt.Errorf("storage error on get: %w", err)
	}

	_, err = writer.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to writer: %w", err)
	}
	return nil
}

func (fs *fileService) ListFiles(ctx context.Context) ([]entity.FileMetadata, error) {
	if err := fs.listLimiter.Acquire(ctx, 1); err != nil {
		return nil, fmt.Errorf("failed to acquire list semaphore: %w", err)
	}
	defer fs.listLimiter.Release(1)

	metadata, err := fs.storage.ListAllFilesMetadata(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage error list: %w", err)
	}
	return metadata, nil
}
