package usecases

import (
	"context"

	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/pkg/idgenerator"
	"ecommerce-platform/services/inventory-service/internal/domain/entities"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/repository"
)

// CreateProductUseCase handles product creation
type CreateProductUseCase struct{}

// NewCreateProductUseCase creates a new CreateProductUseCase
func NewCreateProductUseCase() *CreateProductUseCase {
	return &CreateProductUseCase{}
}

// Execute creates a new product
func (uc *CreateProductUseCase) Execute(ctx context.Context, name string, price valueobjects.Money, imageURL string, repo repository.InventoryRepositoryFacade) (*entities.Product, error) {
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
	if err := repo.CreateProduct(ctx, product); err != nil {
		return nil, err
	}

	return product, nil
}
