package app

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"ecommerce-platform/pkg/config"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/pkg/metrics"
	appsvc "ecommerce-platform/services/payment-service/internal/application/services"
	"ecommerce-platform/services/payment-service/internal/domain/ports/grpc"
	"ecommerce-platform/services/payment-service/internal/domain/ports/publisher"
	paymentmetrics "ecommerce-platform/services/payment-service/internal/metrics"

	grpcLib "google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// Application represents the payment service application
type Application struct {
	config *config.PaymentConfig
	logger logger.Logger
	ctx    context.Context
	cancel context.CancelFunc

	// Dependencies (injected via constructor)
	paymentProcessedPub publisher.PaymentProcessedPublisher
	paymentGrpcServer   grpc.PaymentServer
	metrics             paymentmetrics.PaymentMetrics
	metricsServer       *metrics.MetricsServer
	grpcServer          *grpcLib.Server

	// Business logic components
	paymentApplicationService *appsvc.PaymentApplicationService

	// Components for graceful shutdown
	closers []io.Closer
}

// NewApplication creates a new payment service application
func NewApplication(
	cfg *config.PaymentConfig,
	logger logger.Logger,
	paymentProcessedPub publisher.PaymentProcessedPublisher,
	paymentGrpcServer grpc.PaymentServer,
	metrics paymentmetrics.PaymentMetrics,
	metricsServer *metrics.MetricsServer,
	grpcServer *grpcLib.Server,
) *Application {
	ctx, cancel := context.WithCancel(context.Background())
	return &Application{
		config:              cfg,
		logger:              logger,
		ctx:                 ctx,
		cancel:              cancel,
		paymentProcessedPub: paymentProcessedPub,
		paymentGrpcServer:   paymentGrpcServer,
		metrics:             metrics,
		metricsServer:       metricsServer,
		grpcServer:          grpcServer,
		closers:             make([]io.Closer, 0),
	}
}

// initialize sets up all application components
func (app *Application) initialize() error {
	// Initialize business logic
	if err := app.initializeBusinessLogic(); err != nil {
		return fmt.Errorf("failed to initialize business logic: %w", err)
	}

	// Register gRPC services
	app.registerGRPCServices()

	// Setup health checks
	app.setupHealthChecks()

	// Setup reflection for development
	reflection.Register(app.grpcServer)

	app.logger.Info("application initialized successfully")
	return nil
}

// initializeBusinessLogic sets up business logic components
func (app *Application) initializeBusinessLogic() error {
	// Create payment application service
	app.paymentApplicationService = appsvc.NewPaymentApplicationService(
		app.paymentProcessedPub,
		app.logger,
		app.metrics,
	)
	app.logger.Info("payment application service initialized")

	return nil
}

// registerGRPCServices registers gRPC services with the server
func (app *Application) registerGRPCServices() {
	// Register gRPC server using the injected payment gRPC server
	app.paymentGrpcServer.RegisterPaymentService(app.grpcServer, app.paymentApplicationService)

	app.logger.Info("gRPC payment service registered")
}

// setupHealthChecks configures health check endpoints
func (app *Application) setupHealthChecks() {
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(app.grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
}

// start starts all application components
func (app *Application) start() error {
	// Start metrics server
	if err := app.startMetricsServer(); err != nil {
		return fmt.Errorf("failed to start metrics server: %w", err)
	}

	// Start gRPC server
	if err := app.startGRPCServer(); err != nil {
		return fmt.Errorf("failed to start gRPC server: %w", err)
	}

	app.logger.Info("application started successfully")
	return nil
}

// startMetricsServer starts the metrics HTTP server
func (app *Application) startMetricsServer() error {
	go func() {
		app.logger.Info("metrics server starting", "port", app.config.MetricsPort)
		if err := app.metricsServer.Start(app.ctx); err != nil {
			app.logger.Error("metrics server failed", "error", err)
		}
	}()

	// Add to closers for graceful shutdown
	app.closers = append(app.closers, app.metricsServer)
	return nil
}

// startGRPCServer starts the gRPC server
func (app *Application) startGRPCServer() error {
	lis, err := net.Listen("tcp", ":"+app.config.Port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", app.config.Port, err)
	}

	app.logger.Info("gRPC server starting", "port", app.config.Port)

	go func() {
		if err := app.grpcServer.Serve(lis); err != nil {
			app.logger.Error("gRPC server failed", "error", err)
		}
	}()

	return nil
}

// waitForShutdown waits for shutdown signal and performs graceful shutdown
func (app *Application) waitForShutdown(ctx context.Context) error {
	// Wait for shutdown signal
	<-ctx.Done()
	app.logger.Info("shutdown signal received, starting graceful shutdown...")

	// Perform graceful shutdown
	app.shutdown()

	app.logger.Info("application shutdown completed")
	return nil
}

// shutdown performs graceful shutdown of all components
func (app *Application) shutdown() {
	// Shutdown gRPC server
	if app.grpcServer != nil {
		app.shutdownGRPCServer()
	}

	// Close all closers
	app.closeComponents()

	// Cancel context
	app.cancel()
}

// shutdownGRPCServer gracefully shuts down the gRPC server
func (app *Application) shutdownGRPCServer() {
	done := make(chan struct{})
	go func() {
		app.grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		app.logger.Info("gRPC server gracefully stopped")
	case <-time.After(5 * time.Second):
		app.logger.Warn("gRPC graceful shutdown timeout, forcing stop")
		app.grpcServer.Stop()
	}
}

// closeComponents closes all components
func (app *Application) closeComponents() {
	for _, closer := range app.closers {
		if err := closer.Close(); err != nil {
			app.logger.Warn("failed to close component", "error", err)
		}
	}
}

// cleanup ensures cleanup on exit
func (app *Application) cleanup() {
	app.shutdown()
}
