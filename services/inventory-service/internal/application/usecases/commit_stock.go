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

// CommitStockUseCase handles stock commit
type CommitStockUseCase struct {
	inventoryRepo repository.InventoryRepositoryFacade
	outboxService outbox.Service
	logger        logger.Logger
}

// NewCommitStockUseCase creates a new commit stock use case
func NewCommitStockUseCase(
	inventoryRepo repository.InventoryRepositoryFacade,
	outboxService outbox.Service,
	logger logger.Logger,
) *CommitStockUseCase {
	return &CommitStockUseCase{
		inventoryRepo: inventoryRepo,
		outboxService: outboxService,
		logger:        logger,
	}
}

// Execute commits stock for an order
func (uc *CommitStockUseCase) Execute(ctx context.Context, orderID string, items []valueobjects.Item) error {
	// Use transaction to ensure both stock commit and event are saved atomically
	err := uc.inventoryRepo.WithTransaction(ctx, func(repo repository.InventoryRepositoryFacade) error {
		// Validate and commit stock for each item
		for _, item := range items {
			// Get stock by product ID
			stock, err := repo.GetStockByProductID(ctx, item.ProductID)
			if err != nil {
				uc.logger.Error("failed to get stock", "productID", item.ProductID, "error", err)
				return err
			}

			// Check if stock can be committed
			if !stock.CanCommit(item.Quantity) {
				uc.logger.Warn("insufficient reserved stock", "productID", item.ProductID, "requested", item.Quantity, "reserved", stock.ReservedQuantity)
				return errors.ErrInsufficientReservedStock
			}

			// Commit the stock
			if err := stock.Commit(item.Quantity); err != nil {
				uc.logger.Error("failed to commit stock", "productID", item.ProductID, "quantity", item.Quantity, "error", err)
				return err
			}

			// Update stock in repository using upsert
			if err := repo.UpsertStock(ctx, stock); err != nil {
				uc.logger.Error("failed to upsert stock after commit", "productID", item.ProductID, "error", err)
				return err
			}
		}

		// Save event to outbox table (to be published later)
		eventData := dto.StockEventDTO{
			OrderID: orderID,
			UserID:  "", // UserID not needed for stock operations
			Items:   items,
		}
		event := outbox.Event{
			Type:    "StockCommitted",
			Payload: eventData,
		}
		if err := uc.outboxService.SaveEvent(ctx, event); err != nil {
			uc.logger.Error("failed to save event to outbox", "orderID", orderID, "error", err)
			return err
		}

		return nil
	})

	if err != nil {
		uc.logger.Error("failed to commit stock", "orderID", orderID, "error", err)
		return err
	}

	return nil
}
