package main

import (
	"context"
	"github.com/PianyCoder/test_file_service/internal/app"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-stop
		log.Println("Shutdown signal received")
		cancel()
	}()

	if err := app.Start(ctx); err != nil {
		if ctx.Err() == context.Canceled {
			log.Println("Application stopped gracefully")
			return
		}
		log.Fatalf("Application stopped with error: %v", err)
	}
}
