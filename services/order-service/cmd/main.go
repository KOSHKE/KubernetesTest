package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"ecommerce-platform/pkg/config"
	"ecommerce-platform/pkg/logger"
	app "ecommerce-platform/services/order-service/internal/app"

	"go.uber.org/zap"
)

func main() {
	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()

	// Convert to our unified logger interface
	appLogger := logger.NewZapLogger(zapLogger.Sugar())

	cfg, err := config.LoadOrderConfig()
	if err != nil {
		zapLogger.Fatal("failed to load configuration", zap.Error(err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()

	if err := app.Run(ctx, cfg, appLogger); err != nil {
		os.Exit(1)
	}
}
