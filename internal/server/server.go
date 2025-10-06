package server

import (
	"fmt"
	"github.com/PianyCoder/test_file_service/internal/controller"
	pb "github.com/PianyCoder/test_file_service/internal/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

type GRPCServer struct {
	server   *grpc.Server
	listener net.Listener
	ctrl     *controller.FileServiceHandler
	addr     string
	logger   *zap.Logger
}

func NewGRPCServer(addr string, ctrl *controller.FileServiceHandler, logger *zap.Logger) (*GRPCServer, error) {
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
		logger:   logger,
	}, nil
}
