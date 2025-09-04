package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/services/inventory-service/internal/app"
)

func main() {
	// Create logger
	logger, err := logger.NewZapLogger()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}

	// Load configuration
	cfg := app.LoadConfigFromEnv()
	logger.Info("configuration loaded", "port", cfg.Inventory.Port, "metricsPort", cfg.Inventory.MetricsPort)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// Start shutdown goroutine
	go func() {
		sig := <-sigChan
		logger.Info("received signal, shutting down", "signal", sig)
		cancel()
	}()

	// Run application
	if err := app.Run(ctx, cfg, logger); err != nil {
		logger.Error("application failed", "error", err)
		os.Exit(1)
	}

	logger.Info("application stopped")
}
