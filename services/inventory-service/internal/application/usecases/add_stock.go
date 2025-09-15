package usecases

import (
	"context"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/services/inventory-service/internal/domain/entities"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/repository"
)

// AddStockUseCase handles adding stock to a product
type AddStockUseCase struct{}

// NewAddStockUseCase creates a new AddStockUseCase
func NewAddStockUseCase() *AddStockUseCase {
	return &AddStockUseCase{}
}

// Execute adds stock to a product
func (uc *AddStockUseCase) Execute(ctx context.Context, productID string, quantity int32, repo repository.InventoryRepositoryFacade) (*entities.Stock, error) {
	// Check if product exists
	product, err := repo.GetProductByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.ErrProductNotFound
	}

	// Get existing stock or create new one
	stock, err := repo.GetStockByProductID(ctx, productID)
	if err != nil {
		// If stock doesn't exist, create new one
		stock = entities.NewStock(productID, 0, 0)
	}

	// Add quantity to existing stock
	stock.AddQuantity(quantity)

	// Validate stock
	if err := stock.Validate(); err != nil {
		return nil, err
	}

	if err := repo.UpsertStock(ctx, stock); err != nil {
		return nil, err
	}

	return stock, nil
}
