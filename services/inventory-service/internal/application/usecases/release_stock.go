package usecases

import (
	"context"
	"fmt"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/proto-go/events"
	"ecommerce-platform/services/inventory-service/internal/application/dto"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/publisher"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/repository"
	"ecommerce-platform/services/inventory-service/internal/domain/services"
)

// ReleaseStockUseCase handles stock release
type ReleaseStockUseCase struct {
	inventoryRepo repository.InventoryRepository
	publisher     publisher.StockEventsPublisher
	domainService *services.InventoryDomainService
	logger        logger.Logger
}

// NewReleaseStockUseCase creates a new release stock use case
func NewReleaseStockUseCase(
	inventoryRepo repository.InventoryRepository,
	publisher publisher.StockEventsPublisher,
	domainService *services.InventoryDomainService,
	logger logger.Logger,
) *ReleaseStockUseCase {
	return &ReleaseStockUseCase{
		inventoryRepo: inventoryRepo,
		publisher:     publisher,
		domainService: domainService,
		logger:        logger,
	}
}

// Execute releases stock for an order
func (uc *ReleaseStockUseCase) Execute(ctx context.Context, req *dto.ReleaseStockRequest) (*dto.ReleaseStockResponse, error) {
	// Validate request
	if err := uc.validateRequest(req); err != nil {
		return nil, fmt.Errorf("failed to validate request: %w", err)
	}

	// Convert DTO to domain service format
	items := make([]services.StockReservationItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = services.StockReservationItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	// Release stock using domain service
	if err := uc.domainService.ReleaseStockForOrder(ctx, req.OrderID, items); err != nil {
		return nil, fmt.Errorf("failed to release stock: %w", err)
	}

	// Publish stock released event
	if err := uc.publishStockReleasedEvent(ctx, req.OrderID, req.UserID, items); err != nil {
		uc.logger.Warn("failed to publish stock released event", "error", err)
		// Don't fail the operation if event publishing fails
	}

	uc.logger.Info("stock released successfully",
		"orderID", req.OrderID,
		"itemsCount", len(items))

	return &dto.ReleaseStockResponse{
		OrderID: req.OrderID,
		Success: true,
		Message: "Stock released successfully",
	}, nil
}

// validateRequest validates the release stock request
func (uc *ReleaseStockUseCase) validateRequest(req *dto.ReleaseStockRequest) error {
	if req == nil {
		return errors.ErrInvalidRequest
	}
	if req.OrderID == "" {
		return errors.ErrInvalidRequest
	}
	if req.UserID == "" {
		return errors.ErrInvalidRequest
	}
	if len(req.Items) == 0 {
		return errors.ErrInvalidRequest
	}
	for _, item := range req.Items {
		if item.ProductID == "" {
			return errors.ErrInvalidProductID
		}
		if item.Quantity <= 0 {
			return errors.ErrInvalidQuantity
		}
	}
	return nil
}

// publishStockReleasedEvent publishes the stock released event
func (uc *ReleaseStockUseCase) publishStockReleasedEvent(ctx context.Context, orderID, userID string, items []services.StockReservationItem) error {
	event := &events.StockReleased{
		OrderId: orderID,
		UserId:  userID,
	}

	return uc.publisher.PublishStockReleased(ctx, event)
}
