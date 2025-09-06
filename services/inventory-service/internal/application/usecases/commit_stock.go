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

// CommitStockUseCase handles stock commit
type CommitStockUseCase struct {
	inventoryRepo repository.InventoryRepository
	publisher     publisher.StockEventsPublisher
	domainService *services.InventoryDomainService
	logger        logger.Logger
}

// NewCommitStockUseCase creates a new commit stock use case
func NewCommitStockUseCase(
	inventoryRepo repository.InventoryRepository,
	publisher publisher.StockEventsPublisher,
	domainService *services.InventoryDomainService,
	logger logger.Logger,
) *CommitStockUseCase {
	return &CommitStockUseCase{
		inventoryRepo: inventoryRepo,
		publisher:     publisher,
		domainService: domainService,
		logger:        logger,
	}
}

// Execute commits stock for an order
func (uc *CommitStockUseCase) Execute(ctx context.Context, req *dto.CommitStockRequest) (*dto.CommitStockResponse, error) {
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

	// Commit stock using domain service
	if err := uc.domainService.CommitStockForOrder(ctx, req.OrderID, items); err != nil {
		return nil, fmt.Errorf("failed to commit stock: %w", err)
	}

	// Publish stock committed event
	if err := uc.publishStockCommittedEvent(ctx, req.OrderID, req.UserID, items); err != nil {
		uc.logger.Warn("failed to publish stock committed event", "error", err)
		// Don't fail the operation if event publishing fails
	}

	uc.logger.Info("stock committed successfully",
		"orderID", req.OrderID,
		"itemsCount", len(items))

	return &dto.CommitStockResponse{
		OrderID: req.OrderID,
		Success: true,
		Message: "Stock committed successfully",
	}, nil
}

// validateRequest validates the commit stock request
func (uc *CommitStockUseCase) validateRequest(req *dto.CommitStockRequest) error {
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

// publishStockCommittedEvent publishes the stock committed event
func (uc *CommitStockUseCase) publishStockCommittedEvent(ctx context.Context, orderID, userID string, items []services.StockReservationItem) error {
	event := &events.StockCommitted{
		OrderId: orderID,
		UserId:  userID,
	}

	return uc.publisher.PublishStockCommitted(ctx, event)
}
