package storage

import (
	"github.com/PianyCoder/test_file_service/internal/entity"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func createTestStorageDir(t *testing.T) (string, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "test-storage-")
	assert.NoError(t, err, "Failed to create temp directory for storage tests")

	cleanup := func() {
		if _, err := os.Stat(tempDir); err == nil {
			os.RemoveAll(tempDir)
		}
	}
	return tempDir, cleanup
}

func TestFileStorage(t *testing.T) {
	tempDir, cleanup := createTestStorageDir(t)
	defer cleanup()

	t.Run("SaveAndGetFile", func(t *testing.T) {
		fileStorage, _ := NewFileStorage(tempDir)

		filename := "mytestfile.txt"
		fileContent := []byte("Content of my test file.")

		err := fileStorage.SaveFile(filename, fileContent)
		assert.NoError(t, err, "SaveFile failed")

		content, err := fileStorage.GetFileContent(filename)
		assert.NoError(t, err, "GetFileContent failed")
		assert.Equal(t, fileContent, content, "Read content mismatch")
	})

	t.Run("GetEmptyFile", func(t *testing.T) {
		fileStorage, _ := NewFileStorage(tempDir)

		filename := "empty.txt"
		emptyContent := []byte("")

		err := fileStorage.SaveFile(filename, emptyContent)
		assert.NoError(t, err, "Failed to save empty file")

		content, err := fileStorage.GetFileContent(filename)
		assert.NoError(t, err, "GetFileContent failed for empty file")

		assert.Equal(t, emptyContent, content, "Reading empty file should return an empty slice")
	})

	t.Run("GetNonExistentFile", func(t *testing.T) {
		fileStorage, _ := NewFileStorage(tempDir)

		filename := "nonexistent_storage.txt"
		content, err := fileStorage.GetFileContent(filename)
		assert.Error(t, err, "GetFileContent should return an error for non-existent file")
		assert.Nil(t, content, "Content should be nil when file not found")
		assert.Contains(t, err.Error(), "file not found", "Error message should indicate file not found")
	})

	t.Run("SaveInvalidFilename", func(t *testing.T) {
		fileStorage, _ := NewFileStorage(tempDir)
		invalidFilename := "../secrets/pass.txt"
		err := fileStorage.SaveFile(invalidFilename, []byte("sensitive data"))
		assert.Error(t, err, "SaveFile should return an error for invalid filename")
		assert.Contains(t, err.Error(), "invalid filename", "Error message should indicate invalid filename")
	})

	t.Run("GetInvalidFilename", func(t *testing.T) {
		fileStorage, _ := NewFileStorage(tempDir)
		invalidFilename := "../secrets/pass.txt"
		_, err := fileStorage.GetFileContent(invalidFilename)
		assert.Error(t, err, "GetFileContent should return an error for invalid filename")
		assert.Contains(t, err.Error(), "invalid filename", "Error message should indicate invalid filename")
	})

	t.Run("SaveEmptyFilename", func(t *testing.T) {
		fileStorage, _ := NewFileStorage(tempDir)
		err := fileStorage.SaveFile("", []byte("some data"))
		assert.Error(t, err, "SaveFile should return error for empty filename")
		assert.Contains(t, err.Error(), "filename cannot be empty")
	})

	t.Run("GetEmptyFilename", func(t *testing.T) {
		fileStorage, _ := NewFileStorage(tempDir)
		_, err := fileStorage.GetFileContent("")
		assert.Error(t, err, "GetFileContent should return error for empty filename")
		assert.Contains(t, err.Error(), "filename cannot be empty")
	})

	t.Run("ListAllFilesMetadata", func(t *testing.T) {
		cleanup()
		tempDir, cleanup = createTestStorageDir(t)
		defer cleanup()

		fileStorage, err := NewFileStorage(tempDir)
		assert.NoError(t, err, "Failed to re-create FileStorage for metadata test")

		file1Name := "file1.txt"
		file2Name := "file2.log"
		file1Content := []byte("content1")
		file2Content := []byte("content2")

		err = fileStorage.SaveFile(file1Name, file1Content)
		assert.NoError(t, err, "Failed to save file1 for metadata test")
		err = fileStorage.SaveFile(file2Name, file2Content)
		assert.NoError(t, err, "Failed to save file2 for metadata test")

		metadataList, err := fileStorage.ListAllFilesMetadata()
		assert.NoError(t, err, "ListAllFilesMetadata failed")

		assert.Len(t, metadataList, 2, "Expected 2 files in metadata list")

		metadataMap := make(map[string]entity.FileMetadata)
		for _, meta := range metadataList {
			metadataMap[meta.Name] = meta
		}

		meta1, ok := metadataMap[file1Name]
		assert.True(t, ok, "Metadata for %s not found", file1Name)
		assert.Equal(t, file1Name, meta1.Name)
		assert.WithinDuration(t, time.Now(), meta1.CreatedAt, 5*time.Second, "CreatedAt for %s is too old", file1Name)

		meta2, ok := metadataMap[file2Name]
		assert.True(t, ok, "Metadata for %s not found", file2Name)
		assert.Equal(t, file2Name, meta2.Name)
		assert.WithinDuration(t, time.Now(), meta2.CreatedAt, 5*time.Second, "CreatedAt for %s is too old", file2Name)
	})

	t.Run("ListAllFilesMetadataEmptyStorage", func(t *testing.T) {
		cleanup()
		tempDir, cleanup = createTestStorageDir(t)
		defer cleanup()

		fileStorage, err := NewFileStorage(tempDir)
		assert.NoError(t, err, "Failed to re-create FileStorage for empty storage test")

		metadataList, err := fileStorage.ListAllFilesMetadata()
		assert.NoError(t, err, "ListAllFilesMetadata should not error on empty storage")
		assert.Len(t, metadataList, 0, "Expected 0 files in metadata list for empty storage")
	})
}
