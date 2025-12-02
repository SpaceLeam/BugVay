package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kokuroshesh/bugvay/internal/config"
	"github.com/kokuroshesh/bugvay/internal/database"
	"github.com/kokuroshesh/bugvay/internal/worker"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to Postgres
	pg, err := database.NewPostgres(&cfg.Postgres)
	if err != nil {
		log.Fatalf("Failed to connect to Postgres: %v", err)
	}
	defer pg.Close()
	log.Println("✓ Connected to PostgreSQL")

	// Connect to ClickHouse
	ch, err := database.NewClickHouse(&cfg.ClickHouse)
	if err != nil {
		log.Printf("⚠ ClickHouse connection failed: %v (optional for MVP)", err)
	} else {
		defer ch.Close()
		log.Println("✓ Connected to ClickHouse")
	}

	// Initialize worker
	w := worker.NewWorker(cfg, pg, ch)
	defer w.Shutdown()

	// Start worker
	go func() {
		log.Printf("🔧 Worker starting with concurrency=%d", cfg.Worker.Concurrency)
		if err := w.Run(); err != nil {
			log.Fatalf("Worker failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down worker...")
}
