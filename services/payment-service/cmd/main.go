package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	app "payment-service/internal/app/services"
	srv "payment-service/internal/grpc"
	mockproc "payment-service/internal/infra/processor"
	pb "proto-go/payment"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type Config struct {
	Port string
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func loadConfig() *Config {
	return &Config{
		Port: getEnv("PORT", "50054"),
	}
}

func main() {
	cfg := loadConfig()

	server := grpc.NewServer()

	// Wire dependencies
	processor := mockproc.NewMockPaymentProcessor()
	paymentService := app.NewPaymentService(processor)

	// Register payment service
	pb.RegisterPaymentServiceServer(server, srv.NewPaymentServer(paymentService))

	// Health service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("Payment Service (mock) starting on port %s", cfg.Port)

	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Shutting down Payment Service...")
	server.GracefulStop()
	log.Println("Payment Service stopped")
}
