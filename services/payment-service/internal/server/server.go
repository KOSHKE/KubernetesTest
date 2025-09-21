package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"sync/atomic"
	"time"

	"ecommerce-platform/pkg/config"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/pkg/metrics"
	appsvc "ecommerce-platform/services/payment-service/internal/application/services"
	"ecommerce-platform/services/payment-service/internal/domain/ports/consumer"
	"ecommerce-platform/services/payment-service/internal/domain/ports/publisher"
	infraconsumer "ecommerce-platform/services/payment-service/internal/infra/consumer"
	paymentgrpc "ecommerce-platform/services/payment-service/internal/infra/grpc"
	publisherimpl "ecommerce-platform/services/payment-service/internal/infra/publisher"
	"ecommerce-platform/services/payment-service/internal/infra/repository"
	paymentmetrics "ecommerce-platform/services/payment-service/internal/metrics"

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
	cfg        *config.PaymentConfig
	log        *zap.Logger
	grpcServer *grpc.Server
	httpSrv    *http.Server
	pprofSrv   *http.Server
	ready      atomic.Bool

	// database
	db *gorm.DB

	// business deps
	eventPublisher publisher.PaymentEventsPublisher
	paymentSvc     *appsvc.PaymentApplicationService

	// consumer manager
	consumerManager consumer.EventConsumerManager

	// metrics
	pm *metrics.MetricsServer
}

func New(cfg *config.PaymentConfig, log *zap.Logger) (*Server, error) {
	if cfg.Port == "" || cfg.MetricsPort == "" {
		return nil, errors.New("empty ports in config")
	}

	// Connect to database
	db, err := connectDatabase(cfg)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	// Auto-migrate database if enabled
	if cfg.Database.AutoMigrate {
		if err := autoMigrateDatabase(db); err != nil {
			return nil, fmt.Errorf("auto-migrate database: %w", err)
		}
	}

	// Create outbox repository
	outboxRepo := repository.NewOutboxRepository(db)

	// Create logger adapter
	loggerAdapter := logger.NewZapLogger(log.Sugar())
	pm := metrics.NewMetricsServer(":"+cfg.MetricsPort, loggerAdapter)

	// Create consumer manager
	consumerManager := infraconsumer.NewConsumerManager(loggerAdapter)

	// Initialize event publisher
	eventPublisher, err := initPublisher(cfg, loggerAdapter)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize event publisher: %w", err)
	}

	// gRPC server
	gs := grpc.NewServer()
	paymentAPI := paymentgrpc.NewPaymentServerImpl()
	paymentMetrics := paymentmetrics.NewPaymentMetrics()
	paymentSvc := appsvc.NewPaymentApplicationService(outboxRepo, loggerAdapter, paymentMetrics)
	paymentAPI.RegisterPaymentService(gs, paymentSvc)

	hs := health.NewServer()
	healthpb.RegisterHealthServer(gs, hs)
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	// reflection only for non-prod if needed
	reflection.Register(gs)

	s := &Server{
		cfg:             cfg,
		log:             log,
		db:              db,
		grpcServer:      gs,
		pm:              pm,
		eventPublisher:  eventPublisher,
		paymentSvc:      paymentSvc,
		consumerManager: consumerManager,
	}

	// Initialize consumers if Kafka is configured
	if len(cfg.Kafka.Brokers) > 0 {
		if err := s.initConsumers(cfg, loggerAdapter); err != nil {
			return nil, fmt.Errorf("failed to initialize consumers: %w", err)
		}
	}

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

func (s *Server) initConsumers(cfg *config.PaymentConfig, logger logger.Logger) error {
	ctx := context.Background()

	// Initialize stock reserved consumer
	stockReservedConfig := consumer.ConsumerConfig{
		BootstrapServers: cfg.Kafka.Brokers,
		GroupID:          "payment-service-stock-reserved",
		AutoOffsetReset:  "earliest",
		Topics:           []string{"inventory.v1.stock_reserved"},
	}

	if err := s.consumerManager.StartStockReservedConsumer(ctx, stockReservedConfig, s.paymentSvc); err != nil {
		logger.Warn("failed to initialize stock reserved consumer", "error", err)
	}

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

	// Close consumer manager
	if s.consumerManager != nil {
		if err := s.consumerManager.Close(); err != nil {
			s.log.Error("consumer manager close error", zap.Error(err))
			if firstErr == nil {
				firstErr = err
			}
		}
	}

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

	// close database connection
	if s.db != nil {
		if sqlDB, err := s.db.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				s.log.Error("database close error", zap.Error(err))
				if firstErr == nil {
					firstErr = err
				}
			}
		}
	}

	// close publisher and other resources
	if s.eventPublisher != nil {
		if err := s.eventPublisher.Close(); err != nil {
			s.log.Error("event publisher close error", zap.Error(err))
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}

// connectDatabase connects to the PostgreSQL database
func connectDatabase(cfg *config.PaymentConfig) (*gorm.DB, error) {
	dsn := cfg.GetDatabaseDSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// autoMigrateDatabase runs database migrations
func autoMigrateDatabase(db *gorm.DB) error {
	// Import the outbox models for auto-migration
	if err := db.AutoMigrate(&struct {
		ID          uint `gorm:"primaryKey"`
		CreatedAt   time.Time
		UpdatedAt   time.Time
		DeletedAt   gorm.DeletedAt `gorm:"index"`
		AggregateID string         `gorm:"not null"`
		Type        string         `gorm:"not null"`
		Payload     string         `gorm:"type:jsonb"`
		Processed   bool           `gorm:"default:false"`
		Error       string
	}{}); err != nil {
		return fmt.Errorf("failed to auto-migrate outbox table: %w", err)
	}

	return nil
}
