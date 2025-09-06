package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"strings"
	"sync/atomic"
	"time"

	"ecommerce-platform/pkg/config"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/pkg/metrics"
	appsvc "ecommerce-platform/services/inventory-service/internal/application/services"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/publisher"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/repository"
	domainservices "ecommerce-platform/services/inventory-service/internal/domain/services"
	"ecommerce-platform/services/inventory-service/internal/infra/consumer"
	"ecommerce-platform/services/inventory-service/internal/infra/consumer/handlers"
	inventoryGrpc "ecommerce-platform/services/inventory-service/internal/infra/grpc"
	"ecommerce-platform/services/inventory-service/internal/infra/migration"
	inventoryPublisher "ecommerce-platform/services/inventory-service/internal/infra/publisher"
	productRepoImpl "ecommerce-platform/services/inventory-service/internal/infra/repository"
	inventoryMetrics "ecommerce-platform/services/inventory-service/internal/metrics"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Server struct {
	cfg        *config.InventoryConfig
	log        *zap.Logger
	grpcServer *grpc.Server
	httpSrv    *http.Server
	pprofSrv   *http.Server
	ready      atomic.Bool

	// database
	db *gorm.DB

	// business dependencies
	inventoryRepo  repository.InventoryRepository
	stockPublisher publisher.StockEventsPublisher
	inventorySvc   *appsvc.InventoryApplicationService

	// metrics
	pm      *metrics.MetricsServer
	metrics inventoryMetrics.InventoryMetrics

	// Kafka consumers
	consumers []interface{ Close() error }
}

func New(cfg *config.InventoryConfig, log *zap.Logger) (*Server, error) {
	if cfg.Port == "" || cfg.MetricsPort == "" {
		return nil, errors.New("empty ports in config")
	}

	// Create logger adapter
	loggerAdapter := logger.NewZapLogger(log.Sugar())

	// Initialize database
	db, err := initDatabase(cfg, loggerAdapter)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Run migrations
	if err := runMigrations(db, loggerAdapter); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Initialize inventory repository
	inventoryRepo := productRepoImpl.NewInventoryRepository(db)

	// Initialize publisher
	stockPublisher, err := initPublisher(cfg, loggerAdapter)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize publisher: %w", err)
	}

	// Initialize metrics
	inventoryMetrics := inventoryMetrics.NewInventoryMetrics()
	pm := metrics.NewMetricsServer(":"+cfg.MetricsPort, loggerAdapter)

	// Initialize inventory application service
	inventorySvc := appsvc.NewInventoryApplicationService(inventoryRepo, stockPublisher, loggerAdapter)

	// Initialize gRPC server
	gs := grpc.NewServer()
	inventoryGrpc.RegisterInventoryServer(gs, inventorySvc, loggerAdapter)

	// Setup health checks
	hs := health.NewServer()
	healthpb.RegisterHealthServer(gs, hs)
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	// Setup reflection for development
	reflection.Register(gs)

	s := &Server{
		cfg:            cfg,
		log:            log,
		grpcServer:     gs,
		pm:             pm,
		db:             db,
		inventoryRepo:  inventoryRepo,
		stockPublisher: stockPublisher,
		inventorySvc:   inventorySvc,
		metrics:        inventoryMetrics,
		consumers:      make([]interface{ Close() error }, 0),
	}

	// Initialize Kafka consumers if configured
	if len(cfg.Kafka.Brokers) > 0 {
		if err := s.initConsumers(cfg, loggerAdapter); err != nil {
			return nil, fmt.Errorf("failed to initialize consumers: %w", err)
		}
	}

	return s, nil
}

func initDatabase(cfg *config.InventoryConfig, logger logger.Logger) (*gorm.DB, error) {
	logger.Info("initializing database connection")

	dsn := cfg.GetDatabaseDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("database connection established")
	return db, nil
}

func runMigrations(db *gorm.DB, logger logger.Logger) error {
	logger.Info("running database migrations")

	// Auto migrate if enabled
	if err := db.AutoMigrate(
		&migration.ProductRecord{},
		&migration.StockRecord{},
	); err != nil {
		return fmt.Errorf("failed to run database migrations: %w", err)
	}

	logger.Info("database migrations completed")
	return nil
}

func initPublisher(cfg *config.InventoryConfig, logger logger.Logger) (publisher.StockEventsPublisher, error) {
	logger.Info("initializing stock events publisher")

	stockPublisher, err := inventoryPublisher.NewStockEventsPublisher(
		cfg.Kafka.Brokers,
		"inventory.v1.stock_reserved",
		"inventory.v1.stock_reservation_failed",
		"inventory.v1.stock_released",
		"inventory.v1.stock_committed",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create stock events publisher: %w", err)
	}

	logger.Info("stock events publisher initialized")
	return stockPublisher, nil
}

func (s *Server) initConsumers(cfg *config.InventoryConfig, logger logger.Logger) error {
	logger.Info("initializing Kafka consumers")

	// Initialize order created consumer
	if err := s.initOrderConsumer(cfg, logger); err != nil {
		logger.Warn("failed to initialize order consumer", "error", err)
	}

	// Initialize payment processed consumer
	if err := s.initPaymentConsumer(cfg, logger); err != nil {
		logger.Warn("failed to initialize payment consumer", "error", err)
	}

	return nil
}

func (s *Server) initOrderConsumer(cfg *config.InventoryConfig, logger logger.Logger) error {
	// Create domain service for inventory business logic
	inventoryDomainService := domainservices.NewInventoryDomainService(s.inventoryRepo, logger)

	// Create infrastructure handler
	orderHandler := handlers.NewOrderCreatedHandler(inventoryDomainService, logger)

	// Create and start order consumer
	orderConsumer, err := consumer.NewOrderCreatedConsumer(
		strings.Join(cfg.Kafka.Brokers, ","),
		"inventory-service-orders",
		"earliest",
		orderHandler,
	)
	if err != nil {
		return fmt.Errorf("failed to create order consumer: %w", err)
	}

	s.consumers = append(s.consumers, orderConsumer)

	// Start consumer in background
	go func() {
		if err := orderConsumer.Run(context.Background(), []string{"orders.v1.order_created"}); err != nil {
			s.log.Error("order consumer failed", zap.Error(err))
		}
	}()
	s.log.Info("order consumer started")

	return nil
}

func (s *Server) initPaymentConsumer(cfg *config.InventoryConfig, logger logger.Logger) error {
	// Create domain service for inventory business logic
	inventoryDomainService := domainservices.NewInventoryDomainService(s.inventoryRepo, logger)

	// Create infrastructure handler
	paymentHandler := handlers.NewPaymentProcessedHandler(inventoryDomainService, logger)

	// Create and start payment consumer
	paymentConsumer, err := consumer.NewPaymentProcessedConsumer(
		strings.Join(cfg.Kafka.Brokers, ","),
		"inventory-service-payments",
		"earliest",
		paymentHandler,
	)
	if err != nil {
		return fmt.Errorf("failed to create payment consumer: %w", err)
	}

	s.consumers = append(s.consumers, paymentConsumer)

	// Start consumer in background
	go func() {
		if err := paymentConsumer.Run(context.Background(), []string{"payments.v1.payment_processed"}); err != nil {
			s.log.Error("payment consumer failed", zap.Error(err))
		}
	}()
	s.log.Info("payment consumer started")

	return nil
}

func (s *Server) Run(ctx context.Context) error {
	// HTTP mux: /metrics, /healthz, /readyz
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		if s.ready.Load() {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, "not ready", http.StatusServiceUnavailable)
	})

	s.httpSrv = &http.Server{
		Addr:              ":" + s.cfg.MetricsPort, // shared port for health/metrics
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// pprof (optional: on localhost:6060)
	s.pprofSrv = &http.Server{Addr: "localhost:6060"}

	errCh := make(chan error, 3)

	// start HTTP (metrics/health)
	go func() {
		s.log.Info("http metrics/health starting", zap.String("addr", s.httpSrv.Addr))
		if err := s.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("http metrics/health: %w", err)
		}
	}()

	// start pprof
	go func() {
		s.log.Info("pprof starting", zap.String("addr", s.pprofSrv.Addr))
		if err := s.pprofSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("pprof: %w", err)
		}
	}()

	// start gRPC
	go func() {
		addr := ":" + s.cfg.Port
		s.log.Info("grpc starting", zap.String("addr", addr))
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			errCh <- fmt.Errorf("grpc listen: %w", err)
			return
		}
		if err := s.grpcServer.Serve(lis); err != nil {
			errCh <- fmt.Errorf("grpc serve: %w", err)
		}
	}()

	// simulate warmup and only then readiness=true
	time.AfterFunc(500*time.Millisecond, func() { s.ready.Store(true) })

	// wait for completion
	select {
	case <-ctx.Done():
		s.log.Info("shutdown signal received")
		shCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return s.shutdown(shCtx)
	case err := <-errCh:
		return err
	}
}

func (s *Server) shutdown(ctx context.Context) error {
	var firstErr error

	// gRPC graceful
	done := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(done)
	}()
	select {
	case <-done:
		s.log.Info("grpc stopped gracefully")
	case <-time.After(5 * time.Second):
		s.log.Warn("grpc graceful timeout, forcing stop")
		s.grpcServer.Stop()
	}

	// HTTP servers
	if s.httpSrv != nil {
		if err := s.httpSrv.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.log.Warn("http shutdown error", zap.Error(err))
			firstErr = err
		}
	}
	if s.pprofSrv != nil {
		if err := s.pprofSrv.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.log.Warn("pprof shutdown error", zap.Error(err))
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	// Close Kafka consumers
	for _, consumer := range s.consumers {
		if err := consumer.Close(); err != nil {
			s.log.Warn("consumer close error", zap.Error(err))
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	// Close publisher
	if closer, ok := s.stockPublisher.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			s.log.Warn("publisher close error", zap.Error(err))
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	// Close database
	if s.db != nil {
		if sqlDB, err := s.db.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				s.log.Warn("database close error", zap.Error(err))
				if firstErr == nil {
					firstErr = err
				}
			}
		}
	}

	return firstErr
}
