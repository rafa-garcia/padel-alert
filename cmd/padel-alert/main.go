package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/padel-alert/internal/api"
	"github.com/yourusername/padel-alert/internal/config"
	"github.com/yourusername/padel-alert/internal/logger"
)

const version = "0.1.0"

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger.Init(cfg.LogLevel)
	logger.Info("Starting PadelAlert service", "version", version)

	// Create router with API keys from config
	r := api.NewRouter(version, cfg.APIKeys)

	// Create server with timeouts
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine so it doesn't block the graceful shutdown handling
	go func() {
		logger.Info("Server listening", "port", cfg.Port)
		logger.Info("Health check available", "url", fmt.Sprintf("http://localhost:%s/api/v1/health", cfg.Port))

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server error", err)
		}
	}()

	// Set up channel to listen for signals to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive a signal
	sig := <-quit
	logger.Info("Shutting down server", "signal", sig.String())

	// Create a deadline to wait for current operations to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", err)
	}

	logger.Info("Server gracefully stopped")
}
