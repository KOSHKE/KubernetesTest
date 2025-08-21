package http

import (
	"api-gateway/internal/clients"
	"api-gateway/pkg/types"
)

// ========== User Requests ==========

// RegisterRequest contains information for registering a user
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email" msg:"Valid email address is required"`
	Password  string `json:"password" binding:"required,min=6" msg:"Password must be at least 6 characters long"`
	FirstName string `json:"first_name" binding:"required,min=2,max=50" msg:"First name must be between 2 and 50 characters"`
	LastName  string `json:"last_name" binding:"required,min=2,max=50" msg:"Last name must be between 2 and 50 characters"`
	Phone     string `json:"phone" binding:"omitempty,len=10" msg:"Phone number must be exactly 10 digits"`
}

// ToClientRequest converts RegisterRequest to clients.RegisterRequest
func (r *RegisterRequest) ToClientRequest() *clients.RegisterRequest {
	return &clients.RegisterRequest{
		Email:     r.Email,
		Password:  r.Password,
		FirstName: r.FirstName,
		LastName:  r.LastName,
		Phone:     r.Phone,
	}
}

// LoginRequest contains information for logging in a user
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" msg:"Valid email address is required"`
	Password string `json:"password" binding:"required" msg:"Password is required"`
}

// ToClientRequest converts LoginRequest to clients.LoginRequest
func (r *LoginRequest) ToClientRequest() *clients.LoginRequest {
	return &clients.LoginRequest{
		Email:    r.Email,
		Password: r.Password,
	}
}

// ========== Order Requests ==========

// CreateOrderRequest contains information for creating an order
type CreateOrderRequest struct {
	UserID          string             `json:"user_id" binding:"required" msg:"User ID is required"`
	Items           []OrderItemRequest `json:"items" binding:"required,min=1,dive" msg:"At least one item is required"`
	ShippingAddress string             `json:"shipping_address" binding:"required,min=10,max=200" msg:"Shipping address must be between 10 and 200 characters"`
	PaymentDetails  PaymentDetails     `json:"payment_details" binding:"required" msg:"Payment details are required"`
	PaymentMethod   string             `json:"payment_method" binding:"required,oneof=credit_card debit_card paypal" msg:"Payment method must be credit_card, debit_card, or paypal"`
}

// ToClientRequest converts CreateOrderRequest to clients.CreateOrderRequest
func (r *CreateOrderRequest) ToClientRequest() *clients.CreateOrderRequest {
	items := make([]clients.OrderItemRequest, len(r.Items))
	for i, item := range r.Items {
		items[i] = *item.ToClientRequest()
	}

	return &clients.CreateOrderRequest{
		UserID:          r.UserID,
		Items:           items,
		ShippingAddress: r.ShippingAddress,
	}
}

// ToStockCheckRequest converts CreateOrderRequest to clients.StockCheckRequest
func (r *CreateOrderRequest) ToStockCheckRequest() *clients.StockCheckRequest {
	items := make([]clients.StockCheckItem, len(r.Items))
	for i, item := range r.Items {
		items[i] = clients.StockCheckItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	return &clients.StockCheckRequest{Items: items}
}

// OrderItemRequest contains information for an order item
type OrderItemRequest struct {
	ProductID string `json:"product_id" binding:"required" msg:"Product ID is required"`
	Quantity  int32  `json:"quantity" binding:"required,min=1,max=100" msg:"Quantity must be between 1 and 100"`
}

// ToClientRequest converts OrderItemRequest to clients.OrderItemRequest
func (r *OrderItemRequest) ToClientRequest() *clients.OrderItemRequest {
	return &clients.OrderItemRequest{
		ProductID: r.ProductID,
		Quantity:  r.Quantity,
	}
}

// PaymentDetails contains payment information
type PaymentDetails struct {
	CardNumber  string `json:"card_number" binding:"required,min=12,max=19" msg:"Card number must be between 12 and 19 digits"`
	CardHolder  string `json:"card_holder" binding:"required,min=2,max=100" msg:"Card holder name must be between 2 and 100 characters"`
	ExpiryMonth string `json:"expiry_month" binding:"required,len=2" msg:"Expiry month must be 2 digits (MM)"`
	ExpiryYear  string `json:"expiry_year" binding:"required,len=4" msg:"Expiry year must be 4 digits (YYYY)"`
	CVV         string `json:"cvv" binding:"required,len=3" msg:"CVV must be 3 digits"`
}

// ========== Payment Requests ==========

// ProcessPaymentRequest contains information for processing payment
type ProcessPaymentRequest struct {
	OrderID string         `json:"order_id" binding:"required,uuid" msg:"Valid order ID is required"`
	UserID  string         `json:"user_id" binding:"required,uuid" msg:"Valid user ID is required"`
	Amount  types.Money    `json:"amount" binding:"required" msg:"Payment amount is required"`
	Details PaymentDetails `json:"details" binding:"required" msg:"Payment details are required"`
}

// ToClientRequest converts ProcessPaymentRequest to clients.ProcessPaymentRequest
func (r *ProcessPaymentRequest) ToClientRequest() *clients.ProcessPaymentRequest {
	return &clients.ProcessPaymentRequest{
		OrderID: r.OrderID,
		UserID:  r.UserID,
		Amount:  r.Amount,
		Details: clients.PaymentDetails{
			CardNumber:  r.Details.CardNumber,
			CardHolder:  r.Details.CardHolder,
			ExpiryMonth: r.Details.ExpiryMonth,
			ExpiryYear:  r.Details.ExpiryYear,
			CVV:         r.Details.CVV,
		},
	}
}

// RefundRequest contains information for refunding payment
type RefundRequest struct {
	Amount types.Money `json:"amount" binding:"required" msg:"Refund amount is required"`
	Reason string      `json:"reason" binding:"required,min=10,max=500" msg:"Refund reason must be between 10 and 500 characters"`
}

// ========== Inventory Requests ==========

// StockCheckRequest contains information for checking stock availability
type StockCheckRequest struct {
	Items []StockCheckItem `json:"items" binding:"required,min=1,dive" msg:"At least one item is required for stock check"`
}

// StockCheckItem contains information for checking stock of a specific item
type StockCheckItem struct {
	ProductID string `json:"product_id" binding:"required" msg:"Product ID is required"`
	Quantity  int32  `json:"quantity" binding:"required,min=1,max=1000" msg:"Quantity must be between 1 and 1000"`
}
