// Package main is the entry point for the TrAIveler backend API server.
//
// This file initializes the application, loads configuration, establishes
// database connections, and starts the HTTP server.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JosemaPereira/TrAIveler/backend/config"
	"github.com/JosemaPereira/TrAIveler/backend/internal/database"
)

// shutdownTimeout bounds how long graceful shutdown waits for in-flight
// requests to complete before the process exits regardless.
const shutdownTimeout = 30 * time.Second

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("starting TrAIveler backend API",
		"version", "0.1.0",
		"environment", getEnv("GO_ENV", "development"),
	)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}
	slog.Info("configuration loaded successfully")

	// NewClient already retries (3 attempts, 2s delay) and pings before
	// returning, so any error here is fatal by construction — no separate
	// health check is needed.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dbClient, err := database.NewClient(ctx, cfg.Database.URL)
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

	apiServer := NewHTTPServer(dbClient, cfg, logger)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      apiServer.Router(),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Buffered so this goroutine can't leak: once Shutdown makes
	// ListenAndServe return, the send below always has room.
	serveErrCh := make(chan error, 1)
	go func() {
		slog.Info("HTTP server listening", "port", cfg.Server.Port)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErrCh <- err
			return
		}
		serveErrCh <- nil
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Race a shutdown signal against an early server failure (e.g. port
	// already in use) so a startup error exits immediately instead of
	// waiting indefinitely for a signal that will never arrive.
	select {
	case sig := <-quit:
		slog.Info("shutting down gracefully...", "signal", sig.String())
	case err := <-serveErrCh:
		log.Fatalf("HTTP server failed: %v", err)
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("error during HTTP server shutdown", "error", err)
	}

	slog.Info("shutdown complete")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
