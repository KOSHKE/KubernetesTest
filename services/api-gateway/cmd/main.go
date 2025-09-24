package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"ecommerce-platform/services/api-gateway/internal/app"
	"ecommerce-platform/services/api-gateway/internal/config"

	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load configuration", zap.Error(err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()

	if err := app.Run(ctx, cfg, logger); err != nil {
		logger.Fatal("failed to run api-gateway", zap.Error(err))
	}
}
