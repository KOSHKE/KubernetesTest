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

// setupUpdateOrderStatusTest creates common test setup for UpdateOrderStatusUseCase tests
func setupUpdateOrderStatusTest(t *testing.T) (*gomock.Controller, *mocks.MockOrderRepositoryFacade, *usecases.UpdateOrderStatusUseCase) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockOrderRepositoryFacade(ctrl)
	useCase := usecases.NewUpdateOrderStatusUseCase()
	return ctrl, mockRepo, useCase
}

func TestUpdateOrderStatusUseCase_Execute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupUpdateOrderStatusTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		orderID := "order-123"
		newStatus := orderValueObjects.OrderStatusConfirmed

		// Create test order
		currency, _ := valueobjects.NewCurrency("USD")
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
		result, err := useCase.Execute(ctx, orderID, newStatus, mockRepo)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, orderID, result.ID)
		assert.Equal(t, newStatus, result.Status)
	})

	t.Run("order not found", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupUpdateOrderStatusTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		orderID := "non-existent-order"
		newStatus := orderValueObjects.OrderStatusConfirmed

		// Setup mocks
		mockRepo.EXPECT().
			GetByID(ctx, orderID).
			Return(nil, assert.AnError).
			Times(1)

		// Act
		result, err := useCase.Execute(ctx, orderID, newStatus, mockRepo)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrOrderRetrievalFailed, err)
		assert.Nil(t, result)
	})

	t.Run("update repository error", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupUpdateOrderStatusTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		orderID := "order-123"
		newStatus := orderValueObjects.OrderStatusConfirmed

		// Create test order
		currency, _ := valueobjects.NewCurrency("USD")
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
		result, err := useCase.Execute(ctx, orderID, newStatus, mockRepo)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrOrderPersistenceFailed, err)
		assert.Nil(t, result)
	})
}
