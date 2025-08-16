package grpc

import (
	appsvc "order-service/internal/app/services"
	orderpb "order-service/internal/pb/order"

	gogrpc "google.golang.org/grpc"
)

// RegisterOrderPBServer registers the protobuf server implementation
func RegisterOrderPBServer(server *gogrpc.Server, svc *appsvc.OrderService) {
	orderpb.RegisterOrderServiceServer(server, NewPBOrderServer(svc))
}
