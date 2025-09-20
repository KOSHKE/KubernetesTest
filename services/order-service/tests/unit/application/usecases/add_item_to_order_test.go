package usecases_test

import (
	"context"
	"testing"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/order-service/internal/application/usecases"
	"ecommerce-platform/services/order-service/internal/domain/aggregates"
	orderValueObjects "ecommerce-platform/services/order-service/internal/domain/valueobjects"
	"ecommerce-platform/services/order-service/tests/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// setupAddItemToOrderTest creates common test setup for AddItemToOrderUseCase tests
func setupAddItemToOrderTest(t *testing.T) (*gomock.Controller, *mocks.MockOrderRepositoryFacade, *usecases.AddItemToOrderUseCase) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockOrderRepositoryFacade(ctrl)
	useCase := usecases.NewAddItemToOrderUseCase()
	return ctrl, mockRepo, useCase
}

func TestAddItemToOrderUseCase_Execute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupAddItemToOrderTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		orderID := "order-123"
		productID := "product-456"
		productName := "Test Product"
		quantity := int32(2)

		// Create test order
		currency, _ := valueobjects.NewCurrency("USD")
		price := valueobjects.NewMoney(1000, currency)
		shippingAddress, _ := orderValueObjects.NewShippingAddress("123 Main St, New York, NY 10001, USA")
		order, _ := aggregates.NewOrder("user-456", shippingAddress, currency)
		order.ID = orderID

		// Setup mocks
		mockRepo.EXPECT().
			GetByID(ctx, orderID).
			Return(order, nil).
			Times(1)

		mockRepo.EXPECT().
			Update(ctx, order).
			Return(nil).
			Times(1)

		// Act
		result, err := useCase.Execute(ctx, orderID, productID, productName, quantity, price, mockRepo)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, orderID, result.ID)
		assert.Len(t, result.Items, 1)
		assert.Equal(t, productID, result.Items[0].ProductID)
		assert.Equal(t, productName, result.Items[0].ProductName)
		assert.Equal(t, quantity, result.Items[0].Quantity)
	})

	t.Run("order not found", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupAddItemToOrderTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		orderID := "non-existent-order"
		productID := "product-456"
		productName := "Test Product"
		quantity := int32(2)
		currency, _ := valueobjects.NewCurrency("USD")
		price := valueobjects.NewMoney(1000, currency)

		// Setup mocks
		mockRepo.EXPECT().
			GetByID(ctx, orderID).
			Return(nil, assert.AnError).
			Times(1)

		// Act
		result, err := useCase.Execute(ctx, orderID, productID, productName, quantity, price, mockRepo)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrOrderRetrievalFailed, err)
		assert.Nil(t, result)
	})

	t.Run("update repository error", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupAddItemToOrderTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		orderID := "order-123"
		productID := "product-456"
		productName := "Test Product"
		quantity := int32(2)

		// Create test order
		currency, _ := valueobjects.NewCurrency("USD")
		price := valueobjects.NewMoney(1000, currency)
		shippingAddress, _ := orderValueObjects.NewShippingAddress("123 Main St, New York, NY 10001, USA")
		order, _ := aggregates.NewOrder("user-456", shippingAddress, currency)
		order.ID = orderID

		// Setup mocks
		mockRepo.EXPECT().
			GetByID(ctx, orderID).
			Return(order, nil).
			Times(1)

		mockRepo.EXPECT().
			Update(ctx, order).
			Return(assert.AnError).
			Times(1)

		// Act
		result, err := useCase.Execute(ctx, orderID, productID, productName, quantity, price, mockRepo)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrOrderPersistenceFailed, err)
		assert.Nil(t, result)
	})
}
