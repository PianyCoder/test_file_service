package controller

import (
	"bytes"
	"context"
	"fmt"
	pb "github.com/PianyCoder/test_file_service/internal/proto"
	"github.com/PianyCoder/test_file_service/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
)

const bufferSize = 1024 * 64 // 64 KiB

type FileServiceHandler struct {
	pb.UnimplementedFileServiceServer
	service service.FileService
}

func NewFileServiceHandler(svc service.FileService) *FileServiceHandler {
	return &FileServiceHandler{service: svc}
}

func (h *FileServiceHandler) UploadFile(stream pb.FileService_UploadFileServer) error {
	ctx := stream.Context()

	var filename string
	var buf bytes.Buffer
	first := true

	for {
		select {
		case <-ctx.Done():
			return status.Errorf(codes.Canceled, "request canceled")
		default:
		}

		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return status.Errorf(codes.Internal, "receive chunk error: %v", err)
		}

		if first {
			filename = req.GetFilename()
			if filename == "" {
				return status.Errorf(codes.InvalidArgument, "filename required in first message")
			}
			first = false
		}
		if len(req.GetChunk()) > 0 {
			if _, err := buf.Write(req.GetChunk()); err != nil {
				return status.Errorf(codes.Internal, "buffer write error: %v", err)
			}
		}
	}

	if filename == "" {
		return status.Errorf(codes.InvalidArgument, "no filename provided")
	}
	if buf.Len() == 0 {
		return status.Errorf(codes.InvalidArgument, "no file data provided")
	}

	if err := h.service.UploadFile(ctx, filename, bytes.NewReader(buf.Bytes())); err != nil {
		return status.Errorf(codes.Internal, "upload failed: %v", err)
	}

	return stream.SendAndClose(&pb.UploadFileResponse{Message: fmt.Sprintf("file '%s' uploaded", filename)})
}

func (h *FileServiceHandler) DownloadFile(req *pb.DownloadFileRequest, stream pb.FileService_DownloadFileServer) error {
	ctx := stream.Context()
	filename := req.GetFilename()
	if filename == "" {
		return status.Errorf(codes.InvalidArgument, "filename required")
	}

	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		if err := h.service.DownloadFile(ctx, filename, pw); err != nil {
			_ = pw.CloseWithError(err)
		}
	}()

	buf := make([]byte, bufferSize)
	for {
		select {
		case <-ctx.Done():
			return status.Errorf(codes.Canceled, "request canceled")
		default:
		}

		n, err := pr.Read(buf)
		if n > 0 {
			if sendErr := stream.Send(&pb.DownloadFileResponse{Chunk: buf[:n]}); sendErr != nil {
				return status.Errorf(codes.Internal, "send chunk error: %v", sendErr)
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return status.Errorf(codes.Internal, "read pipe error: %v", err)
		}
	}

	return nil
}

func (h *FileServiceHandler) ListFiles(ctx context.Context, req *pb.ListFilesRequest) (*pb.ListFilesResponse, error) {
	metadata, err := h.service.ListFiles(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list files error: %v", err)
	}

	resp := &pb.ListFilesResponse{}
	for _, m := range metadata {
		resp.Files = append(resp.Files, &pb.FileMetadata{
			Name:      m.Name,
			CreatedAt: timestamppb.New(m.CreatedAt),
			UpdatedAt: timestamppb.New(m.UpdatedAt),
		})
	}
	return resp, nil
}
