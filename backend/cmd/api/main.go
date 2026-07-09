// Package main is the entry point for the TrAIveler backend API server.
//
// This file initializes the application, loads configuration, establishes
// database connections, and starts the HTTP server.
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JosemaPereira/capstone-project-ai-bootcamp/backend/config"
	"github.com/JosemaPereira/capstone-project-ai-bootcamp/backend/internal/database"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("starting TrAIveler backend API",
		"version", "0.1.0",
		"environment", getEnv("GO_ENV", "development"),
	)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}
	slog.Info("configuration loaded successfully")

	// Initialize database client
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var dbClient database.Client
	dbClient, err = database.NewClient(ctx, cfg.Database.URL)
	if err != nil {
		log.Fatalf("failed to initialize database client: %v", err)
	}
	defer func() {
		if err := dbClient.Close(); err != nil {
			slog.Error("error closing database connection", "error", err)
		}
	}()

	slog.Info("database connection established",
		"max_connections", cfg.Database.MaxConnections,
		"min_connections", cfg.Database.MinConnections,
	)

	// Verify database health
	if err := dbClient.Ping(ctx); err != nil {
		log.Fatalf("database health check failed: %v", err)
	}
	slog.Info("database health check passed")

	// TODO: Initialize HTTP router and middleware
	// TODO: Register route handlers
	// TODO: Start HTTP server

	// For now, just demonstrate successful startup
	fmt.Println("✅ Backend initialization complete!")
	fmt.Printf("📊 Database: connected (min: %d, max: %d connections)\n",
		cfg.Database.MinConnections,
		cfg.Database.MaxConnections,
	)
	fmt.Printf("🚀 Ready to start HTTP server on port %d (TODO)\n", cfg.Server.Port)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down gracefully...")
	fmt.Println("👋 Goodbye!")
}

// getEnv retrieves an environment variable with a fallback default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
