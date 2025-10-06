package main

import (
	"context"
	"github.com/PianyCoder/test_file_service/internal/app"
	"log"
)

func main() {
	ctx := context.Background()
	if err := app.Start(ctx); err != nil {
		log.Fatalf("start app error: %v", err)
	}
}
