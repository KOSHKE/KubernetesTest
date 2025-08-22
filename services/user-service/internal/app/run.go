package app

import (
	"context"
	"fmt"
	"net"

	"user-service/internal/app/services"
	"user-service/internal/infra/auth"
	grpcsvc "user-service/internal/infra/grpc"
	"user-service/internal/infra/repository"

	"go.uber.org/zap"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Run wires dependencies, starts gRPC server and blocks until context cancellation.
func Run(ctx context.Context, cfg *Config, logger *zap.Logger) error {
	log := logger.Sugar()

	// Connect DB
	db, err := connectDB(cfg)
	if err != nil {
		log.Errorw("db connect failed", "error", err)
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Errorw("db pool obtain failed", "error", err)
		return err
	}
	defer sqlDB.Close()

	// Repo + Service
	userRepo := repository.NewGormUserRepository(db)
	if getEnv("AUTO_MIGRATE", "") == "true" {
		if err := userRepo.AutoMigrate(); err != nil {
			log.Errorw("automigrate failed", "error", err)
			return err
		}
	}

	// Initialize JWT auth service
	authConfig := &auth.Config{
		AccessTokenSecret:  cfg.JWTSecret,
		RefreshTokenSecret: cfg.JWTRefreshSecret,
		AccessTokenTTL:     cfg.AccessTokenTTL,
		RefreshTokenTTL:    cfg.RefreshTokenTTL,
		RedisURL:           cfg.RedisURL,
	}

	authService, err := auth.NewJWTAuthService(authConfig, userRepo)
	if err != nil {
		log.Errorw("failed to initialize auth service", "error", err)
		return err
	}
	defer authService.Close()

	userService := services.NewUserService(userRepo, authService)

	// gRPC server
	server := gogrpc.NewServer()
	grpcsvc.RegisterUserPBServer(server, userService)

	// Health
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Listen
	lis, err := netListen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Errorw("listen failed", "error", err)
		return err
	}
	log.Infow("user-service starting", "port", cfg.Port)

	// Serve
	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(lis) }()

	select {
	case <-ctx.Done():
		log.Infow("shutting down user-service...")
		server.GracefulStop()
		<-serveErr
		return nil
	case err := <-serveErr:
		log.Errorw("serve failed", "error", err)
		return err
	}
}

// connectDB creates DB connection and verifies it.
func connectDB(cfg *Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		return nil, err
	}
	if sqlDB, err := db.DB(); err != nil {
		return nil, err
	} else if err := sqlDB.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

// netListen is a small indirection for testability.
var netListen = func(network, address string) (net.Listener, error) {
	return net.Listen(network, address)
}
