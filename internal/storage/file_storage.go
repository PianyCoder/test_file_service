package storage

import (
	"context"
	"github.com/PianyCoder/test_file_service/internal/entity"
)

type FileStorage interface {
	SaveFile(ctx context.Context, filename string, data []byte) error
	GetFileContent(ctx context.Context, filename string) ([]byte, error)
	ListAllFilesMetadata(ctx context.Context) ([]entity.FileMetadata, error)
}
