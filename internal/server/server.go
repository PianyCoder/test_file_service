package server

import (
	"fmt"
	"github.com/PianyCoder/test_file_service/internal/service"
	"github.com/PianyCoder/test_file_service/internal/storage"

	"github.com/PianyCoder/test_file_service/internal/controller"
	pb "github.com/PianyCoder/test_file_service/internal/proto"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"google.golang.org/grpc"
)

const (
	DefaultListenAddr = ":50051"
	DefaultStorageDir = "./storage_data"
	DefaultChunkSize  = 1024 * 1024

	UploadDownloadLimit = 10
	ListLimit           = 100
)

type Config struct {
	ListenAddr    string
	StorageDir    string
	ChunkSize     int
	UploadLimit   int64
	DownloadLimit int64
	ListLimit     int64
}

func (c *Config) SetDefaults() {
	if c.ListenAddr == "" {
		c.ListenAddr = DefaultListenAddr
	}
	if c.StorageDir == "" {
		c.StorageDir = DefaultStorageDir
	}
	if c.ChunkSize <= 0 {
		c.ChunkSize = DefaultChunkSize
	}
	if c.UploadLimit <= 0 {
		c.UploadLimit = UploadDownloadLimit
	}
	if c.DownloadLimit <= 0 {
		c.DownloadLimit = UploadDownloadLimit
	}
	if c.ListLimit <= 0 {
		c.ListLimit = ListLimit
	}
}

type GrpcServer struct {
	cfg        Config
	grpcServer *grpc.Server
	lis        net.Listener
	wg         sync.WaitGroup
}

func NewGrpcServer(cfg Config) (*GrpcServer, error) {
	cfg.SetDefaults()

	fileStorage, err := storage.NewFileStorage(cfg.StorageDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize file storage: %w", err)
	}
	log.Printf("File storage initialized at: %s", cfg.StorageDir)

	fileUseCase := service.NewFileService(
		fileStorage,
		cfg.UploadLimit,
		cfg.DownloadLimit,
		cfg.ListLimit,
		cfg.StorageDir,
		cfg.ChunkSize,
	)
	log.Printf("File use case initialized with limits: Upload/Download=%d, List=%d", cfg.UploadLimit, cfg.ListLimit)

	grpcServer := grpc.NewServer()
	fileServiceHandler := controller.NewFileServiceHandler(fileUseCase)
	pb.RegisterFileServiceServer(grpcServer, fileServiceHandler)

	lis, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", cfg.ListenAddr, err)
	}
	log.Printf("gRPC server configured to listen on %s", cfg.ListenAddr)

	return &GrpcServer{
		cfg:        cfg,
		grpcServer: grpcServer,
		lis:        lis,
	}, nil
}

func (s *GrpcServer) Start() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		log.Println("gRPC server started successfully.")
		if serveErr := s.grpcServer.Serve(s.lis); serveErr != nil {
			log.Fatalf("Failed to serve gRPC: %v", serveErr)
		}
	}()
}

func (s *GrpcServer) Shutdown() {
	log.Println("Shutting down gRPC server...")
	s.grpcServer.GracefulStop()
	s.wg.Wait()
	log.Println("gRPC server stopped.")
}

func (s *GrpcServer) WaitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	s.Shutdown()
}
