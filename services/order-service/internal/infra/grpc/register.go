package grpc

import (
	appsvc "github.com/kubernetestest/ecommerce-platform/services/order-service/internal/application/services"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/metrics"
	orderpb "github.com/kubernetestest/ecommerce-platform/proto-go/order"

	gogrpc "google.golang.org/grpc"
)

// RegisterOrderPBServer registers the protobuf server implementation
func RegisterOrderPBServer(server *gogrpc.Server, svc *appsvc.OrderService, defaultCurrency string, m metrics.OrderMetrics) {
	orderpb.RegisterOrderServiceServer(server, NewPBOrderServer(svc, defaultCurrency, m))
}
