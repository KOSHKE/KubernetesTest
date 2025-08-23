package grpc

import (
	appsvc "github.com/kubernetestest/ecommerce-platform/services/payment-service/internal/app/services"
	pb "github.com/kubernetestest/ecommerce-platform/proto-go/payment"

	"google.golang.org/grpc"
)

// RegisterPaymentPBServer registers the protobuf server implementation
func RegisterPaymentPBServer(server *grpc.Server, svc *appsvc.PaymentService) {
	pb.RegisterPaymentServiceServer(server, NewPBPaymentServer(svc))
}
