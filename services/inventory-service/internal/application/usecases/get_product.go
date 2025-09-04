package usecases

import (
	"context"
	"fmt"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/services/inventory-service/internal/application/dto"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/repository"
)

// GetProductUseCase handles product retrieval
type GetProductUseCase struct {
	repo   repository.InventoryRepository
	logger logger.Logger
}

// NewGetProductUseCase creates a new get product use case
func NewGetProductUseCase(repo repository.InventoryRepository, logger logger.Logger) *GetProductUseCase {
	return &GetProductUseCase{
		repo:   repo,
		logger: logger,
	}
}

// Execute retrieves a product by ID
func (uc *GetProductUseCase) Execute(ctx context.Context, productID string) (*dto.ProductResponse, error) {
	// Validate input
	if productID == "" {
		return nil, errors.ErrInvalidProductID
	}

	// Get product from repository
	product, err := uc.repo.GetProductByID(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	if product == nil {
		return nil, errors.ErrProductNotFound
	}

	uc.logger.Info("product retrieved successfully", "productID", productID)

	// Convert to response
	response := &dto.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		ImageURL:    product.ImageURL,
		IsActive:    product.IsActive,
		Stock: dto.StockInfo{
			AvailableQuantity: product.GetAvailableQuantity(),
			ReservedQuantity:  product.GetReservedQuantity(),
			TotalQuantity:     product.GetTotalQuantity(),
		},
		CreatedAt: product.CreatedAt,
		UpdatedAt: product.UpdatedAt,
	}

	return response, nil
}
