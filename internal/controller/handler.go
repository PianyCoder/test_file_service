package controller

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	pb "github.com/PianyCoder/test_file_service/internal/proto"
	"github.com/PianyCoder/test_file_service/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"os"
)

const (
	MaxUploadFileSize = 100 * 1024 * 1024
)

type FileServiceHandler struct {
	pb.UnimplementedFileServiceServer

	fileUseCase service.FileService
}

func NewFileServiceHandler(fileUseCase service.FileService) *FileServiceHandler {
	return &FileServiceHandler{
		fileUseCase: fileUseCase,
	}
}

func (h *FileServiceHandler) UploadFile(stream pb.FileService_UploadFileServer) error {
	var filename string
	var fullData []byte
	var firstMessage = true

	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			s, ok := status.FromError(err)
			if ok {
				if s.Code() == codes.Canceled || s.Code() == codes.DeadlineExceeded {
					return err
				}
				return status.Errorf(codes.Internal, "error receiving chunk: %v", err)
			}
			return status.Errorf(codes.Internal, "error receiving chunk: %v", err)
		}

		if firstMessage {
			filename = req.GetFilename()
			if filename == "" {
				return status.Errorf(codes.InvalidArgument, "filename must be provided in the first message")
			}
			firstMessage = false
		}

		fullData = append(fullData, req.GetChunk()...)

		if len(fullData) > MaxUploadFileSize {
			return status.Errorf(codes.InvalidArgument, "file size exceeds the upload limit of %d bytes", MaxUploadFileSize)
		}
	}

	if filename == "" || len(fullData) == 0 {
		return status.Errorf(codes.InvalidArgument, "no file data received or filename not set")
	}

	dataReader := io.NopCloser(bytes.NewReader(fullData))
	err := h.fileUseCase.UploadFile(filename, dataReader)
	if err != nil {
		if _, ok := err.(*os.PathError); ok || errors.Is(err, os.ErrNotExist) || errors.Is(err, os.ErrPermission) {
			return status.Errorf(codes.Internal, "failed to save file: %v", err)
		}
		if _, ok := err.(*os.PathError); ok || errors.Is(err, os.ErrExist) {
			return status.Errorf(codes.AlreadyExists, "file already exists: %v", err)
		}
		return status.Errorf(codes.Internal, "failed to upload file: %v", err)
	}

	return stream.SendAndClose(&pb.UploadFileResponse{
		Message: fmt.Sprintf("File '%s' uploaded successfully", filename),
	})
}

func (h *FileServiceHandler) DownloadFile(req *pb.DownloadFileRequest, stream pb.FileService_DownloadFileServer) error {
	filename := req.GetFilename()
	if filename == "" {
		return status.Errorf(codes.InvalidArgument, "filename must be provided for download")
	}

	reader, writer := io.Pipe()

	go func() {
		defer writer.Close()
		err := h.fileUseCase.DownloadFile(filename, writer)
		if err != nil {
			writer.CloseWithError(err)
			return
		}
	}()

	buf := make([]byte, service.DefaultChunkSize)
	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return status.Errorf(codes.Internal, "error reading file content: %v", err)
		}

		err = stream.Send(&pb.DownloadFileResponse{
			Chunk: buf[:n],
		})
		if err != nil {
			return status.Errorf(codes.Internal, "error sending file chunk: %v", err)
		}
	}

	return nil
}

func (h *FileServiceHandler) ListFiles(ctx context.Context, req *pb.ListFilesRequest) (*pb.ListFilesResponse, error) {
	metadataList, err := h.fileUseCase.ListFiles()
	if err != nil {
		if _, ok := err.(*os.PathError); ok || errors.Is(err, os.ErrNotExist) {
			return nil, status.Errorf(codes.NotFound, "no files found or storage error: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to list files: %v", err)
	}

	protoFiles := make([]*pb.FileMetadata, len(metadataList))
	for i, meta := range metadataList {
		protoFiles[i] = &pb.FileMetadata{
			Name:      meta.Name,
			CreatedAt: timestamppb.New(meta.CreatedAt),
			UpdatedAt: timestamppb.New(meta.UpdatedAt),
		}
	}

	return &pb.ListFilesResponse{
		Files: protoFiles,
	}, nil
}
