package usecases

import (
	"context"
	"fmt"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/services/inventory-service/internal/application/dto"
	"ecommerce-platform/services/inventory-service/internal/domain/aggregates"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/repository"
)

// CreateProductUseCase handles product creation
type CreateProductUseCase struct {
	repo   repository.InventoryRepository
	logger logger.Logger
}

// NewCreateProductUseCase creates a new create product use case
func NewCreateProductUseCase(repo repository.InventoryRepository, logger logger.Logger) *CreateProductUseCase {
	return &CreateProductUseCase{
		repo:   repo,
		logger: logger,
	}
}

// Execute creates a new product
func (uc *CreateProductUseCase) Execute(ctx context.Context, req *dto.CreateProductRequest) (*dto.ProductResponse, error) {
	// Validate request
	if err := uc.validateRequest(req); err != nil {
		return nil, fmt.Errorf("failed to validate request: %w", err)
	}

	// Check if product already exists
	existingProduct, err := uc.repo.GetProductByID(ctx, req.Name) // Using name as ID for now
	if err == nil && existingProduct != nil {
		return nil, errors.ErrProductAlreadyExists
	}

	// Create product aggregate
	product := aggregates.NewProduct(
		req.Name, // Using name as ID for now, should be generated
		req.Name,
		req.Description,
		req.Price,
		req.ImageURL,
		req.Stock,
		0, // No reserved quantity initially
	)

	// Validate product
	if err := product.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate product: %w", err)
	}

	// Save product
	if err := uc.repo.CreateProduct(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	uc.logger.Info("product created successfully", "productID", product.ID, "name", product.Name)

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

// validateRequest validates the create product request
func (uc *CreateProductUseCase) validateRequest(req *dto.CreateProductRequest) error {
	if req == nil {
		return errors.ErrInvalidRequest
	}
	if req.Name == "" {
		return errors.ErrInvalidProductName
	}
	if req.Price == nil {
		return errors.ErrInvalidProductPrice
	}
	if req.Stock < 0 {
		return errors.ErrInvalidQuantity
	}
	return nil
}
