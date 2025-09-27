package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"time"

	"ecommerce-platform/pkg/config"
	pkghealth "ecommerce-platform/pkg/health"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/pkg/metrics"
	"ecommerce-platform/pkg/outbox"
	appsvc "ecommerce-platform/services/payment-service/internal/application/services"
	"ecommerce-platform/services/payment-service/internal/domain/ports/consumer"
	"ecommerce-platform/services/payment-service/internal/domain/ports/publisher"
	infraconsumer "ecommerce-platform/services/payment-service/internal/infra/consumer"
	paymentgrpc "ecommerce-platform/services/payment-service/internal/infra/grpc"
	"ecommerce-platform/services/payment-service/internal/infra/migration"
	publisherimpl "ecommerce-platform/services/payment-service/internal/infra/publisher"
	"ecommerce-platform/services/payment-service/internal/infra/repository"
	paymentmetrics "ecommerce-platform/services/payment-service/internal/metrics"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Server struct {
	cfg        *config.PaymentConfig
	log        *zap.Logger
	grpcServer *grpc.Server
	httpSrv    *http.Server
	pprofSrv   *http.Server
	health     *pkghealth.Manager

	// database
	db *gorm.DB

	// business dependencies
	eventPublisher publisher.PaymentEventsPublisher
	paymentSvc     *appsvc.PaymentApplicationService

	// outbox publisher
	outboxPublisher *outbox.BackgroundPublisher

	// metrics
	pm      *metrics.MetricsServer
	metrics paymentmetrics.PaymentMetrics

	// Kafka consumers
	consumerManager consumer.EventConsumerManager
}

func New(cfg *config.PaymentConfig, log *zap.Logger) (*Server, error) {
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

	// Run migrations (consistent with inventory-service)
	if err := runMigrations(db, loggerAdapter); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Create outbox repository
	outboxRepo := repository.NewOutboxRepository(db)

	// Initialize event publisher
	eventPublisher, err := initPublisher(cfg, loggerAdapter)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize event publisher: %w", err)
	}

	// Initialize outbox publisher
	outboxPublisher := outbox.NewBackgroundPublisher(
		outboxRepo,
		eventPublisher,
		loggerAdapter,
		5*time.Second, // interval
		10,            // batch size
	)

	// Initialize metrics
	paymentMetrics := paymentmetrics.NewPaymentMetrics()
	pm := metrics.NewMetricsServer(":"+cfg.MetricsPort, loggerAdapter)

	// Initialize payment application service
	paymentSvc := appsvc.NewPaymentApplicationService(outboxRepo, loggerAdapter, paymentMetrics)

	// Initialize health manager
	healthManager := pkghealth.NewManager()

	// Add database health check
	healthManager.AddChecker(pkghealth.NewDatabaseChecker(db))

	// Create gRPC health checker that syncs with our health manager
	grpcHealthChecker := pkghealth.NewGRPCHealthChecker(healthManager)
	healthManager.AddChecker(grpcHealthChecker)

	// Initialize gRPC server
	gs := grpc.NewServer()
	paymentgrpc.RegisterPaymentServer(gs, paymentSvc, paymentMetrics)

	// Register gRPC health server
	healthpb.RegisterHealthServer(gs, grpcHealthChecker.GetGRPCHealthServer())

	// Setup reflection for development
	reflection.Register(gs)

	s := &Server{
		cfg:             cfg,
		log:             log,
		grpcServer:      gs,
		pm:              pm,
		db:              db,
		eventPublisher:  eventPublisher,
		paymentSvc:      paymentSvc,
		outboxPublisher: outboxPublisher,
		metrics:         paymentMetrics,
		consumerManager: infraconsumer.NewConsumerManager(loggerAdapter),
		health:          healthManager,
	}

	// Initialize Kafka consumers if configured
	if len(cfg.Kafka.Brokers) > 0 {
		if err := s.initConsumers(cfg, loggerAdapter); err != nil {
			return nil, fmt.Errorf("failed to initialize consumers: %w", err)
		}

		// Add consumer health checks
		s.health.AddChecker(pkghealth.NewConsumerChecker("order", func() bool {
			return s.consumerManager != nil
		}))
	}

	// Add outbox health check
	s.health.AddChecker(pkghealth.NewOutboxChecker(func() bool {
		return s.outboxPublisher != nil
	}))

	// All components initialized successfully - check health status
	s.health.IsHealthy(context.Background())

	return s, nil
}

func initPublisher(cfg *config.PaymentConfig, logger logger.Logger) (publisher.PaymentEventsPublisher, error) {
	logger.Info("initializing payment events publisher")

	topics := map[string]string{
		"PaymentProcessed": "payments.v1.payment_processed",
	}

	eventPublisher, err := publisherimpl.NewPaymentEventsPublisher(
		cfg.Kafka.Brokers,
		topics,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment events publisher: %w", err)
	}

	logger.Info("payment events publisher initialized")
	return eventPublisher, nil
}

func runMigrations(db *gorm.DB, logger logger.Logger) error {
	logger.Info("running database migrations")

	migrationService := migration.NewMigrationService(db)
	if err := migrationService.Migrate(context.Background()); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info("database migrations completed")
	return nil
}

func (s *Server) initConsumers(cfg *config.PaymentConfig, logger logger.Logger) error {
	logger.Info("initializing Kafka consumers")

	ctx := context.Background()

	// Consumer configuration
	consumerConfig := consumer.ConsumerConfig{
		BootstrapServers: cfg.Kafka.Brokers,
		AutoOffsetReset:  "earliest",
	}

	// Initialize critical stock reserved consumer - failure should stop service
	stockReservedConfig := consumerConfig
	stockReservedConfig.GroupID = "payment-service-stock-reserved"
	stockReservedConfig.Topics = []string{"inventory.v1.stock_reserved"}

	if err := s.consumerManager.StartStockReservedConsumer(ctx, stockReservedConfig, s.paymentSvc); err != nil {
		return fmt.Errorf("failed to start critical stock reserved consumer: %w", err)
	}
	logger.Info("stock reserved consumer started successfully")

	return nil
}

func (s *Server) Run(ctx context.Context) error {
	// HTTP mux: /metrics, /healthz, /readyz
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		w.Header().Set("Content-Type", "application/json")

		if s.health.IsReady(ctx) {
			w.WriteHeader(http.StatusOK)
			response := map[string]interface{}{
				"status": "ready",
				"checks": make(map[string]string),
			}

			// Get status of all components for complete info
			status := s.health.GetStatus(ctx)
			for name, err := range status {
				if err != nil {
					response["checks"].(map[string]string)[name] = err.Error()
				} else {
					response["checks"].(map[string]string)[name] = "ok"
				}
			}

			json.NewEncoder(w).Encode(response)
			return
		}

		// Return detailed status for debugging
		w.WriteHeader(http.StatusServiceUnavailable)
		response := map[string]interface{}{
			"status": "not ready",
			"checks": make(map[string]string),
		}

		status := s.health.GetStatus(ctx)
		for name, err := range status {
			if err != nil {
				response["checks"].(map[string]string)[name] = err.Error()
			} else {
				response["checks"].(map[string]string)[name] = "ok"
			}
		}

		json.NewEncoder(w).Encode(response)
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

	// start outbox publisher
	go func() {
		s.log.Info("outbox publisher starting")
		s.outboxPublisher.Start(ctx)
	}()

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
	if err := s.consumerManager.Close(); err != nil {
		s.log.Warn("consumer close error", zap.Error(err))
		if firstErr == nil {
			firstErr = err
		}
	}

	// Close outbox publisher
	if s.outboxPublisher != nil {
		if err := s.outboxPublisher.Close(); err != nil {
			s.log.Warn("outbox publisher close error", zap.Error(err))
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	// Close publisher
	if closer, ok := s.eventPublisher.(interface{ Close() error }); ok {
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

func initDatabase(cfg *config.PaymentConfig, logger logger.Logger) (*gorm.DB, error) {
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

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.Database.MaxConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	logger.Info("database connection established")
	return db, nil
}
