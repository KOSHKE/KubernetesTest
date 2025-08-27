package dto

import (
	"time"

	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/aggregates"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/valueobjects"
)

// OrderResponse represents the order response
type OrderResponse struct {
	ID              string               `json:"id"`
	UserID          string               `json:"user_id"`
	Number          int64                `json:"number"`
	Status          string               `json:"status"`
	Items           []*OrderItemResponse `json:"items"`
	ShippingAddress string               `json:"shipping_address"`
	Currency        string               `json:"currency"`
	TotalAmount     *valueobjects.Money  `json:"total_amount"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
}

// OrderItemResponse represents an order item response
type OrderItemResponse struct {
	ProductID   string              `json:"product_id"`
	ProductName string              `json:"product_name"`
	Quantity    int32               `json:"quantity"`
	UnitPrice   *valueobjects.Money `json:"unit_price"`
	TotalPrice  *valueobjects.Money `json:"total_price"`
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

// NewOrderResponse creates a new OrderResponse from Order aggregate
func NewOrderResponse(order *aggregates.Order) *OrderResponse {
	items := make([]*OrderItemResponse, len(order.Items))
	for i, item := range order.Items {
		items[i] = &OrderItemResponse{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			TotalPrice:  item.TotalPrice(),
		}
	}

	return &OrderResponse{
		ID:              order.ID,
		UserID:          order.UserID,
		Number:          order.Number,
		Status:          string(order.Status),
		Items:           items,
		ShippingAddress: order.ShippingAddress,
		Currency:        order.Currency,
		TotalAmount:     order.TotalAmount,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
	}
}

// NewOrdersListResponse creates a new OrdersListResponse
func NewOrdersListResponse(orders []*aggregates.Order, total, page, limit int64) *OrdersListResponse {
	orderResponses := make([]*OrderResponse, len(orders))
	for i, order := range orders {
		orderResponses[i] = NewOrderResponse(order)
	}

	return &OrdersListResponse{
		Orders: orderResponses,
		Total:  total,
		Page:   int(page),
		Limit:  int(limit),
	}
}

// NewOrderStatusResponse creates a new OrderStatusResponse
func NewOrderStatusResponse(orderID string, newStatus valueobjects.OrderStatus, updatedAt time.Time) *OrderStatusResponse {
	return &OrderStatusResponse{
		OrderID:   orderID,
		NewStatus: string(newStatus),
		UpdatedAt: updatedAt,
	}
}
