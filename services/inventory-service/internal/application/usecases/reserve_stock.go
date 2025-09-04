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

// ReserveStockUseCase handles stock reservation
type ReserveStockUseCase struct {
	repo          repository.InventoryRepository
	publisher     publisher.StockEventsPublisher
	domainService *services.InventoryDomainService
	logger        logger.Logger
}

// NewReserveStockUseCase creates a new reserve stock use case
func NewReserveStockUseCase(
	repo repository.InventoryRepository,
	publisher publisher.StockEventsPublisher,
	domainService *services.InventoryDomainService,
	logger logger.Logger,
) *ReserveStockUseCase {
	return &ReserveStockUseCase{
		repo:          repo,
		publisher:     publisher,
		domainService: domainService,
		logger:        logger,
	}
}

// Execute reserves stock for an order
func (uc *ReserveStockUseCase) Execute(ctx context.Context, req *dto.ReserveStockRequest) (*dto.ReserveStockResponse, error) {
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

	// Reserve stock using domain service
	failedProducts, err := uc.domainService.ReserveStock(ctx, req.OrderID, req.UserID, items)
	if err != nil {
		return nil, fmt.Errorf("failed to reserve stock: %w", err)
	}

	// Determine success
	success := len(failedProducts) == 0
	var message string
	if success {
		message = "Stock reserved successfully"
	} else {
		message = fmt.Sprintf("Stock reservation partially failed. Failed products: %v", failedProducts)
	}

	// Get reserved products (those that succeeded)
	var reservedProducts []string
	for _, item := range req.Items {
		isFailed := false
		for _, failed := range failedProducts {
			if failed == item.ProductID {
				isFailed = true
				break
			}
		}
		if !isFailed {
			reservedProducts = append(reservedProducts, item.ProductID)
		}
	}

	// Publish appropriate event
	if err := uc.publishEvent(ctx, req.OrderID, req.UserID, items, failedProducts); err != nil {
		uc.logger.Warn("failed to publish stock event", "error", err)
		// Don't fail the operation if event publishing fails
	}

	uc.logger.Info("stock reservation completed",
		"orderID", req.OrderID,
		"reservedCount", len(reservedProducts),
		"failedCount", len(failedProducts),
		"success", success)

	return &dto.ReserveStockResponse{
		OrderID:       req.OrderID,
		ReservedItems: reservedProducts,
		FailedItems:   failedProducts,
		Success:       success,
		Message:       message,
	}, nil
}

// validateRequest validates the reserve stock request
func (uc *ReserveStockUseCase) validateRequest(req *dto.ReserveStockRequest) error {
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

// publishEvent publishes the appropriate stock event
func (uc *ReserveStockUseCase) publishEvent(ctx context.Context, orderID, userID string, items []services.StockReservationItem, failedProducts []string) error {
	if len(failedProducts) == 0 {
		// All items reserved successfully
		event := &events.StockReserved{
			OrderId: orderID,
			UserId:  userID,
			Items:   uc.convertItemsToProto(items),
		}
		return uc.publisher.PublishStockReserved(ctx, event)
	} else {
		// Some items failed to reserve
		event := &events.StockReservationFailed{
			OrderId:       orderID,
			UserId:        userID,
			FailedItems:   failedProducts,
			ReservedItems: uc.getReservedItems(items, failedProducts),
		}
		return uc.publisher.PublishStockReservationFailed(ctx, event)
	}
}

// convertItemsToProto converts domain items to proto items
func (uc *ReserveStockUseCase) convertItemsToProto(items []services.StockReservationItem) []*events.StockItem {
	protoItems := make([]*events.StockItem, len(items))
	for i, item := range items {
		protoItems[i] = &events.StockItem{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
		}
	}
	return protoItems
}

// getReservedItems gets the list of successfully reserved items
func (uc *ReserveStockUseCase) getReservedItems(items []services.StockReservationItem, failedProducts []string) []*events.StockItem {
	var reservedItems []*events.StockItem
	for _, item := range items {
		isFailed := false
		for _, failed := range failedProducts {
			if failed == item.ProductID {
				isFailed = true
				break
			}
		}
		if !isFailed {
			reservedItems = append(reservedItems, &events.StockItem{
				ProductId: item.ProductID,
				Quantity:  item.Quantity,
			})
		}
	}
	return reservedItems
}
