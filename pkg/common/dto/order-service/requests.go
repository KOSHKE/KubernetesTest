package dto

// CreateOrderRequest represents the request to create a new order
type CreateOrderRequest struct {
	UserID          string             `json:"user_id" validate:"required"`
	Items           []OrderItemRequest `json:"items" validate:"required"`
	ShippingAddress ShippingAddressDTO `json:"shipping_address" validate:"required"`
	Currency        string             `json:"currency" validate:"required"`
}

// OrderItemRequest represents an item in the order creation request
type OrderItemRequest struct {
	ProductID   string `json:"product_id" validate:"required"`
	ProductName string `json:"product_name" validate:"required"`
	Quantity    int32  `json:"quantity" validate:"required,gt=0"`
	Price       int64  `json:"price" validate:"required"`
}

// ShippingAddressDTO represents shipping address in DTO format
type ShippingAddressDTO struct {
	Address string `json:"address" validate:"required"`
}

// UpdateOrderStatusRequest represents the request to update order status
type UpdateOrderStatusRequest struct {
	OrderID string `json:"order_id" validate:"required,min=1"`
	Status  string `json:"status" validate:"required"`
}

// AddItemToOrderRequest represents the request to add an item to an order
type AddItemToOrderRequest struct {
	OrderID     string `json:"order_id" validate:"required,min=1"`
	UserID      string `json:"user_id" validate:"required,min=1"`
	ProductID   string `json:"product_id" validate:"required,min=1"`
	ProductName string `json:"product_name" validate:"required,min=1"`
	Quantity    int32  `json:"quantity" validate:"required,gt=0"`
	Price       int64  `json:"price" validate:"required"`
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
	Reason  string `json:"reason,omitempty"` // cancellation reason
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

// ProcessOrderRequest represents the request for processing an order
type ProcessOrderRequest struct {
	OrderID string `json:"order_id" validate:"required,min=1"`
}
