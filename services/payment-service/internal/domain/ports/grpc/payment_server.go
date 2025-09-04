package grpc

import (
	"ecommerce-platform/services/payment-service/internal/application/services"

	"google.golang.org/grpc"
)

// PaymentServer defines the interface for payment gRPC server
type PaymentServer interface {
	// RegisterPaymentService registers payment service with gRPC server
	RegisterPaymentService(server *grpc.Server, service *services.PaymentApplicationService)
}
