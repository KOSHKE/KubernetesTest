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
	appsvc "ecommerce-platform/services/payment-service/internal/application/services"
	"ecommerce-platform/services/payment-service/internal/domain/ports/publisher"
	infraconsumer "ecommerce-platform/services/payment-service/internal/infra/consumer"
	"ecommerce-platform/services/payment-service/internal/infra/consumer/handlers"
	paymentgrpc "ecommerce-platform/services/payment-service/internal/infra/grpc"
	publisherimpl "ecommerce-platform/services/payment-service/internal/infra/publisher"
	paymentmetrics "ecommerce-platform/services/payment-service/internal/metrics"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	cfg        *config.PaymentConfig
	log        *zap.Logger
	grpcServer *grpc.Server
	httpSrv    *http.Server
	pprofSrv   *http.Server
	ready      atomic.Bool

	// business deps
	paymentProcessedPub publisher.PaymentProcessedPublisher
	paymentSvc          *appsvc.PaymentApplicationService

	// consumers
	consumers []interface{ Close() error }

	// metrics
	pm *metrics.MetricsServer
}

func New(cfg *config.PaymentConfig, log *zap.Logger) (*Server, error) {
	if cfg.Port == "" || cfg.MetricsPort == "" {
		return nil, errors.New("empty ports in config")
	}

	// dependencies
	pub, err := publisherimpl.NewPaymentProcessedPublisher(cfg.GetKafkaBrokers(), "payments.v1.payment_processed")
	if err != nil {
		return nil, fmt.Errorf("build publisher: %w", err)
	}

	// Create logger adapter
	loggerAdapter := logger.NewZapLogger(log.Sugar())
	pm := metrics.NewMetricsServer(":"+cfg.MetricsPort, loggerAdapter)

	// gRPC server
	gs := grpc.NewServer()
	paymentAPI := paymentgrpc.NewPaymentServerImpl()
	paymentMetrics := paymentmetrics.NewPaymentMetrics()
	paymentSvc := appsvc.NewPaymentApplicationService(pub, loggerAdapter, paymentMetrics)
	paymentAPI.RegisterPaymentService(gs, paymentSvc)

	hs := health.NewServer()
	healthpb.RegisterHealthServer(gs, hs)
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	// reflection only for non-prod if needed
	reflection.Register(gs)

	s := &Server{
		cfg:                 cfg,
		log:                 log,
		grpcServer:          gs,
		pm:                  pm,
		paymentProcessedPub: pub,
		paymentSvc:          paymentSvc,
		consumers:           make([]interface{ Close() error }, 0),
	}

	// Initialize consumers if Kafka is configured
	if len(cfg.Kafka.Brokers) > 0 {
		if err := s.initConsumers(cfg, loggerAdapter); err != nil {
			return nil, fmt.Errorf("failed to initialize consumers: %w", err)
		}
	}

	return s, nil
}

func (s *Server) initConsumers(cfg *config.PaymentConfig, logger logger.Logger) error {
	// Initialize stock released consumer
	if err := s.initStockReleasedConsumer(cfg, logger); err != nil {
		logger.Warn("failed to initialize stock released consumer", "error", err)
	}

	return nil
}

func (s *Server) initStockReleasedConsumer(cfg *config.PaymentConfig, logger logger.Logger) error {
	// Create application service
	stockEventService := appsvc.NewStockEventService(s.paymentSvc.GetProcessPaymentUseCase(), logger)

	// Create infrastructure handler
	stockReleasedHandler := handlers.NewStockReleasedHandler(stockEventService, logger)

	// Create and start stock released consumer
	stockReleasedConsumer, err := infraconsumer.NewStockReleasedConsumer(
		strings.Join(cfg.Kafka.Brokers, ","),
		"payment-service-stock-released",
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

	// Close consumers
	for _, consumer := range s.consumers {
		if err := consumer.Close(); err != nil {
			s.log.Error("consumer close error", zap.Error(err))
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

	// close publisher and other resources
	if closer, ok := any(s.paymentProcessedPub).(interface{ Close() error }); ok {
		_ = closer.Close()
	}

	return firstErr
}
