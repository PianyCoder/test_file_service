package service

import (
	"context"
	"fmt"
	"github.com/PianyCoder/test_file_service/internal/storage"

	"github.com/PianyCoder/test_file_service/internal/entity"
	"golang.org/x/sync/semaphore"
	"io"
	"path/filepath"
	"sync"
)

const (
	DefaultStorageDir = "./storage_data"
	DefaultChunkSize  = 1024 * 1024
)

type FileService interface {
	UploadFile(filename string, reader io.Reader) error
	DownloadFile(filename string, writer io.Writer) error
	ListFiles() ([]entity.FileMetadata, error)
}

type fileService struct {
	storage         storage.FileStorage
	storageDir      string
	chunkSize       int
	uploadLimiter   *semaphore.Weighted
	downloadLimiter *semaphore.Weighted
	listLimiter     *semaphore.Weighted
	mu              sync.RWMutex
}

func NewFileService(storage storage.FileStorage, uploadLimit int64, downloadLimit int64, listLimit int64, storageDir string, chunkSize int) *fileService {
	if storageDir == "" {
		storageDir = DefaultStorageDir
	}
	if chunkSize <= 0 {
		chunkSize = DefaultChunkSize
	}

	return &fileService{
		storage:         storage,
		storageDir:      storageDir,
		chunkSize:       chunkSize,
		uploadLimiter:   semaphore.NewWeighted(uploadLimit),
		downloadLimiter: semaphore.NewWeighted(downloadLimit),
		listLimiter:     semaphore.NewWeighted(listLimit),
	}
}

func (fs *fileService) UploadFile(filename string, reader io.Reader) error {
	if err := fs.uploadLimiter.Acquire(context.Background(), 1); err != nil {
		return fmt.Errorf("failed to acquire upload semaphore: %w", err)
	}
	defer fs.uploadLimiter.Release(1)

	if filename == "" {
		return fmt.Errorf("filename cannot be empty for upload")
	}
	if filepath.Clean(filename) != filepath.Base(filename) {
		return fmt.Errorf("invalid filename, prevents directory traversal: %s", filename)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read file data: %w", err)
	}

	if err := fs.storage.SaveFile(filename, data); err != nil {
		return fmt.Errorf("storage error during upload: %w", err)
	}

	return nil
}

func (fs *fileService) DownloadFile(filename string, writer io.Writer) error {
	if err := fs.downloadLimiter.Acquire(context.Background(), 1); err != nil {
		return fmt.Errorf("failed to acquire download semaphore: %w", err)
	}
	defer fs.downloadLimiter.Release(1)

	if filename == "" {
		return fmt.Errorf("filename cannot be empty for download")
	}
	if filepath.Clean(filename) != filepath.Base(filename) {
		return fmt.Errorf("invalid filename, prevents directory traversal: %s", filename)
	}

	fileContent, err := fs.storage.GetFileContent(filename)
	if err != nil {
		return fmt.Errorf("storage error during download: %w", err)
	}

	_, err = writer.Write(fileContent)
	if err != nil {
		return fmt.Errorf("failed to write file content to writer: %w", err)
	}

	return nil
}

func (fs *fileService) ListFiles() ([]entity.FileMetadata, error) {
	if err := fs.listLimiter.Acquire(context.Background(), 1); err != nil {
		return nil, fmt.Errorf("failed to acquire list semaphore: %w", err)
	}
	defer fs.listLimiter.Release(1)

	metadataList, err := fs.storage.ListAllFilesMetadata()
	if err != nil {
		return nil, fmt.Errorf("storage error listing files: %w", err)
	}

	return metadataList, nil
}
