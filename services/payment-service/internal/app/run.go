package app

import (
	"context"
	"fmt"

	"ecommerce-platform/pkg/config"
	"ecommerce-platform/pkg/logger"
)

// Run starts the payment service application
func Run(ctx context.Context, cfg *config.PaymentConfig, logger logger.Logger) error {
	builder := NewBuilder(cfg, logger)

	// Build all dependencies
	paymentProcessedPub, err := builder.BuildPaymentProcessedPublisher()
	if err != nil {
		return fmt.Errorf("failed to build payment processed publisher: %w", err)
	}

	paymentGrpcServer := builder.BuildPaymentGrpcServer()
	metrics := builder.BuildMetrics()
	metricsServer := builder.BuildMetricsServer()
	grpcServer := builder.BuildGrpcServer()

	app := NewApplication(
		cfg,
		logger,
		paymentProcessedPub,
		paymentGrpcServer,
		metrics,
		metricsServer,
		grpcServer,
	)
	defer app.cleanup()

	if err := app.initialize(); err != nil {
		return fmt.Errorf("failed to initialize application: %w", err)
	}

	if err := app.start(); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	return app.waitForShutdown(ctx)
}
