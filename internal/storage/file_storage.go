package storage

import (
	"fmt"
	"github.com/PianyCoder/test_file_service/internal/entity"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type FileStorage interface {
	SaveFile(filename string, data []byte) error
	GetFileContent(filename string) ([]byte, error)
	ListAllFilesMetadata() ([]entity.FileMetadata, error)
}

type fileStorage struct {
	storageDir string
	mu         sync.RWMutex
}

func NewFileStorage(storageDir string) (*fileStorage, error) {
	err := os.MkdirAll(storageDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage directory '%s': %w", storageDir, err)
	}
	return &fileStorage{
		storageDir: storageDir,
	}, nil
}

func (fs *fileStorage) SaveFile(filename string, data []byte) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	fullPath := filepath.Join(fs.storageDir, filename)
	if filepath.Clean(fullPath) != filepath.Join(fs.storageDir, filepath.Base(filename)) {
		return fmt.Errorf("invalid filename: %s (potential path traversal)", filename)
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file '%s': %w", filename, err)
	}
	defer func() {
		_ = file.Close()
	}()

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write data to file '%s': %w", filename, err)
	}

	return nil
}

func (fs *fileStorage) GetFileContent(filename string) ([]byte, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	if filename == "" {
		return nil, fmt.Errorf("filename cannot be empty")
	}

	fullPath := filepath.Join(fs.storageDir, filename)
	if filepath.Clean(fullPath) != filepath.Join(fs.storageDir, filepath.Base(filename)) {
		return nil, fmt.Errorf("invalid filename: %s (potential path traversal)", filename)
	}

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", filename)
		}
		return nil, fmt.Errorf("failed to open file '%s': %w", filename, err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			fmt.Printf("Warning: error closing file '%s': %v\n", filename, cerr)
		}
	}()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file '%s': %w", filename, err)
	}

	return content, nil
}

func (fs *fileStorage) ListAllFilesMetadata() ([]entity.FileMetadata, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var metadataList []entity.FileMetadata

	entries, err := os.ReadDir(fs.storageDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read storage directory '%s': %w", fs.storageDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			fileInfo, err := entry.Info()
			if err != nil {
				fmt.Printf("Warning: Could not get info for file '%s': %v\n", entry.Name(), err)
				continue
			}

			metadataList = append(metadataList, entity.FileMetadata{
				Name:      entry.Name(),
				CreatedAt: fileInfo.ModTime(),
				UpdatedAt: fileInfo.ModTime(),
			})
		}
	}

	return metadataList, nil
}
