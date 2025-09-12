package usecases

import (
	"context"

	"ecommerce-platform/services/inventory-service/internal/domain/entities"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/repository"
)

// GetProductUseCase handles product retrieval
type GetProductUseCase struct {
	inventoryRepo repository.InventoryRepositoryFacade
}

// NewGetProductUseCase creates a new get product use case
func NewGetProductUseCase(inventoryRepo repository.InventoryRepositoryFacade) *GetProductUseCase {
	return &GetProductUseCase{
		inventoryRepo: inventoryRepo,
	}
}

// Execute retrieves a product by ID
func (uc *GetProductUseCase) Execute(ctx context.Context, productID string) (*entities.Product, error) {
	// Get product from repository
	product, err := uc.inventoryRepo.GetProductByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	return product, nil
}
