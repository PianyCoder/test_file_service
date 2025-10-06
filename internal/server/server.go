package server

import (
	"context"
	"fmt"
	"github.com/PianyCoder/test_file_service/infrastructure/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc/reflection"
	"time"

	"github.com/PianyCoder/test_file_service/internal/controller"
	pb "github.com/PianyCoder/test_file_service/internal/proto"
	"google.golang.org/grpc"
	"net"
)

const gracefulTimeout = 5 * time.Second

type GRPCServer struct {
	server   *grpc.Server
	listener net.Listener
	ctrl     *controller.FileServiceHandler
	addr     string
	logger   *zap.Logger
}

func NewGRPCServer(addr string, ctrl *controller.FileServiceHandler) (*GRPCServer, error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen error: %w", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterFileServiceServer(grpcServer, ctrl)
	reflection.Register(grpcServer)

	return &GRPCServer{
		server:   grpcServer,
		listener: lis,
		ctrl:     ctrl,
		addr:     addr,
	}, nil
}

func (s *GRPCServer) StartServer(ctx context.Context) error {
	l := logger.FromContext(ctx)
	l.Infof("gRPC listening on %s", s.addr)

	go func() {
		if err := s.server.Serve(s.listener); err != nil {
			l.Fatalf("gRPC Serve error: %v", err)
		}
	}()

	<-ctx.Done()
	l.Info("Shutdown requested, stopping gRPC server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	select {
	case <-shutdownCtx.Done():
		l.Warn("Graceful shutdown timed out, forcing stop")
		s.server.Stop()
	case <-done:
		l.Info("gRPC server stopped gracefully")
	}

	_ = s.logger.Sync()
	return nil
}
