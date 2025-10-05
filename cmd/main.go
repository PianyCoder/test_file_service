package main

import (
	"github.com/PianyCoder/test_file_service/internal/server"
	"log"
)

func main() {
	cfg := server.Config{
		ListenAddr: ":50051",
		StorageDir: "./storage_data",
	}

	grpcServer, err := server.NewGrpcServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create gRPC server: %v", err)
	}

	grpcServer.Start()

	grpcServer.WaitForShutdown()
}
