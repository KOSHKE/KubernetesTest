package app

import (
	"context"
	"net"
	"time"

	appsvc "inventory-service/internal/app/services"
	grpcsvr "inventory-service/internal/infra/grpc"
	"inventory-service/internal/infra/kafka"
	repo "inventory-service/internal/infra/repository"
	events "proto-go/events"
	invpb "proto-go/inventory"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func Run(ctx context.Context, cfg *Config, logger *zap.Logger) error {
	log := logger.Sugar()

	repository := repo.NewInMemoryInventoryRepository()
	svc := appsvc.NewInventoryService(repository)

	// Kafka producer for inventory events (best-effort)
	if brokers := getEnv("KAFKA_BROKERS", "kafka:9092"); brokers != "" {
		if p, err := kafka.NewProducer(brokers, "inventory.v1.stock_reserved", "inventory.v1.stock_reservation_failed"); err == nil {
			defer p.Close()
			svc = svc.WithPublisher(p)
		}
	}

	// Kafka consumer for order created (best-effort)
	if brokers := getEnv("KAFKA_BROKERS", "kafka:9092"); brokers != "" {
		handler := kafka.OrderCreatedHandlerFunc(func(cctx context.Context, evt *events.OrderCreated) error {
			items := make([]appsvc.StockCheckItem, 0, len(evt.Items))
			for _, it := range evt.Items {
				items = append(items, appsvc.StockCheckItem{ProductID: it.ProductId, Quantity: it.Quantity})
			}
			_, _ = svc.ReserveStock(cctx, evt.OrderId, items)
			return nil
		})
		if cons, err := kafka.NewConsumer(brokers, "inventory-service", handler); err == nil {
			defer cons.Close()
			go cons.Run(ctx, "orders.v1.order_created")
		}
	}

	grpcServer := grpc.NewServer()
	invpb.RegisterInventoryServiceServer(grpcServer, grpcsvr.NewPBInventoryServer(svc))

	h := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, h)

	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Errorw("listen failed", "error", err)
		return err
	}
	log.Infow("inventory-service starting", "port", cfg.Port)

	serveErr := make(chan error, 1)
	go func() { serveErr <- grpcServer.Serve(lis) }()

	select {
	case <-ctx.Done():
		log.Infow("shutting down inventory-service...")
		grpcServer.GracefulStop()
		time.Sleep(100 * time.Millisecond)
		return nil
	case err := <-serveErr:
		log.Errorw("serve failed", "error", err)
		return err
	}
}
