package usecases

import (
	"context"
	"fmt"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/services/inventory-service/internal/application/dto"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/repository"
)

// GetProductUseCase handles product retrieval
type GetProductUseCase struct {
	inventoryRepo repository.InventoryRepository
	logger        logger.Logger
}

// NewGetProductUseCase creates a new get product use case
func NewGetProductUseCase(inventoryRepo repository.InventoryRepository, logger logger.Logger) *GetProductUseCase {
	return &GetProductUseCase{
		inventoryRepo: inventoryRepo,
		logger:        logger,
	}
}

// Execute retrieves a product by ID
func (uc *GetProductUseCase) Execute(ctx context.Context, productID string) (*dto.ProductResponse, error) {
	// Validate input
	if productID == "" {
		return nil, errors.ErrInvalidProductID
	}

	// Get product from repository
	product, err := uc.inventoryRepo.GetProductByID(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	if product == nil {
		return nil, errors.ErrProductNotFound
	}

	// Get stock information for this product
	stock, err := uc.inventoryRepo.GetStockByProductID(ctx, productID)
	var stockInfo dto.StockInfo
	if err != nil {
		uc.logger.Warn("failed to get stock info for product", "productID", productID, "error", err)
		stockInfo = dto.StockInfo{
			AvailableQuantity: 0,
			ReservedQuantity:  0,
			TotalQuantity:     0,
		}
	} else {
		stockInfo = dto.StockInfo{
			AvailableQuantity: stock.AvailableQuantity,
			ReservedQuantity:  stock.ReservedQuantity,
			TotalQuantity:     stock.AvailableQuantity + stock.ReservedQuantity,
		}
	}

	uc.logger.Info("product retrieved successfully", "productID", productID)

	// Convert to response
	response := &dto.ProductResponse{
		ID:        product.ID,
		Name:      product.Name,
		Price:     product.Price,
		ImageURL:  product.ImageURL,
		Stock:     stockInfo,
		CreatedAt: product.CreatedAt,
		UpdatedAt: product.UpdatedAt,
	}

	return response, nil
}
