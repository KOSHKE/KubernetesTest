package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"order-service/internal/app/services"
	ordergrpc "order-service/internal/infra/grpc"
	"order-service/internal/infra/repository"

	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	Port   string
	DBHost string
	DBPort string
	DBName string
	DBUser string
	DBPass string
}

func loadConfig() *Config {
	return &Config{
		Port:   getEnv("PORT", "50052"),
		DBHost: getEnv("DB_HOST", "localhost"),
		DBPort: getEnv("DB_PORT", "5432"),
		DBName: getEnv("DB_NAME", "orderdb"),
		DBUser: getEnv("DB_USER", "admin"),
		DBPass: getEnv("DB_PASSWORD", "password"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	cfg := loadConfig()

	db, err := connectDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	orderRepo := repository.NewGormOrderRepository(db)
	if err := orderRepo.AutoMigrate(); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	orderService := services.NewOrderService(orderRepo)

	server := gogrpc.NewServer()
	ordergrpc.RegisterOrderPBServer(server, orderService)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("Order Service starting on port %s", cfg.Port)
	log.Printf("Database: %s:%s/%s", cfg.DBHost, cfg.DBPort, cfg.DBName)

	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Shutting down Order Service...")
	server.GracefulStop()
	log.Println("Order Service stopped")
}

func connectDB(cfg *Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
