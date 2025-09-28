package main

import (
	"context"
	"os/signal"
	"syscall"

	"ecommerce-platform/pkg/config"
	"ecommerce-platform/services/payment-service/internal/server"

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

	log.Info("starting payment-service",
		zap.String("version", Version),
		zap.String("commit", Commit),
		zap.String("build_date", BuildDate),
	)

	cfg, err := config.LoadPaymentConfig()
	if err != nil {
		log.Fatal("failed to load configuration", zap.Error(err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()

	srv, err := server.New(cfg, log)
	if err != nil {
		log.Fatal("failed to build server", zap.Error(err))
	}

	if err := srv.Run(ctx); err != nil {
		log.Fatal("server terminated with error", zap.Error(err))
	}

	log.Info("payment-service stopped gracefully")
}
