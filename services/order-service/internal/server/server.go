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
	appservices "ecommerce-platform/services/order-service/internal/application/services"
	appsvc "ecommerce-platform/services/order-service/internal/application/services"
	"ecommerce-platform/services/order-service/internal/domain/ports/publisher"
	"ecommerce-platform/services/order-service/internal/domain/ports/repository"
	domainservices "ecommerce-platform/services/order-service/internal/domain/services"
	infraconsumer "ecommerce-platform/services/order-service/internal/infra/consumer"
	"ecommerce-platform/services/order-service/internal/infra/consumer/handlers"
	orderGrpc "ecommerce-platform/services/order-service/internal/infra/grpc"
	"ecommerce-platform/services/order-service/internal/infra/migration"
	orderPublisher "ecommerce-platform/services/order-service/internal/infra/publisher"
	orderRepoImpl "ecommerce-platform/services/order-service/internal/infra/repository"
	ordermetrics "ecommerce-platform/services/order-service/internal/metrics"

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
	cfg        *config.OrderConfig
	log        *zap.Logger
	grpcServer *grpc.Server
	httpSrv    *http.Server
	pprofSrv   *http.Server
	ready      atomic.Bool

	// database
	db *gorm.DB

	// business dependencies
	orderRepo      repository.OrderRepository
	orderPublisher publisher.OrderCreatedPublisher
	orderSvc       *appsvc.OrderApplicationService

	// metrics
	pm      *metrics.MetricsServer
	metrics ordermetrics.OrderMetrics

	// Kafka consumers
	consumers []interface{ Close() error }
}

func New(cfg *config.OrderConfig, log *zap.Logger) (*Server, error) {
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

	// Initialize repository
	orderRepo := orderRepoImpl.NewGormOrderRepository(db)

	// Initialize publisher
	orderPublisher, err := initPublisher(cfg, loggerAdapter)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize publisher: %w", err)
	}

	// Initialize metrics
	orderMetrics := ordermetrics.NewOrderMetrics()
	pm := metrics.NewMetricsServer(":"+cfg.MetricsPort, loggerAdapter)

	// Initialize order application service
	orderSvc := appsvc.NewOrderApplicationService(orderRepo, orderPublisher, loggerAdapter, orderMetrics)

	// Initialize gRPC server
	gs := grpc.NewServer()
	orderGrpc.RegisterOrderPBServer(gs, orderSvc)

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
		orderRepo:      orderRepo,
		orderPublisher: orderPublisher,
		orderSvc:       orderSvc,
		metrics:        orderMetrics,
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

func initDatabase(cfg *config.OrderConfig, logger logger.Logger) (*gorm.DB, error) {
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
		&migration.OrderRecord{},
		&migration.OrderItemRecord{},
	); err != nil {
		return fmt.Errorf("failed to run database migrations: %w", err)
	}

	logger.Info("database migrations completed")
	return nil
}

func initPublisher(cfg *config.OrderConfig, logger logger.Logger) (publisher.OrderCreatedPublisher, error) {
	logger.Info("initializing order created publisher")

	orderCreatedPublisher, err := orderPublisher.NewOrderCreatedPublisher(
		cfg.GetKafkaBrokers(),
		"orders.v1.order_created",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create order created publisher: %w", err)
	}

	logger.Info("order created publisher initialized")
	return orderCreatedPublisher, nil
}

func (s *Server) initConsumers(cfg *config.OrderConfig, logger logger.Logger) error {
	logger.Info("initializing Kafka consumers")

	// Initialize payment processed consumer
	if err := s.initPaymentConsumer(cfg, logger); err != nil {
		logger.Warn("failed to initialize payment consumer", "error", err)
	}

	// Initialize stock reserved consumer
	if err := s.initStockConsumer(cfg, logger); err != nil {
		logger.Warn("failed to initialize stock consumer", "error", err)
	}

	return nil
}

func (s *Server) initPaymentConsumer(cfg *config.OrderConfig, logger logger.Logger) error {
	// Create domain service for order business logic
	orderDomainService := domainservices.NewOrderDomainService(s.orderRepo, logger)

	// Create application service for payment event processing
	paymentEventService := appsvc.NewPaymentEventService(orderDomainService, logger)

	// Create infrastructure handler
	paymentHandler := handlers.NewPaymentProcessedHandler(paymentEventService, logger)

	// Create and start payment consumer
	paymentConsumer, err := infraconsumer.NewPaymentProcessedConsumer(
		cfg.GetKafkaBrokers(),
		"order-service-payment",
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

func (s *Server) initStockConsumer(cfg *config.OrderConfig, logger logger.Logger) error {
	// Create domain service for order business logic
	orderDomainService := domainservices.NewOrderDomainService(s.orderRepo, logger)

	// Create application service for stock event processing
	stockEventService := appsvc.NewStockEventService(orderDomainService, logger)

	// Create infrastructure handler
	stockHandler := handlers.NewStockReservedHandler(stockEventService, logger)

	// Create and start stock consumer
	stockConsumer, err := infraconsumer.NewStockReservedConsumer(
		cfg.GetKafkaBrokers(),
		"order-service-stock",
		"earliest",
		stockHandler,
	)
	if err != nil {
		return fmt.Errorf("failed to create stock consumer: %w", err)
	}

	s.consumers = append(s.consumers, stockConsumer)

	// Start consumer in background
	go func() {
		if err := stockConsumer.Run(context.Background(), []string{"inventory.v1.stock_reserved"}); err != nil {
			s.log.Error("stock consumer failed", zap.Error(err))
		}
	}()
	s.log.Info("stock consumer started")

	// Initialize stock reservation failed consumer
	if err := s.initStockReservationFailedConsumer(cfg, logger); err != nil {
		logger.Warn("failed to initialize stock reservation failed consumer", "error", err)
	}

	// Initialize stock released consumer
	if err := s.initStockReleasedConsumer(cfg, logger); err != nil {
		logger.Warn("failed to initialize stock released consumer", "error", err)
	}

	// Initialize stock committed consumer
	if err := s.initStockCommittedConsumer(cfg, logger); err != nil {
		logger.Warn("failed to initialize stock committed consumer", "error", err)
	}

	return nil
}

func (s *Server) initStockReservationFailedConsumer(cfg *config.OrderConfig, logger logger.Logger) error {
	// Create domain service for order business logic
	orderDomainService := domainservices.NewOrderDomainService(s.orderRepo, logger)

	// Create application service
	stockEventService := appservices.NewStockEventService(orderDomainService, logger)

	// Create infrastructure handler
	stockReservationFailedHandler := handlers.NewStockReservationFailedHandler(stockEventService, logger)

	// Create and start stock reservation failed consumer
	stockReservationFailedConsumer, err := infraconsumer.NewStockReservationFailedConsumer(
		strings.Join(cfg.Kafka.Brokers, ","),
		"order-service-stock-failed",
		"earliest",
		stockReservationFailedHandler,
	)
	if err != nil {
		return fmt.Errorf("failed to create stock reservation failed consumer: %w", err)
	}

	s.consumers = append(s.consumers, stockReservationFailedConsumer)

	// Start consumer in background
	go func() {
		if err := stockReservationFailedConsumer.Run(context.Background(), []string{"inventory.v1.stock_reservation_failed"}); err != nil {
			s.log.Error("stock reservation failed consumer failed", zap.Error(err))
		}
	}()
	s.log.Info("stock reservation failed consumer started")

	return nil
}

func (s *Server) initStockReleasedConsumer(cfg *config.OrderConfig, logger logger.Logger) error {
	// Create domain service for order business logic
	orderDomainService := domainservices.NewOrderDomainService(s.orderRepo, logger)

	// Create application service
	stockEventService := appservices.NewStockEventService(orderDomainService, logger)

	// Create infrastructure handler
	stockReleasedHandler := handlers.NewStockReleasedHandler(stockEventService, logger)

	// Create and start stock released consumer
	stockReleasedConsumer, err := infraconsumer.NewStockReleasedConsumer(
		strings.Join(cfg.Kafka.Brokers, ","),
		"order-service-stock-released",
		"earliest",
		stockReleasedHandler,
	)
	if err != nil {
		return fmt.Errorf("failed to create stock released consumer: %w", err)
	}

	s.consumers = append(s.consumers, stockReleasedConsumer)

	// Start consumer in background
	go func() {
		if err := stockReleasedConsumer.Run(context.Background(), []string{"inventory.v1.stock_released"}); err != nil {
			s.log.Error("stock released consumer failed", zap.Error(err))
		}
	}()
	s.log.Info("stock released consumer started")

	return nil
}

func (s *Server) initStockCommittedConsumer(cfg *config.OrderConfig, logger logger.Logger) error {
	// Create domain service for order business logic
	orderDomainService := domainservices.NewOrderDomainService(s.orderRepo, logger)

	// Create application service
	stockEventService := appservices.NewStockEventService(orderDomainService, logger)

	// Create infrastructure handler
	stockCommittedHandler := handlers.NewStockCommittedHandler(stockEventService, logger)

	// Create and start stock committed consumer
	stockCommittedConsumer, err := infraconsumer.NewStockCommittedConsumer(
		strings.Join(cfg.Kafka.Brokers, ","),
		"order-service-stock-committed",
		"earliest",
		stockCommittedHandler,
	)
	if err != nil {
		return fmt.Errorf("failed to create stock committed consumer: %w", err)
	}

	s.consumers = append(s.consumers, stockCommittedConsumer)

	// Start consumer in background
	go func() {
		if err := stockCommittedConsumer.Run(context.Background(), []string{"inventory.v1.stock_committed"}); err != nil {
			s.log.Error("stock committed consumer failed", zap.Error(err))
		}
	}()
	s.log.Info("stock committed consumer started")

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
	if closer, ok := s.orderPublisher.(interface{ Close() error }); ok {
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
