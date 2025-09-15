package usecases

import (
	"context"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/inventory-service/internal/domain/entities"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/repository"
)

// ReleaseStockUseCase handles stock release
type ReleaseStockUseCase struct{}

// NewReleaseStockUseCase creates a new release stock use case
func NewReleaseStockUseCase() *ReleaseStockUseCase {
	return &ReleaseStockUseCase{}
}

// Execute releases stock for an order
func (uc *ReleaseStockUseCase) Execute(ctx context.Context, orderID string, items []valueobjects.Item, repo repository.InventoryRepositoryFacade) error {
	// Get all product IDs for batch query
	productIDs := make([]string, len(items))
	for i, item := range items {
		productIDs[i] = item.ProductID
	}

	// Fetch all stocks in a single query with FOR UPDATE to prevent race conditions
	stocks, err := repo.GetStocksByProductIDs(ctx, productIDs, true)
	if err != nil {
		return err
	}

	// Validate all items first (fail fast)
	for _, item := range items {
		stock, exists := stocks[item.ProductID]
		if !exists {
			return errors.ErrProductNotFound
		}

		if !stock.CanRelease(item.Quantity) {
			return errors.ErrInsufficientReservedStock
		}
	}

	// Perform releases on all stocks
	updatedStocks := make([]*entities.Stock, 0, len(items))
	for _, item := range items {
		stock := stocks[item.ProductID]
		if err := stock.Release(item.Quantity); err != nil {
			return err
		}
		updatedStocks = append(updatedStocks, stock)
	}

	// Save all stock changes in a single batch operation
	if err := repo.UpsertStocks(ctx, updatedStocks); err != nil {
		return err
	}

	return nil
}
