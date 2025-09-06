package services

import (
	"context"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/repository"
)

// InventoryDomainService handles inventory business logic
type InventoryDomainService struct {
	inventoryRepo repository.InventoryRepository
	logger        logger.Logger
}

// NewInventoryDomainService creates a new inventory domain service
func NewInventoryDomainService(inventoryRepo repository.InventoryRepository, logger logger.Logger) *InventoryDomainService {
	return &InventoryDomainService{
		inventoryRepo: inventoryRepo,
		logger:        logger,
	}
}

// ReserveStock reserves stock for an order (all or nothing)
func (s *InventoryDomainService) ReserveStock(ctx context.Context, orderID, userID string, items []StockReservationItem) ([]string, error) {
	// Step 1: Validate all items first (fail fast)
	for _, item := range items {
		if err := s.validateItemForReservation(ctx, item); err != nil {
			s.logger.Warn("item validation failed", "productID", item.ProductID, "error", err)
			return []string{item.ProductID}, nil
		}
	}

	// Step 2: Reserve all items in a single transaction
	var unavailableProducts []string
	err := s.inventoryRepo.WithTransaction(ctx, func(repo repository.InventoryRepository) error {
		for _, item := range items {
			stock, err := repo.GetStockByProductID(ctx, item.ProductID)
			if err != nil {
				s.logger.Error("failed to get stock for reservation", "productID", item.ProductID, "error", err)
				unavailableProducts = append(unavailableProducts, item.ProductID)
				continue
			}

			if err := stock.Reserve(item.Quantity); err != nil {
				s.logger.Warn("failed to reserve stock", "productID", item.ProductID, "quantity", item.Quantity, "error", err)
				unavailableProducts = append(unavailableProducts, item.ProductID)
				continue
			}

			if err := repo.UpdateStock(ctx, stock); err != nil {
				s.logger.Error("failed to update stock after reservation", "productID", item.ProductID, "error", err)
				unavailableProducts = append(unavailableProducts, item.ProductID)
				continue
			}
		}

		// If any items failed, return error to rollback transaction
		if len(unavailableProducts) > 0 {
			return errors.ErrInsufficientStock
		}

		return nil
	})

	if err != nil {
		s.logger.Error("failed to reserve stock in transaction", "orderID", orderID, "error", err)
		return unavailableProducts, err
	}

	return nil, nil
}

// ReleaseStock releases reserved stock (all or nothing)
func (s *InventoryDomainService) ReleaseStock(ctx context.Context, orderID string, items []StockReservationItem) error {
	err := s.inventoryRepo.WithTransaction(ctx, func(repo repository.InventoryRepository) error {
		for _, item := range items {
			stock, err := repo.GetStockByProductID(ctx, item.ProductID)
			if err != nil {
				s.logger.Error("failed to get stock for release", "productID", item.ProductID, "error", err)
				return err
			}

			if err := stock.Release(item.Quantity); err != nil {
				s.logger.Error("failed to release stock", "productID", item.ProductID, "quantity", item.Quantity, "error", err)
				return err
			}

			if err := repo.UpdateStock(ctx, stock); err != nil {
				s.logger.Error("failed to update stock after release", "productID", item.ProductID, "error", err)
				return err
			}
		}

		return nil
	})

	if err != nil {
		s.logger.Error("failed to release stock in transaction", "orderID", orderID, "error", err)
		return err
	}

	return nil
}

// CommitStock commits reserved stock (all or nothing)
func (s *InventoryDomainService) CommitStock(ctx context.Context, orderID string, items []StockReservationItem) error {
	err := s.inventoryRepo.WithTransaction(ctx, func(repo repository.InventoryRepository) error {
		for _, item := range items {
			stock, err := repo.GetStockByProductID(ctx, item.ProductID)
			if err != nil {
				s.logger.Error("failed to get stock for commit", "productID", item.ProductID, "error", err)
				return err
			}

			if err := stock.Commit(item.Quantity); err != nil {
				s.logger.Error("failed to commit stock", "productID", item.ProductID, "quantity", item.Quantity, "error", err)
				return err
			}

			if err := repo.UpdateStock(ctx, stock); err != nil {
				s.logger.Error("failed to update stock after commit", "productID", item.ProductID, "error", err)
				return err
			}
		}

		return nil
	})

	if err != nil {
		s.logger.Error("failed to commit stock in transaction", "orderID", orderID, "error", err)
		return err
	}

	return nil
}

// CheckStockAvailability checks if products have sufficient stock
func (s *InventoryDomainService) CheckStockAvailability(ctx context.Context, items []StockReservationItem) ([]string, error) {
	var unavailableProducts []string

	for _, item := range items {
		product, err := s.inventoryRepo.GetProductByID(ctx, item.ProductID)
		if err != nil {
			s.logger.Error("failed to get product for availability check", "productID", item.ProductID, "error", err)
			unavailableProducts = append(unavailableProducts, item.ProductID)
			continue
		}

		if !product.IsAvailable() {
			unavailableProducts = append(unavailableProducts, item.ProductID)
			continue
		}

		// Check stock availability
		stock, err := s.inventoryRepo.GetStockByProductID(ctx, item.ProductID)
		if err != nil {
			s.logger.Error("failed to get stock for availability check", "productID", item.ProductID, "error", err)
			unavailableProducts = append(unavailableProducts, item.ProductID)
			continue
		}

		if !stock.CanReserve(item.Quantity) {
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

// validateItemForReservation validates if an item can be reserved
func (s *InventoryDomainService) validateItemForReservation(ctx context.Context, item StockReservationItem) error {
	// Check if product exists and is available
	product, err := s.inventoryRepo.GetProductByID(ctx, item.ProductID)
	if err != nil {
		return err
	}

	if !product.IsAvailable() {
		return errors.ErrProductNotAvailable
	}

	// Check if stock exists and has sufficient quantity
	stock, err := s.inventoryRepo.GetStockByProductID(ctx, item.ProductID)
	if err != nil {
		return err
	}

	if !stock.CanReserve(item.Quantity) {
		return errors.ErrInsufficientStock
	}

	return nil
}

// ReleaseStockForOrder releases stock for a specific order
func (s *InventoryDomainService) ReleaseStockForOrder(ctx context.Context, orderID string, items []StockReservationItem) error {
	return s.inventoryRepo.WithTransaction(ctx, func(repo repository.InventoryRepository) error {
		for _, item := range items {
			stock, err := repo.GetStockByProductID(ctx, item.ProductID)
			if err != nil {
				s.logger.Error("failed to get stock for release", "productID", item.ProductID, "error", err)
				return err
			}

			if err := stock.Release(item.Quantity); err != nil {
				s.logger.Error("failed to release stock", "productID", item.ProductID, "quantity", item.Quantity, "error", err)
				return err
			}

			if err := repo.UpdateStock(ctx, stock); err != nil {
				s.logger.Error("failed to update stock after release", "productID", item.ProductID, "error", err)
				return err
			}
		}

		return nil
	})
}

// CommitStockForOrder commits stock for a specific order
func (s *InventoryDomainService) CommitStockForOrder(ctx context.Context, orderID string, items []StockReservationItem) error {
	return s.inventoryRepo.WithTransaction(ctx, func(repo repository.InventoryRepository) error {
		for _, item := range items {
			stock, err := repo.GetStockByProductID(ctx, item.ProductID)
			if err != nil {
				s.logger.Error("failed to get stock for commit", "productID", item.ProductID, "error", err)
				return err
			}

			if err := stock.Commit(item.Quantity); err != nil {
				s.logger.Error("failed to commit stock", "productID", item.ProductID, "quantity", item.Quantity, "error", err)
				return err
			}

			if err := repo.UpdateStock(ctx, stock); err != nil {
				s.logger.Error("failed to update stock after commit", "productID", item.ProductID, "error", err)
				return err
			}
		}

		return nil
	})
}
