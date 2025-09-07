package dto

import "ecommerce-platform/pkg/common/valueobjects"

// CreateProductRequest represents a request to create a product
type CreateProductRequest struct {
	Name     string             `json:"name" validate:"required,min=1,max=255"`
	Price    valueobjects.Money `json:"price" validate:"required"`
	ImageURL string             `json:"image_url" validate:"max=500"`
	Stock    int32              `json:"stock" validate:"min=0"`
}

// UpdateProductRequest represents a request to update a product
type UpdateProductRequest struct {
	Name     string             `json:"name" validate:"omitempty,min=1,max=255"`
	Price    valueobjects.Money `json:"price" validate:"omitempty"`
	ImageURL string             `json:"image_url" validate:"omitempty,max=500"`
}

// ReserveStockRequest represents a request to reserve stock
type ReserveStockRequest struct {
	OrderID string                 `json:"order_id" validate:"required,min=1,max=255"`
	Items   []StockReservationItem `json:"items" validate:"required,min=1,dive"`
}

// StockReservationItem represents an item for stock reservation
type StockReservationItem struct {
	ProductID string `json:"product_id" validate:"required,min=1,max=255"`
	Quantity  int32  `json:"quantity" validate:"required,min=1"`
}

// ListProductsRequest represents a request to list products
type ListProductsRequest struct {
	Page   int    `json:"page" validate:"min=1"`
	Limit  int    `json:"limit" validate:"min=1,max=100"`
	Search string `json:"search" validate:"max=255"`
}

// ReleaseStockRequest represents a request to release stock
type ReleaseStockRequest struct {
	OrderID string                 `json:"order_id" validate:"required,min=1,max=255"`
	Items   []StockReservationItem `json:"items" validate:"required,min=1,dive"`
}

// CommitStockRequest represents a request to commit stock
type CommitStockRequest struct {
	OrderID string                 `json:"order_id" validate:"required,min=1,max=255"`
	Items   []StockReservationItem `json:"items" validate:"required,min=1,dive"`
}
