package usecases_test

import (
	"context"
	"testing"

	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/inventory-service/internal/application/usecases"
	"ecommerce-platform/services/inventory-service/tests/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// setupCreateProductTest creates common test setup for CreateProductUseCase tests
func setupCreateProductTest(t *testing.T) (*gomock.Controller, *mocks.MockInventoryRepositoryFacade, *usecases.CreateProductUseCase) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockInventoryRepositoryFacade(ctrl)
	useCase := usecases.NewCreateProductUseCase()
	return ctrl, mockRepo, useCase
}

func TestCreateProductUseCase_Execute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupCreateProductTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		name := "Test Product"
		currency, _ := valueobjects.NewCurrency("USD")
		price := valueobjects.NewMoney(1000, currency)
		imageURL := "https://example.com/image.jpg"

		// Setup mocks
		mockRepo.EXPECT().
			CreateProduct(ctx, gomock.Any()).
			Return(nil).
			Times(1)

		// Act
		result, err := useCase.Execute(ctx, name, price, imageURL, mockRepo)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, name, result.Name)
		assert.Equal(t, price, result.Price)
		assert.Equal(t, imageURL, result.ImageURL)
		assert.NotEmpty(t, result.ID)
		assert.False(t, result.CreatedAt.IsZero())
		assert.False(t, result.UpdatedAt.IsZero())
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupCreateProductTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		name := "Test Product"
		currency, _ := valueobjects.NewCurrency("USD")
		price := valueobjects.NewMoney(1000, currency)
		imageURL := "https://example.com/image.jpg"

		// Setup mocks
		mockRepo.EXPECT().
			CreateProduct(ctx, gomock.Any()).
			Return(assert.AnError).
			Times(1)

		// Act
		result, err := useCase.Execute(ctx, name, price, imageURL, mockRepo)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
