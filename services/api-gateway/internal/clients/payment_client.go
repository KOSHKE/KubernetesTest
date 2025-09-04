package clients

import (
	"context"

	"ecommerce-platform/pkg/common/valueobjects"
	paymentpb "ecommerce-platform/proto-go/payment"
	"ecommerce-platform/services/api-gateway/pkg/grpc"
)

// ---------------- Payment Client Interface ----------------

type PaymentClient interface {
	Close() error
	ProcessPayment(ctx context.Context, req *ProcessPaymentRequest) (*ProcessPaymentResponse, error)
}

// ---------------- Payment Models ----------------

// Payment represents a payment in the system
type Payment struct {
	ID            string             `json:"id"`
	OrderID       string             `json:"order_id"`
	UserID        string             `json:"user_id"`
	Amount        valueobjects.Money `json:"amount"`
	Status        string             `json:"status"`
	Method        string             `json:"method"`
	TransactionID string             `json:"transaction_id"`
	CreatedAt     string             `json:"created_at"`
	UpdatedAt     string             `json:"updated_at"`
}

// ProcessPaymentRequest represents a request to process a payment
type ProcessPaymentRequest struct {
	OrderID string             `json:"order_id"`
	UserID  string             `json:"user_id"`
	Amount  valueobjects.Money `json:"amount"`
	Method  string             `json:"method"`
	Details PaymentDetails     `json:"details"`
}

// PaymentDetails represents payment method specific details
type PaymentDetails struct {
	CardNumber  string `json:"card_number"`
	CardHolder  string `json:"card_holder"`
	ExpiryMonth string `json:"expiry_month"`
	ExpiryYear  string `json:"expiry_year"`
	CVV         string `json:"cvv"`
}

// ProcessPaymentResponse represents the response after processing a payment
type ProcessPaymentResponse struct {
	Payment *Payment `json:"payment"`
	Success bool     `json:"success"`
	Message string   `json:"message"`
}

// paymentClient implements PaymentClient interface
type paymentClient struct {
	*grpc.BaseClient
	client paymentpb.PaymentServiceClient
}

// ---------------- Constructor ----------------

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

// ---------------- Payment Methods ----------------

// ProcessPayment processes a payment
func (c *paymentClient) ProcessPayment(ctx context.Context, req *ProcessPaymentRequest) (*ProcessPaymentResponse, error) {
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

	return &ProcessPaymentResponse{
		Payment: mapPaymentFromPB(resp.GetPayment()),
		Success: resp.GetSuccess(),
		Message: resp.GetMessage(),
	}, nil
}

// ---------------- Mapping Helpers ----------------

func mapMethodToEnum(method string) paymentpb.PaymentMethod {
	switch method {
	case "CREDIT_CARD", "credit_card":
		return paymentpb.PaymentMethod_CREDIT_CARD
	default:
		return paymentpb.PaymentMethod_CREDIT_CARD
	}
}

func mapMethodFromEnum(method paymentpb.PaymentMethod) string {
	switch method {
	case paymentpb.PaymentMethod_CREDIT_CARD:
		return "CREDIT_CARD"
	default:
		return "CREDIT_CARD"
	}
}

func mapStatusFromEnum(status paymentpb.PaymentStatus) string {
	switch status {
	case paymentpb.PaymentStatus_PAYMENT_PENDING:
		return "PENDING"
	case paymentpb.PaymentStatus_PAYMENT_COMPLETED:
		return "COMPLETED"
	case paymentpb.PaymentStatus_PAYMENT_FAILED:
		return "FAILED"
	default:
		return "PENDING"
	}
}

func mapPaymentFromPB(p *paymentpb.Payment) *Payment {
	if p == nil {
		return nil
	}
	amt := valueobjects.Money{}
	if p.Amount != nil {
		amt = valueobjects.Money{Amount: p.Amount.Amount, Currency: p.Amount.Currency}
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
