package grpc

import (
	"ecommerce-platform/proto-go/inventory"
	"ecommerce-platform/services/inventory-service/internal/application/services"
	"ecommerce-platform/services/inventory-service/internal/metrics"

	gogrpc "google.golang.org/grpc"
)

// RegisterInventoryServer hides proto dependency from main
func RegisterInventoryServer(server *gogrpc.Server, svc *services.InventoryApplicationService, metrics metrics.InventoryMetrics) {
	inventoryServer := NewPBInventoryServer(svc, metrics)
	inventory.RegisterInventoryServiceServer(server, inventoryServer)
}
