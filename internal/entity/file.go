package entity

import (
	"os"
	"time"

	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

type FileMetadata struct {
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func FileMetadataFromProto(protoMeta *timestamppb.Timestamp, updateTime *timestamppb.Timestamp, name string) FileMetadata {
	fm := FileMetadata{
		Name: name,
	}
	if protoMeta != nil {
		fm.CreatedAt = protoMeta.AsTime()
	}
	if updateTime != nil {
		fm.UpdatedAt = updateTime.AsTime()
	}
	return fm
}

func (fm *FileMetadata) ToProto() *timestamppb.Timestamp {
	if fm.CreatedAt.IsZero() {
		return nil
	}
	return timestamppb.New(fm.CreatedAt)
}

func (fm *FileMetadata) ToProtoUpdate() *timestamppb.Timestamp {
	if fm.UpdatedAt.IsZero() {
		return nil
	}
	return timestamppb.New(fm.UpdatedAt)
}

func GetFileMetadataFromOSStat(filePath string) (*FileMetadata, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	if fileInfo.IsDir() {
		return nil, nil
	}

	return &FileMetadata{
		Name:      fileInfo.Name(),
		CreatedAt: fileInfo.ModTime(),
		UpdatedAt: fileInfo.ModTime(),
	}, nil
}
