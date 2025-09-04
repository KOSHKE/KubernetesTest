package dto

import (
	"time"

	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/inventory-service/internal/domain/entities"
)

// ProductResponse represents a product response
type ProductResponse struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Price       *valueobjects.Money `json:"price"`
	ImageURL    string              `json:"image_url"`
	IsActive    bool                `json:"is_active"`
	Stock       StockInfo           `json:"stock"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
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

// FromProductEntity converts a product entity to response DTO
func FromProductEntity(product *entities.Product) ProductResponse {
	return ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		ImageURL:    product.ImageURL,
		IsActive:    product.IsActive,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}
}
