package usecases_test

import (
	"context"
	"testing"

	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/order-service/internal/application/usecases"
	"ecommerce-platform/services/order-service/internal/domain/aggregates"
	orderValueObjects "ecommerce-platform/services/order-service/internal/domain/valueobjects"
	"ecommerce-platform/services/order-service/tests/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// setupGetUserOrdersTest creates common test setup for GetUserOrdersUseCase tests
func setupGetUserOrdersTest(t *testing.T) (*gomock.Controller, *mocks.MockOrderRepositoryFacade, *usecases.GetUserOrdersUseCase) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockOrderRepositoryFacade(ctrl)
	useCase := usecases.NewGetUserOrdersUseCase()
	return ctrl, mockRepo, useCase
}

func TestGetUserOrdersUseCase_Execute(t *testing.T) {
	t.Run("success with orders", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupGetUserOrdersTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		userID := "user-456"
		page := 1
		limit := 10

		// Create test orders
		currency, _ := valueobjects.NewCurrency("USD")
		shippingAddress, _ := orderValueObjects.NewShippingAddress("123 Main St, New York, NY 10001, USA")

		order1, _ := aggregates.NewOrder(userID, shippingAddress, currency)
		order1.ID = "order-1"
		order2, _ := aggregates.NewOrder(userID, shippingAddress, currency)
		order2.ID = "order-2"

		expectedOrders := []*aggregates.Order{order1, order2}
		expectedTotal := int64(2)

		// Setup mocks
		mockRepo.EXPECT().
			GetByUserID(ctx, userID, page, limit).
			Return(expectedOrders, expectedTotal, nil).
			Times(1)

		// Act
		orders, total, err := useCase.Execute(ctx, userID, page, limit, mockRepo)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, orders)
		assert.Equal(t, expectedTotal, total)
		assert.Len(t, orders, 2)
		assert.Equal(t, "order-1", orders[0].ID)
		assert.Equal(t, "order-2", orders[1].ID)
	})

	t.Run("success with no orders", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupGetUserOrdersTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		userID := "user-456"
		page := 1
		limit := 10

		expectedOrders := []*aggregates.Order{}
		expectedTotal := int64(0)

		// Setup mocks
		mockRepo.EXPECT().
			GetByUserID(ctx, userID, page, limit).
			Return(expectedOrders, expectedTotal, nil).
			Times(1)

		// Act
		orders, total, err := useCase.Execute(ctx, userID, page, limit, mockRepo)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, orders)
		assert.Equal(t, expectedTotal, total)
		assert.Len(t, orders, 0)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupGetUserOrdersTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		userID := "user-456"
		page := 1
		limit := 10

		// Setup mocks
		mockRepo.EXPECT().
			GetByUserID(ctx, userID, page, limit).
			Return(nil, int64(0), assert.AnError).
			Times(1)

		// Act
		orders, total, err := useCase.Execute(ctx, userID, page, limit, mockRepo)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, orders)
		assert.Equal(t, int64(0), total)
	})
}
