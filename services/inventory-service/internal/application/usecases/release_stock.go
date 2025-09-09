package usecases

import (
	"context"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/pkg/outbox"
	"ecommerce-platform/services/inventory-service/internal/application/dto"
	"ecommerce-platform/services/inventory-service/internal/domain/entities"
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
		// Get all product IDs for batch query
		productIDs := make([]string, len(items))
		for i, item := range items {
			productIDs[i] = item.ProductID
		}

		// Fetch all stocks in a single query with FOR UPDATE to prevent race conditions
		stocks, err := repo.GetStocksByProductIDs(ctx, productIDs, true)
		if err != nil {
			uc.logger.Error("failed to get stocks", "productIDs", productIDs, "error", err)
			return err
		}

		// Validate all items first (fail fast)
		for _, item := range items {
			stock, exists := stocks[item.ProductID]
			if !exists {
				uc.logger.Warn("product not found", "productID", item.ProductID)
				return errors.ErrProductNotFound
			}

			if !stock.CanRelease(item.Quantity) {
				uc.logger.Warn("insufficient reserved stock to release", "productID", item.ProductID, "requested", item.Quantity, "reserved", stock.ReservedQuantity)
				return errors.ErrInsufficientReservedStock
			}
		}

		// Perform releases on all stocks
		updatedStocks := make([]*entities.Stock, 0, len(items))
		for _, item := range items {
			stock := stocks[item.ProductID]
			if err := stock.Release(item.Quantity); err != nil {
				uc.logger.Error("failed to release stock", "productID", item.ProductID, "quantity", item.Quantity, "error", err)
				return err
			}
			updatedStocks = append(updatedStocks, stock)
		}

		// Save all stock changes in a single batch operation
		if err := repo.UpsertStocks(ctx, updatedStocks); err != nil {
			uc.logger.Error("failed to save stock releases", "error", err)
			return err
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
