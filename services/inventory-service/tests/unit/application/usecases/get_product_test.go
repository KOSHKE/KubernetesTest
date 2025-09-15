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

// setupGetProductTest creates common test setup for GetProductUseCase tests
func setupGetProductTest(t *testing.T) (*gomock.Controller, *mocks.MockInventoryRepositoryFacade, *usecases.GetProductUseCase) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockInventoryRepositoryFacade(ctrl)
	useCase := usecases.NewGetProductUseCase()
	return ctrl, mockRepo, useCase
}

func TestGetProductUseCase_Execute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupGetProductTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		productID := "product-123"

		// Create test product
		currency, _ := valueobjects.NewCurrency("USD")
		price := valueobjects.NewMoney(1000, currency)
		expectedProduct := entities.NewProduct(productID, "Test Product", price, "https://example.com/image.jpg")

		// Setup mocks
		mockRepo.EXPECT().
			GetProductByID(ctx, productID).
			Return(expectedProduct, nil).
			Times(1)

		// Act
		result, err := useCase.Execute(ctx, productID, mockRepo)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedProduct.ID, result.ID)
		assert.Equal(t, expectedProduct.Name, result.Name)
		assert.Equal(t, expectedProduct.Price, result.Price)
		assert.Equal(t, expectedProduct.ImageURL, result.ImageURL)
	})

	t.Run("product not found", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupGetProductTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		productID := "non-existent-product"

		// Setup mocks
		mockRepo.EXPECT().
			GetProductByID(ctx, productID).
			Return(nil, assert.AnError).
			Times(1)

		// Act
		result, err := useCase.Execute(ctx, productID, mockRepo)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("empty product id", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupGetProductTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		productID := ""

		// Setup mocks
		mockRepo.EXPECT().
			GetProductByID(ctx, productID).
			Return(nil, assert.AnError).
			Times(1)

		// Act
		result, err := useCase.Execute(ctx, productID, mockRepo)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
