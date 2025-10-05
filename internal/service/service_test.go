// service_test.go
package service

import (
	"bytes"
	"github.com/PianyCoder/test_file_service/internal/entity"
	"github.com/PianyCoder/test_file_service/internal/storage"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func createTestStorageDir(t *testing.T) (string, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "test-storage-")
	assert.NoError(t, err, "Failed to create temp directory")

	cleanup := func() {
		if _, err := os.Stat(tempDir); err == nil {
			os.RemoveAll(tempDir)
		}
	}
	return tempDir, cleanup
}

func TestFileService(t *testing.T) {
	tempDir, cleanup := createTestStorageDir(t)
	defer cleanup()

	fileStorage, err := storage.NewFileStorage(tempDir)
	assert.NoError(t, err, "Failed to create file storage")

	fileService := NewFileService(fileStorage, 10, 10, 10, tempDir, 1024)
	assert.NotNil(t, fileService, "Failed to create file service")

	t.Run("UploadAndDownloadFile", func(t *testing.T) {
		filename := "testfile.txt"
		fileContent := []byte("This is the content of the test file.")

		reader := bytes.NewReader(fileContent)
		err := fileService.UploadFile(filename, reader)
		assert.NoError(t, err, "UploadFile failed")

		var downloadedContent bytes.Buffer
		err = fileService.DownloadFile(filename, &downloadedContent)
		assert.NoError(t, err, "DownloadFile failed")

		assert.Equal(t, fileContent, downloadedContent.Bytes(), "Downloaded content does not match original content")
	})

	t.Run("UploadAndDownloadEmptyFile", func(t *testing.T) {
		filename := "emptyfile.txt"
		fileContent := []byte("")

		reader := bytes.NewReader(fileContent)
		err := fileService.UploadFile(filename, reader)
		assert.NoError(t, err, "UploadFile failed for empty file")

		var downloadedContent bytes.Buffer
		err = fileService.DownloadFile(filename, &downloadedContent)
		assert.NoError(t, err, "DownloadFile failed for empty file")

		actualContent := downloadedContent.Bytes()
		assert.True(t, actualContent == nil || len(actualContent) == 0,
			"Downloaded empty file content mismatch: expected nil or empty slice, got %v", actualContent)
	})

	t.Run("DownloadNonExistentFile", func(t *testing.T) {
		filename := "nonexistent.txt"
		var writer bytes.Buffer
		err := fileService.DownloadFile(filename, &writer)
		assert.Error(t, err, "DownloadFile should return an error for non-existent file")
		assert.Contains(t, err.Error(), "file not found", "Error message should indicate file not found")
	})

	t.Run("UploadInvalidFilename", func(t *testing.T) {
		invalidFilename := "../sensitive/secrets.txt"
		reader := bytes.NewReader([]byte("some data"))
		err := fileService.UploadFile(invalidFilename, reader)
		assert.Error(t, err, "UploadFile should return an error for invalid filename")
		assert.Contains(t, err.Error(), "invalid filename", "Error message should indicate invalid filename")
	})

	t.Run("DownloadInvalidFilename", func(t *testing.T) {
		invalidFilename := "../sensitive/secrets.txt"
		var writer bytes.Buffer
		err := fileService.DownloadFile(invalidFilename, &writer)
		assert.Error(t, err, "DownloadFile should return an error for invalid filename")
		assert.Contains(t, err.Error(), "invalid filename", "Error message should indicate invalid filename")
	})

	t.Run("ListFilesAfterUpload", func(t *testing.T) {
		cleanup()
		tempDir, cleanup = createTestStorageDir(t)
		defer cleanup()

		fileStorage, err = storage.NewFileStorage(tempDir)
		assert.NoError(t, err, "Failed to re-create file storage for ListFiles test")
		fileService = NewFileService(fileStorage, 10, 10, 10, tempDir, 1024)

		err = fileService.UploadFile("file1.txt", bytes.NewReader([]byte("content1")))
		assert.NoError(t, err)
		err = fileService.UploadFile("file2.log", bytes.NewReader([]byte("content2")))
		assert.NoError(t, err)

		metadataList, err := fileService.ListFiles()
		assert.NoError(t, err, "ListFiles failed")

		assert.Len(t, metadataList, 2, "Expected 2 files in the list")

		metadataMap := make(map[string]entity.FileMetadata)
		for _, meta := range metadataList {
			metadataMap[meta.Name] = meta
		}

		meta1, ok := metadataMap["file1.txt"]
		assert.True(t, ok, "Metadata for file1.txt not found")
		assert.Equal(t, "file1.txt", meta1.Name)
		assert.WithinDuration(t, time.Now(), meta1.CreatedAt, 5*time.Second, "CreatedAt for file1.txt is too old")
		assert.WithinDuration(t, time.Now(), meta1.UpdatedAt, 5*time.Second, "UpdatedAt for file1.txt is too old")

		meta2, ok := metadataMap["file2.log"]
		assert.True(t, ok, "Metadata for file2.log not found")
		assert.Equal(t, "file2.log", meta2.Name)
		assert.WithinDuration(t, time.Now(), meta2.CreatedAt, 5*time.Second, "CreatedAt for file2.log is too old")
		assert.WithinDuration(t, time.Now(), meta2.UpdatedAt, 5*time.Second, "UpdatedAt for file2.log is too old")
	})

	t.Run("ListFilesEmptyStorage", func(t *testing.T) {
		cleanup()
		tempDir, cleanup = createTestStorageDir(t)
		defer cleanup()

		fileStorage, err = storage.NewFileStorage(tempDir)
		assert.NoError(t, err, "Failed to re-create file storage for empty test")
		fileService = NewFileService(fileStorage, 10, 10, 10, tempDir, 1024)

		metadataList, err := fileService.ListFiles()
		assert.NoError(t, err, "ListFiles should not error on empty storage")
		assert.Len(t, metadataList, 0, "Expected 0 files in empty storage")
	})

	t.Run("UploadFileEmptyFilename", func(t *testing.T) {
		reader := bytes.NewReader([]byte("some data"))
		err := fileService.UploadFile("", reader)
		assert.Error(t, err, "UploadFile should return error for empty filename")
		assert.Contains(t, err.Error(), "filename cannot be empty")
	})

	t.Run("DownloadFileEmptyFilename", func(t *testing.T) {
		var writer bytes.Buffer
		err := fileService.DownloadFile("", &writer)
		assert.Error(t, err, "DownloadFile should return error for empty filename")
		assert.Contains(t, err.Error(), "filename cannot be empty")
	})
}
