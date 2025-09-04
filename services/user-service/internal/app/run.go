package app

import (
	"context"
	"fmt"
	"io"
	"net"

	"time"

	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/pkg/metrics"

	appsvc "ecommerce-platform/services/user-service/internal/application/services"
	authPorts "ecommerce-platform/services/user-service/internal/domain/ports/auth"
	"ecommerce-platform/services/user-service/internal/domain/ports/repository"
	"ecommerce-platform/services/user-service/internal/infra/auth"
	userGrpc "ecommerce-platform/services/user-service/internal/infra/grpc"
	"ecommerce-platform/services/user-service/internal/infra/migration"
	userRepoImpl "ecommerce-platform/services/user-service/internal/infra/repository"
	usermetrics "ecommerce-platform/services/user-service/internal/metrics"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Application represents the user service application
type Application struct {
	config *Config
	logger logger.Logger
	ctx    context.Context
	cancel context.CancelFunc

	// Infrastructure components
	db          *gorm.DB
	userRepo    repository.UserRepository
	authService authPorts.AuthService

	// Business logic components
	userService *appsvc.UserApplicationService

	// Metrics
	metrics *usermetrics.UserPrometheusMetrics

	// Servers
	grpcServer    *grpc.Server
	metricsServer *metrics.MetricsServer

	// Components for graceful shutdown
	closers []io.Closer
}

// NewApplication creates a new user service application
func NewApplication(cfg *Config, logger logger.Logger) *Application {
	ctx, cancel := context.WithCancel(context.Background())
	return &Application{
		config:  cfg,
		logger:  logger,
		closers: make([]io.Closer, 0),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Run starts the user service application
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

	// Register gRPC services
	app.registerGRPCServices()

	// Setup health checks
	app.setupHealthChecks()

	// Setup reflection for development
	if app.config.IsDevelopment() {
		reflection.Register(app.grpcServer)
	}

	app.logger.Info("application initialized successfully")
	return nil
}

// initializeInfrastructure sets up database and other infrastructure components
func (app *Application) initializeInfrastructure() error {
	// Initialize database
	if err := app.initializeDatabase(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize migration service
	if err := app.initializeMigrationService(); err != nil {
		return fmt.Errorf("failed to initialize migration service: %w", err)
	}

	return nil
}

// initializeDatabase sets up database connection
func (app *Application) initializeDatabase() error {
	app.logger.Info("initializing database connection")

	dsn := app.config.GetDatabaseDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	app.db = db
	app.logger.Info("database connection established")
	return nil
}

// initializeMigrationService sets up and runs database migrations
func (app *Application) initializeMigrationService() error {
	app.logger.Info("initializing migration service")

	migrationService := migration.NewMigrationService(app.db)
	if err := migrationService.Migrate(context.Background()); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	app.logger.Info("database migrations completed")
	return nil
}

// initializeBusinessLogic sets up business logic components
func (app *Application) initializeBusinessLogic() error {
	// Initialize repository
	if err := app.initializeRepository(); err != nil {
		return fmt.Errorf("failed to initialize repository: %w", err)
	}

	// Initialize auth service
	if err := app.initializeAuthService(); err != nil {
		return fmt.Errorf("failed to initialize auth service: %w", err)
	}

	// Initialize metrics
	app.metrics = usermetrics.NewUserMetrics().(*usermetrics.UserPrometheusMetrics)
	app.logger.Info("business metrics initialized")

	// Create user application service
	app.userService = appsvc.NewUserApplicationService(
		app.userRepo,
		app.authService,
		app.logger,
		app.metrics,
	)
	app.logger.Info("user application service initialized")

	return nil
}

// initializeRepository sets up user repository
func (app *Application) initializeRepository() error {
	app.logger.Info("initializing user repository")

	userRepo := userRepoImpl.NewGormUserRepository(app.db)
	app.userRepo = userRepo

	app.logger.Info("user repository initialized")
	return nil
}

// initializeAuthService sets up authentication service
func (app *Application) initializeAuthService() error {
	app.logger.Info("initializing authentication service")

	authConfig := auth.Config{
		AccessTokenSecret:  app.config.Auth.AccessTokenSecret,
		RefreshTokenSecret: app.config.Auth.RefreshTokenSecret,
		AccessTokenTTL:     app.config.Auth.AccessTokenTTL,
		RefreshTokenTTL:    app.config.Auth.RefreshTokenTTL,
		StorageURL:         app.config.Auth.StorageURL,
		StorageTimeout:     app.config.Auth.StorageTimeout,
	}

	authService, err := auth.NewJWTAuthService(&authConfig, app.userRepo, app.logger)
	if err != nil {
		return fmt.Errorf("failed to create JWT auth service: %w", err)
	}

	app.authService = authService
	app.logger.Info("authentication service initialized")
	return nil
}

// registerGRPCServices registers gRPC services with the server
func (app *Application) registerGRPCServices() {
	// Register gRPC server using the initialized user application service
	userGrpc.RegisterUserPBServer(app.grpcServer, app.userService)
	app.logger.Info("gRPC user service registered")
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
	app.logger.Info("waiting for shutdown signal")

	// Wait for shutdown signal
	<-ctx.Done()
	app.logger.Info("shutdown signal received, starting graceful shutdown")

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
