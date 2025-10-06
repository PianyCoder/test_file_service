package controller

import (
	"context"
	"fmt"
	"github.com/PianyCoder/test_file_service/infrastructure/logger"
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
	l := logger.FromContext(ctx)
	l.Info("UploadFile called")

	req, err := stream.Recv()
	if err != nil {
		l.Errorw("receive first message error", "error", err)
		if err == io.EOF {
			return status.Errorf(codes.InvalidArgument, "no messages received")
		}
		return status.Errorf(codes.Internal, "failed to receive first message")
	}

	filename := req.GetFilename()
	if filename == "" {
		l.Warn("filename required in first message")
		return status.Errorf(codes.InvalidArgument, "filename required in first message")
	}
	l.Infow("upload metadata received", "filename", filename)

	pr, pw := io.Pipe()

	go func(firstChunk []byte) {
		defer func() {
			_ = pw.Close()
			l.Debug("uploader goroutine finished and pipe writer closed")
		}()
		if len(firstChunk) > 0 {
			if _, werr := pw.Write(firstChunk); werr != nil {
				_ = pw.CloseWithError(werr)
				l.Errorw("failed to write first chunk to pipe", "error", werr)
				return
			}
		}
		for {
			msg, rerr := stream.Recv()
			if rerr != nil {
				if rerr == io.EOF {
					l.Debug("uploader goroutine received EOF from stream")
					return
				}
				_ = pw.CloseWithError(rerr)
				l.Errorw("error receiving chunk from stream", "error", rerr)
				return
			}
			ch := msg.GetChunk()
			if len(ch) > 0 {
				if _, werr := pw.Write(ch); werr != nil {
					_ = pw.CloseWithError(werr)
					l.Errorw("failed to write chunk to pipe", "error", werr)
					return
				}
			}
		}
	}(req.GetChunk())

	if err := h.service.UploadFile(ctx, filename, pr); err != nil {
		l.Errorw("upload failed", "filename", filename, "error", err)
		return status.Errorf(codes.Internal, "upload failed")
	}

	l.Infow("upload finished successfully", "filename", filename)
	return stream.SendAndClose(&pb.UploadFileResponse{Message: fmt.Sprintf("file '%s' uploaded", filename)})
}

func (h *FileServiceHandler) DownloadFile(req *pb.DownloadFileRequest, stream pb.FileService_DownloadFileServer) error {
	ctx := stream.Context()
	l := logger.FromContext(ctx)
	l.Infow("DownloadFile called", "req_filename", req.GetFilename())

	filename := req.GetFilename()
	if filename == "" {
		l.Warn("download filename required")
		return status.Errorf(codes.InvalidArgument, "filename required")
	}

	pr, pw := io.Pipe()
	go func() {
		defer func() {
			_ = pw.Close()
			l.Debug("downloader goroutine finished and pipe writer closed")
		}()
		if err := h.service.DownloadFile(ctx, filename, pw); err != nil {
			_ = pw.CloseWithError(err)
			l.Errorw("service.DownloadFile error", "filename", filename, "error", err)
			return
		}
	}()

	buf := make([]byte, bufferSize)
	for {
		select {
		case <-ctx.Done():
			l.Warn("download context canceled", "filename", filename)
			return status.Errorf(codes.Canceled, "request canceled")
		default:
		}

		n, err := pr.Read(buf)
		if n > 0 {
			if sendErr := stream.Send(&pb.DownloadFileResponse{Chunk: buf[:n]}); sendErr != nil {
				l.Errorw("failed to send chunk to client", "error", sendErr)
				return status.Errorf(codes.Internal, "failed to send chunk")
			}
		}
		if err != nil {
			if err == io.EOF {
				l.Infow("download finished", "filename", filename)
				break
			}
			l.Errorw("read pipe error during download", "error", err, "filename", filename)
			return status.Errorf(codes.Internal, "failed to read file")
		}
	}
	return nil
}

func (h *FileServiceHandler) ListFiles(ctx context.Context, req *pb.ListFilesRequest) (*pb.ListFilesResponse, error) {
	l := logger.FromContext(ctx)
	l.Info("ListFiles called")
	metadata, err := h.service.ListFiles(ctx)
	if err != nil {
		l.Errorw("list files error", "error", err)
		return nil, status.Errorf(codes.Internal, "list files error")
	}

	resp := &pb.ListFilesResponse{}
	for _, m := range metadata {
		resp.Files = append(resp.Files, &pb.FileMetadata{
			Name:      m.Name,
			CreatedAt: timestamppb.New(m.CreatedAt),
			UpdatedAt: timestamppb.New(m.UpdatedAt),
		})
	}
	l.Infow("ListFiles finished", "count", len(resp.Files))
	return resp, nil
}
