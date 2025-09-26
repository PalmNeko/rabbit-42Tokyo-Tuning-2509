package main

import (
	"backend/internal/server"
	"backend/internal/telemetry"
	"context"
	"log"
)

func main() {
	ctx := context.Background()

	// テレメトリー初期化
	shutdown, err := telemetry.Init(ctx)
	if err != nil {
		log.Printf("Failed to initialize telemetry: %v", err)
	}
	defer shutdown(ctx)

	srv, dbConn, err := server.NewServer()
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}
	if dbConn != nil {
		defer dbConn.Close()
	}

	srv.Run()
}
