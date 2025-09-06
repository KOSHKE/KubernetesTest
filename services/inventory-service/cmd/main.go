package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"ecommerce-platform/pkg/config"
	"ecommerce-platform/services/inventory-service/internal/server"

	"go.uber.org/zap"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

func main() {
	log, err := zap.NewProduction()
	if err != nil {
		panic("failed to init logger: " + err.Error())
	}
	defer log.Sync()

	log.Info("starting inventory-service",
		zap.String("version", Version),
		zap.String("commit", Commit),
		zap.String("build_date", BuildDate),
	)

	cfg, err := config.LoadInventoryConfig()
	if err != nil {
		log.Fatal("failed to load configuration", zap.Error(err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()

	// General timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	srv, err := server.New(cfg, log)
	if err != nil {
		log.Fatal("failed to build server", zap.Error(err))
	}

	if err := srv.Run(ctx); err != nil {
		log.Fatal("server terminated with error", zap.Error(err))
	}

	log.Info("inventory-service stopped gracefully")
}
