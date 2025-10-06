package storage

import (
	"context"
	"github.com/PianyCoder/test_file_service/internal/entity"
	"io"
)

type FileStorage interface {
	SaveFile(ctx context.Context, filename string, r io.Reader, size int64) error
	GetFileReader(ctx context.Context, filename string) (io.ReadCloser, error)
	ListAllFilesMetadata(ctx context.Context) ([]entity.FileMetadata, error)
}
