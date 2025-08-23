package clients

import (
	"context"

	"github.com/kubernetestest/ecommerce-platform/services/api-gateway/pkg/grpc"
	"github.com/kubernetestest/ecommerce-platform/services/api-gateway/pkg/types"
	paymentpb "github.com/kubernetestest/ecommerce-platform/proto-go/payment"
)

type PaymentClient interface {
	Close() error
	ProcessPayment(ctx context.Context, req *ProcessPaymentRequest) (*PaymentResponse, error)
	GetPayment(ctx context.Context, paymentID string) (*Payment, error)
	RefundPayment(ctx context.Context, paymentID string, amount types.Money, reason string) (*PaymentResponse, error)
}

type paymentClient struct {
	*grpc.BaseClient
	client paymentpb.PaymentServiceClient
}

type Payment struct {
	ID            string      `json:"id"`
	OrderID       string      `json:"order_id"`
	UserID        string      `json:"user_id"`
	Amount        types.Money `json:"amount"`
	Status        string      `json:"status"`
	Method        string      `json:"method"`
	TransactionID string      `json:"transaction_id"`
	CreatedAt     string      `json:"created_at"`
	UpdatedAt     string      `json:"updated_at"`
}

type ProcessPaymentRequest struct {
	OrderID string         `json:"order_id"`
	UserID  string         `json:"user_id"`
	Amount  types.Money    `json:"amount"`
	Method  string         `json:"method"`
	Details PaymentDetails `json:"details"`
}

type PaymentDetails struct {
	CardNumber  string `json:"card_number"`
	CardHolder  string `json:"card_holder"`
	ExpiryMonth string `json:"expiry_month"`
	ExpiryYear  string `json:"expiry_year"`
	CVV         string `json:"cvv"`
}

type PaymentResponse struct {
	Payment *Payment `json:"payment"`
	Success bool     `json:"success"`
	Message string   `json:"message"`
}

func NewPaymentClient(address string) (PaymentClient, error) {
	baseClient, err := grpc.NewBaseClient(address)
	if err != nil {
		return nil, err
	}
	return &paymentClient{
		BaseClient: baseClient,
		client:     paymentpb.NewPaymentServiceClient(baseClient.GetConn()),
	}, nil
}

func (c *paymentClient) ProcessPayment(ctx context.Context, req *ProcessPaymentRequest) (*PaymentResponse, error) {
	grpcReq := &paymentpb.ProcessPaymentRequest{
		OrderId: req.OrderID,
		UserId:  req.UserID,
		Amount:  &paymentpb.Money{Amount: req.Amount.Amount, Currency: req.Amount.Currency},
		Method:  mapMethodToEnum(req.Method),
		Details: &paymentpb.PaymentDetails{
			CardNumber:  req.Details.CardNumber,
			CardHolder:  req.Details.CardHolder,
			ExpiryMonth: req.Details.ExpiryMonth,
			ExpiryYear:  req.Details.ExpiryYear,
			Cvv:         req.Details.CVV,
		},
	}

	resp, err := grpc.WithTimeoutResult(ctx, func(ctx context.Context) (*paymentpb.ProcessPaymentResponse, error) {
		return c.client.ProcessPayment(ctx, grpcReq)
	})
	if err != nil {
		return nil, err
	}
	return &PaymentResponse{
		Payment: mapPaymentFromPB(resp.Payment),
		Success: resp.Success,
		Message: resp.Message,
	}, nil
}

func (c *paymentClient) GetPayment(ctx context.Context, paymentID string) (*Payment, error) {
	resp, err := grpc.WithTimeoutResult(ctx, func(ctx context.Context) (*paymentpb.GetPaymentResponse, error) {
		return c.client.GetPayment(ctx, &paymentpb.GetPaymentRequest{Id: paymentID})
	})
	if err != nil {
		return nil, err
	}
	return mapPaymentFromPB(resp.Payment), nil
}

func (c *paymentClient) RefundPayment(ctx context.Context, paymentID string, amount types.Money, reason string) (*PaymentResponse, error) {
	resp, err := grpc.WithTimeoutResult(ctx, func(ctx context.Context) (*paymentpb.RefundPaymentResponse, error) {
		return c.client.RefundPayment(ctx, &paymentpb.RefundPaymentRequest{
			PaymentId: paymentID,
			Amount:    &paymentpb.Money{Amount: amount.Amount, Currency: amount.Currency},
			Reason:    reason,
		})
	})
	if err != nil {
		return nil, err
	}
	return &PaymentResponse{
		Payment: mapPaymentFromPB(resp.Payment),
		Success: resp.Success,
		Message: resp.Message,
	}, nil
}

func mapMethodToEnum(method string) paymentpb.PaymentMethod {
	switch method {
	case "CREDIT_CARD", "credit_card":
		return paymentpb.PaymentMethod_CREDIT_CARD
	case "DEBIT_CARD", "debit_card":
		return paymentpb.PaymentMethod_DEBIT_CARD
	case "PAYPAL", "paypal":
		return paymentpb.PaymentMethod_PAYPAL
	case "BANK_TRANSFER", "bank_transfer":
		return paymentpb.PaymentMethod_BANK_TRANSFER
	default:
		return paymentpb.PaymentMethod_CREDIT_CARD
	}
}

func mapMethodFromEnum(method paymentpb.PaymentMethod) string {
	switch method {
	case paymentpb.PaymentMethod_CREDIT_CARD:
		return "CREDIT_CARD"
	case paymentpb.PaymentMethod_DEBIT_CARD:
		return "DEBIT_CARD"
	case paymentpb.PaymentMethod_PAYPAL:
		return "PAYPAL"
	case paymentpb.PaymentMethod_BANK_TRANSFER:
		return "BANK_TRANSFER"
	default:
		return "CREDIT_CARD"
	}
}

func mapStatusFromEnum(status paymentpb.PaymentStatus) string {
	switch status {
	case paymentpb.PaymentStatus_PAYMENT_COMPLETED:
		return "COMPLETED"
	case paymentpb.PaymentStatus_PAYMENT_FAILED:
		return "FAILED"
	case paymentpb.PaymentStatus_PAYMENT_REFUNDED:
		return "REFUNDED"
	case paymentpb.PaymentStatus_PAYMENT_PROCESSING:
		return "PROCESSING"
	default:
		return "PENDING"
	}
}

func mapPaymentFromPB(p *paymentpb.Payment) *Payment {
	if p == nil {
		return nil
	}
	amt := types.Money{}
	if p.Amount != nil {
		amt = types.Money{Amount: p.Amount.Amount, Currency: p.Amount.Currency}
	}
	return &Payment{
		ID:            p.Id,
		OrderID:       p.OrderId,
		UserID:        p.UserId,
		Amount:        amt,
		Status:        mapStatusFromEnum(p.Status),
		Method:        mapMethodFromEnum(p.Method),
		TransactionID: p.TransactionId,
		CreatedAt:     grpc.FormatTimestamp(p.CreatedAt),
		UpdatedAt:     grpc.FormatTimestamp(p.UpdatedAt),
	}
}
