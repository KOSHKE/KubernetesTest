package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"ecommerce-platform/pkg/logger"
	app "ecommerce-platform/services/user-service/internal/app"

	"go.uber.org/zap"
)

func main() {
	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()

	// Convert to our unified logger interface
	logger := logger.NewZapLogger(zapLogger.Sugar())

	cfg, err := app.LoadConfigFromEnv()
	if err != nil {
		zapLogger.Fatal("failed to load configuration", zap.Error(err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()

	if err := app.Run(ctx, cfg, logger); err != nil {
		zapLogger.Fatal("application failed", zap.Error(err))
		os.Exit(1)
	}
}
