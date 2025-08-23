package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/kubernetestest/ecommerce-platform/services/api-gateway/internal/app"
	"github.com/kubernetestest/ecommerce-platform/services/api-gateway/internal/config"

	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()

	if err := app.Run(ctx, cfg, logger); err != nil {
		os.Exit(1)
	}
}
