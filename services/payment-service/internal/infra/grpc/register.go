package grpc

import (
	pb "ecommerce-platform/proto-go/payment"
	"ecommerce-platform/services/payment-service/internal/application/services"

	"google.golang.org/grpc"
)

// RegisterPaymentPBServer registers the protobuf server implementation
func RegisterPaymentPBServer(
	server *grpc.Server,
	paymentAppService *services.PaymentApplicationService,
) {
	pb.RegisterPaymentServiceServer(server, NewPBPaymentServer(paymentAppService))
}
