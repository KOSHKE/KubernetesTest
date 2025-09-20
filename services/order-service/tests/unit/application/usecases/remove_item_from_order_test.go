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

// setupRemoveItemFromOrderTest creates common test setup for RemoveItemFromOrderUseCase tests
func setupRemoveItemFromOrderTest(t *testing.T) (*gomock.Controller, *mocks.MockOrderRepositoryFacade, *usecases.RemoveItemFromOrderUseCase) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockOrderRepositoryFacade(ctrl)
	useCase := usecases.NewRemoveItemFromOrderUseCase()
	return ctrl, mockRepo, useCase
}

func TestRemoveItemFromOrderUseCase_Execute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupRemoveItemFromOrderTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		orderID := "order-123"
		userID := "user-456"
		productID := "product-456"

		// Create test order with an item
		currency, _ := valueobjects.NewCurrency("USD")
		price := valueobjects.NewMoney(1000, currency)
		shippingAddress, _ := orderValueObjects.NewShippingAddress("123 Main St, New York, NY 10001, USA")
		order, _ := aggregates.NewOrder(userID, shippingAddress, currency)
		order.ID = orderID
		// Add item to order first
		_ = order.AddItem(productID, "Test Product", 2, price)

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
		result, err := useCase.Execute(ctx, orderID, userID, productID, mockRepo)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, orderID, result.ID)
		assert.Equal(t, userID, result.UserID)
		assert.Len(t, result.Items, 0) // Item should be removed
	})

	t.Run("order not found", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupRemoveItemFromOrderTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		orderID := "non-existent-order"
		userID := "user-456"
		productID := "product-456"

		// Setup mocks
		mockRepo.EXPECT().
			GetByID(ctx, orderID).
			Return(nil, assert.AnError).
			Times(1)

		// Act
		result, err := useCase.Execute(ctx, orderID, userID, productID, mockRepo)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrOrderRetrievalFailed, err)
		assert.Nil(t, result)
	})

	t.Run("access denied - different user", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupRemoveItemFromOrderTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		orderID := "order-123"
		userID := "user-456"
		differentUserID := "user-789"
		productID := "product-456"

		// Create test order with different user
		currency, _ := valueobjects.NewCurrency("USD")
		shippingAddress, _ := orderValueObjects.NewShippingAddress("123 Main St, New York, NY 10001, USA")
		order, _ := aggregates.NewOrder(differentUserID, shippingAddress, currency)
		order.ID = orderID

		// Setup mocks
		mockRepo.EXPECT().
			GetByID(ctx, orderID).
			Return(order, nil).
			Times(1)

		// Act
		result, err := useCase.Execute(ctx, orderID, userID, productID, mockRepo)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrOrderAccessDenied, err)
		assert.Nil(t, result)
	})

	t.Run("update repository error", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupRemoveItemFromOrderTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		orderID := "order-123"
		userID := "user-456"
		productID := "product-456"

		// Create test order with an item
		currency, _ := valueobjects.NewCurrency("USD")
		price := valueobjects.NewMoney(1000, currency)
		shippingAddress, _ := orderValueObjects.NewShippingAddress("123 Main St, New York, NY 10001, USA")
		order, _ := aggregates.NewOrder(userID, shippingAddress, currency)
		order.ID = orderID
		_ = order.AddItem(productID, "Test Product", 2, price)

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
		result, err := useCase.Execute(ctx, orderID, userID, productID, mockRepo)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrOrderPersistenceFailed, err)
		assert.Nil(t, result)
	})
}
