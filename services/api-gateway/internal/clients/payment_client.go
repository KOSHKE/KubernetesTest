package clients

import (
	"context"
	"time"

	"api-gateway/internal/types"
	paymentpb "proto-go/payment"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PaymentClient interface {
	Close() error
	ProcessPayment(ctx context.Context, req *ProcessPaymentRequest) (*PaymentResponse, error)
	GetPayment(ctx context.Context, paymentID string) (*Payment, error)
	RefundPayment(ctx context.Context, paymentID string, amount types.Money, reason string) (*PaymentResponse, error)
}

type paymentClient struct {
	conn   *grpc.ClientConn
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}

	return &paymentClient{conn: conn, client: paymentpb.NewPaymentServiceClient(conn)}, nil
}

func (c *paymentClient) Close() error { return c.conn.Close() }

func (c *paymentClient) ProcessPayment(ctx context.Context, req *ProcessPaymentRequest) (*PaymentResponse, error) {
	method := mapMethodToEnum(req.Method)
	grpcReq := &paymentpb.ProcessPaymentRequest{OrderId: req.OrderID, UserId: req.UserID, Amount: &paymentpb.Money{Amount: req.Amount.Amount, Currency: req.Amount.Currency}, Method: method, Details: &paymentpb.PaymentDetails{CardNumber: req.Details.CardNumber, CardHolder: req.Details.CardHolder, ExpiryMonth: req.Details.ExpiryMonth, ExpiryYear: req.Details.ExpiryYear, Cvv: req.Details.CVV}}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res, err := c.client.ProcessPayment(ctx, grpcReq)
	if err != nil {
		return nil, err
	}
	return &PaymentResponse{Payment: mapPaymentFromPB(res.Payment), Success: res.Success, Message: res.Message}, nil
}

func (c *paymentClient) GetPayment(ctx context.Context, paymentID string) (*Payment, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res, err := c.client.GetPayment(ctx, &paymentpb.GetPaymentRequest{Id: paymentID})
	if err != nil {
		return nil, err
	}
	return mapPaymentFromPB(res.Payment), nil
}

func (c *paymentClient) RefundPayment(ctx context.Context, paymentID string, amount types.Money, reason string) (*PaymentResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res, err := c.client.RefundPayment(ctx, &paymentpb.RefundPaymentRequest{PaymentId: paymentID, Amount: &paymentpb.Money{Amount: amount.Amount, Currency: amount.Currency}, Reason: reason})
	if err != nil {
		return nil, err
	}
	return &PaymentResponse{Payment: mapPaymentFromPB(res.Payment), Success: res.Success, Message: res.Message}, nil
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

func mapPaymentFromPB(p *paymentpb.Payment) *Payment {
	if p == nil {
		return nil
	}
	return &Payment{ID: p.Id, OrderID: p.OrderId, UserID: p.UserId, Amount: types.Money{Amount: p.Amount.Amount, Currency: p.Amount.Currency}, Status: mapStatusFromEnum(p.Status), Method: mapMethodFromEnum(p.Method), TransactionID: p.TransactionId, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt}
}
