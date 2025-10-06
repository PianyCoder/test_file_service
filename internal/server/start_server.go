package server

import (
	"context"
	"fmt"
	"github.com/PianyCoder/test_file_service/infrastructure/logger"
)

func (s *GRPCServer) StartServer(ctx context.Context) error {
	l := logger.FromContext(ctx)
	if l == nil && s.logger != nil {
		l = s.logger.Sugar()
	}

	l.Infof("gRPC listening on %s", s.addr)

	serveErr := make(chan error, 1)
	go func() {
		if err := s.server.Serve(s.listener); err != nil {
			serveErr <- err
			return
		}
		serveErr <- nil
	}()

	select {
	case <-ctx.Done():
		l.Info("Context canceled — stopping gRPC server")
		s.server.GracefulStop()
		return nil
	case err := <-serveErr:
		if err != nil {
			l.Errorf("gRPC Serve error: %v", err)
			return fmt.Errorf("grpc serve error: %w", err)
		}
		return nil
	}
}
