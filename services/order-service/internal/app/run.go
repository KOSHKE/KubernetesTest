package app

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/kubernetestest/ecommerce-platform/proto-go/events"
	invpb "github.com/kubernetestest/ecommerce-platform/proto-go/inventory"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/app/services"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/models"
	clockimpl "github.com/kubernetestest/ecommerce-platform/services/order-service/internal/infra/clock"
	ordergrpc "github.com/kubernetestest/ecommerce-platform/services/order-service/internal/infra/grpc"
	con "github.com/kubernetestest/ecommerce-platform/services/order-service/internal/infra/kafka/consumer"
	pub "github.com/kubernetestest/ecommerce-platform/services/order-service/internal/infra/kafka/publisher"
	productinfoimpl "github.com/kubernetestest/ecommerce-platform/services/order-service/internal/infra/productinfo"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/infra/repository"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/ports/productinfo"

	pkglogger "github.com/kubernetestest/ecommerce-platform/pkg/logger"
	"go.uber.org/zap"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

)

func Run(ctx context.Context, cfg *Config, logger *zap.Logger) error {
	log := logger.Sugar()

	db, err := connectDB(cfg)
	if err != nil {
		log.Errorw("db connect failed", "error", err)
		return fmt.Errorf("connect db: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Errorw("db pool obtain failed", "error", err)
		return fmt.Errorf("db pool: %w", err)
	}
	defer sqlDB.Close()

	orderRepo := repository.NewGormOrderRepository(db)
	if getEnv("AUTO_MIGRATE", "") == "true" {
		if err := orderRepo.AutoMigrate(); err != nil {
			log.Errorw("automigrate failed", "error", err)
			return fmt.Errorf("automigrate: %w", err)
		}
	}

	// Load brokers once
	brokers := getEnv("KAFKA_BROKERS", "kafka:9092")

	// Optional Kafka producer
	prod, prodErr := createKafkaProducer(brokers, logger)
	if prodErr != nil {
		log.Warnw("kafka producer init failed", "error", prodErr)
	}
	if prod == nil {
		log.Infow("kafka producer disabled or not configured")
	}
	if prod != nil {
		_ = prod.WithLogger(pkglogger.NewZapLogger(log))
		defer prod.Close()
	}

	// Optional Inventory provider
	var (
		provider productinfo.Provider
		invConn  *gogrpc.ClientConn
	)
	if cfg.InventoryServiceURL != "" {
		if conn, err := gogrpc.DialContext(ctx, cfg.InventoryServiceURL, gogrpc.WithInsecure()); err == nil {
			invConn = conn
			invClient := invpb.NewInventoryServiceClient(conn)
			provider = productinfoimpl.NewInventoryProvider(invClient, cfg.InventoryProviderTimeout)
		} else {
			log.Warnw("inventory grpc dial failed", "url", cfg.InventoryServiceURL, "error", err)
		}
	} else {
		log.Infow("inventory provider not configured")
	}

	// Build service with the new constructor
	orderService := services.NewOrderService(
		orderRepo,
		clockimpl.NewSystemClock(),
		prod,
		provider,
		logger,
	)

	// Optional Kafka consumer (payments)
	var wg sync.WaitGroup
	if brokers != "" {
		if cons, err := con.NewConsumer(brokers, "order-service", cfg.KafkaAutoOffsetReset, con.PaymentProcessedHandlerFunc(func(cctx context.Context, evt *events.PaymentProcessed) error {
			status := models.OrderStatusConfirmed
			if !evt.Success {
				status = models.OrderStatusCancelled
			}
			if _, err := orderService.UpdateOrderStatus(cctx, &services.UpdateOrderStatusRequest{OrderID: evt.OrderId, Status: status}); err != nil {
				log.Warnw("update order status failed", "orderID", evt.OrderId, "status", status, "error", err)
			}
			return nil
		})); err == nil {
			defer cons.Close()
			cons.WithLogger(pkglogger.NewZapLogger(log))
			wg.Add(1)
			go func() { defer wg.Done(); cons.Run(ctx, []string{"payments.v1.payment_processed"}) }()
		} else {
			log.Warnw("kafka consumer init failed", "error", err)
		}
	}

	server := gogrpc.NewServer()
	ordergrpc.RegisterOrderPBServer(server, orderService, cfg.DefaultCurrency)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Errorw("listen failed", "error", err)
		return fmt.Errorf("listen: %w", err)
	}
	log.Infow("order-service starting", "port", cfg.Port)

	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(lis) }()

	shutdown := func() {
		server.GracefulStop()
		wg.Wait()
		if invConn != nil {
			_ = invConn.Close()
		}
	}

	select {
	case <-ctx.Done():
		log.Infow("shutting down order-service...")
		shutdown()
		return nil
	case err := <-serveErr:
		log.Errorw("serve failed", "error", err)
		shutdown()
		return err
	}
}

func connectDB(cfg *Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: nil})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	return db, nil
}

func createKafkaProducer(brokers string, logger *zap.Logger) (*pub.OrderCreatedPublisher, error) {
	if brokers == "" {
		return nil, nil
	}
	prod, err := pub.NewOrderCreatedPublisher(brokers, "orders.v1.order_created")
	if err != nil {
		return nil, err
	}
	return prod, nil
}
