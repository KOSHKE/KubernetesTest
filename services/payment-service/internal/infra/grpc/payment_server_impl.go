package grpc

import (
	"ecommerce-platform/services/payment-service/internal/application/services"
	"ecommerce-platform/services/payment-service/internal/domain/ports/grpc"

	grpcLib "google.golang.org/grpc"
)

// PaymentServerImpl implements the domain PaymentServer interface using existing implementation
type PaymentServerImpl struct{}

// NewPaymentServerImpl creates a new payment server implementation
func NewPaymentServerImpl() grpc.PaymentServer {
	return &PaymentServerImpl{}
}

// RegisterPaymentService registers payment service with gRPC server using existing implementation
func (p *PaymentServerImpl) RegisterPaymentService(server *grpcLib.Server, service *services.PaymentApplicationService) {
	RegisterPaymentPBServer(server, service)
}
