package grpc

import (
	"context"
	"time"

	pb "github.com/kubernetestest/ecommerce-platform/proto-go/payment"
	appsvc "github.com/kubernetestest/ecommerce-platform/services/payment-service/internal/app/services"
	"github.com/kubernetestest/ecommerce-platform/services/payment-service/internal/domain/entities"
	"github.com/kubernetestest/ecommerce-platform/services/payment-service/internal/domain/valueobjects"
	"github.com/kubernetestest/ecommerce-platform/services/payment-service/internal/metrics"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type PBPaymentServer struct {
	pb.UnimplementedPaymentServiceServer
	svc     *appsvc.PaymentService
	metrics metrics.PaymentMetrics
}

func NewPBPaymentServer(svc *appsvc.PaymentService, metrics metrics.PaymentMetrics) *PBPaymentServer {
	return &PBPaymentServer{
		svc:     svc,
		metrics: metrics,
	}
}

// helpers
func toPBStatus(s entities.PaymentStatus) pb.PaymentStatus {
	switch s {
	case entities.PaymentCompleted:
		return pb.PaymentStatus_PAYMENT_COMPLETED
	case entities.PaymentFailed:
		fallthrough
	default:
		return pb.PaymentStatus_PAYMENT_FAILED
	}
}

func toPBMethod(m entities.PaymentMethod) pb.PaymentMethod {
	switch m {
	case entities.MethodCreditCard:
		return pb.PaymentMethod_CREDIT_CARD
	default:
		return pb.PaymentMethod_CREDIT_CARD
	}
}

func fromPBMethod(m pb.PaymentMethod) entities.PaymentMethod {
	switch m {
	case pb.PaymentMethod_CREDIT_CARD:
		return entities.MethodCreditCard
	default:
		return entities.MethodCreditCard
	}
}

func toPBPayment(p *entities.Payment) *pb.Payment {
	if p == nil {
		return nil
	}
	return &pb.Payment{
		Id:            p.ID,
		OrderId:       p.OrderID,
		UserId:        p.UserID,
		Amount:        &pb.Money{Amount: p.Amount.Amount, Currency: p.Amount.Currency},
		Status:        toPBStatus(p.Status),
		Method:        toPBMethod(p.Method),
		TransactionId: p.TransactionID,
		CreatedAt:     timestamppb.New(p.CreatedAt.UTC()),
		UpdatedAt:     timestamppb.New(p.UpdatedAt.UTC()),
	}
}

func (s *PBPaymentServer) ProcessPayment(ctx context.Context, req *pb.ProcessPaymentRequest) (*pb.ProcessPaymentResponse, error) {
	start := time.Now()

	amt, err := valueobjects.NewMoney(req.Amount.Amount, req.Amount.Currency)
	if err != nil {
		s.metrics.HTTPRequestsTotal("POST", "/ProcessPayment", "400")
		s.metrics.HTTPRequestDuration("POST", "/ProcessPayment", time.Since(start))
		return nil, err
	}
	method := fromPBMethod(req.Method)
	card := ""
	if req.Details != nil {
		card = req.Details.CardNumber
	}
	resp, err := s.svc.ProcessPayment(ctx, &appsvc.ProcessPaymentRequest{
		OrderID:    req.OrderId,
		UserID:     req.UserId,
		Amount:     amt,
		Method:     method,
		CardNumber: card,
	})
	if err != nil {
		s.metrics.HTTPRequestsTotal("POST", "/ProcessPayment", "500")
		s.metrics.HTTPRequestDuration("POST", "/ProcessPayment", time.Since(start))
		return nil, err
	}

	s.metrics.HTTPRequestsTotal("POST", "/ProcessPayment", "200")
	s.metrics.HTTPRequestDuration("POST", "/ProcessPayment", time.Since(start))

	return &pb.ProcessPaymentResponse{
		Payment: toPBPayment(resp.Payment),
		Success: resp.Success,
		Message: resp.Message,
	}, nil
}

func (s *PBPaymentServer) GetPayment(ctx context.Context, req *pb.GetPaymentRequest) (*pb.GetPaymentResponse, error) {
	start := time.Now()

	pay, err := s.svc.GetPayment(ctx, req.Id)
	if err != nil {
		s.metrics.HTTPRequestsTotal("GET", "/GetPayment", "404")
		s.metrics.HTTPRequestDuration("GET", "/GetPayment", time.Since(start))
		return nil, err
	}

	s.metrics.HTTPRequestsTotal("GET", "/GetPayment", "200")
	s.metrics.HTTPRequestDuration("GET", "/GetPayment", time.Since(start))

	return &pb.GetPaymentResponse{Payment: toPBPayment(pay)}, nil
}
