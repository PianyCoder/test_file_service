package service

import (
	"context"
	"github.com/PianyCoder/test_file_service/internal/entity"
	"io"
)

type FileService interface {
	UploadFile(ctx context.Context, filename string, reader io.Reader) error
	DownloadFile(ctx context.Context, filename string, writer io.Writer) error
	ListFiles(ctx context.Context) ([]entity.FileMetadata, error)
}
