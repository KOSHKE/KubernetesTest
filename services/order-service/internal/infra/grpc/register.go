package grpc

import (
	orderpb "ecommerce-platform/proto-go/order"
	"ecommerce-platform/services/order-service/internal/application/services"

	gogrpc "google.golang.org/grpc"
)

// RegisterOrderPBServer hides proto dependency from main
func RegisterOrderPBServer(server *gogrpc.Server, svc *services.OrderApplicationService) {
	pbServer := NewPBOrderServer(svc)
	orderpb.RegisterOrderServiceServer(server, pbServer)
}
