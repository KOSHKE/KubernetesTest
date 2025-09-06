package grpc

import (
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/proto-go/inventory"
	"ecommerce-platform/services/inventory-service/internal/application/services"

	gogrpc "google.golang.org/grpc"
)

// RegisterInventoryServer hides proto dependency from main
func RegisterInventoryServer(server *gogrpc.Server, svc *services.InventoryApplicationService, logger logger.Logger) {
	inventoryServer := NewPBInventoryServer(svc, logger)
	inventory.RegisterInventoryServiceServer(server, inventoryServer)
}
