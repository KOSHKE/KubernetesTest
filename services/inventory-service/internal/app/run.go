package app

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/pkg/metrics"
	appsvc "ecommerce-platform/services/inventory-service/internal/application/services"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/repository"
	"ecommerce-platform/services/inventory-service/internal/infra/consumer"
	"ecommerce-platform/services/inventory-service/internal/infra/migration"
	"ecommerce-platform/services/inventory-service/internal/infra/publisher"
	repositoryimpl "ecommerce-platform/services/inventory-service/internal/infra/repository"
	inventorymetrics "ecommerce-platform/services/inventory-service/internal/metrics"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Component represents a component that can be started and stopped
type Component interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Application represents the inventory service application
type Application struct {
	config *Config
	logger logger.Logger
	ctx    context.Context
	cancel context.CancelFunc

	// Infrastructure components
	db             *gorm.DB
	inventoryRepo  repository.InventoryRepository
	stockPublisher *publisher.StockEventsPublisher

	// Business logic components
	inventoryApplicationService *appsvc.InventoryApplicationService

	// Metrics
	metrics *inventorymetrics.InventoryMetrics

	// Servers
	grpcServer    *grpc.Server
	metricsServer *metrics.MetricsServer

	// Components for graceful shutdown
	components []Component
	closers    []io.Closer
}

// NewApplication creates a new inventory service application
func NewApplication(cfg *Config, logger logger.Logger) *Application {
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

// Run starts the inventory service application
func Run(ctx context.Context, cfg *Config, logger logger.Logger) error {
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
	app.metricsServer = metrics.NewMetricsServer(":"+app.config.Inventory.MetricsPort, app.logger)
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
	migrationService := migration.NewMigrationService(db)

	// Run migrations
	if err := migrationService.RunMigrations(app.ctx); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Seed data
	if err := migrationService.SeedData(app.ctx); err != nil {
		return fmt.Errorf("failed to seed data: %w", err)
	}

	return nil
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
	// Initialize stock events publisher
	stockPublisher, err := publisher.NewStockEventsPublisher(
		app.config.GetKafkaBrokers(),
		"inventory.v1.stock_reserved",
		"inventory.v1.stock_reservation_failed",
	)
	if err != nil {
		app.logger.Warn("failed to create stock events publisher", "error", err)
		app.stockPublisher = nil
	} else {
		app.stockPublisher = stockPublisher
		app.logger.Info("stock events publisher initialized successfully")
	}

	return nil
}

// initializeConsumers sets up Kafka consumers
func (app *Application) initializeConsumers() error {
	if app.inventoryRepo == nil {
		return fmt.Errorf("inventoryRepo is not initialized - cannot create consumers")
	}

	app.logger.Info("creating Kafka consumers with inventoryRepo", "inventoryRepo_available", app.inventoryRepo != nil)

	// Initialize order created consumer
	if err := app.initializeOrderCreatedConsumer(); err != nil {
		app.logger.Warn("failed to initialize order created consumer", "error", err)
	}

	// Initialize payment processed consumer
	if err := app.initializePaymentProcessedConsumer(); err != nil {
		app.logger.Warn("failed to initialize payment processed consumer", "error", err)
	}

	return nil
}

// initializeOrderCreatedConsumer initializes order created consumer
func (app *Application) initializeOrderCreatedConsumer() error {
	// Create domain service for inventory business logic
	domainService := app.inventoryApplicationService.GetDomainService()

	// Create and start order created consumer
	orderConsumer, err := consumer.NewOrderCreatedConsumer(
		app.config.GetKafkaBrokers(),
		"inventory-service-orders",
		"earliest",
		domainService,
	)
	if err != nil {
		return fmt.Errorf("failed to create order created consumer: %w", err)
	}

	app.closers = append(app.closers, orderConsumer)

	// Start consumer in background
	go func() {
		if err := orderConsumer.Run(app.ctx, []string{"orders.v1.order_created"}); err != nil {
			app.logger.Error("order created consumer failed", "error", err)
		}
	}()
	app.logger.Info("order created consumer started")

	return nil
}

// initializePaymentProcessedConsumer initializes payment processed consumer
func (app *Application) initializePaymentProcessedConsumer() error {
	// Create domain service for inventory business logic
	domainService := app.inventoryApplicationService.GetDomainService()

	// Create and start payment processed consumer
	paymentConsumer, err := consumer.NewPaymentProcessedConsumer(
		app.config.GetKafkaBrokers(),
		"inventory-service-payments",
		"earliest",
		domainService,
	)
	if err != nil {
		return fmt.Errorf("failed to create payment processed consumer: %w", err)
	}

	app.closers = append(app.closers, paymentConsumer)

	// Start consumer in background
	go func() {
		if err := paymentConsumer.Run(app.ctx, []string{"payments.v1.payment_processed"}); err != nil {
			app.logger.Error("payment processed consumer failed", "error", err)
		}
	}()
	app.logger.Info("payment processed consumer started")

	return nil
}

// initializeBusinessLogic sets up business logic components
func (app *Application) initializeBusinessLogic() error {
	// Create inventory repository
	app.inventoryRepo = repositoryimpl.NewGormInventoryRepository(app.db)
	app.logger.Info("inventory repository initialized")

	// Initialize metrics
	app.metrics = inventorymetrics.NewInventoryMetrics()
	app.logger.Info("metrics initialized")

	// Create inventory application service
	app.inventoryApplicationService = appsvc.NewInventoryApplicationService(
		app.inventoryRepo,
		app.stockPublisher,
		app.logger,
	)
	app.logger.Info("inventory application service initialized")

	return nil
}

// registerGRPCServices registers gRPC services with the server
func (app *Application) registerGRPCServices() {
	// Register gRPC server using the initialized inventory application service
	// Note: This would need to be implemented based on the actual proto service registration
	// inventorygrpc.RegisterInventoryServiceServer(app.grpcServer, app.inventoryApplicationService)

	app.logger.Info("gRPC inventory service registered")
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
		app.logger.Info("metrics server starting", "port", app.config.Inventory.MetricsPort)
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
	lis, err := net.Listen("tcp", ":"+app.config.Inventory.Port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", app.config.Inventory.Port, err)
	}

	app.logger.Info("gRPC server starting", "port", app.config.Inventory.Port)

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
