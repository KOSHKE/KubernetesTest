package grpc

import (
	orderpb "ecommerce-platform/proto-go/order"
	"ecommerce-platform/services/order-service/internal/application/services"
	"ecommerce-platform/services/order-service/internal/metrics"

	gogrpc "google.golang.org/grpc"
)

// RegisterOrderPBServer hides proto dependency from main
func RegisterOrderPBServer(server *gogrpc.Server, svc *services.OrderApplicationService, m metrics.OrderMetrics) {
	pbServer := NewPBOrderServer(svc, m)
	orderpb.RegisterOrderServiceServer(server, pbServer)
}
