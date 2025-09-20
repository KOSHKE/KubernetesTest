package dto

import (
	"time"

	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/proto-go/common"
	"ecommerce-platform/services/order-service/internal/domain/aggregates"
	"ecommerce-platform/services/order-service/internal/domain/entities"
)

// OrderResponse represents the order response
type OrderResponse struct {
	ID              string               `json:"id"`
	UserID          string               `json:"user_id"`
	Status          string               `json:"status"`
	Items           []*OrderItemResponse `json:"items"`
	ShippingAddress string               `json:"shipping_address"`
	Currency        string               `json:"currency"`
	TotalAmount     valueobjects.Money   `json:"total_amount"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
}

// OrderItemResponse represents an order item response
type OrderItemResponse struct {
	ProductID   string             `json:"product_id"`
	ProductName string             `json:"product_name"`
	Quantity    int32              `json:"quantity"`
	UnitPrice   valueobjects.Money `json:"unit_price"`
	TotalPrice  valueobjects.Money `json:"total_price"`
}

// OrdersListResponse represents a paginated list of orders
type OrdersListResponse struct {
	Orders []*OrderResponse `json:"orders"`
	Total  int64            `json:"total"`
	Page   int              `json:"page"`
	Limit  int              `json:"limit"`
}

// OrderStatusResponse represents order status update response
type OrderStatusResponse struct {
	OrderID   string    `json:"order_id"`
	NewStatus string    `json:"new_status"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ProcessOrderResponse represents the response after processing an order
type ProcessOrderResponse struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// OrderEventDTO represents order event data for outbox pattern
type OrderEventDTO struct {
	UserID      string              `json:"user_id"`
	Items       []*common.OrderItem `json:"items"`
	TotalAmount int64               `json:"total_amount"`
	Currency    string              `json:"currency"`
	Reason      string              `json:"reason,omitempty"` // for OrderCancelled events
}

// NewOrderItemResponse creates an OrderItemResponse from domain entity
func NewOrderItemResponse(item *entities.OrderItem) *OrderItemResponse {
	return &OrderItemResponse{
		ProductID:   item.ProductID,
		ProductName: item.ProductName,
		Quantity:    item.Quantity,
		UnitPrice:   item.UnitPrice,
		TotalPrice:  item.TotalPrice(),
	}
}

// NewOrderResponse creates an OrderResponse from Order aggregate
func NewOrderResponse(order *aggregates.Order) *OrderResponse {
	// Create items slice using the single item constructor
	items := make([]*OrderItemResponse, len(order.Items))
	for i, item := range order.Items {
		items[i] = NewOrderItemResponse(item)
	}

	return &OrderResponse{
		ID:              order.ID,
		UserID:          order.UserID,
		Status:          string(order.Status),
		Items:           items,
		ShippingAddress: order.ShippingAddress.Value,
		Currency:        order.Currency.Code,
		TotalAmount:     order.TotalAmount,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
	}
}

// CreateOrderResponse represents the response after creating an order
type CreateOrderResponse struct {
	OrderID     string `json:"order_id"`
	TotalAmount int64  `json:"total_amount"`
	Currency    string `json:"currency"`
	Status      string `json:"status"`
}
