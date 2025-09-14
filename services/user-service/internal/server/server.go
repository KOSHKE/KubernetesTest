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
	"ecommerce-platform/pkg/jwt"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/pkg/metrics"
	"ecommerce-platform/pkg/redisclient"
	appsvc "ecommerce-platform/services/user-service/internal/application/services"
	"ecommerce-platform/services/user-service/internal/domain/ports/repository"
	"ecommerce-platform/services/user-service/internal/domain/ports/services"
	userGrpc "ecommerce-platform/services/user-service/internal/infra/grpc"
	"ecommerce-platform/services/user-service/internal/infra/migration"
	userRepoImpl "ecommerce-platform/services/user-service/internal/infra/repository"
	infraServices "ecommerce-platform/services/user-service/internal/infra/services"
	usermetrics "ecommerce-platform/services/user-service/internal/metrics"

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
	cfg        *config.UserConfig
	log        *zap.Logger
	grpcServer *grpc.Server
	httpSrv    *http.Server
	pprofSrv   *http.Server
	ready      atomic.Bool

	// database
	db *gorm.DB

	// business dependencies
	userRepo       repository.UserRepository
	sessionRepo    repository.SessionRepository
	tokenGenerator services.TokenGenerator
	userSvc        *appsvc.UserApplicationService

	// metrics
	pm      *metrics.MetricsServer
	metrics usermetrics.UserMetrics
}

func New(cfg *config.UserConfig, log *zap.Logger) (*Server, error) {
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

	// Initialize repositories
	userRepo := userRepoImpl.NewGormUserRepository(db)

	// Initialize Redis client for session storage
	redisClient := redisclient.New(cfg.Auth.StorageURL, "", 0) // No password, default DB
	sessionRepo := userRepoImpl.NewRedisSessionRepository(redisClient, cfg.Auth.RefreshTokenTTL)

	// Initialize JWT manager
	jwtConfig := jwt.Config{
		AccessTokenSecret:  cfg.Auth.AccessTokenSecret,
		RefreshTokenSecret: cfg.Auth.RefreshTokenSecret,
		AccessTokenTTL:     cfg.Auth.AccessTokenTTL,
		RefreshTokenTTL:    cfg.Auth.RefreshTokenTTL,
		Issuer:             "user-service",
		Audience:           "ecommerce-platform",
	}
	jwtManager := jwt.NewManager(jwtConfig)

	// Initialize token generator
	tokenGenerator := infraServices.NewJWTTokenGenerator(jwtManager)

	// Initialize metrics
	userMetrics := usermetrics.NewUserMetrics()
	pm := metrics.NewMetricsServer(":"+cfg.MetricsPort, loggerAdapter)

	// Initialize user application service
	userSvc := appsvc.NewUserApplicationService(userRepo, sessionRepo, tokenGenerator, loggerAdapter, userMetrics)

	// Initialize gRPC server
	gs := grpc.NewServer()
	userGrpc.RegisterUserPBServer(gs, userSvc, userMetrics)

	// Setup health checks
	hs := health.NewServer()
	healthpb.RegisterHealthServer(gs, hs)
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	// Setup reflection for development
	if cfg.IsDevelopment() {
		reflection.Register(gs)
	}

	s := &Server{
		cfg:            cfg,
		log:            log,
		grpcServer:     gs,
		pm:             pm,
		db:             db,
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		tokenGenerator: tokenGenerator,
		userSvc:        userSvc,
		metrics:        userMetrics,
	}
	return s, nil
}

func initDatabase(cfg *config.UserConfig, logger logger.Logger) (*gorm.DB, error) {
	logger.Info("initializing database connection")

	dsn := cfg.GetDatabaseDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	logger.Info("database connection established")
	return db, nil
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
		Addr:              ":" + s.cfg.MetricsPort, // общий порт для health/metrics
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// pprof (опционально: на localhost:6060)
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

	// имитация прогрева и только потом readiness=true
	time.AfterFunc(500*time.Millisecond, func() { s.ready.Store(true) })

	// ожидание завершения
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

	// Close token generator
	if closer, ok := s.tokenGenerator.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			s.log.Warn("token generator close error", zap.Error(err))
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
