package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	app "user-service/internal/app"

	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg := app.LoadConfigFromEnv()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()

	if err := app.Run(ctx, cfg, logger); err != nil {
		os.Exit(1)
	}
}
