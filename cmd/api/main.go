package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.API.Host, cfg.API.Port),
		Handler: router.Engine(),
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 API server starting on %s:%s", cfg.API.Host, cfg.API.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
