package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	app "payment-service/internal/app"

	"go.uber.org/zap"
)

type Config struct {
	Port string
}

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
