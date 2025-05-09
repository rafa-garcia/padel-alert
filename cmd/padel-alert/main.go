package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rafa-garcia/padel-alert/internal/api"
	"github.com/rafa-garcia/padel-alert/internal/config"
	"github.com/rafa-garcia/padel-alert/internal/logger"
	"github.com/rafa-garcia/padel-alert/internal/scheduler"
	"github.com/rafa-garcia/padel-alert/internal/storage"
)

const version = "0.3.0"

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

	// Initialize Redis client
	redisClient, err := storage.NewRedisClient(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", err)
	}
	defer redisClient.Close()

	// Create rule and user storage
	ruleStorage := storage.NewRedisRuleStorage(redisClient)
	userStorage := storage.NewRedisUserStorage(redisClient)

	// Initialize scheduler
	sched := scheduler.NewScheduler(cfg, ruleStorage)
	if err := sched.Start(); err != nil {
		logger.Fatal("Failed to start scheduler", err)
	}
	defer sched.Stop()

	// Create router with API keys from config
	r := api.NewRouter(version, cfg.APIKeys, ruleStorage, userStorage)

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
