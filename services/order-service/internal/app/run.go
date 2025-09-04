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
	appsvc "ecommerce-platform/services/order-service/internal/application/services"
	"ecommerce-platform/services/order-service/internal/domain/ports/repository"
	domainservices "ecommerce-platform/services/order-service/internal/domain/services"
	"ecommerce-platform/services/order-service/internal/infra/consumer/handlers"
	orderGrpc "ecommerce-platform/services/order-service/internal/infra/grpc"
	"ecommerce-platform/services/order-service/internal/infra/migration"
	publisher "ecommerce-platform/services/order-service/internal/infra/publisher"
	orderRepoImpl "ecommerce-platform/services/order-service/internal/infra/repository"
	ordermetrics "ecommerce-platform/services/order-service/internal/metrics"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"ecommerce-platform/services/order-service/internal/infra/consumer"
)

// Component represents a component that can be started and stopped
type Component interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Application represents the order service application
type Application struct {
	config *config.OrderConfig
	logger logger.Logger
	ctx    context.Context
	cancel context.CancelFunc

	// Infrastructure components
	db             *gorm.DB
	orderRepo      repository.OrderRepository
	orderPublisher *publisher.OrderCreatedPublisher

	// Business logic components
	orderApplicationService *appsvc.OrderApplicationService

	// Metrics
	metrics *ordermetrics.OrderPrometheusMetrics

	// Servers
	grpcServer    *grpc.Server
	metricsServer *metrics.MetricsServer

	// Components for graceful shutdown
	components []Component
	closers    []io.Closer
}

// NewApplication creates a new order service application
func NewApplication(cfg *config.OrderConfig, logger logger.Logger) *Application {
	ctx, cancel := context.WithCancel(context.Background())
	return &Application{
		config:     cfg,
		logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
		components: make([]Component, 0),
		closers:    make([]io.Closer, 0),
	}
}

// Run starts the order service application
func Run(ctx context.Context, cfg *config.OrderConfig, logger logger.Logger) error {
	app := NewApplication(cfg, logger)
	defer app.cleanup()

	if err := app.initialize(); err != nil {
		return fmt.Errorf("failed to initialize application: %w", err)
	}

	if err := app.start(); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	return app.waitForShutdown(ctx)
}

// initialize sets up all application components
func (app *Application) initialize() error {
	// Initialize metrics
	app.metricsServer = metrics.NewMetricsServer(":"+app.config.MetricsPort, app.logger)
	app.logger.Info("metrics initialized")

	// Initialize gRPC server
	app.grpcServer = grpc.NewServer()

	// Initialize infrastructure components
	if err := app.initializeInfrastructure(); err != nil {
		return fmt.Errorf("failed to initialize infrastructure: %w", err)
	}

	// Initialize business logic
	if err := app.initializeBusinessLogic(); err != nil {
		return fmt.Errorf("failed to initialize business logic: %w", err)
	}

	// Initialize Kafka components AFTER business logic
	if len(app.config.Kafka.Brokers) > 0 {
		if err := app.initializeKafka(); err != nil {
			return fmt.Errorf("failed to initialize Kafka: %w", err)
		}
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

// initializeInfrastructure sets up infrastructure components
func (app *Application) initializeInfrastructure() error {
	// Initialize database
	if err := app.initializeDatabase(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize inventory service connection
	if err := app.initializeInventoryService(); err != nil {
		app.logger.Warn("inventory service initialization failed", "error", err)
	}

	return nil
}

// initializeDatabase sets up database connection
func (app *Application) initializeDatabase() error {
	dsn := app.config.GetDatabaseDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Auto migrate if enabled
	if app.config.Database.AutoMigrate {
		if err := app.runMigrations(db); err != nil {
			return fmt.Errorf("failed to run database migrations: %w", err)
		}
		app.logger.Info("database schema migrated successfully")
	}

	app.db = db
	app.logger.Info("database initialized successfully")
	return nil
}

// runMigrations runs database migrations
func (app *Application) runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&migration.OrderRecord{},
		&migration.OrderItemRecord{},
	)
}

// initializeKafka sets up Kafka producers and consumers
func (app *Application) initializeKafka() error {
	// Initialize Kafka publishers
	if err := app.initializePublishers(); err != nil {
		return fmt.Errorf("failed to initialize publishers: %w", err)
	}

	// Initialize Kafka consumers
	if err := app.initializeConsumers(); err != nil {
		return fmt.Errorf("failed to initialize consumers: %w", err)
	}

	return nil
}

// initializePublishers sets up Kafka publishers
func (app *Application) initializePublishers() error {
	// Initialize order created publisher
	orderCreatedPublisher, err := publisher.NewOrderCreatedPublisher(
		app.config.GetKafkaBrokers(),
		"orders.v1.order_created",
	)
	if err != nil {
		app.logger.Warn("failed to create order created publisher", "error", err)
		app.orderPublisher = nil
	} else {
		app.orderPublisher = orderCreatedPublisher
		app.logger.Info("order created publisher initialized successfully")
	}

	return nil
}

// initializeInventoryService sets up inventory service connection
func (app *Application) initializeInventoryService() error {
	if app.config.Inventory.URL == "" {
		app.logger.Info("inventory service not configured")
		return nil
	}

	app.logger.Info("inventory service connection initialized", "url", app.config.Inventory.URL)
	return nil
}

// initializeConsumers sets up Kafka consumers
func (app *Application) initializeConsumers() error {
	if app.orderRepo == nil {
		return fmt.Errorf("orderRepo is not initialized - cannot create consumers")
	}

	app.logger.Info("creating Kafka consumers with orderRepo", "orderRepo_available", app.orderRepo != nil)

	// Initialize payment processed consumer
	if err := app.initializePaymentConsumer(); err != nil {
		app.logger.Warn("failed to initialize payment consumer", "error", err)
	}

	// Initialize stock reserved consumer
	if err := app.initializeStockConsumer(); err != nil {
		app.logger.Warn("failed to initialize stock consumer", "error", err)
	}

	return nil
}

// initializePaymentConsumer initializes payment processed consumer
func (app *Application) initializePaymentConsumer() error {
	// Create domain service for order business logic
	orderDomainService := domainservices.NewOrderDomainService(app.orderRepo, app.logger)

	// Create application service for payment event processing
	paymentEventService := appsvc.NewPaymentEventService(orderDomainService, app.logger)

	// Create infrastructure handler
	paymentHandler := handlers.NewPaymentProcessedHandler(paymentEventService, app.logger)

	// Create and start payment consumer
	paymentConsumer, err := consumer.NewPaymentProcessedConsumer(
		app.config.GetKafkaBrokers(),
		"order-service-payment",
		"earliest",
		paymentHandler,
	)
	if err != nil {
		return fmt.Errorf("failed to create payment consumer: %w", err)
	}

	app.closers = append(app.closers, paymentConsumer)

	// Start consumer in background
	go func() {
		if err := paymentConsumer.Run(app.ctx, []string{"payments.v1.payment_processed"}); err != nil {
			app.logger.Error("payment consumer failed", "error", err)
		}
	}()
	app.logger.Info("payment consumer started")

	return nil
}

// initializeStockConsumer initializes stock reserved consumer
func (app *Application) initializeStockConsumer() error {
	// Create domain service for order business logic
	orderDomainService := domainservices.NewOrderDomainService(app.orderRepo, app.logger)

	// Create application service for stock event processing
	stockEventService := appsvc.NewStockEventService(orderDomainService, app.logger)

	// Create infrastructure handler
	stockHandler := handlers.NewStockReservedHandler(stockEventService, app.logger)

	// Create and start stock consumer
	stockConsumer, err := consumer.NewStockReservedConsumer(
		app.config.GetKafkaBrokers(),
		"order-service-stock",
		"earliest",
		stockHandler,
	)
	if err != nil {
		return fmt.Errorf("failed to create stock consumer: %w", err)
	}

	app.closers = append(app.closers, stockConsumer)

	// Start consumer in background
	go func() {
		if err := stockConsumer.Run(app.ctx, []string{"inventory.v1.stock_reserved"}); err != nil {
			app.logger.Error("stock consumer failed", "error", err)
		}
	}()
	app.logger.Info("stock consumer started")

	return nil
}

// initializeBusinessLogic sets up business logic components
func (app *Application) initializeBusinessLogic() error {
	// Create order repository
	app.orderRepo = orderRepoImpl.NewGormOrderRepository(app.db)
	app.logger.Info("order repository initialized")

	// Inventory client is not used in current implementation
	// inventoryClient, err := inventoryImpl.NewInventoryClient(app.config.Inventory.URL, app.logger)
	// if err != nil {
	// 	app.logger.Warn("failed to create inventory client", "error", err)
	// 	inventoryClient = nil
	// }

	// Initialize metrics
	app.metrics = ordermetrics.NewOrderMetrics().(*ordermetrics.OrderPrometheusMetrics)
	app.logger.Info("metrics initialized")

	// Create order application service
	app.orderApplicationService = appsvc.NewOrderApplicationService(
		app.orderRepo,
		app.orderPublisher,
		app.logger,
		app.metrics,
	)
	app.logger.Info("order application service initialized")

	return nil
}

// registerGRPCServices registers gRPC services with the server
func (app *Application) registerGRPCServices() {
	// Register gRPC server using the initialized order application service
	orderGrpc.RegisterOrderPBServer(app.grpcServer, app.orderApplicationService)

	app.logger.Info("gRPC order service registered")
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

	// Add to closers for graceful shutdown - MetricsServer now implements io.Closer directly
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
