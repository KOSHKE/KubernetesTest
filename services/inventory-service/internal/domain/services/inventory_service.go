package services

import (
	"context"

	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/repository"
)

// InventoryDomainService handles inventory business logic
type InventoryDomainService struct {
	repo   repository.InventoryRepository
	logger logger.Logger
}

// NewInventoryDomainService creates a new inventory domain service
func NewInventoryDomainService(repo repository.InventoryRepository, logger logger.Logger) *InventoryDomainService {
	return &InventoryDomainService{
		repo:   repo,
		logger: logger,
	}
}

// ReserveStock reserves stock for an order
func (s *InventoryDomainService) ReserveStock(ctx context.Context, orderID, userID string, items []StockReservationItem) ([]string, error) {
	var failedProducts []string

	for _, item := range items {
		product, err := s.repo.GetProductByID(ctx, item.ProductID)
		if err != nil {
			s.logger.Error("failed to get product for reservation", "productID", item.ProductID, "error", err)
			failedProducts = append(failedProducts, item.ProductID)
			continue
		}

		if !product.IsAvailableForPurchase() {
			s.logger.Warn("product not available for purchase", "productID", item.ProductID)
			failedProducts = append(failedProducts, item.ProductID)
			continue
		}

		if err := product.ReserveStock(item.Quantity); err != nil {
			s.logger.Warn("failed to reserve stock", "productID", item.ProductID, "quantity", item.Quantity, "error", err)
			failedProducts = append(failedProducts, item.ProductID)
			continue
		}

		// Update stock in repository
		if err := s.repo.UpdateStock(ctx, item.ProductID, product.GetAvailableQuantity(), product.GetReservedQuantity()); err != nil {
			s.logger.Error("failed to update stock after reservation", "productID", item.ProductID, "error", err)
			// Rollback the reservation
			product.ReleaseStock(item.Quantity)
			failedProducts = append(failedProducts, item.ProductID)
			continue
		}

		s.logger.Info("stock reserved successfully", "productID", item.ProductID, "quantity", item.Quantity)
	}

	return failedProducts, nil
}

// ReleaseStock releases reserved stock
func (s *InventoryDomainService) ReleaseStock(ctx context.Context, orderID string, items []StockReservationItem) error {
	for _, item := range items {
		product, err := s.repo.GetProductByID(ctx, item.ProductID)
		if err != nil {
			s.logger.Error("failed to get product for release", "productID", item.ProductID, "error", err)
			continue
		}

		product.ReleaseStock(item.Quantity)

		// Update stock in repository
		if err := s.repo.UpdateStock(ctx, item.ProductID, product.GetAvailableQuantity(), product.GetReservedQuantity()); err != nil {
			s.logger.Error("failed to update stock after release", "productID", item.ProductID, "error", err)
			continue
		}

		s.logger.Info("stock released successfully", "productID", item.ProductID, "quantity", item.Quantity)
	}

	return nil
}

// CommitStock commits reserved stock (removes it completely)
func (s *InventoryDomainService) CommitStock(ctx context.Context, orderID string, items []StockReservationItem) error {
	for _, item := range items {
		product, err := s.repo.GetProductByID(ctx, item.ProductID)
		if err != nil {
			s.logger.Error("failed to get product for commit", "productID", item.ProductID, "error", err)
			continue
		}

		product.CommitStock(item.Quantity)

		// Update stock in repository
		if err := s.repo.UpdateStock(ctx, item.ProductID, product.GetAvailableQuantity(), product.GetReservedQuantity()); err != nil {
			s.logger.Error("failed to update stock after commit", "productID", item.ProductID, "error", err)
			continue
		}

		s.logger.Info("stock committed successfully", "productID", item.ProductID, "quantity", item.Quantity)
	}

	return nil
}

// CheckStockAvailability checks if products have sufficient stock
func (s *InventoryDomainService) CheckStockAvailability(ctx context.Context, items []StockReservationItem) ([]string, error) {
	var unavailableProducts []string

	for _, item := range items {
		product, err := s.repo.GetProductByID(ctx, item.ProductID)
		if err != nil {
			s.logger.Error("failed to get product for availability check", "productID", item.ProductID, "error", err)
			unavailableProducts = append(unavailableProducts, item.ProductID)
			continue
		}

		if !product.IsAvailableForPurchase() || !product.CanReserve(item.Quantity) {
			unavailableProducts = append(unavailableProducts, item.ProductID)
		}
	}

	return unavailableProducts, nil
}

// StockReservationItem represents an item for stock reservation
type StockReservationItem struct {
	ProductID string
	Quantity  int32
}
