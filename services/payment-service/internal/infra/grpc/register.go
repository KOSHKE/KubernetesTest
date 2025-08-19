package grpc

import (
	appsvc "payment-service/internal/app/services"
	pb "proto-go/payment"

	"google.golang.org/grpc"
)

// RegisterPaymentPBServer registers the protobuf server implementation
func RegisterPaymentPBServer(server *grpc.Server, svc *appsvc.PaymentService) {
	pb.RegisterPaymentServiceServer(server, NewPBPaymentServer(svc))
}
