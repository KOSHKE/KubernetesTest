package dto

import (
	"github.com/go-playground/validator/v10"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/valueobjects"
)

// CreateOrderRequest represents the request to create a new order
type CreateOrderRequest struct {
	UserID          string             `json:"user_id" validate:"required"`
	Items           []OrderItemRequest `json:"items" validate:"required"`
	ShippingAddress string             `json:"shipping_address" validate:"required"`
	Currency        string             `json:"currency" validate:"required"`
}

// OrderItemRequest represents an item in the order creation request
type OrderItemRequest struct {
	ProductID   string              `json:"product_id" validate:"required"`
	ProductName string              `json:"product_name" validate:"required"`
	Quantity    int32               `json:"quantity" validate:"required"`
	Price       *valueobjects.Money `json:"price,omitempty" validate:"omitempty"`
}

// UpdateOrderStatusRequest represents the request to update order status
type UpdateOrderStatusRequest struct {
	OrderID string                   `json:"order_id" validate:"required,min=1"`
	Status  valueobjects.OrderStatus `json:"status" validate:"required,valid_order_status"`
}

// AddItemToOrderRequest represents the request to add an item to an order
type AddItemToOrderRequest struct {
	OrderID     string              `json:"order_id" validate:"required,min=1"`
	UserID      string              `json:"user_id" validate:"required,min=1"`
	ProductID   string              `json:"product_id" validate:"required,min=1"`
	ProductName string              `json:"product_name" validate:"required,min=1"`
	Quantity    int32               `json:"quantity" validate:"required,gt=0"`
	Price       *valueobjects.Money `json:"price" validate:"required"`
}

// RemoveItemFromOrderRequest represents the request to remove an item from an order
type RemoveItemFromOrderRequest struct {
	OrderID   string `json:"order_id" validate:"required,min=1"`
	UserID    string `json:"user_id" validate:"required,min=1"`
	ProductID string `json:"product_id" validate:"required,min=1"`
}

// CancelOrderRequest represents the request to cancel an order
type CancelOrderRequest struct {
	OrderID string `json:"order_id" validate:"required,min=1"`
	UserID  string `json:"user_id" validate:"required,min=1"`
}

// GetOrderRequest represents the request to get an order
type GetOrderRequest struct {
	OrderID string `json:"order_id" validate:"required,min=1"`
	UserID  string `json:"user_id" validate:"required,min=1"`
}

// GetUserOrdersRequest represents the request to get user orders with pagination
type GetUserOrdersRequest struct {
	UserID string `json:"user_id" validate:"required,min=1"`
	Page   int    `json:"page" validate:"gte=1"`
	Limit  int    `json:"limit" validate:"gte=1,lte=100"`
}

// RegisterCustomValidators registers custom validation functions
func RegisterCustomValidators(v *validator.Validate) {
	v.RegisterValidation("valid_order_status", validateOrderStatus)
	v.RegisterValidation("valid_money", validateMoney)
}

// validateOrderStatus validates that the order status is one of the valid statuses
func validateOrderStatus(fl validator.FieldLevel) bool {
	status, ok := fl.Field().Interface().(valueobjects.OrderStatus)
	if !ok {
		return false
	}
	return status.IsValid()
}

// validateMoney validates that the money value is valid
func validateMoney(fl validator.FieldLevel) bool {
	money, ok := fl.Field().Interface().(*valueobjects.Money)
	if !ok || money == nil {
		return true // nil is valid for omitempty
	}
	return money.Amount >= 0
}
