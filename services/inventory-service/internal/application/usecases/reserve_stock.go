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

// ReserveStockUseCase handles stock reservation
type ReserveStockUseCase struct {
	inventoryRepo repository.InventoryRepositoryFacade
	outboxService outbox.Service
	logger        logger.Logger
}

// NewReserveStockUseCase creates a new reserve stock use case
func NewReserveStockUseCase(
	inventoryRepo repository.InventoryRepositoryFacade,
	outboxService outbox.Service,
	logger logger.Logger,
) *ReserveStockUseCase {
	return &ReserveStockUseCase{
		inventoryRepo: inventoryRepo,
		outboxService: outboxService,
		logger:        logger,
	}
}

// Execute reserves stock for an order
func (uc *ReserveStockUseCase) Execute(ctx context.Context, orderID string, items []valueobjects.Item) error {
	// Use transaction to ensure both stock reservation and event are saved atomically
	err := uc.inventoryRepo.WithTransaction(ctx, func(repo repository.InventoryRepositoryFacade) error {
		// Reserve stock using the repository from transaction context
		if err := uc.reserveStock(ctx, repo, items); err != nil {
			return err
		}

		// Save event to outbox table (to be published later)
		eventData := dto.StockEventDTO{
			OrderID: orderID,
			Items:   items,
		}
		event := outbox.Event{
			AggregateID: orderID,
			Type:        "StockReserved",
			Payload:     eventData,
		}
		if err := uc.outboxService.SaveEvent(ctx, event); err != nil {
			uc.logger.Error("failed to save event to outbox", "orderID", orderID, "error", err)
			return err
		}

		return nil
	})

	if err != nil {
		uc.logger.Error("failed to reserve stock", "orderID", orderID, "error", err)
		return err
	}

	return nil
}

// reserveStock performs stock reservation using the provided repository (within transaction)
func (uc *ReserveStockUseCase) reserveStock(ctx context.Context, repo repository.InventoryRepositoryFacade, items []valueobjects.Item) error {
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

		if !stock.CanReserve(item.Quantity) {
			uc.logger.Warn("insufficient stock", "productID", item.ProductID, "requested", item.Quantity, "available", stock.AvailableQuantity)
			return errors.ErrInsufficientStock
		}
	}

	// Perform reservations on all stocks
	updatedStocks := make([]*entities.Stock, 0, len(items))
	for _, item := range items {
		stock := stocks[item.ProductID]
		if err := stock.Reserve(item.Quantity); err != nil {
			uc.logger.Error("failed to reserve stock", "productID", item.ProductID, "quantity", item.Quantity, "error", err)
			return err
		}
		updatedStocks = append(updatedStocks, stock)
	}

	// Save all stock changes in a single batch operation
	if err := repo.UpsertStocks(ctx, updatedStocks); err != nil {
		uc.logger.Error("failed to save stock reservations", "error", err)
		return err
	}

	return nil
}
