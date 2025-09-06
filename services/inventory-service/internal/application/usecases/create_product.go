package usecases

import (
	"context"

	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/pkg/idgenerator"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/services/inventory-service/internal/domain/entities"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/repository"
)

// CreateProductUseCase handles product creation
type CreateProductUseCase struct {
	inventoryRepo repository.InventoryRepositoryFacade
	logger        logger.Logger
}

// NewCreateProductUseCase creates a new CreateProductUseCase
func NewCreateProductUseCase(inventoryRepo repository.InventoryRepositoryFacade, logger logger.Logger) *CreateProductUseCase {
	return &CreateProductUseCase{
		inventoryRepo: inventoryRepo,
		logger:        logger,
	}
}

// Execute creates a new product
func (uc *CreateProductUseCase) Execute(ctx context.Context, name string, price valueobjects.Money, imageURL string) (*entities.Product, error) {
	// Generate product ID
	productID := idgenerator.GenerateID("product")

	// Create product entity
	product := entities.NewProduct(
		productID,
		name,
		price,
		imageURL,
	)

	// Validate product
	if err := product.Validate(); err != nil {
		return nil, err
	}

	// Save product
	if err := uc.inventoryRepo.CreateProduct(ctx, product); err != nil {
		uc.logger.Error("failed to create product", "productID", productID, "error", err)
		return nil, err
	}

	return product, nil
}
