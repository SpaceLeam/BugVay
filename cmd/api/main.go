package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kokuroshesh/bugvay/internal/api"
	"github.com/kokuroshesh/bugvay/internal/config"
	"github.com/kokuroshesh/bugvay/internal/database"
	"github.com/kokuroshesh/bugvay/internal/queue"
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

	// Initialize queue client
	queueClient := queue.NewClient(&cfg.Redis)
	defer queueClient.Close()
	log.Println("✓ Connected to Redis (Asynq)")

	// Initialize API router
	router := api.NewRouter(pg, ch, queueClient)

	// Graceful shutdown
	go func() {
		addr := fmt.Sprintf("%s:%s", cfg.API.Host, cfg.API.Port)
		log.Printf("🚀 API server starting on %s", addr)
		if err := router.Run(addr); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
}
