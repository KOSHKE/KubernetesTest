package usecases_test

import (
	"context"
	"testing"

	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/inventory-service/internal/application/usecases"
	"ecommerce-platform/services/inventory-service/internal/domain/entities"
	"ecommerce-platform/services/inventory-service/tests/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGetProductUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockInventoryRepositoryFacade(ctrl)

	useCase := usecases.NewGetProductUseCase(mockRepo)

	ctx := context.Background()
	productID := "product-123"

	// Create test product
	currency, _ := valueobjects.NewCurrency("USD")
	price := valueobjects.NewMoney(1000, currency)
	expectedProduct := entities.NewProduct(productID, "Test Product", price, "https://example.com/image.jpg")

	// Setup mocks
	mockRepo.EXPECT().GetProductByID(ctx, productID).Return(expectedProduct, nil)

	// Act
	result, err := useCase.Execute(ctx, productID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedProduct.ID, result.ID)
	assert.Equal(t, expectedProduct.Name, result.Name)
	assert.Equal(t, expectedProduct.Price, result.Price)
	assert.Equal(t, expectedProduct.ImageURL, result.ImageURL)
}

func TestGetProductUseCase_Execute_ProductNotFound(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockInventoryRepositoryFacade(ctrl)

	useCase := usecases.NewGetProductUseCase(mockRepo)

	ctx := context.Background()
	productID := "non-existent-product"

	// Setup mocks
	mockRepo.EXPECT().GetProductByID(ctx, productID).Return(nil, assert.AnError)

	// Act
	result, err := useCase.Execute(ctx, productID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetProductUseCase_Execute_EmptyProductID(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockInventoryRepositoryFacade(ctrl)

	useCase := usecases.NewGetProductUseCase(mockRepo)

	ctx := context.Background()
	productID := ""

	// Setup mocks
	mockRepo.EXPECT().GetProductByID(ctx, productID).Return(nil, assert.AnError)

	// Act
	result, err := useCase.Execute(ctx, productID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
}
