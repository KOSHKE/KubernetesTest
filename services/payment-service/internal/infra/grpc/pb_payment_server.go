package grpc

import (
	"context"

	appsvc "payment-service/internal/app/services"
	"payment-service/internal/domain/entities"
	"payment-service/internal/domain/valueobjects"
	pb "proto-go/payment"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type PBPaymentServer struct {
	pb.UnimplementedPaymentServiceServer
	svc *appsvc.PaymentService
}

func NewPBPaymentServer(svc *appsvc.PaymentService) *PBPaymentServer {
	return &PBPaymentServer{svc: svc}
}

// helpers
func toPBStatus(s entities.PaymentStatus) pb.PaymentStatus {
	switch s {
	case entities.PaymentCompleted:
		return pb.PaymentStatus_PAYMENT_COMPLETED
	case entities.PaymentRefunded:
		return pb.PaymentStatus_PAYMENT_REFUNDED
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
	amt, err := valueobjects.NewMoney(req.Amount.Amount, req.Amount.Currency)
	if err != nil {
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
		return nil, err
	}
	return &pb.ProcessPaymentResponse{
		Payment: toPBPayment(resp.Payment),
		Success: resp.Success,
		Message: resp.Message,
	}, nil
}

func (s *PBPaymentServer) GetPayment(ctx context.Context, req *pb.GetPaymentRequest) (*pb.GetPaymentResponse, error) {
	pay, err := s.svc.GetPayment(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.GetPaymentResponse{Payment: toPBPayment(pay)}, nil
}

func (s *PBPaymentServer) RefundPayment(ctx context.Context, req *pb.RefundPaymentRequest) (*pb.RefundPaymentResponse, error) {
	amt, err := valueobjects.NewMoney(req.Amount.Amount, req.Amount.Currency)
	if err != nil {
		return nil, err
	}
	pay, message, success, err := s.svc.RefundPayment(ctx, req.PaymentId, amt, req.Reason)
	if err != nil {
		return nil, err
	}
	return &pb.RefundPaymentResponse{Payment: toPBPayment(pay), Success: success, Message: message}, nil
}
