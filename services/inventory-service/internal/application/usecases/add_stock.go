package usecases

import (
	"context"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/services/inventory-service/internal/domain/entities"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/repository"
)

// AddStockUseCase handles adding stock to a product
type AddStockUseCase struct {
	inventoryRepo repository.InventoryRepositoryFacade
	logger        logger.Logger
}

// NewAddStockUseCase creates a new AddStockUseCase
func NewAddStockUseCase(inventoryRepo repository.InventoryRepositoryFacade, logger logger.Logger) *AddStockUseCase {
	return &AddStockUseCase{
		inventoryRepo: inventoryRepo,
		logger:        logger,
	}
}

// Execute adds stock to a product
func (uc *AddStockUseCase) Execute(ctx context.Context, productID string, quantity int32) (*entities.Stock, error) {
	// Check if product exists
	product, err := uc.inventoryRepo.GetProductByID(ctx, productID)
	if err != nil {
		uc.logger.Error("failed to get product", "productID", productID, "error", err)
		return nil, err
	}
	if product == nil {
		return nil, errors.ErrProductNotFound
	}

	// Create new stock with the quantity to add
	stock := entities.NewStock(productID, quantity, 0)

	// Validate stock
	if err := stock.Validate(); err != nil {
		return nil, err
	}

	if err := uc.inventoryRepo.UpsertStock(ctx, stock); err != nil {
		uc.logger.Error("failed to upsert stock", "productID", productID, "error", err)
		return nil, err
	}

	return stock, nil
}
