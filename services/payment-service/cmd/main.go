package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"ecommerce-platform/pkg/config"
	app "ecommerce-platform/services/payment-service/internal/app"

	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg, err := config.LoadPaymentConfig()
	if err != nil {
		logger.Fatal("failed to load configuration", zap.Error(err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()

	if err := app.Run(ctx, cfg, logger); err != nil {
		os.Exit(1)
	}
}
