package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"time"

	"ecommerce-platform/pkg/config"
	pkghealth "ecommerce-platform/pkg/health"
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
	"encoding/json"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
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
	health     *pkghealth.Manager

	// database
	db *gorm.DB

	// redis
	redisClient *redisclient.Client

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

	// Initialize health manager
	healthManager := pkghealth.NewManager()

	// Add database health check
	healthManager.AddChecker(pkghealth.NewDatabaseChecker(db))

	// Add Redis health check
	healthManager.AddChecker(pkghealth.NewRedisChecker(redisClient))

	// Create gRPC health checker that syncs with our health manager
	grpcHealthChecker := pkghealth.NewGRPCHealthChecker(healthManager)
	healthManager.AddChecker(grpcHealthChecker)

	// Initialize user application service
	userSvc := appsvc.NewUserApplicationService(userRepo, sessionRepo, tokenGenerator, loggerAdapter)

	// Initialize gRPC server
	gs := grpc.NewServer()
	userGrpc.RegisterUserPBServer(gs, userSvc, userMetrics)

	// Register gRPC health server
	healthpb.RegisterHealthServer(gs, grpcHealthChecker.GetGRPCHealthServer())

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
		health:         healthManager,
		redisClient:    redisClient,
	}

	// All components initialized successfully - check health status
	s.health.IsHealthy(context.Background())

	return s, nil
}

func initDatabase(cfg *config.UserConfig, logger logger.Logger) (*gorm.DB, error) {
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

	// Close token generator
	if closer, ok := s.tokenGenerator.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			s.log.Warn("token generator close error", zap.Error(err))
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	// Close Redis client
	if s.redisClient != nil {
		if err := s.redisClient.Close(); err != nil {
			s.log.Warn("redis client close error", zap.Error(err))
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	// Close Redis session repository
	if closer, ok := s.sessionRepo.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			s.log.Warn("session repository close error", zap.Error(err))
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
