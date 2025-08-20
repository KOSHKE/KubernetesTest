package app

import (
	"context"
	"fmt"
	"net"
	"time"

	appsvc "inventory-service/internal/app/services"
	"inventory-service/internal/domain/models"
	grpcsvr "inventory-service/internal/infra/grpc"
	con "inventory-service/internal/infra/kafka/consumer"
	pub "inventory-service/internal/infra/kafka/publisher"
	repo "inventory-service/internal/infra/repository"
	"proto-go/events"
	invpb "proto-go/inventory"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

func Run(ctx context.Context, cfg *Config, logger *zap.Logger) error {
	log := logger.Sugar()

	// Connect DB
	db, err := connectDB(cfg)
	if err != nil {
		log.Errorw("failed to connect to db", "error", err)
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Errorw("failed to obtain db pool", "error", err)
		return err
	}
	defer sqlDB.Close()

	// Repository + migrations
	gormRepo := repo.NewGormInventoryRepository(db)
	if getEnv("AUTO_MIGRATE", "") == "true" {
		if err := gormRepo.AutoMigrate(); err != nil {
			log.Errorw("failed to automigrate", "error", err)
			return err
		}
		// Seed default catalog and stock: 3 of prod-1, 1 of prod-2 (Smart Watch), 2 of prod-3 (Mug)
		seedProducts := []*models.Product{
			{ID: "prod-1", Name: "Wireless Headphones", Description: "High-quality wireless headphones with noise cancellation", PriceMinor: 9999, Currency: "USD", CategoryID: "cat-1", CategoryName: "Electronics", ImageURL: "/images/headphones.jpg", IsActive: true},
			{ID: "prod-2", Name: "Smart Watch", Description: "Fitness tracking smart watch with heart rate monitor", PriceMinor: 19999, Currency: "USD", CategoryID: "cat-1", CategoryName: "Electronics", ImageURL: "/images/smartwatch.jpg", IsActive: true},
			{ID: "prod-3", Name: "Coffee Mug", Description: "Ceramic coffee mug with custom design", PriceMinor: 1599, Currency: "USD", CategoryID: "cat-2", CategoryName: "Home & Kitchen", ImageURL: "/images/mug.jpg", IsActive: true},
		}
		for _, p := range seedProducts {
			_ = db.Clauses(clause.OnConflict{UpdateAll: true}).Create(p).Error
		}
		seedStocks := []*models.Stock{
			{ProductID: "prod-1", AvailableQuantity: 3, ReservedQuantity: 0},
			{ProductID: "prod-2", AvailableQuantity: 1, ReservedQuantity: 0},
			{ProductID: "prod-3", AvailableQuantity: 2, ReservedQuantity: 0},
		}
		for _, s := range seedStocks {
			_ = db.Clauses(clause.OnConflict{UpdateAll: true}).Create(s).Error
		}
	}

	svc := appsvc.NewInventoryService(gormRepo).WithReservationTTL(time.Duration(cfg.ReservationTTLSeconds) * time.Second)

	// Kafka producer for inventory events (best-effort)
	if brokers := getEnv("KAFKA_BROKERS", "kafka:9092"); brokers != "" {
		if p, err := pub.NewStockEventsPublisher(brokers, "inventory.v1.stock_reserved", "inventory.v1.stock_reservation_failed"); err == nil {
			defer p.Close()
			p = p.WithLogger(log)
			svc = svc.WithPublisher(p)
		}
	}

	// Kafka consumer for order created (best-effort)
	if brokers := getEnv("KAFKA_BROKERS", "kafka:9092"); brokers != "" {
		handler := con.OrderCreatedHandlerFunc(func(cctx context.Context, evt *events.OrderCreated) error {
			items := make([]appsvc.StockCheckItem, 0, len(evt.Items))
			for _, it := range evt.Items {
				items = append(items, appsvc.StockCheckItem{ProductID: it.ProductId, Quantity: it.Quantity})
			}
			failed, rerr := svc.ReserveStock(cctx, evt.OrderId, evt.UserId, items)
			if rerr != nil {
				log.Errorw("failed to reserve stock", "orderId", evt.OrderId, "failedProducts", failed, "error", rerr)
			}
			return nil
		})
		if cons, err := con.NewConsumer(brokers, "inventory-service", handler); err == nil {
			defer cons.Close()
			cons.WithLogger(log)
			go cons.Run(ctx, []string{"orders.v1.order_created"})
		}
		// Consume payment outcomes to finalize or release reservations
		if payCons, err := con.NewPaymentConsumer(brokers, "inventory-service", con.PaymentProcessedHandlerFunc(func(cctx context.Context, evt *events.PaymentProcessed) error {
			if ferr := svc.FinalizeReservation(cctx, evt.OrderId, evt.Success); ferr != nil {
				log.Errorw("failed to finalize reservation", "orderId", evt.OrderId, "success", evt.Success, "error", ferr)
			}
			return nil
		})); err == nil {
			defer payCons.Close()
			payCons.WithLogger(log)
			go payCons.Run(ctx, []string{"payments.v1.payment_processed"})
		}
	}

	grpcServer := grpc.NewServer()
	invpb.RegisterInventoryServiceServer(grpcServer, grpcsvr.NewPBInventoryServer(svc, log))

	h := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, h)

	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Errorw("failed to listen", "error", err)
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
		log.Errorw("failed to serve", "error", err)
		return err
	}
}

func connectDB(cfg *Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
}
