package dto

import (
	"time"

	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/inventory-service/internal/domain/entities"
)

// ProductResponse represents a product response
type ProductResponse struct {
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	Price     valueobjects.Money `json:"price"`
	ImageURL  string             `json:"image_url"`
	Stock     StockInfo          `json:"stock"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

// StockInfo represents stock information
type StockInfo struct {
	AvailableQuantity int32 `json:"available_quantity"`
	ReservedQuantity  int32 `json:"reserved_quantity"`
	TotalQuantity     int32 `json:"total_quantity"`
}

// ListProductsResponse represents a paginated list of products
type ListProductsResponse struct {
	Products []ProductResponse `json:"products"`
	Total    int32             `json:"total"`
	Page     int               `json:"page"`
	Limit    int               `json:"limit"`
}

// ReserveStockResponse represents a response to stock reservation
type ReserveStockResponse struct {
	OrderID       string   `json:"order_id"`
	ReservedItems []string `json:"reserved_items"`
	FailedItems   []string `json:"failed_items"`
	Success       bool     `json:"success"`
	Message       string   `json:"message"`
}

// ReleaseStockResponse represents a response to stock release
type ReleaseStockResponse struct {
	OrderID string `json:"order_id"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// CommitStockResponse represents a response to stock commit
type CommitStockResponse struct {
	OrderID string `json:"order_id"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// NewProductResponse creates a ProductResponse from product entity with optional stock info
func NewProductResponse(product *entities.Product, stockInfo ...StockInfo) *ProductResponse {
	var stock StockInfo
	if len(stockInfo) > 0 {
		stock = stockInfo[0]
	}

	return &ProductResponse{
		ID:        product.ID,
		Name:      product.Name,
		Price:     product.Price,
		ImageURL:  product.ImageURL,
		Stock:     stock,
		CreatedAt: product.CreatedAt,
		UpdatedAt: product.UpdatedAt,
	}
}
