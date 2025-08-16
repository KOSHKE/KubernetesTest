package grpc

import (
	"context"
	"time"

	appsvc "payment-service/internal/app/services"
	"payment-service/internal/domain/entities"
	"payment-service/internal/domain/valueobjects"
	pb "proto-go/payment"
)

type PaymentServer struct {
	pb.UnimplementedPaymentServiceServer
	svc *appsvc.PaymentService
}

func NewPaymentServer(svc *appsvc.PaymentService) *PaymentServer { return &PaymentServer{svc: svc} }

func (s *PaymentServer) ProcessPayment(ctx context.Context, req *pb.ProcessPaymentRequest) (*pb.ProcessPaymentResponse, error) {
	method := entities.MethodCreditCard
	if req.Method == pb.PaymentMethod_CREDIT_CARD {
		method = entities.MethodCreditCard
	}

	amt, _ := valueobjects.NewMoney(req.Amount.Amount, req.Amount.Currency)
	resp, err := s.svc.ProcessPayment(ctx, &appsvc.ProcessPaymentRequest{
		OrderID: req.OrderId,
		UserID:  req.UserId,
		Amount:  amt,
		Method:  method,
		CardNumber: func() string {
			if req.Details != nil {
				return req.Details.CardNumber
			}
			return ""
		}(),
	})
	if err != nil {
		return nil, err
	}

	pbStatus := pb.PaymentStatus_PAYMENT_FAILED
	if resp.Payment.Status == entities.PaymentCompleted {
		pbStatus = pb.PaymentStatus_PAYMENT_COMPLETED
	}

	return &pb.ProcessPaymentResponse{
		Payment: &pb.Payment{
			Id:            resp.Payment.ID,
			OrderId:       resp.Payment.OrderID,
			UserId:        resp.Payment.UserID,
			Amount:        &pb.Money{Amount: resp.Payment.Amount.Amount, Currency: resp.Payment.Amount.Currency},
			Status:        pbStatus,
			Method:        req.Method,
			TransactionId: resp.Payment.TransactionID,
			CreatedAt:     resp.Payment.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     resp.Payment.UpdatedAt.Format(time.RFC3339),
		},
		Success: resp.Success,
		Message: resp.Message,
	}, nil
}

func (s *PaymentServer) GetPayment(ctx context.Context, req *pb.GetPaymentRequest) (*pb.GetPaymentResponse, error) {
	pay, err := s.svc.GetPayment(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.GetPaymentResponse{Payment: &pb.Payment{
		Id:            pay.ID,
		OrderId:       pay.OrderID,
		UserId:        pay.UserID,
		Amount:        &pb.Money{Amount: pay.Amount.Amount, Currency: pay.Amount.Currency},
		Status:        pb.PaymentStatus_PAYMENT_COMPLETED,
		Method:        pb.PaymentMethod_CREDIT_CARD,
		TransactionId: pay.TransactionID,
		CreatedAt:     pay.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     pay.UpdatedAt.Format(time.RFC3339),
	}}, nil
}

func (s *PaymentServer) RefundPayment(ctx context.Context, req *pb.RefundPaymentRequest) (*pb.RefundPaymentResponse, error) {
	amt, _ := valueobjects.NewMoney(req.Amount.Amount, req.Amount.Currency)
	pay, message, success, err := s.svc.RefundPayment(ctx, req.PaymentId, amt, req.Reason)
	if err != nil {
		return nil, err
	}
	return &pb.RefundPaymentResponse{Payment: &pb.Payment{
		Id:            pay.ID,
		OrderId:       pay.OrderID,
		UserId:        pay.UserID,
		Amount:        &pb.Money{Amount: pay.Amount.Amount, Currency: pay.Amount.Currency},
		Status:        pb.PaymentStatus_PAYMENT_REFUNDED,
		Method:        pb.PaymentMethod_CREDIT_CARD,
		TransactionId: pay.TransactionID,
		CreatedAt:     pay.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     pay.UpdatedAt.Format(time.RFC3339),
	}, Success: success, Message: message}, nil
}
