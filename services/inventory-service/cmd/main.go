package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	appsvc "inventory-service/internal/app/services"
	"inventory-service/internal/domain/models"
	grpcsvr "inventory-service/internal/infra/grpc"
	repo "inventory-service/internal/infra/repository"
	invpb "proto-go/inventory"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type Config struct{ Port string }

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	cfg := Config{Port: getEnv("PORT", "50053")}

	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	repository := repo.NewInMemoryInventoryRepository()
	// Seed demo data for development
	products := []*models.Product{
		{ID: "p-1001", Name: "Ultra Laptop 14", Description: "14-inch ultrabook with 16GB RAM, 512GB SSD", PriceMinor: 1299900, Currency: "USD", CategoryID: "electronics", CategoryName: "Electronics", ImageURL: "https://picsum.photos/seed/laptop14/600/400", IsActive: true},
		{ID: "p-1002", Name: "Gaming Laptop 15", Description: "15-inch gaming laptop, RTX 4060, 32GB RAM", PriceMinor: 1999900, Currency: "USD", CategoryID: "electronics", CategoryName: "Electronics", ImageURL: "https://picsum.photos/seed/laptop15/600/400", IsActive: true},
		{ID: "p-2001", Name: "Wireless Headphones", Description: "ANC over-ear wireless headphones", PriceMinor: 19900, Currency: "USD", CategoryID: "audio", CategoryName: "Audio", ImageURL: "https://picsum.photos/seed/headphones/600/400", IsActive: true},
		{ID: "p-3001", Name: "Smartphone Pro", Description: "6.1-inch OLED, 256GB", PriceMinor: 89900, Currency: "USD", CategoryID: "mobile", CategoryName: "Mobile", ImageURL: "https://picsum.photos/seed/phone/600/400", IsActive: true},
		{ID: "p-4001", Name: "Mechanical Keyboard", Description: "Hot‑swap, RGB, tactile switches", PriceMinor: 12900, Currency: "USD", CategoryID: "peripherals", CategoryName: "Peripherals", ImageURL: "https://picsum.photos/seed/keyboard/600/400", IsActive: true},
		{ID: "p-5001", Name: "4K Monitor 27\"", Description: "27-inch 4K IPS display", PriceMinor: 34900, Currency: "USD", CategoryID: "monitors", CategoryName: "Monitors", ImageURL: "https://picsum.photos/seed/monitor27/600/400", IsActive: true},
	}
	stocks := []*models.Stock{
		{ProductID: "p-1001", AvailableQuantity: 12},
		{ProductID: "p-1002", AvailableQuantity: 7},
		{ProductID: "p-2001", AvailableQuantity: 50},
		{ProductID: "p-3001", AvailableQuantity: 20},
		{ProductID: "p-4001", AvailableQuantity: 35},
		{ProductID: "p-5001", AvailableQuantity: 15},
	}
	repository.Seed(products, stocks)

	svc := appsvc.NewInventoryService(repository)

	grpcServer := grpc.NewServer()
	invpb.RegisterInventoryServiceServer(grpcServer, grpcsvr.NewPBInventoryServer(svc))

	h := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, h)

	log.Printf("inventory-service gRPC started on :%s", cfg.Port)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Println("Shutting down inventory-service...")
	grpcServer.GracefulStop()
}
