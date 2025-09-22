package grpc

import (
	pb "ecommerce-platform/proto-go/payment"
	"ecommerce-platform/services/payment-service/internal/application/services"
	"ecommerce-platform/services/payment-service/internal/metrics"

	"google.golang.org/grpc"
)

// RegisterPaymentServer hides proto dependency from main
func RegisterPaymentServer(server *grpc.Server, svc *services.PaymentApplicationService, metrics metrics.PaymentMetrics) {
	paymentServer := NewPBPaymentServer(svc, metrics)
	pb.RegisterPaymentServiceServer(server, paymentServer)
}
