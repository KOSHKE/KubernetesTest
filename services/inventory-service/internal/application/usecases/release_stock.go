package usecases

import (
	"context"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/pkg/outbox"
	"ecommerce-platform/services/inventory-service/internal/application/dto"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/repository"
)

// ReleaseStockUseCase handles stock release
type ReleaseStockUseCase struct {
	inventoryRepo repository.InventoryRepositoryFacade
	outboxService outbox.Service
	logger        logger.Logger
}

// NewReleaseStockUseCase creates a new release stock use case
func NewReleaseStockUseCase(
	inventoryRepo repository.InventoryRepositoryFacade,
	outboxService outbox.Service,
	logger logger.Logger,
) *ReleaseStockUseCase {
	return &ReleaseStockUseCase{
		inventoryRepo: inventoryRepo,
		outboxService: outboxService,
		logger:        logger,
	}
}

// Execute releases stock for an order
func (uc *ReleaseStockUseCase) Execute(ctx context.Context, orderID string, items []valueobjects.Item) error {
	// Use transaction to ensure both stock release and event are saved atomically
	err := uc.inventoryRepo.WithTransaction(ctx, func(repo repository.InventoryRepositoryFacade) error {
		// Release stock for each item
		for _, item := range items {
			// Get stock by product ID
			stock, err := repo.GetStockByProductID(ctx, item.ProductID)
			if err != nil {
				uc.logger.Error("failed to get stock", "productID", item.ProductID, "error", err)
				return err
			}

			// Check if stock can be released
			if !stock.CanRelease(item.Quantity) {
				uc.logger.Warn("insufficient reserved stock to release", "productID", item.ProductID, "requested", item.Quantity, "reserved", stock.ReservedQuantity)
				return errors.ErrInsufficientReservedStock
			}

			// Release the stock
			if err := stock.Release(item.Quantity); err != nil {
				uc.logger.Error("failed to release stock", "productID", item.ProductID, "quantity", item.Quantity, "error", err)
				return err
			}

			// Update stock in repository using upsert
			if err := repo.UpsertStock(ctx, stock); err != nil {
				uc.logger.Error("failed to upsert stock after release", "productID", item.ProductID, "error", err)
				return err
			}
		}

		// Save event to outbox table (to be published later)
		eventData := dto.StockEventDTO{
			OrderID: orderID,
			Items:   items,
		}
		event := outbox.Event{
			AggregateID: orderID,
			Type:        "StockReleased",
			Payload:     eventData,
		}
		if err := uc.outboxService.SaveEvent(ctx, event); err != nil {
			uc.logger.Error("failed to save event to outbox", "orderID", orderID, "error", err)
			return err
		}

		return nil
	})

	if err != nil {
		uc.logger.Error("failed to release stock", "orderID", orderID, "error", err)
		return err
	}

	return nil
}
