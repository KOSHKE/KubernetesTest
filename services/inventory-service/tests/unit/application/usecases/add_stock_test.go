package usecases_test

import (
	"context"
	"testing"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/inventory-service/internal/application/usecases"
	"ecommerce-platform/services/inventory-service/internal/domain/entities"
	"ecommerce-platform/services/inventory-service/tests/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// setupAddStockTest creates common test setup for AddStockUseCase tests
func setupAddStockTest(t *testing.T) (*gomock.Controller, *mocks.MockInventoryRepositoryFacade, *usecases.AddStockUseCase) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockInventoryRepositoryFacade(ctrl)
	useCase := usecases.NewAddStockUseCase()
	return ctrl, mockRepo, useCase
}

func TestAddStockUseCase_Execute(t *testing.T) {
	t.Run("success with existing stock", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupAddStockTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		productID := "product-123"
		quantity := int32(50)

		// Setup test data
		currency, _ := valueobjects.NewCurrency("USD")
		money := valueobjects.NewMoney(1000, currency)
		product := entities.NewProduct(productID, "Test Product", money, "image.jpg")
		existingStock := entities.NewStock(productID, 100, 0) // Already has 100 units

		// Setup mocks
		mockRepo.EXPECT().
			GetProductByID(ctx, productID).
			Return(product, nil).
			Times(1)

		mockRepo.EXPECT().
			GetStockByProductID(ctx, productID).
			Return(existingStock, nil).
			Times(1)

		mockRepo.EXPECT().
			UpsertStock(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, stock *entities.Stock) error {
				assert.Equal(t, int32(150), stock.AvailableQuantity) // 100 + 50
				assert.Equal(t, int32(0), stock.ReservedQuantity)
				return nil
			}).
			Times(1)

		// Act
		result, err := useCase.Execute(ctx, productID, quantity, mockRepo)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int32(150), result.AvailableQuantity) // 100 + 50
		assert.Equal(t, int32(0), result.ReservedQuantity)
	})

	t.Run("success with new stock", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupAddStockTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		productID := "product-123"
		quantity := int32(50)

		// Setup test data
		currency, _ := valueobjects.NewCurrency("USD")
		money := valueobjects.NewMoney(1000, currency)
		product := entities.NewProduct(productID, "Test Product", money, "image.jpg")

		// Setup mocks - stock doesn't exist
		mockRepo.EXPECT().
			GetProductByID(ctx, productID).
			Return(product, nil).
			Times(1)

		mockRepo.EXPECT().
			GetStockByProductID(ctx, productID).
			Return(nil, errors.ErrStockNotFound).
			Times(1)

		mockRepo.EXPECT().
			UpsertStock(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, stock *entities.Stock) error {
				assert.Equal(t, int32(50), stock.AvailableQuantity) // New stock with 50 units
				assert.Equal(t, int32(0), stock.ReservedQuantity)
				return nil
			}).
			Times(1)

		// Act
		result, err := useCase.Execute(ctx, productID, quantity, mockRepo)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int32(50), result.AvailableQuantity) // New stock with 50 units
		assert.Equal(t, int32(0), result.ReservedQuantity)
	})

	t.Run("product not found", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupAddStockTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		productID := "non-existent-product"
		quantity := int32(50)

		// Setup mocks
		mockRepo.EXPECT().
			GetProductByID(ctx, productID).
			Return(nil, errors.ErrProductNotFound).
			Times(1)

		// Act
		result, err := useCase.Execute(ctx, productID, quantity, mockRepo)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrProductNotFound, err)
		assert.Nil(t, result)
	})

	t.Run("product is nil", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupAddStockTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		productID := "product-123"
		quantity := int32(50)

		// Setup mocks
		mockRepo.EXPECT().
			GetProductByID(ctx, productID).
			Return(nil, nil).
			Times(1)

		// Act
		result, err := useCase.Execute(ctx, productID, quantity, mockRepo)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrProductNotFound, err)
		assert.Nil(t, result)
	})

	t.Run("upsert stock error", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupAddStockTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		productID := "product-123"
		quantity := int32(50)

		// Setup test data
		currency, _ := valueobjects.NewCurrency("USD")
		money := valueobjects.NewMoney(1000, currency)
		product := entities.NewProduct(productID, "Test Product", money, "image.jpg")
		existingStock := entities.NewStock(productID, 100, 0)

		// Setup mocks
		mockRepo.EXPECT().
			GetProductByID(ctx, productID).
			Return(product, nil).
			Times(1)

		mockRepo.EXPECT().
			GetStockByProductID(ctx, productID).
			Return(existingStock, nil).
			Times(1)

		mockRepo.EXPECT().
			UpsertStock(ctx, gomock.Any()).
			Return(assert.AnError).
			Times(1)

		// Act
		result, err := useCase.Execute(ctx, productID, quantity, mockRepo)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
