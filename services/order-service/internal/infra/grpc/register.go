package grpc

import (
	appsvc "github.com/kubernetestest/ecommerce-platform/services/order-service/internal/app/services"
	orderpb "github.com/kubernetestest/ecommerce-platform/services/order-service/internal/pb/order"

	gogrpc "google.golang.org/grpc"
)

// RegisterOrderPBServer registers the protobuf server implementation
func RegisterOrderPBServer(server *gogrpc.Server, svc *appsvc.OrderService, defaultCurrency string) {
	orderpb.RegisterOrderServiceServer(server, NewPBOrderServer(svc, defaultCurrency))
}
