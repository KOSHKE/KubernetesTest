package app

import (
	"context"
	"fmt"
	"net"
	"time"

	"order-service/internal/app/services"
	"order-service/internal/domain/models"
	ordergrpc "order-service/internal/infra/grpc"
	"order-service/internal/infra/kafka"
	"order-service/internal/infra/repository"
	events "proto-go/events"

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

	// Kafka producer (best-effort)
	if brokers := getEnv("KAFKA_BROKERS", "kafka:9092"); brokers != "" {
		if prod, err := kafka.NewProducer(brokers, "orders.v1.order_created"); err == nil {
			defer prod.Close()
			orderService = orderService.WithPublisher(prod)
		} else {
			log.Warnw("kafka producer init failed", "error", err)
		}
		// Kafka consumer for payments.v1.payment_processed
		if cons, err := kafka.NewConsumer(brokers, "order-service", kafka.PaymentProcessedHandlerFunc(func(cctx context.Context, evt *events.PaymentProcessed) error {
			status := models.OrderStatusConfirmed
			if !evt.Success {
				status = models.OrderStatusCancelled
			}
			_, _ = orderService.UpdateOrderStatus(cctx, &services.UpdateOrderStatusRequest{OrderID: evt.OrderId, Status: status})
			return nil
		})); err == nil {
			defer cons.Close()
			go cons.Run(ctx, "payments.v1.payment_processed")
		}
	}

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
