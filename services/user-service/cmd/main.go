package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"user-service/internal/app/services"
	grpcsvc "user-service/internal/infra/grpc"
	"user-service/internal/infra/repository"

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
		Port:   getEnv("PORT", "50051"),
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

	// Connect to database with GORM
	db, err := connectDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize repository and service
	userRepo := repository.NewGormUserRepository(db)
	if err := userRepo.AutoMigrate(); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	userService := services.NewUserService(userRepo, nil)

	// Create gRPC server
	server := gogrpc.NewServer()
	grpcsvc.RegisterUserPBServer(server, userService)

	// Health service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Start gRPC server
	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Printf("User Service starting on port %s", cfg.Port)

	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Shutting down User Service...")
	server.GracefulStop()
	log.Println("User Service stopped")
}

func connectDB(cfg *Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	// Test connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
