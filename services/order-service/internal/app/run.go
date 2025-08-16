package app

import (
	"context"
	"fmt"
	"net"
	"time"

	"order-service/internal/app/services"
	ordergrpc "order-service/internal/infra/grpc"
	"order-service/internal/infra/repository"

	"go.uber.org/zap"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Run(ctx context.Context, cfg *Config, logger *zap.Logger) error {
	log := logger.Sugar()

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

	orderRepo := repository.NewGormOrderRepository(db)
	if getEnv("AUTO_MIGRATE", "") == "true" {
		if err := orderRepo.AutoMigrate(); err != nil {
			log.Errorw("automigrate failed", "error", err)
			return err
		}
	}
	orderService := services.NewOrderService(orderRepo)

	server := gogrpc.NewServer()
	ordergrpc.RegisterOrderPBServer(server, orderService)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Errorw("listen failed", "error", err)
		return err
	}
	log.Infow("order-service starting", "port", cfg.Port)

	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(lis) }()

	select {
	case <-ctx.Done():
		log.Infow("shutting down order-service...")
		server.GracefulStop()
		time.Sleep(100 * time.Millisecond)
		return nil
	case err := <-serveErr:
		log.Errorw("serve failed", "error", err)
		return err
	}
}

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
