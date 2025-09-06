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
		// Reserve stock
		if err := uc.reserveStock(ctx, items); err != nil {
			return err
		}

		// Save event to outbox table (to be published later)
		eventData := dto.StockEventDTO{
			OrderID: orderID,
			UserID:  "", // UserID not needed for stock operations
			Items:   items,
		}
		event := outbox.Event{
			Type:    "StockReserved",
			Payload: eventData,
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

// reserveStock performs stock reservation using repository directly
func (uc *ReserveStockUseCase) reserveStock(ctx context.Context, items []valueobjects.Item) error {
	// Validate all items first (fail fast)
	for _, item := range items {
		// Check stock availability
		stock, err := uc.inventoryRepo.GetStockByProductID(ctx, item.ProductID)
		if err != nil {
			uc.logger.Error("failed to get stock", "productID", item.ProductID, "error", err)
			return err
		}

		if !stock.CanReserve(item.Quantity) {
			uc.logger.Warn("insufficient stock", "productID", item.ProductID, "requested", item.Quantity, "available", stock.AvailableQuantity)
			return errors.ErrInsufficientStock
		}
	}

	// Perform reservations
	for _, item := range items {
		stock, err := uc.inventoryRepo.GetStockByProductID(ctx, item.ProductID)
		if err != nil {
			uc.logger.Error("failed to get stock for reservation", "productID", item.ProductID, "error", err)
			return err
		}

		if err := stock.Reserve(item.Quantity); err != nil {
			uc.logger.Error("failed to reserve stock", "productID", item.ProductID, "quantity", item.Quantity, "error", err)
			return err
		}

		if err := uc.inventoryRepo.UpsertStock(ctx, stock); err != nil {
			uc.logger.Error("failed to upsert stock after reservation", "productID", item.ProductID, "error", err)
			return err
		}
	}

	return nil
}
