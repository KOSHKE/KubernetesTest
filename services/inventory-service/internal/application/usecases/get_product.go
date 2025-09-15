package usecases

import (
	"context"

	"ecommerce-platform/services/inventory-service/internal/domain/entities"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/repository"
)

// GetProductUseCase handles product retrieval
type GetProductUseCase struct{}

// NewGetProductUseCase creates a new get product use case
func NewGetProductUseCase() *GetProductUseCase {
	return &GetProductUseCase{}
}

// Execute retrieves a product by ID
func (uc *GetProductUseCase) Execute(ctx context.Context, productID string, repo repository.InventoryRepositoryFacade) (*entities.Product, error) {
	// Get product from repository
	product, err := repo.GetProductByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	return product, nil
}
