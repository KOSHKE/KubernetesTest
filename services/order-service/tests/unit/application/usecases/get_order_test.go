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

// setupGetOrderTest creates common test setup for GetOrderUseCase tests
func setupGetOrderTest(t *testing.T) (*gomock.Controller, *mocks.MockOrderRepositoryFacade, *usecases.GetOrderUseCase) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockOrderRepositoryFacade(ctrl)
	useCase := usecases.NewGetOrderUseCase()
	return ctrl, mockRepo, useCase
}

func TestGetOrderUseCase_Execute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupGetOrderTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		orderID := "order-123"
		userID := "user-456"

		// Create test order
		currency, _ := valueobjects.NewCurrency("USD")
		shippingAddress, _ := orderValueObjects.NewShippingAddress("123 Main St, New York, NY 10001, USA")
		expectedOrder, _ := aggregates.NewOrder(userID, shippingAddress, currency)
		expectedOrder.ID = orderID

		// Setup mocks
		mockRepo.EXPECT().
			GetByID(ctx, orderID).
			Return(expectedOrder, nil).
			Times(1)

		// Act
		result, err := useCase.Execute(ctx, orderID, userID, mockRepo)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedOrder.ID, result.ID)
		assert.Equal(t, expectedOrder.UserID, result.UserID)
		assert.Equal(t, expectedOrder.Status, result.Status)
		assert.Equal(t, expectedOrder.Currency, result.Currency)
	})

	t.Run("order not found", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupGetOrderTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		orderID := "non-existent-order"
		userID := "user-456"

		// Setup mocks
		mockRepo.EXPECT().
			GetByID(ctx, orderID).
			Return(nil, assert.AnError).
			Times(1)

		// Act
		result, err := useCase.Execute(ctx, orderID, userID, mockRepo)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrOrderRetrievalFailed, err)
		assert.Nil(t, result)
	})

	t.Run("access denied - different user", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupGetOrderTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		orderID := "order-123"
		userID := "user-456"
		differentUserID := "user-789"

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
		result, err := useCase.Execute(ctx, orderID, userID, mockRepo)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrOrderAccessDenied, err)
		assert.Nil(t, result)
	})
}
